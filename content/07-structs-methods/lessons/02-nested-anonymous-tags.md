---
title: Nested & Anonymous Structs, Tags
minutes: 7
---
Structs can contain other structs, and you can even write a struct type without naming it. This lesson also gives a first look at struct tags.

## Nested structs

A field can be any type, including another struct.

```go run
package main

import "fmt"

type Address struct {
	City    string
	Country string
}

type Person struct {
	Name string
	Home Address
}

func main() {
	p := Person{
		Name: "Grace",
		Home: Address{City: "Arlington", Country: "US"},
	}

	fmt.Println(p.Home.City)
	p.Home.City = "New York"
	fmt.Printf("%+v\n", p)

	// Zero value: nested structs are zeroed too.
	var empty Person
	fmt.Printf("%q\n", empty.Home.City)
}
```

Nested composite literals need their type written out (`Address{...}`), and a trailing comma on the last line is required when the closing brace is on its own line.

If you want the nested value to be optional, use a pointer (`Home *Address`). Its zero value is `nil`, which tells you "not set". Remember to check for `nil` before using it.

## Anonymous structs

When you need a one-off shape, skip the type declaration:

```go run
package main

import "fmt"

func main() {
	config := struct {
		Host string
		Port int
	}{
		Host: "localhost",
		Port: 8080,
	}
	fmt.Println(config.Host, config.Port)

	// Very common in table-driven tests.
	cases := []struct {
		in   int
		want int
	}{
		{1, 2},
		{2, 4},
		{5, 10},
	}
	for _, c := range cases {
		fmt.Println(c.in*2 == c.want)
	}
}
```

Anonymous structs shine for local data you will never reuse: test tables, JSON snippets, grouping a few values. If you find yourself passing one around, give it a name.

## Struct tags (a teaser)

A field may be followed by a string literal, its **tag**. The compiler ignores it; libraries read it with reflection.

```go
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	token string // unexported: encoding/json skips it
}
```

Here `encoding/json` uses the tags to choose JSON key names, and `omitempty` drops the field when it is empty. Other packages (database mappers, validators, YAML parsers) define their own tag keys. You will use JSON tags in practice later in the course.

The tag convention is `key:"value"` pairs separated by spaces, written in backticks. `go vet` checks that tags are well-formed.

## Field names and visibility

Capitalised field names are **exported** (visible to other packages); lowercase ones are **unexported** (visible only inside the package). This matters for libraries: `encoding/json`, for example, can only see exported fields.

| Field | Visible from other packages? |
|---|---|
| `Name` | yes |
| `name` | no |

> **Common mistake:** writing `name string` in a struct you intend to marshal to JSON, then wondering why the output is `{}`. Lowercase fields are invisible to other packages, including the JSON encoder.

## Key takeaways

- Structs nest by value; nested zero values are zeroed too. Use a pointer field for "optional".
- Anonymous structs are good for one-off data such as test tables.
- Struct tags are metadata read by libraries via reflection; the compiler ignores them.
- Capitalised fields are exported, lowercase fields are not.
