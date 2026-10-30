---
title: if and else
minutes: 6
---
Go's `if` looks familiar, with three differences: no parentheses around the condition, braces are mandatory, and you can run a short statement before the condition.

```go
if x > 10 {
	fmt.Println("big")
} else if x > 5 {
	fmt.Println("medium")
} else {
	fmt.Println("small")
}
```

- The condition must be a `bool`. `if x {` with an integer or string does not compile.
- Braces are required even for a one-line body, which removes a whole class of bugs.
- `else` must be on the same line as the closing `}` of the previous block; otherwise the automatic semicolon insertion breaks the syntax.

## The init statement

An `if` can begin with a simple statement, followed by a semicolon. Variables declared there exist only inside the `if`, including all of its `else if` and `else` branches.

```go run
package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, text := range []string{"42", "forty-two"} {
		if n, err := strconv.Atoi(text); err != nil {
			fmt.Println("not a number:", text)
		} else if n > 40 {
			fmt.Println("big number:", n)
		} else {
			fmt.Println("small number:", n)
		}
	}
}
```

This is the idiomatic way to call a function that returns `(value, error)`: the pair lives exactly as long as it is needed, and does not leak into the surrounding function. `n` is even visible in the `else if` branch.

## The "early return" style

Go code usually handles the unhappy path first and returns, leaving the main logic un-indented. Instead of deeply nested `if`/`else`, you will see this shape:

```go
func parsePort(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n < 1 || n > 65535 {
		return 0, fmt.Errorf("port %d out of range", n)
	}
	return n, nil
}
```

Notice there is no `else` after a `return`. Linters will even suggest removing it. (Functions and errors get full lessons later; just notice the shape.)

> **Tip:** When a variable is only needed for the check, put it in the init statement. When you need it after the `if`, declare it before.

## Logical operators and short-circuiting

| Operator | Meaning |
|---|---|
| `&&` | and (stops at the first false) |
| `\|\|` | or (stops at the first true) |
| `!` | not |

Short-circuiting lets you guard an expression with a check on its left. The right side is not evaluated if the left already decides the result:

```go run
package main

import "fmt"

func check(name string, result bool) bool {
	fmt.Println("checking", name)
	return result
}

func main() {
	if check("a", false) && check("b", true) {
		fmt.Println("both")
	}
	if check("c", true) || check("d", true) {
		fmt.Println("at least one")
	}

	items := []int{}
	if len(items) > 0 && items[0] == 1 {
		fmt.Println("first is one")
	} else {
		fmt.Println("nothing to test")
	}
}
```

`check("b", ...)` and `check("d", ...)` never run. The last `if` is safe because `items[0]` is evaluated only when the slice is non-empty.

## No ternary operator

Go has no `cond ? a : b`. Use a plain `if`:

```go
label := "off"
if enabled {
	label = "on"
}
```

It is a few lines longer, but nobody ever misreads it.

> **Common mistake:** writing `if x = 5 {`. Assignment is a statement, not an expression, so this does not compile. Comparison uses `==`. This removes the classic C bug of assigning inside a condition.

## Key takeaways

- No parentheses around conditions; braces are always required.
- The condition must be a `bool`; there is no truthiness.
- `if v, err := f(); err != nil { ... }` scopes `v` and `err` to the `if` and its `else` branches.
- Prefer early returns over nested `else` blocks.
- `&&` and `||` short-circuit; there is no ternary operator.
