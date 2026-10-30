---
title: Constraints and Type Sets
minutes: 10
---
A **constraint** is an interface that describes which types may be used as a type argument. With generics, interfaces gained a second role: besides listing methods they can list **types** (a *type set*).

## Built-in and standard constraints

| Constraint | Allows | Lets you use |
|---|---|---|
| `any` | every type (alias for `interface{}`) | assignment and passing around; nothing else |
| `comparable` | types usable with `==` and `!=` (and as map keys) | `==`, `!=` |
| `cmp.Ordered` | integers, floats, strings (and types built on them) | `<`, `<=`, `>`, `>=`, `==` |

`cmp.Ordered` lives in the standard `cmp` package (Go 1.21+). Before that people used `golang.org/x/exp/constraints`.

```go run
package main

import (
	"cmp"
	"fmt"
)

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Index[T comparable](xs []T, target T) int {
	for i, x := range xs {
		if x == target {
			return i
		}
	}
	return -1
}

func main() {
	fmt.Println(Max(3, 7), Max("apple", "pear"), Max(2.5, 1.5))
	fmt.Println(Index([]string{"a", "b", "c"}, "c"))
	fmt.Println(Index([]int{10, 20}, 99))
}
```

Without `cmp.Ordered`, `a > b` would not compile: with `any` the compiler cannot know that `>` makes sense for `T`.

## Custom constraints with type sets

Write an interface whose body lists types joined with `|`.

```go
type Number interface {
	int | int64 | float64
}
```

Such an interface can only be used as a constraint, not as the type of an ordinary variable.

### The tilde `~`

`int` in a type set means exactly `int`. A named type like `type Celsius float64` has underlying type `float64` but is a different type, so it would be rejected. `~float64` means "any type whose **underlying type** is `float64`".

```go run
package main

import "fmt"

type Number interface {
	~int | ~int64 | ~float64
}

type Celsius float64
type Score int

func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

func main() {
	fmt.Println(Sum([]Celsius{20.5, 1.5}))
	fmt.Println(Sum([]Score{10, 20, 30}))
	fmt.Printf("%T\n", Sum([]Score{1}))
}
```

Without the `~`, `Sum([]Celsius{...})` would fail to compile because `Celsius` is not literally `float64`. The result keeps the named type: `main.Score`.

Constraints can be composed: embed another constraint such as `cmp.Ordered`, or add methods next to the type list.

## Constraints with methods

A constraint can require methods, exactly like a normal interface. The body of the generic function may then call those methods on values of type `T`.

```go run
package main

import (
	"fmt"
	"strings"
)

type Named interface {
	Name() string
}

type Dog struct{}
type Cat struct{}

func (Dog) Name() string { return "dog" }
func (Cat) Name() string { return "cat" }

func Names[T Named](xs []T) string {
	parts := make([]string, 0, len(xs))
	for _, x := range xs {
		parts = append(parts, x.Name())
	}
	return strings.Join(parts, ",")
}

// A method and a type set together: only integer-based types with a String method.
type StringableInt interface {
	~int | ~int64
	String() string
}

type Level int

func (l Level) String() string { return fmt.Sprintf("L%d", int(l)) }

func Describe[T StringableInt](v T) string {
	return v.String() + "/" + fmt.Sprint(int64(v)+1)
}

func main() {
	fmt.Println(Names([]Dog{{}, {}}))
	fmt.Println(Describe(Level(4)))
}
```

Notice that `Names` takes `[]T`, not `[]Named`. A `[]Dog` cannot be passed where a `[]Named` is expected, but it is a valid `[]T` with `T = Dog`. That is one real benefit of generics here: no copying into an interface slice.

## What the constraint allows

Inside the function you may only do what **every** type in the constraint supports. With `~int | ~float64` you may use `+`, `-`, `*`, `/` and comparisons. You cannot use `%` (invalid for floats), and you cannot index, call or range over `T` unless the whole set supports it.

> **Common mistake:** using `any` and then writing `a + b` or `a < b`. Pick the narrowest constraint that expresses what the body needs.

> **Tip:** since Go 1.20 interface types satisfy `comparable`, but comparing two interface values panics at run time if their dynamic type is not comparable (a slice, for example). The compiler cannot catch that.

## Key takeaways

- A constraint is an interface; it may list methods, types, or both.
- `any`, `comparable` and `cmp.Ordered` cover most needs.
- A union like `int | float64` lists exact types; `~T` includes every type whose underlying type is `T`.
- Interfaces with type lists are usable only as constraints.
- A generic function body may only use operations valid for all types in the constraint.
