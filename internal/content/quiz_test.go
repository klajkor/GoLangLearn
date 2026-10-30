package content

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

const sampleQuiz = `
passPercent: 60
questions:
  - id: s
    type: single
    prompt: Pick b
    options: [a, b, c]
    answer: 1
    explanation: b is right
  - id: m
    type: multi
    prompt: Pick a and c
    options: [a, b, c]
    answer: [0, 2]
    explanation: a and c
  - id: o
    type: output
    prompt: What prints?
    code: |
      fmt.Println("a", "b")
    answer: ["a b", "a  b"]
    explanation: spaces
  - id: f
    type: fill
    prompt: Fill in
    code: |
      import "___"
      fmt.___("x")
    answer: [fmt, [Println, Print]]
    explanation: fmt
  - id: b
    type: bug
    prompt: Which line?
    code: |
      a := 1
      b := 2
      fmt.Println(a)
    answer: 2
    explanation: b unused
`

func mustQuiz(t *testing.T) *Quiz {
	t.Helper()
	qz, err := parseQuiz([]byte(sampleQuiz))
	if err != nil {
		t.Fatal(err)
	}
	return qz
}

func TestGrade(t *testing.T) {
	qz := mustQuiz(t)
	if qz.PassPercent != 60 {
		t.Errorf("PassPercent = %d", qz.PassPercent)
	}
	byID := map[string]*Question{}
	for _, q := range qz.Questions {
		byID[q.ID] = q
	}

	tests := []struct {
		id   string
		ans  any
		want bool
	}{
		{"s", 1.0, true},
		{"s", 0.0, false},
		{"s", 1.5, false},
		{"s", "1", false},
		{"s", nil, false},

		{"m", []any{0.0, 2.0}, true},
		{"m", []any{2.0, 0.0, 2.0}, true}, // order and duplicates don't matter
		{"m", []any{0.0}, false},
		{"m", []any{0.0, 1.0, 2.0}, false},
		{"m", []any{"0", "2"}, false},
		{"m", 0.0, false},

		{"o", "a b", true},
		{"o", "  a b  \r\n", true},
		{"o", "a  b", true},
		{"o", "A b", false},
		{"o", 1.0, false},

		{"f", []any{"fmt", "Println"}, true},
		{"f", []any{" fmt ", "Print"}, true},
		{"f", []any{"fmt", "println"}, false}, // case matters
		{"f", []any{"fmt"}, false},
		{"f", []any{"fmt", 1.0}, false},

		{"b", 2.0, true},
		{"b", 1.0, false},
	}
	for _, tt := range tests {
		if got := byID[tt.id].Grade(tt.ans); got != tt.want {
			t.Errorf("Grade(%s, %#v) = %v, want %v", tt.id, tt.ans, got, tt.want)
		}
	}
}

func TestCorrectAnswerGradesAsCorrect(t *testing.T) {
	for _, q := range mustQuiz(t).Questions {
		// Round-trip through JSON-like types, as the client would submit it.
		var ans any
		switch v := q.CorrectAnswer().(type) {
		case int:
			ans = float64(v)
		case []int:
			list := []any{}
			for _, n := range v {
				list = append(list, float64(n))
			}
			ans = list
		case []string:
			list := []any{}
			for _, s := range v {
				list = append(list, s)
			}
			ans = list
		default:
			ans = v
		}
		if !q.Grade(ans) {
			t.Errorf("%s: CorrectAnswer %#v does not grade as correct", q.ID, q.CorrectAnswer())
		}
	}
}

func TestPublicHidesAnswers(t *testing.T) {
	qz := mustQuiz(t)
	for _, q := range qz.Questions {
		p := q.Public()
		if p.ID != q.ID || p.Prompt != q.Prompt {
			t.Errorf("Public() lost fields: %+v", p)
		}
	}
	if got := qz.Questions[3].Public().Blanks; got != 2 {
		t.Errorf("Blanks = %d, want 2", got)
	}
}

// q wraps question fields (YAML, indented under the question) in a quiz document.
func q(fields string) string {
	return "questions:\n  - id: a\n    prompt: p\n    explanation: e\n" + fields
}

