---
title: Formatting with fmt Verbs
minutes: 9
---
The `fmt` package prints and builds text using **verbs**: placeholders starting with `%`. You have used `Println`; now let's use `Printf`, which takes a format string.

## The Print family

| Function | Output goes to | Notes |
|---|---|---|
| `Println` | stdout | spaces between args, newline at end |
| `Printf` | stdout | format string, no automatic newline |
| `Sprintf` | returns a `string` | build text without printing |
| `Errorf` | returns an `error` | like `Sprintf`, but makes an error |
| `Fprintf` | any `io.Writer` | files, buffers, `os.Stderr` |

## Common verbs

```go run
package main

import "fmt"

type Point struct {
	X, Y int
}

func main() {
	p := Point{1, 2}
	name := "Gopher"

	fmt.Printf("%v\n", p)
	fmt.Printf("%+v\n", p)
	fmt.Printf("%#v\n", p)
	fmt.Printf("%T %T %T\n", p, name, 3.5)

	fmt.Printf("%d %s %q\n", 42, name, name)
	fmt.Printf("%x %X %b %o\n", 255, 255, 5, 8)
	fmt.Printf("%c %U\n", 'G', 'G')
	fmt.Printf("%t\n", true)
	fmt.Printf("%f %.2f %e\n", 3.14159, 3.14159, 1234.5)
	fmt.Printf("100%%\n")
}
```

- `%v` is the default format for any value. `%+v` adds field names to structs and `%#v` prints Go-syntax, handy for debugging.
- `%T` prints the **type**.
- `%d` integers, `%s` strings, `%q` quoted strings, `%x` hexadecimal (also works on strings and byte slices), `%c` a character, `%f` a float, `%t` a bool.
- `%%` prints a literal percent sign.

## Width, precision and flags

A number after `%` sets the minimum **width**; `.N` sets precision (digits after the point for floats).

```go run
package main

import "fmt"

func main() {
	fmt.Printf("[%8.2f]\n", 3.14159)
	fmt.Printf("[%-8.2f]\n", 3.14159)
	fmt.Printf("[%08.2f]\n", 3.14159)
	fmt.Printf("[%5d]\n", 42)
	fmt.Printf("[%-5d]\n", 42)
	fmt.Printf("[%05d]\n", 42)
	fmt.Printf("[%+d]\n", 42)
	fmt.Printf("[%8s]\n", "go")
	fmt.Printf("[%-8s]\n", "go")
	fmt.Printf("[%*d]\n", 6, 7)
}
```

| Flag | Effect |
|---|---|
| `-` | left-align (pad on the right) |
| `0` | pad numbers with zeros |
| `+` | always show the sign |
| `#` | alternate form (`%#x` gives `0xff`) |
| space | leave a space for a missing `+` sign |

A `*` takes the width from an argument, as in the last line.

## Aligned tables

Width and the `-` flag make neat columns.

```go run
package main

import "fmt"

func main() {
	items := []struct {
		Name  string
		Qty   int
		Price float64
	}{
		{"apple", 3, 0.5},
		{"watermelon", 1, 3.25},
		{"fig", 12, 0.2},
	}

	fmt.Printf("%-12s %4s %8s\n", "ITEM", "QTY", "PRICE")
	for _, it := range items {
		fmt.Printf("%-12s %4d %8.2f\n", it.Name, it.Qty, it.Price)
	}
}
```

> **Tip:** width counts **runes**, not bytes, for `%s`, so accented text still lines up in most cases (wide East Asian characters may not).

## Sprintf and Errorf

`Sprintf` returns the string. `Errorf` creates an error value with a formatted message, and `%w` wraps another error so it can be inspected later.

```go run
package main

import (
	"errors"
	"fmt"
)

var ErrEmpty = errors.New("empty name")

func validate(name string) error {
	if name == "" {
		return fmt.Errorf("validate %q: %w", name, ErrEmpty)
	}
	return nil
}

func main() {
	label := fmt.Sprintf("user-%03d", 7)
	fmt.Println(label)

	err := validate("")
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrEmpty))
}
```

> **Common mistake:** a mismatched verb and argument, such as `Printf("%d", "text")`, does not fail at compile time. It prints `%!d(string=text)`. `go vet` catches these, so run it often.

## Key takeaways

- `%v`, `%+v`, `%#v` and `%T` show values and types; `%d`, `%s`, `%q`, `%x`, `%f` format specific kinds.
- `%8.2f` means width 8, two decimals; `-` left-aligns, `0` zero-pads, `+` forces a sign.
- `Sprintf` builds a string; `Errorf` builds an error and `%w` wraps one.
- Use width with `Printf` to align columns.
- Run `go vet` to catch verb and argument mismatches.
