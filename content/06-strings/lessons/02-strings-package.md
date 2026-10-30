---
title: The strings Package
minutes: 8
---
Most text processing in Go uses functions from the standard `strings` package. They are plain functions (not methods), and since strings are immutable they always **return a new string** rather than changing the original.

## Searching

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "the quick brown fox"

	fmt.Println(strings.Contains(s, "quick"))
	fmt.Println(strings.HasPrefix(s, "the"), strings.HasSuffix(s, "fox"))
	fmt.Println(strings.Index(s, "brown"))
	fmt.Println(strings.Index(s, "cat"))
	fmt.Println(strings.Count("banana", "a"))
}
```

`Index` returns the byte offset of the first match, or `-1` if absent.

## Splitting and joining

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	csv := "a,b,,c"
	parts := strings.Split(csv, ",")
	fmt.Println(len(parts), parts)
	fmt.Printf("%q\n", parts)

	text := "  many   spaces\there\n"
	fmt.Printf("%q\n", strings.Fields(text))

	fmt.Println(strings.Join([]string{"x", "y", "z"}, " | "))
}
```

`Split` keeps empty pieces (note the empty string between the two commas). `Fields` splits around any run of whitespace and never returns empty strings, which is usually what you want for words.

## Transforming

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.ToUpper("go"), strings.ToLower("GO"))
	fmt.Printf("%q\n", strings.TrimSpace("  padded \n"))
	fmt.Println(strings.Trim("xxhixx", "x"))
	fmt.Println(strings.TrimPrefix("prefix-name", "prefix-"))
	fmt.Println(strings.TrimSuffix("file.go", ".go"))
	fmt.Println(strings.Replace("a-b-c-d", "-", "+", 2))
	fmt.Println(strings.ReplaceAll("a-b-c-d", "-", "+"))
	fmt.Println(strings.Repeat("ab", 3))
}
```

> **Common mistake:** calling `strings.ToUpper(s)` and ignoring the result. The original `s` is unchanged because strings are immutable.

`Trim` takes a **set of characters** to remove from both ends, whereas `TrimPrefix` and `TrimSuffix` remove an exact string.

## Cut: split once

`strings.Cut` (Go 1.18+) splits around the first occurrence of a separator and reports whether it was found. It is perfect for `key=value` parsing.

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	line := "name=Gopher=extra"

	key, value, found := strings.Cut(line, "=")
	fmt.Println(key, value, found)

	before, after, ok := strings.Cut("novalue", "=")
	fmt.Printf("%q %q %v\n", before, after, ok)
}
```

Only the first `=` splits, so `value` is `Gopher=extra`. When the separator is missing, `before` is the whole string, `after` is empty and `found` is `false`.

## Building strings efficiently

Repeated `s += "x"` in a loop copies the string every time. `strings.Builder` appends into a growing buffer.

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	var sb strings.Builder
	for i := 1; i <= 5; i++ {
		fmt.Fprintf(&sb, "%d,", i)
	}
	sb.WriteString("done")
	sb.WriteByte('!')
	sb.WriteRune('✓')

	fmt.Println(sb.String(), sb.Len())
}
```

Declare a `strings.Builder` with `var`, write to it, and call `String()` at the end. `fmt.Fprintf` writes formatted text to any writer, including a Builder (pass its address).

## Key takeaways

- `strings` functions return new strings; they never modify the input.
- `Contains`, `HasPrefix`, `HasSuffix`, `Index`, `Count` search; `Index` gives `-1` when absent.
- `Split` keeps empty parts, `Fields` splits on whitespace runs; `Join` is the inverse of `Split`.
- `TrimSpace`, `Trim`, `TrimPrefix`, `TrimSuffix`, `Replace`, `ReplaceAll` transform text.
- `Cut` splits once around a separator; `strings.Builder` builds big strings efficiently.
