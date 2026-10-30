---
title: Cancellation with context
minutes: 11
---
How do you tell a goroutine to stop, give up after two seconds, or abandon work because the user closed the browser tab? With a **`context.Context`**. A context carries a cancellation signal, an optional deadline, and a few request-scoped values through a call chain.

## The interface

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{} // closed when the context is cancelled or times out
	Err() error            // nil until Done is closed, then why
	Value(key any) any
}
```

The key is `Done()`: a channel you can `select` on. Once it is closed, `Err()` returns either `context.Canceled` or `context.DeadlineExceeded`.

## Creating contexts

Every tree of contexts starts at a root:

- `context.Background()`: the empty root, used in `main`, tests and top-level code.
- `context.TODO()`: the same, but signals "I have not decided which context to use here yet".

From a parent you derive children:

| Function | Result |
|---|---|
| `WithCancel(parent)` | child plus a `cancel` function |
| `WithTimeout(parent, d)` | cancelled automatically after duration `d` |
| `WithDeadline(parent, t)` | cancelled at time `t` |
| `WithValue(parent, key, val)` | child carrying a value |

`WithTimeout` is just `WithDeadline(parent, time.Now().Add(d))`.

> **Always call `cancel`**, usually with `defer cancel()` right after creating the context. It releases resources even if the work finished normally. `go vet` reports a missing cancel.

## Stopping a goroutine

```go run
package main

import (
	"context"
	"fmt"
)

func numbers(ctx context.Context) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 0; ; i++ {
			select {
			case ch <- i:
			case <-ctx.Done():
				return // stop when cancelled
			}
		}
	}()
	return ch
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := numbers(ctx)
	fmt.Println(<-ch, <-ch, <-ch)

	cancel() // tell the producer to stop
	for range ch {
		// drain until the producer closes the channel
	}
	fmt.Println("stopped:", ctx.Err())
}
```

The producer would otherwise run forever, blocked on a send nobody receives: a classic leak. Selecting on `ctx.Done()` alongside every blocking operation is the pattern to remember.

## Timeouts and deadlines

```go run
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func slowOperation(ctx context.Context) error {
	select {
	case <-time.After(2 * time.Second): // pretend work
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := slowOperation(ctx)
	fmt.Println(err)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
}
```

Use `errors.Is` to recognise `context.Canceled` and `context.DeadlineExceeded`, even when wrapped. The standard library respects contexts: `http.NewRequestWithContext`, `sql.DB.QueryContext`, `exec.CommandContext` and many more abort when the context ends.

## Propagation

Cancelling a parent cancels **all** of its children, but never the other way round. A child can have a shorter deadline than its parent, never a longer one.

```go run
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	parent, cancelParent := context.WithCancel(context.Background())
	child, cancelChild := context.WithTimeout(parent, time.Hour)
	defer cancelChild()

	cancelParent()
	<-child.Done() // closed because the parent was cancelled
	fmt.Println(child.Err())
}
```

## Rules for using context

1. **Pass it as the first parameter**, named `ctx`: `func Fetch(ctx context.Context, id int) (T, error)`.
2. **Do not store a context in a struct** (the rare exception is a request-like value object); pass it through calls.
3. **Never pass a nil context**; use `context.TODO()` if unsure.
4. Check `ctx.Done()` (or pass `ctx` on) in long loops and before blocking operations.
5. `WithValue` is for **request-scoped data that crosses API boundaries** (a request ID, an authenticated user), never for optional function parameters. Use an unexported custom key type to avoid collisions:

```go
type ctxKey struct{}

ctx = context.WithValue(ctx, ctxKey{}, "req-123")
id, _ := ctx.Value(ctxKey{}).(string)
```

## A few newer helpers

- `context.WithCancelCause(parent)` gives a `cancel(err)` function; `context.Cause(ctx)` returns that error, which explains *why* the context ended (`ctx.Err()` still says `Canceled`). Go 1.20.
- `context.AfterFunc(ctx, f)` runs `f` in its own goroutine once `ctx` is done, and returns a `stop` function. Go 1.21.
- `context.WithoutCancel(ctx)` keeps the values but detaches from cancellation, for example to finish cleanup work. Go 1.21.

## errgroup (read-only mention)

Running several goroutines, collecting the first error and cancelling the rest is so common that the extended library offers `golang.org/x/sync/errgroup`. It is not part of the standard library, so you add it with `go get golang.org/x/sync`:

```go
g, ctx := errgroup.WithContext(ctx)
for _, url := range urls {
	g.Go(func() error {
		return fetch(ctx, url) // ctx is cancelled when any call fails
	})
}
if err := g.Wait(); err != nil { // first non-nil error
	return err
}
```

`g.SetLimit(n)` caps how many goroutines run at once. It combines the WaitGroup, error collection and context cancellation you have just learned about.

## Key takeaways

- A context carries cancellation, a deadline and request-scoped values; `ctx.Done()` is a channel to `select` on.
- Roots are `Background()` and `TODO()`; derive children with `WithCancel`, `WithTimeout`, `WithDeadline`, `WithValue`.
- Always `defer cancel()`; cancelling a parent cancels every child.
- `ctx` is the first parameter, never stored in structs, never nil.
- `ctx.Err()` is `context.Canceled` or `context.DeadlineExceeded`; `context.Cause` gives the specific reason.
- `errgroup` (golang.org/x/sync) bundles goroutines, first-error and cancellation.
