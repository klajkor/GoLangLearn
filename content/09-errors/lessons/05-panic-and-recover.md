---
title: panic & recover
minutes: 8
---
Errors are for expected failures: bad input, missing files, network trouble. **Panics** are for situations the program cannot sensibly continue from, usually programming mistakes.

## What panic does

`panic(v)` stops the normal flow of the current function. Deferred functions run as the stack unwinds, and if nothing recovers, the program prints the panic message with a stack trace and exits with status 2.

The runtime itself panics on bugs such as:

- index out of range
- nil map write
- nil pointer dereference
- integer division by zero
- failed type assertion without comma-ok

```go
func main() {
	var m map[string]int
	m["a"] = 1 // panic: assignment to entry in nil map
}
```

```text
panic: assignment to entry in nil map

goroutine 1 [running]:
main.main()
	/tmp/main.go:5 +0x...
exit status 2
```

## defer, then recover

`recover()` stops a panic and returns the value that was passed to `panic`. It only works when called **directly inside a deferred function**. Called anywhere else it returns `nil`.

```go run
package main

import "fmt"

func risky() {
	defer fmt.Println("deferred 1 runs even during a panic")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r)
		}
	}()

	fmt.Println("about to panic")
	panic("something went wrong")
}

func main() {
	risky()
	fmt.Println("main continues normally")
}
```

Deferred calls run in last-in, first-out order: the recovering function runs first, then `deferred 1`. After recovery, `risky` returns normally to its caller.

## Converting a panic to an error

A common use of `recover` is a boundary that turns a panic into an ordinary `error`, so one bad input or one bug does not take down the whole program (an HTTP server does this per request). Use a **named result** so the deferred function can set it:

```go run
package main

import (
	"errors"
	"fmt"
)

func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return a / b, nil // panics when b == 0
}

func mustPositive(n int) int {
	if n <= 0 {
		panic(errors.New("n must be positive"))
	}
	return n
}

func try(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()
	f()
	return nil
}

func main() {
	fmt.Println(safeDivide(10, 2))
	fmt.Println(safeDivide(1, 0))

	fmt.Println(try(func() { mustPositive(5) }))
	fmt.Println(try(func() { mustPositive(-1) }))
	fmt.Println(try(func() { panic("plain string") }))
}
```

`recover()` returns `any`, so type-assert it to find out what was panicked. Runtime panics carry a `runtime.Error`, which implements `error`.

> **Common mistake:** calling `recover()` in a normal function body or in a helper called *by* the deferred function. It must be called directly in the deferred function itself, otherwise it returns `nil` and the panic continues.

## When is panic appropriate?

| Situation | Do |
|---|---|
| File missing, bad user input, network error | return an `error` |
| Impossible state, broken invariant (a true bug) | `panic` is acceptable |
| Program cannot start (invalid static config, bad regexp literal) | panic during init, like `regexp.MustCompile` |
| A library meeting a failure | return `error`; never panic across your public API |

The `MustXxx` naming convention (for example `regexp.MustCompile`, `template.Must`) signals "panics on failure; use only with constants known at write time".

Rules of thumb:

- **Do not** use panic/recover as exceptions for control flow.
- **Do not** let a panic cross your package boundary; recover internally and return an error.
- **Do** recover at the top of a goroutine or request handler if one failure must not kill the whole process. A panic in any goroutine, if not recovered in that goroutine, crashes the entire program.
- A recovered panic hides a bug; log it with a stack trace (`runtime/debug.Stack()`) so it can be fixed.

## Key takeaways

- Return `error` for expected failures; `panic` signals bugs and impossible states.
- Deferred functions run during a panic; `recover()` inside one stops it and returns the panic value.
- Use a named `err` result to turn a recovered panic into an error.
- `recover` works only when called directly by a deferred function, and only in the panicking goroutine.
- Never use panic as normal control flow, and never let one escape your public API.
