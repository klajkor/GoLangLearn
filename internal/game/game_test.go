package game

import (
	"testing"
	"time"
)

func TestLevelFor(t *testing.T) {
	tests := []struct {
		xp    int
		level int
		title string
	}{
		{-5, 1, "Gopher Hatchling"},
		{0, 1, "Gopher Hatchling"},
		{99, 1, "Gopher Hatchling"},
		{100, 2, "Curious Gopher"},
		{249, 2, "Curious Gopher"},
		{250, 3, "Syntax Scout"},
		{700, 5, "Goroutine Wrangler"},
		{1_000_000, 199, "Go Guru"}, // beyond the title list the last one sticks
	}
	for _, tt := range tests {
		got := LevelFor(tt.xp)
		if got.Level != tt.level || got.Title != tt.title {
			t.Errorf("LevelFor(%d) = level %d %q, want %d %q", tt.xp, got.Level, got.Title, tt.level, tt.title)
		}
		if got.TotalXP >= 0 && (got.TotalXP < got.Floor || got.TotalXP >= got.Next) {
			t.Errorf("LevelFor(%d): xp outside [%d, %d)", tt.xp, got.Floor, got.Next)
		}
	}
}

func day(s string) time.Time {
	t, err := time.Parse(dayLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestComputeStreak(t *testing.T) {
	tests := []struct {
		name    string
		days    []string
		today   string
		current int
		longest int
		active  bool
		freeze  bool
	}{
		{"no activity", nil, "2026-03-10", 0, 0, false, true},
		{"only today", []string{"2026-03-10"}, "2026-03-10", 1, 1, true, true},
		{"three in a row", []string{"2026-03-08", "2026-03-09", "2026-03-10"}, "2026-03-10", 3, 3, true, true},
		{"today not started yet keeps streak", []string{"2026-03-08", "2026-03-09"}, "2026-03-10", 2, 2, false, true},
		{"two days off breaks it", []string{"2026-03-06", "2026-03-07"}, "2026-03-10", 0, 2, false, true},
		{"one missed day is frozen", []string{"2026-03-07", "2026-03-08", "2026-03-10"}, "2026-03-10", 3, 3, true, false},
		{"yesterday missed, saved by freeze", []string{"2026-03-08"}, "2026-03-10", 1, 1, false, false},
		{
			"second gap within a week is not forgiven",
			[]string{"2026-03-01", "2026-03-02", "2026-03-04", "2026-03-06", "2026-03-07"},
			"2026-03-07", 3, 3, true, false,
		},
		{
			"gaps a week apart are both forgiven",
			[]string{"2026-03-01", "2026-03-03", "2026-03-04", "2026-03-05", "2026-03-06", "2026-03-07", "2026-03-08", "2026-03-10", "2026-03-11"},
			"2026-03-11", 9, 9, true, false,
		},
		{"old record is remembered as longest", []string{"2026-01-01", "2026-01-02", "2026-01-03", "2026-01-04", "2026-03-10"}, "2026-03-10", 1, 4, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeStreak(tt.days, day(tt.today))
			if got.Current != tt.current || got.Longest != tt.longest || got.ActiveToday != tt.active || got.FreezeReady != tt.freeze {
				t.Errorf("got %+v, want current=%d longest=%d active=%v freeze=%v", got, tt.current, tt.longest, tt.active, tt.freeze)
			}
		})
	}
}

func TestNewlyEarned(t *testing.T) {
	ids := func(bs []Badge) map[string]bool {
		m := map[string]bool{}
		for _, b := range bs {
			m[b.ID] = true
		}
		return m
	}

	got := ids(NewlyEarned(State{Level: 1, LessonsDone: 5, RunLessons: 1}, nil))
	for _, want := range []string{"first-steps", "bookworm", "hello-gopher"} {
		if !got[want] {
			t.Errorf("expected badge %q to be earned", want)
		}
	}
	if len(got) != 3 {
		t.Errorf("earned %v, want exactly 3 badges", got)
	}

	// Already-held badges are not reported again.
	again := ids(NewlyEarned(State{Level: 1, LessonsDone: 5, RunLessons: 1}, map[string]bool{"first-steps": true, "bookworm": true}))
	if len(again) != 1 || !again["hello-gopher"] {
		t.Errorf("with two held, got %v, want only hello-gopher", again)
	}

	tiers := ids(NewlyEarned(State{Level: 1, MasteredTiers: map[string]bool{"beginner": true}}, nil))
	if !tiers["beginner-graduate"] || tiers["intermediate-graduate"] {
		t.Errorf("tier badges wrong: %v", tiers)
	}
}

func TestBadgeCatalogueIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range Badges {
		if b.ID == "" || b.Name == "" || b.Description == "" || b.Icon == "" || b.earned == nil {
			t.Errorf("badge %+v is incomplete", b)
		}
		if seen[b.ID] {
			t.Errorf("duplicate badge ID %q", b.ID)
		}
		seen[b.ID] = true
		if got, ok := BadgeByID(b.ID); !ok || got.ID != b.ID {
			t.Errorf("BadgeByID(%q) failed", b.ID)
		}
	}
	if len(Badges) < 20 {
		t.Errorf("only %d badges, plan calls for ~20", len(Badges))
	}
}
