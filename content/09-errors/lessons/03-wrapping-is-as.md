---
title: Wrapping, errors.Is & errors.As
minutes: 9
---
As an error travels up the call stack, each layer should add context ("loading config", "opening file"). But the original error must stay identifiable. **Wrapping** does both.

## Wrapping with %w

`fmt.Errorf` with the `%w` verb creates a new error that wraps the original:

```go run
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func find(id int) error {
	return ErrNotFound
}

func loadUser(id int) error {
	if err := find(id); err != nil {
		return fmt.Errorf("loadUser %d: %w", id, err)
	}
	return nil
}

func handler() error {
	if err := loadUser(7); err != nil {
		return fmt.Errorf("handler: %w", err)
	}
	return nil
}

func main() {
	err := handler()
	fmt.Println(err)
	fmt.Println(err == ErrNotFound)
	fmt.Println(errors.Is(err, ErrNotFound))
	fmt.Println(errors.Unwrap(err))
}
```

The message reads like a trail: `handler: loadUser 7: not found`. The `==` check fails because `err` is now a different value, but `errors.Is` walks the chain and finds the sentinel.

`%v` also prints the message, but it does **not** wrap: the original is lost for `Is` and `As`. Use `%w` when callers might need to inspect the cause, and `%v` when you intentionally want to hide it (to avoid making an internal error part of your API).

## errors.Is

`errors.Is(err, target)` reports whether any error in the chain equals `target`. It unwraps repeatedly. It also calls an `Is(error) bool` method if an error provides one.

## errors.As

`errors.As(err, &target)` finds the first error in the chain that is assignable to the type of `target`, and sets `target` to it. Pass a **pointer** to a variable of the wanted type.

```go run
package main

import (
	"errors"
	"fmt"
	"strconv"
)

type StatusError struct {
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("status %d", e.Code)
}

func call() error {
	return fmt.Errorf("call api: %w", &StatusError{Code: 503})
}

func main() {
	err := call()

	var se *StatusError
	if errors.As(err, &se) {
		fmt.Println("got status:", se.Code)
	}

	// The standard library uses the same pattern.
	_, err = strconv.Atoi("12x")
	var ne *strconv.NumError
	if errors.As(err, &ne) {
		fmt.Println("func:", ne.Func, "input:", ne.Num)
	}
	fmt.Println(errors.Is(err, strconv.ErrSyntax))
}
```

> **Common mistake:** passing a non-pointer or the wrong pointer type to `errors.As`. The second argument must be a non-nil pointer to a type that implements `error` (or to any interface type). `go vet` reports this mistake.

## Unwrap and custom wrapping

`errors.Unwrap(err)` returns the result of calling `err.Unwrap()`, or `nil`. You rarely call it directly; `Is` and `As` do it for you.

A custom error type can take part in wrapping by implementing `Unwrap() error`:

```go run
package main

import (
	"errors"
	"fmt"
	"io"
)

type QueryError struct {
	Query string
	Err   error
}

func (e *QueryError) Error() string { return e.Query + ": " + e.Err.Error() }
func (e *QueryError) Unwrap() error { return e.Err }

func main() {
	err := &QueryError{Query: "SELECT 1", Err: io.EOF}
	fmt.Println(err)
	fmt.Println(errors.Is(err, io.EOF))
}
```

## Which tool when?

| Question | Tool |
|---|---|
| Is this (or something it wraps) the specific sentinel? | `errors.Is(err, ErrX)` |
| Is it of this type, and I need its fields? | `errors.As(err, &target)` |
| I need to add context and keep the cause | `fmt.Errorf("...: %w", err)` |
| I need to add context but hide the cause | `fmt.Errorf("...: %v", err)` |

## Where to wrap

- Wrap when crossing a meaningful boundary and you can add useful context: which operation, which input.
- Do not wrap at every single line; the message becomes noisy.
- Do not log **and** return the same error. Handle it once: either log it (at the top of the program) or return it.

> **Tip:** `fmt.Errorf` accepts multiple `%w` verbs (Go 1.20+), producing an error that wraps all of them. `errors.Is` and `As` check each.

## Key takeaways

- Wrap with `fmt.Errorf("context: %w", err)` to add context and preserve the cause.
- Use `errors.Is` for sentinels and `errors.As` for types; they see through wrapping, while `==` does not.
- `%v` formats without wrapping.
- Implement `Unwrap() error` on custom errors that wrap another.
- Handle each error once: log or return, not both.
