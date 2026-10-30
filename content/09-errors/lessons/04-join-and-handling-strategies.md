---
title: errors.Join & Handling Strategies
minutes: 7
---
Sometimes more than one thing goes wrong, and you want to report all of it. Go 1.20 added `errors.Join` for exactly this.

## errors.Join

`errors.Join(errs...)` combines several errors into one. `nil` entries are discarded, and if every argument is `nil` the result is `nil`. The message is each error's text on its own line.

```go run
package main

import (
	"errors"
	"fmt"
)

var (
	ErrNameEmpty = errors.New("name is empty")
	ErrAgeRange  = errors.New("age out of range")
)

func validate(name string, age int) error {
	var errs []error
	if name == "" {
		errs = append(errs, ErrNameEmpty)
	}
	if age < 0 || age > 150 {
		errs = append(errs, ErrAgeRange)
	}
	return errors.Join(errs...) // nil when errs is empty
}

func main() {
	err := validate("", 200)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrNameEmpty), errors.Is(err, ErrAgeRange))

	fmt.Println(validate("Ada", 36) == nil)
}
```

The joined error supports `errors.Is` and `errors.As` for every member, so callers can test for each individual problem. It is ideal for validation, cleanup (closing several resources) and fan-out work where several tasks can fail.

Calling `errors.Join(errs...)` with an empty slice returns `nil`, which means you can return it directly without an `if len(errs) > 0` check.

## Handling strategies

When you get an error you have four options:

1. **Return it**, possibly wrapped with context. The most common choice.
2. **Handle it**: retry, use a fallback, or choose a default.
3. **Log it** and carry on (only at a boundary that can safely continue, such as a request handler).
4. **Stop**: terminate the program (`log.Fatal`) in `main` for unrecoverable startup failures.

Do **not** both log and return: the same failure then appears several times in the logs.

```go run
package main

import (
	"errors"
	"fmt"
)

var ErrTemporary = errors.New("temporary failure")

func flaky(attempt int) error {
	if attempt < 3 {
		return fmt.Errorf("attempt %d: %w", attempt, ErrTemporary)
	}
	return nil
}

func withRetry(max int) error {
	var err error
	for i := 1; i <= max; i++ {
		err = flaky(i)
		if err == nil {
			fmt.Println("succeeded on attempt", i)
			return nil
		}
		if !errors.Is(err, ErrTemporary) {
			return err // permanent: do not retry
		}
		fmt.Println("retrying after:", err)
	}
	return fmt.Errorf("giving up after %d attempts: %w", max, err)
}

func main() {
	fmt.Println(withRetry(5))
	fmt.Println(withRetry(2))
}
```

This shows `errors.Is` doing real work: the retry logic depends on the *kind* of error, not its text.

> **Common mistake:** deciding what to do by matching `err.Error()` strings, e.g. `strings.Contains(err.Error(), "not found")`. Messages are for humans and change freely. Use sentinels with `errors.Is` or types with `errors.As`.

## Deferred cleanup errors

A `defer f.Close()` hides the error from `Close`. For files you write to, `Close` can report real failures. Capture it with a named result:

```go
func write(path string, data []byte) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()

	_, err = f.Write(data)
	return err
}
```

`errors.Join(err, nil)` simply gives `err`, so there is no extra noise on the success path.

## A checklist for good error handling

- Check every error; ignore explicitly and rarely.
- Add context that the caller does not already know (the operation and key inputs).
- Preserve causes with `%w` when callers may need them.
- Compare with `errors.Is` and `errors.As`.
- Handle once: return or log, not both.
- Export sentinels and types only if callers truly need them.

## Key takeaways

- `errors.Join` merges several errors; `Is`/`As` work on each member and an all-nil input gives `nil`.
- Choose deliberately: return, handle, log or stop.
- Never branch on error message text.
- Capture errors from deferred calls (like `Close`) with named results when they matter.
