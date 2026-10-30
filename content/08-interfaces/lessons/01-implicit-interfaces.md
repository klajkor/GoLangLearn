---
title: Implicit Interfaces
minutes: 8
---
An **interface** is a set of method signatures. Any type that has those methods *is* an implementation of the interface. There is no `implements` keyword; satisfaction is **implicit** and checked by the compiler.

## A first interface

```go run
package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

type Circle struct{ R float64 }

func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

func describe(s Shape) {
	fmt.Printf("area=%.2f perimeter=%.2f\n", s.Area(), s.Perimeter())
}

func main() {
	shapes := []Shape{Rect{3, 4}, Circle{1}}
	for _, s := range shapes {
		describe(s)
	}
}
```

Neither `Rect` nor `Circle` mentions `Shape`. They simply have the right methods. That decoupling is the point: the interface can be defined *later*, and *by the consumer*, without touching the types it describes. You can even write an interface for a type from a third-party package.

## Why implicit?

- **Decoupling.** The package that defines a type does not have to import the package that defines the interface.
- **Small, local interfaces.** A function can declare exactly the behaviour it needs, nothing more.
- **Easy testing.** Replace a real database with a tiny fake that has the same methods.

```go run
package main

import "fmt"

// The consumer defines the interface it needs.
type Notifier interface {
	Notify(msg string) error
}

type Email struct{ To string }

func (e Email) Notify(msg string) error {
	fmt.Printf("email to %s: %s\n", e.To, msg)
	return nil
}

// A test double: no email, just records.
type Recorder struct{ Sent []string }

func (r *Recorder) Notify(msg string) error {
	r.Sent = append(r.Sent, msg)
	return nil
}

func alertOnFailure(n Notifier, failed bool) {
	if failed {
		_ = n.Notify("job failed")
	}
}

func main() {
	alertOnFailure(Email{To: "ops@example.com"}, true)

	rec := &Recorder{}
	alertOnFailure(rec, true)
	alertOnFailure(rec, false)
	fmt.Println(len(rec.Sent), rec.Sent)
}
```

## Method sets decide satisfaction

Remember from the structs chapter: a value of type `T` has only the value-receiver methods, while `*T` has both. Here `Recorder.Notify` has a pointer receiver, so only `*Recorder` satisfies `Notifier`. This does not compile:

```go
alertOnFailure(Recorder{}, true)
// cannot use Recorder{} (value of type Recorder) as Notifier value:
// Recorder does not implement Notifier (method Notify has pointer receiver)
```

> **Common mistake:** putting a struct value into an interface when the methods have pointer receivers. Pass `&T{}` instead.

## Interface-typed values

A variable of interface type can hold any value whose type satisfies it, and calling a method dispatches dynamically to that type's implementation:

```go
var s Shape      // nil interface for now
s = Rect{2, 3}   // holds a Rect
s = Circle{5}    // now holds a Circle
```

An interface type is a *contract*, not a data type. You cannot access fields through it, only call the methods it lists. The next lessons show how to get back to the concrete type.

## Interface naming

By convention, one-method interfaces are named after the method with an `-er` suffix: `Reader`, `Writer`, `Stringer`, `Closer`. Multi-method interfaces get descriptive names, like `Shape`.

## Key takeaways

- A type satisfies an interface by having all its methods; there is no `implements`.
- Define interfaces where they are *used*, and keep them small.
- Pointer-receiver methods belong only to `*T`, so only `*T` satisfies an interface that needs them.
- Interface values give dynamic dispatch: the method of the stored concrete type runs.
