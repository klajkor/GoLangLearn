---
title: Your First Tests with go test
minutes: 9
---
Testing is built into Go: there is no framework to install. The `testing` package and the `go test` command are all you need, and they are what nearly every Go project uses.

> The in-browser Playground cannot run `go test`, so the test files in this chapter are shown as read-only snippets. Copy them into a module on your machine to try them.

## The rules

1. Test code lives in files whose names end in **`_test.go`**. They are compiled only when testing, never in your normal build.
2. A test is a function named **`TestXxx`** (`Test` followed by an upper-case letter or digit) with the signature `func(t *testing.T)`.
3. Tests usually sit next to the code they test, in the same directory.

```text
mathx/
  go.mod
  mathx.go          // package mathx
  mathx_test.go     // package mathx (or mathx_test)
```

## A function and its test

```go
// mathx.go
package mathx

// Abs returns the absolute value of n.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
```

```go
// mathx_test.go
package mathx

import "testing"

func TestAbs(t *testing.T) {
	got := Abs(-5)
	want := 5
	if got != want {
		t.Errorf("Abs(-5) = %d, want %d", got, want)
	}
}
```

Run it from the module directory:

```bash
go test
```

```text
PASS
ok  	example.com/mathx	0.002s
```

A test **passes** unless it calls a failure method or panics. There are no assertion functions in the standard library; you write plain `if` statements. This keeps tests ordinary Go code.

## Errorf versus Fatalf

| Method | Effect |
|---|---|
| `t.Errorf(format, ...)` | marks the test failed, logs a message, **continues** |
| `t.Fatalf(format, ...)` | marks the test failed, logs a message, **stops this test** immediately |
| `t.Error`, `t.Fatal` | same, using `Sprint`-style arguments |
| `t.Log`, `t.Logf` | log a message (shown on failure or with `-v`) |
| `t.Skip` | skip the test, for example when a resource is missing |

Use `Fatalf` when continuing makes no sense, such as when setup failed or a pointer you are about to dereference is nil. Use `Errorf` to report several independent problems in one run.

> **Convention:** failure messages read `FuncName(input) = got, want expected`. Put the actual value first, then the expectation, and include the input so the failure is understandable without re-reading the test.

```go
func TestParse(t *testing.T) {
	cfg, err := Parse("a=1")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err) // cfg is unusable, so stop
	}
	if cfg.Count != 1 {
		t.Errorf("Count = %d, want 1", cfg.Count)
	}
}
```

## Running tests

```bash
go test                  # the package in the current directory
go test ./...            # every package in the module
go test -v               # verbose: print each test name and result
go test -run TestAbs     # only tests whose names match this regexp
go test -count=1         # bypass the test result cache
go test -race            # enable the data race detector
go test -failfast        # stop at the first failure
```

Go **caches** successful results when the code and inputs have not changed; you will see `(cached)` in the output. Use `-count=1` when you want a fresh run, for example for tests that touch the environment.

Verbose output looks like this:

```text
=== RUN   TestAbs
--- PASS: TestAbs (0.00s)
PASS
ok  	example.com/mathx	0.002s
```

## Same package or external package?

A test file may declare `package mathx` (an **internal** test with access to unexported names) or `package mathx_test` (an **external** test that sees only the exported API, just as a user would). Both can coexist in one directory. Prefer external tests for public behaviour and internal ones when you need to reach implementation details.

```go
package mathx_test

import (
	"testing"

	"example.com/mathx"
)

func TestAbsPublic(t *testing.T) {
	if mathx.Abs(-1) != 1 {
		t.Fatal("Abs(-1) != 1")
	}
}
```

## Reproducing the idea in a program

Here is a plain program with the same shape: call, compare, report. It runs in the Playground, but a real test lets `go test` do the reporting.

```go run
package main

import "fmt"

func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func main() {
	inputs := []int{-5, 0, 7}
	wants := []int{5, 0, 7}
	for i, in := range inputs {
		if got := Abs(in); got != wants[i] {
			fmt.Printf("FAIL: Abs(%d) = %d, want %d\n", in, got, wants[i])
			continue
		}
		fmt.Printf("ok: Abs(%d) = %d\n", in, wants[i])
	}
}
```

## Key takeaways

- Tests live in `_test.go` files and are named `TestXxx(t *testing.T)`.
- Go has no assertion library; compare with `if` and report with `t.Errorf` or `t.Fatalf`.
- `Errorf` continues the test; `Fatalf` stops it.
- Use `go test ./...`, `-v`, `-run`, `-count=1` and `-race` constantly.
- Results are cached; `-count=1` forces a re-run.
- Use `package foo_test` to test only the public API.
