---
title: Interface Design & Composition
minutes: 8
---
Now that you know how interfaces work, here is how Go programmers use them well.

## Accept interfaces, return structs

A widely quoted guideline: **functions should accept interfaces and return concrete types.**

- *Accepting* an interface lets callers pass anything with the right behaviour, including test fakes.
- *Returning* a concrete type gives callers the full API and does not force an abstraction on them. They can always assign the result to an interface of their own choosing.

```go run
package main

import (
	"fmt"
	"io"
	"strings"
)

// Accepts the smallest interface that does the job.
func countWords(r io.Reader) (int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	return len(strings.Fields(string(data))), nil
}

type Report struct {
	Words int
}

// Returns a concrete struct.
func NewReport(r io.Reader) (*Report, error) {
	n, err := countWords(r)
	if err != nil {
		return nil, err
	}
	return &Report{Words: n}, nil
}

func main() {
	rep, err := NewReport(strings.NewReader("small interfaces are great"))
	fmt.Println(rep.Words, err)
}
```

> **Tip:** the main exception is when the whole point of a function is to hide its implementation, for example returning `error`. Even then, the interface is tiny.

## Embedding interfaces

An interface can include other interfaces. The result requires all of their methods.

```go
type Reader interface{ Read(p []byte) (int, error) }
type Closer interface{ Close() error }

type ReadCloser interface {
	Reader
	Closer
}
```

The standard library does exactly this: `io.ReadWriter`, `io.ReadCloser` and `io.ReadWriteCloser` are all built from `Reader`, `Writer` and `Closer`.

```go run
package main

import (
	"fmt"
	"strings"
)

type Greeter interface{ Greet() string }
type Namer interface{ Name() string }

// Requires both behaviours.
type Person interface {
	Greeter
	Namer
}

type English struct{ name string }

func (e English) Name() string  { return e.name }
func (e English) Greet() string { return "Hello, " + e.name }

func introduce(p Person) string {
	return strings.ToUpper(p.Greet()) + " / " + p.Name()
}

func main() {
	fmt.Println(introduce(English{"Ada"}))

	// A Person can be used where only a Greeter is needed.
	var g Greeter = English{"Bob"}
	fmt.Println(g.Greet())
}
```

A value of a larger interface can always be assigned to a variable of a smaller interface whose methods it includes.

## Compile-time interface assertions

Because satisfaction is implicit, you may not notice that a type has stopped implementing an interface until some far-away call site breaks. To catch this at the declaration, add a compile-time assertion:

```go
var _ Shape = (*Circle)(nil)
var _ io.Reader = upperReader{}
var _ fmt.Stringer = Level(0)
```

The blank identifier `_` discards the value, but the assignment still must type-check. If `*Circle` loses a method of `Shape`, the build fails right here with a clear message. The `(*Circle)(nil)` form builds a typed nil pointer without allocating anything.

```go run
package main

import "fmt"

type Speaker interface{ Speak() string }

type Dog struct{}

func (*Dog) Speak() string { return "woof" }

// Compile-time check that *Dog implements Speaker.
// Replacing *Dog with Dog here would be a compile error,
// because Speak has a pointer receiver.
var _ Speaker = (*Dog)(nil)

func main() {
	var s Speaker = &Dog{}
	fmt.Println(s.Speak())
}
```

## Where to define interfaces

| Place | When |
|---|---|
| In the **consumer** package | the default; it states what it needs |
| Next to the implementation | only when the interface is the package's public contract (e.g. `io.Reader`) |

Go interfaces describe what a caller *needs*, so defining them on the producer side ("Java style") usually leads to large interfaces that nobody uses fully.

## Practical guidelines

1. Keep interfaces small (1 to 3 methods).
2. Do not export an interface merely to have one; start with concrete types.
3. Write interfaces when you have at least two real implementations, or when you need a seam for testing.
4. Name the argument types after what the function does with them (`io.Reader`, not `FileLike`).
5. Do not return interface types from constructors unless hiding the type is the explicit goal.

> **Common mistake:** adding an interface for every struct "for mocking" by default. This doubles the code you maintain and often hurts readability. Add one when a test or a second implementation actually needs it.

## Key takeaways

- Accept interfaces, return structs.
- Compose larger interfaces by embedding smaller ones.
- `var _ I = (*T)(nil)` verifies at compile time that `*T` implements `I`.
- Define interfaces in the consumer, keep them small, and add them only when needed.
