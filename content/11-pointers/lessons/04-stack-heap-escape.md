---
title: Stack, Heap and Escape Analysis
minutes: 8
---
Where does a variable live? In C you decide with `malloc` versus a local. In Go you **do not decide**: the compiler does, using *escape analysis*. This lesson gives you the mental model so that performance discussions make sense.

## Stack versus heap

- The **stack** is per goroutine, fast, and automatically reclaimed when a function returns. Allocation is nearly free.
- The **heap** is shared memory managed by the garbage collector (GC). Allocation costs more, and the GC has to find and free unreachable values later.

Go's goroutine stacks start small and grow as needed, so deep recursion and many goroutines are fine.

## The rule: values that outlive their function must escape

A local variable can stay on the stack only if the compiler can prove nothing references it after the function returns. If something might (a returned pointer, a stored closure, a value placed in an interface the compiler cannot track), the variable *escapes to the heap*.

```go
type Point struct{ X, Y int }

func onStack() int {
	p := Point{1, 2} // does not escape
	return p.X + p.Y
}

func onHeap() *Point {
	p := Point{1, 2}
	return &p // escapes: the caller keeps using it
}
```

Both are correct Go; you never free memory yourself. The difference is only cost.

## Seeing the compiler's decisions

Ask the compiler to print its escape analysis with `-gcflags=-m`:

```bash
go build -gcflags=-m ./...
# or for one run:
go run -gcflags=-m main.go
```

Typical output:

```text
./main.go:9:2: moved to heap: p
./main.go:13:13: ... argument does not escape
./main.go:13:14: x escapes to heap
```

`moved to heap: p` is the line of interest for `onHeap`. Add `-m -m` for more detail about why. Messages vary slightly between Go versions, but the idea is stable.

## Common causes of escape

| Cause | Example |
|---|---|
| Returning a pointer to a local | `return &p` |
| Storing a pointer in something longer-lived | assigning to a global or a field of a heap object |
| Converting to an interface | passing a value to `fmt.Println(x)` |
| Closures capturing variables that outlive the function | returning `func() { x++ }` |
| Size unknown at compile time or very large | `make([]int, n)` with variable `n`, huge arrays |
| Sending a pointer on a channel | the receiver may use it later |

Many of these are unavoidable and perfectly fine. Escape analysis is something you *read when profiling shows an allocation problem*, not something to optimise by default.

## Closures and escaping variables

```go run
package main

import "fmt"

func counter() func() int {
	n := 0 // captured by the closure, so it must outlive counter()
	return func() int {
		n++
		return n
	}
}

type Point struct{ X, Y int }

func newPoint(x, y int) *Point {
	p := Point{x, y}
	return &p // safe: p escapes to the heap automatically
}

func main() {
	next := counter()
	fmt.Println(next(), next(), next()) // 1 2 3

	a := newPoint(1, 2)
	b := newPoint(1, 2)
	fmt.Println(*a == *b, a == b) // true false: equal values, different addresses
}
```

`n` is shared by every call of `next`; it clearly cannot die when `counter` returns. Go handles this transparently. Likewise `a` and `b` are different allocations, so comparing the pointers gives `false` even though the pointed-to values are equal.

## Pointers are not automatically faster

A common beginner assumption is "pointers avoid copying, so they are faster". Often the opposite holds:

- Copying a small struct (a few words) is extremely cheap and stays on the stack.
- Passing a pointer can force the value to the heap, adding allocation and GC work, and an extra indirection on each access.
- Pointer-heavy data structures hurt cache locality.

Measure before deciding. The testing chapter shows how to benchmark with `-benchmem` to see allocations per operation.

> **Tip:** rule of thumb: use values for small data, pointers when you need mutation, sharing, `nil`, or when a type is large or must not be copied. Do not use pointers purely as a speed trick.

## The garbage collector in one paragraph

Go uses a concurrent, non-generational, mark-and-sweep collector. It finds everything reachable from live goroutine stacks and global variables, and frees the rest. Fewer heap allocations means less work for it. You rarely tune it, but `GOGC` and `GOMEMLIMIT` environment variables exist for when you must.

## Key takeaways

- Go decides stack versus heap through escape analysis; you cannot and need not choose.
- A value that outlives its function (returned pointer, captured by a closure, stored globally) is moved to the heap.
- `go build -gcflags=-m` prints those decisions (`moved to heap`, `escapes to heap`).
- Returning `&local` is safe; the garbage collector reclaims it once unreachable.
- Pointers do not guarantee speed; small values on the stack are often faster. Benchmark to decide.
