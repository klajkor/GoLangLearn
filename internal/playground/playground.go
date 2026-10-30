// Package playground runs Go programs on the public Go Playground
// (https://go.dev/play) via its compile endpoint.
package playground

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrUnavailable means the Playground could not be reached or answered
// with an unexpected status (offline, blocked by a proxy, rate-limited).
var ErrUnavailable = errors.New("go playground unavailable")

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New() *Client {
	return &Client{
		BaseURL: "https://play.golang.org",
		HTTP:    &http.Client{Timeout: 20 * time.Second},
	}
}

// Event is one chunk of program output. DelayMs is the simulated time
// since the previous event (the Playground fakes time.Sleep).
type Event struct {
	Message string `json:"message"`
	Kind    string `json:"kind"` // "stdout" or "stderr"
	DelayMs int64  `json:"delayMs"`
}

type Result struct {
	Events        []Event `json:"events"`
	CompileErrors string  `json:"compileErrors,omitempty"`
	VetErrors     string  `json:"vetErrors,omitempty"`
	ExitCode      int     `json:"exitCode"`
}

// response mirrors the Playground's JSON (Delay is in nanoseconds).
type response struct {
	Errors string
	Events []struct {
		Message string
		Kind    string
		Delay   int64
	}
	Status    int
	VetErrors string
}

// Run compiles and runs code, with go vet enabled.
func (c *Client) Run(ctx context.Context, code string) (*Result, error) {
	form := url.Values{"version": {"2"}, "body": {code}, "withVet": {"true"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/compile", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %s", ErrUnavailable, res.Status)
	}

	var r response
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("%w: bad response: %v", ErrUnavailable, err)
	}
	out := &Result{
		Events:        []Event{},
		CompileErrors: r.Errors,
		VetErrors:     r.VetErrors,
		ExitCode:      r.Status,
	}
	for _, e := range r.Events {
		out.Events = append(out.Events, Event{Message: e.Message, Kind: e.Kind, DelayMs: e.Delay / int64(time.Millisecond)})
	}
	return out, nil
}
