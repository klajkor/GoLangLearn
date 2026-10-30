---
title: Profiling, Escape Analysis and GC Tuning
minutes: 14
---
"Measure, don't guess" is the rule of performance work. Go ships with excellent tools to measure; this lesson tours them. Nothing here runs on the Playground, so the snippets are read-only.

## Benchmark recap

Start with a benchmark (see the testing chapter). `-benchmem` adds allocation counts, which are often the real story:

```go
func BenchmarkJoin(b *testing.B) {
	parts := []string{"a", "b", "c", "d"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = strings.Join(parts, ",")
	}
}
```

```bash
go test -bench=Join -benchmem -count=5
go test -bench=. -cpuprofile cpu.out -memprofile mem.out
```

Compare runs with `benchstat` (from `golang.org/x/perf`) instead of eyeballing numbers.

## Kinds of profile

| Profile | Answers | Notes |
|---|---|---|
| CPU | where is CPU time spent? | sampled about 100 times per second while running |
| Heap | who allocated memory that is still live (`inuse_*`) or allocated overall (`alloc_*`)? | sampled by allocation size |
| Goroutine | what are all goroutines doing right now? | great for leaks and deadlocks |
| Block | where do goroutines wait on channels and mutexes? | off by default: `runtime.SetBlockProfileRate` |
| Mutex | who holds contended locks? | off by default: `runtime.SetMutexProfileFraction` |

## Collecting profiles

In a long-running service, import `net/http/pprof` for its side effect and serve it on a **private** address:

```go
import (
	"log"
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/* on http.DefaultServeMux
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// ... your real server, on its own mux ...
}
```

> **Warning:** pprof endpoints expose internals and can be used to slow your service down. Never put them on a public listener.

In a command-line tool, use `runtime/pprof` directly:

```go
f, _ := os.Create("cpu.out")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
```

## Reading a profile

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof -http=:8081 cpu.out        # browser UI with flame graph
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

Inside the interactive prompt, `top` shows the heaviest functions:

```text
(pprof) top
Showing nodes accounting for 1.80s, 90% of 2.00s total
      flat  flat%   sum%        cum   cum%
     0.90s 45.00% 45.00%      0.90s 45.00%  runtime.memmove
     0.50s 25.00% 70.00%      1.40s 70.00%  main.buildReport
     0.40s 20.00% 90.00%      0.40s 20.00%  strings.(*Builder).WriteString
```

- **flat**: time spent in the function itself.
- **cum** (cumulative): time in the function plus everything it called.

A high `cum` with a low `flat` means "look at what it calls". `list buildReport` prints the source with per-line costs, and `web` or the `-http` UI draws the call graph. In a **flame graph**, each box is a function, its width is its share of samples, and callees stack on top of callers: hunt for wide boxes at the top edge. Use `-sample_index=alloc_space` on a heap profile to see total allocations rather than live memory.

## The execution tracer

Profiles are statistical summaries. The tracer records individual events (goroutine scheduling, GC, syscalls, blocking) on a timeline, which helps with latency and concurrency questions such as "why is my CPU idle?".

```bash
go test -trace trace.out
go tool trace trace.out
```

Add the `/debug/pprof/trace?seconds=5` endpoint for live services. Traces are large; capture a few seconds only.

## Escape analysis and fewer allocations

The compiler decides whether a variable lives on the stack (cheap, freed automatically) or **escapes** to the heap (costs allocation and GC work). Ask it why:

```bash
go build -gcflags='-m' ./...
```

```text
./main.go:9:6: moved to heap: x
./main.go:14:13: ... argument does not escape
```

Typical reasons for escaping: returning a pointer to a local, storing in an interface or a captured closure variable, or sending a pointer through a channel. Practical tips:

- Preallocate: `make([]T, 0, n)` and `make(map[K]V, n)` when you know the size.
- Reuse buffers with `strings.Builder`, `bytes.Buffer`, or `sync.Pool` (for short-lived, same-shaped objects).
- Avoid `fmt.Sprintf` in hot paths; use `strconv` or append-style functions.
- Pass small structs by value; pointers do not always mean faster.
- Do not trade clarity for allocations you have not measured.

## GC tuning basics

Go's garbage collector is concurrent and tuned by two knobs, both settable as environment variables or in code:

| Knob | Meaning |
|---|---|
| `GOGC` (default 100) | next GC starts when the heap has grown by this percent over the live data after the last GC. Higher: fewer GCs, more memory. `GOGC=off` disables it. |
| `GOMEMLIMIT` (Go 1.19) | soft total-memory cap, such as `GOMEMLIMIT=900MiB`. The GC works harder as it nears the limit. |

A common container recipe: set `GOMEMLIMIT` to roughly 90 percent of the container's memory limit so the GC reacts before the OOM killer does. Check effects with `GODEBUG=gctrace=1`. Most services need neither; reducing allocations beats tuning.

> **Common mistake:** optimizing from a CPU profile of a benchmark that does not look like production traffic. Profile realistic load.

## Key takeaways

- Benchmark with `-benchmem`; profile with `-cpuprofile`/`-memprofile` or `net/http/pprof`.
- `go tool pprof`: `top` (flat versus cum), `list`, and the `-http` flame graph.
- Block and mutex profiles are off until you enable them.
- `go tool trace` shows a timeline; pprof shows aggregates.
- `-gcflags=-m` explains escapes; fewer allocations usually win more than GC tuning.
- `GOGC` trades memory for CPU; `GOMEMLIMIT` is a soft cap.
