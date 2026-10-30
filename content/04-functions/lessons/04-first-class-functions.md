---
title: Functions as Values
minutes: 7
---
In Go, functions are **first-class values**. You can store them in variables, pass them as arguments, return them from other functions and put them in maps and slices.

## Function types

A function's type is described by its parameter and result types.

```go run
package main

import "fmt"

func double(n int) int { return n * 2 }
func square(n int) int { return n * n }

func apply(nums []int, f func(int) int) []int {
	out := make([]int, 0, len(nums))
	for _, n := range nums {
		out = append(out, f(n))
	}
	return out
}

func main() {
	var f func(int) int = double
	fmt.Println(f(21))

	data := []int{1, 2, 3}
	fmt.Println(apply(data, double))
	fmt.Println(apply(data, square))
}
```

`apply` is a *higher-order function*: it takes behaviour as an argument. The zero value of a function variable is `nil`; calling a nil function panics.

## Anonymous functions

You can write a function without a name, called a function literal, and call it right away or assign it.

```go run
package main

import "fmt"

func main() {
	result := func(a, b int) int {
		return a * b
	}(6, 7) // called immediately
	fmt.Println(result)

	ops := map[string]func(int, int) int{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}
	fmt.Println(ops["add"](5, 3), ops["sub"](5, 3))
}
```

A map of functions is a neat replacement for a long `switch` when you dispatch on a name.

## A named function type

Giving a function signature a name makes code easier to read and lets you attach methods later.

```go
type Transform func(string) string
```

## Using functions to customise behaviour

The standard library does this all the time. For example, `sort.Slice` takes a "less" function that decides the order:

```go run
package main

import (
	"fmt"
	"sort"
)

func main() {
	words := []string{"banana", "fig", "apple", "kiwi"}
	sort.Slice(words, func(i, j int) bool {
		if len(words[i]) != len(words[j]) {
			return len(words[i]) < len(words[j])
		}
		return words[i] < words[j]
	})
	fmt.Println(words)
}
```

Sorting by length, then alphabetically, gives `[fig kiwi apple banana]`.

> **Tip:** if a function literal grows beyond ten lines or so, give it a name. Readers will thank you.

## Key takeaways

- Functions have types such as `func(int, int) int` and can be assigned, passed and returned.
- Function literals (anonymous functions) can be called immediately or stored.
- A nil function value panics when called.
- Maps of functions work well for dispatch tables.
- `type Name func(...)` gives a signature a readable name.
