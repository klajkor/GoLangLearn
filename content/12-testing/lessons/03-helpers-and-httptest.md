---
title: Helpers, Cleanup and httptest
minutes: 9
---
As tests grow you will want to share setup and assertions. The `testing` package provides a few small tools that make this clean.

## Test helpers and t.Helper

A helper is a function that performs a check or setup on behalf of a test. Call `t.Helper()` at its top, so a failure is reported at the **caller's** line, not inside the helper:

```go
func assertEqual(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestSum(t *testing.T) {
	assertEqual(t, Sum(1, 2), 3) // a failure points here
}
```

Without `t.Helper()`, every failure would point at the `t.Errorf` line inside `assertEqual`, which is useless when ten tests call it. Helpers take `t` as their first parameter. Accept `testing.TB` instead of `*testing.T` if the same helper must work for benchmarks and fuzz tests as well.

## Cleanup

`t.Cleanup(f)` registers a function to run when the test (and its subtests) finish, in last-in-first-out order. It works even when a test fails with `Fatalf`, and it keeps setup and teardown in one place:

```go
func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open("test.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
```

Compare this with `defer db.Close()` inside the test: a `defer` in the *helper* would run when the helper returns, which is too early. `t.Cleanup` runs at the end of the *test*.

## Temporary files and environment

| Helper | Purpose |
|---|---|
| `t.TempDir()` | a fresh directory, removed automatically after the test |
| `t.Setenv(k, v)` | sets an environment variable, restored afterwards (not usable with `t.Parallel`) |
| `t.Chdir(dir)` | changes the working directory for the test (Go 1.24+) |
| `t.Skip(...)`, `testing.Short()` | skip slow tests when `go test -short` is used |

```go
func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")

	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello" {
		t.Errorf("got %q, want %q", b, "hello")
	}
}
```

You never delete the directory yourself, and parallel tests each get their own.

## TestMain

When a whole package needs one-time setup (for example starting a container), define `func TestMain(m *testing.M)`. It runs instead of the default entry point and must call `m.Run()`:

```go
func TestMain(m *testing.M) {
	// setup
	code := m.Run()
	// teardown
	os.Exit(code)
}
```

Use it sparingly; per-test setup with helpers and `t.Cleanup` is easier to reason about.

## Testing HTTP with httptest

The `net/http/httptest` package lets you test handlers without a real network.

**Testing a handler directly**, using `httptest.NewRecorder`, which implements `http.ResponseWriter` and records the response:

```go
func Hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Hello, %s!", name)
}
```

```go
func TestHello(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello?name=Go", nil)
	rec := httptest.NewRecorder()

	Hello(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if string(body) != "Hello, Go!" {
		t.Errorf("body = %q, want %q", body, "Hello, Go!")
	}
}
```

**Testing a client** against a fake server with `httptest.NewServer`. It starts a real server on a random local port and gives you its URL:

```go
func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	// assert on resp ...
}
```

Handler tests are fast and need no ports; `NewServer` is useful when the code under test makes HTTP calls itself.

## Fakes over mocks

Go tests rarely need a mocking framework. Define a small interface where your code consumes a dependency, and implement it with a tiny struct in the test:

```go
type Clock interface{ Now() time.Time }

type fakeClock struct{ t time.Time }

func (f fakeClock) Now() time.Time { return f.t }
```

Passing `fakeClock{...}` makes time-dependent code deterministic.

> **Tip:** keep tests independent. Each test sets up its own data (via helpers and `t.TempDir()`), so tests can run in any order and in parallel.

## Key takeaways

- Call `t.Helper()` in helper functions so failures report the caller's line.
- `t.Cleanup` runs teardown at the end of the test, even after `Fatalf`.
- `t.TempDir()`, `t.Setenv()` and `t.Skip()` handle files, environment and slow tests.
- `httptest.NewRecorder` tests handlers directly; `httptest.NewServer` provides a throwaway server for clients.
- Prefer small interfaces and hand-written fakes to heavy mocking libraries.
