---
title: Constants and iota
minutes: 8
---
A constant is a value fixed at compile time. Declare one with `const`; trying to assign to it later is a compile error.

```go
const Pi = 3.14159
const Greeting = "hello"

const (
	MaxUsers = 100
	Debug    = false
)
```

Constants can only be numbers, strings, booleans, or runes, and their values must be computable at compile time. You cannot make a constant slice, map, or the result of a normal function call.

## Typed and untyped constants

This is the subtle and useful part. A constant without a type is **untyped**: it has a kind (integer, floating-point, rune, string, bool) but no fixed type. An untyped constant adapts to whatever context it is used in, as long as the value fits.

```go run
package main

import "fmt"

const Big = 1 << 40 // untyped integer constant, far bigger than int32

const Typed int32 = 7

func main() {
	var f float64 = Big // fine: untyped constant converts
	var i int64 = Big
	fmt.Println(f, i)

	var x float64 = 2
	fmt.Println(x * 3) // 3 becomes float64 here

	var t int32 = Typed
	fmt.Println(t)
	// var u int64 = Typed // compile error: Typed is an int32, not untyped

	d := 5 / 2   // both untyped integers: integer division
	e := 5 / 2.0 // 2.0 is an untyped float, so the result is a float constant
	fmt.Println(d, e)
}
```

Untyped constants are also computed with arbitrary precision at compile time. That is why `Big` is fine even though it would overflow a 32-bit integer: it is only checked when assigned to a real type.

A **typed** constant (`const Typed int32 = 7`) behaves like a variable of that type: it will not mix with other types without conversion. Prefer untyped constants for plain numbers, and typed constants when the type is part of the meaning (as with enums below).

> **Common mistake:** `var u uint8 = 300` is a compile error ("constant 300 overflows uint8"), unlike a runtime value, which would wrap. Constant expressions are checked strictly.

## iota: counting constants

Inside a `const ( ... )` block, `iota` is an integer that starts at 0 and increases by 1 for each line (each constant specification). If you omit the expression on later lines, the previous expression is repeated. That makes enumerations concise:

```go
const (
	Red   = iota // 0
	Green        // 1
	Blue         // 2
)
```

Each new `const` block restarts `iota` at 0. Combine it with expressions and `_` to skip values:

```go run
package main

import "fmt"

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
)

const (
	_  = iota             // skip 0
	KB = 1 << (10 * iota) // 1 << 10
	MB                    // 1 << 20
	GB                    // 1 << 30
)

type Perm uint8

const (
	Read Perm = 1 << iota // 1
	Write                 // 2
	Exec                  // 4
)

func main() {
	fmt.Println(Sunday, Monday, Tuesday)
	fmt.Println(KB, MB, GB)

	p := Read | Exec // combine bit flags
	fmt.Println(p, p&Write != 0, p&Exec != 0)
}
```

Look at what happened:

- `Weekday` is a *typed* enum: `Monday` is a `Weekday`, so you cannot accidentally add it to an unrelated `int` without a conversion.
- `KB`, `MB`, `GB` show the "skip zero, then scale" idiom.
- `Read`, `Write`, `Exec` are bit flags: each is a different power of two, so they can be combined with `|` and tested with `&`.

Go has no `enum` keyword; a named type plus an `iota` block is the idiom. Methods on the type (such as a `String()` method) can later print names instead of numbers.

## const vs var

| | `const` | `var` |
|---|---|---|
| Value known at | compile time | run time |
| Can change | no | yes |
| Allowed types | numbers, strings, bools | any |
| Untyped form | yes | no |

> **Tip:** Constants have no address, so you cannot take `&MaxUsers`. If you need a pointer, copy it into a variable first.

## Key takeaways

- `const` values are fixed at compile time and limited to basic types.
- Untyped constants adapt to their context and have arbitrary precision; typed constants do not mix freely.
- `iota` counts lines within one `const` block, restarting at 0 per block; omitted expressions repeat the previous one.
- Use `_` to skip values, and shifts such as `1 << iota` for bit flags and size units.
- Typed `iota` constants are Go's idiomatic enums.
