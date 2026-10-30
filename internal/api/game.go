package api

import (
	"fmt"
	"log"
	"net/http"

	"golearn/internal/content"
	"golearn/internal/game"
	"golearn/internal/store"
)

// Chapter states shown on the skill map.
const (
	statusSoon       = "soon"        // no content written yet
	statusLocked     = "locked"      // an earlier chapter isn't mastered yet
	statusAvailable  = "available"   // unlocked, not started
	statusInProgress = "in-progress" // started, not mastered
	statusMastered   = "mastered"    // every lesson done and the quiz passed
)

// award is one XP grant. Each (kind, ref) can only be granted once.
type award struct {
	kind   string
	amount int
	ref    string
}

// reward tells the UI what a learner's action earned, for celebrations.
type reward struct {
	XPGained  int            `json:"xpGained"`
	TotalXP   int            `json:"totalXP"`
	Level     game.LevelInfo `json:"level"`
	LeveledUp bool           `json:"leveledUp"`
	NewBadges []game.Badge   `json:"newBadges"`
}

// chapterStatuses classifies every chapter. A chapter unlocks once every
// earlier chapter that has content is mastered; chapters without content are
// skipped over so they never block the path. Locks are advisory: the UI still
// lets learners open any chapter.
func (s *Server) chapterStatuses(p *store.Progress) map[string]string {
	out := make(map[string]string, len(s.cur.Chapters))
	prevOK := true
	for _, ch := range s.cur.Chapters {
		var st string
		switch {
		case len(ch.Lessons) == 0 && ch.Quiz == nil:
			st = statusSoon
		case s.mastered(ch, p):
			st = statusMastered
		case len(p.CompletedLessons[ch.ID]) > 0 || p.Quizzes[ch.ID].Attempts > 0:
			st = statusInProgress
		case prevOK:
			st = statusAvailable
		default:
			st = statusLocked
		}
		out[ch.ID] = st
		prevOK = st == statusMastered || (st == statusSoon && prevOK)
	}
	return out
}

func (s *Server) mastered(ch *content.Chapter, p *store.Progress) bool {
	done := map[string]bool{}
	for _, id := range p.CompletedLessons[ch.ID] {
		done[id] = true
	}
	for _, l := range ch.Lessons {
		if !done[l.ID] {
			return false
		}
	}
	return ch.Quiz == nil || p.Quizzes[ch.ID].Passed
}

// snapshot loads everything the badge rules and the UI need in one place.
func (s *Server) snapshot() (*store.Progress, game.State, map[string]string, game.Streak, error) {
	var st game.State
	p, err := s.store.Progress()
	if err != nil {
		return nil, st, nil, game.Streak{}, err
	}
	counts, err := s.store.Counts()
	if err != nil {
		return nil, st, nil, game.Streak{}, err
	}
	days, err := s.store.ActivityDays()
	if err != nil {
		return nil, st, nil, game.Streak{}, err
	}
	streak := game.ComputeStreak(days, s.now())
	statuses := s.chapterStatuses(p)

	st = game.State{
		Level:          game.LevelFor(p.TotalXP).Level,
		RunLessons:     counts.RunLessons,
		PerfectQuizzes: counts.PerfectQuizzes,
		RetryPassed:    counts.RetryPassed,
		LongestStreak:  streak.Longest,
		MasteredTiers:  map[string]bool{},
	}
	for _, l := range p.CompletedLessons {
		st.LessonsDone += len(l)
	}
	for _, q := range p.Quizzes {
		if q.Passed {
			st.QuizzesPassed++
		}
	}
	tierTotal, tierMastered := map[string]int{}, map[string]int{}
	for _, ch := range s.cur.Chapters {
		tierTotal[ch.Tier]++
		if statuses[ch.ID] == statusMastered {
			st.ChaptersMastered++
			tierMastered[ch.Tier]++
		}
	}
	for tier, n := range tierTotal {
		st.MasteredTiers[tier] = tierMastered[tier] == n
	}
	return p, st, statuses, streak, nil
}

