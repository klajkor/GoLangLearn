---
title: "select, Timeouts and Tickers"
minutes: 10
---
A goroutine often has to wait on **several** things at once: a result, a timeout, a cancellation signal. The `select` statement does exactly that. It looks like a `switch`, but every case is a channel operation.

```go
select {
case v := <-a:
	fmt.Println("got", v)
case b <- 5:
	fmt.Println("sent")
case <-time.After(time.Second):
	fmt.Println("timeout")
default:
	fmt.Println("nothing ready")
}
```

How it works:

- `select` blocks until at least one case can proceed, then runs that case.
- If **several** are ready, it picks one **at random**. Never assume priority by order.
- With a `default` case, `select` never blocks: if nothing is ready, `default` runs. That gives you a non-blocking send or receive.
- `select {}` with no cases blocks forever.

## Waiting for the first of several

```go run
package main

import "fmt"

func main() {
	a := make(chan string, 1)
	b := make(chan string, 1)
	b <- "from b" // only b has a value

	select {
	case v := <-a:
		fmt.Println(v)
	case v := <-b:
		fmt.Println(v)
	}

	// Non-blocking receive with default.
	select {
	case v := <-a:
		fmt.Println(v)
	default:
		fmt.Println("a is empty")
	}
}
```

## Timeouts with time.After

`time.After(d)` returns a channel that receives one value after `d`. Combined with `select`, it limits how long you wait:

```go run
package main

import (
	"fmt"
	"time"
)

func slowWorker() <-chan string {
	ch := make(chan string, 1) // buffered: the worker can finish even if nobody listens
	go func() {
		time.Sleep(500 * time.Millisecond)
		ch <- "finished"
	}()
	return ch
}

func main() {
	select {
	case res := <-slowWorker():
		fmt.Println(res)
	case <-time.After(10 * time.Millisecond):
		fmt.Println("timeout: gave up waiting")
	}
}
```

The timeout (10 ms) fires long before the worker (500 ms), so the result is certain. Notice the worker's channel is buffered; otherwise the abandoned worker would be a goroutine leak.

> **Tip:** in a loop, `time.After` creates a new timer on every iteration. Before Go 1.23 unreferenced timers lingered until they fired; since 1.23 they are garbage-collected, but for hot loops a single reused `time.Timer` is still cleaner. For cancellation prefer `context` (next lesson).

## The nil-channel trick

A `nil` channel is never ready, so a `select` case on it is skipped. Set a channel variable to `nil` when it is finished, and the select ignores it. This is how you merge two channels until **both** are closed:

```go run
package main

import (
	"fmt"
	"sort"
)

func emit(vals ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range vals {
			ch <- v
		}
	}()
	return ch
}

func main() {
	a := emit(1, 2, 3)
	b := emit(10, 20)

	var got []int
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil // closed: never select this case again
				continue
			}
			got = append(got, v)
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			got = append(got, v)
		}
	}

	sort.Ints(got)
	fmt.Println(got)
}
```

Without the `nil` assignment, a closed channel would be "ready" forever (it returns zero values), and the loop would spin.

## time.Timer and time.Ticker

- A **`time.Timer`** fires once. `time.NewTimer(d)` gives a timer with a channel `C`; `Stop()` cancels it; `Reset(d)` re-arms it. `time.AfterFunc(d, f)` runs `f` in its own goroutine after `d`.
- A **`time.Ticker`** fires repeatedly at an interval. **Always call `Stop`** when done (use `defer`), or it keeps ticking until the program ends (before Go 1.23 it could not even be garbage-collected).

```go run
package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	count := 0
	for range ticker.C {
		count++
		fmt.Println("tick", count)
		if count == 3 {
			break
		}
	}
	fmt.Println("done")
}
```

Tickers drop ticks for slow receivers instead of queueing them up, so do not rely on counting ticks for exact elapsed time.

## A common loop shape

A worker loop that processes jobs until told to stop:

```go
for {
	select {
	case job, ok := <-jobs:
		if !ok {
			return // no more work
		}
		handle(job)
	case <-quit:
		return
	}
}
```

## Key takeaways

- `select` waits on several channel operations and runs one that is ready; ties are broken randomly.
- `default` makes a `select` non-blocking.
- Timeouts: `case <-time.After(d)`; prefer `context` for cancellation across call chains.
- Assign `nil` to a channel variable to disable its case, for example after it is closed.
- `Ticker`s must be stopped; a `Timer` fires once and can be `Stop`ped or `Reset`.
