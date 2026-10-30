---
title: Reflect Basics - Type and Value
minutes: 10
---
Normally the compiler knows every type in your program. **Reflection** lets code inspect and manipulate values whose types are only known at run time. The `reflect` package is how `encoding/json`, `fmt` and many validators work under the hood.

## Two entry points

Everything starts by converting a value to an `any` and handing it to one of two functions:

| Function | Returns | Describes |
|---|---|---|
| `reflect.TypeOf(x)` | `reflect.Type` | the static shape: name, kind, fields, methods |
| `reflect.ValueOf(x)` | `reflect.Value` | the actual data, which you can read (and sometimes write) |

## Type versus Kind

This distinction trips up almost everyone. The **Type** is the exact named type (`main.MyInt`, `main.User`). The **Kind** is the underlying category (`int`, `struct`, `slice`, `ptr`, `map`, `func` ...). Code that handles "any integer" switches on `Kind`; code that cares about identity compares `Type`.

```go run
package main

import (
	"fmt"
	"reflect"
)

type MyInt int

type User struct {
	Name string
	Age  int
}

func describe(x any) {
	t := reflect.TypeOf(x)
	fmt.Printf("%-10v name=%-6q kind=%v\n", t, t.Name(), t.Kind())
}

func main() {
	describe(42)
	describe(MyInt(7))
	describe("hi")
	describe([]int{1, 2})
	describe(User{})
	describe(&User{})
	describe(map[string]bool{})
}
```

Note that `*User` has an empty `Name()` (pointers are unnamed types) and kind `ptr`. Use `t.Elem()` to get the type it points to.

## Reading values

A `reflect.Value` has typed getters such as `Int()`, `String()`, `Float()`, `Bool()`, and `Len()` / `Index(i)` for collections. Calling the wrong getter panics, so check `Kind()` first. `v.Interface()` converts a `Value` back to an `any`.

## Iterating struct fields

`NumField` and `Field(i)` exist on both `Type` (metadata) and `Value` (data). Combine them to walk a struct:

```go run
package main

import (
	"fmt"
	"reflect"
)

type Config struct {
	Host    string
	Port    int
	Debug   bool
	private string
}

func dump(x any) {
	v := reflect.ValueOf(x)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			fmt.Printf("%s: (unexported, skipped)\n", f.Name)
			continue
		}
		fmt.Printf("%s (%s) = %v\n", f.Name, f.Type, v.Field(i).Interface())
	}
}

func main() {
	dump(Config{Host: "localhost", Port: 8080, Debug: true, private: "x"})
}
```

Calling `Interface()` on an unexported field panics, which is why the loop checks `IsExported()` first.

## Pointers and Elem

`reflect.ValueOf(&cfg)` has kind `Ptr`. Call `.Elem()` to reach the struct it points to. `Elem` also unwraps an interface value. Calling `NumField` directly on a pointer panics, so a common helper is:

```go
v := reflect.ValueOf(x)
for v.Kind() == reflect.Pointer {
	v = v.Elem()
}
```

(`reflect.Pointer` was added in Go 1.18 as the preferred name for `reflect.Ptr`.)

> **Common mistake:** assuming `Kind()` returns the named type. `reflect.TypeOf(MyInt(1)).Kind()` is `int`, not `MyInt`.

> **Tip:** most reflection bugs are panics from wrong-kind operations. Switch on `Kind()` and handle the cases you support; return an error for the rest.

## Key takeaways

- `reflect.TypeOf` gives metadata; `reflect.ValueOf` gives data.
- `Kind` is the underlying category; `Type` is the exact named type.
- Use `NumField`/`Field(i)` to walk structs, and skip unexported fields before calling `Interface()`.
- `Elem()` dereferences pointers (and unwraps interfaces).
- Wrong-kind getters panic, so always check `Kind()` first.
