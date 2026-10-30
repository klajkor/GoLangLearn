---
title: Map, Filter, Reduce and the Standard Library
minutes: 10
---
Generics make small, reusable algorithms possible. In this lesson you write the classic trio yourself, then meet the standard packages `slices`, `maps` and `cmp` that already do most of this for you (all require Go 1.21 or newer; the iterator helpers in `maps` need Go 1.23).

## Map, Filter, Reduce

```go run
package main

import (
	"fmt"
	"strconv"
)

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func Filter[T any](xs []T, keep func(T) bool) []T {
	var out []T
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}

func main() {
	nums := []int{1, 2, 3, 4, 5, 6}

	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println(evens)

	strs := Map(evens, func(n int) string { return "n" + strconv.Itoa(n) })
	fmt.Println(strs)

	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Println(sum)

	// The accumulator type can differ from the element type.
	joined := Reduce(nums, "", func(acc string, n int) string { return acc + strconv.Itoa(n) })
	fmt.Println(joined)
}
```

All type arguments are inferred. `Map` has two type parameters (`T` in, `U` out); `Reduce` has the element type `T` and the accumulator type `A`. The function literals have explicit parameter types, which is how the compiler learns `T` and `U`.

## The `slices` package

`slices` contains the generic helpers that you would otherwise write by hand. A selection:

| Function | What it does |
|---|---|
| `slices.Contains(s, v)` / `slices.Index(s, v)` | membership / position (`comparable` elements) |
| `slices.Sort(s)` | sort in place (`cmp.Ordered` elements) |
| `slices.SortFunc(s, cmp)` | sort with a comparison function |
| `slices.BinarySearch(s, v)` | search a sorted slice, returns `(index, found)` |
| `slices.Max(s)` / `slices.Min(s)` | extremes (panic on an empty slice) |
| `slices.Reverse(s)` | reverse in place |
| `slices.Equal(a, b)` | element-wise equality |

## `cmp` and `maps`

`cmp.Compare(a, b)` returns -1, 0 or +1 and is exactly what `slices.SortFunc` wants. `cmp.Or` returns the first non-zero argument, handy for tie-breaking (Go 1.22+). The `maps` package works on any map type; `maps.Keys` returns an iterator (Go 1.23+), and `slices.Sorted` turns an iterator into a sorted slice.

```go run
package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	nums := []int{5, 2, 8, 1}
	slices.Sort(nums)
	fmt.Println(nums, slices.Contains(nums, 8), slices.Index(nums, 5))
	i, found := slices.BinarySearch(nums, 5)
	fmt.Println(i, found, slices.Max(nums))

	people := []Person{{"Ann", 30}, {"Bob", 25}, {"Cid", 30}}
	slices.SortFunc(people, func(a, b Person) int {
		// Older first; ties broken by name.
		return cmp.Or(cmp.Compare(b.Age, a.Age), cmp.Compare(a.Name, b.Name))
	})
	fmt.Println(people)

	ages := map[string]int{"ann": 30, "bob": 25, "cid": 41}
	keys := slices.Sorted(maps.Keys(ages))
	fmt.Println(keys)

	clone := maps.Clone(ages)
	clone["dan"] = 7
	fmt.Println(len(ages), len(clone), maps.Equal(ages, clone))
}
```

Because `slices.Sort` needs `<`, it requires `cmp.Ordered`. For anything else (structs, custom order) use `slices.SortFunc`.

> **Common mistake:** relying on map iteration order. `maps.Keys` yields in random order; wrap it in `slices.Sorted` whenever the order matters (for printing, tests, deterministic output).

> **Tip:** before writing a helper, check `slices` and `maps`. `slices.IndexFunc`, `slices.ContainsFunc`, `slices.DeleteFunc`, `slices.Insert`, `slices.Compact` and `maps.DeleteFunc` all exist. There is deliberately **no** `slices.Map` or `slices.Filter`: the Go team prefers a plain `for` loop for those, which is often just as short and easier to debug.

## Inference in practice

Passing a function literal or a named function usually lets the compiler infer all arguments. When it cannot (no arguments to learn from), or when you want a different type than the default (an untyped constant `1` becomes `int`), give the type argument:

```go
total := Reduce(nums, 0.0, func(acc float64, n int) float64 { return acc + float64(n) })
big := Sum[int64]([]int64{1, 2})
```

## Key takeaways

- Generic `Map`, `Filter` and `Reduce` are a few lines each; type arguments are inferred from the arguments.
- The standard `slices`, `maps` and `cmp` packages are generic and cover sorting, searching, cloning and comparing.
- `slices.Sort` needs ordered elements; `slices.SortFunc` takes a comparison function returning a negative, zero or positive int.
- `slices.Sorted(maps.Keys(m))` is the idiom for a deterministic list of keys (Go 1.23+).
- The standard library intentionally omits `Map`/`Filter`; a simple loop is idiomatic Go.