func TestQuizValidation(t *testing.T) {
	bad := []struct{ name, src, wantErr string }{
		{"no questions", "questions: []", "no questions"},
		{"bad pass", "passPercent: 150\n" + q("    type: single\n    options: [x, y]\n    answer: 0\n"), "passPercent"},
		{"missing explanation", "questions:\n  - id: a\n    prompt: p\n    type: single\n    options: [x, y]\n    answer: 0\n", "explanation is required"},
		{"missing id", "questions:\n  - prompt: p\n    explanation: e\n    type: single\n    options: [x, y]\n    answer: 0\n", "id and prompt"},
		{"unknown type", q("    type: essay\n"), "unknown type"},
		{"single out of range", q("    type: single\n    options: [x, y]\n    answer: 2\n"), "option index"},
		{"single not int", q("    type: single\n    options: [x, y]\n    answer: x\n"), "option index"},
		{"one option", q("    type: single\n    options: [x]\n    answer: 0\n"), "at least 2 options"},
		{"multi empty", q("    type: multi\n    options: [x, y]\n    answer: []\n"), "non-empty list"},
		{"multi range", q("    type: multi\n    options: [x, y]\n    answer: [0, 5]\n"), "not an option index"},
		{"output no code", q("    type: output\n    answer: x\n"), "code is required"},
		{"output bad answer", q("    type: output\n    code: c\n    answer: 5\n"), "string or a non-empty list"},
		{"fill no gap", q("    type: fill\n    code: abc\n    answer: [x]\n"), "gap"},
		{"fill wrong count", q("    type: fill\n    code: '___ ___'\n    answer: [x]\n"), "one entry per gap"},
		{"fill bare string", q("    type: fill\n    code: '___'\n    answer: x\n"), "one entry per gap"},
		{"bug out of range", q("    type: bug\n    code: one\n    answer: 2\n"), "line number"},
		{
			"duplicate id",
			q("    type: single\n    options: [x, y]\n    answer: 0\n") +
				"  - id: a\n    prompt: p\n    explanation: e\n    type: single\n    options: [x, y]\n    answer: 1\n",
			"duplicate",
		},
	}
	for _, tt := range bad {
		_, err := parseQuiz([]byte(tt.src))
		if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
			t.Errorf("%s: err = %v, want it to contain %q", tt.name, err, tt.wantErr)
		}
	}
}

func TestLoadAttachesQuiz(t *testing.T) {
	cur, err := Load(fstest.MapFS{
		"01-a/chapter.yaml": {Data: []byte("title: A\ntier: beginner\n")},
		"01-a/quiz.yaml":    {Data: []byte(sampleQuiz)},
		"02-b/chapter.yaml": {Data: []byte("title: B\ntier: beginner\n")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if a := cur.Chapter("a"); a.Quiz == nil || a.QuizQuestions != 5 {
		t.Errorf("chapter a quiz = %+v, questions = %d", a.Quiz, a.QuizQuestions)
	}
	if b := cur.Chapter("b"); b.Quiz != nil || b.QuizQuestions != 0 {
		t.Errorf("chapter b should have no quiz, got %+v", b.Quiz)
	}

	_, err = Load(fstest.MapFS{
		"01-a/chapter.yaml": {Data: []byte("title: A\ntier: beginner\n")},
		"01-a/quiz.yaml":    {Data: []byte("questions: []")},
	})
	if err == nil || !strings.Contains(err.Error(), "quiz.yaml") {
		t.Errorf("expected quiz.yaml error, got %v", err)
	}
}

// TestRealContent makes sure everything under content/ parses and validates,
// so a typo in a lesson or quiz fails `go test` instead of the running app.
func TestRealContent(t *testing.T) {
	cur, err := Load(os.DirFS("../../content"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cur.Chapters) == 0 {
		t.Fatal("no chapters loaded")
	}
	for _, ch := range cur.Chapters {
		for _, l := range ch.Lessons {
			if l.Title == "" {
				t.Errorf("%s/%s: lesson has no title", ch.ID, l.ID)
			}
		}
	}
}

// Shuffling must move options around without changing which text is correct.
func TestShuffledOptionsKeepAnswerKey(t *testing.T) {
	plain, err := Load(os.DirFS("../../content"))
	if err != nil {
		t.Fatal(err)
	}
	shuffled, err := Load(os.DirFS("../../content"), WithShuffledOptions())
	if err != nil {
		t.Fatal(err)
	}
	moved, checked := 0, 0
	for i, ch := range plain.Chapters {
		if ch.Quiz == nil {
			continue
		}
		for j, q := range ch.Quiz.Questions {
			s := shuffled.Chapters[i].Quiz.Questions[j]
			if q.Type != "single" && q.Type != "multi" {
				continue
			}
			checked++
			if !s.Grade(viaJSON(t, s.CorrectAnswer())) {
				t.Errorf("%s/%s: shuffled key doesn't grade as correct", ch.ID, q.ID)
			}
			want := correctTexts(q)
			got := correctTexts(s)
			if !slices.Equal(want, got) {
				t.Errorf("%s/%s: correct options changed from %v to %v", ch.ID, q.ID, want, got)
			}
			if !slices.Equal(q.Options, s.Options) {
				moved++
			}
		}
	}
	if checked == 0 || moved < checked/2 {
		t.Errorf("only %d of %d choice questions were reordered", moved, checked)
	}
}

func correctTexts(q *Question) []string {
	var out []string
	switch a := q.CorrectAnswer().(type) {
	case int:
		out = append(out, q.Options[a])
	case []int:
		for _, n := range a {
			out = append(out, q.Options[n])
		}
	}
	slices.Sort(out)
	return out
}

// viaJSON round-trips v the way an answer arrives from the browser.
func viaJSON(t *testing.T, v any) any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
