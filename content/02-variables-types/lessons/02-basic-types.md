---
title: Basic Types
minutes: 9
---
Go's built-in types are few and precise. Knowing exactly how big each is saves you from surprises.

## Integers

| Type | Size | Range |
|---|---|---|
| `int8` / `uint8` | 8 bits | -128..127 / 0..255 |
| `int16` / `uint16` | 16 bits | -32768..32767 / 0..65535 |
| `int32` / `uint32` | 32 bits | about +-2.1 billion / 0..4.29 billion |
| `int64` / `uint64` | 64 bits | about +-9.2e18 / 0..1.8e19 |
| `int` / `uint` | 32 or 64 bits, depending on the platform | usually 64 bits today |

Use plain `int` unless you have a reason not to: it is what slice indexes, `len`, and untyped integer constants use. Pick a sized type when you care about layout (binary formats, large arrays) or need specific overflow behaviour.

Integers wrap around silently when they overflow; there is no exception:

```go run
package main

import "fmt"

func main() {
	var small uint8 = 250
	small += 10
	fmt.Println(small) // wraps around modulo 256

	var i int8 = 127
	i++
	fmt.Println(i)

	fmt.Println(7/2, -7/2, 7%3, -7%3) // integer division truncates toward zero
	fmt.Println(7.0/2, float64(7)/2)
}
```

Integer division truncates toward zero, and the remainder takes the sign of the dividend. `7.0/2` is a constant expression with a float operand, so it gives `3.5`.

## Floats and complex numbers

`float64` (the default for decimal literals) and `float32` follow IEEE 754. They cannot represent every decimal exactly, so never compare computed floats with `==` when exactness matters, and never use them for money (use integer cents instead).

```go
fmt.Println(0.1 + 0.2) // might surprise you; try it
```

There are also `complex64` and `complex128`, rarely needed.

## Booleans

`bool` has two values, `true` and `false`. Unlike C or JavaScript, there is no truthiness: `if 1 {` and `if name {` do not compile. A condition must be a real `bool`. Operators: `&&`, `||`, `!`, with short-circuit evaluation.

## Strings, bytes and runes

A `string` is an immutable sequence of **bytes**, conventionally UTF-8 text. `len(s)` counts bytes, not characters. Two aliases matter:

| Alias | Underlying | Represents |
|---|---|---|
| `byte` | `uint8` | one byte |
| `rune` | `int32` | one Unicode code point |

Indexing a string gives a byte; ranging over it decodes runes. A character literal like `'a'` is a rune constant; a double-quoted `"a"` is a string.

```go run
package main

import "fmt"

func main() {
	s := "héllo"
	fmt.Println(len(s), len([]rune(s))) // bytes vs. characters

	fmt.Println(s[0], string(s[0])) // a byte: 104 is 'h'

	for i, r := range s {
		fmt.Println(i, r, string(r))
	}

	raw := `C:\path\no escapes`
	fmt.Println(raw)
	fmt.Println("tab:\there", "quote:\"")
}
```

Notice that `é` takes two bytes, so the byte index jumps from 1 to 3 in the loop, and that `len(s)` is 6 although there are only 5 characters. Backtick strings are *raw*: no escape sequences are processed.

> **Common mistake:** assuming `len("héllo")` is the character count. Use `utf8.RuneCountInString(s)` or `len([]rune(s))` to count characters. Strings are immutable, so `s[0] = 'H'` does not compile.

## Choosing a type

- Whole numbers: `int`. Fractional: `float64`.
- Text: `string`. A single character: `rune`. Raw data: `byte` or `[]byte`.
- Yes/no: `bool`.
- Reach for `int64`, `uint32` and friends only when the size is part of the requirement.

## Key takeaways

- `int` is 32 or 64 bits depending on the platform; sized variants exist for exact widths.
- Integer overflow wraps silently; integer division truncates toward zero.
- `bool` is strictly `true` or `false`; there is no truthiness.
- A `string` is an immutable byte sequence; `len` counts bytes, `range` yields runes.
- `byte` is `uint8`, `rune` is `int32`.
