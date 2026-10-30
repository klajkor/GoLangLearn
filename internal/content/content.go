// Package content loads the curriculum (chapters and lessons) from a
// directory tree of YAML and Markdown files.
//
// Layout:
//
//	content/
//	  01-intro/
//	    chapter.yaml          # title, tier, summary
//	    lessons/
//	      01-hello-world.md   # YAML front matter + Markdown body
package content

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Tiers in learning order.
var Tiers = []string{"beginner", "intermediate", "advanced", "expert"}

type Curriculum struct {
	Chapters []*Chapter `json:"chapters"`
	byID     map[string]*Chapter
}

type Chapter struct {
	ID      string    `json:"id"`
	Order   int       `json:"order"`
	Title   string    `json:"title" yaml:"title"`
	Tier    string    `json:"tier" yaml:"tier"`
	Summary string    `json:"summary" yaml:"summary"`
	Icon    string    `json:"icon" yaml:"icon"`
	Lessons []*Lesson `json:"lessons" yaml:"-"`

	Quiz          *Quiz `json:"-" yaml:"-"`
	QuizQuestions int   `json:"quizQuestions" yaml:"-"`
}

type Lesson struct {
	ID      string `json:"id"`
	Title   string `json:"title" yaml:"title"`
	Minutes int    `json:"minutes" yaml:"minutes"`
	Body    string `json:"body,omitempty"`
}

// Chapter returns the chapter with the given ID, or nil.
func (c *Curriculum) Chapter(id string) *Chapter {
	return c.byID[id]
}

// Lesson returns the lesson with the given ID, or nil.
func (ch *Chapter) Lesson(id string) *Lesson {
	for _, l := range ch.Lessons {
		if l.ID == id {
			return l
		}
	}
	return nil
}

// Load reads every chapter directory in fsys. Chapter and lesson IDs are
// derived from directory and file names with the numeric prefix removed,
// so "01-intro" becomes "intro"; the prefix only controls ordering.
func Load(fsys fs.FS, opts ...Option) (*Curriculum, error) {
	var cfg loadConfig
	for _, o := range opts {
		o(&cfg)
	}
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	cur := &Curriculum{byID: map[string]*Chapter{}}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ch, err := loadChapter(fsys, e.Name(), cfg)
		if err != nil {
			return nil, fmt.Errorf("chapter %s: %w", e.Name(), err)
		}
		if _, dup := cur.byID[ch.ID]; dup {
			return nil, fmt.Errorf("duplicate chapter id %q", ch.ID)
		}
		ch.Order = len(cur.Chapters) + 1
		cur.Chapters = append(cur.Chapters, ch)
		cur.byID[ch.ID] = ch
	}
	return cur, nil
}

func loadChapter(fsys fs.FS, dir string, cfg loadConfig) (*Chapter, error) {
	raw, err := fs.ReadFile(fsys, path.Join(dir, "chapter.yaml"))
	if err != nil {
		return nil, err
	}
	ch := &Chapter{ID: stripPrefix(dir)}
	if err := yaml.Unmarshal(raw, ch); err != nil {
		return nil, err
	}
	if !validTier(ch.Tier) {
		return nil, fmt.Errorf("unknown tier %q", ch.Tier)
	}

	lessonDir := path.Join(dir, "lessons")
	files, err := fs.ReadDir(fsys, lessonDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	for _, f := range files {
		if f.IsDir() || path.Ext(f.Name()) != ".md" {
			continue
		}
		raw, err := fs.ReadFile(fsys, path.Join(lessonDir, f.Name()))
		if err != nil {
			return nil, err
		}
		l, err := parseLesson(raw)
		if err != nil {
			return nil, fmt.Errorf("lesson %s: %w", f.Name(), err)
		}
		l.ID = stripPrefix(strings.TrimSuffix(f.Name(), ".md"))
		ch.Lessons = append(ch.Lessons, l)
	}

	raw, err = fs.ReadFile(fsys, path.Join(dir, "quiz.yaml"))
	switch {
	case err == nil:
		if ch.Quiz, err = parseQuiz(raw); err != nil {
			return nil, fmt.Errorf("quiz.yaml: %w", err)
		}
		if cfg.shuffleOptions {
			ch.Quiz.shuffleOptions()
		}
		ch.QuizQuestions = len(ch.Quiz.Questions)
	case !errors.Is(err, fs.ErrNotExist):
		return nil, err
	}
	return ch, nil
}

// parseLesson splits optional "---" delimited YAML front matter from the body.
func parseLesson(raw []byte) (*Lesson, error) {
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	l := &Lesson{}
	if rest, ok := bytes.CutPrefix(raw, []byte("---\n")); ok {
		meta, body, found := bytes.Cut(rest, []byte("\n---\n"))
		if !found {
			return nil, errors.New("unterminated front matter")
		}
		if err := yaml.Unmarshal(meta, l); err != nil {
			return nil, err
		}
		raw = body
	}
	l.Body = strings.TrimSpace(string(raw))
	return l, nil
}

func stripPrefix(name string) string {
	if i := strings.IndexByte(name, '-'); i > 0 && strings.Trim(name[:i], "0123456789") == "" {
		return name[i+1:]
	}
	return name
}

func validTier(t string) bool {
	for _, v := range Tiers {
		if v == t {
			return true
		}
	}
	return false
}

type loadConfig struct{ shuffleOptions bool }

// Option customises Load.
type Option func(*loadConfig)

// WithShuffledOptions reorders the options of single and multi quiz
// questions, deterministically per question.
func WithShuffledOptions() Option { return func(c *loadConfig) { c.shuffleOptions = true } }
