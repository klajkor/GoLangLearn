---
title: Struct Basics
minutes: 8
---
A **struct** groups named fields into one type. Go has no classes; structs plus methods (next lessons) cover most of what you would use classes for elsewhere.

## Declaring and creating structs

You declare a struct as a new named type, then create values of it with a *composite literal*.

```go run
package main

import "fmt"

type Point struct {
	X, Y int
}

type User struct {
	Name  string
	Email string
	Age   int
}

func main() {
	// Named fields: preferred, order does not matter.
	u1 := User{Name: "Ada", Email: "ada@example.com", Age: 36}

	// Positional: every field, in declaration order. Brittle.
	p := Point{3, 4}

	// Omitted fields get their zero value.
	u2 := User{Name: "Bob"}

	// The zero value of a struct is every field zeroed.
	var u3 User

	fmt.Println(u1)
	fmt.Println(p)
	fmt.Printf("%+v\n", u2)
	fmt.Printf("%+v\n", u3)

	u3.Name = "Cy"
	fmt.Println(u3.Name, p.X+p.Y)
}
```

Use `%v` to print values, `%+v` to include field names and `%#v` for Go-syntax output.

> **Tip:** Prefer named fields in literals. If someone later adds a field to the struct, positional literals stop compiling, while named literals keep working.

## Useful zero values

Every field starts at its zero value (`0`, `""`, `false`, `nil`). Good Go types are designed so that the zero value is ready to use. A `sync.Mutex` or a `bytes.Buffer` works straight after `var m sync.Mutex`; there is no constructor to forget.

## Pointers to structs

Taking the address of a literal is very common:

```go
p := &Point{X: 1, Y: 2}
p.X = 10 // same as (*p).X = 10; Go dereferences automatically
```

Go uses `.` for both values and pointers; there is no `->`.

## Structs are values

Assigning a struct copies all its fields. Passing it to a function copies it too.

```go run
package main

import "fmt"

type Counter struct {
	N int
}

func bump(c Counter) {
	c.N++ // changes the copy only
}

func bumpPtr(c *Counter) {
	c.N++ // changes the original
}

func main() {
	a := Counter{N: 1}
	b := a // copy
	b.N = 99

	bump(a)
	fmt.Println(a.N, b.N)

	bumpPtr(&a)
	fmt.Println(a.N)
}
```

> **Common mistake:** expecting a function that takes a struct by value to modify the caller's struct. It receives a copy. Pass a pointer when the function must change the original.

Note that the copy is *shallow*: if a field is a slice, map or pointer, both copies share the underlying data.

## Comparing structs

Structs are comparable with `==` if **all** their fields are comparable. The comparison checks field by field.

```go run
package main

import "fmt"

type Point struct{ X, Y int }

type Tagged struct {
	Name string
	Tags []string // slices are NOT comparable
}

func main() {
	fmt.Println(Point{1, 2} == Point{1, 2})
	fmt.Println(Point{1, 2} == Point{2, 1})

	// Comparable structs can be map keys.
	visited := map[Point]bool{{0, 0}: true}
	fmt.Println(visited[Point{0, 0}], visited[Point{5, 5}])

	// Tagged{} == Tagged{} would not compile:
	// invalid operation: struct containing []string cannot be compared
	_ = Tagged{}
}
```

Because comparable structs can be map keys, `map[Point]bool` is a neat way to represent a grid or a set of coordinates.

## Key takeaways

- A struct is a named collection of fields; create values with composite literals, preferably with named fields.
- Omitted fields and `var x T` give zero values; design types so the zero value is useful.
- Structs are copied on assignment and when passed to functions; use pointers to share or mutate.
- `==` works only if every field is comparable (no slices, maps or functions).
- Comparable structs can be used as map keys.
