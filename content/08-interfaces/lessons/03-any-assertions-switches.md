---
title: any, Type Assertions & Type Switches
minutes: 9
---
Sometimes you hold an interface value and need the concrete type back. Go gives you the empty interface, type assertions and type switches.

## The empty interface

An interface with no methods is satisfied by **every** type. It is written `interface{}`, or since Go 1.18 as the alias `any`.

```go
var x any
x = 42
x = "hello"
x = []int{1, 2, 3}
```

`any` is useful for things that truly accept anything (`fmt.Println(a ...any)`, JSON decoding into unknown shapes), but it throws away type safety. You cannot do much with an `any` value directly; you must recover its type first.

> **Tip:** treat `any` as a last resort. Prefer a specific interface or, for functions that work on many types, generics (a later chapter).

## Type assertions

`x.(T)` asserts that interface value `x` holds a `T` and returns it.

```go run
package main

import "fmt"

func main() {
	var x any = "gopher"

	s := x.(string) // ok
	fmt.Println(s, len(s))

	// Comma-ok form never panics.
	n, ok := x.(int)
	fmt.Println(n, ok)

	if str, ok := x.(string); ok {
		fmt.Println("it is a string:", str)
	}
}
```

Without the comma-ok form, a failed assertion **panics**:

```go
var x any = "gopher"
n := x.(int) // panic: interface conversion: interface {} is string, not int
```

Use the single-value form only when you are certain, and the `v, ok` form when you are not. On failure `v` is the zero value of `T` and `ok` is `false`.

You can also assert to another **interface** type, which checks whether the dynamic value has those methods:

```go run
package main

import (
	"fmt"
	"strings"
)

type Closer interface{ Close() error }

type file struct{ name string }

func (f file) Close() error {
	fmt.Println("closing", f.name)
	return nil
}

func main() {
	values := []any{file{"a.txt"}, strings.NewReader("x"), 7}
	for _, v := range values {
		if c, ok := v.(Closer); ok {
			_ = c.Close()
		} else {
			fmt.Printf("%T cannot be closed\n", v)
		}
	}
}
```

This "optional interface" technique is used widely, for example `http.Flusher` and `io.WriterTo`.

## Type switches

When there are several possible types, a type switch is cleaner than a chain of assertions.

```go run
package main

import "fmt"

func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int %d (doubled %d)", x, x*2)
	case string:
		return "string of length " + fmt.Sprint(len(x))
	case []int:
		return fmt.Sprintf("slice with %d ints", len(x))
	case error:
		return "error: " + x.Error()
	case fmt.Stringer:
		return "stringer: " + x.String()
	default:
		return fmt.Sprintf("other type %T", x)
	}
}

func main() {
	fmt.Println(describe(21))
	fmt.Println(describe("go"))
	fmt.Println(describe([]int{1, 2}))
	fmt.Println(describe(fmt.Errorf("boom")))
	fmt.Println(describe(3.14))
	fmt.Println(describe(nil))
}
```

Inside each `case`, `x` has the case's type (in a case listing several types, or in `default`, it keeps the interface type). Cases are checked top to bottom, so put specific types before general interfaces. `case nil` matches a nil interface value.

## Comparison of the tools

| Tool | Use when | On mismatch |
|---|---|---|
| `x.(T)` | you are certain of the type | panics |
| `v, ok := x.(T)` | you want to test for one type | `ok == false` |
| `switch x.(type)` | several possible types | falls to `default` |

> **Common mistake:** asserting on a non-interface value. `x.(T)` is only legal when `x` itself has an interface type. For `n := 5`, `n.(int)` is a compile error.

## Key takeaways

- `any` is an alias for `interface{}`; every type satisfies it, but it loses static type checking.
- `v, ok := x.(T)` is the safe assertion; `x.(T)` alone panics on mismatch.
- Assert to an interface type to test for optional capabilities.
- Use a type switch for several types, with the specific cases first.
