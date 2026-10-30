---
title: Fuzzing, Coverage and Examples
minutes: 9
---
Three more features of `go test` round out the toolkit: fuzzing finds inputs you did not think of, coverage shows what your tests exercise, and examples are documentation that is verified by the test runner.

## Fuzzing

A **fuzz test** is a function named `FuzzXxx` with the signature `func(f *testing.F)`. Instead of fixed inputs, the engine generates and mutates inputs, hunting for crashes or violated properties. Fuzzing has been built into `go test` since Go 1.18.

Suppose we have:

```go
// Reverse reverses a string rune by rune.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
```

A fuzz test states a **property** that must hold for every input, because the engine cannot know the "right answer" for random data:

```go
func FuzzReverse(f *testing.F) {
	// seed corpus: starting inputs
	f.Add("hello")
	f.Add("")
	f.Add("héllo, 世界")

	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			t.Skip() // Reverse is only specified for valid UTF-8
		}
		twice := Reverse(Reverse(s))
		if twice != s {
			t.Errorf("Reverse(Reverse(%q)) = %q", s, twice)
		}
		if utf8.RuneCountInString(Reverse(s)) != utf8.RuneCountInString(s) {
			t.Errorf("rune count changed for %q", s)
		}
	})
}
```

Typical properties: round trips (encode then decode), idempotence, "never panics", and agreement between two implementations.

### Running it

```bash
go test                       # runs only the seed corpus, like a normal test
go test -fuzz=FuzzReverse     # actually fuzzes, until stopped (Ctrl+C) or a failure
go test -fuzz=FuzzReverse -fuzztime=30s
```

When the engine finds a failing input, it writes it to `testdata/fuzz/FuzzReverse/<hash>` in your package. From then on, a plain `go test` re-runs that input as a regression test, so commit these files. Fuzz arguments may be `string`, `[]byte`, ints, floats, `bool`, and so on. Only one package target can be fuzzed per command, and `-fuzz` accepts a single test.

## Coverage

Coverage tells you which statements your tests executed:

```bash
go test -cover
```

```text
PASS
coverage: 83.3% of statements
ok  	example.com/strx	0.003s
```

For detail, write a profile and view it:

```bash
go test -coverprofile=cover.out ./...
go tool cover -func=cover.out      # per-function percentages
go tool cover -html=cover.out      # opens an annotated source view
```

Red lines in the HTML view were never executed, which usually points to a missing test case, often an error path.

> **Tip:** coverage is a flashlight, not a goal. 100% coverage does not mean the tests check anything meaningful, and chasing the number leads to weak tests. Use it to find *untested* behaviour. Add `-covermode=atomic` together with `-race`.

## Examples

An **example** is a function named `ExampleXxx` with no parameters. If it ends with an `// Output:` comment, `go test` runs it and compares what it prints with the comment. The example then appears in the generated documentation (`pkg.go.dev` and `go doc`).

```go
func ExampleReverse() {
	fmt.Println(Reverse("Go"))
	fmt.Println(Reverse("héllo"))
	// Output:
	// oG
	// olléh
}
```

Naming ties an example to what it documents:

| Function | Documents |
|---|---|
| `ExampleReverse` | the function `Reverse` |
| `ExampleStack` | the type `Stack` |
| `ExampleStack_Push` | the method `Push` on `Stack` |
| `ExampleReverse_unicode` | a second example for `Reverse` (suffix starts lower-case) |
| `Example` | the package as a whole |

Use `// Unordered output:` when lines can come out in any order, such as when ranging over a map. An example with **no** output comment is compiled but not run.

Examples are the best of both worlds: documentation that cannot silently rot, because a wrong `// Output:` makes the test fail.

## The function under test, runnable

The Playground cannot run `go test`, but you can still run the function being tested and see the values the fuzz properties and example describe:

```go run
package main

import (
	"fmt"
	"unicode/utf8"
)

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func main() {
	for _, s := range []string{"Go", "héllo", "", "世界"} {
		rev := Reverse(s)
		fmt.Printf("%q -> %q  roundtrip=%v  runes=%d\n",
			s, rev, Reverse(rev) == s, utf8.RuneCountInString(rev))
	}
}
```

If `Reverse` had worked on bytes instead of runes, the fuzzer would quickly find that `"é"` (two bytes) comes out as invalid UTF-8. That is exactly the kind of bug fuzzing excels at.

## Putting it together

A healthy Go package usually has:

1. Table-driven unit tests for specified behaviour.
2. Examples for the public API.
3. Fuzz tests for parsers and encoders.
4. Benchmarks for hot paths.
5. A CI step running `go vet ./...` and `go test -race -cover ./...`.

## Key takeaways

- Fuzz tests (`FuzzXxx`) check properties over generated inputs; run with `go test -fuzz=Name`.
- Failing fuzz inputs are saved in `testdata/fuzz/` and become regression tests.
- `go test -cover` and `-coverprofile` plus `go tool cover -html` reveal untested code; treat the number as a hint, not a target.
- `ExampleXxx` functions with `// Output:` are run and verified, and show up in the docs.
- Combine unit tests, examples, fuzzing and benchmarks, and run `go test -race ./...` in CI.
