---
title: "Limitations and When Not to Use Generics"
minutes: 9
---
Generics are a tool, not a style. Go's designers added them carefully and kept them small. Knowing the limits helps you decide when a plain interface or a plain function is the better answer.

## Limitations to know

| Limitation | Detail |
|---|---|
| No type parameters on methods | Only the receiver's parameters are available; use a generic function instead. |
| No operators beyond the constraint | You cannot use `+` on `any`, `<` on `comparable`, `%` on a set that includes floats. |
| No specialization | You cannot write one version for `int` and another for `string` of the same generic function. |
| No type switch on `T` itself | A type switch needs an interface value: write `switch any(v).(type)`. |
| No field access | A constraint cannot say "has a field `Name`"; use a method such as `Name() string`. |

A type switch on a type parameter looks like this. It works, but treat it as a smell: if you branch on the concrete type, you probably wanted separate functions or an interface.

```go run
package main

import "fmt"

func Describe[T any](v T) string {
	switch x := any(v).(type) {
	case int:
		return fmt.Sprintf("int %d", x)
	case string:
		return "string " + x
	default:
		return fmt.Sprintf("other %T", x)
	}
}

func main() {
	fmt.Println(Describe(7))
	fmt.Println(Describe("go"))
	fmt.Println(Describe(2.5))
}
```

## Generics versus interfaces

Both abstract over types, but they answer different questions.

- **Interfaces abstract behaviour**: "anything that can `Write`". Different types do different things behind the same method.
- **Generics abstract algorithms and data structures**: "the same steps for any element type". The code does not care what the type does, only that it can be stored, compared or ordered.

```go run
package main

import (
	"fmt"
	"io"
	"strings"
)

// Behaviour: an interface is the right tool. The function only needs Read.
func countBytes(r io.Reader) int {
	data, _ := io.ReadAll(r)
	return len(data)
}

// Generic is overkill here, even though it compiles:
func countBytesGeneric[R io.Reader](r R) int {
	data, _ := io.ReadAll(r)
	return len(data)
}

// Algorithm over elements: generics are the right tool.
func Uniq[T comparable](xs []T) []T {
	seen := make(map[T]bool)
	var out []T
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

func main() {
	fmt.Println(countBytes(strings.NewReader("hello")))
	fmt.Println(countBytesGeneric(strings.NewReader("hello")))
	fmt.Println(Uniq([]string{"a", "b", "a", "c", "b"}))
	fmt.Println(Uniq([]int{3, 3, 3}))
}
```

`countBytesGeneric` adds syntax and gains nothing. `Uniq`, on the other hand, would need `any` and type assertions without generics.

## When generics are a good fit

- **Containers**: stacks, queues, sets, trees, caches, ring buffers.
- **Algorithms over slices and maps**: filter, group, min/max, binary search, de-duplication.
- **Helpers around channels, pointers and functions** that must preserve the element type: `func Ptr[T any](v T) *T`, a generic `Merge[T any](chans ...<-chan T) <-chan T`.
- **Avoiding duplicated code** that differs only in type.

## When to avoid them

- You only call methods on the argument: take an interface.
- There is one concrete type today. Write it concretely; generalize when a second caller appears ("duplicate before abstracting").
- The generic version needs `any(v).(type)` switches or reflection.
- The signature becomes hard to read (`[K comparable, V any, M ~map[K]V, S ~[]V]` and so on). Clever type-level code costs your teammates more than a little duplication.
- Performance tuning: generic code is compiled using "GC shape stenciling", where types with the same memory layout share code and method calls on type-parameter values can go through a dictionary. It is usually fast enough but not always as fast as hand-written concrete code; measure before assuming either way.

> **Common mistake:** replacing `io.Reader`-style interfaces with `[R io.Reader]` "because generics are modern". It makes the API noisier without adding type safety.

> **Tip:** a good rule of thumb from the Go team: write code, and when you find yourself writing the same function twice for different types, or reaching for `any` plus type assertions, reach for generics.

## Putting it together

A last example combines a constraint, a generic type and the zero-value idiom into a small, genuinely useful type: a bounded cache keeping the most recent entries.

```go run
package main

import "fmt"

type Cache[K comparable, V any] struct {
	limit int
	order []K
	data  map[K]V
}

func NewCache[K comparable, V any](limit int) *Cache[K, V] {
	return &Cache[K, V]{limit: limit, data: make(map[K]V)}
}

func (c *Cache[K, V]) Put(k K, v V) {
	if _, ok := c.data[k]; !ok {
		if len(c.order) == c.limit {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.data, oldest)
		}
		c.order = append(c.order, k)
	}
	c.data[k] = v
}

func (c *Cache[K, V]) Get(k K) (V, bool) {
	v, ok := c.data[k]
	return v, ok
}

func main() {
	c := NewCache[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3) // evicts "a"

	_, okA := c.Get("a")
	vc, okC := c.Get("c")
	fmt.Println(okA, vc, okC)
}
```

## Key takeaways

- Methods cannot have type parameters, you cannot specialize, and a constraint limits you to operations common to its whole type set.
- Use interfaces for **behaviour**, generics for **containers and algorithms** that are identical for every element type.
- Needing `any(v).(type)` inside a generic function is a warning sign.
- Do not generalize early; start concrete and add type parameters when duplication appears.
- Prefer the standard `slices`, `maps` and `cmp` helpers before writing your own.
