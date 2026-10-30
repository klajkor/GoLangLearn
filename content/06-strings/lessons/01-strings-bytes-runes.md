---
title: Strings, Bytes and Runes
minutes: 9
---
A Go `string` is an **immutable sequence of bytes**, conventionally holding UTF-8 encoded text. Understanding that one sentence prevents most string bugs.

## Bytes versus characters

UTF-8 encodes each Unicode character (a *code point*) in one to four bytes. ASCII letters take one byte; many other characters take more. `len(s)` counts **bytes**, not characters.

```go run
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "héllo, 世界"

	fmt.Println(len(s))                    // bytes
	fmt.Println(utf8.RuneCountInString(s)) // characters
	fmt.Println(len("a"), len("é"), len("世"))
}
```

`é` takes 2 bytes and `世` takes 3, so the string above has 9 characters but 14 bytes.

## byte and rune

| Type | Alias for | Represents |
|---|---|---|
| `byte` | `uint8` | one byte of data |
| `rune` | `int32` | one Unicode code point |

A rune literal uses single quotes: `'a'`, `'世'`. A string literal uses double quotes (or backticks for raw strings spanning lines without escapes).

## Indexing gives bytes, range gives runes

`s[i]` returns the **byte** at position `i`. `for i, r := range s` decodes UTF-8 and yields **runes**, along with the byte offset where each starts.

```go run
package main

import "fmt"

func main() {
	s := "aé世"

	for i := 0; i < len(s); i++ {
		fmt.Print(s[i], " ")
	}
	fmt.Println()

	for i, r := range s {
		fmt.Printf("%d:%c(%d) ", i, r, r)
	}
	fmt.Println()
}
```

The first loop prints the six raw bytes (`97 195 169 228 184 150`). The second prints three runes at byte offsets 0, 1 and 3.

> **Common mistake:** looping with `for i := 0; i < len(s); i++` and treating `s[i]` as a character. That breaks on any non-ASCII text. Use `range`, or convert to `[]rune`.

## Immutability and conversions

You cannot assign to `s[0]`. To modify text, convert to a byte or rune slice, change it, and convert back (which copies).

```go run
package main

import "fmt"

func main() {
	s := "héllo"

	r := []rune(s)
	r[0] = 'H'
	fmt.Println(string(r), len(r))

	b := []byte("go")
	b[0] = 'G'
	fmt.Println(string(b))

	// Reverse by runes, not bytes
	word := "añb"
	rs := []rune(word)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	fmt.Println(string(rs))
}
```

Slicing a string with `s[a:b]` works on **byte** offsets and is cheap (no copy), but it can cut a multi-byte character in half. Only slice at positions you know are character boundaries.

> **Tip:** `string(65)` is not `"65"`; it converts a number to the character with that code point (`"A"`). `go vet` warns about it for ints. Use `strconv.Itoa` for digits (next lessons).

## Comparing and concatenating

Strings compare with `==`, `<` and friends, byte by byte. `+` concatenates and allocates a new string, so for building large strings in a loop use `strings.Builder` (coming up).

## Key takeaways

- A string is an immutable byte sequence, usually UTF-8.
- `len(s)` is bytes; `utf8.RuneCountInString(s)` is characters.
- `s[i]` is a `byte`; `range s` yields `rune`s with byte offsets.
- `byte` is `uint8`, `rune` is `int32`; convert with `[]byte(s)` and `[]rune(s)`.
- Slice strings only at character boundaries.
