---
title: "Slices: Length, Capacity and append"
minutes: 9
---
A **slice** is a flexible view onto an array. It is the Go collection you will use every day. Internally a slice is a tiny struct with three fields: a pointer to the first element, a **length** and a **capacity**.

## Creating slices

```go run
package main

import "fmt"

func main() {
	a := []int{1, 2, 3}         // literal
	b := make([]int, 3)         // len 3, cap 3, zeros
	c := make([]int, 0, 10)     // len 0, cap 10
	var d []int                 // nil slice

	fmt.Println(a, len(a), cap(a))
	fmt.Println(b, len(b), cap(b))
	fmt.Println(c, len(c), cap(c))
	fmt.Println(d, len(d), cap(d), d == nil)
}
```

- **len** is how many elements the slice currently holds.
- **cap** is how many elements fit before Go must allocate a bigger array.

## Growing with append

`append` adds elements and **returns the new slice**. You must use its result.

```go run
package main

import "fmt"

func main() {
	var s []int
	for i := 1; i <= 5; i++ {
		s = append(s, i*i)
		fmt.Println(len(s), s)
	}

	s = append(s, 100, 200)
	more := []int{7, 8}
	s = append(s, more...)
	fmt.Println(s)
}
```

When the capacity is exhausted, `append` allocates a larger array (roughly doubling for small slices), copies the elements and returns a slice pointing at the new array. Do not rely on exact growth numbers; they are an implementation detail.

> **Common mistake:** calling `append(s, 4)` and ignoring the result. The original `s` is unchanged. Always write `s = append(s, 4)`.

## Nil slice vs empty slice

A nil slice has no backing array; an empty slice does (of length 0). Both have `len` 0, you can `append` to both, and `range` over both does nothing.

```go run
package main

import "fmt"

func main() {
	var nilS []int
	emptyS := []int{}

	fmt.Println(nilS == nil, emptyS == nil)
	fmt.Println(len(nilS), len(emptyS))

	nilS = append(nilS, 1)
	fmt.Println(nilS)
}
```

The difference only matters in a few places, for example JSON encoding (`null` vs `[]`). In your own code, `len(s) == 0` is the check that works for both.

## Preallocating

If you know roughly how many items you will add, give `make` a capacity. This avoids repeated reallocations.

```go
result := make([]string, 0, len(input))
for _, v := range input {
	result = append(result, strings.ToUpper(v))
}
```

## Removing and inserting

Go has no built-in `remove`. The `slices` package (next lessons) provides `slices.Delete` and `slices.Insert`; the classic manual way to delete index `i` while keeping order is:

```go
s = append(s[:i], s[i+1:]...)
```

## Key takeaways

- A slice is (pointer, length, capacity) describing part of an array.
- `make([]T, len, cap)` creates one; `var s []T` gives a nil slice, which is usable.
- `append` returns a new slice; always assign the result.
- Use `len(s) == 0` to test for emptiness.
- Preallocate capacity when you know the size.
