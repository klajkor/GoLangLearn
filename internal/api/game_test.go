package api

import (
	"net/http"
	"testing"
	"testing/fstest"
	"time"

	"golearn/internal/content"
	"golearn/internal/store"
)

// newGameMux serves three chapters: intro (two lessons + quiz), types (one
// lesson, no quiz) and later (no content yet). The returned pointer controls
// the clock.
func newGameMux(t *testing.T) (*http.ServeMux, *time.Time) {
	t.Helper()
	cur, err := content.Load(fstest.MapFS{
		"01-intro/chapter.yaml":        {Data: []byte("title: Intro\ntier: beginner\n")},
		"01-intro/lessons/01-hello.md": {Data: []byte("---\ntitle: Hello\n---\nHi")},
		"01-intro/lessons/02-more.md":  {Data: []byte("---\ntitle: More\n---\nMore")},
		"01-intro/quiz.yaml":           {Data: []byte(testQuiz)},
		"02-types/chapter.yaml":        {Data: []byte("title: Types\ntier: beginner\n")},
		"02-types/lessons/01-vars.md":  {Data: []byte("---\ntitle: Vars\n---\nVars")},
		"03-later/chapter.yaml":        {Data: []byte("title: Later\ntier: intermediate\n")},
	})
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	srv := New(cur, st, &fakeRunner{})
	now := time.Date(2026, 3, 10, 9, 0, 0, 0, time.Local)
	srv.now = func() time.Time { return now }
	mux := http.NewServeMux()
	srv.Register(mux)
	return mux, &now
}

const (
	perfectAnswers = `{"answers":{"q1":1,"q2":"1\n"}}`
	halfAnswers    = `{"answers":{"q1":1}}`
	wrongAnswers   = `{"answers":{"q1":0}}`
)

func numberAt(t *testing.T, m map[string]any, path ...string) float64 {
	t.Helper()
	var cur any = m
	for _, k := range path {
		cur = cur.(map[string]any)[k]
	}
	f, ok := cur.(float64)
	if !ok {
		t.Fatalf("%v is %T, want number", path, cur)
	}
	return f
}

func rewardOf(m map[string]any) map[string]any { return m["reward"].(map[string]any) }

func badgeIDs(reward map[string]any) map[string]bool {
	ids := map[string]bool{}
	for _, b := range reward["newBadges"].([]any) {
		ids[b.(map[string]any)["id"].(string)] = true
	}
	return ids
}

func TestCompleteLessonAwardsXPOnce(t *testing.T) {
	mux, _ := newGameMux(t)

	rec, rw := post(t, mux, "/api/chapters/intro/lessons/hello/complete", "")
	if rec.Code != http.StatusOK || rw["xpGained"] != float64(20) || rw["totalXP"] != float64(20) {
		t.Fatalf("first completion: status %d, %v", rec.Code, rw)
	}
	if !badgeIDs(rw)["first-steps"] {
		t.Errorf("expected First Steps badge, got %v", rw["newBadges"])
	}

	_, again := post(t, mux, "/api/chapters/intro/lessons/hello/complete", "")
	if again["xpGained"] != float64(0) || again["totalXP"] != float64(20) || len(again["newBadges"].([]any)) != 0 {
		t.Errorf("repeat completion must not pay out again: %v", again)
	}

	for _, p := range []string{"/api/chapters/nope/lessons/hello/complete", "/api/chapters/intro/lessons/nope/complete"} {
		if rec, _ := post(t, mux, p, ""); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: status = %d, want 404", p, rec.Code)
		}
	}

	_, prog := get(t, mux, "/api/progress")
	done := prog["completedLessons"].(map[string]any)["intro"].([]any)
	if len(done) != 1 || done[0] != "hello" {
		t.Errorf("completedLessons = %v", done)
	}
}

func TestLevelUpIsReported(t *testing.T) {
	mux, _ := newGameMux(t)
	// 20 (lesson) + 50 (quiz pass) + 25 (perfect) = 95 XP, just under level 2.
	post(t, mux, "/api/chapters/intro/lessons/hello/complete", "")
	_, quiz := post(t, mux, "/api/chapters/intro/quiz", perfectAnswers)
	if numberAt(t, quiz, "reward", "totalXP") != 95 || rewardOf(quiz)["leveledUp"] != false {
		t.Fatalf("after quiz: %v", quiz["reward"])
	}
	_, rw := post(t, mux, "/api/chapters/intro/lessons/more/complete", "")
	if rw["leveledUp"] != true || numberAt(t, rw, "level", "level") != 2 {
		t.Errorf("expected level up to 2: %v", rw)
	}
}

