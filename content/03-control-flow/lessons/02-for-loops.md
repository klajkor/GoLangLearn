---
title: for - Go's Only Loop
minutes: 9
---
Go has no `while`, no `do ... while`, and no `foreach`. One keyword, `for`, covers every case, in three shapes.

## 1. The classic three-part loop

```go run
package main

import "fmt"

func main() {
	for i := 0; i < 5; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()
}
```

Init statement, condition, post statement, just like C, but without parentheses. The variable `i` exists only inside the loop. Since Go 1.22, each iteration gets its **own copy** of `i`, which matters when closures or goroutines capture it. Older versions shared one variable across all iterations, which caused famous bugs.

## 2. The condition-only loop (while)

Drop the init and post statements and you have a `while`:

```go
n := 1
for n < 100 {
	n *= 2
}
```

## 3. The infinite loop

With no condition at all, the loop runs until you `break` or `return`:

```go run
package main

import "fmt"

func main() {
	n := 0
	for {
		n++
		if n%2 == 0 {
			continue // skip the rest of this iteration
		}
		if n > 7 {
			break // leave the loop
		}
		fmt.Print(n, " ")
	}
	fmt.Println("\nstopped at", n)
}
```

- `break` exits the innermost loop.
- `continue` jumps to the next iteration (running the post statement in a three-part loop).

## range

`range` iterates over collections. It yields two values per step: an index (or key) and a copy of the element.

```go run
package main

import "fmt"

func main() {
	fruits := []string{"apple", "banana", "cherry"}

	for i, f := range fruits {
		fmt.Println(i, f)
	}

	for _, f := range fruits { // _ discards the index
		fmt.Print(f[:1], " ")
	}
	fmt.Println()

	for i := range fruits { // index only
		fmt.Print(i, " ")
	}
	fmt.Println()

	for i, r := range "go🙂" { // strings yield runes with byte offsets
		fmt.Print(i, ":", string(r), " ")
	}
	fmt.Println()

	ages := map[string]int{"ann": 30}
	for name, age := range ages { // maps yield key, value
		fmt.Println(name, age)
	}

	for i := range 3 { // Go 1.22+: range over an integer
		fmt.Print(i, " ")
	}
	fmt.Println()
}
```

What `range` yields:

| Ranging over | First value | Second value |
|---|---|---|
| slice or array | index | element copy |
| string | byte offset | rune |
| map | key | value |
| channel | element | (none) |
| integer `n` (1.22+) | `0 .. n-1` | (none) |

Two notes. Map iteration order is deliberately **random**, so never rely on it. And the element is a copy, so assigning to the loop variable does not modify the slice; write `fruits[i] = ...` for that.

> **Common mistake:** modifying the loop variable and expecting the collection to change:
>
> ```go
> for _, v := range nums { v *= 2 } // nums is unchanged
> ```
>
> Use `for i := range nums { nums[i] *= 2 }` instead.

## Labels: break and continue in nested loops

A plain `break` only leaves the innermost loop. To leave an outer one, put a **label** in front of it:

```go run
package main

import "fmt"

func main() {
outer:
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 3; j++ {
			if j == 2 {
				continue outer // next i
			}
			if i == 3 {
				break outer // leave both loops
			}
			fmt.Println(i, j)
		}
	}
	fmt.Println("done")
}
```

It prints `1 1` and `2 1`, then `done`. When `j` reaches 2, `continue outer` skips ahead to the next `i`; when `i` is 3, `break outer` ends everything. Labels are used rarely, but they beat flag variables.

> **Tip:** A `break` inside a `switch` or `select` leaves that statement, not the surrounding loop. Use a labelled break when you want to leave the loop.

## Key takeaways

- `for` is the only loop: three-part, condition-only (while), or infinite.
- `range` works on slices, arrays, strings, maps, channels, and (Go 1.22+) integers.
- `range` gives copies of elements; use the index to modify the collection.
- Map iteration order is random.
- `break` and `continue` act on the innermost loop unless you give them a label.
