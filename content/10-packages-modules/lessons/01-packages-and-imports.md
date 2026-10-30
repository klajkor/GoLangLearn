---
title: Packages and Imports
minutes: 9
---
Every Go file starts with a `package` clause. A **package** is a directory of `.go` files that are compiled together and share one namespace. Every file in the same directory must declare the same package name (the one exception is an external `_test` package, which you will meet in the testing chapter).

## The main package

A package named `main` is special: it produces an executable, and it must contain a `func main()`. Every other package is a library that other packages import.

```text
hello/
  go.mod
  main.go          // package main
  greet/
    greet.go       // package greet
```

The package name is usually the last element of the import path, but it does not have to be. By convention it is short, lower-case, and has no underscores: `strings`, `http`, `greet`.

## Importing

You import a package by its **import path** (a string), and then use it through its **package name**:

```go
import (
    "fmt"
    "strings"

    "example.com/hello/greet"
)
```

Grouping imports in one parenthesised block is idiomatic. `gofmt` sorts each group, and tools like `goimports` separate standard library imports from third-party ones with a blank line.

Go is strict: **an unused import is a compile error**. This keeps dependency lists honest.

> **Common mistake:** importing a package "just in case" and then getting `"os" imported and not used`. Remove it, or let your editor (`goimports`) manage imports for you.

## Three special import forms

| Form | Example | Effect |
|---|---|---|
| Alias | `str "strings"` | Use `str.ToUpper` instead of `strings.ToUpper` |
| Blank | `_ "embed"` | Import only for side effects (its `init` functions run) |
| Dot | `. "fmt"` | Names are used unqualified: `Println(...)` |

**Aliases** resolve name clashes (two packages both called `rand`) or shorten long names. **Blank imports** are common for drivers and registries: a database driver registers itself in its `init()`, and you import it as `_ "github.com/lib/pq"` without calling anything directly. **Dot imports** are discouraged outside rare test situations because they hide where a name comes from.

```go run
package main

import (
	"fmt"
	m "math"
	str "strings"
)

func main() {
	fmt.Println(str.ToUpper("gopher"))
	fmt.Println(m.Sqrt(144))
	fmt.Println(m.MaxInt8)
}
```

Here `fmt` is used normally, while `math` and `strings` are accessed through the aliases `m` and `str`. The original names `math` and `strings` are **not** available in this file.

## Package names versus import paths

The import path says *where* the code lives; the package name says *how you refer to it*. For example `"math/rand"` has package name `rand`, and `"crypto/rand"` has package name `rand` too. Importing both in one file needs an alias:

```go
import (
    "crypto/rand"
    mrand "math/rand"
)
```

## One package, many files

All files in a directory with the same package name act as one unit. A function declared in `a.go` is visible in `b.go` without any import. There are no header files and no declaration order across files: Go resolves package-level declarations regardless of the order they appear in.

```text
greet/
  greet.go        // func Hello(name string) string
  format.go       // func decorate(s string) string   (used by Hello)
```

`Hello` can call `decorate` freely because they are in the same package, even though they live in different files.

## No import cycles

Package A may not import package B if B (directly or indirectly) imports A. The compiler rejects it with `import cycle not allowed`. If you hit one, it usually means a shared type belongs in a third, lower-level package, or an interface should be defined on the consumer side so the dependency direction flips.

## Key takeaways

- A package is a directory; all its files share one `package` name and one namespace.
- `package main` plus `func main()` builds an executable; everything else is a library.
- Unused imports are compile errors.
- Use an alias to resolve clashes, `_` to import for side effects, and avoid `.` imports.
- The import path locates the code; the package name is how you refer to it.
- Import cycles are forbidden; restructure rather than fight them.
