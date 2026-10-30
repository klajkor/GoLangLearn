---
title: Visibility and init()
minutes: 9
---
Go has no `public` or `private` keywords. Visibility is decided by a single rule about the **first letter** of an identifier.

## Exported vs unexported

- Starts with an **upper-case** letter: *exported*, visible to other packages.
- Starts with a lower-case letter (or `_`): *unexported*, visible only inside its own package.

This applies to functions, types, variables, constants, struct fields and methods.

```go
package account

type Account struct {
    Owner   string // exported field
    balance int    // unexported field
}

func New(owner string) *Account { // exported constructor
    return &Account{Owner: owner}
}

func (a *Account) Deposit(n int) { a.balance += n }
func (a *Account) Balance() int  { return a.balance }
```

From another package you can write `acc.Owner` and `acc.Deposit(5)`, but `acc.balance` is a compile error: `acc.balance undefined (cannot refer to unexported field balance)`. This is how Go does encapsulation: keep fields unexported and expose behaviour through methods.

> **Tip:** an exported identifier is part of your package's public API. Once other code depends on it, renaming it is a breaking change. Export as little as possible.

Visibility is per **package**, not per type or file. Two files in the same package see each other's unexported names, and so do functions that are unrelated to the type.

## Documenting exported names

By convention, a doc comment on an exported name starts with that name:

```go
// Deposit adds n to the account balance.
func (a *Account) Deposit(n int) { a.balance += n }
```

`go doc` and pkg.go.dev render these comments.

## The init function

A package can declare a function named `init` that takes no arguments and returns nothing. The runtime calls it automatically, once, before `main` starts. You cannot call it yourself, and you may declare **several** `init` functions, even in the same file.

Typical uses: registering a driver, validating configuration, or building a lookup table that cannot be expressed as a simple expression.

## Initialization order

Go initializes a program in a well-defined order:

1. Imported packages are initialized first (dependencies before dependents, each package only once).
2. Within a package, **package-level variables** are initialized in declaration order, but a variable that depends on another is initialized after it, whatever order they are written in.
3. Then every `init()` function runs, in the order files were presented to the compiler (the `go` tool sorts files by name) and in source order within a file.
4. Finally `main.main` runs.

```go run
package main

import "fmt"

var a = c + b // depends on c and b
var b = initB()
var c = initC()

func initB() int {
	fmt.Println("initB")
	return 2
}

func initC() int {
	fmt.Println("initC")
	return 3
}

func init() {
	fmt.Println("init #1, a =", a)
}

func init() {
	fmt.Println("init #2")
}

func main() {
	fmt.Println("main")
}
```

The output shows the variable initializers (`initB`, `initC`) running before any `init()`, and `a` being computed only once both `b` and `c` are ready.

## Keep init() small

`init` is implicit: readers cannot see where it is called, it cannot return an error, and it makes testing harder because it runs on import. Prefer explicit constructors like `NewClient(cfg)`. Reserve `init()` for registration-style side effects.

> **Common mistake:** doing network calls or file reads in `init`. A failure there crashes the program before `main` even starts, and nobody can handle the error.

## Package-level variables are shared state

Package-level `var`s are global to the package. Exported ones are global to the whole program. Mutating them from several goroutines without synchronization is a data race. Prefer passing dependencies explicitly.

## Key takeaways

- Upper-case first letter means exported; lower-case means package-private.
- Encapsulate by unexporting fields and exposing methods.
- Visibility is per package, not per file or type.
- Initialization order: imports, then package variables (dependency-ordered), then `init()` functions, then `main`.
- A package may have many `init` functions; they cannot be called or return values.
- Prefer explicit constructors to `init()` for anything beyond registration.
