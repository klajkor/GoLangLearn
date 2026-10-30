---
title: Slices, Maps and Channels
minutes: 9
---
People often say that slices, maps and channels are "reference types" in Go. The precise story is more useful: each is a small **header value** that contains a pointer to data stored elsewhere. Go still passes the header **by value**, but copies of the header share the same underlying data.

## Maps and channels

A map value is a pointer-like handle to a hash table the runtime manages. A channel value is likewise a handle. Copying either gives another handle to the same thing.

```go run
package main

import "fmt"

func addKey(m map[string]int) {
	m["added"] = 1 // visible to the caller
}

func replace(m map[string]int) {
	m = map[string]int{"new": 1} // only rebinds the local variable
}

func main() {
	m := map[string]int{"a": 1}
	addKey(m)
	replace(m)
	fmt.Println(m) // map[a:1 added:1]

	ch := make(chan int, 1)
	send := func(c chan int) { c <- 42 }
	send(ch)
	fmt.Println(<-ch)

	var nilMap map[string]int
	fmt.Println(nilMap == nil, len(nilMap), nilMap["x"]) // true 0 0
}
```

A nil map can be read (you get zero values) but writing to it panics with `assignment to entry in nil map`. Always create maps with `make` or a literal before writing.

## Slices: pointer, length, capacity

A slice header has three fields:

```text
slice header:  [ ptr | len | cap ]
                  |
                  v
backing array: [ 1 | 2 | 3 | 4 | 5 ]
```

Passing a slice copies those three words. The function sees the same backing array, so element writes are visible to the caller. But `len` and `cap` are copied: an `append` inside the function changes only the function's header unless you return it.

```go run
package main

import "fmt"

func setFirst(s []int) {
	s[0] = 99 // shared backing array: caller sees this
}

func appendOne(s []int) {
	s = append(s, 4) // caller's len is unchanged
}

func appendReturn(s []int) []int {
	return append(s, 4)
}

func main() {
	s := make([]int, 3, 10)
	setFirst(s)
	appendOne(s)
	fmt.Println(s, len(s)) // [99 0 0] 3

	s = appendReturn(s)
	fmt.Println(s, len(s)) // [99 0 0 4] 4
}
```

Notice `appendOne` did write `4` into the shared array (there was spare capacity), but the caller's `len` is still 3, so it cannot see it. That is why you always write `s = append(s, ...)`.

## Aliasing gotchas

Because slices can share memory, two slices may overlap unexpectedly.

```go run
package main

import "fmt"

func main() {
	a := []int{1, 2, 3, 4, 5}
	b := a[1:3] // [2 3], shares a's array, cap 4

	b = append(b, 100) // fits in capacity: overwrites a[3]
	fmt.Println(a)     // [1 2 3 100 5]
	fmt.Println(b)

	c := make([]int, len(a))
	copy(c, a) // independent copy
	c[0] = -1
	fmt.Println(a[0], c[0]) // 1 -1
}
```

Appending to a sub-slice can overwrite elements of the original. When you need independence, use `copy`, `slices.Clone`, or the full slice expression `a[1:3:3]`, which limits capacity so `append` must reallocate.

> **Common mistake:** keeping a tiny sub-slice of a huge slice, which keeps the whole backing array alive for the garbage collector. Copy what you need if the original is large.

## Arrays are values

Unlike slices, **arrays** are plain values. Passing `[3]int` copies all three elements, and `==` compares them element by element. If you really want to share an array, pass a pointer `*[3]int` or a slice of it.

## Summary table

| Type | What a copy shares | Zero value |
|---|---|---|
| pointer | the pointee | `nil` |
| slice | the backing array (not len/cap) | `nil` (usable with `append`) |
| map | the hash table | `nil` (read-only) |
| channel | the channel | `nil` (blocks forever) |
| array, struct | nothing (full copy) | all zero |
| `func` | the captured variables of the closure | `nil` |

## Key takeaways

- Slices, maps and channels are small headers passed by value, but they point at shared data.
- Element writes through a slice parameter are visible to the caller; `append` results must be assigned back.
- Writing to a nil map panics; appending to a nil slice is fine.
- Sub-slices share memory, so `append` can overwrite the original; use `copy`, `slices.Clone` or a three-index slice when you need independence.
- Arrays and structs are copied in full when assigned or passed.
