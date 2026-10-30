---
title: Errors Are Values
minutes: 8
---
Go has no exceptions for ordinary failures. Functions that can fail return an **error** as an extra result, and the caller decides what to do. It is verbose by design: every failure path is visible in the code.

## The error interface

`error` is a built-in interface with a single method:

```go
type error interface {
	Error() string
}
```

Any type with an `Error() string` method is an error. The zero value of the interface, `nil`, means "no error".

## Creating errors

The two everyday tools are `errors.New` for a fixed message and `fmt.Errorf` for a formatted one.

```go run
package main

import (
	"errors"
	"fmt"
	"strconv"
)

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func parseAge(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parseAge %q: %v", s, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("parseAge: negative age %d", n)
	}
	return n, nil
}

func main() {
	q, err := divide(10, 2)
	fmt.Println(q, err)

	_, err = divide(1, 0)
	fmt.Println("error:", err)

	for _, in := range []string{"42", "abc", "-5"} {
		age, err := parseAge(in)
		if err != nil {
			fmt.Println("failed:", err)
			continue
		}
		fmt.Println("age:", age)
	}
}
```

## The idiom

You will write this shape thousands of times:

```go
result, err := doSomething()
if err != nil {
	return fmt.Errorf("doing something: %w", err) // or handle it
}
// use result
```

Conventions to follow:

- The error is the **last** return value.
- Check `err != nil` **immediately** after the call, before using other results.
- On error, the other results are usually zero values and should be ignored (unless the docs say otherwise, as with `io.Reader`, which can return `n > 0` together with an error).
- Return early; keep the happy path unindented ("line of sight").

## Error messages

- Start lowercase and do not end with punctuation, because messages are often chained: `open config: parse port: invalid syntax`.
- Include context about what you were doing, not stack-trace noise.
- Do not repeat the word "error" or "failed" in every layer.

## Errors are ordinary values

Because errors are just values you can compare them, store them, put them in structs or send them over channels. That is the idea behind Rob Pike's line "errors are values": handle them with ordinary programming, not special syntax.

```go run
package main

import (
	"errors"
	"fmt"
)

// errWriter remembers the first error so callers check once at the end.
type errWriter struct {
	err   error
	total int
}

func (w *errWriter) write(s string) {
	if w.err != nil {
		return // skip after the first failure
	}
	if s == "" {
		w.err = errors.New("empty chunk")
		return
	}
	w.total += len(s)
}

func main() {
	w := &errWriter{}
	w.write("abc")
	w.write("")
	w.write("never counted")
	fmt.Println(w.total, w.err)
}
```

This removes repetitive `if err != nil` blocks when many steps share one failure policy.

## Do not ignore errors

Assigning an error to `_` discards it silently:

```go
data, _ := os.ReadFile("config.json") // if this fails, data is empty and you will not know why
```

> **Common mistake:** ignoring an error "because it never fails". It will, eventually, and the failure will surface far from its cause. If you really mean to ignore it, do so visibly (`_ = f.Close()`) and ideally leave a comment saying why.

Linters such as `errcheck` flag unchecked errors for this reason.

## Key takeaways

- `error` is an interface with `Error() string`; `nil` means success.
- Create errors with `errors.New` or `fmt.Errorf`.
- Return errors as the last value and check them right away, returning early.
- Write lowercase, unpunctuated messages with useful context.
- Errors are ordinary values; do not discard them silently.
