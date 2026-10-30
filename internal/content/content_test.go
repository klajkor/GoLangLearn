package content

import (
	"testing"
	"testing/fstest"
)

func TestLoad(t *testing.T) {
	fsys := fstest.MapFS{
		"02-types/chapter.yaml":         {Data: []byte("title: Types\ntier: beginner\n")},
		"01-intro/chapter.yaml":         {Data: []byte("title: Intro\ntier: beginner\nsummary: Start here\n")},
		"01-intro/lessons/02-second.md": {Data: []byte("---\ntitle: Second\nminutes: 5\n---\nBody two")},
		"01-intro/lessons/01-first.md":  {Data: []byte("---\r\ntitle: First\r\n---\r\n# Hello\r\n")},
		"01-intro/lessons/notes.txt":    {Data: []byte("ignored")},
		"README.md":                     {Data: []byte("not a chapter")},
	}

	cur, err := Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(cur.Chapters) != 2 || cur.Chapters[0].ID != "intro" || cur.Chapters[1].ID != "types" {
		t.Fatalf("unexpected chapters: %+v", cur.Chapters)
	}
	if cur.Chapters[1].Order != 2 {
		t.Errorf("order = %d, want 2", cur.Chapters[1].Order)
	}

	intro := cur.Chapter("intro")
	if len(intro.Lessons) != 2 {
		t.Fatalf("lessons = %d, want 2", len(intro.Lessons))
	}
	first := intro.Lesson("first")
	if first == nil || first.Title != "First" || first.Body != "# Hello" {
		t.Errorf("first lesson = %+v", first)
	}
	if second := intro.Lesson("second"); second == nil || second.Minutes != 5 {
		t.Errorf("second lesson = %+v", second)
	}
	if cur.Chapter("missing") != nil {
		t.Error("expected nil for missing chapter")
	}
}

func TestLoadRejectsUnknownTier(t *testing.T) {
	fsys := fstest.MapFS{"01-x/chapter.yaml": {Data: []byte("title: X\ntier: wizard\n")}}
	if _, err := Load(fsys); err == nil {
		t.Fatal("expected error for unknown tier")
	}
}

func TestStripPrefix(t *testing.T) {
	cases := map[string]string{"01-intro": "intro", "10-a-b": "a-b", "intro": "intro", "go-basics": "go-basics"}
	for in, want := range cases {
		if got := stripPrefix(in); got != want {
			t.Errorf("stripPrefix(%q) = %q, want %q", in, got, want)
		}
	}
}
