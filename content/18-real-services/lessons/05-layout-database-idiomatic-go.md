---
title: Project Layout, database/sql and Idiomatic Go
minutes: 12
---
This last lesson ties the course together: how to organize a service, a read-only overview of talking to a database, and a checklist of what "idiomatic Go" means.

## Project layout

Go does not mandate a layout, and small projects should stay small. A single `main.go` is a fine start. When a service grows, this structure is common and has real compiler support:

```text
myservice/
  go.mod
  cmd/
    myservice/
      main.go          # tiny: parse config, wire dependencies, run
    migrate/
      main.go          # a second binary
  internal/
    config/            # loading settings
    httpapi/           # handlers, routes, middleware
    store/             # persistence
    orders/            # domain logic
  README.md
```

- **`cmd/<name>/main.go`**: one directory per binary, with as little code as possible.
- **`internal/`**: the compiler **forbids** importing packages under `internal/` from outside the parent module tree. It is how you keep implementation details private without worrying about someone depending on them.
- Avoid a catch-all `pkg/` or `utils` / `common` packages. Name packages for what they provide (`orders`, `httpapi`), short and lower case, with no underscores.
- Avoid import cycles by letting the **consumer** define the interface it needs. A domain package declares `type Store interface{...}`; the `store` package implements it without importing the domain.
- Wire everything in `main` ("composition root"): build config, logger, store, server; pass them down explicitly.

## database/sql overview (read-only)

`database/sql` is a driver-agnostic API. You import a driver for its side effects, and open a `*sql.DB`, which is a **connection pool** that is safe for concurrent use. Create it once, share it, and close it on shutdown. These snippets need a database, so they are read-only:

```go
import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
)

db, err := sql.Open("pgx", dsn) // does not connect yet
if err != nil {
	return err
}
defer db.Close()
db.SetMaxOpenConns(25)
db.SetConnMaxLifetime(30 * time.Minute)

ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()
if err := db.PingContext(ctx); err != nil { // verifies connectivity
	return err
}
```

Querying a single row and a list:

```go
var name string
err := db.QueryRowContext(ctx, `SELECT name FROM users WHERE id = $1`, id).Scan(&name)
if errors.Is(err, sql.ErrNoRows) {
	return ErrNotFound // translate to a domain error
}

rows, err := db.QueryContext(ctx, `SELECT id, name FROM users WHERE active = $1`, true)
if err != nil {
	return nil, err
}
defer rows.Close()
var users []User
for rows.Next() {
	var u User
	if err := rows.Scan(&u.ID, &u.Name); err != nil {
		return nil, err
	}
	users = append(users, u)
}
return users, rows.Err() // always check Err after the loop
```

Rules to remember:

- **Always use placeholders** (`$1` or `?`, depending on the driver) and never build SQL with string concatenation: that is how SQL injection happens.
- Use the `...Context` methods so queries stop when the request is cancelled.
- Close `rows`, and check `rows.Err()`.
- Use `db.BeginTx` for transactions; `defer tx.Rollback()` right after beginning (it is a no-op after `Commit`).
- Convert `sql.ErrNoRows` into your own error; do not leak it past the store layer.
- Libraries such as `sqlx`, `pgx` (native API), and generators like `sqlc` build on this foundation.

## The idiomatic Go checklist

**Naming and style**
- Run `gofmt` (or `goimports`) always; do not argue about formatting.
- Short names in small scopes (`i`, `r`, `ctx`); descriptive names for exported identifiers. No `Get` prefix on getters (`user.Name()`), no stuttering (`orders.Order`, not `orders.OrderOrder`).
- Acronyms keep one case: `ServeHTTP`, `userID`, `URL`.
- Doc comments start with the identifier name.

**Errors**
- Return errors; do not panic for expected failures.
- Wrap with context using `%w`; compare with `errors.Is` / `errors.As`.
- Handle an error once: either log it or return it, not both.

**Interfaces and types**
- Accept interfaces, return concrete types. Keep interfaces small, and define them where they are used.
- Make the zero value useful. Use functional options or a config struct for constructors with many settings.
- Do not use `any` when a generic or concrete type works.

**Concurrency**
- Share memory by communicating, but a mutex is fine for simple state.
- Every goroutine needs a known way to stop (context or channel close) and an owner who waits for it (`WaitGroup`, `errgroup`).
- Pass `context.Context` as the first parameter; never store it in a struct.
- Run tests with `-race`.

**Testing and tooling**
- Table-driven tests with subtests; use `testdata/` and `t.TempDir()`.
- Run `go vet`, `staticcheck` or `golangci-lint` in CI; keep `go.mod` tidy with `go mod tidy`.
- Benchmark and profile before optimizing.

> **Tip:** read the standard library. `net/http`, `io`, `sort` and `encoding/json` are some of the best examples of idiomatic Go you will find.

## Where to go next

- *Effective Go* and the *Go Code Review Comments* wiki page.
- The Go blog (go.dev/blog), especially posts on errors, context, pipelines and generics.
- *The Go Programming Language* (Donovan and Kernighan) and *100 Go Mistakes* (Teiva Harsanyi).
- "Go Proverbs" and the talks by Rob Pike, Dave Cheney and Bill Kennedy.
- Build something real: a small JSON API with tests, structured logs, graceful shutdown and a profile, then read other people's open-source Go.

## Key takeaways

- Put binaries in `cmd/`, private code in `internal/`, and wire dependencies in `main`.
- `*sql.DB` is a pool: create once, use contexts, placeholders, close rows, check `rows.Err()`.
- Idiomatic Go is simple, explicit and consistent: small interfaces, errors as values, disciplined concurrency, and tests.
- When in doubt, choose clarity over cleverness.
