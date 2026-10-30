---
title: "unsafe: Sizes, Padding and Pointers"
minutes: 11
---
The `unsafe` package lets you step around Go's type system. Most programs never need it. But one part is useful even to ordinary developers: measuring **struct layout** and understanding why field order matters. Everything below assumes a **64-bit platform** (such as amd64 or arm64).

> **Warning:** code using `unsafe` is not covered by the Go 1 compatibility promise, can corrupt memory instead of failing with a clean panic, and may break between Go versions. Do not use it to "optimize" unless a profile proves you need to.

## Sizeof, Alignof, Offsetof

| Function | Meaning |
|---|---|
| `unsafe.Sizeof(x)` | bytes occupied by the value itself (not what it points to) |
| `unsafe.Alignof(x)` | the alignment the type requires |
| `unsafe.Offsetof(s.f)` | byte offset of field `f` inside struct `s` |

All three are evaluated at compile time and are safe to call.

```go run
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	var (
		b  bool
		i  int
		s  string
		sl []int
		m  map[string]int
		p  *int
		a  any
	)
	fmt.Println("bool:", unsafe.Sizeof(b))
	fmt.Println("int:", unsafe.Sizeof(i))
	fmt.Println("string:", unsafe.Sizeof(s))
	fmt.Println("slice:", unsafe.Sizeof(sl))
	fmt.Println("map:", unsafe.Sizeof(m))
	fmt.Println("pointer:", unsafe.Sizeof(p))
	fmt.Println("interface:", unsafe.Sizeof(a))
}
```

On 64-bit: a string is 16 bytes (pointer + length), a slice 24 (pointer + length + capacity), an interface 16, and maps and pointers 8. These sizes never include the pointed-to data.

## Padding and field order

The CPU wants each value at an address that is a multiple of its alignment. The compiler inserts **padding** bytes to guarantee that, and rounds the struct's total size up to a multiple of its largest alignment. So field order changes the size:

```go run
package main

import (
	"fmt"
	"unsafe"
)

type Bad struct {
	A bool  // 1 byte, then 7 padding
	B int64 // 8
	C bool  // 1 byte, then 7 trailing padding
}

type Good struct {
	B int64 // 8
	A bool  // 1
	C bool  // 1, then 6 trailing padding
}

func main() {
	fmt.Println("Bad: ", unsafe.Sizeof(Bad{}), unsafe.Alignof(Bad{}))
	fmt.Println("Good:", unsafe.Sizeof(Good{}), unsafe.Alignof(Good{}))
	fmt.Println("Bad offsets: ", unsafe.Offsetof(Bad{}.A), unsafe.Offsetof(Bad{}.B), unsafe.Offsetof(Bad{}.C))
	fmt.Println("Good offsets:", unsafe.Offsetof(Good{}.B), unsafe.Offsetof(Good{}.A), unsafe.Offsetof(Good{}.C))
}
```

Sorting fields from largest to smallest alignment usually minimizes padding. It matters when you hold millions of values; otherwise prefer the order that reads best. The `fieldalignment` analyzer from `golang.org/x/tools` can suggest better orderings. Also remember that the zero-size type `struct{}` has size 0, which is why `map[string]struct{}` is the idiomatic set.

## unsafe.Pointer rules

`unsafe.Pointer` is a pointer that can convert to and from any pointer type and to `uintptr`. The language allows only a short list of valid patterns (documented on the `unsafe` package page). The most common is reinterpreting memory of the same layout:

```go
f := 1.5
bits := *(*uint64)(unsafe.Pointer(&f)) // reinterpret float64 bits
```

(In real code use `math.Float64bits`.) The big rules:

- A `uintptr` is just an integer. The garbage collector does **not** treat it as a reference, so the object may be freed (or its stack moved) while you hold it. Never keep a `uintptr` computed from a pointer across statements.
- Pointer arithmetic is only valid within the same allocated object, using `unsafe.Add` (Go 1.17).
- Never cast between types of different sizes or layouts and then read beyond the real object.

## unsafe.String and unsafe.Slice

`unsafe.Slice` (Go 1.17) and `unsafe.String` (Go 1.20) build a slice or string around existing memory without copying:

```go
b := []byte("hello")
s := unsafe.String(&b[0], len(b)) // shares memory with b
b[0] = 'J'                        // s now reads "Jello": strings must be immutable!
```

That example shows the danger: Go assumes strings are immutable, so the compiler and runtime may cache or share them. The trick is only safe if the bytes are never modified afterward. Prefer a normal conversion and accept the copy unless profiling says otherwise.

## cgo, briefly

`import "C"` lets Go call C libraries. It requires a C toolchain, slows builds, complicates cross-compiling, and each Go-to-C call costs far more than a Go call. Use it only when you must link to an existing C library; look for a pure-Go port first.

## Key takeaways

- `Sizeof`, `Alignof` and `Offsetof` are safe, compile-time functions; sizes here assume 64-bit.
- Field order affects padding; largest-first usually minimizes size.
- `unsafe.Pointer` and `unsafe.String`/`Slice` bypass type safety; follow the documented patterns exactly or avoid them.
- Never store a pointer as a `uintptr`.
- cgo is a last resort.
