---
title: Generic Types
minutes: 10
---
Functions are not the only things that can take type parameters. Structs, slices, maps and function types can too. This is where generics shine: **containers** such as stacks, queues, sets and linked lists that were painful before.

## A generic stack

```go run
package main

import "fmt"

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items = s.items[:last]
	return v, true
}

func (s *Stack[T]) Len() int { return len(s.items) }

func main() {
	var ints Stack[int]
	ints.Push(1)
	ints.Push(2)
	v, _ := ints.Pop()
	fmt.Println(v, ints.Len())

	words := &Stack[string]{}
	words.Push("go")
	w, ok := words.Pop()
	fmt.Println(w, ok)
	_, ok = words.Pop()
	fmt.Println(ok)
}
```

Points to notice:

- The type is declared as `Stack[T any]`; you must **instantiate** it with a type argument when you use it: `Stack[int]`. The compiler does not infer type arguments for types in a variable declaration.
- Methods repeat the receiver's type parameters: `func (s *Stack[T]) Push(v T)`. The name `T` is local to the method; you could call it `E`, but consistent names read better.
- `Stack[int]` and `Stack[string]` are two different, unrelated types.

## Several parameters: `Pair[K, V]`

```go run
package main

import "fmt"

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

func (p Pair[K, V]) String() string {
	return fmt.Sprintf("%v=%v", p.Key, p.Value)
}

// A constructor lets the compiler infer the type arguments.
func NewPair[K comparable, V any](k K, v V) Pair[K, V] {
	return Pair[K, V]{Key: k, Value: v}
}

func main() {
	p := NewPair("age", 42)
	fmt.Println(p)
	fmt.Printf("%T\n", p)

	explicit := Pair[int, bool]{Key: 1, Value: true}
	fmt.Println(explicit)
}
```

Because `NewPair` is a function, `NewPair("age", 42)` infers `Pair[string, int]`. That is why generic types usually come with `NewX` constructors.

## A generic linked list

A generic type may refer to itself, using the same type arguments.

```go run
package main

import "fmt"

type node[T any] struct {
	val  T
	next *node[T]
}

type List[T any] struct {
	head *node[T]
	size int
}

func (l *List[T]) Prepend(v T) {
	l.head = &node[T]{val: v, next: l.head}
	l.size++
}

// All returns the values in order, front to back.
func (l *List[T]) All() []T {
	out := make([]T, 0, l.size)
	for n := l.head; n != nil; n = n.next {
		out = append(out, n.val)
	}
	return out
}

func main() {
	var l List[string]
	l.Prepend("c")
	l.Prepend("b")
	l.Prepend("a")
	fmt.Println(l.All(), l.size)
}
```

## Methods cannot have their own type parameters

A method can use the type parameters of its receiver type, but it **cannot declare new ones**.

```go
// Does not compile: "methods cannot have type parameters"
func (s *Stack[T]) Map[U any](f func(T) U) *Stack[U] { ... }
```

The workaround is an ordinary generic **function** that takes the container as an argument:

```go run
package main

import "fmt"

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func MapStack[T, U any](s *Stack[T], f func(T) U) *Stack[U] {
	out := &Stack[U]{}
	for _, v := range s.items {
		out.Push(f(v))
	}
	return out
}

func main() {
	s := &Stack[int]{}
	s.Push(1)
	s.Push(2)
	strs := MapStack(s, func(n int) string { return fmt.Sprint("#", n) })
	fmt.Println(strs.items)
}
```

## Constraints on generic types

The constraint of a type parameter applies wherever the type is used. A `Set` needs comparable elements because it is backed by a map:

```go
type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(v T)           { s[v] = struct{}{} }
func (s Set[T]) Has(v T) bool      { _, ok := s[v]; return ok }
```

`Set[[]int]` would not compile, because slices are not comparable.

> **Common mistake:** forgetting that the zero value of a generic container is usually usable. Design `Stack[T]` so `var s Stack[int]` works without a constructor, as above (a nil slice is fine to `append` to).

> **Tip:** type parameters on a struct are part of its identity. If you only need one operation to vary, a function field or an interface is often simpler than a type parameter.

## Key takeaways

- Types take type parameters too: `type Stack[T any] struct{ ... }`; instantiate with `Stack[int]`.
- Methods repeat the receiver's parameters (`func (s *Stack[T]) ...`) but cannot add new ones.
- Use a generic function (or a `NewX` constructor) when you need inference or extra type parameters.
- Generic types can be recursive, e.g. a linked list node pointing to `*node[T]`.
- Each instantiation is a distinct type; constraints on the type parameters are enforced wherever it is used.
