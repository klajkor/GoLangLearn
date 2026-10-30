---
title: When to Use Reflection (and What It Costs)
minutes: 8
---
Reflection is powerful, but Go's proverb says it plainly: *"Clear is better than clever. Reflection is never clear."* This lesson helps you decide when it earns its place.

## Where reflection belongs

Reflection is the right tool when the code must work for types it cannot know at compile time **and** generics or interfaces cannot express it:

| Use case | Why reflection |
|---|---|
| `encoding/json`, `encoding/xml`, YAML libraries | map arbitrary structs to text through field names and tags |
| ORMs and SQL scanners | map columns to struct fields |
| Validators, config loaders | read tags such as `validate:"required"` or `env:"PORT"` |
| `fmt` | print any value with `%v` |
| Dependency-injection and test frameworks | inspect function signatures |

Library authors use it so that application authors do not have to. If you are writing application logic, you almost never need it.

## Alternatives to try first

1. **Interfaces.** If every type can implement `Validate() error`, call that. No reflection.
2. **Generics.** `func Map[T, U any](s []T, f func(T) U) []U` works for any element type with full type safety.
3. **Type switches.** If you only support five concrete types, a `switch v := x.(type)` is faster and clearer.
4. **Code generation.** Tools run by `go generate` (such as `stringer`) produce plain code ahead of time.

## The costs

- **Speed.** Reflective access is typically many times slower than a direct field access and often allocates, because values get boxed into `any`.
- **Safety.** The compiler can no longer check you. Mistakes become run-time panics ("call of reflect.Value.Int on string Value").
- **Readability.** Readers cannot follow the types, and tools like "find usages" and refactoring miss reflective uses (field names in strings, tags).
- **Binary size.** A reflective call such as `MethodByName` with a non-constant name forces the linker to keep many methods that it could otherwise drop.

You can measure the speed difference yourself with a benchmark. Run this locally with `go test -bench .`; it is not a playground example because benchmarks need `go test`:

```go
package main

import (
	"reflect"
	"testing"
)

type T struct{ N int }

func BenchmarkDirect(b *testing.B) {
	t := &T{}
	for i := 0; i < b.N; i++ {
		t.N = i
	}
}

func BenchmarkReflect(b *testing.B) {
	t := &T{}
	f := reflect.ValueOf(t).Elem().FieldByName("N")
	for i := 0; i < b.N; i++ {
		f.SetInt(int64(i))
	}
}
```

Even in this best case (the field lookup is hoisted out of the loop), the reflective version is slower. Doing `FieldByName` inside the loop is slower still, because it searches by string each time.

## Using it responsibly

If you do write reflective code:

- **Cache.** Parse tags and field layouts **once** per type (for example into a `sync.Map` keyed by `reflect.Type`) and reuse. `encoding/json` does exactly this.
- **Check kinds and return errors.** Do not let a bad input become a panic in a server.
- **Keep it at the edge.** Wrap it in a small, well-tested package with a typed public API.
- **Test with many shapes:** pointers, nested structs, nil values, unexported fields, embedded structs.

> **Common mistake:** reaching for `reflect` to avoid writing a few lines of repetitive code. A few explicit lines are usually better than a clever reflective helper.

> **Tip:** if a function accepts `any` and immediately calls `reflect`, consider whether a generic constraint or an interface would document the contract better.

## Key takeaways

- Reflection suits libraries that handle arbitrary user types (JSON, ORMs, validators).
- Prefer interfaces, generics, type switches or code generation first.
- It is slower, defers errors to run time, and hurts readability.
- Cache per-type metadata and keep reflective code behind a typed API.
