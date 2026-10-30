// Package api exposes the JSON HTTP API used by the web UI.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"go/format"
	"log"
	"net/http"
	"time"

	"golearn/internal/content"
	"golearn/internal/game"
	"golearn/internal/playground"
	"golearn/internal/store"
)

// maxCodeBytes caps the size of submitted programs.
const maxCodeBytes = 64 << 10

// Runner executes Go programs; *playground.Client implements it.
type Runner interface {
	Run(ctx context.Context, code string) (*playground.Result, error)
}

type Server struct {
	cur    *content.Curriculum
	store  *store.Store
	runner Runner
	now    func() time.Time // injectable for tests; drives streaks
}

func New(cur *content.Curriculum, st *store.Store, runner Runner) *Server {
	return &Server{cur: cur, store: st, runner: runner, now: time.Now}
}

// Register mounts all API routes on mux under /api/.
func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/curriculum", s.curriculum)
	mux.HandleFunc("GET /api/chapters/{chapter}/lessons/{lesson}", s.lesson)
	mux.HandleFunc("GET /api/progress", s.progress)
	mux.HandleFunc("POST /api/chapters/{chapter}/lessons/{lesson}/complete", s.completeLesson)
	mux.HandleFunc("GET /api/chapters/{chapter}/quiz", s.quiz)
	mux.HandleFunc("POST /api/chapters/{chapter}/quiz", s.submitQuiz)
	mux.HandleFunc("POST /api/run", s.run)
	mux.HandleFunc("POST /api/format", s.format)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// curriculum returns all chapters with lesson metadata but without bodies.
func (s *Server) curriculum(w http.ResponseWriter, r *http.Request) {
	type lessonMeta struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Minutes int    `json:"minutes"`
	}
	type chapterMeta struct {
		*content.Chapter
		Lessons []lessonMeta `json:"lessons"`
	}
	out := struct {
		Tiers    []string      `json:"tiers"`
		Chapters []chapterMeta `json:"chapters"`
	}{Tiers: content.Tiers, Chapters: []chapterMeta{}}

	for _, ch := range s.cur.Chapters {
		cm := chapterMeta{Chapter: ch, Lessons: []lessonMeta{}}
		for _, l := range ch.Lessons {
			cm.Lessons = append(cm.Lessons, lessonMeta{l.ID, l.Title, l.Minutes})
		}
		out.Chapters = append(out.Chapters, cm)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) lesson(w http.ResponseWriter, r *http.Request) {
	ch := s.cur.Chapter(r.PathValue("chapter"))
	if ch == nil {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	}
	l := ch.Lesson(r.PathValue("lesson"))
	if l == nil {
		writeError(w, http.StatusNotFound, "lesson not found")
		return
	}
	writeJSON(w, http.StatusOK, l)
}

// chapterWithQuiz resolves the {chapter} path value, writing a 404 if the
// chapter or its quiz doesn't exist.
func (s *Server) chapterWithQuiz(w http.ResponseWriter, r *http.Request) (*content.Chapter, bool) {
	ch := s.cur.Chapter(r.PathValue("chapter"))
	if ch == nil {
		writeError(w, http.StatusNotFound, "chapter not found")
		return nil, false
	}
	if ch.Quiz == nil {
		writeError(w, http.StatusNotFound, "this chapter has no quiz yet")
		return nil, false
	}
	return ch, true
}

// quiz returns the questions without answers or explanations.
func (s *Server) quiz(w http.ResponseWriter, r *http.Request) {
	ch, ok := s.chapterWithQuiz(w, r)
	if !ok {
		return
	}
	out := struct {
		ChapterID   string                   `json:"chapterId"`
		PassPercent int                      `json:"passPercent"`
		Questions   []content.PublicQuestion `json:"questions"`
	}{ch.ID, ch.Quiz.PassPercent, make([]content.PublicQuestion, 0, len(ch.Quiz.Questions))}
	for _, q := range ch.Quiz.Questions {
		out.Questions = append(out.Questions, q.Public())
	}
	writeJSON(w, http.StatusOK, out)
}

