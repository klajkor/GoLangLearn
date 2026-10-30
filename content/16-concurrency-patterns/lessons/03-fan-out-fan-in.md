---
title: Fan-out and Fan-in
minutes: 9
---
When one stage of a pipeline is slow, you can run several copies of it in parallel. That is **fan-out**: multiple goroutines reading from the same channel. **Fan-in** is the reverse: merging several channels into one so that a single consumer can read everything.

```text
                 ┌─► worker A ─┐
 source ──► in ──┼─► worker B ─┼──► merge ──► out ──► consumer
                 └─► worker C ─┘
      (fan-out: same channel)     (fan-in: many channels become one)
```

## Fan-out is just "many receivers"

A channel can have any number of goroutines receiving from it. Each value goes to exactly one receiver, so the work is automatically shared: a busy worker simply takes fewer items. You do not need to split the data yourself.

## Fan-in: merge

Fan-in needs one goroutine per input channel, all forwarding into a single output channel. A `WaitGroup` tells us when every input is exhausted so that a final goroutine can close the output.

```go run
package main

import (
	"fmt"
	"sort"
	"sync"
)

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// merge is generic: it works for channels of any element type.
func merge[T any](cs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan T) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	source := gen(1, 2, 3, 4, 5, 6, 7, 8, 9)

	// Fan-out: three workers read from the same channel.
	w1 := square(source)
	w2 := square(source)
	w3 := square(source)

	// Fan-in: merge their outputs.
	var got []int
	for v := range merge(w1, w2, w3) {
		got = append(got, v)
	}

	// Arrival order is not deterministic, so sort before looking at it.
	sort.Ints(got)
	fmt.Println(got)
}
```

Look at the closing logic: `merge` must not close `out` until **all** forwarding goroutines have finished, so the closer goroutine waits on the `WaitGroup`. The same trick appeared in the worker pool lesson.

> **Common mistake:** assuming the output of a fan-out stage is in input order. With three workers, 9 may well arrive before 1. If the order matters, carry an index with each item (a struct `{id, value}`) and sort or place by id at the end, as in the worker pool lesson.

## Fan-in with cancellation

A real merge should stop when the consumer gives up. Add `ctx` and select around the forwarding send:

```go
func merge[T any](ctx context.Context, cs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan T) {
			defer wg.Done()
			for v := range c {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}(c)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}
```

Without the `select`, an abandoned `out` would strand every forwarding goroutine. Note that the upstream producers need the same cancellation, or they would be blocked trying to feed workers that returned.

## Fan-out with a function: bounded parallel map

Fan-out/fan-in often appears in one function: "run `f` on every item with at most N workers and give me the results". Combining the index trick from the last lesson makes the order deterministic:

```go run
package main

import (
	"fmt"
	"strings"
	"sync"
)

func ParallelMap[T, U any](items []T, workers int, f func(T) U) []U {
	results := make([]U, len(items))
	idx := make(chan int)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range idx {
				results[i] = f(items[i])
			}
		}()
	}
	for i := range items {
		idx <- i
	}
	close(idx)
	wg.Wait()
	return results
}

func main() {
	words := []string{"go", "is", "fun", "and", "fast"}
	upper := ParallelMap(words, 3, func(s string) string {
		return strings.ToUpper(s)
	})
	fmt.Println(upper)
}
```

It combines three earlier ideas: generics for the element types, a pool for bounded concurrency, and slot-per-index results for ordering. Only the return value crosses back to the caller, after `wg.Wait()` has established that all writes happened.

> **Tip:** fan-out only helps when the work in each item is larger than the cost of passing it through a channel (a few hundred nanoseconds). For tiny operations like `n*n`, a plain loop is faster; the examples use trivial work purely for clarity.

## Key takeaways

- Fan-out: several goroutines receive from one channel; each item is processed by exactly one of them.
- Fan-in: `merge` starts one forwarder per input and closes the output once a `WaitGroup` says all are done.
- Fan-out destroys ordering; restore it with ids or by sorting at the end.
- Put `select` with `ctx.Done()` around sends so abandoned pipelines do not leak goroutines.
- Make sure the work per item justifies the channel overhead.
