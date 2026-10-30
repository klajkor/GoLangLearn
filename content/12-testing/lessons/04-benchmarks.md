---
title: Benchmarks
minutes: 9
---
Benchmarks measure how fast code runs and how much it allocates. They live in `_test.go` files next to your tests and are named **`BenchmarkXxx`** with the signature `func(b *testing.B)`.

## The classic form

```go
func BenchmarkConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = strings.Repeat("go", 100)
	}
}
```

The framework chooses `b.N`: it runs your function with a small `N`, then increases it until the measurement is stable (about a second by default). The loop **must** run the code `b.N` times. The result is reported as nanoseconds per operation.

## Running benchmarks

Benchmarks do not run during a normal `go test`. Ask for them with `-bench`, which takes a regular expression (`.` means all):

```bash
go test -bench=.
go test -bench=Concat -benchmem
go test -bench=. -run='^$'          # skip the unit tests
go test -bench=. -benchtime=3s      # run longer
go test -bench=. -count=5           # repeat, to see variance
```

Sample output:

```text
goos: linux
goarch: amd64
pkg: example.com/strx
cpu: Example CPU
BenchmarkConcat-8     8504617    141.2 ns/op    208 B/op    1 allocs/op
PASS
ok  	example.com/strx	1.352s
```

Reading the line: `-8` is `GOMAXPROCS`, `8504617` is the final `b.N`, `141.2 ns/op` is time per operation. The last two columns, `B/op` and `allocs/op`, appear with **`-benchmem`** (or `b.ReportAllocs()`) and show heap bytes and allocation counts per operation. Often allocations matter more than raw time.

## Setup and timers

Expensive setup should not be counted. Use `b.ResetTimer()` after it:

```go
func BenchmarkSort(b *testing.B) {
	data := makeBigSlice() // setup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := slices.Clone(data)
		slices.Sort(tmp)
	}
}
```

`b.StopTimer()` and `b.StartTimer()` pause and resume timing, but they add overhead. Prefer restructuring.

## The compiler can optimise your benchmark away

If the result of the code under test is unused and the compiler can see that the function has no side effects, it may delete the call. Protect the result by assigning it to a package-level sink variable:

```go
var sink int

func BenchmarkFib(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		r = Fib(20)
	}
	sink = r
}
```

## b.Loop (Go 1.24+)

Newer Go versions provide `b.Loop()`, which is shorter and avoids several of those pitfalls. Setup before the loop is excluded from timing automatically, and the compiler keeps the loop body alive:

```go
func BenchmarkFib(b *testing.B) {
	for b.Loop() {
		Fib(20)
	}
}
```

Prefer `for b.Loop()` when your Go version supports it; the `for i := 0; i < b.N; i++` form remains valid and is common in existing code.

## Sub-benchmarks

`b.Run` works like `t.Run` and is perfect for comparing input sizes or implementations:

```go
func BenchmarkJoin(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		parts := make([]string, n)
		for i := range parts {
			parts[i] = "x"
		}
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			for b.Loop() {
				strings.Join(parts, ",")
			}
		})
	}
}
```

## Benchmarks in a runnable program

You normally need `go test` for benchmarks, but `testing.Benchmark` and `testing.AllocsPerRun` can be called from `main`. Timings differ on every machine, so this program only prints stable facts:

```go run
package main

import (
	"fmt"
	"testing"
)

var sink []int

func noAlloc(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}

func oneAlloc() {
	sink = make([]int, 128) // stored in a global, so it is heap allocated
}

func main() {
	xs := []int{1, 2, 3, 4}

	fmt.Println("noAlloc allocs:", testing.AllocsPerRun(100, func() { noAlloc(xs) }))
	fmt.Println("oneAlloc allocs:", testing.AllocsPerRun(100, oneAlloc))

	res := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			noAlloc(xs)
		}
	})
	fmt.Println("benchmark ran:", res.N > 0)
}
```

`AllocsPerRun` reports the average number of heap allocations per call, which is great for asserting "this hot path must not allocate" inside an ordinary test.

## Comparing results

Single runs are noisy. Run each version several times (`-count=10`) and compare with `benchstat` (`golang.org/x/perf/cmd/benchstat`), which reports whether a difference is statistically meaningful. Also close other heavy programs, and keep the laptop plugged in.

> **Common mistake:** benchmarking before you know there is a problem. Profile first (`go test -bench=. -cpuprofile cpu.out` then `go tool pprof cpu.out`) and optimise the part that matters.

## Key takeaways

- Benchmarks are `BenchmarkXxx(b *testing.B)` and run only with `-bench`.
- Loop `b.N` times, or use `for b.Loop()` in Go 1.24+.
- `-benchmem` adds `B/op` and `allocs/op`, which are often the key numbers.
- Use `b.ResetTimer()` after setup, and a sink variable to stop dead-code elimination.
- `b.Run` creates sub-benchmarks; `testing.AllocsPerRun` checks allocation counts.
- Repeat runs and compare with `benchstat` before trusting a difference.
