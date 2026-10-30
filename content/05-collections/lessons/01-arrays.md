---
title: Arrays
minutes: 6
---
An **array** is a fixed-length sequence of elements of one type. The length is part of the type, so `[3]int` and `[4]int` are different types. You will use slices far more often, but arrays are the foundation slices are built on.

## Declaring arrays

```go run
package main

import "fmt"

func main() {
	var a [3]int // zero values: [0 0 0]
	b := [3]string{"x", "y", "z"}
	c := [...]int{10, 20, 30, 40} // compiler counts the elements
	d := [5]int{1: 100, 3: 300}   // set specific indexes

	fmt.Println(a, b, c, d)
	fmt.Println(len(c), c[0], c[len(c)-1])

	a[1] = 7
	fmt.Println(a)
}
```

Indexes start at 0. An out-of-range constant index is a compile error; a runtime index that is out of range panics.

## Arrays are values

Assigning an array, or passing it to a function, **copies every element**. This is very different from arrays in Java, JavaScript or Python.

```go run
package main

import "fmt"

func modify(arr [3]int) {
	arr[0] = 999
}

func main() {
	original := [3]int{1, 2, 3}
	copyOf := original
	copyOf[0] = 42

	modify(original)

	fmt.Println(original)
	fmt.Println(copyOf)
	fmt.Println(original == [3]int{1, 2, 3})
}
```

`original` is untouched by both the copy and the function call. Because they are values, arrays of the same type can be compared with `==` (when their elements are comparable).

> **Common mistake:** expecting a function that takes an array to modify the caller's array. It receives a copy. Use a slice or a pointer if you want sharing.

## Iterating

```go run
package main

import "fmt"

func main() {
	days := [...]string{"Mon", "Tue", "Wed"}
	for i, d := range days {
		fmt.Println(i, d)
	}
}
```

`range` gives the index and a copy of each element.

## Multi-dimensional arrays

```go
var grid [3][3]int
grid[1][2] = 5
```

This is an array of three arrays of three ints, laid out contiguously in memory.

## When to use arrays

- When the size is truly fixed and known at compile time (a chessboard `[8][8]rune`, an RGB colour `[3]byte`, a hash value `[32]byte`).
- As map keys: arrays are comparable, slices are not.
- Almost always, otherwise, use a slice.

## Key takeaways

- `[N]T` has a fixed length that is part of its type.
- Arrays are values: assignment and function calls copy them.
- `[...]T{...}` lets the compiler count; `len(a)` gives the length.
- Arrays of comparable elements can be compared with `==` and used as map keys.
- Prefer slices for everyday sequences.
