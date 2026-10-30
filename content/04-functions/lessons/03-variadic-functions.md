---
title: Variadic Functions
minutes: 6
---
A **variadic** function accepts any number of trailing arguments of one type. You have been using one all along: `fmt.Println(a ...any)`.

## Declaring one

Put `...` before the type of the last parameter. Inside the function, that parameter is an ordinary slice.

```go run
package main

import "fmt"

func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func main() {
	fmt.Println(sum())
	fmt.Println(sum(1))
	fmt.Println(sum(1, 2, 3, 4))
}
```

With zero arguments, `nums` is a `nil` slice, so the loop does nothing and `sum()` returns `0`.

## Mixing fixed and variadic parameters

Only the **last** parameter can be variadic, and it can follow regular ones.

```go run
package main

import (
	"fmt"
	"strings"
)

func join(sep string, parts ...string) string {
	return strings.Join(parts, sep)
}

func main() {
	fmt.Println(join("-", "a", "b", "c"))
	fmt.Printf("%q\n", join("/"))
}
```

The second call passes no parts, so the result is the empty string (printed as `""` by `%q`).

## Passing an existing slice with `...`

If you already have a slice, "spread" it into the call with `...` after the argument.

```go run
package main

import "fmt"

func count(items ...string) int {
	return len(items)
}

func main() {
	names := []string{"ann", "bob", "cy"}
	fmt.Println(count(names...))
	fmt.Println(count("only one"))
}
```

Writing `count(names)` would not compile: a `[]string` is not a `string`.

When you spread a slice, the function receives **that same slice**, not a copy. If the function modifies its elements, the caller sees it. When you list arguments individually, Go creates a fresh slice.

> **Common mistake:** forgetting the `...` when passing a slice, which gives `cannot use names (variable of type []string) as string value`.

## Rules to remember

| Rule | Example |
|---|---|
| Only the last parameter is variadic | `func f(a int, rest ...string)` |
| Inside the function it is `[]T` | `len(rest)`, `range rest` |
| Spread a slice with `...` | `f(1, words...)` |
| You cannot mix individual values and a spread | `f(1, "x", words...)` is an error |

## Key takeaways

- `func f(xs ...T)` accepts zero or more `T` values; `xs` has type `[]T`.
- Only the final parameter may be variadic.
- Use `f(slice...)` to pass an existing slice; it is shared, not copied.
- With no arguments the variadic parameter is `nil`.
