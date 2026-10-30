---
title: Slicing, Sharing and copy
minutes: 9
---
The slice expression `s[low:high]` creates a new slice that **shares the same backing array** as `s`. This is cheap and powerful, and the source of the most common slice surprises.

## The slice expression

`s[low:high]` includes index `low` and excludes `high`. Both are optional: `s[:3]`, `s[2:]`, `s[:]`.

```go run
package main

import "fmt"

func main() {
	s := []int{0, 1, 2, 3, 4, 5}

	fmt.Println(s[1:4])
	fmt.Println(s[:2])
	fmt.Println(s[4:])

	t := s[1:4]
	fmt.Println(len(t), cap(t))
}
```

`t` has length 3 (indexes 1, 2, 3 of `s`) and capacity 5: it can grow until the end of the original array.

## Shared memory

Because the arrays are shared, writing through one slice is visible through the other.

```go run
package main

import "fmt"

func main() {
	s := []int{1, 2, 3, 4}
	t := s[1:3]

	t[0] = 99
	fmt.Println(s, t)

	// Appending within capacity overwrites the shared array!
	t = append(t, 77)
	fmt.Println(s, t)
}
```

The first print shows `[1 99 3 4] [99 3]`. Then `t` still had room (capacity 3), so `append` wrote `77` into the shared array, overwriting `s[3]`: `[1 99 3 77] [99 3 77]`.

> **Common mistake:** keeping a small sub-slice of a huge slice. The whole big array stays in memory. Copy what you need instead.

## The full slice expression

`s[low:high:max]` also limits the **capacity** to `max - low`. This forces a later `append` to allocate a new array instead of overwriting the original.

```go run
package main

import "fmt"

func main() {
	s := []int{1, 2, 3, 4}
	t := s[1:3:3] // len 2, cap 2

	t = append(t, 77) // must reallocate
	fmt.Println(s, t)
	fmt.Println(len(t), cap(t) >= 3)
}
```

Now `s` stays `[1 2 3 4]`, and `t` is `[2 3 77]`.

## copy: independent duplicates

`copy(dst, src)` copies `min(len(dst), len(src))` elements and returns how many it copied. It is the way to get a slice that does not share memory.

```go run
package main

import "fmt"

func main() {
	src := []int{1, 2, 3}
	dst := make([]int, len(src))
	n := copy(dst, src)

	dst[0] = 100
	fmt.Println(n, src, dst)

	short := make([]int, 2)
	copy(short, src)
	fmt.Println(short)
}
```

`dst` needs a **length**, not only a capacity: `copy` into `make([]int, 0, 3)` copies nothing.

## Passing slices to functions

A slice header is copied when passed, but it still points at the same array. A function can modify the *elements* and the caller sees it. But if the function appends, the caller's slice header (and its length) does not change; return the new slice if needed.

```go run
package main

import "fmt"

func double(s []int) {
	for i := range s {
		s[i] *= 2
	}
}

func addOne(s []int) {
	s = append(s, 1) // caller's length is unchanged
}

func main() {
	nums := []int{1, 2, 3}
	double(nums)
	addOne(nums)
	fmt.Println(nums)
}
```

The output is `[2 4 6]`.

## Key takeaways

- `s[a:b]` shares memory with `s`; mutations are visible through both.
- Appending to a sub-slice with spare capacity can overwrite the original's elements.
- `s[a:b:c]` caps the capacity so `append` reallocates.
- `copy(dst, src)` copies up to `min(len(dst), len(src))` elements; `dst` must have length.
- Functions can change elements of a slice argument, but not the caller's length.
