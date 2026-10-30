// Package game holds the pure gamification rules: XP amounts, levels,
// streaks and badges. It knows nothing about storage or HTTP.
package game

import "time"

// XP awarded for each kind of progress. Every award is one-off per subject
// (lesson, chapter quiz), so nothing can be farmed by repeating it.
const (
	XPLesson        = 20 // marking a lesson complete
	XPRun           = 5  // first successful run of an example in a lesson
	XPQuizPass      = 50 // first pass of a chapter quiz, on the first attempt
	XPQuizRetryPass = 25 // first pass of a chapter quiz after failing before
	XPQuizPerfect   = 25 // bonus for the first 100% on a chapter quiz
)

// levelTitles are Go-themed names for levels 1..len; higher levels keep the last.
var levelTitles = []string{
	"Gopher Hatchling",
	"Curious Gopher",
	"Syntax Scout",
	"Func Apprentice",
	"Goroutine Wrangler",
	"Interface Artisan",
	"Channel Master",
	"Concurrency Sage",
	"Runtime Whisperer",
	"Go Guru",
}

// threshold returns the total XP needed to reach level l (level 1 is 0 XP).
// The curve is quadratic: 100, 250, 450, 700, … so each level takes a bit
// more effort than the last.
func threshold(l int) int { return 25 * (l - 1) * (l + 2) }

// LevelInfo describes where a total XP amount sits on the level curve.
type LevelInfo struct {
	Level   int    `json:"level"`
	Title   string `json:"title"`
	TotalXP int    `json:"totalXP"`
	Floor   int    `json:"floor"` // XP at which this level started
	Next    int    `json:"next"`  // XP at which the next level starts
}

// LevelFor maps total XP to a level.
func LevelFor(xp int) LevelInfo {
	if xp < 0 {
		xp = 0
	}
	l := 1
	for threshold(l+1) <= xp {
		l++
	}
	title := levelTitles[len(levelTitles)-1]
	if l <= len(levelTitles) {
		title = levelTitles[l-1]
	}
	return LevelInfo{Level: l, Title: title, TotalXP: xp, Floor: threshold(l), Next: threshold(l + 1)}
}

// Streak summarises consecutive-day learning activity.
type Streak struct {
	Current     int  `json:"current"`
	Longest     int  `json:"longest"`
	ActiveToday bool `json:"activeToday"`
	// FreezeReady is true when a missed day would still be forgiven: one
	// freeze is available per rolling week.
	FreezeReady bool `json:"freezeReady"`
}

const dayLayout = "2006-01-02"

// Day formats t as the YYYY-MM-DD key used for activity days (local time).
func Day(t time.Time) string { return t.Format(dayLayout) }

// ComputeStreak derives the streak from the set of active days (YYYY-MM-DD)
// as of today. A streak survives a single missed day once per week (the
// "freeze"), as long as the learner comes back the day after. Today being
// empty does not break a streak: the day isn't over yet.
func ComputeStreak(days []string, today time.Time) Streak {
	active := make(map[string]bool, len(days))
	for _, d := range days {
		active[d] = true
	}
	today = time.Date(today.Year(), today.Month(), today.Day(), 12, 0, 0, 0, time.UTC)

	cur, freezeReady := streakEnding(active, today)
	st := Streak{Current: cur, ActiveToday: active[Day(today)], FreezeReady: freezeReady}
	st.Longest = st.Current
	for _, d := range days {
		t, err := time.Parse(dayLayout, d)
		if err != nil {
			continue
		}
		if n, _ := streakEnding(active, t.Add(12*time.Hour)); n > st.Longest {
			st.Longest = n
		}
	}
	return st
}

// streakEnding counts active days walking back from end. A day with no
// activity is only skipped when it's end itself, or when it's a single gap
// (freeze) not within a week of the previous freeze and preceded by activity.
func streakEnding(active map[string]bool, end time.Time) (count int, freezeReady bool) {
	d := end
	if !active[Day(d)] {
		d = d.AddDate(0, 0, -1) // today isn't over yet
	}
	var lastFreeze time.Time
	freezeReady = true
	for {
		switch {
		case active[Day(d)]:
			count++
		case active[Day(d.AddDate(0, 0, -1))] && (lastFreeze.IsZero() || lastFreeze.Sub(d) >= 7*24*time.Hour):
			lastFreeze = d
			if end.Sub(d) < 7*24*time.Hour {
				freezeReady = false
			}
		default:
			return count, freezeReady
		}
		d = d.AddDate(0, 0, -1)
	}
}
