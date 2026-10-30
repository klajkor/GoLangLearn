---
title: Table-Driven Tests and Subtests
minutes: 9
---
When a function has many interesting inputs, copy-pasting `if` blocks gets noisy. The idiomatic Go answer is the **table-driven test**: a slice of cases and one loop that runs them all.

## The pattern

```go
func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, 5},
		{"negative", -5, 5},
		{"zero", 0, 0},
		{"negative one", -1, 1},
	}

	for _, tt := range tests {
		got := Abs(tt.in)
		if got != tt.want {
			t.Errorf("%s: Abs(%d) = %d, want %d", tt.name, tt.in, got, tt.want)
		}
	}
}
```

Adding a case is one line. The anonymous struct declares exactly the fields the test needs: inputs, expected outputs, and a descriptive `name`. You can include an `wantErr bool` or `wantErr error` field to cover failure paths too.

## Subtests with t.Run

`t.Run(name, func(t *testing.T) {...})` runs each case as a **subtest**. You get independent pass/fail status, readable names, and the ability to run one case.

```go
func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, 5},
		{"negative", -5, 5},
		{"zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Abs(tt.in); got != tt.want {
				t.Errorf("Abs(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
```

With `-v`, the output shows the hierarchy:

```text
=== RUN   TestAbs
=== RUN   TestAbs/positive
=== RUN   TestAbs/negative
=== RUN   TestAbs/zero
--- PASS: TestAbs (0.00s)
    --- PASS: TestAbs/positive (0.00s)
    --- PASS: TestAbs/negative (0.00s)
    --- PASS: TestAbs/zero (0.00s)
```

Subtest names with spaces are rewritten with underscores (`negative one` becomes `negative_one`). Select a subtest with a slash-separated pattern:

```bash
go test -run 'TestAbs/negative' -v
go test -run 'TestAbs/^zero$'
```

Because each subtest has its own `*testing.T`, `Fatalf` in a case stops only that case, and the rest keep running. That is a major advantage over the plain loop.

## Parallel subtests

Calling `t.Parallel()` inside a subtest lets it run concurrently with other parallel subtests of the same parent:

```go
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		t.Parallel()
		if got := Slow(tt.in); got != tt.want {
			t.Errorf("Slow(%d) = %d, want %d", tt.in, got, tt.want)
		}
	})
}
```

A parallel subtest pauses until the parent's sequential portion finishes, then runs alongside the other parallel ones; the parent test waits for them all. The default parallelism is `GOMAXPROCS`, adjustable with `-parallel N`.

> **Common mistake (pre-1.22):** before Go 1.22 the loop variable `tt` was shared by all iterations, so parallel subtests all saw the last case. The old fix was `tt := tt` at the top of the loop body. With `go 1.22` or later in `go.mod` it is no longer needed, but you will still see it in older code.

Only use `t.Parallel()` when the tests do not share mutable state. Shared files, globals, environment variables and ports are all common sources of flaky parallel tests.

## Comparing complex values

For slices, maps and structs, `!=` does not compile or compares the wrong thing. Options:

| Tool | Use |
|---|---|
| `slices.Equal(a, b)` | slices of comparable elements |
| `maps.Equal(a, b)` | maps with comparable values |
| `reflect.DeepEqual(a, b)` | anything (slow, strict about nil vs empty) |
| `cmp.Diff(want, got)` from `github.com/google/go-cmp` | third-party, readable diffs |

```go
if !slices.Equal(got, tt.want) {
	t.Errorf("Split(%q) = %v, want %v", tt.in, got, tt.want)
}
```

Note: `reflect.DeepEqual([]int{}, []int(nil))` is `false`, which surprises people. Decide whether nil and empty should be treated the same for your function.

## Testing errors

```go
func TestParseAge(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"42", 42, false},
		{"-1", 0, true},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseAge(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseAge(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseAge(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
```

For sentinel errors use `errors.Is(err, ErrNegative)`; for typed errors, `errors.As`. Do not compare error message strings unless the message *is* the contract.

## A runnable mini-table

Here is the same table idea in a program you can run. It is not `go test`, but it shows the data-driven structure and the style of failure message:

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"lower", "go", "GO"},
		{"mixed", "GoLang", "GOLANG"},
		{"empty", "", ""},
		{"wrong on purpose", "x", "y"},
	}

	failed := 0
	for _, tt := range tests {
		if got := strings.ToUpper(tt.in); got != tt.want {
			fmt.Printf("FAIL %s: ToUpper(%q) = %q, want %q\n", tt.name, tt.in, got, tt.want)
			failed++
			continue
		}
		fmt.Printf("PASS %s\n", tt.name)
	}
	fmt.Println("failed:", failed)
}
```

## Key takeaways

- Table-driven tests put cases in a slice and run them in one loop; adding cases is cheap.
- `t.Run` creates named subtests with independent results; select them with `-run 'Test/name'`.
- `t.Parallel()` runs subtests concurrently; avoid it with shared mutable state.
- Since Go 1.22, loop variables are per-iteration, so `tt := tt` is no longer required.
- Use `slices.Equal`, `maps.Equal` or `reflect.DeepEqual` for non-comparable types.
- Test errors with a `wantErr` flag, `errors.Is` or `errors.As`.
