package playground

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/compile" || r.FormValue("version") != "2" || r.FormValue("withVet") != "true" {
			t.Errorf("unexpected request %s %v", r.URL.Path, r.Form)
		}
		if r.FormValue("body") != "package main" {
			t.Errorf("body = %q", r.FormValue("body"))
		}
		w.Write([]byte(`{"Errors":"","Events":[{"Message":"hi\n","Kind":"stdout","Delay":0},{"Message":"bye\n","Kind":"stderr","Delay":1500000000}],"Status":1,"VetErrors":""}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	res, err := c.Run(context.Background(), "package main")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Events) != 2 || res.Events[1].Kind != "stderr" || res.Events[1].DelayMs != 1500 {
		t.Errorf("events = %+v", res.Events)
	}
	if res.ExitCode != 1 {
		t.Errorf("exit code = %d", res.ExitCode)
	}
}

func TestRunCompileErrorHasEmptyEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Errors":"prog.go:3:1: syntax error","Events":null,"Status":0}`))
	}))
	defer srv.Close()

	res, err := (&Client{BaseURL: srv.URL, HTTP: srv.Client()}).Run(context.Background(), "x")
	if err != nil {
		t.Fatal(err)
	}
	if res.CompileErrors == "" || res.Events == nil {
		t.Errorf("result = %+v", res)
	}
}

func TestRunUnavailable(t *testing.T) {
	busy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "slow down", http.StatusTooManyRequests)
	}))
	defer busy.Close()
	down := httptest.NewServer(nil)
	down.Close() // connection refused

	for name, c := range map[string]*Client{
		"rate-limited": {BaseURL: busy.URL, HTTP: busy.Client()},
		"unreachable":  {BaseURL: down.URL, HTTP: http.DefaultClient},
	} {
		if _, err := c.Run(context.Background(), "x"); !errors.Is(err, ErrUnavailable) {
			t.Errorf("%s: err = %v, want ErrUnavailable", name, err)
		}
	}
}
