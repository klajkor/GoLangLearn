---
title: Why Generics, and Type Parameters
minutes: 9
---
Before Go 1.18 you had two bad choices when the same logic had to work for several types: copy and paste the function for each type, or accept `any` and lose type safety. **Generics** give you a third choice: write the function once and let the compiler check it for each concrete type.

## The problem: same logic, different types

```go
func SumInts(xs []int) int {
	var total int
	for _, x := range xs {
		total += x
	}
	return total
}

func SumFloats(xs []float64) float64 {
	var total float64
	for _, x := range xs {
		total += x
	}
	return total
}
```

The bodies are identical; only the types differ. The `any` workaround is worse: the caller must type-assert the result, and the compiler can no longer stop you from passing a slice of strings.

## The fix: a type parameter

A **type parameter** is declared in square brackets before the normal parameters. It has a name and a **constraint** that says which types are allowed.

```go run
package main

import "fmt"

// T is the type parameter; "int | float64" is its constraint.
func Sum[T int | float64](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

func main() {
	fmt.Println(Sum([]int{1, 2, 3}))
	fmt.Println(Sum([]float64{1.5, 2.5}))
	// Sum([]string{"a"}) would not compile: string does not satisfy int | float64
}
```

Read the signature as: "for any type `T` that is `int` or `float64`, take a `[]T` and return a `T`".

## Type inference

You rarely write the type argument yourself. The compiler infers `T` from the arguments, so `Sum([]int{1, 2, 3})` means `Sum[int]([]int{1, 2, 3})`. You can still be explicit, and sometimes you must.

```go run
package main

import "fmt"

func Zero[T any]() T {
	var zero T
	return zero
}

func First[T any](xs []T) (T, bool) {
	if len(xs) == 0 {
		var zero T // the zero value of whatever T is
		return zero, false
	}
	return xs[0], true
}

func main() {
	// Nothing to infer from, so the type argument is required.
	fmt.Printf("%q %v %v\n", Zero[string](), Zero[int](), Zero[*int]())

	v, ok := First([]string{"go", "rust"})
	fmt.Println(v, ok)

	n, ok := First([]int{})
	fmt.Println(n, ok)
}
```

Inference works from function arguments. It cannot guess from the variable you assign the result to, which is why `Zero[string]()` needs help.

## The zero value: `var zero T`

You do not know what `T` is, so you cannot write `return 0` or `return ""`. Declare `var zero T` and return that: it is `0` for numbers, `""` for strings, `nil` for pointers and a zeroed struct for structs. You will use this idiom constantly when a function has "nothing to return".

## Several type parameters

Separate them with commas. Parameters that share a constraint can be grouped like ordinary parameters (`[K, V comparable]`).

```go run
package main

import "fmt"

func Keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func main() {
	m := map[string]int{"a": 1}
	fmt.Println(Keys(m))
}
```

`K comparable` is needed because map keys must support `==`. For a bigger map the order of the result is unspecified, so sort it before printing (the `slices` package, covered later, does this neatly).

> **Common mistake:** writing `func Sum[T](xs []T) T`. Every type parameter needs a constraint; use `any` if you truly accept every type.

> **Tip:** a type argument that does not fit its constraint is a compile error, not a run-time panic.

## Key takeaways

- Generics let one function work for many types while keeping static type checking.
- Type parameters live in `[...]` before the regular parameter list, each with a constraint.
- The compiler usually infers type arguments from the call arguments; supply them explicitly when there is nothing to infer from.
- Use `var zero T` to produce the zero value of a type parameter.
- Prefer generics over `any` plus type assertions for algorithms that are the same for every type.
