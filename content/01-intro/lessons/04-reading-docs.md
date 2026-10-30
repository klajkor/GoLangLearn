---
title: Docs, go doc and the Playground
minutes: 6
---
Go's standard library is large, and knowing how to look things up is a core skill. The good news: documentation is generated from the source code itself, so it is always consistent and always in the same format.

## pkg.go.dev

[pkg.go.dev](https://pkg.go.dev) hosts the documentation for the standard library and for every public Go module. Search for a package (for example `strings`) and you get an overview, an index of all functions and types, and runnable examples. Each function links to its source code, which in Go is usually short and readable. Reading the standard library is one of the best ways to learn idiomatic Go.

## go doc in the terminal

You do not even need a browser:

```bash
go doc strings            # package overview and exported names
go doc strings.ToUpper    # one function
go doc strings.Builder    # one type and its methods
go doc -all strings       # everything in the package
```

```text
package strings // import "strings"

func ToUpper(s string) string
    ToUpper returns s with all Unicode letters mapped to their upper case.
```

(The output is shortened.) This is the quickest way to check an argument order or return type while you are coding.

## How docs are written

Documentation is just comments placed directly above a declaration, starting with the name of the thing described:

```go
// Greet returns a friendly greeting for the given name.
func Greet(name string) string {
	return "Hello, " + name
}
```

Tools turn that comment into the page on pkg.go.dev. A comment immediately before `package` documents the whole package. Because the rule is so simple, almost every library is documented.

## Reading a signature

Most of your time in the docs is spent decoding signatures. A few patterns you will see constantly:

| Signature | Reading |
|---|---|
| `func Contains(s, substr string) bool` | takes two strings, returns a bool |
| `func Atoi(s string) (int, error)` | returns two values: a result and an error |
| `func (b *Builder) WriteString(s string) (int, error)` | a method on `*Builder` |
| `func Println(a ...any) (n int, err error)` | variadic: accepts any number of arguments |

The `(T, error)` shape is everywhere. Go does not use exceptions, so a function that can fail returns an error as its last result.

## Examples are runnable

Most packages include examples that show real output. Try one yourself. Here is `strings` doing a few typical jobs, exactly the sort of thing you would find after a quick docs search:

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.Contains("seafood", "foo"))
	fmt.Println(strings.Split("a,b,c", ","))
	fmt.Println(strings.Repeat("ab", 3))
	fmt.Println(strings.TrimSpace("  padded  ") + "|")
	fmt.Println(strings.Fields(" one  two three "))
}
```

Guess each result before you press **Run**, then check the docs for `Fields` to see how it splits on runs of whitespace.

## The Go Playground

[go.dev/play](https://go.dev/play) is a free online editor that compiles and runs Go on Google's servers. It is how the editable blocks in this course work. It is ideal for trying a snippet or sharing a reproducible example via a short link (the "Share" button). Limits to remember: no network access, a fixed clock, and short execution time, so it is for experiments, not for real programs.

## Other places to learn

- **A Tour of Go** (go.dev/tour): an interactive introduction.
- **Effective Go** and the **Go spec** (go.dev/ref/spec): the language definition; surprisingly readable.
- **Go by Example** (gobyexample.com): short annotated snippets.

> **Tip:** When you are unsure what a function does, run `go doc pkg.Func` first. It takes one second and the answer is authoritative.

## Key takeaways

- Documentation lives in code comments and is rendered on pkg.go.dev.
- `go doc <package>[.<name>]` shows docs in the terminal.
- Learn to read signatures; `(value, error)` returns are the norm.
- The Playground runs snippets online and is shareable, but is limited (no network, short runs).
