---
title: A Standard Library Tour
minutes: 9
---
Go ships with a large, high-quality standard library. Before reaching for a third-party dependency, check whether the standard library already covers the job. Every package below needs only an `import`, no `go get`.

## A map of the territory

| Area | Packages |
|---|---|
| Text and strings | `strings`, `strconv`, `bytes`, `unicode`, `regexp`, `text/template` |
| Collections and sorting | `slices`, `maps`, `sort`, `container/list`, `container/heap` |
| Math | `math`, `math/rand/v2`, `math/big`, `math/bits` |
| I/O and files | `io`, `bufio`, `os`, `io/fs`, `path/filepath`, `embed` |
| Time | `time` |
| Encoding | `encoding/json`, `encoding/csv`, `encoding/xml`, `encoding/base64` |
| Networking | `net`, `net/http`, `net/http/httptest`, `net/url` |
| Concurrency | `sync`, `sync/atomic`, `context` |
| Errors and logging | `errors`, `log`, `log/slog` |
| Testing | `testing`, `testing/fstest`, `testing/quick` |
| Crypto and hashing | `crypto/sha256`, `crypto/rand`, `hash/fnv` |

The two packages you will use the most are `fmt` and `strings`, followed closely by `errors`, `os` and `time`.

## Discovering documentation

You do not need a browser to read the docs:

```bash
go doc strings.Fields
go doc -all time.Duration
go doc encoding/json
```

The same documentation is on [pkg.go.dev](https://pkg.go.dev/std). Reading the standard library source is also encouraged; it is clear, idiomatic Go.

## Strings, slices and sorting

The `strings` package covers almost everything text-related, and the generic `slices` package (Go 1.21+) provides sorting, searching and comparing for any slice type.

```go run
package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

func main() {
	words := strings.Fields("the quick brown fox jumps over the lazy dog")
	fmt.Println(len(words), "words")

	slices.Sort(words)
	words = slices.Compact(words) // removes consecutive duplicates
	fmt.Println(words)

	i, found := slices.BinarySearch(words, "fox")
	fmt.Println("fox at", i, found)

	people := []struct {
		Name string
		Age  int
	}{{"Ana", 31}, {"Bo", 25}, {"Cy", 40}}
	sort.Slice(people, func(i, j int) bool { return people[i].Age < people[j].Age })
	fmt.Println(people)

	fmt.Println(strings.Repeat("=", 10))
	fmt.Println(strings.Join([]string{"a", "b", "c"}, "-"))
}
```

## Time, conversion and math

`time.Duration` is an `int64` count of nanoseconds that prints in a human-friendly form. You can do arithmetic and conversions without any clock access, which makes the code deterministic. `strconv` converts between strings and numbers and reports errors rather than guessing.

```go run
package main

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func main() {
	d := 90*time.Minute + 30*time.Second
	fmt.Println(d)
	fmt.Println(d.Hours(), "hours")
	fmt.Println(d.Round(time.Hour))

	fmt.Println(2500 * time.Millisecond)

	n, err := strconv.Atoi("42")
	fmt.Println(n+1, err)

	_, err = strconv.Atoi("forty-two")
	fmt.Println(err)

	fmt.Printf("%.4f\n", math.Pi)
	fmt.Println(math.Floor(-2.5), math.Ceil(2.1), math.Abs(-7))
}
```

> **Tip:** `time.Duration` constants like `time.Second` are typed. Write `5 * time.Second`, never `5 * 1000` with a mystery unit.

## A peek at HTTP

`net/http` is production-grade: many Go services use it directly without a framework. A minimal server looks like this (shown read-only because it needs a network):

```go
package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "hello from Go")
    })
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Choosing between standard library and dependencies

- Start with the standard library; it is stable and covered by the Go 1 compatibility promise.
- Add a third-party module when it clearly saves effort and is well maintained (for example a database driver or a specialised router).
- Each dependency adds update, security and review cost. `go mod tidy` and `go list -m all` help you see what you carry.

The `golang.org/x/...` modules (such as `x/text`, `x/sync`, `x/tools`) are maintained by the Go team but live outside the standard library, so they are not under the compatibility promise and are added with `go get`.

## Key takeaways

- The standard library is broad; check it before adding a dependency.
- Use `go doc` for fast, offline documentation.
- `slices`, `maps`, `strings`, `strconv`, `time` and `sort` cover everyday data handling.
- `time.Duration` is a typed int64 of nanoseconds with readable printing.
- `golang.org/x/...` packages are official but outside the compatibility promise.
