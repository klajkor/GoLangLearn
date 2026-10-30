---
title: Pointer Patterns and Gotchas
minutes: 9
---
## When to use a pointer

| Situation | Choose |
|---|---|
| A function or method must **modify** its argument | pointer |
| The struct is **large**, or copying it is wasteful | pointer |
| The value must express **"absent"** (`nil`) | pointer |
| The type holds a mutex, file handle, or other uncopyable state | pointer |
| Small, read-only data (`Point`, `Duration`, config copies) | value |
| Slices, maps, channels, funcs | usually the value itself |
| You want to prevent unexpected mutation | value |

Values are simpler to reason about: no `nil`, no aliasing, no accidental sharing. Start with values and reach for a pointer when you have one of the reasons above.

## Optional values with pointers

A `*T` can encode "maybe a `T`". This is common in JSON structs and database models, where a missing field and a zero value are different things:

```go run
package main

import "fmt"

type Settings struct {
	Name    string
	Retries *int // nil means "use the default"
}

func retries(s Settings) int {
	if s.Retries == nil {
		return 3 // default
	}
	return *s.Retries
}

func ptr[T any](v T) *T { return &v }

func main() {
	a := Settings{Name: "a"}
	b := Settings{Name: "b", Retries: ptr(0)} // explicitly zero

	fmt.Println(retries(a), retries(b)) // 3 0
}
```

The generic helper `ptr` is a handy way to take the address of a literal, since `&0` is illegal.

## Swap and out-parameters

Classic pointer example, still useful to understand:

```go run
package main

import "fmt"

func swap(a, b *int) {
	*a, *b = *b, *a
}

func main() {
	x, y := 1, 2
	swap(&x, &y)
	fmt.Println(x, y) // 2 1
}
```

In idiomatic Go you more often **return multiple values** (`a, b = b, a`) than use out-parameters. Use pointer parameters when the data has to be filled in place, such as `json.Unmarshal(data, &v)` or `fmt.Sscanf`.

## Loop variables and pointers (Go 1.22+)

Before Go 1.22 a `for` loop reused **one** variable for every iteration, so taking its address inside the loop produced the same pointer each time. This caused one of Go's most famous bugs. Since Go 1.22 (for modules whose `go.mod` says `go 1.22` or later), **each iteration gets its own copy** of the loop variable.

```go run
package main

import "fmt"

func main() {
	var ptrs []*int
	for i := 0; i < 3; i++ {
		ptrs = append(ptrs, &i)
	}
	for _, p := range ptrs {
		fmt.Print(*p, " ")
	}
	fmt.Println()

	words := []string{"a", "b", "c"}
	var wp []*string
	for _, w := range words {
		wp = append(wp, &w)
	}
	for _, p := range wp {
		fmt.Print(*p, " ")
	}
	fmt.Println()
}
```

With Go 1.22 semantics this prints `0 1 2` and `a b c`. Under the old semantics it printed `3 3 3` and `c c c`. Closures and goroutines started inside loops benefit in the same way.

> **Common mistake:** relying on 1.22 behaviour in a module whose `go.mod` still says `go 1.21` or lower. The language version in `go.mod` decides the semantics, not the toolchain you happen to run. Check that line when you see the old behaviour.

## Aliasing: two names, one value

```go
a := &Config{Port: 80}
b := a // same Config
b.Port = 8080
fmt.Println(a.Port) // 8080
```

Aliasing is the price of sharing. When a function stores a pointer you gave it, later changes you make through your pointer are visible to it, and vice versa. If a struct should own its data, copy on the way in:

```go
func NewServer(cfg Config) *Server { // takes a copy
	return &Server{cfg: cfg}
}
```

## Pointer-to-slice is rarely needed

Beginners sometimes write `func add(s *[]int)` to make `append` work. Prefer returning the new slice (`s = add(s, 1)`), which is how `append` itself works. A pointer to a slice is only worthwhile in special cases, such as methods on a named slice type that must replace the slice.

## Pointers to interfaces: almost never

An interface value already holds a pointer (or a copy of a value) plus its type. `*io.Reader` is almost always a mistake. Take or return the interface itself.

## Typed nil trap, in one line

A nil `*T` stored in an interface is **not** a nil interface:

```go
var p *MyErr = nil
var err error = p
fmt.Println(err == nil) // false
```

Return a literal `nil` for "no error" rather than a typed nil pointer. Later chapters on interfaces and errors revisit this.

## Checklist

1. Can the zero value or a copy do the job? Use a value.
2. Does the callee need to change the caller's data? Pass a pointer.
3. Does the struct hold a lock or resource? Always pass `*T`.
4. Is `nil` a meaningful state? A pointer may fit.
5. Worried about speed? Benchmark first.

## Key takeaways

- Use pointers for mutation, large or uncopyable types, and optional values; otherwise prefer values.
- A `*T` field can distinguish "unset" from the zero value, but remember to nil-check.
- Since Go 1.22 every loop iteration has its own variable, so `&i` and closures capture per-iteration values (depends on `go` version in `go.mod`).
- Aliasing means one change is seen through every pointer; copy on the way in when a type should own its data.
- Avoid pointers to slices and interfaces unless you have a specific reason.
- A nil pointer inside an interface is not a nil interface.
