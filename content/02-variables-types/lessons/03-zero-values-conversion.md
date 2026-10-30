---
title: Zero Values and Type Conversion
minutes: 7
---
Two ideas shape how Go code feels: variables are never uninitialised, and types never convert behind your back.

## Zero values

A variable declared without an initial value gets the **zero value** of its type. There is no "undefined" and no garbage memory.

| Type | Zero value |
|---|---|
| numbers (`int`, `float64`, ...) | `0` |
| `bool` | `false` |
| `string` | `""` (empty string) |
| pointers, slices, maps, channels, functions, interfaces | `nil` |

Good Go code often designs types so that the zero value is already useful, so you can declare a variable and use it without a constructor.

```go run
package main

import "fmt"

func main() {
	var i int
	var f float64
	var b bool
	var s string
	var p *int
	var sl []int
	var m map[string]int

	fmt.Printf("%v %v %v %q %v %v %v\n", i, f, b, s, p, sl, m)
	fmt.Println(p == nil, sl == nil, m == nil, len(sl), len(m))

	fmt.Println(m["missing"]) // reading a nil map is fine
}
```

The `%q` verb prints a string with quotes, which makes the empty string visible. Reading from a nil map gives the zero value of the element type, and `len` of a nil slice or map is 0. (Writing to a nil map panics; maps get a proper lesson later.)

## No implicit conversion

In many languages `int + float` quietly converts. Go refuses: mixing types is a compile error, even between `int` and `int64`.

```go
var a int = 5
var b int64 = 10
fmt.Println(a + b) // compile error: mismatched types int and int64
```

You convert explicitly with the type name used like a function, `T(value)`:

```go run
package main

import "fmt"

func main() {
	var a int = 5
	var b int64 = 10
	fmt.Println(int64(a) + b)

	price := 19.99
	whole := int(price) // truncates, does not round
	fmt.Println(whole)

	n := 7
	avg := float64(n) / 2
	fmt.Println(avg)

	big := 300
	fmt.Println(uint8(big)) // value does not fit: wraps to 300 mod 256
}
```

Be aware of what conversions do:

- Float to integer **truncates** toward zero (`int(-2.9)` is `-2`).
- Narrowing integer conversions keep the low bits, so values that do not fit wrap silently.
- The explicit conversion is the point: you decide, and a reader can see it.

## Converting to and from strings

Numbers and strings are different kinds of thing, so a plain conversion does something you might not expect:

```go
n := 65
fmt.Println(string(rune(n))) // "A": the character with code point 65
```

`string(65)` treats the number as a Unicode code point, not as digits; `go vet` warns about `string(int)` for exactly that reason. To get the digits, use the `strconv` package:

```go run
package main

import (
	"fmt"
	"strconv"
)

func main() {
	s := strconv.Itoa(65) // int to decimal string
	fmt.Println(s + "!")

	n, err := strconv.Atoi("123")
	fmt.Println(n+1, err)

	_, err = strconv.Atoi("12a")
	fmt.Println(err)

	f, _ := strconv.ParseFloat("2.5", 64)
	fmt.Println(f * 2)

	fmt.Println(string([]byte{72, 105}), []byte("Hi"))
}
```

Parsing returns an error as a second value, because user text may not be a number. Always check it in real code. You can also convert between `string`, `[]byte` and `[]rune`, as the last line shows.

> **Tip:** `fmt.Sprintf("%d", n)` also builds a string from a number. `strconv.Itoa` is simpler and faster for the plain-integer case.

## Named types need conversions too

A type defined with `type Celsius float64` is a different type from `float64`, even though they share a representation. You must write `Celsius(x)` or `float64(c)` to move between them. This is how Go helps stop you adding metres to seconds. You will meet defined types properly in a later chapter.

## Key takeaways

- Every variable has a zero value (`0`, `false`, `""`, `nil`), so there are no uninitialised variables.
- Go never converts between types implicitly; write `T(x)`.
- Float to int truncates; narrowing integer conversions wrap.
- `string(n)` gives a character, not digits; use `strconv.Itoa` and `strconv.Atoi` for numbers.
- Parsing functions return an `error` that you should check.
