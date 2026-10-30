---
title: Pointers, Structs and Passing
minutes: 9
---
## Everything is passed by value

Go has exactly one parameter-passing rule: **arguments are copied**. When you call `f(x)`, the function receives its own copy of `x`. For an `int` that means changes inside the function are invisible to the caller. For a pointer, the *address* is copied, so both sides refer to the same underlying value.

```go run
package main

import "fmt"

type Counter struct {
	Name string
	N    int
}

func incByValue(c Counter) {
	c.N++ // modifies the copy
}

func incByPointer(c *Counter) {
	c.N++ // modifies the original
}

func main() {
	c := Counter{Name: "hits"}

	incByValue(c)
	fmt.Println(c.N) // 0

	incByPointer(&c)
	incByPointer(&c)
	fmt.Println(c.N) // 2

	fmt.Printf("%+v\n", c)
}
```

## Automatic dereferencing for structs

With a pointer to a struct you write `p.Field`, not `(*p).Field`. Go dereferences for you. The same applies to method calls: Go takes the address or dereferences as needed.

```go
p := &Counter{Name: "x"}
p.N = 5        // same as (*p).N = 5
fmt.Println(p.Name)
```

This only works for field and method selectors. For a plain `*int` you must write `*p` explicitly.

## Pointer receivers: a recap

Methods can have a value receiver or a pointer receiver:

```go
func (c Counter) Describe() string { return c.Name }  // gets a copy
func (c *Counter) Inc()            { c.N++ }          // can modify
```

| Use a pointer receiver when... | Use a value receiver when... |
|---|---|
| the method must modify the receiver | the type is small and immutable-like (e.g. a `Point`, `time.Time`) |
| the struct is large and copying is costly | the method only reads |
| the struct contains a `sync.Mutex` or similar that must not be copied | the type is a map, func or small basic type |

> **Tip:** pick one style per type. If any method needs a pointer receiver, make them all pointer receivers. Mixing causes surprises with interfaces (a `Counter` value does not have `Inc` in its method set, only `*Counter` does).

Because Go auto-addresses variables, `c.Inc()` works on an addressable variable `c` of type `Counter`. It does not work on a non-addressable value such as a map element or a return value:

```go
m := map[string]Counter{"a": {}}
m["a"].Inc() // compile error: cannot take address of m["a"]
```

## Mutating through a slice or map of structs

Ranging over a slice of structs gives you a **copy** of each element:

```go
items := []Counter{{"a", 0}, {"b", 0}}

for _, it := range items {
    it.N++ // changes the copy only
}

for i := range items {
    items[i].N++ // changes the element
}
```

A slice of pointers (`[]*Counter`) shares the objects, so changing `it.N` inside a `range` modifies the original there.

## Returning pointers

Returning a pointer to a local variable is perfectly safe in Go. The compiler notices the pointer outlives the function and keeps the value alive:

```go
func NewCounter(name string) *Counter {
	c := Counter{Name: name}
	return &c
}
```

This is exactly how constructor functions are written. You will see in a later lesson how the compiler makes it work.

## Copying structs

Assigning a struct copies all of its fields. If a field is a pointer, slice or map, the copy shares what it points to; this is a **shallow** copy.

```go run
package main

import "fmt"

type Profile struct {
	Name string
	Tags []string
}

func main() {
	a := Profile{Name: "ann", Tags: []string{"go", "dev"}}
	b := a // shallow copy

	b.Name = "bob"
	b.Tags[0] = "rust"

	fmt.Println(a.Name, a.Tags) // ann [rust dev]
	fmt.Println(b.Name, b.Tags) // bob [rust dev]
}
```

`Name` was copied, but both `Tags` slices refer to the same backing array. To get independent data you must copy the slice explicitly (`slices.Clone` or `append([]string(nil), a.Tags...)`).

## Key takeaways

- Go passes everything by value; passing a pointer copies the address.
- Use a pointer parameter or receiver to let a function mutate its argument.
- `p.Field` works through a pointer; Go dereferences automatically.
- Prefer one receiver kind per type; use pointer receivers for mutation, large structs, and types with mutexes.
- Ranging over `[]T` yields copies; index into the slice to modify elements.
- Returning `&local` is safe and idiomatic.
- Struct assignment is a shallow copy.
