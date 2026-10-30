---
title: "Closing, Directions and nil Channels"
minutes: 10
---
This lesson covers the finer points of channels: telling receivers that no more values are coming, restricting what a function may do with a channel, and what a `nil` channel is good for.

## Closing a channel

`close(ch)` says: "I will send nothing more." Rules worth memorising:

- Only the **sender** side should close a channel, never a receiver.
- Sending on a closed channel **panics**. Closing a closed channel panics. Closing a `nil` channel panics.
- Receiving from a closed channel never blocks: you first get any values still in the buffer, then the zero value forever.
- Closing is optional. You only need it when receivers must be told that the stream ended (for example so `range` stops). The garbage collector reclaims unused channels anyway.

## Receive with ok

How do you tell a real zero from "channel closed"? Use the two-value receive form:

```go
v, ok := <-ch
```

`ok` is `true` if `v` is a real sent value and `false` if the channel is closed and drained.

```go run
package main

import "fmt"

func main() {
	ch := make(chan int, 2)
	ch <- 10
	ch <- 20
	close(ch)

	v, ok := <-ch
	fmt.Println(v, ok)
	v, ok = <-ch
	fmt.Println(v, ok)
	v, ok = <-ch // closed and empty
	fmt.Println(v, ok)
}
```

## Range over a channel

`for v := range ch` receives until the channel is closed and empty. It is the idiomatic consumer loop.

```go run
package main

import "fmt"

func producer(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // runs when the goroutine finishes
		for i := 1; i <= n; i++ {
			out <- i * 10
		}
	}()
	return out
}

func main() {
	for v := range producer(3) {
		fmt.Println(v)
	}
	fmt.Println("done")
}
```

> **Common mistake:** forgetting to close the channel. The `range` loop then blocks forever once the producer stops, and the program dies with a deadlock (or leaks a goroutine).

## Directional channel types

A parameter can say whether the function may only send or only receive:

| Type | Meaning |
|---|---|
| `chan T` | send and receive |
| `chan<- T` | send-only |
| `<-chan T` | receive-only |

A bidirectional channel converts implicitly to a directional one, but not the other way round. The compiler then enforces your design: a consumer cannot accidentally send, and `close` is not allowed on a receive-only channel, which neatly encodes "only the sender closes".

```go run
package main

import "fmt"

func produce(out chan<- int) {
	for i := 1; i <= 3; i++ {
		out <- i
	}
	close(out)
}

func consume(in <-chan int) int {
	total := 0
	for v := range in {
		total += v
	}
	return total
}

func main() {
	ch := make(chan int)
	go produce(ch)
	fmt.Println(consume(ch))
}
```

Reading signatures becomes easier too: `func gen() <-chan int` tells callers "you can only read what I give you".

## nil channels

A channel variable that was never made is `nil`:

```go
var ch chan int // nil
```

On a nil channel, **send and receive block forever**, and `close` panics. That sounds useless, but in a `select` statement (next chapter) a blocked case is simply never chosen, so setting a channel variable to `nil` is a way to *switch a case off*. You will use this trick to merge channels.

Outside `select`, a nil channel is almost always a bug: you forgot `make`.

```go
var results chan int
results <- 1 // blocks forever -> "fatal error: all goroutines are asleep - deadlock!"
```

## Key takeaways

- The sender closes; receivers detect it with `v, ok := <-ch` or `range`.
- Sending on or double-closing a closed channel panics; receiving from one yields zero values.
- Use `chan<- T` and `<-chan T` in signatures to document and enforce intent.
- A nil channel blocks forever on send and receive; it is a bug unless used deliberately inside `select`.
