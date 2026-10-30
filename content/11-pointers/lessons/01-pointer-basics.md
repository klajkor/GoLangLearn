---
title: Pointer Basics
minutes: 9
---
A **pointer** holds the memory address of a value. Instead of copying a value around, you can pass its address, letting other code read or change the original. Go has pointers, but they are deliberately simple: no pointer arithmetic, and the garbage collector keeps memory safe.

## The two operators

| Syntax | Meaning |
|---|---|
| `&x` | the address of `x` (produces a pointer) |
| `*p` | the value `p` points to (dereference) |
| `*T` | in a type: "pointer to `T`" |

```go run
package main

import "fmt"

func main() {
	x := 42
	p := &x // p has type *int

	fmt.Println(x, *p) // 42 42

	*p = 100 // write through the pointer
	fmt.Println(x)     // 100: x changed

	x = 7
	fmt.Println(*p) // 7: p still points at x

	q := p // copies the address, not the int
	*q = 1
	fmt.Println(x, p == q) // 1 true
}
```

`p` and `q` are two pointers to the same variable. Pointer equality compares addresses: `p == q` is true when both point to the same variable.

## The zero value is nil

A pointer that has not been assigned is `nil`. It points nowhere, and **dereferencing a nil pointer panics** at run time with `invalid memory address or nil pointer dereference`.

```go run
package main

import "fmt"

func safeRead(p *int) (v int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return *p, nil
}

func main() {
	var p *int
	fmt.Println(p == nil) // true

	_, err := safeRead(p)
	fmt.Println(err)

	n := 5
	v, err := safeRead(&n)
	fmt.Println(v, err)
}
```

Normally you do not recover from nil dereferences; you **check** before using a pointer that might be nil:

```go
if p != nil {
    fmt.Println(*p)
}
```

> **Common mistake:** assuming a function returning `*T` always returns a valid pointer. If a function documents that it can return `nil` (for example "not found"), check it every time.

## Creating pointers

There are three common ways to get a pointer:

```go
x := 10
p1 := &x           // address of an existing variable
p2 := new(int)     // allocates a zeroed int, returns *int
p3 := &Point{1, 2} // address of a composite literal
```

`new(T)` allocates a zero value of `T` and returns a pointer to it, so `*p2` is `0`. It is rarely needed; composite literals such as `&Point{...}` are much more common. You can take the address of a composite literal, but you cannot take the address of a constant or of a map element:

```go
p := &10          // compile error
m := map[string]int{"a": 1}
q := &m["a"]      // compile error: map elements are not addressable
```

## Pointers to pointers

`**int` is legal ("a pointer to a pointer to an int") and used occasionally, for example when a function needs to replace a pointer owned by its caller. Reach for it rarely.

## No pointer arithmetic

Unlike C, you cannot add to a pointer or step through memory with `p++`. If you need to walk a sequence, use a slice and an index. (The `unsafe` package can do it, but it is outside normal Go, and not covered here.) This restriction is a large part of why Go programs are memory-safe.

## Pointers are values too

A pointer is itself a value (typically one machine word) that is copied when you pass it. Two functions given the same pointer see the same underlying variable, but reassigning the pointer *variable* inside a function does not affect the caller's pointer:

```go
func reset(p *int) {
    p = nil // only changes the local copy of the pointer
}
```

To change what the caller's variable holds, write through it: `*p = 0`.

## Key takeaways

- `&x` gives the address of `x`; `*p` reads or writes the value `p` points to.
- The zero value of a pointer is `nil`; dereferencing it panics.
- `new(T)` and `&T{...}` create pointers to fresh values; `&T{...}` is the common form.
- You cannot take the address of constants or map elements.
- There is no pointer arithmetic in Go.
- Pointers are copied by value; write through `*p` to change the original.
