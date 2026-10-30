---
title: Declaring Variables
minutes: 8
---
Go is statically typed: every variable has a type fixed at compile time. But you rarely have to write that type out, because Go can infer it. There are two main ways to declare a variable.

## var

`var` works anywhere, inside or outside functions. You give a name, and a type, a value, or both:

```go
var count int          // type only: starts at the zero value (0)
var name string = "Go" // type and value
var ratio = 0.5        // value only: type inferred (float64)
```

You can declare several at once, and group related declarations in a block:

```go
var x, y int = 1, 2

var (
	host = "localhost"
	port = 8080
)
```

## Short declaration :=

Inside functions you can use the short form. It declares and initialises in one step, and infers the type from the right-hand side:

```go
total := 10      // int
price := 9.99    // float64
label := "sale"  // string
a, b := 1, "two" // two variables, two different types
```

`:=` is not available at package level; there, you must use `var`.

> **Common mistake:** `:=` declares *new* variables, while `=` assigns to existing ones. `x := 1` followed by `x := 2` in the same scope is a compile error ("no new variables on left side"). However, `a, err := f()` followed by `b, err := g()` is fine, because at least one variable on the left (`b`) is new; `err` is simply reused.

## Unused variables are errors

Like imports, a local variable that is declared but never used makes the program fail to compile (`declared and not used`). If you truly want to ignore a value, assign it to the **blank identifier** `_`:

```go
_, err := fmt.Println("hi") // ignore the byte count
```

## Scope and shadowing

A variable lives in the innermost block `{ ... }` where it was declared. An inner block can declare a new variable with the same name as an outer one. That *shadows* the outer variable for the duration of the inner block, and the outer variable is untouched.

```go run
package main

import "fmt"

func main() {
	x := 1
	fmt.Println("outer before:", x)

	{
		x := 2 // a brand-new variable that hides the outer x
		x++
		fmt.Println("inner:", x)
	}

	fmt.Println("outer after:", x)

	if x := x * 10; x > 5 {
		fmt.Println("if-scoped x:", x)
	}
	fmt.Println("still outer:", x)
}
```

The inner `x` was incremented to 3, yet the outer `x` stayed at 1. The `if x := ...` line also created a new, scoped `x`. Shadowing is legal and sometimes useful, but it is the source of a classic bug: you think you are updating a variable, but `:=` quietly created a new one.

```go
var result string
if ok {
	result := "yes" // BUG: shadows result; the outer one stays ""
	_ = result
}
```

> **Tip:** `go vet` has an optional shadow analyzer, and many editors highlight shadowed variables. Prefer `=` when you mean to assign.

## Package-level variables

Variables declared outside any function are visible to the whole package. They can be declared in any order, because Go resolves their dependencies for you, but use them sparingly: global mutable state makes code harder to test.

```go run
package main

import "fmt"

var greeting = "Hello"

func main() {
	var name = "Gopher"
	count, unit := 3, "times"
	fmt.Println(greeting, name, count, unit)

	count = 4 // plain assignment
	name, extra := "Gophers", true
	fmt.Println(greeting, name, count, extra)
}
```

Here `name, extra := ...` reuses `name` (same scope) and creates `extra`, since at least one new variable appears on the left.

## Swapping and multiple assignment

Go evaluates the whole right side before assigning, so swapping needs no temporary:

```go
a, b := 1, 2
a, b = b, a // a == 2, b == 1
```

## Key takeaways

- `var name type = value` works everywhere; either the type or the value may be omitted, not both.
- `name := value` declares and infers the type, but only inside functions.
- `:=` needs at least one new variable on its left; `=` assigns to existing ones.
- Unused local variables and imports are compile errors; use `_` to discard a value.
- Inner blocks can shadow outer variables; the outer variable is not modified.
