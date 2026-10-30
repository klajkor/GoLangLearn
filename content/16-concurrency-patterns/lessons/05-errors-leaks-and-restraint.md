---
title: Errors, Leaks and When Not to Use Concurrency
minutes: 10
---
The last piece of the toolbox is correctness: how to report errors from goroutines, how to make sure none of them get stuck forever, and when to skip concurrency altogether.

## Errors from goroutines

A goroutine cannot `return` an error to its starter. Common solutions are an `errs` channel, or a struct containing both value and error sent on the results channel. When you start a group of tasks and want "the first error, and stop the rest", you want an **errgroup**.

The standard solution is `golang.org/x/sync/errgroup`:

```go
g, ctx := errgroup.WithContext(parent)
for _, url := range urls {
	g.Go(func() error { return fetch(ctx, url) })
}
if err := g.Wait(); err != nil { ... } // first non-nil error
```

It is a separate module (`go get golang.org/x/sync/errgroup`), so the examples here use a tiny hand-written version. The whole idea is just three pieces: a `WaitGroup`, a `sync.Once` to remember the **first** error, and a `cancel` function that tells the siblings to stop.

```go run
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type Group struct {
	wg     sync.WaitGroup
	once   sync.Once
	err    error
	cancel context.CancelFunc
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.once.Do(func() {
				g.err = err // only the first error is kept
				g.cancel()  // tell everyone else to stop
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel() // release context resources
	return g.err
}

func main() {
	g, ctx := WithContext(context.Background())
	var stopped atomic.Int32

	for i := 0; i < 5; i++ {
		g.Go(func() error {
			if i == 2 {
				return errors.New("task 2 failed")
			}
			<-ctx.Done() // siblings wait until cancelled
			stopped.Add(1)
			return ctx.Err()
		})
	}

	err := g.Wait()
	fmt.Println("error:", err)
	fmt.Println("tasks that stopped early:", stopped.Load())
}
```

(The loop variable `i` is per-iteration since Go 1.22, so each closure sees its own value.) Why is the result deterministic? Only task 2 returns an error without waiting, and the siblings only return (with `context.Canceled`) **after** `cancel()` was called, which happens inside the `once.Do` of task 2. So the recorded error is always "task 2 failed", and all four siblings observe the cancellation before `Wait` returns.

> **Tip:** `errgroup` also has `SetLimit(n)` to bound the number of active goroutines, replacing the semaphore from the previous lesson, and Go 1.20+ offers `errors.Join` if you want to report all errors instead of the first.

## Goroutine leaks

A **leak** is a goroutine that can never finish, usually because it is blocked on a channel nobody will use again. It is not caught by the compiler or garbage collector: blocked goroutines are considered alive. The usual causes:

- sending on a channel whose receiver gave up (early return, timeout);
- receiving on a channel that is never closed;
- a `for { select { ... } }` loop with no exit case;
- a forgotten `ticker.Stop()` or `cancel()`.

Here is the classic first-result-wins leak:

```go
func firstReply(servers []string) string {
	ch := make(chan string) // unbuffered
	for _, s := range servers {
		go func() { ch <- query(s) }()
	}
	return <-ch // the other goroutines block on `ch <-` forever
}
```

The fix is a **buffered channel** large enough for every sender (so the losers can deposit their value and exit), or a context so that losers abort.

```go run
package main

import "fmt"

func query(server string) string { return "answer" } // same from every server

func firstReply(servers []string) string {
	ch := make(chan string, len(servers)) // room for every sender
	for _, s := range servers {
		go func() { ch <- query(s) }()
	}
	return <-ch // the rest finish in the background; nobody is stuck
}

func main() {
	fmt.Println(firstReply([]string{"a", "b", "c"}))
}
```

### A leak checklist

For every goroutine you start, you should be able to answer: **how does it stop, and who makes that happen?** Good answers are "its input channel is closed", "the context is cancelled", or "it finishes its fixed work and the `WaitGroup` tracks it". In tests, tools such as `go.uber.org/goleak` fail a test if goroutines remain after it.

## Bounded concurrency, summarised

| Technique | Bounds |
|---|---|
| Worker pool | number of goroutines |
| Semaphore (buffered channel) | number of concurrently active tasks |
| `errgroup.SetLimit` | the same, plus error propagation |
| Rate limiter | operations per time |

## When NOT to use concurrency

Concurrency is a structuring tool first and a speed-up second. Skip it when:

- The work is tiny. Starting goroutines and passing values costs more than computing `n*n`.
- The steps depend on each other. A strictly sequential algorithm gains nothing.
- There is a shared bottleneck (a single lock, a single database connection): the goroutines just queue.
- The code becomes harder to reason about than a loop. Races and leaks are bugs that tests find rarely.
- A single-threaded design with one `for` loop and a map is enough for your throughput.

A good workflow: write it sequentially, measure, then add concurrency where the profile shows I/O waiting or idle cores. Run tests with `go test -race`, and remember that a data race is a bug even when the program seems to work.

> **Common mistake:** "more goroutines means faster". CPU-bound work saturates at `GOMAXPROCS` cores; beyond that you only add scheduling overhead.

## Key takeaways

- Goroutines cannot return errors directly; collect them via channels or an errgroup.
- An errgroup is a `WaitGroup` plus `sync.Once` for the first error plus a `cancel` for the siblings; use `golang.org/x/sync/errgroup` in real projects.
- A leaked goroutine is blocked forever; always know how each goroutine ends.
- Buffered channels sized for all senders and context cancellation are the standard cures.
- Do not use concurrency for tiny, dependent or bottlenecked work; measure first and test with `-race`.
