---
title: The slices and maps Packages
minutes: 7
---
Since Go 1.21 the standard library ships generic helpers in the `slices` and `maps` packages. They remove a lot of boilerplate loops. Both work with any element type thanks to generics, which you will learn about later; for now just use them.

## slices

```go run
package main

import (
	"fmt"
	"slices"
)

func main() {
	nums := []int{5, 2, 8, 1, 9}

	slices.Sort(nums)
	fmt.Println(nums)

	fmt.Println(slices.Contains(nums, 8))
	fmt.Println(slices.Index(nums, 8))
	fmt.Println(slices.Max(nums), slices.Min(nums))

	idx, found := slices.BinarySearch(nums, 5)
	fmt.Println(idx, found)

	slices.Reverse(nums)
	fmt.Println(nums)
}
```

`slices.Sort` sorts in place (ascending). `BinarySearch` requires a sorted slice. `Index` returns `-1` when the value is absent.

More handy functions:

| Function | What it does |
|---|---|
| `slices.Equal(a, b)` | same length and equal elements |
| `slices.Clone(s)` | independent copy |
| `slices.Insert(s, i, v...)` | insert values at index `i` |
| `slices.Delete(s, i, j)` | remove `s[i:j]` |
| `slices.IndexFunc(s, f)` | first index where `f` is true |
| `slices.SortFunc(s, cmp)` | sort with a comparison function |
| `slices.Compact(s)` | remove consecutive duplicates |

## Custom sorting

`SortFunc` takes a function returning negative, zero or positive, like `strings.Compare` or `cmp.Compare`.

```go run
package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

func main() {
	words := []string{"pear", "fig", "banana", "kiwi"}

	slices.SortFunc(words, func(a, b string) int {
		if c := cmp.Compare(len(a), len(b)); c != 0 {
			return c
		}
		return strings.Compare(a, b)
	})
	fmt.Println(words)

	c := slices.Clone(words)
	c[0] = "changed"
	fmt.Println(words[0], c[0])

	words = slices.Insert(words, 1, "plum")
	words = slices.Delete(words, 0, 1)
	fmt.Println(words)
}
```

> **Common mistake:** assuming `slices.Delete` and `slices.Insert` modify the slice in place. Always assign the result: `s = slices.Delete(s, i, j)`.

## maps

`maps.Keys` and `maps.Values` return **iterators** (Go 1.23+), not slices. Combine them with `slices.Collect` or `slices.Sorted` to get a slice.

```go run
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	stock := map[string]int{"pear": 4, "apple": 0, "fig": 12}

	keys := slices.Sorted(maps.Keys(stock))
	fmt.Println(keys)

	values := slices.Sorted(maps.Values(stock))
	fmt.Println(values)

	clone := maps.Clone(stock)
	clone["kiwi"] = 1
	fmt.Println(len(stock), len(clone))

	fmt.Println(maps.Equal(stock, clone))

	maps.DeleteFunc(clone, func(k string, v int) bool { return v == 0 })
	fmt.Println(slices.Sorted(maps.Keys(clone)))
}
```

`slices.Sorted(maps.Keys(m))` is the modern one-liner for "iterate a map in key order", and it replaces the manual collect-and-sort loop from the previous lesson. Other useful functions are `maps.Copy(dst, src)` and `maps.Equal`.

> **Tip:** if you must support Go older than 1.23, `maps.Keys` is not in the standard library; use the manual collect-and-sort loop.

## Key takeaways

- `slices` offers `Sort`, `Contains`, `Index`, `Max`, `Min`, `Reverse`, `Clone`, `Insert`, `Delete`, `BinarySearch` and more.
- `slices.SortFunc` uses a comparison function returning an int; `cmp.Compare` helps.
- `maps.Keys` and `maps.Values` return iterators; wrap them in `slices.Sorted` or `slices.Collect`.
- Functions that change length (`Insert`, `Delete`) return the new slice; assign it.
- For deterministic output from maps, sort the keys first.
