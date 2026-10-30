---
title: "Data Races and Loop Variables"
minutes: 9
---
Goroutines that share memory need care. This lesson introduces the **data race**, the tool that finds it, and a closely related gotcha: how loop variables behave inside goroutines.

## What is a data race?

A data race happens when two goroutines access the same variable at the same time, at least one of them writes, and there is no synchronisation. The result is unpredictable in practice: lost updates, corrupted values, and bugs that disappear when you add a `Println`.

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	count := 0
	var wg sync.WaitGroup
	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count++ // read, add, write: three steps, not atomic
		}()
	}
	wg.Wait()
	fmt.Println(count) // often less than 1000, and different on each run
}
```

```text
943
```

`count++` is really "load, add one, store". Two goroutines can load the same value and both store the same result, losing one increment. (`sync.WaitGroup` is explained in the next chapter.)

## The race detector

Go ships a race detector. Add `-race` to run, build or test:

```bash
go run -race main.go
go test -race ./...
```

When it sees a race at runtime it prints a report with the two conflicting accesses and the goroutines that made them:

```text
==================
WARNING: DATA RACE
Read at 0x00c000012345 by goroutine 8:
  main.main.func1()
      /tmp/main.go:14 +0x...

Previous write at 0x00c000012345 by goroutine 7:
  main.main.func1()
      /tmp/main.go:14 +0x...
==================
```

Things to know:

- It finds races that **actually happen** during that run; it does not prove their absence. Run it on tests that exercise concurrent paths.
- It makes programs several times slower and hungrier for memory, so use it in development and CI, not in production builds.
- It requires cgo support on most platforms. Treat every report as a real bug.

## Ways to fix a race

1. **Don't share**: keep data owned by one goroutine and pass values or results through channels.
2. **Lock**: protect shared data with a `sync.Mutex` (next chapter).
3. **Atomics**: for single counters and flags use `sync/atomic`.

Here is a channel-based fix, where only one goroutine ever touches `count`:

```go run
package main

import "fmt"

func main() {
	inc := make(chan struct{})
	done := make(chan int)

	go func() { // the single owner of count
		count := 0
		for range inc {
			count++
		}
		done <- count
	}()

	for range 1000 {
		inc <- struct{}{}
	}
	close(inc)
	fmt.Println(<-done)
}
```

This always prints `1000`. Each increment is a message, and the owner handles one message at a time.

## Loop variables and goroutines (Go 1.22)

Before Go 1.22 a `for` loop declared **one** variable shared by all iterations. Closures that captured it all saw its final value, a famous bug:

```go
for i := 0; i < 3; i++ {
	go func() { fmt.Println(i) }() // old Go: usually prints 3 3 3
}
```

Since **Go 1.22**, every iteration has its **own copy** of the loop variable, for both three-clause loops and `range` loops. The closure above now captures a fresh `i` each time.

```go run
package main

import "fmt"

func main() {
	var funcs []func() int
	for i := 0; i < 3; i++ {
		funcs = append(funcs, func() int { return i })
	}
	for _, f := range funcs {
		fmt.Print(f(), " ")
	}
	fmt.Println()
}
```

The program prints `0 1 2`. Under the old semantics it printed `3 3 3`.

> **Important:** the new behaviour depends on the `go` line in your `go.mod`. A module declaring `go 1.21` or lower keeps the old semantics even when built with a newer toolchain. Old code often contains the workaround `i := i` as the first line of the loop body; it is harmless but no longer needed with `go 1.22` or newer.

Per-iteration copies only fix the captured loop variable itself. A goroutine that modifies some *other* shared variable declared outside the loop is still a race.

## Key takeaways

- A data race is an unsynchronised concurrent access where at least one side writes; its effects are unpredictable.
- `go run -race`, `go test -race` and `go build -race` detect races that occur at runtime; use them in tests and CI.
- Fix races by not sharing, by locking, or by using atomics.
- Go 1.22+ gives each loop iteration its own loop variable; this depends on the `go` version in `go.mod`.
