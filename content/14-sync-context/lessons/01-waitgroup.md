---
title: Waiting with sync.WaitGroup
minutes: 8
---
In the last chapter we waited for goroutines by closing or receiving on channels. When you simply need to say "wait until all of these finished", the `sync` package offers a tidier tool: **`sync.WaitGroup`**.

A WaitGroup is a counter with three operations:

| Method | Effect |
|---|---|
| `wg.Add(n)` | add `n` to the counter (usually 1 per goroutine) |
| `wg.Done()` | subtract one (same as `Add(-1)`) |
| `wg.Wait()` | block until the counter is zero |

## The standard pattern

```go run
package main

import (
	"fmt"
	"sync"
)

func main() {
	squares := make([]int, 5)
	var wg sync.WaitGroup

	for i := range squares {
		wg.Add(1) // before the go statement
		go func() {
			defer wg.Done()
			squares[i] = (i + 1) * (i + 1) // each goroutine writes its own slot
		}()
	}

	wg.Wait()
	fmt.Println(squares)
}
```

Two details make this correct and race-free:

- `Add(1)` runs in the **calling** goroutine, before `go`. If it were inside the new goroutine, `Wait` might see a zero counter and return too early.
- Every goroutine writes to a **different** slice element, so no two goroutines touch the same memory. After `Wait` returns, everything the goroutines wrote is visible to `main`.

`defer wg.Done()` guarantees the counter is decremented even if the function returns early or panics.

> **Common mistake:** passing a `WaitGroup` by value. Copying it copies the counter, so `Done` is called on a different group and `Wait` hangs. Pass `*sync.WaitGroup`, or capture the variable in a closure as above. `go vet` warns about copied sync types.

## Collecting results through a channel

If results do not fit a pre-sized slice, send them on a channel. A tiny extra goroutine closes the channel once everyone is done, so the consumer can `range` over it:

```go run
package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	results := make(chan int)
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- i * 100
		}()
	}

	go func() {
		wg.Wait()
		close(results) // lets the range below finish
	}()

	var all []int
	for r := range results {
		all = append(all, r)
	}
	sort.Ints(all)
	fmt.Println(all)
}
```

Note that `Wait` and `close` must live in their own goroutine. If `main` called `wg.Wait()` first, the workers would block forever sending on the unbuffered channel and nobody would be receiving: a deadlock.

## wg.Go (Go 1.25)

The `Add(1)` / `go` / `defer Done()` dance is so common that Go 1.25 added a helper. `wg.Go(f)` increments the counter, runs `f` in a new goroutine, and calls `Done` when `f` returns:

```go
var wg sync.WaitGroup
for i := range 5 {
	wg.Go(func() {
		work(i)
	})
}
wg.Wait()
```

It needs Go 1.25 or newer, so older codebases and the examples above still use the classic form. Behaviour is identical.

## Panics and negative counters

If the counter goes below zero, `Done` panics with `sync: negative WaitGroup counter`. That is a bug: more `Done` calls than `Add` calls.

A WaitGroup can be reused after `Wait` returns, but only once all previous `Wait` calls are finished. Do not call `Add` with a positive value concurrently with `Wait` when the counter is zero.

## Errors from workers

A WaitGroup only counts. It collects neither results nor errors, and it cannot stop other goroutines when one fails. For that you combine it with channels, or use `errgroup` (see the last lesson of this chapter).

## Key takeaways

- `Add` before `go`, `defer Done` inside, `Wait` in the goroutine that needs the results.
- Never copy a WaitGroup; pass a pointer or capture it.
- To `range` over a result channel, close it from a goroutine that does `wg.Wait()` first.
- Go 1.25 adds `wg.Go(func())` as a shortcut for the Add/go/Done pattern.
- A WaitGroup does not carry errors or cancellation; other tools handle those.
