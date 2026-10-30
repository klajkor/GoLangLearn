---
title: Mutexes and Atomics
minutes: 10
---
Channels are great for passing ownership of data. Sometimes, though, several goroutines really do need to read and update one shared structure: a cache, a counter, a map of sessions. Then you protect it with a lock.

## sync.Mutex

A `sync.Mutex` is a mutual-exclusion lock: only one goroutine can hold it at a time. Everyone else calling `Lock` waits.

```go run
package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

func main() {
	var c Counter
	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				c.Inc()
			}
		}()
	}

	wg.Wait()
	fmt.Println(c.Value())
}
```

It always prints `100000`. The code between `Lock` and `Unlock` is the **critical section**.

Conventions that keep mutex code healthy:

- Put the mutex **next to the data it guards**, in the same struct, and say so in a comment.
- Use `defer mu.Unlock()` right after `Lock()` so every return path unlocks.
- Keep critical sections short; never call unknown code or do slow I/O while holding a lock.
- The zero value is an unlocked mutex, ready to use. **Never copy** a mutex after first use: use pointer receivers, as above.
- A Go mutex is **not reentrant**: locking it twice from the same goroutine deadlocks.

> **Common mistake:** locking only the writes. A read that happens while another goroutine writes is still a data race, so every access to the guarded data must hold the lock.

## sync.RWMutex

When reads vastly outnumber writes, `sync.RWMutex` allows many readers at once but only one writer, who has the data to itself:

| Method | Meaning |
|---|---|
| `RLock` / `RUnlock` | shared (read) lock; many holders at once |
| `Lock` / `Unlock` | exclusive (write) lock |

```go run
package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]int
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]int)}
}

func (c *Cache) Get(k string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok
}

func (c *Cache) Set(k string, v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
}

func main() {
	c := NewCache()
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Set(fmt.Sprint("key", i), i*i)
		}()
	}
	wg.Wait()

	v, ok := c.Get("key3")
	fmt.Println(v, ok)
	_, ok = c.Get("missing")
	fmt.Println(ok)
}
```

Maps are not safe for concurrent use: a concurrent write and read can crash the program with `fatal error: concurrent map read and map write`. Guard them with a mutex (the standard library also has `sync.Map`, intended for a few special cases such as write-once, read-many caches; a plain map plus a mutex is usually clearer).

An `RWMutex` has more overhead than a `Mutex`; prefer the plain `Mutex` unless profiling shows read contention.

## sync/atomic

For a single number or flag, a lock is overkill. The `sync/atomic` package performs indivisible operations at the hardware level. Since Go 1.19 there are convenient typed values:

```go run
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var hits atomic.Int64
	var ready atomic.Bool
	var wg sync.WaitGroup

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				hits.Add(1)
			}
		}()
	}
	wg.Wait()
	ready.Store(true)

	fmt.Println(hits.Load(), ready.Load())

	// Compare-and-swap: set to 99 only if the current value is 10000.
	swapped := hits.CompareAndSwap(10000, 99)
	fmt.Println(swapped, hits.Load())
}
```

The typed values (`atomic.Int32`, `Int64`, `Uint32`, `Uint64`, `Bool`, `Pointer[T]`, `Value`) have methods `Load`, `Store`, `Add` (numbers), `Swap` and `CompareAndSwap`.

Atomics protect **one variable**. If two fields must change together to stay consistent, use a mutex.

## Which tool?

| Situation | Use |
|---|---|
| Passing data or events between goroutines | channel |
| Protecting a struct or map with several fields | `Mutex` / `RWMutex` |
| A counter, flag or a pointer swap | `sync/atomic` |
| Detecting mistakes | `go test -race` |

## Key takeaways

- A `Mutex` guards shared data: `Lock`, then `defer Unlock`, and keep the critical section small.
- Never copy a mutex; it is not reentrant; guard every access, reads included.
- `RWMutex` lets many readers in at once; use it only when reads dominate.
- Plain maps are not goroutine-safe; wrap them with a lock.
- `atomic.Int64`, `atomic.Bool` and friends handle single-variable updates without a lock.