// grant applies awards, records today's activity, hands out any newly earned
// badges and reports what changed.
func (s *Server) grant(awards ...award) (*reward, error) {
	before, err := s.store.Progress()
	if err != nil {
		return nil, err
	}
	for _, a := range awards {
		if _, err := s.store.AddXP(a.kind, a.amount, a.ref); err != nil {
			return nil, err
		}
	}
	if err := s.store.MarkActive(game.Day(s.now())); err != nil {
		return nil, err
	}

	p, state, _, _, err := s.snapshot()
	if err != nil {
		return nil, err
	}
	have := make(map[string]bool, len(p.Badges))
	for _, id := range p.Badges {
		have[id] = true
	}
	earned := []game.Badge{}
	for _, b := range game.NewlyEarned(state, have) {
		if ok, err := s.store.GrantBadge(b.ID); err != nil {
			return nil, err
		} else if ok {
			earned = append(earned, b)
		}
	}

	oldLevel := game.LevelFor(before.TotalXP)
	level := game.LevelFor(p.TotalXP)
	return &reward{
		XPGained:  p.TotalXP - before.TotalXP,
		TotalXP:   p.TotalXP,
		Level:     level,
		LeveledUp: level.Level > oldLevel.Level,
		NewBadges: earned,
	}, nil
}

// progressResponse is the learner's full state for the dashboard and map.
type progressResponse struct {
	*store.Progress
	Level         game.LevelInfo    `json:"level"`
	Streak        game.Streak       `json:"streak"`
	ActivityDays  []string          `json:"activityDays"`
	ChapterStatus map[string]string `json:"chapterStatus"`
	BadgeCatalog  []game.Badge      `json:"badgeCatalog"`
}

func (s *Server) progress(w http.ResponseWriter, r *http.Request) {
	p, _, statuses, streak, err := s.snapshot()
	if err != nil {
		log.Printf("progress: %v", err)
		writeError(w, http.StatusInternalServerError, "could not load progress")
		return
	}
	days, err := s.store.ActivityDays()
	if err != nil {
		log.Printf("progress: %v", err)
		writeError(w, http.StatusInternalServerError, "could not load progress")
		return
	}
	writeJSON(w, http.StatusOK, progressResponse{
		Progress:      p,
		Level:         game.LevelFor(p.TotalXP),
		Streak:        streak,
		ActivityDays:  days,
		ChapterStatus: statuses,
		BadgeCatalog:  game.Badges,
	})
}

// completeLesson marks a lesson done (idempotent) and awards XP the first time.
func (s *Server) completeLesson(w http.ResponseWriter, r *http.Request) {
	ch := s.cur.Chapter(r.PathValue("chapter"))
	if ch == nil || ch.Lesson(r.PathValue("lesson")) == nil {
		writeError(w, http.StatusNotFound, "lesson not found")
		return
	}
	lessonID := r.PathValue("lesson")
	if _, err := s.store.CompleteLesson(ch.ID, lessonID); err != nil {
		log.Printf("complete lesson: %v", err)
		writeError(w, http.StatusInternalServerError, "could not save progress")
		return
	}
	rw, err := s.grant(award{"lesson", game.XPLesson, lessonRef(ch.ID, lessonID)})
	if err != nil {
		log.Printf("complete lesson: %v", err)
		writeError(w, http.StatusInternalServerError, "could not save progress")
		return
	}
	writeJSON(w, http.StatusOK, rw)
}

func lessonRef(chapterID, lessonID string) string { return fmt.Sprintf("%s/%s", chapterID, lessonID) }

// lessonExists reports whether ref ("chapter/lesson") names a real lesson.
func (s *Server) lessonExists(chapterID, lessonID string) bool {
	ch := s.cur.Chapter(chapterID)
	return ch != nil && ch.Lesson(lessonID) != nil
}
