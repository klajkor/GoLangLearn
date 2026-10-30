// Package store persists learner progress in a local SQLite database.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// schema is applied on every start; statements must be idempotent.
const schema = `
CREATE TABLE IF NOT EXISTS lesson_progress (
	chapter_id   TEXT NOT NULL,
	lesson_id    TEXT NOT NULL,
	completed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (chapter_id, lesson_id)
);
CREATE TABLE IF NOT EXISTS quiz_attempts (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	chapter_id TEXT NOT NULL,
	score      INTEGER NOT NULL,
	max_score  INTEGER NOT NULL,
	passed     BOOLEAN NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS xp_events (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	kind       TEXT NOT NULL,
	amount     INTEGER NOT NULL,
	ref        TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS xp_events_kind_ref ON xp_events (kind, ref);
CREATE TABLE IF NOT EXISTS badges (
	badge_id  TEXT PRIMARY KEY,
	earned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS activity_days (
	day TEXT PRIMARY KEY -- YYYY-MM-DD, local time
);
`

type Store struct {
	db *sql.DB
}

// Open creates (if needed) and opens the database at path.
func Open(path string) (*Store, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	// SQLite allows one writer; a single connection avoids "database is locked".
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// QuizStat summarises all attempts at one chapter's quiz.
type QuizStat struct {
	Attempts    int  `json:"attempts"`
	BestPercent int  `json:"bestPercent"`
	Passed      bool `json:"passed"`
}

// Progress is a summary of the learner's state.
type Progress struct {
	TotalXP          int                 `json:"totalXP"`
	CompletedLessons map[string][]string `json:"completedLessons"` // chapter ID -> lesson IDs
	Quizzes          map[string]QuizStat `json:"quizzes"`          // chapter ID -> stats
	Badges           []string            `json:"badges"`
}

// RecordQuizAttempt saves one graded attempt. maxScore must be positive.
func (s *Store) RecordQuizAttempt(chapterID string, score, maxScore int, passed bool) error {
	_, err := s.db.Exec(
		`INSERT INTO quiz_attempts (chapter_id, score, max_score, passed) VALUES (?, ?, ?, ?)`,
		chapterID, score, maxScore, passed,
	)
	return err
}

func (s *Store) Progress() (*Progress, error) {
	p := &Progress{CompletedLessons: map[string][]string{}, Quizzes: map[string]QuizStat{}, Badges: []string{}}

	if err := s.db.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM xp_events`).Scan(&p.TotalXP); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`SELECT chapter_id, lesson_id FROM lesson_progress ORDER BY completed_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ch, l string
		if err := rows.Scan(&ch, &l); err != nil {
			return nil, err
		}
		p.CompletedLessons[ch] = append(p.CompletedLessons[ch], l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	qrows, err := s.db.Query(`
		SELECT chapter_id, COUNT(*), MAX(score * 100 / max_score), MAX(passed)
		FROM quiz_attempts GROUP BY chapter_id`)
	if err != nil {
		return nil, err
	}
	defer qrows.Close()
	for qrows.Next() {
		var ch string
		var st QuizStat
		if err := qrows.Scan(&ch, &st.Attempts, &st.BestPercent, &st.Passed); err != nil {
			return nil, err
		}
		p.Quizzes[ch] = st
	}
	if err := qrows.Err(); err != nil {
		return nil, err
	}

	brows, err := s.db.Query(`SELECT badge_id FROM badges ORDER BY earned_at, rowid`)
	if err != nil {
		return nil, err
	}
	defer brows.Close()
	for brows.Next() {
		var b string
		if err := brows.Scan(&b); err != nil {
			return nil, err
		}
		p.Badges = append(p.Badges, b)
	}
	return p, brows.Err()
}

// CompleteLesson marks a lesson done. It reports whether it was newly completed.
func (s *Store) CompleteLesson(chapterID, lessonID string) (bool, error) {
	return s.insertOnce(`INSERT OR IGNORE INTO lesson_progress (chapter_id, lesson_id) VALUES (?, ?)`, chapterID, lessonID)
}

// AddXP records an XP award. Each (kind, ref) pair can only be awarded once;
// it reports whether this call created the award.
func (s *Store) AddXP(kind string, amount int, ref string) (bool, error) {
	return s.insertOnce(`INSERT OR IGNORE INTO xp_events (kind, amount, ref) VALUES (?, ?, ?)`, kind, amount, ref)
}

// GrantBadge records an earned badge, reporting whether it was new.
func (s *Store) GrantBadge(id string) (bool, error) {
	return s.insertOnce(`INSERT OR IGNORE INTO badges (badge_id) VALUES (?)`, id)
}

// MarkActive records that the learner was active on day (YYYY-MM-DD).
func (s *Store) MarkActive(day string) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO activity_days (day) VALUES (?)`, day)
	return err
}

func (s *Store) insertOnce(query string, args ...any) (bool, error) {
	res, err := s.db.Exec(query, args...)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// ActivityDays lists every active day, oldest first.
func (s *Store) ActivityDays() ([]string, error) {
	rows, err := s.db.Query(`SELECT day FROM activity_days ORDER BY day`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	days := []string{}
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		days = append(days, d)
	}
	return days, rows.Err()
}

// QuizAttempts returns how many times chapterID's quiz has been attempted.
func (s *Store) QuizAttempts(chapterID string) (n int, err error) {
	err = s.db.QueryRow(`SELECT COUNT(*) FROM quiz_attempts WHERE chapter_id = ?`, chapterID).Scan(&n)
	return n, err
}

// Counts holds aggregate figures used by badge rules.
type Counts struct {
	RunLessons     int  // lessons whose examples ran successfully
	PerfectQuizzes int  // chapters with a 100% attempt
	RetryPassed    bool // some quiz was passed after an earlier failure
}

func (s *Store) Counts() (c Counts, err error) {
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM xp_events WHERE kind = 'run'`).Scan(&c.RunLessons); err != nil {
		return
	}
	if err = s.db.QueryRow(`SELECT COUNT(DISTINCT chapter_id) FROM quiz_attempts WHERE score = max_score`).Scan(&c.PerfectQuizzes); err != nil {
		return
	}
	err = s.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM quiz_attempts a
			WHERE a.passed AND EXISTS (
				SELECT 1 FROM quiz_attempts b
				WHERE b.chapter_id = a.chapter_id AND b.id < a.id AND NOT b.passed))`).Scan(&c.RetryPassed)
	return
}
