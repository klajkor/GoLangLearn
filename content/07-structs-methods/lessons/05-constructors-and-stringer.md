---
title: Constructors, Visibility & Stringer
minutes: 8
---
Go has no constructors as a language feature. Instead you follow a naming convention and rely on exported/unexported visibility.

## The NewX convention

A function named `NewThing` that returns a ready-to-use value is a constructor by convention. Use one when the zero value is not enough, for example to validate input, set defaults or allocate a map.

```go run
package main

import (
	"errors"
	"fmt"
)

type Account struct {
	owner   string // unexported: only this package can touch it
	balance int
}

func NewAccount(owner string) (*Account, error) {
	if owner == "" {
		return nil, errors.New("owner required")
	}
	return &Account{owner: owner}, nil
}

func (a *Account) Owner() string { return a.owner }
func (a *Account) Balance() int  { return a.balance }

func (a *Account) Deposit(n int) error {
	if n <= 0 {
		return errors.New("deposit must be positive")
	}
	a.balance += n
	return nil
}

func main() {
	acc, err := NewAccount("Ada")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_ = acc.Deposit(50)
	fmt.Println(acc.Owner(), acc.Balance())

	_, err = NewAccount("")
	fmt.Println("error:", err)
}
```

When a package has one main type, the constructor is just `New` (for example `errors.New`).

## Exported vs unexported fields

Because `balance` is lowercase, other packages cannot write `acc.balance = 1_000_000`. They must go through `Deposit`, which enforces the rules. This is Go's encapsulation: it happens at the **package** level, not the type level. Code inside the same package can see all fields.

Getters are named after the field without a `Get` prefix: `Balance()`, not `GetBalance()`. Setters, if you need them, use `SetBalance`.

> **Common mistake:** returning a struct with unexported fields and expecting callers to build it themselves with a literal. Outside the package, they cannot set those fields, so they get a half-initialised value. Document whether the zero value is usable, and provide `NewX` if it is not.

## Setting defaults

Constructors are the right place to set defaults:

```go
func NewServer(addr string) *Server {
	return &Server{
		addr:    addr,
		timeout: 30 * time.Second,
		routes:  make(map[string]Handler), // a nil map would panic on write
	}
}
```

## The Stringer interface

`fmt` checks whether a value has a `String() string` method (the `fmt.Stringer` interface). If so, `Println`, `%v` and `%s` use it.

```go run
package main

import "fmt"

type Level int

const (
	Debug Level = iota
	Info
	Error
)

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Error:
		return "ERROR"
	}
	return fmt.Sprintf("Level(%d)", int(l))
}

type Point struct{ X, Y int }

func (p Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func main() {
	fmt.Println(Info, Error)
	fmt.Printf("%v %d %s\n", Debug, Debug, Debug)
	fmt.Println(Level(7))
	fmt.Println(Point{1, 2})
	fmt.Println([]Point{{1, 2}, {3, 4}})
}
```

`%d` still prints the number because it asks for the integer form. Slices of Stringers print each element using `String`.

> **Common mistake:** formatting the receiver itself with `%v` or `%s` *inside* its own `String` method. That calls `String` again and recurses forever. Convert to a different type or format the fields directly.

Also note the receiver kind: if `String` has a pointer receiver, printing a plain value (not a pointer) will **not** use it, because the value's method set lacks that method.

## Key takeaways

- Constructors are plain functions named `NewX`; use them to validate and set defaults.
- Visibility is per package: lowercase fields are hidden from other packages.
- Name getters `Balance()`, not `GetBalance()`.
- Implement `String() string` to control how `fmt` prints your type.
- Never format the receiver with `%v` inside its own `String` method.
