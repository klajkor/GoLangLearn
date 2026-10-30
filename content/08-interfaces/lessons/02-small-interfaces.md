---
title: Small Interfaces in the Standard Library
minutes: 9
---
The Go proverb says: *the bigger the interface, the weaker the abstraction.* The standard library is full of tiny interfaces that make code wonderfully composable.

## io.Reader and io.Writer

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}
```

`Read` fills the buffer `p` with up to `len(p)` bytes and returns how many it wrote. When the data runs out it returns `io.EOF`. Files, network connections, HTTP bodies, compressed streams, and in-memory strings all implement these, so functions that accept an `io.Reader` work with all of them.

```go run
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// upperReader wraps another Reader and upper-cases what it reads.
type upperReader struct {
	src io.Reader
}

func (u upperReader) Read(p []byte) (int, error) {
	n, err := u.src.Read(p)
	for i := 0; i < n; i++ {
		if p[i] >= 'a' && p[i] <= 'z' {
			p[i] -= 'a' - 'A'
		}
	}
	return n, err
}

func main() {
	r := upperReader{src: strings.NewReader("hello, interfaces!\n")}

	// io.Copy works with any Reader and any Writer.
	n, err := io.Copy(os.Stdout, r)
	fmt.Println("bytes:", n, "err:", err)

	// fmt.Fprintf takes any io.Writer.
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d-%s", 42, "answer")
	fmt.Println(sb.String())
}
```

`upperReader` wraps another reader, a pattern called *decorating*. You can stack such wrappers to build pipelines (decompress, then decrypt, then count bytes) without any of the pieces knowing about each other.

> **Tip:** when you write a function that reads data, accept an `io.Reader` rather than a filename or `*os.File`. It becomes trivially testable with `strings.NewReader`.

## fmt.Stringer

```go
type Stringer interface {
	String() string
}
```

You met it in the structs chapter. `fmt` uses it whenever it prints a value.

## sort.Interface

`sort.Sort` can order any collection that can report its length, compare two elements and swap them:

```go
type Interface interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}
```

```go run
package main

import (
	"fmt"
	"sort"
)

type Person struct {
	Name string
	Age  int
}

type ByAge []Person

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func main() {
	people := []Person{
		{"Alice", 30},
		{"Bob", 25},
		{"Carol", 35},
	}
	sort.Sort(ByAge(people))
	fmt.Println(people)

	// Often simpler: pass a comparison function.
	sort.Slice(people, func(i, j int) bool {
		return people[i].Name > people[j].Name
	})
	fmt.Println(people)
}
```

Conversion `ByAge(people)` is free: it only changes the static type, so `Len`, `Less` and `Swap` become available. In everyday code `sort.Slice` (or the newer `slices.SortFunc`) is more convenient; `sort.Interface` is a great example of how a small interface lets one algorithm sort anything.

## Other interfaces worth knowing

| Interface | Method | Used for |
|---|---|---|
| `io.Closer` | `Close() error` | files, connections |
| `io.ReadWriter` | `Read` + `Write` | combined (built by embedding) |
| `error` | `Error() string` | all errors (next chapter) |
| `http.Handler` | `ServeHTTP(w, r)` | web servers |
| `json.Marshaler` | `MarshalJSON() ([]byte, error)` | custom JSON encoding |

## Designing small interfaces

- One or two methods is typical. Three or more should make you ask whether the caller really needs them all.
- Combine small interfaces by embedding if you need more (see the last lesson).
- Do not create an interface until you have two implementations or a concrete need such as testing.

> **Common mistake:** creating a big `UserService` interface with ten methods "just in case". Callers then depend on more than they use, and fakes become tedious. Declare a one-method interface in the code that consumes it.

## Key takeaways

- Tiny interfaces like `io.Reader`, `io.Writer`, `fmt.Stringer` and `sort.Interface` power much of the standard library.
- Wrapping a Reader or Writer in another lets you compose behaviour without coupling.
- Accept `io.Reader`/`io.Writer` in your APIs for flexibility and easy testing.
- Prefer interfaces with few methods; add them when you need them.