func TestQuizXP(t *testing.T) {
	mux, _ := newGameMux(t)

	_, r := post(t, mux, "/api/chapters/intro/quiz", wrongAnswers)
	if numberAt(t, r, "reward", "xpGained") != 0 {
		t.Fatalf("failed quiz paid out: %v", r["reward"])
	}
	// Passing (50%) on a retry pays 25, not 50, and isn't perfect.
	_, r = post(t, mux, "/api/chapters/intro/quiz", halfAnswers)
	if numberAt(t, r, "reward", "xpGained") != 25 {
		t.Fatalf("retry pass should pay 25: %v", r["reward"])
	}
	if !badgeIDs(rewardOf(r))["comeback-kid"] {
		t.Errorf("expected Comeback Kid: %v", r["reward"])
	}
	// Perfect: the bonus is paid once; the pass itself was already paid.
	_, r = post(t, mux, "/api/chapters/intro/quiz", perfectAnswers)
	if numberAt(t, r, "reward", "xpGained") != 25 {
		t.Fatalf("perfect bonus: %v", r["reward"])
	}
	_, r = post(t, mux, "/api/chapters/intro/quiz", perfectAnswers)
	if numberAt(t, r, "reward", "xpGained") != 0 {
		t.Errorf("repeat perfect must not pay again: %v", r["reward"])
	}
}

func TestFirstTryPerfectQuizPaysPassAndBonus(t *testing.T) {
	mux, _ := newGameMux(t)
	_, r := post(t, mux, "/api/chapters/intro/quiz", perfectAnswers)
	if numberAt(t, r, "reward", "xpGained") != 75 {
		t.Errorf("first-try perfect should pay 50+25: %v", r["reward"])
	}
	ids := badgeIDs(rewardOf(r))
	if !ids["quiz-taker"] || !ids["perfectionist"] || ids["comeback-kid"] {
		t.Errorf("badges = %v", ids)
	}
}

func TestRunXPNeedsRealLessonAndSuccess(t *testing.T) {
	mux, _ := newGameMux(t)

	_, r := post(t, mux, "/api/run", `{"code":"x"}`)
	if _, has := r["reward"]; has {
		t.Errorf("run without lesson context got a reward: %v", r["reward"])
	}
	_, r = post(t, mux, "/api/run", `{"code":"x","chapter":"intro","lesson":"nope"}`)
	if _, has := r["reward"]; has {
		t.Errorf("run for unknown lesson got a reward: %v", r["reward"])
	}

	_, r = post(t, mux, "/api/run", `{"code":"x","chapter":"intro","lesson":"hello"}`)
	if numberAt(t, r, "reward", "xpGained") != 5 || !badgeIDs(rewardOf(r))["hello-gopher"] {
		t.Errorf("first run: %v", r["reward"])
	}
	if events := r["events"].([]any); len(events) != 1 {
		t.Errorf("run result lost its events: %v", r)
	}
	_, r = post(t, mux, "/api/run", `{"code":"y","chapter":"intro","lesson":"hello"}`)
	if numberAt(t, r, "reward", "xpGained") != 0 {
		t.Errorf("second run in the same lesson paid again: %v", r["reward"])
	}
}

func TestChapterStatusUnlocksInOrder(t *testing.T) {
	mux, _ := newGameMux(t)
	status := func() map[string]any {
		_, p := get(t, mux, "/api/progress")
		return p["chapterStatus"].(map[string]any)
	}

	if s := status(); s["intro"] != "available" || s["types"] != "locked" || s["later"] != "soon" {
		t.Fatalf("fresh state: %v", s)
	}

	post(t, mux, "/api/chapters/intro/lessons/hello/complete", "")
	if s := status(); s["intro"] != "in-progress" || s["types"] != "locked" {
		t.Errorf("after one lesson: %v", s)
	}

	post(t, mux, "/api/chapters/intro/lessons/more/complete", "")
	if s := status(); s["intro"] != "in-progress" {
		t.Errorf("lessons done but quiz not passed must not be mastered: %v", s)
	}

	_, r := post(t, mux, "/api/chapters/intro/quiz", perfectAnswers)
	if s := status(); s["intro"] != "mastered" || s["types"] != "available" {
		t.Errorf("after quiz: %v", s)
	}
	if !badgeIDs(rewardOf(r))["chapter-champion"] {
		t.Errorf("mastering a chapter should earn Chapter Champion: %v", r["reward"])
	}

	post(t, mux, "/api/chapters/types/lessons/vars/complete", "")
	if s := status(); s["types"] != "mastered" || s["later"] != "soon" {
		t.Errorf("types has no quiz, so finishing its lessons masters it: %v", s)
	}
}

func TestStreakAcrossDays(t *testing.T) {
	mux, now := newGameMux(t)
	post(t, mux, "/api/chapters/intro/lessons/hello/complete", "")
	*now = now.AddDate(0, 0, 1)
	post(t, mux, "/api/chapters/intro/lessons/more/complete", "")
	*now = now.AddDate(0, 0, 1)
	_, r := post(t, mux, "/api/run", `{"code":"x","chapter":"intro","lesson":"hello"}`)
	if !badgeIDs(rewardOf(r))["on-a-roll"] {
		t.Errorf("three days in a row should earn On a Roll: %v", r["reward"])
	}

	_, p := get(t, mux, "/api/progress")
	if numberAt(t, p, "streak", "current") != 3 || p["streak"].(map[string]any)["activeToday"] != true {
		t.Errorf("streak = %v", p["streak"])
	}
	if days := p["activityDays"].([]any); len(days) != 3 || days[0] != "2026-03-10" {
		t.Errorf("activityDays = %v", days)
	}
	if len(p["badgeCatalog"].([]any)) < 20 {
		t.Error("progress should include the badge catalogue")
	}
}
