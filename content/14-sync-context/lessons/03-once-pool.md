---
title: "Run Once and Reuse: Once and Pool"
minutes: 8
---
Two more members of the `sync` package solve everyday problems: doing something exactly one time, and recycling temporary objects.

## sync.Once

`sync.Once` guarantees that a function runs **exactly once**, no matter how many goroutines call it at the same time. Every caller of `Do` returns only after the one execution has completed, so afterwards the initialised data is safe to use. It is the classic tool for lazy initialisation.

```go run
package main

import (
	"fmt"
	"sync"
)

var (
	once   sync.Once
	config map[string]string
)

func loadConfig() {
	fmt.Println("loading config")
	config = map[string]string{"env": "prod"}
}

func get(key string) string {
	once.Do(loadConfig)
	return config[key]
}

func main() {
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = get("env")
		}()
	}
	wg.Wait()
	fmt.Println(get("env"))
}
```

"loading config" is printed exactly once, followed by `prod`.

Details:

- If the function panics, `Do` still counts it as done; later calls do nothing.
- `Do` with a function that itself calls `once.Do` on the same Once deadlocks.
- Like all sync types, a Once must not be copied.

## OnceFunc, OnceValue, OnceValues (Go 1.21)

Newer helpers wrap this pattern in a function, so you do not need a package-level Once plus a separate variable:

| Function | Returns |
|---|---|
| `sync.OnceFunc(f)` | a `func()` that runs `f` once |
| `sync.OnceValue(f)` | a `func() T` that computes `f()` once and caches the result |
| `sync.OnceValues(f)` | a function returning two cached results, typically `(T, error)` |

If `f` panics, the wrapper panics the same way on every call.

```go run
package main

import (
	"fmt"
	"sync"
)

func main() {
	expensive := sync.OnceValue(func() int {
		fmt.Println("computing...")
		return 6 * 7
	})

	fmt.Println(expensive())
	fmt.Println(expensive()) // cached, no second "computing..."

	hello := sync.OnceFunc(func() { fmt.Println("hello, once") })
	hello()
	hello()
}
```

Output:

```text
computing...
42
42
hello, once
```

`OnceValues` is handy for things like loading a file once and keeping both the data and the error:

```go
var readConfig = sync.OnceValues(func() ([]byte, error) {
	return os.ReadFile("config.json")
})

data, err := readConfig() // reads the file the first time only
```

> **Tip:** the function runs only once even if it fails: an error returned through `OnceValues` is cached too. If you need retry-on-failure, use a mutex and your own flag instead.

## sync.Pool

A `sync.Pool` is a set of temporary objects that can be reused to take pressure off the garbage collector. You `Get` an object (the pool calls `New` if it has none) and `Put` it back when finished.

```go run
package main

import (
	"bytes"
	"fmt"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func render(name string) string {
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset() // always reset a reused object

	buf.WriteString("Hello, ")
	buf.WriteString(name)
	return buf.String()
}

func main() {
	fmt.Println(render("Ann"))
	fmt.Println(render("Bob"))
}
```

Rules of thumb:

- The pool may drop items at any time (it is emptied across garbage collections), so it is a cache of *spare* objects, never storage for state you need.
- Always reset an object before use or before putting it back; it may carry old data.
- Do not pool objects that hold external resources (files, connections).
- Measure first. The standard library itself uses pools for formatting buffers (`fmt`), but for most application code a pool is a premature optimisation.

## Key takeaways

- `sync.Once.Do(f)` runs `f` exactly once and makes all callers wait for it to finish.
- `OnceFunc`, `OnceValue` and `OnceValues` (Go 1.21+) wrap lazy initialisation in a function, with the result cached.
- `sync.Pool` recycles temporary objects: `Get`, reset, use, `Put`.
- Pools can be emptied at any time; use them only for performance, never for correctness.
