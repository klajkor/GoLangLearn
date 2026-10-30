package store

import (
	"path/filepath"
	"testing"
)

func TestOpenAndEmptyProgress(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "progress.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p, err := s.Progress()
	if err != nil {
		t.Fatal(err)
	}
	if p.TotalXP != 0 || len(p.CompletedLessons) != 0 || len(p.Badges) != 0 {
		t.Errorf("expected empty progress, got %+v", p)
	}

	// Reopening must not fail: the schema is idempotent.
	s.Close()
	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s2.Close()
}

func TestProgressAggregates(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	for _, q := range []string{
		`INSERT INTO xp_events (kind, amount) VALUES ('lesson', 20), ('quiz', 50)`,
		`INSERT INTO lesson_progress (chapter_id, lesson_id) VALUES ('intro', 'hello')`,
		`INSERT INTO badges (badge_id) VALUES ('first-run')`,
	} {
		if _, err := s.db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}

	p, err := s.Progress()
	if err != nil {
		t.Fatal(err)
	}
	if p.TotalXP != 70 {
		t.Errorf("TotalXP = %d, want 70", p.TotalXP)
	}
	if got := p.CompletedLessons["intro"]; len(got) != 1 || got[0] != "hello" {
		t.Errorf("CompletedLessons = %v", p.CompletedLessons)
	}
	if len(p.Badges) != 1 || p.Badges[0] != "first-run" {
		t.Errorf("Badges = %v", p.Badges)
	}
}

func TestQuizStats(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	for _, a := range []struct {
		score, max int
		passed     bool
	}{{3, 6, false}, {5, 6, true}, {4, 6, false}} {
		if err := s.RecordQuizAttempt("intro", a.score, a.max, a.passed); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.RecordQuizAttempt("types", 1, 4, false); err != nil {
		t.Fatal(err)
	}

	p, err := s.Progress()
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Quizzes["intro"]; got != (QuizStat{Attempts: 3, BestPercent: 83, Passed: true}) {
		t.Errorf("intro = %+v", got)
	}
	if got := p.Quizzes["types"]; got != (QuizStat{Attempts: 1, BestPercent: 25, Passed: false}) {
		t.Errorf("types = %+v", got)
	}
}

func TestAwardsAreOneOff(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if ok, err := s.CompleteLesson("intro", "hello"); err != nil || !ok {
		t.Fatalf("first CompleteLesson = %v, %v; want true", ok, err)
	}
	if ok, _ := s.CompleteLesson("intro", "hello"); ok {
		t.Error("completing a lesson twice reported it as new")
	}

	if ok, _ := s.AddXP("lesson", 20, "intro/hello"); !ok {
		t.Error("first AddXP should be new")
	}
	if ok, _ := s.AddXP("lesson", 20, "intro/hello"); ok {
		t.Error("same (kind, ref) must not be awarded twice")
	}
	if ok, _ := s.AddXP("run", 5, "intro/hello"); !ok {
		t.Error("a different kind with the same ref is a separate award")
	}

	if ok, _ := s.GrantBadge("first-steps"); !ok {
		t.Error("first GrantBadge should be new")
	}
	if ok, _ := s.GrantBadge("first-steps"); ok {
		t.Error("badge granted twice")
	}

	p, err := s.Progress()
	if err != nil {
		t.Fatal(err)
	}
	if p.TotalXP != 25 || len(p.Badges) != 1 || len(p.CompletedLessons["intro"]) != 1 {
		t.Errorf("unexpected progress %+v", p)
	}
}

func TestActivityDays(t *testing.T) {
	s, _ := Open(":memory:")
	defer s.Close()
	for _, d := range []string{"2026-03-02", "2026-03-01", "2026-03-02"} {
		if err := s.MarkActive(d); err != nil {
			t.Fatal(err)
		}
	}
	days, err := s.ActivityDays()
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 2 || days[0] != "2026-03-01" || days[1] != "2026-03-02" {
		t.Errorf("days = %v, want sorted and de-duplicated", days)
	}
}

func TestCounts(t *testing.T) {
	s, _ := Open(":memory:")
	defer s.Close()

	s.AddXP("run", 5, "a/1")
	s.AddXP("run", 5, "a/2")
	s.RecordQuizAttempt("a", 2, 4, false)
	if c, _ := s.Counts(); c.RetryPassed || c.PerfectQuizzes != 0 || c.RunLessons != 2 {
		t.Errorf("after a fail: %+v", c)
	}
	s.RecordQuizAttempt("a", 4, 4, true)
	s.RecordQuizAttempt("b", 3, 3, true) // passed first time: not a comeback, but perfect
	c, err := s.Counts()
	if err != nil {
		t.Fatal(err)
	}
	if !c.RetryPassed || c.PerfectQuizzes != 2 {
		t.Errorf("after retry pass: %+v", c)
	}

	if n, _ := s.QuizAttempts("a"); n != 2 {
		t.Errorf("QuizAttempts(a) = %d, want 2", n)
	}
}
