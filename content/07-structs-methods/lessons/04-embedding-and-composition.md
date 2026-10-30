---
title: Embedding & Composition
minutes: 8
---
Go has no inheritance. To reuse behaviour, you **compose** types, and **embedding** makes that convenient.

## Embedding a struct

If you list a type in a struct without a field name, it is *embedded*. Its fields and methods are **promoted** to the outer type.

```go run
package main

import "fmt"

type Animal struct {
	Name string
}

func (a Animal) Describe() string {
	return "I am " + a.Name
}

type Dog struct {
	Animal // embedded
	Breed  string
}

func main() {
	d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Collie"}

	fmt.Println(d.Name)        // promoted field
	fmt.Println(d.Describe())  // promoted method
	fmt.Println(d.Animal.Name) // the explicit path still works
}
```

The embedded field's name is the type name (`Animal`), which is why the literal says `Animal: Animal{...}`.

## Shadowing and overriding

If the outer type defines a method or field with the same name, it wins. The embedded one is still reachable explicitly.

```go run
package main

import "fmt"

type Base struct{}

func (Base) Hello() string { return "hello from Base" }
func (b Base) Run() string { return b.Hello() }

type Derived struct{ Base }

func (Derived) Hello() string { return "hello from Derived" }

func main() {
	d := Derived{}
	fmt.Println(d.Hello())      // Derived's version
	fmt.Println(d.Base.Hello()) // Base's version
	fmt.Println(d.Run())        // Base.Run calls Base.Hello
}
```

The last line is the key difference from inheritance: `Base.Run` has no idea it is embedded, so it calls `Base.Hello`, not `Derived.Hello`. There is no virtual dispatch. Embedding is automatic delegation, not subclassing.

## Embedding is not "is-a"

A `Dog` is **not** an `Animal` in the type system. You cannot pass a `Dog` where an `Animal` parameter is expected. What you get is:

- shorter code (no `d.Animal.Name` boilerplate), and
- promoted methods, which count towards interface satisfaction (the interfaces chapter).

> **Tip:** prefer *composition over inheritance*. Embed a type when you want its behaviour as part of your type's API. Use a named field (`logger Logger`) when it is just an implementation detail you do not want to expose.

## Embedding a pointer

You can embed `*T` as well. The promoted methods then act on the shared value. Remember it is `nil` unless you initialise it.

```go
type Service struct {
	*log.Logger // methods like Println are promoted
}

s := Service{Logger: log.New(os.Stderr, "svc: ", 0)}
s.Println("started")
```

## Name conflicts

If two embedded types provide the same name at the same depth, the name is *ambiguous*. It is only an error if you actually use it:

```go
type A struct{}

func (A) Hi() {}

type B struct{}

func (B) Hi() {}

type C struct {
	A
	B
}

// c.Hi()   // compile error: ambiguous selector c.Hi
// c.A.Hi() // fine
```

## Embedding to extend a type

A common pattern is embedding to add a feature while keeping everything else:

```go run
package main

import (
	"fmt"
	"sync"
)

type SafeCounter struct {
	sync.Mutex // Lock and Unlock are promoted
	counts     map[string]int
}

func (c *SafeCounter) Inc(key string) {
	c.Lock()
	defer c.Unlock()
	c.counts[key]++
}

func main() {
	c := SafeCounter{counts: map[string]int{}}
	c.Inc("a")
	c.Inc("a")
	c.Inc("b")
	fmt.Println(c.counts["a"], c.counts["b"])
}
```

Embedding `sync.Mutex` also exports `Lock`/`Unlock` to users of `SafeCounter`, which is often not what you want in a public type. A named field `mu sync.Mutex` is usually tidier for exported types.

## Key takeaways

- Embedding promotes fields and methods of the inner type to the outer type.
- The outer type may shadow promoted names; the inner one stays reachable via its type name.
- There is no virtual dispatch: inner methods never call the outer overrides.
- Embedding is composition, not inheritance; the outer type is not a subtype of the inner.
- Ambiguous promoted names are only an error when used.
