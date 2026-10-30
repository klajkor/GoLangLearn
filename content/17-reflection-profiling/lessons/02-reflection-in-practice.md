---
title: Tags, Setting Values, Methods and DeepEqual
minutes: 11
---
This lesson covers the four things you will actually do with reflection: read struct tags, modify values, call methods, and compare deeply.

## Struct tags

A tag is a raw string literal after a field. By convention it holds `key:"value"` pairs. `reflect.StructTag` parses them with `Get` and `Lookup`.

```go run
package main

import (
	"fmt"
	"reflect"
	"strings"
)

type Signup struct {
	Email string `json:"email" validate:"required"`
	Name  string `json:"name,omitempty" validate:"required"`
	Bio   string `json:"bio"`
}

func validate(x any) []string {
	var problems []string
	v := reflect.ValueOf(x)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("validate") == "required" && v.Field(i).String() == "" {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			problems = append(problems, name+" is required")
		}
	}
	return problems
}

func main() {
	fmt.Println(validate(Signup{Email: "a@b.c"}))
	fmt.Println(validate(Signup{Email: "a@b.c", Name: "Ann"}))
	_, ok := reflect.TypeOf(Signup{}).Field(2).Tag.Lookup("validate")
	fmt.Println("Bio has validate tag:", ok)
}
```

`Get` returns `""` for a missing key; `Lookup` also tells you whether the key exists, which separates "absent" from "empty".

## Settable values

`reflect.ValueOf(x)` holds a **copy** of `x`, so changing it would be pointless and is forbidden. A value is *settable* only if it is addressable, which in practice means you reached it through a pointer with `Elem()`. Unexported fields are never settable.

```go run
package main

import (
	"fmt"
	"reflect"
)

type Point struct {
	X, Y int
	z    int
}

func main() {
	p := Point{1, 2, 3}

	v := reflect.ValueOf(p)
	fmt.Println("copy settable:", v.Field(0).CanSet())

	e := reflect.ValueOf(&p).Elem()
	fmt.Println("via pointer:", e.Field(0).CanSet())
	fmt.Println("unexported:", e.Field(2).CanSet())

	e.Field(0).SetInt(10)
	e.FieldByName("Y").Set(reflect.ValueOf(20))
	fmt.Println(p)
}
```

> **Common mistake:** calling `Set` on a value obtained from `ValueOf(x)` rather than `ValueOf(&x).Elem()`. It panics with "using unaddressable value".

## Calling methods

`Value.MethodByName` finds an exported method and `Call` invokes it with a slice of `reflect.Value` arguments, returning a slice of results. Only methods in the value's method set are visible: for pointer-receiver methods, start from a pointer.

```go run
package main

import (
	"fmt"
	"reflect"
)

type Greeter struct{ Prefix string }

func (g Greeter) Hello(name string) string { return g.Prefix + ", " + name }
func (g *Greeter) SetPrefix(p string)      { g.Prefix = p }

func main() {
	g := &Greeter{Prefix: "Hi"}
	v := reflect.ValueOf(g)

	t := v.Type()
	for i := 0; i < t.NumMethod(); i++ {
		fmt.Println("method:", t.Method(i).Name)
	}

	v.MethodByName("SetPrefix").Call([]reflect.Value{reflect.ValueOf("Hello")})
	out := v.MethodByName("Hello").Call([]reflect.Value{reflect.ValueOf("Gopher")})
	fmt.Println(out[0].String())

	fmt.Println(reflect.ValueOf(*g).MethodByName("SetPrefix").IsValid())
}
```

Methods are listed in alphabetical order. The last line prints `false`: `SetPrefix` is not in the method set of the plain `Greeter` value.

## DeepEqual

`==` does not work on slices, maps or functions. `reflect.DeepEqual` compares recursively:

```go run
package main

import (
	"fmt"
	"reflect"
)

func main() {
	a := map[string][]int{"x": {1, 2}}
	b := map[string][]int{"x": {1, 2}}
	fmt.Println(reflect.DeepEqual(a, b))

	var s1 []int
	s2 := []int{}
	fmt.Println(reflect.DeepEqual(s1, s2))
	fmt.Println(reflect.DeepEqual(1, int64(1)))
}
```

Gotchas: a `nil` slice is **not** deeply equal to an empty non-nil slice, and values of different types are never equal. In new code, prefer `slices.Equal` / `maps.Equal` (Go 1.21) for plain collections; they are faster and type-safe. For tests, `go-cmp` gives readable diffs.

## Key takeaways

- `field.Tag.Get("key")` and `Lookup` read struct tags; they are how validators and codecs are configured.
- A value is settable only if addressable: use `reflect.ValueOf(&x).Elem()`, and never for unexported fields.
- `MethodByName(...).Call(args)` invokes methods dynamically; check `IsValid()` first.
- `DeepEqual` treats nil and empty slices as different and requires identical types.
