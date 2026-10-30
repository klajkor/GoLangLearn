---
title: defer
minutes: 8
---
`defer` schedules a function call to run **when the surrounding function returns**, whether it returns normally, by a `return` statement, or because of a panic. It is Go's tool for cleanup, keeping the cleanup right next to the code that needs it.

```go run
package main

import "fmt"

func main() {
	fmt.Println("start")
	defer fmt.Println("deferred: runs last")
	fmt.Println("middle")
}
```

The deferred line prints after `middle`, just before `main` returns.

## Order: last in, first out

Multiple defers form a stack. The most recently deferred call runs first:

```go run
package main

import "fmt"

func main() {
	for i := 1; i <= 3; i++ {
		defer fmt.Println("deferred", i)
	}
	fmt.Println("loop finished")
}
```

The output is `loop finished`, then `deferred 3`, `deferred 2`, `deferred 1`. This reverse order makes sense for cleanup: if you open A and then B, you want to close B before A.

## Arguments are evaluated immediately

This surprises almost everyone once. When the `defer` statement executes, the function value and its **arguments** are evaluated right then. Only the call itself is postponed.

```go run
package main

import "fmt"

func main() {
	x := 1
	defer fmt.Println("argument captured at defer time:", x)

	defer func() {
		fmt.Println("closure sees the final value:", x)
	}()

	x = 100
	fmt.Println("x is now", x)
}
```

The first deferred call printed `1`, because `x` was copied when `defer` ran. The closure, in contrast, refers to the variable `x` itself, so it sees `100`. Both are useful; just know which one you have written.

## Typical uses

Pair acquisition with release on the next line, so you cannot forget:

```go
f, err := os.Open("data.txt")
if err != nil {
	return err
}
defer f.Close() // runs on every return path below

// ... read from f ...
```

```go
mu.Lock()
defer mu.Unlock() // always unlocked, even on early return or panic
```

Other common cases are closing network connections and database rows, finishing a timer, and recovering from panics (a later chapter).

> **Common mistake:** `defer` inside a long loop. Deferred calls run when the *function* ends, not at the end of each iteration, so a loop opening a thousand files would hold all of them open until the function returns. Move the loop body into its own function, or close explicitly.

## Defer and return values

Deferred closures run after the return value has been set, and they can modify **named result values**:

```go run
package main

import "fmt"

func double(n int) (result int) {
	defer func() {
		result *= 2 // runs after "return n + 1" has set result
	}()
	return n + 1
}

func main() {
	fmt.Println(double(4))
}
```

`return n + 1` stores `5` in `result`, then the deferred function doubles it, so the caller receives `10`. This is how libraries turn a recovered panic into an ordinary error. For everyday code, remember only that defers run after the return expression is evaluated.

## Performance

A defer costs a few nanoseconds in modern Go, which is irrelevant in nearly all code. Prefer correctness: if there is cleanup to do, `defer` it.

> **Tip:** `defer` takes a function *call*, so `defer f.Close` (without parentheses) does not compile. To run several statements, wrap them in a closure: `defer func() { ... }()`.

## Key takeaways

- `defer` runs a call when the enclosing function returns, even on panic.
- Deferred calls execute in last-in, first-out order.
- Arguments are evaluated when `defer` executes; a closure sees later changes to variables.
- Use it right after acquiring a resource: files, locks, connections.
- Beware of `defer` in loops: it waits for the function, not the iteration.
- Deferred closures can modify named return values.
