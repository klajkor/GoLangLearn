---
title: Deadlocks and Goroutine Leaks
minutes: 9
---
Two failure modes dominate concurrent Go code: goroutines that wait forever and the runtime notices (a **deadlock**), and goroutines that wait forever while the program carries on (a **leak**).

## Deadlock

The runtime watches for the situation where *every* goroutine is blocked and none can ever be woken. It then aborts the program. This is the simplest possible deadlock:

```go
package main

func main() {
	ch := make(chan int)
	ch <- 1 // nobody will ever receive: main blocks forever
}
```

```text
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/tmp/main.go:5 +0x...
exit status 2
```

The goroutine trace tells you what each goroutine is waiting on (`chan send`, `chan receive`, `semacquire`, ...). Read it first.

Typical causes:

| Cause | Fix |
|---|---|
| Unbuffered send with no receiver | start the receiver in a goroutine first |
| `range` over a channel nobody closes | the sender calls `close` |
| Receiving from a `nil` channel | `make` the channel |
| `wg.Wait()` while the counter never reaches zero | match every `Add` with a `Done` |
| Two goroutines each waiting for the other | agree on one direction or lock order |

> **Note:** the runtime only detects deadlocks where *all* goroutines are stuck. If anything else is alive (a timer, a network listener, another healthy goroutine), a stuck goroutine just hangs silently.

## Goroutine leaks

A leaked goroutine is one that never finishes, typically blocked on a channel that nobody will ever use again. It keeps its stack and everything it references alive. A few leaks are harmless in a short-lived tool, but in a server they pile up until memory runs out.

```go
func firstResult(urls []string) string {
	ch := make(chan string) // unbuffered
	for _, u := range urls {
		go func() { ch <- fetch(u) }()
	}
	return <-ch // returns after the first answer...
} // ...the other goroutines block on send forever
```

Only one value is ever received; every other goroutine is stuck on `ch <- ...` for good. Two common cures:

1. **Buffer the channel** so every sender can complete: `make(chan string, len(urls))`.
2. **Give goroutines a way to stop**, using cancellation with `context` (next chapter).

Here is the buffered idea in runnable form. Every sender has room in the buffer, so none can get stuck; we sort the results because arrival order is unspecified.

```go run
package main

import (
	"fmt"
	"sort"
)

func collectAll(names []string) []string {
	ch := make(chan string, len(names)) // room for every sender
	for _, n := range names {
		go func() { ch <- "hello " + n }() // never blocks
	}
	var out []string
	for range names {
		out = append(out, <-ch)
	}
	sort.Strings(out)
	return out
}

func main() {
	fmt.Println(collectAll([]string{"ann", "bob", "cy"}))
}
```

## A rule for every goroutine you start

Before writing `go`, answer: **how does this goroutine end?** And: **who waits for it?** If you cannot answer both, you are probably creating a leak. Typical exits are:

- the work loop finishes;
- the input channel is closed, so `range` ends;
- a `ctx.Done()` or "quit" channel fires.

## Spotting leaks

- `runtime.NumGoroutine()` returns the current count; a number that keeps climbing is a red flag.
- Send `SIGQUIT` to a running program, or use the `net/http/pprof` goroutine profile, to dump all goroutine stacks and see where they are parked.
- In tests, libraries such as `go.uber.org/goleak` fail a test if goroutines remain at the end.

## Key takeaways

- "all goroutines are asleep - deadlock!" means every goroutine is blocked; read the trace to see on what.
- A leak is a goroutine blocked forever while the program lives on; it is not detected automatically.
- For every goroutine, know how it stops and who waits for it.
- Buffering a result channel or adding cancellation are the usual cures for abandoned senders.
