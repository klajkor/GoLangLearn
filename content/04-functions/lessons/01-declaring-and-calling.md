---
title: Declaring and Calling Functions
minutes: 7
---
Functions are the basic building block of every Go program. You have already met `main`; now let's write our own.

## The shape of a function

```go
func name(param1 type1, param2 type2) returnType {
    // body
}
```

The keyword `func` comes first, then the name, then the parameters in parentheses (name **before** type, as with variables), then the return type. If neighbouring parameters share a type you can write it once: `func add(a, b int) int`.

```go run
package main

import "fmt"

func add(a, b int) int {
	return a + b
}

func greet(name string) string {
	return "Hello, " + name + "!"
}

func logLine(msg string) {
	fmt.Println("[log]", msg)
}

func main() {
	fmt.Println(add(2, 3))
	fmt.Println(greet("Gopher"))
	logLine("done")
}
```

A function with no result simply omits the return type (`logLine` above). A function with a result **must** end in a `return` on every path, or the compiler complains with `missing return`.

## Arguments are passed by value

Go always copies arguments. When you pass an `int`, the function gets its own copy and cannot change your variable.

```go run
package main

import "fmt"

func tryToChange(n int) {
	n = 100
	fmt.Println("inside:", n)
}

func main() {
	x := 1
	tryToChange(x)
	fmt.Println("outside:", x)
}
```

The output shows `inside: 100` and `outside: 1`. (Later you will see that pointers, slices and maps let a function affect shared data, but the *argument itself* is still a copy.)

> **Common mistake:** expecting `n = 100` to modify the caller's variable. It only modifies the function's local copy.

## Naming and visibility

- Names starting with an **uppercase** letter are exported (visible to other packages), like `fmt.Println`.
- Names starting with lowercase are private to the package.
- Go has **no overloading**: you cannot have two functions with the same name and different parameters in one package.
- Functions can be declared in any order in a file; `main` can call a function defined below it.

## Unused parameters and the blank identifier

Unused *parameters* are allowed (unlike unused local variables). If you want to make it explicit, name the parameter `_`:

```go
func handler(_ int, name string) string {
	return name
}
```

## A quick look at defer

`defer` schedules a call to run when the surrounding function returns. It is perfect for cleanup. Deferred calls run in last-in, first-out order.

```go run
package main

import "fmt"

func main() {
	defer fmt.Println("deferred 1")
	defer fmt.Println("deferred 2")
	fmt.Println("working...")
}
```

The output is `working...`, `deferred 2`, `deferred 1`. Later you will use `defer` for closing files and unlocking mutexes.

## Key takeaways

- Syntax: `func name(a, b int) string { ... }`; types come after names.
- Arguments are always copied; assigning to a parameter does not affect the caller.
- Every code path of a function with a result must return.
- No overloading; uppercase names are exported.
- `defer` runs a call when the function exits, last-in first-out.
