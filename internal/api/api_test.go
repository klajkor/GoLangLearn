package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"golearn/internal/content"
	"golearn/internal/playground"
	"golearn/internal/store"
)

func newTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	cur, err := content.Load(fstest.MapFS{
		"01-intro/chapter.yaml":        {Data: []byte("title: Intro\ntier: beginner\n")},
		"01-intro/lessons/01-hello.md": {Data: []byte("---\ntitle: Hello\n---\nHi there")},
		"01-intro/quiz.yaml":           {Data: []byte(testQuiz)},
		"02-types/chapter.yaml":        {Data: []byte("title: Types\ntier: beginner\n")},
	})
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	mux := http.NewServeMux()
	New(cur, st, &fakeRunner{}).Register(mux)
	return mux
}

// fakeRunner echoes the code back as stdout, or fails if code is "offline".
type fakeRunner struct{}

func (fakeRunner) Run(_ context.Context, code string) (*playground.Result, error) {
	if code == "offline" {
		return nil, fmt.Errorf("%w: test", playground.ErrUnavailable)
	}
	return &playground.Result{Events: []playground.Event{{Message: code, Kind: "stdout"}}}, nil
}

func post(t *testing.T, mux *http.ServeMux, path, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("%s: invalid JSON: %v", path, err)
	}
	return rec, out
}

func get(t *testing.T, mux *http.ServeMux, path string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("%s: invalid JSON: %v", path, err)
	}
	return rec, body
}

func TestCurriculumOmitsBodies(t *testing.T) {
	rec, body := get(t, newTestMux(t), "/api/curriculum")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	chapters := body["chapters"].([]any)
	lesson := chapters[0].(map[string]any)["lessons"].([]any)[0].(map[string]any)
	if lesson["title"] != "Hello" {
		t.Errorf("lesson = %v", lesson)
	}
	if _, ok := lesson["body"]; ok {
		t.Error("curriculum should not include lesson bodies")
	}
}

func TestLesson(t *testing.T) {
	mux := newTestMux(t)
	rec, body := get(t, mux, "/api/chapters/intro/lessons/hello")
	if rec.Code != http.StatusOK || body["body"] != "Hi there" {
		t.Errorf("status = %d, body = %v", rec.Code, body)
	}
	for _, p := range []string{"/api/chapters/nope/lessons/hello", "/api/chapters/intro/lessons/nope", "/api/unknown"} {
		if rec, _ := get(t, mux, p); rec.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404", p, rec.Code)
		}
	}
}

func TestProgress(t *testing.T) {
	rec, body := get(t, newTestMux(t), "/api/progress")
	if rec.Code != http.StatusOK || body["totalXP"] != float64(0) {
		t.Errorf("status = %d, body = %v", rec.Code, body)
	}
}

