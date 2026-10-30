---
title: Closures, defer and panic
minutes: 9
---
A **closure** is a function literal that refers to variables from the surrounding scope. The function "closes over" those variables and keeps them alive for as long as it exists.

## A counter generator

```go run
package main

import "fmt"

func newCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	a := newCounter()
	b := newCounter()

	fmt.Println(a(), a(), a())
	fmt.Println(b())
}
```

Each call to `newCounter` creates a fresh `count`. `a` and `b` have independent state, so the output is `1 2 3` then `1`. The variable `count` outlives `newCounter` because the closure still references it; the compiler moves it to the heap automatically.

Closures capture **variables, not values**. If the variable changes later, the closure sees the new value.

## Loop variables in Go 1.22+

Before Go 1.22, a `for` loop reused one variable for every iteration, so closures created in the loop all shared it, a famous bug. Since Go 1.22 each iteration gets its **own copy** of the loop variable.

```go run
package main

import "fmt"

func main() {
	var funcs []func() int
	for i := 0; i < 3; i++ {
		funcs = append(funcs, func() int { return i })
	}
	for _, f := range funcs {
		fmt.Print(f(), " ")
	}
	fmt.Println()
}
```

With Go 1.22 or newer (the module's `go` line must say 1.22 or higher) this prints `0 1 2`. In older versions it printed `3 3 3`.

> **Tip:** in old code you may see `i := i` at the top of the loop body. That line is no longer needed with Go 1.22+.

## defer in more detail

Deferred calls run when the function returns, in last-in first-out order. Arguments to the deferred call are **evaluated immediately**, when the `defer` statement executes.

```go run
package main

import "fmt"

func main() {
	x := 1
	defer fmt.Println("deferred sees", x) // x evaluated now: 1
	x = 2
	defer func() {
		fmt.Println("closure sees", x) // reads x when it runs: 2
	}()
	fmt.Println("body sees", x)
}
```

The output is `body sees 2`, `closure sees 2`, `deferred sees 1`.

## panic and recover, briefly

`panic` stops normal execution, runs deferred calls, and crashes the program with a message, unless something calls `recover()` inside a deferred function. Go reserves panics for programmer errors (index out of range, nil map write, nil pointer). For expected failures, return an `error` instead.

```go run
package main

import "fmt"

func safeIndex(s []int, i int) (v int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return s[i], nil
}

func main() {
	fmt.Println(safeIndex([]int{10, 20}, 1))
	fmt.Println(safeIndex([]int{10, 20}, 5))
}
```

This is also a nice use of a named result: the deferred closure sets `err` after the panic.

> **Common mistake:** using `panic` for ordinary errors like "file not found". Return an `error`; keep `panic` for bugs.

## Key takeaways

- A closure captures variables from its enclosing scope and keeps them alive.
- Each call to a closure factory gets its own captured state.
- Since Go 1.22, every loop iteration has its own loop variable.
- `defer` evaluates its arguments immediately but runs the call at function exit, LIFO.
- `recover` only works inside a deferred function; prefer returning errors over panicking.
