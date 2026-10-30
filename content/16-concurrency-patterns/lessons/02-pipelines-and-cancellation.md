---
title: Pipelines and Cancellation
minutes: 10
---
A **pipeline** is a series of stages connected by channels. Each stage is a function that receives values from an inbound channel, does some work, and sends results on an outbound channel it owns. Stages run concurrently, so while stage 2 handles item 1, stage 1 can already produce item 2.

## The shape of a stage

```go
func stage(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)        // the stage owns and closes its output
		for v := range in {     // runs until the previous stage closes `in`
			out <- transform(v)
		}
	}()
	return out
}
```

Three rules make pipelines composable:

1. A stage **creates, owns and closes** its output channel.
2. A stage keeps receiving until its input is closed.
3. Stages take `<-chan T` and return `<-chan T`, so they chain like `square(double(gen(...)))`.

```go run
package main

import "fmt"

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

func addOne(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n + 1
		}
	}()
	return out
}

func main() {
	// Each stage has a single goroutine, so order is preserved.
	for v := range addOne(square(gen(1, 2, 3, 4))) {
		fmt.Println(v)
	}
}
```

Because every stage here has exactly one goroutine, items stay in order. Order is only at risk when you run several workers in one stage (next lesson).

## The leak problem and cancellation

What if the consumer stops early? Imagine `main` reads only the first value and walks away. The stages upstream are blocked on `out <- v` forever: they are **leaked goroutines**, and they keep their memory for the life of the program.

The fix is a cancellation signal that every send can race against. In modern Go that signal is a `context.Context`; every blocking send becomes a `select` with `ctx.Done()`.

```go run
package main

import (
	"context"
	"fmt"
	"sync"
)

// naturals produces 0, 1, 2, ... until ctx is cancelled.
func naturals(ctx context.Context, wg *sync.WaitGroup) <-chan int {
	out := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for i := 0; ; i++ {
			select {
			case out <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func square(ctx context.Context, in <-chan int, wg *sync.WaitGroup) <-chan int {
	out := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	squares := square(ctx, naturals(ctx, &wg), &wg)

	for i := 0; i < 5; i++ {
		fmt.Println(<-squares)
	}

	cancel()  // tell every stage to stop
	wg.Wait() // prove they all returned: nothing is leaked
	fmt.Println("all stages stopped")
}
```

The `WaitGroup` here is only for demonstration (and for tests): it shows that cancellation really terminates each stage. In real code you usually just call `cancel()` (often with `defer`) and the stages unwind on their own.

> **Tip:** always `defer cancel()` right after `context.WithCancel`, even if you also cancel explicitly. It guarantees the context's resources are released on every return path.

## `done` channels, the pre-context style

Before `context`, the same idea used a `done <-chan struct{}` that was closed to broadcast cancellation. Closing a channel wakes **every** receiver, which is exactly what you want for "stop all". You will still see this in older code; `ctx.Done()` is that same channel with extras (deadlines, error reason, values).

## The or-done helper

Reading from a channel while also watching for cancellation gets repetitive. A small generic helper wraps it:

```go run
package main

import (
	"context"
	"fmt"
)

func orDone[T any](ctx context.Context, in <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := make(chan string)
	go func() {
		defer close(src)
		for _, s := range []string{"a", "b", "c"} {
			src <- s
		}
	}()

	for s := range orDone(ctx, src) {
		fmt.Println(s)
	}
}
```

Now the consumer is a simple `range` loop, and cancelling `ctx` ends it without leaving the helper stuck. Note the inner `select` around the send: forgetting it would let the helper block forever if the consumer has stopped listening.

> **Common mistake:** writing `out <- v` as a bare statement in a long-lived stage. If nobody receives any more, that goroutine is stuck permanently. Every send or receive that can block indefinitely should have a cancellation escape.

## Key takeaways

- A pipeline stage owns its output channel, closes it when done, and ranges over its input.
- Stages that are single goroutines preserve order; stages with several workers do not.
- An abandoned pipeline leaks goroutines unless every blocking operation can be interrupted.
- Use `select` with `ctx.Done()` around sends (and receives that may wait forever); cancel with `defer cancel()`.
- An `orDone` wrapper keeps consumer code readable while staying cancellable.
