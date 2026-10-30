---
title: Hello, World and Program Structure
minutes: 8
---
Time for the traditional first program. Press **Run**, then change the message and run again.

```go run
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
```

Every part of this tiny file matters, so let us go through it top to bottom.

## package main

Every Go file starts with a `package` clause. Packages are Go's unit of code organisation: all files in one folder belong to the same package. The special package name `main` tells the toolchain "this is an executable program, not a library". Library packages use any other name, like `strings` or `geometry`.

## import

`import "fmt"` makes the standard-library package `fmt` (format) available. Multiple imports are grouped in parentheses, and `gofmt` keeps them sorted:

```go
import (
	"fmt"
	"strings"
)
```

> **Common mistake:** Go refuses to compile if you import a package and do not use it (`"strings" imported and not used`). The same rule applies to unused local variables. It feels strict at first, but it keeps code tidy.

## func main

`main` in package `main` is the entry point. It takes no arguments and returns nothing; when it returns, the program exits. The opening brace must be on the same line as `func`; a brace on its own line is a syntax error, because of how Go inserts semicolons automatically at line ends.

## Calling into a package

`fmt.Println` means "the `Println` function from package `fmt`". Names that start with an **uppercase** letter are *exported*: visible to other packages. Lowercase names are private to their own package. This one rule replaces `public` and `private` keywords.

```go run
package main

import (
	"fmt"
	"strings"
)

func main() {
	name := "gopher"
	fmt.Println("Hello,", strings.ToUpper(name)+"!")
	fmt.Printf("%s has %d letters\n", name, len(name))
}
```

Notice:

- `Println` prints its arguments separated by spaces and adds a newline.
- `Printf` takes a format string. `%s` is a string, `%d` an integer, `%v` any value in its default format, and `\n` is the newline you must add yourself.
- There are no semicolons. The compiler inserts them for you.
- `:=` declares and initialises a variable (covered in the next chapter).

## Comments and layout

```go
// A line comment.

/*
   A block comment.
*/
```

A comment directly above a declaration becomes its documentation, which we use in the next lesson.

## Several files, one package

A package can be split across files. If `main.go` calls a function `greet` defined in `greet.go` in the same folder (and same `package main`), no import is needed; they share one namespace. That is why `go run .` is preferred over naming one file.

```go
// greet.go
package main

func greet(name string) string {
	return "Hello, " + name
}
```

## init functions

A package may also contain `func init()`, which runs automatically before `main`. It is used sparingly, for setup that cannot be done in a declaration. For now, just recognise it when you see it.

## Key takeaways

- Every file begins with `package`; `package main` plus `func main()` makes an executable.
- Imports must all be used, or compilation fails.
- Exported identifiers start with an uppercase letter; lowercase ones stay private to the package.
- `fmt.Println` prints with a newline; `fmt.Printf` formats with verbs like `%s`, `%d`, `%v`.
- No semicolons, and the opening brace goes on the same line.
