---
title: Converting with strconv
minutes: 6
---
Go never converts between strings and numbers implicitly, and `string(n)` does not do what you might expect. The `strconv` package ("string conversion") is the right tool.

## Numbers to strings

`strconv.Itoa` ("integer to ASCII") turns an `int` into its decimal text.

```go run
package main

import (
	"fmt"
	"strconv"
)

func main() {
	n := 42
	s := strconv.Itoa(n)
	fmt.Println(s+"!", len(s))

	fmt.Println(strconv.FormatInt(255, 2))  // binary
	fmt.Println(strconv.FormatInt(255, 16)) // hex
	fmt.Println(strconv.FormatFloat(3.14159, 'f', 2, 64))
	fmt.Println(strconv.FormatBool(true))
}
```

## Strings to numbers

Parsing can fail, so these functions return `(value, error)`: exactly the idiom from the Functions chapter.

```go run
package main

import (
	"fmt"
	"strconv"
)

func main() {
	n, err := strconv.Atoi("123")
	fmt.Println(n+1, err)

	_, err = strconv.Atoi("12a")
	fmt.Println(err)

	f, err := strconv.ParseFloat("2.5", 64)
	fmt.Println(f*2, err)

	b, err := strconv.ParseBool("true")
	fmt.Println(b, err)

	big, err := strconv.ParseInt("ff", 16, 64)
	fmt.Println(big, err)
}
```

`Atoi` is shorthand for `ParseInt(s, 10, 0)` returning an `int`. The error message tells you the function, the input and the reason (for example `strconv.Atoi: parsing "12a": invalid syntax`).

> **Common mistake:** ignoring the error with `n, _ := strconv.Atoi(input)` on user input. A failed parse returns `0`, which looks like a valid value. Check `err`.

`ParseInt(s, base, bitSize)` takes a base (0 means "detect from prefix such as `0x`") and the integer size in bits it must fit into; out-of-range values produce an error.

## Quoting

`strconv.Quote` returns a double-quoted Go string literal with escapes, which is useful for debugging invisible characters. `Unquote` reverses it.

```go run
package main

import (
	"fmt"
	"strconv"
)

func main() {
	raw := "tab\there \"quoted\"\n"
	q := strconv.Quote(raw)
	fmt.Println(q)

	back, err := strconv.Unquote(q)
	fmt.Println(back == raw, err)
}
```

`fmt.Printf("%q", s)` produces the same result as `Quote`.

## Choosing the right tool

| Task | Use |
|---|---|
| `int` to text | `strconv.Itoa(n)` |
| text to `int` | `strconv.Atoi(s)` |
| text to `float64` | `strconv.ParseFloat(s, 64)` |
| number to text with a format | `fmt.Sprintf` or `strconv.FormatFloat` |
| code point to text | `string(rune(n))` |

## Key takeaways

- `Itoa` converts int to string; `Atoi` converts string to int and returns an error.
- `ParseInt`, `ParseFloat`, `ParseBool` cover other types and bases.
- Always check the error from a parse; the value is not trustworthy when it fails.
- `strconv.Quote` shows a string with escapes made visible.
- `string(65)` yields `"A"`, not `"65"`.
