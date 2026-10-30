---
title: Sentinel Errors & Custom Error Types
minutes: 8
---
Sometimes a caller needs to know *which* error happened, not just that there was one. Go gives you two ways to make errors identifiable: sentinel values and custom types.

## Sentinel errors

A **sentinel error** is a package-level variable that callers compare against. The convention is to name them `ErrSomething`.

```go run
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")
var ErrEmptyName = errors.New("empty name")

var users = map[string]int{"ada": 36, "bob": 25}

func age(name string) (int, error) {
	if name == "" {
		return 0, ErrEmptyName
	}
	a, ok := users[name]
	if !ok {
		return 0, ErrNotFound
	}
	return a, nil
}

func main() {
	for _, name := range []string{"ada", "zed", ""} {
		a, err := age(name)
		switch {
		case err == ErrNotFound:
			fmt.Printf("%q: no such user\n", name)
		case err == ErrEmptyName:
			fmt.Println("please give a name")
		case err != nil:
			fmt.Println("unexpected:", err)
		default:
			fmt.Printf("%q is %d\n", name, a)
		}
	}
}
```

Well-known examples from the standard library: `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`.

Two caveats:

- Sentinels become part of your package's public API. Callers will depend on them.
- Comparing with `==` breaks as soon as the error is wrapped with extra context. The next lesson fixes this with `errors.Is`, which is what you should normally use instead of `==`.

> **Tip:** two `errors.New("x")` calls produce two *different* errors, even with identical text. Comparison is by identity, which is why sentinels are declared once as variables.

## Custom error types

When callers need more than identity, such as the failing field, a status code or a retry hint, define your own type that implements `error`.

```go run
package main

import (
	"errors"
	"fmt"
)

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}

func validate(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Reason: "must not be negative"}
	}
	if age > 150 {
		return &ValidationError{Field: "age", Reason: "unrealistic"}
	}
	return nil
}

func main() {
	for _, a := range []int{30, -1, 200} {
		err := validate(a)
		if err == nil {
			fmt.Println(a, "ok")
			continue
		}

		var ve *ValidationError
		if errors.As(err, &ve) {
			fmt.Printf("%d: field=%s reason=%s\n", a, ve.Field, ve.Reason)
		}
	}
}
```

`errors.As` finds an error of a given type and gives you access to its fields; it is explained properly in the next lesson.

Notes on custom types:

- Use a **pointer receiver** and return `&ValidationError{...}` so identity and `nil` checks behave well, matching how the standard library (`*os.PathError`, `*strconv.NumError`) does it.
- Name the type `XxxError`.
- Return the `error` interface from functions, never the concrete type (see the nil-interface gotcha in the interfaces chapter).

```go
// BAD: a nil *ValidationError returned as error is non-nil!
func validate(age int) *ValidationError { ... }

// GOOD
func validate(age int) error { ... }
```

## Sentinel or type?

| Need | Use |
|---|---|
| Caller only needs to know *which* condition | sentinel (`ErrNotFound`) |
| Caller needs extra data (field, code, path) | custom type |
| Message is for humans only | plain `errors.New` / `fmt.Errorf` |

Do not expose more error API than you must. Every sentinel and type you export is a promise to your users.

## Key takeaways

- Sentinel errors are package-level `ErrXxx` variables, compared by identity.
- Custom error types carry structured data; implement `Error() string` on a pointer receiver.
- Return `error`, never a concrete error pointer type.
- Prefer `errors.Is` and `errors.As` over `==` and type assertions so wrapping does not break your checks.
