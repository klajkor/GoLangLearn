---
title: Semaphores and Rate Limiting
minutes: 9
---
Sometimes you do want one goroutine per task, but you do not want all of them active at once: 1,000 HTTP calls should not become 1,000 simultaneous connections. Two related tools control that: **bounding concurrency** (how many at the same time) and **rate limiting** (how many per unit of time).

## Semaphore with a buffered channel

A buffered channel with capacity N is a counting semaphore. Sending a value **acquires** a slot (blocking when all N are taken); receiving **releases** it.

```go
sem := make(chan struct{}, 3) // at most 3 holders at once

sem <- struct{}{}        // acquire (blocks when full)
defer func() { <-sem }() // release
```

`struct{}` carries no data and takes no memory; only the count matters. This is how you get **bounded concurrency** while still spawning one goroutine per task.

Here is a runnable proof. Eight tasks try to run, but an atomic counter shows that never more than 3 are active at the same moment:

```go run
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	const limit = 3
	const tasks = 8

	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup

	var active, peak atomic.Int32
	results := make([]int, tasks) // one slot per task: deterministic order

	for i := 0; i < tasks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			n := active.Add(1)
			for { // record the highest value seen
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond) // simulate work
			active.Add(-1)

			results[i] = i * 10
		}()
	}
	wg.Wait()

	fmt.Println(results)
	fmt.Println("never above the limit:", peak.Load() <= limit)
}
```

The printed claim `peak <= limit` is guaranteed by the semaphore, not by timing. (The sleep is only there to make tasks overlap; correctness does not depend on it.) Compare with the worker pool: the pool starts N goroutines and feeds them; the semaphore starts a goroutine per task and makes them queue up. Pools are cheaper for huge workloads, semaphores are simpler when tasks are few or naturally independent.

> **Tip:** the library `golang.org/x/sync/semaphore` offers a weighted semaphore (`Acquire(ctx, n)`), useful when tasks need different amounts of a resource.

> **Common mistake:** forgetting to release on one code path (an early `return` or an error). The slot is lost, and after N such losses every task blocks forever. Always `defer` the release right after the acquire.

## Rate limiting with a ticker

Limiting *rate* means spacing work out in time. The simplest limiter is a `time.Ticker`: wait for the next tick before every request.

```go run
package main

import (
	"fmt"
	"time"
)

func main() {
	limiter := time.NewTicker(10 * time.Millisecond)
	defer limiter.Stop()

	start := time.Now()
	handled := 0
	for i := 1; i <= 5; i++ {
		<-limiter.C // at most one request per tick
		handled++
	}

	elapsed := time.Since(start)
	fmt.Println("handled:", handled)
	fmt.Println("took at least 40ms:", elapsed >= 40*time.Millisecond)
}
```

Five requests need five ticks, so the loop cannot finish in under 40 ms. We only assert the **lower bound**; the upper bound depends on the scheduler, so a deterministic program must not print it.

## Token bucket: bursts allowed

A strict ticker never lets a burst through. A **token bucket** does: the bucket holds up to `burst` tokens, each request takes one, and a background goroutine adds a token at a steady rate (dropping it if the bucket is full). A buffered channel makes a neat bucket.

```go run
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const burst = 3
	bucket := make(chan struct{}, burst)
	for i := 0; i < burst; i++ { // start full
		bucket <- struct{}{}
	}

	// Refill one token every 10ms until cancelled.
	go func() {
		t := time.NewTicker(10 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				select {
				case bucket <- struct{}{}:
				default: // bucket full: drop the token
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	start := time.Now()
	for i := 1; i <= 5; i++ {
		<-bucket // take a token (waits if empty)
	}
	elapsed := time.Since(start)

	// 3 requests used the burst; 2 more had to wait for refills.
	fmt.Println("5 requests served")
	fmt.Println("waited for at least two refills:", elapsed >= 20*time.Millisecond)
}
```

The first three receives succeed immediately. The fourth and fifth must wait for two ticks, so at least ~20 ms must pass. In production code, prefer the maintained `golang.org/x/time/rate` package, which implements this algorithm (`rate.NewLimiter(rate.Every(...), burst)` and `limiter.Wait(ctx)`) with correct handling of time, fractional tokens and cancellation.

## Choosing the right tool

| Need | Tool |
|---|---|
| At most N tasks running at once | buffered-channel semaphore or a worker pool |
| At most N operations per second | `time.Ticker` or `x/time/rate` |
| Allow short bursts, enforce average rate | token bucket |
| Weighted resources (memory, bytes) | `x/sync/semaphore` |

## Key takeaways

- A buffered channel of `struct{}` is a counting semaphore: send to acquire, receive to release.
- Always release with `defer` immediately after acquiring.
- A `time.Ticker` gives a simple rate limit; stop it with `defer ticker.Stop()`.
- A token bucket is a buffered channel plus a refill goroutine; it allows bursts up to its capacity.
- In tests and demos assert lower bounds on elapsed time, never upper bounds.
