---
title: Testing Handlers and Writing HTTP Clients
minutes: 11
---
Handlers are plain functions of `(ResponseWriter, *Request)`, so they are easy to test without any network. Clients are harder to get right: the default settings are dangerous.

## Testing handlers with httptest

`httptest.NewRecorder()` is a fake `ResponseWriter` that records the status, headers and body. Combine it with `httptest.NewRequest`. A typical test file (read-only here, because it uses `testing`):

```go
func TestGetItem(t *testing.T) {
	srv := &Server{store: fakeStore{"1": Item{ID: "1", Name: "pen"}}}
	h := srv.routes()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"found", "/items/1", http.StatusOK, `{"id":"1","name":"pen"}`},
		{"missing", "/items/9", http.StatusNotFound, `{"error":"not found"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest("GET", tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}
```

Test the **router** (`h`), not just the handler function, so path patterns and `PathValue` are exercised too. If you need a real listener (for example to test a client or TLS), `httptest.NewServer(handler)` starts one on a random local port and gives you `srv.URL`; remember to `defer srv.Close()`. That needs sockets, so it cannot run on the Playground.

## http.Client: never use the defaults

`http.Get` uses `http.DefaultClient`, which has **no timeout**. A server that never answers will hang your goroutine forever. Always build a client:

```go
client := &http.Client{Timeout: 10 * time.Second} // covers connect, redirects, and reading the body

req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil {
	return err
}
req.Header.Set("Accept", "application/json")

resp, err := client.Do(req)
if err != nil {
	return fmt.Errorf("fetch: %w", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
	return fmt.Errorf("fetch: unexpected status %s", resp.Status)
}
var out Result
if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
	return fmt.Errorf("decode: %w", err)
}
```

The rules to remember:

- **Always close `resp.Body`**, or connections leak. If you do not need the body, still drain and close it so the connection can be reused.
- A non-2xx status is **not** an error from `Do`; check `StatusCode` yourself.
- Reuse one `Client` for the whole program. Its `Transport` pools connections and is safe for concurrent use.
- Tie requests to a `context.Context`. Timeouts per call: `ctx, cancel := context.WithTimeout(parent, 3*time.Second)`. `Client.Timeout` is the overall safety net.
- Retries need care: only retry idempotent requests, use backoff with jitter, and stop when the context is done.

## A deterministic client demo

`Client.Transport` is an interface (`RoundTripper`), so a test can replace the network with a function. This also runs without sockets:

```go run
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fakeTransport() http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if err := r.Context().Err(); err != nil {
			return nil, err // a real transport also stops when the context ends
		}
		body := `{"city":"Oslo","temp":3}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})
}

type weather struct {
	City string `json:"city"`
	Temp int    `json:"temp"`
}

func fetch(ctx context.Context, c *http.Client, url string) (weather, error) {
	var w weather
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return w, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return w, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return w, fmt.Errorf("fetch: status %s", resp.Status)
	}
	return w, json.NewDecoder(resp.Body).Decode(&w)
}

func main() {
	client := &http.Client{Timeout: 5 * time.Second, Transport: fakeTransport()}

	w, err := fetch(context.Background(), client, "http://weather.invalid/oslo")
	fmt.Println(w, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = fetch(ctx, client, "http://weather.invalid/oslo")
	fmt.Println(errors.Is(err, context.Canceled))
}
```

The second call fails with an error that wraps `context.Canceled` (`Do` wraps errors in `*url.Error`, which supports `errors.Is`). Putting the HTTP call behind a small interface in your own code (`type WeatherAPI interface { Fetch(...) }`) lets the rest of your tests avoid HTTP entirely.

> **Common mistake:** `defer resp.Body.Close()` placed before checking `err`. If `Do` fails, `resp` is nil and the deferred call panics. Check the error first.

## Key takeaways

- Test handlers with `httptest.NewRecorder` and `NewRequest`; test the router so patterns are covered.
- `httptest.NewServer` gives a real local server for client tests.
- `http.DefaultClient` has no timeout: build a `Client` with `Timeout` and use `NewRequestWithContext`.
- Close response bodies, check status codes yourself, and reuse clients.
- A fake `RoundTripper` makes client code testable without a network.
