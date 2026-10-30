---
title: Channels
minutes: 9
---
Go's concurrency motto is: *"Do not communicate by sharing memory; share memory by communicating."* The tool for that is the **channel**, a typed conduit through which goroutines send and receive values.

## Creating and using channels

```go
ch := make(chan int) // unbuffered channel of int
ch <- 5              // send 5 into the channel
v := <-ch            // receive a value from the channel
```

The arrow shows the direction the data flows. A channel's type is `chan T`, and `make` is required: the zero value of a channel is `nil`, which behaves very differently (see the next lesson).

## Unbuffered channels synchronise

On an unbuffered channel, a send blocks until another goroutine is ready to receive, and a receive blocks until another goroutine sends. The two goroutines **meet** at the channel. That makes an unbuffered channel both a data pipe and a synchronisation point.

```go run
package main

import "fmt"

func main() {
	ping := make(chan string)
	pong := make(chan string)

	go func() {
		msg := <-ping // wait for a message
		pong <- msg + " pong"
	}()

	ping <- "ping"
	fmt.Println(<-pong)
}
```

The output is always `ping pong`: the goroutine cannot answer before it receives, and `main` cannot print before the answer arrives. The channel operations create a strict order, with no sleeping needed.

## Buffered channels

`make(chan T, n)` creates a channel with room for `n` values. A send blocks only when the buffer is full; a receive blocks only when it is empty.

```go run
package main

import "fmt"

func main() {
	ch := make(chan string, 2)

	ch <- "a"
	ch <- "b"
	// A third send here would block forever: the buffer is full.

	fmt.Println(len(ch), cap(ch))
	fmt.Println(<-ch, <-ch)
	fmt.Println(len(ch), cap(ch))
}
```

- `len(ch)` is the number of values currently queued.
- `cap(ch)` is the buffer size (0 for unbuffered).
- Values come out in first-in, first-out order.

Because the sender can run ahead, a buffered channel decouples the two sides a little. It is useful for absorbing short bursts, for limiting concurrency (a buffered channel used as a counting semaphore), and for letting a goroutine hand over a result and exit even if nobody is receiving yet.

> **Common mistake:** adding a buffer to "fix" a deadlock. If the program only works because the buffer happens to be large enough, you have hidden the problem, not solved it. Start with unbuffered channels and add a buffer only for a reason you can state.

## Channels as values

A channel variable is a handle: copying it gives you another handle to the same channel. You can pass channels as arguments, store them in structs, send them over other channels, and compare them with `==`.

A common pattern is a function that starts a goroutine and returns the channel the results will arrive on:

```go run
package main

import "fmt"

func countTo(n int) chan int {
	ch := make(chan int)
	go func() {
		for i := 1; i <= n; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

func main() {
	sum := 0
	for v := range countTo(4) {
		sum += v
	}
	fmt.Println("sum:", sum)
}
```

`close` and `range` over a channel are covered in the next lesson.

## What can go on a channel?

Any type: `chan int`, `chan error`, `chan []byte`, `chan func()`, `chan struct{}`. The empty struct `struct{}` takes no memory and is the idiom for pure signals, where the value carries no information and only the event matters.

## Key takeaways

- `ch <- v` sends, `<-ch` receives; both block until they can proceed.
- Unbuffered channels hand a value over directly and synchronise both goroutines.
- Buffered channels block only when full (send) or empty (receive); `len` and `cap` report the state.
- Channel values are handles; the zero value is `nil`.
- Use `chan struct{}` for signals that carry no data.
