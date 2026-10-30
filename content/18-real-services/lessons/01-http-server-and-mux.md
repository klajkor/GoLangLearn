---
title: net/http Servers and the ServeMux
minutes: 11
---
Go's standard library contains a production-grade HTTP server. Many Go services use nothing else: no framework, just `net/http`. In this chapter, examples that would need a real socket are shown as read-only code; the runnable ones drive handlers directly with `httptest`, which needs no network.

## The Handler interface

The whole server model rests on one interface:

```go
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

`http.HandlerFunc` adapts an ordinary function to this interface, which is why you can write `mux.HandleFunc("/x", func(w http.ResponseWriter, r *http.Request) {...})`.

Inside a handler: read from `r` (method, URL, headers, body, context), write to `w`. The order matters: set headers first, then call `w.WriteHeader(status)` (optional; the default is 200), then write the body. After the first body write, headers are frozen.

## Starting a server

A real server looks like this (read-only: it opens a socket):

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout:       10 * time.Second,
	WriteTimeout:      15 * time.Second,
	IdleTimeout:       60 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```

> **Common mistake:** using `http.ListenAndServe(addr, nil)` in production. It has no timeouts, so slow or malicious clients can hold connections open forever, and `nil` uses the global `DefaultServeMux`, which any imported package can add routes to. Build your own `http.Server` and `ServeMux`.

## ServeMux patterns (Go 1.22+)

Since Go 1.22 the standard `ServeMux` understands **HTTP methods** and **path wildcards**, which removes the main reason people used third-party routers. Pattern syntax: `[METHOD ][HOST]/[PATH]`.

| Pattern | Matches |
|---|---|
| `GET /items` | `GET` (and `HEAD`) for exactly `/items` |
| `GET /items/{id}` | one path segment bound to `id` |
| `GET /files/{path...}` | the rest of the path, including slashes |
| `GET /{$}` | only the root `/` exactly |
| `/static/` | any path under `/static/` (any method) |

Read wildcard values with `r.PathValue("id")`. The most specific pattern wins, and a request that matches a path but not the method gets an automatic `405 Method Not Allowed` with an `Allow` header. Unknown paths get `404`. Conflicting patterns panic at registration time, so mistakes show up at startup.

Let's try it without a network, using `httptest`:

```go run
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "home")
	})
	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s!\n", r.PathValue("name"))
	})
	mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "file %q\n", r.PathValue("path"))
	})
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "created")
	})
	return mux
}

func do(h http.Handler, method, target string) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	h.ServeHTTP(rec, req)
	fmt.Printf("%-5s %-16s -> %d %s\n", method, target, rec.Code, strings.TrimSpace(rec.Body.String()))
}

func main() {
	mux := newMux()
	do(mux, "GET", "/")
	do(mux, "GET", "/hello/Gopher")
	do(mux, "GET", "/files/a/b/c.txt")
	do(mux, "POST", "/items")
	do(mux, "GET", "/items")
	do(mux, "GET", "/nope")
	do(mux, "GET", "/other")
}
```

Notice `GET /items` yields 405 (the path exists for POST), while `/nope` is 404. `httptest.NewRequest` builds a server-side request (unlike `http.NewRequest`, which builds a client request), and `ServeHTTP` runs the whole mux in-process.

## Handlers that return errors

A popular pattern keeps handlers short by letting them return an error that one wrapper converts to a response:

```go
type appHandler func(http.ResponseWriter, *http.Request) error

func (f appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := f(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

Real code maps error types to status codes (we do that next lesson). Also note that `http.Error` sets the content type to `text/plain` and appends a newline.

## Dependencies via structs

Avoid global variables. Group handlers on a struct that holds dependencies (logger, store) and register methods as handlers:

```go
type Server struct {
	log   *slog.Logger
	store Store
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", s.getItem)
	return mux
}
```

This makes handlers easy to test: construct `Server` with a fake `Store`.

## Key takeaways

- Everything is an `http.Handler`; `HandlerFunc` adapts functions.
- Set headers, then status, then body.
- Build your own `http.Server` with timeouts and your own `ServeMux`; avoid the global default.
- Go 1.22 patterns: `"GET /items/{id}"`, `{path...}`, `{$}`, with `r.PathValue`.
- Wrong method gives 405, unknown path gives 404.
- Put dependencies on a struct and test handlers in-process.
