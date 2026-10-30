---
title: Maps
minutes: 9
---
A **map** is an unordered collection of key-value pairs, Go's built-in hash table. The type is written `map[KeyType]ValueType`.

## Creating and using maps

```go run
package main

import "fmt"

func main() {
	ages := map[string]int{
		"alice": 30,
		"bob":   25,
	}
	ages["carol"] = 41 // insert
	ages["bob"] = 26   // update

	fmt.Println(ages["alice"], len(ages))

	scores := make(map[string]int) // empty, ready to use
	scores["go"]++                 // missing key reads as 0, so this becomes 1
	scores["go"]++
	fmt.Println(scores)
}
```

Reading a missing key gives the **zero value** of the value type (`0`, `""`, `nil`...), which makes counting idioms like `m[k]++` work without initialisation.

Keys must be **comparable**: numbers, strings, bools, pointers, arrays and structs of those. Slices, maps and functions cannot be keys.

## The comma-ok idiom

How do you tell a stored zero from a missing key? Use the two-value form.

```go run
package main

import "fmt"

func main() {
	stock := map[string]int{"apple": 0, "pear": 4}

	n, ok := stock["apple"]
	fmt.Println(n, ok) // present, value 0

	n, ok = stock["plum"]
	fmt.Println(n, ok) // absent

	if v, found := stock["pear"]; found {
		fmt.Println("pear:", v)
	}
}
```

## delete

`delete(m, key)` removes an entry. It is safe to delete a missing key or to call it on a nil map.

```go run
package main

import "fmt"

func main() {
	m := map[string]bool{"a": true, "b": true}
	delete(m, "a")
	delete(m, "zzz") // no-op
	fmt.Println(len(m), m)
}
```

Printing a map with `fmt` shows keys in sorted order, which is a convenience of the `fmt` package only.

## Iteration order is random

`for k, v := range m` visits entries in an **unspecified order**, and the order can change between runs. Never depend on it. To get deterministic output, collect the keys and sort them.

```go run
package main

import (
	"fmt"
	"sort"
)

func main() {
	m := map[string]int{"banana": 3, "apple": 5, "cherry": 7}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%s=%d\n", k, m[k])
	}
}
```

## Nil maps

A nil map behaves like an empty map for reading, `len`, `range` and `delete`, but **writing to it panics**.

```go
var m map[string]int
fmt.Println(m["x"]) // 0, fine
m["x"] = 1          // panic: assignment to entry in nil map
```

> **Common mistake:** declaring `var m map[string]int` and then assigning into it. Create it with `make` or a literal first.

## Maps of slices (and sets)

Because a missing key yields a nil slice, `append` just works:

```go run
package main

import "fmt"

func main() {
	groups := map[int][]string{}
	for _, w := range []string{"go", "rust", "c", "java", "zig"} {
		n := len(w)
		groups[n] = append(groups[n], w)
	}
	fmt.Println(groups[2], groups[4])

	seen := map[string]struct{}{} // a set: empty struct takes no space
	seen["a"] = struct{}{}
	_, ok := seen["a"]
	fmt.Println(ok, len(seen))
}
```

Note that map values are **not addressable**: `m[k].field = x` does not compile for struct values; copy out, modify, and store back.

## Key takeaways

- `map[K]V`, created with `make` or a literal; `m[k] = v` inserts or updates.
- Missing keys read as the zero value; use `v, ok := m[k]` to detect them.
- `delete(m, k)` removes entries; maps are unordered, so sort keys for stable output.
- Writing to a nil map panics; reading from one is fine.
- Maps are reference-like: passing a map to a function shares the same data.
