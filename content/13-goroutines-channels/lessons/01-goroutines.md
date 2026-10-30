---
title: Goroutines and the go Statement
minutes: 9
---
Concurrency is one of the reasons people pick Go. A **goroutine** is a function running independently of the code that started it. You create one by putting the keyword `go` in front of a function call.

```go
go doWork(42)             // runs doWork concurrently
go func() { /* ... */ }() // anonymous function, also very common
```

The `go` statement returns immediately. The caller carries on while the new goroutine runs on its own. Any return value of the called function is discarded, so goroutines report back through channels (next lessons) or through shared memory protected by locks (next chapter).

## main does not wait

The program ends when `main` returns, and it does not care whether other goroutines are still running. They are simply killed. So the example below needs a way to wait; here we use a channel, which we will study properly in a moment.

```go run
package main

import "fmt"

func main() {
	done := make(chan struct{})

	go func() {
		fmt.Println("hello from a goroutine")
		close(done) // tell main we are finished
	}()

	<-done // block until the goroutine closes the channel
	fmt.Println("main exits")
}
```

Without the `<-done` line, `main` would usually print nothing from the goroutine, because the program exits first.

> **Common mistake:** using `time.Sleep` to "wait" for goroutines. It is a guess: too short and you lose work, too long and you waste time. Always wait with a real synchronisation tool (a channel or `sync.WaitGroup`).

## Goroutines are cheap

A goroutine starts with a tiny stack (a few KB) that grows and shrinks as needed, and creating one costs far less than creating an OS thread. It is normal for a Go program to run thousands, even hundreds of thousands, of goroutines.

| | OS thread | Goroutine |
|---|---|---|
| Initial stack | ~1 MB (typically fixed) | a few KB, grows on demand |
| Scheduled by | the operating system | the Go runtime |
| Creation cost | high | very low |

## The scheduler in one minute

The Go runtime multiplexes many goroutines onto a small number of OS threads. It uses three kinds of objects, often called **GMP**:

- **G**: a goroutine (its stack and state).
- **M**: a "machine", an OS thread that executes code.
- **P**: a processor, a token that lets an M run Go code and owns a queue of runnable Gs.

The number of Ps defaults to the number of CPU cores (`runtime.GOMAXPROCS(0)` returns it). When a goroutine blocks, for example while waiting on a channel, the scheduler parks it and runs another one on the same thread. Since Go 1.14 the scheduler can also preempt goroutines that run for too long, so a tight loop cannot hog a core forever.

What matters for you: **you do not control the order in which goroutines run.** Never write code whose correctness depends on it.

## Gathering results

Here five goroutines each compute a square and send it back. Because the arrival order is not defined, we collect everything and sort before printing.

```go run
package main

import (
	"fmt"
	"sort"
)

func main() {
	results := make(chan int)

	for i := 1; i <= 5; i++ {
		go func() {
			results <- i * i
		}()
	}

	var squares []int
	for range 5 {
		squares = append(squares, <-results)
	}

	sort.Ints(squares)
	fmt.Println(squares)
}
```

Each goroutine uses `i` directly. That is safe in Go 1.22 and newer, as the last lesson of this chapter explains.

> **Tip:** `for range 5` (range over an integer) also needs Go 1.22+.

## Key takeaways

- `go f(x)` starts `f` concurrently; the arguments are evaluated immediately, in the calling goroutine.
- When `main` returns, the program exits and all other goroutines die with it.
- Goroutines are cheap, runtime-scheduled, and multiplexed onto OS threads (the GMP model).
- Execution order between goroutines is unspecified: synchronise, do not sleep.
- A goroutine cannot return a value; use channels or shared state to get results out.