type questionResult struct {
	ID            string `json:"id"`
	Correct       bool   `json:"correct"`
	CorrectAnswer any    `json:"correctAnswer"`
	Explanation   string `json:"explanation"`
}

// submitQuiz grades a full attempt on the server and records it. Questions
// left unanswered count as wrong.
func (s *Server) submitQuiz(w http.ResponseWriter, r *http.Request) {
	ch, ok := s.chapterWithQuiz(w, r)
	if !ok {
		return
	}
	var req struct {
		Answers map[string]any `json:"answers"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCodeBytes)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	qz := ch.Quiz
	results := make([]questionResult, 0, len(qz.Questions))
	score := 0
	for _, q := range qz.Questions {
		correct := q.Grade(req.Answers[q.ID])
		if correct {
			score++
		}
		results = append(results, questionResult{q.ID, correct, q.CorrectAnswer(), q.Explanation})
	}
	total := len(qz.Questions)
	passed := score*100 >= qz.PassPercent*total

	earlier, err := s.store.QuizAttempts(ch.ID)
	if err == nil {
		err = s.store.RecordQuizAttempt(ch.ID, score, total, passed)
	}
	var rw *reward
	if err == nil {
		var awards []award
		if passed {
			amount := game.XPQuizPass
			if earlier > 0 {
				amount = game.XPQuizRetryPass
			}
			awards = append(awards, award{"quiz-pass", amount, ch.ID})
		}
		if score == total {
			awards = append(awards, award{"quiz-perfect", game.XPQuizPerfect, ch.ID})
		}
		rw, err = s.grant(awards...)
	}
	if err != nil {
		log.Printf("record quiz attempt: %v", err)
		writeError(w, http.StatusInternalServerError, "could not save your result")
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Score       int              `json:"score"`
		Total       int              `json:"total"`
		Percent     int              `json:"percent"`
		PassPercent int              `json:"passPercent"`
		Passed      bool             `json:"passed"`
		Results     []questionResult `json:"results"`
		Reward      *reward          `json:"reward"`
	}{score, total, score * 100 / total, qz.PassPercent, passed, results, rw})
}

type codeRequest struct {
	Code string `json:"code"`
}

// decodeCode reads a {"code": "..."} body, writing an error response on failure.
func decodeCode(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req codeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCodeBytes)).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "program is too large (max 64 KB)")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return "", false
	}
	return req.Code, true
}

type runRequest struct {
	Code string `json:"code"`
	// Chapter and Lesson say which lesson's example this is, so a successful
	// run can earn XP. Both are optional.
	Chapter string `json:"chapter"`
	Lesson  string `json:"lesson"`
}

func (s *Server) run(w http.ResponseWriter, r *http.Request) {
	var req runRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCodeBytes)).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "program is too large (max 64 KB)")
		} else {
			writeError(w, http.StatusBadRequest, "invalid request body")
		}
		return
	}
	code := req.Code
	res, err := s.runner.Run(r.Context(), code)
	if err != nil {
		log.Printf("run: %v", err)
		if errors.Is(err, playground.ErrUnavailable) {
			writeError(w, http.StatusServiceUnavailable, "Can't reach the Go Playground. Check your internet connection and try again.")
		} else {
			writeError(w, http.StatusInternalServerError, "could not run program")
		}
		return
	}

	out := struct {
		*playground.Result
		Reward *reward `json:"reward,omitempty"`
	}{Result: res}
	if res.CompileErrors == "" && res.ExitCode == 0 && s.lessonExists(req.Chapter, req.Lesson) {
		rw, err := s.grant(award{"run", game.XPRun, lessonRef(req.Chapter, req.Lesson)})
		if err != nil {
			log.Printf("run reward: %v", err) // the run itself succeeded; don't fail it
		} else {
			out.Reward = rw
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// format runs gofmt locally; it works offline.
func (s *Server) format(w http.ResponseWriter, r *http.Request) {
	code, ok := decodeCode(w, r)
	if !ok {
		return
	}
	out, err := format.Source([]byte(code))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, codeRequest{Code: string(out)})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
