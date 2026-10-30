---
title: Graceful Shutdown, Configuration and slog
minutes: 12
---
A service that works in a demo is not yet a service you can deploy. It must stop cleanly, take its settings from the environment, and log in a way machines can read.

## Graceful shutdown

When a deploy restarts your process, the orchestrator sends `SIGTERM` (and a developer pressing Ctrl+C sends `SIGINT`). If you ignore them, in-flight requests are cut off. The recipe: listen for signals using `signal.NotifyContext` (Go 1.16), then call `srv.Shutdown`, which stops accepting new connections and waits for active requests to finish, up to a deadline. This needs real sockets and signals, so it is read-only:

```go
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: ":8080", Handler: routes()}

	errCh := make(chan error, 1)
	go func() {
		// ErrServerClosed is the normal result of Shutdown, not a failure.
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err // failed to start, or crashed
	case <-ctx.Done():
		stop() // restore default signal behavior: a second Ctrl+C kills us
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return <-errCh
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}
```

Points worth noticing:

- `ListenAndServe` returns `http.ErrServerClosed` immediately after `Shutdown` is called, so the goroutine must not treat that as an error.
- The shutdown context is a **fresh** one with a timeout; the signal context is already cancelled.
- `main` stays tiny, and `run()` returns an error, so deferred calls run before `os.Exit`.
- Background workers should also take a context so that they stop at the same time. To tie request contexts to shutdown, set `BaseContext` on the server.
- `Shutdown` does not wait for hijacked connections such as WebSockets; use `srv.RegisterOnShutdown` for those.

## Configuration

Follow the twelve-factor habit: configuration comes from the environment (with flags for local overrides), never from constants in the code. Parse it once at startup into a struct and pass the struct down.

```go run
package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	Addr    string
	Timeout time.Duration
	Debug   bool
}

// load reads flags from args; env supplies defaults. In a real program
// pass os.Args[1:] and os.LookupEnv instead of the fakes used here.
func load(args []string, lookup func(string) (string, bool)) (Config, error) {
	getenv := func(key, def string) string {
		if v, ok := lookup(key); ok {
			return v
		}
		return def
	}

	fs := flag.NewFlagSet("svc", flag.ContinueOnError)
	var cfg Config
	fs.StringVar(&cfg.Addr, "addr", getenv("APP_ADDR", ":8080"), "listen address")
	fs.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "request timeout")
	fs.BoolVar(&cfg.Debug, "debug", false, "debug logging")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if v, ok := lookup("APP_PORT_CHECK"); ok {
		if _, err := strconv.Atoi(v); err != nil {
			return Config{}, fmt.Errorf("APP_PORT_CHECK: %w", err)
		}
	}
	return cfg, nil
}

func main() {
	env := map[string]string{"APP_ADDR": ":9000"}
	lookup := func(k string) (string, bool) { v, ok := env[k]; return v, ok }

	cfg, err := load([]string{"-timeout", "250ms"}, lookup)
	fmt.Printf("%+v %v\n", cfg, err)

	cfg, err = load([]string{"-addr", ":7000", "-debug"}, lookup)
	fmt.Printf("%+v %v\n", cfg, err)
}
```

Injecting the lookup function (instead of calling `os.Getenv` deep inside) keeps the loader testable. Fail fast on invalid configuration, and never log secrets.

## Structured logging with log/slog

`log/slog` (Go 1.21) writes **key-value records** with levels. Text output suits a terminal; JSON suits log aggregators. Handlers decide the format, so your code only calls `logger.Info("msg", "key", value, ...)`.

```go run
package main

import (
	"log/slog"
	"os"
)

func main() {
	// Drop the timestamp so the output is stable for this demo.
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}
			return a
		},
	}

	text := slog.New(slog.NewTextHandler(os.Stdout, opts))
	text.Info("request", "method", "GET", "path", "/items/7", "status", 200)
	text.Debug("cache miss", "key", "item:7")

	jsonLog := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	reqLog := jsonLog.With("request_id", "abc123")
	reqLog.Warn("slow query", slog.Int("ms", 840), slog.Group("db", "name", "orders"))
	reqLog.Error("failed", "err", os.ErrNotExist)
}
```

Good habits: construct a logger once and inject it; use `With` to attach request-scoped fields such as a request ID; use `InfoContext(ctx, ...)` variants so handlers can pull values from the context; and log an error **once**, where you handle it, instead of at every layer it passes through. In production use `slog.NewJSONHandler(os.Stderr, nil)` and `slog.SetDefault` so the old `log` package also writes through it.

> **Common mistake:** an odd number of key-value arguments, such as `logger.Info("x", "key")`. It produces a `!BADKEY` entry. The `go vet` `slog` analyzer catches this; or use typed `slog.String`, `slog.Int` attrs.

## Key takeaways

- Use `signal.NotifyContext` plus `srv.Shutdown(ctx)` with a timeout; `ErrServerClosed` is not a failure.
- Keep `main` small: `run() error` makes cleanup and testing easy.
- Load configuration once into a struct from env and flags; inject lookups for testability.
- `log/slog` gives leveled, structured logs; choose text or JSON through the handler.
- Add context with `logger.With(...)` and log each error once.
