---
title: Methods & Receivers
minutes: 9
---
A **method** is a function with a special *receiver* parameter. You can attach methods to any type you define in your own package, not only structs.

## Defining methods

```go run
package main

import "fmt"

type Rect struct {
	W, H float64
}

// Value receiver: works on a copy.
func (r Rect) Area() float64 {
	return r.W * r.H
}

// Pointer receiver: can modify the original.
func (r *Rect) Scale(k float64) {
	r.W *= k
	r.H *= k
}

func main() {
	r := Rect{W: 2, H: 3}
	fmt.Println(r.Area())

	r.Scale(2)
	fmt.Println(r.W, r.H, r.Area())
}
```

Notice `r.Scale(2)` works even though `r` is a plain `Rect` and `Scale` wants a `*Rect`. Go rewrites it to `(&r).Scale(2)` automatically, but only if `r` is **addressable** (a variable, a slice element, a field of an addressable struct).

## Value vs pointer receivers

| | Value receiver `(r Rect)` | Pointer receiver `(r *Rect)` |
|---|---|---|
| Gets | a copy | the original |
| Can mutate caller's value | no | yes |
| Cost | copies the struct | copies one pointer |
| Called through a nil pointer | panics (needs to dereference) | works; the method can handle `nil` |

Rules of thumb:

- Use a **pointer receiver** if the method modifies the receiver, if the struct is large, or if it contains something that must not be copied (a `sync.Mutex`, for example).
- Use a **value receiver** for small, immutable-feeling types (points, money amounts, `time.Time`).
- **Be consistent:** if any method needs a pointer receiver, make all methods on that type pointer receivers.

## The mutation gotcha

```go run
package main

import "fmt"

type Counter struct{ N int }

func (c Counter) IncBroken() { c.N++ } // modifies a copy
func (c *Counter) Inc()      { c.N++ }

func main() {
	var c Counter
	c.IncBroken()
	c.IncBroken()
	fmt.Println(c.N)

	c.Inc()
	c.Inc()
	fmt.Println(c.N)
}
```

> **Common mistake:** a value-receiver method that "updates" a field. It silently changes a copy and the caller never sees it. The compiler will not warn you, so choose receivers deliberately.

## Addressability

Pointer-receiver methods need an addressable value. These cases are *not* addressable:

```go
type T struct{}

func (t *T) Ptr() {}

func main() {
	T{}.Ptr()                 // compile error: cannot call pointer method Ptr on T
	m := map[string]T{"a": {}}
	m["a"].Ptr()              // compile error: map elements are not addressable
}
```

A slice element `s[i]` is addressable, and so is a variable. A map element and the result of a function call are not.

## Method sets (a teaser)

Every type has a **method set**: the methods you can call through a value of that type when it is stored in an interface.

- The method set of `T` contains the methods with receiver `T`.
- The method set of `*T` contains methods with receiver `T` **and** `*T`.

So a `*Rect` can use both `Area` and `Scale`, but a plain `Rect` stored in an interface only has `Area`. This becomes important in the interfaces chapter: if `Scale` is needed to satisfy an interface, you must store a `*Rect`, not a `Rect`.

## Methods on non-struct types

Any named type declared in your package can have methods:

```go run
package main

import "fmt"

type Celsius float64

func (c Celsius) Fahrenheit() float64 {
	return float64(c)*9/5 + 32
}

type IntList []int

func (l IntList) Sum() int {
	total := 0
	for _, v := range l {
		total += v
	}
	return total
}

func main() {
	fmt.Println(Celsius(100).Fahrenheit())
	fmt.Println(IntList{1, 2, 3}.Sum())
}
```

You cannot add methods to types from other packages (such as `int` or `time.Time`) directly; define your own named type instead.

## Key takeaways

- Methods are functions with a receiver; any local named type can have them.
- Value receivers get a copy; pointer receivers can mutate the original and avoid copying.
- Do not mix receiver kinds on one type without a good reason.
- Pointer-method calls on addressable values are rewritten for you; map elements and call results are not addressable.
- Method sets: `T` has value-receiver methods, `*T` has both. Interfaces will make this matter.