func TestRun(t *testing.T) {
	mux := newTestMux(t)

	rec, body := post(t, mux, "/api/run", `{"code":"hello"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", rec.Code, body)
	}
	if ev := body["events"].([]any)[0].(map[string]any); ev["message"] != "hello" {
		t.Errorf("events = %v", body["events"])
	}

	if rec, _ := post(t, mux, "/api/run", `{"code":"offline"}`); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("offline: status = %d, want 503", rec.Code)
	}
	if rec, _ := post(t, mux, "/api/run", `not json`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad body: status = %d, want 400", rec.Code)
	}
	big := `{"code":"` + strings.Repeat("x", maxCodeBytes) + `"}`
	if rec, _ := post(t, mux, "/api/run", big); rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("big body: status = %d, want 413", rec.Code)
	}
}

func TestFormat(t *testing.T) {
	mux := newTestMux(t)

	rec, body := post(t, mux, "/api/format", `{"code":"package main\nfunc main(){x:=1;_=x}"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", rec.Code, body)
	}
	want := "package main\n\nfunc main() { x := 1; _ = x }\n"
	if body["code"] != want {
		t.Errorf("code = %q, want %q", body["code"], want)
	}

	rec, body = post(t, mux, "/api/format", `{"code":"package main\nfunc {"}`)
	if rec.Code != http.StatusUnprocessableEntity || !strings.HasPrefix(body["error"].(string), "2:") {
		t.Errorf("syntax error: status = %d, body = %v", rec.Code, body)
	}
}

const testQuiz = `
passPercent: 50
questions:
  - id: q1
    type: single
    prompt: Pick b
    options: [a, b]
    answer: 1
    explanation: because b
  - id: q2
    type: output
    prompt: Prints?
    code: fmt.Println(1)
    answer: "1"
    explanation: one
`

func TestQuizHidesAnswers(t *testing.T) {
	mux := newTestMux(t)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/chapters/intro/quiz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	raw := rec.Body.String()
	for _, secret := range []string{"because b", `"answer"`, `"explanation"`, "correctAnswer"} {
		if strings.Contains(raw, secret) {
			t.Errorf("quiz response leaks %q: %s", secret, raw)
		}
	}
	var body struct {
		PassPercent int
		Questions   []map[string]any
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.PassPercent != 50 || len(body.Questions) != 2 || body.Questions[0]["prompt"] != "Pick b" {
		t.Errorf("body = %+v", body)
	}
}

func TestQuizNotFound(t *testing.T) {
	mux := newTestMux(t)
	for _, p := range []string{"/api/chapters/types/quiz", "/api/chapters/nope/quiz"} {
		if rec, _ := get(t, mux, p); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", p, rec.Code)
		}
		if rec, _ := post(t, mux, p, `{"answers":{}}`); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: status = %d, want 404", p, rec.Code)
		}
	}
}

func TestSubmitQuiz(t *testing.T) {
	mux := newTestMux(t)

	rec, body := post(t, mux, "/api/chapters/intro/quiz", `{"answers":{"q1":1,"q2":"1\n"}}`)
	if rec.Code != http.StatusOK || body["score"] != float64(2) || body["passed"] != true || body["percent"] != float64(100) {
		t.Fatalf("perfect attempt: status = %d, body = %v", rec.Code, body)
	}
	results := body["results"].([]any)
	if r := results[0].(map[string]any); r["correct"] != true || r["correctAnswer"] != float64(1) || r["explanation"] != "because b" {
		t.Errorf("result[0] = %v", r)
	}

	// One right, one unanswered: 50% meets passPercent 50.
	_, body = post(t, mux, "/api/chapters/intro/quiz", `{"answers":{"q1":1}}`)
	if body["score"] != float64(1) || body["passed"] != true {
		t.Errorf("half attempt: %v", body)
	}

	// Nothing right: fails.
	_, body = post(t, mux, "/api/chapters/intro/quiz", `{"answers":{"q1":0,"q2":"2","unknown":5}}`)
	if body["score"] != float64(0) || body["passed"] != false {
		t.Errorf("failed attempt: %v", body)
	}

	if rec, _ := post(t, mux, "/api/chapters/intro/quiz", `nope`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad body: status = %d, want 400", rec.Code)
	}

	_, prog := get(t, mux, "/api/progress")
	stat := prog["quizzes"].(map[string]any)["intro"].(map[string]any)
	if stat["attempts"] != float64(3) || stat["bestPercent"] != float64(100) || stat["passed"] != true {
		t.Errorf("progress quizzes = %v", stat)
	}
}

func TestCurriculumReportsQuizSize(t *testing.T) {
	_, body := get(t, newTestMux(t), "/api/curriculum")
	chapters := body["chapters"].([]any)
	if n := chapters[0].(map[string]any)["quizQuestions"]; n != float64(2) {
		t.Errorf("intro quizQuestions = %v", n)
	}
	if n := chapters[1].(map[string]any)["quizQuestions"]; n != float64(0) {
		t.Errorf("types quizQuestions = %v", n)
	}
}
