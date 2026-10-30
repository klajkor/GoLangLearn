---
title: Middleware and JSON APIs
minutes: 13
---
Cross-cutting concerns such as logging, panic recovery and timeouts belong in **middleware**, not in every handler. JSON APIs add a second set of habits: strict decoding and consistent error responses.

## Middleware is just a function

A middleware takes a handler and returns a new handler that wraps it:

```go
type Middleware func(http.Handler) http.Handler
```

It can run code before, after, or instead of the wrapped handler. The first example builds logging and panic recovery. The logging middleware needs the status code, which `ResponseWriter` does not expose, so we wrap it.

```go run
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type Middleware func(http.Handler) http.Handler

// Chain applies middleware so the first one listed is outermost.
func Chain(h http.Handler, ms ...Middleware) http.Handler {
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		fmt.Printf("log: %s %s -> %d\n", r.Method, r.URL.Path, rec.status)
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				fmt.Println("recovered:", v)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "fine")
	})
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		panic("something broke")
	})

	h := Chain(mux, Logging, Recover) // Logging wraps Recover wraps mux

	for _, path := range []string{"/ok", "/boom"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		fmt.Printf("client got %d %s\n", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
}
```

Order matters: `Logging` is outside `Recover`, so it still sees the 500 that `Recover` produced.

Two notes on real servers. The `net/http` server already recovers panics per request (and logs them), but explicit recovery lets you return a proper response and log with your own logger. And **timeouts**: `http.TimeoutHandler(h, 2*time.Second, "timeout")` is a ready-made middleware that returns 503 if the handler is slow. Remember that handlers should also respect `r.Context()` so slow work stops when the client leaves.

> **Tip:** middleware that needs configuration is a function returning a `Middleware`, for example `RateLimit(100) Middleware`. Pass request-scoped values (request ID, user) with `context.WithValue` and a private key type.

## JSON APIs

`encoding/json` turns structs into JSON and back using field **tags**:

| Tag | Effect |
|---|---|
| `json:"name"` | use `name` as the JSON key |
| `json:"name,omitempty"` | omit when the value is empty (`""`, `0`, `nil`, empty slice or map) |
| `json:"-"` | never encode or decode this field |

Only exported fields take part. Good JSON handlers do four things: limit the body size, decode strictly, validate, and answer with a consistent shape.

```go run
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type createUser struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

type user struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

type apiError struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB cap
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var in createUser
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}
	if in.Name == "" {
		writeJSON(w, http.StatusUnprocessableEntity, apiError{Error: "name is required"})
		return
	}
	writeJSON(w, http.StatusCreated, user{ID: 1, Name: in.Name, Email: in.Email})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", createHandler)

	bodies := []string{
		`{"name":"Ann","email":"ann@example.com"}`,
		`{"name":"Bob"}`,
		`{"name":"Cy","age":30}`,
		`{"email":"x@y.z"}`,
		`{"name":`,
	}
	for _, b := range bodies {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/users", strings.NewReader(b))
		mux.ServeHTTP(rec, req)
		fmt.Printf("%d %s", rec.Code, rec.Body.String())
	}
}
```

Notice `omitempty` dropping Bob's empty email, and `DisallowUnknownFields` rejecting the `age` key (a typo in a client request is better caught than silently ignored).

## Status codes that matter

| Code | Use it for |
|---|---|
| 200 / 201 / 204 | success / created / success with no body |
| 400 | malformed request (bad JSON, bad parameter) |
| 401 / 403 | not authenticated / not allowed |
| 404 | no such resource |
| 409 | conflict (duplicate) |
| 422 | well-formed but fails validation |
| 500 | your bug; never leak internal error text to clients, log it instead |

Decide on one error format (such as `{"error": "..."}` or RFC 9457 problem details) and use it everywhere via one helper. Map domain errors to codes with `errors.Is` / `errors.As` in one place.

> **Common mistake:** ignoring the error from `Decode`, or calling `WriteHeader` after writing the body (the status is already sent). Also, `json.Marshal` fails on channels, funcs and cyclic data, so check its error when the data is not under your control.

## Key takeaways

- Middleware is `func(http.Handler) http.Handler`; wrap `ResponseWriter` to observe the status.
- Chain order matters: the outermost middleware sees everything inside it.
- Use `MaxBytesReader`, `DisallowUnknownFields`, then validate.
- Tags `omitempty` and `-` control encoding; only exported fields are encoded.
- Return consistent JSON errors with the right status codes and hide internals from clients.
