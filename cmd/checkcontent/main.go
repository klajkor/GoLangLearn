// Command checkcontent verifies the code in the course content:
//
//   - every ```go run example in a lesson compiles, passes go vet and exits 0;
//   - every quiz "output" question's answer matches what its program prints;
//   - every quiz "fill" question compiles once its gaps are filled in.
//
// Usage: go run ./cmd/checkcontent [-content content] [chapter-id ...]
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"golearn/internal/content"
)

type job struct {
	name string
	code string
	// wantOutput, when non-nil, is compared with the program's output.
	wantOutput *string
	// compileOnly skips running (used for fill-in questions).
	compileOnly bool
}

var fence = regexp.MustCompile("(?ms)^```go run[^\n]*\n(.*?)^```")

func main() {
	dir := flag.String("content", "content", "content directory")
	flag.Parse()
	only := map[string]bool{}
	for _, a := range flag.Args() {
		only[a] = true
	}

	cur, err := content.Load(os.DirFS(*dir))
	if err != nil {
		fmt.Fprintln(os.Stderr, "load content:", err)
		os.Exit(1)
	}

	var jobs []job
	for _, ch := range cur.Chapters {
		if len(only) > 0 && !only[ch.ID] && !only[fmt.Sprintf("%02d-%s", ch.Order, ch.ID)] {
			continue
		}
		for _, l := range ch.Lessons {
			for i, m := range fence.FindAllStringSubmatch(l.Body, -1) {
				jobs = append(jobs, job{name: fmt.Sprintf("%s/%s example %d", ch.ID, l.ID, i+1), code: m[1]})
			}
		}
		if ch.Quiz == nil {
			continue
		}
		for _, q := range ch.Quiz.Questions {
			name := fmt.Sprintf("%s quiz %s", ch.ID, q.ID)
			switch q.Type {
			case "output":
				want, _ := q.Answer.(string)
				jobs = append(jobs, job{name: name, code: q.Code, wantOutput: &want})
			case "fill":
				answers, _ := q.CorrectAnswer().([]string)
				code := q.Code
				for _, a := range answers {
					code = strings.Replace(code, content.BlankMarker, a, 1)
				}
				jobs = append(jobs, job{name: name, code: code, compileOnly: true})
			}
		}
	}

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		failed int
		sem    = make(chan struct{}, 4)
	)
	for _, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if msg := check(j); msg != "" {
				mu.Lock()
				failed++
				fmt.Printf("FAIL %s\n%s\n\n", j.name, indent(msg))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	fmt.Printf("checked %d programs, %d failed\n", len(jobs), failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func indent(s string) string {
	return "    " + strings.ReplaceAll(strings.TrimSpace(s), "\n", "\n    ")
}

// check returns "" on success or a description of the failure.
func check(j job) string {
	dir, err := os.MkdirTemp("", "checkcontent")
	if err != nil {
		return err.Error()
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte(j.code), fs.FileMode(0o644)); err != nil {
		return err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if out, err := command(ctx, dir, "go", "vet", file); err != nil {
		return "go vet:\n" + out
	}
	if j.compileOnly {
		return ""
	}
	out, err := command(ctx, dir, "go", "run", file)
	if err != nil {
		return "go run:\n" + out
	}
	if j.wantOutput != nil && normalize(out) != normalize(*j.wantOutput) {
		return fmt.Sprintf("output mismatch\n--- quiz answer:\n%s\n--- program printed:\n%s", normalize(*j.wantOutput), normalize(out))
	}
	return ""
}

func command(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	err := cmd.Run()
	return buf.String(), err
}

// normalize trims trailing space on each line and blank lines at the ends,
// matching how quiz output answers are graded.
func normalize(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
