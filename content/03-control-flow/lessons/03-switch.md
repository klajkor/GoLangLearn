---
title: switch
minutes: 8
---
Go's `switch` is a cleaner version of the C-style one. It replaces long `if`/`else if` chains and is more flexible than you might expect.

## Basics: no fall-through

```go run
package main

import "fmt"

func describe(day string) string {
	switch day {
	case "sat", "sun": // several values in one case
		return "weekend"
	case "mon":
		return "ugh"
	default:
		return "weekday"
	}
}

func main() {
	fmt.Println(describe("sun"), describe("mon"), describe("wed"))
}
```

Key differences from C and Java:

- Each case ends automatically. You do **not** write `break`, and execution does **not** run into the next case.
- A case can list several values, separated by commas.
- Cases do not have to be constants; they can be any expressions, evaluated top to bottom until one matches.
- `default` runs when nothing matches. It can appear anywhere, though last is conventional.
- The value of a `switch` does not need to be an integer. Strings, structs, and so on work if they are comparable.

## Tagless switch

Leave out the value and each case becomes a boolean condition. This reads much better than a ladder of `else if`:

```go run
package main

import "fmt"

func grade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	default:
		return "F"
	}
}

func main() {
	for _, s := range []int{95, 85, 72, 10} {
		fmt.Print(grade(s), " ")
	}
	fmt.Println()
}
```

The first matching case wins, which is why the thresholds are ordered from highest to lowest.

## Init statement

Like `if`, a switch can start with a short statement:

```go
switch os := runtime.GOOS; os {
case "linux":
	fmt.Println("Linux")
case "windows":
	fmt.Println("Windows")
default:
	fmt.Println("something else")
}
```

## fallthrough

If you really do want C-style behaviour, the `fallthrough` keyword transfers control to the **body of the next case**, without testing its condition:

```go run
package main

import "fmt"

func main() {
	n := 2
	switch n {
	case 1:
		fmt.Println("one")
	case 2:
		fmt.Println("two")
		fallthrough
	case 3:
		fmt.Println("three (reached by fallthrough)")
	case 4:
		fmt.Println("four")
	}
}
```

It prints `two` and then the `three` line, but not `four`. `fallthrough` must be the last statement in a case, and it is rarely needed; if you find yourself using it, consider listing several values in one case instead.

> **Common mistake:** expecting `break` to be required, as in other languages. Here it is unnecessary. But a `break` inside a switch that sits in a `for` loop only leaves the switch, not the loop. Use a label (`break outer`) to leave the loop.

## Type switch (a preview)

A switch can also branch on the *dynamic type* of an interface value. The syntax uses `.(type)`:

```go
func kind(v any) string {
	switch x := v.(type) {
	case int:
		return fmt.Sprint("int ", x+1)
	case string:
		return "string of length " + fmt.Sprint(len(x))
	case nil:
		return "nil"
	default:
		return "something else"
	}
}
```

Inside each case, `x` has the matching concrete type (`int` in the first case, `string` in the second). You will use this a lot once you learn interfaces and `any`; for now just recognise the shape.

## switch or if?

Use `switch` when you compare one value against several alternatives, or when a tagless switch makes a ladder of conditions clearer. Use `if` for a single yes/no decision.

## Key takeaways

- Cases do not fall through; no `break` is needed.
- A case can list multiple values, and the first match wins.
- A tagless `switch { case cond: ... }` replaces `if / else if` chains.
- `fallthrough` explicitly continues into the next case body and is rarely needed.
- `switch v := x.(type)` branches on an interface's dynamic type.
