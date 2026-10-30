---
title: Interface Values & the nil Gotcha
minutes: 8
---
To understand interfaces deeply, you need to know what is actually stored in an interface variable.

## The (type, value) pair

An interface value has two parts: a **dynamic type** and a **dynamic value**.

```text
var s Shape = Rect{3, 4}      ->  (type: Rect,  value: {3 4})
var e error = (*MyErr)(nil)   ->  (type: *MyErr, value: nil)
var x any                      ->  (type: nil,   value: nil)
```

An interface is `== nil` **only when both parts are nil**, meaning no type has been set at all.

```go run
package main

import "fmt"

func main() {
	var x any
	fmt.Println(x == nil)
	fmt.Printf("%T %v\n", x, x)

	x = 42
	fmt.Println(x == nil)
	fmt.Printf("%T %v\n", x, x)

	var p *int // nil pointer
	x = p
	fmt.Println(x == nil) // false! type is *int, value is nil
	fmt.Printf("%T %v\n", x, x)
}
```

The last case is the classic surprise: a nil pointer stored in an interface produces a **non-nil interface**.

## The famous nil-interface bug

```go run
package main

import "fmt"

type MyErr struct{ Msg string }

func (e *MyErr) Error() string { return e.Msg }

// Buggy: returns a typed nil pointer wrapped in a non-nil error.
func mayFailBuggy(fail bool) error {
	var err *MyErr // nil
	if fail {
		err = &MyErr{"it failed"}
	}
	return err
}

// Correct: return a literal nil when there is no error.
func mayFail(fail bool) error {
	if fail {
		return &MyErr{"it failed"}
	}
	return nil
}

func main() {
	err := mayFailBuggy(false)
	fmt.Println(err == nil) // false, even though nothing failed!
	fmt.Printf("%T\n", err)

	err = mayFail(false)
	fmt.Println(err == nil)
}
```

In `mayFailBuggy` the function's return type is `error` (an interface), and `err` is a `*MyErr`. Returning it converts it into an interface value with type `*MyErr` and value `nil`, which is not equal to `nil`. The caller's `if err != nil` check fires on success.

> **Common mistake:** declaring a variable of a concrete error type, then returning it as `error`. Always `return nil` explicitly on the success path, and declare variables as `error`, not as a concrete pointer type, when they will be returned as one.

## Calling methods on a nil interface

A truly nil interface has no type, so calling a method on it panics with a nil-pointer dereference:

```go
var s Shape
s.Area() // panic: invalid memory address or nil pointer dereference
```

A *non-nil* interface holding a nil pointer is different: the method is called, with a nil receiver. That works if the method handles `nil`, and panics if it dereferences the receiver.

```go run
package main

import "fmt"

type Tree struct {
	Left, Right *Tree
	Val         int
}

// Safe on a nil receiver.
func (t *Tree) Sum() int {
	if t == nil {
		return 0
	}
	return t.Val + t.Left.Sum() + t.Right.Sum()
}

func main() {
	t := &Tree{Val: 1, Left: &Tree{Val: 2}, Right: &Tree{Val: 3}}
	fmt.Println(t.Sum())

	var empty *Tree
	fmt.Println(empty.Sum())
}
```

Handling a nil receiver is an occasional, elegant technique for recursive structures.

## Comparing interface values

Two interface values are equal if their dynamic types are identical **and** their dynamic values are equal. If the dynamic type is not comparable (a slice, map or function), comparing panics at run time:

```go
var a, b any = []int{1}, []int{1}
fmt.Println(a == b) // panic: runtime error: comparing uncomparable type []int
```

Interface values holding comparable types (ints, strings, pointers, comparable structs) can be compared and used as map keys.

## Size of the pair

Because an interface stores a type descriptor plus a value (or a pointer to one), converting a value to an interface may allocate memory. That is usually fine, but it explains why tight loops in performance-critical code avoid unnecessary interface conversions.

## Key takeaways

- An interface value is a (dynamic type, dynamic value) pair.
- It is `nil` only if the type part is nil too; a nil `*T` in an interface is a non-nil interface.
- Return a literal `nil` for "no error"; do not return a nil concrete pointer as `error`.
- Calling a method on a nil interface panics; on a nil pointer inside an interface, it runs with a nil receiver.
- Comparing interfaces holding uncomparable types (slices, maps) panics.
