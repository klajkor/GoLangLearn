---
title: Multiple and Named Return Values
minutes: 8
---
Go functions can return more than one value. This is not a gimmick: it is how Go handles errors, lookups and many everyday operations.

## Returning several values

List the result types in parentheses. The caller receives them in order.

```go run
package main

import "fmt"

func divmod(a, b int) (int, int) {
	return a / b, a % b
}

func main() {
	q, r := divmod(17, 5)
	fmt.Println(q, r)

	// Ignore a value with the blank identifier
	quotient, _ := divmod(9, 2)
	fmt.Println(quotient)
}
```

When you assign the results you must name every one of them; use `_` to discard a value you do not need.

## The (value, error) idiom

By convention, a function that can fail returns its result **and** an `error` as the last value. If the error is `nil`, the call worked.

```go run
package main

import (
	"errors"
	"fmt"
)

func safeDivide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func main() {
	if result, err := safeDivide(10, 4); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("result:", result)
	}

	_, err := safeDivide(1, 0)
	fmt.Println("error:", err)
}
```

There are no exceptions in everyday Go code: you check `err` right after the call. A later chapter covers errors in depth; for now just learn the pattern `value, err := f()` followed by `if err != nil { ... }`.

> **Tip:** when returning an error, return the zero value for the other results (`0`, `""`, `nil`). Callers should not look at them.

## Named return values

You can give the results names. They behave like variables declared at the top of the function, initialised to their zero values, and a bare `return` returns their current values.

```go run
package main

import "fmt"

func minMax(nums []int) (lo, hi int) {
	lo, hi = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < lo {
			lo = n
		}
		if n > hi {
			hi = n
		}
	}
	return // returns lo, hi
}

func main() {
	a, b := minMax([]int{4, 9, -2, 7})
	fmt.Println(a, b)
}
```

Named results are most useful as **documentation** (`(width, height int)` is clearer than `(int, int)`) and with `defer`, which can inspect or modify them as the function exits.

> **Common mistake:** overusing bare `return` in long functions. It makes readers scroll up to figure out what is returned. Prefer an explicit `return lo, hi` unless the function is short.

## Shadowing trap

If you write `x := ...` inside an inner block, you create a *new* variable that hides the named result of the same name. Use `=` to assign to the named result.

## Passing results straight through

If a function returns exactly the types another function accepts, you can nest the call: `fmt.Println(divmod(17, 5))` prints `3 2`. This only works when the call is the sole argument.

## Key takeaways

- Return multiple values with `(T1, T2)`; receive them with `a, b := f()`.
- Discard unwanted values with `_`.
- Fallible functions return `(result, error)`; check `err != nil` immediately.
- Named results are zero-initialised and enable a bare `return`; use them mostly for clarity.
- Beware of shadowing named results with `:=` in inner scopes.
