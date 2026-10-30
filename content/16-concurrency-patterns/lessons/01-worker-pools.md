---
title: Worker Pools and Producer/Consumer
minutes: 10
---
You already know goroutines, channels and `sync.WaitGroup`. This chapter combines them into reusable **patterns**. The first and most common is the **worker pool**: a fixed number of goroutines pulling jobs from a shared channel.

## Why a pool?

Starting one goroutine per job is cheap, but not free of consequences: ten million jobs means ten million goroutines, each holding memory and possibly a database connection or a file handle. A pool keeps the number of concurrent workers **bounded** while the queue of work can be as long as you like.

```text
producer ──► jobs channel ──► worker 1 ─┐
                          ├─► worker 2 ─┼──► results
                          └─► worker 3 ─┘
```

Workers `range` over the jobs channel. When the producer closes it, every worker's loop ends and the worker returns. This is the **producer/consumer** pattern: one side produces values, the other consumes them, and the channel decouples their speed.

## The ordering problem

Workers finish in an unpredictable order. If you need results in input order, do **not** try to make the goroutines take turns. Instead, give every job an **id** (its index) and let each worker write into its own slot of a pre-sized slice. Different goroutines writing to different elements is not a data race, and after `wg.Wait()` all writes are visible to the reader.

```go run
package main

import (
	"fmt"
	"sync"
)

func main() {
	const workers = 3
	inputs := []int{4, 8, 15, 16, 23, 42}

	jobs := make(chan int) // carries indexes into inputs
	results := make([]int, len(inputs))

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				results[id] = inputs[id] * inputs[id] // own slot, no lock needed
			}
		}()
	}

	for id := range inputs {
		jobs <- id
	}
	close(jobs) // no more work: workers' range loops end
	wg.Wait()

	fmt.Println(results)
}
```

The output order always matches the input order, whichever worker handled which job. Closing `jobs` is the producer's job; workers only receive.

## Collecting results through a channel

When results are not index-addressable, or you want to start processing them while others are still running, send them on a results channel. The tricky part is **who closes it**: no worker may, because the others still send. Use one extra goroutine that waits for the pool and then closes the channel.

```go run
package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type job struct {
	id   int
	text string
}

type result struct {
	id    int
	upper string
}

func worker(jobs <-chan job, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		results <- result{id: j.id, upper: strings.ToUpper(j.text)}
	}
}

func main() {
	words := []string{"alpha", "beta", "gamma", "delta", "epsilon"}

	jobs := make(chan job)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// Producer: its own goroutine so it can send while we consume below.
	go func() {
		defer close(jobs)
		for i, w := range words {
			jobs <- job{id: i, text: w}
		}
	}()

	// Closer: the only goroutine that knows when all senders are finished.
	go func() {
		wg.Wait()
		close(results)
	}()

	var all []result
	for r := range results { // ends when results is closed
		all = append(all, r)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].id < all[j].id })

	for _, r := range all {
		fmt.Println(r.id, r.upper)
	}
}
```

Notice the directions in the signature: `jobs <-chan job` (receive only) and `results chan<- result` (send only). They document intent and the compiler enforces it.

## Why the producer runs in its own goroutine

With unbuffered channels, `jobs <- job` blocks until a worker receives. If `main` sent every job itself before reading any result, workers would block sending results while `main` blocks sending jobs: a **deadlock**. Either run the producer concurrently (as above), or buffer the channels so the whole batch fits.

```go
// Deadlock-prone: main sends all jobs before reading any results.
for _, j := range allJobs {
	jobs <- j // blocks once the workers are stuck on `results <-`
}
close(jobs)
for r := range results { ... }
```

> **Common mistake:** closing the results channel from a worker, or from `main` before `wg.Wait()`. A send on a closed channel panics. Exactly one goroutine, the one that knows all senders are done, should close.

> **Tip:** pick the pool size from the nature of the work. CPU-bound work rarely benefits from more than `runtime.NumCPU()` workers; I/O-bound work (HTTP calls, queries) can use many more, limited by the remote system.

## Key takeaways

- A worker pool is N goroutines ranging over one jobs channel; closing the channel stops them.
- Only the sending side closes a channel; with several senders, close from a goroutine that waits on a `WaitGroup`.
- For ordered results, send job ids and have each worker write `results[id]`; no lock is needed for distinct elements.
- Use directional channel types (`<-chan`, `chan<-`) in function signatures.
- Run the producer in its own goroutine, or buffer the channels, to avoid a send/receive deadlock.
