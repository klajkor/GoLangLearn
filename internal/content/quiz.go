package content

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultPassPercent = 70

	// BlankMarker marks a gap to fill in a "fill" question's code.
	BlankMarker = "___"
)

// Quiz is the end-of-chapter test loaded from a chapter's quiz.yaml.
type Quiz struct {
	PassPercent int         `yaml:"passPercent"`
	Questions   []*Question `yaml:"questions"`
}

// Question types:
//
//	single  pick one option                       answer: option index
//	multi   pick every correct option             answer: [option indexes]
//	output  type what the program prints          answer: "text" or [accepted texts]
//	fill    complete the ___ gaps in code         answer: [per gap: "text" or [accepted texts]]
//	bug     click the line that contains the bug  answer: 1-based line number of Code
type Question struct {
	ID          string   `yaml:"id"`
	Type        string   `yaml:"type"`
	Prompt      string   `yaml:"prompt"`
	Code        string   `yaml:"code"`
	Options     []string `yaml:"options"`
	Answer      any      `yaml:"answer"`
	Explanation string   `yaml:"explanation"`

	// Parsed from Answer by validate.
	index    int
	set      map[int]bool
	accepted [][]string // output: one entry; fill: one entry per gap
}

// PublicQuestion is what learners see before answering: no answer, no explanation.
type PublicQuestion struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"`
	Prompt  string   `json:"prompt"`
	Code    string   `json:"code,omitempty"`
	Options []string `json:"options,omitempty"`
	Blanks  int      `json:"blanks,omitempty"`
}

func (q *Question) Public() PublicQuestion {
	p := PublicQuestion{ID: q.ID, Type: q.Type, Prompt: q.Prompt, Code: q.Code, Options: q.Options}
	if q.Type == "fill" {
		p.Blanks = strings.Count(q.Code, BlankMarker)
	}
	return p
}

// Grade reports whether ans, as decoded from JSON (float64, string, []any),
// is correct. Malformed answers are simply wrong.
func (q *Question) Grade(ans any) bool {
	switch q.Type {
	case "single", "bug":
		n, ok := ans.(float64)
		return ok && int(n) == q.index && float64(int(n)) == n
	case "multi":
		list, ok := ans.([]any)
		if !ok {
			return false
		}
		got := map[int]bool{}
		for _, v := range list {
			n, ok := v.(float64)
			if !ok {
				return false
			}
			got[int(n)] = true
		}
		return len(got) == len(q.set) && func() bool {
			for k := range got {
				if !q.set[k] {
					return false
				}
			}
			return true
		}()
	case "output":
		s, ok := ans.(string)
		return ok && matches(normalizeOutput(s), q.accepted[0], normalizeOutput)
	case "fill":
		list, ok := ans.([]any)
		if !ok || len(list) != len(q.accepted) {
			return false
		}
		for i, v := range list {
			s, ok := v.(string)
			if !ok || !matches(s, q.accepted[i], normalizeToken) {
				return false
			}
		}
		return true
	}
	return false
}

// CorrectAnswer returns the answer in the same shape the client submits.
func (q *Question) CorrectAnswer() any {
	switch q.Type {
	case "single", "bug":
		return q.index
	case "multi":
		out := make([]int, 0, len(q.set))
		for k := range q.set {
			out = append(out, k)
		}
		slices.Sort(out)
		return out
	case "output":
		return q.accepted[0][0]
	case "fill":
		out := make([]string, len(q.accepted))
		for i, alts := range q.accepted {
			out[i] = alts[0]
		}
		return out
	}
	return nil
}

func matches(got string, accepted []string, norm func(string) string) bool {
	got = norm(got)
	for _, a := range accepted {
		if got == norm(a) {
			return true
		}
	}
	return false
}

// normalizeOutput ignores line-ending style and trailing whitespace.
func normalizeOutput(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// normalizeToken ignores surrounding and repeated whitespace; case matters.
func normalizeToken(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func parseQuiz(raw []byte) (*Quiz, error) {
	qz := &Quiz{}
	if err := yaml.Unmarshal(raw, qz); err != nil {
		return nil, err
	}
	if qz.PassPercent == 0 {
		qz.PassPercent = defaultPassPercent
	}
	if qz.PassPercent < 1 || qz.PassPercent > 100 {
		return nil, fmt.Errorf("passPercent %d out of range 1-100", qz.PassPercent)
	}
	if len(qz.Questions) == 0 {
		return nil, errors.New("quiz has no questions")
	}
	seen := map[string]bool{}
	for i, q := range qz.Questions {
		if q == nil {
			return nil, fmt.Errorf("question %d is empty", i+1)
		}
		if err := q.validate(); err != nil {
			return nil, fmt.Errorf("question %d (%s): %w", i+1, q.ID, err)
		}
		if seen[q.ID] {
			return nil, fmt.Errorf("duplicate question id %q", q.ID)
		}
		seen[q.ID] = true
	}
	return qz, nil
}

func (q *Question) validate() error {
	if q.ID == "" || q.Prompt == "" {
		return errors.New("id and prompt are required")
	}
	if q.Explanation == "" {
		return errors.New("explanation is required")
	}
	switch q.Type {
	case "single":
		if len(q.Options) < 2 {
			return errors.New("needs at least 2 options")
		}
		n, ok := toInt(q.Answer)
		if !ok || n < 0 || n >= len(q.Options) {
			return fmt.Errorf("answer must be an option index 0-%d", len(q.Options)-1)
		}
		q.index = n
	case "multi":
		if len(q.Options) < 2 {
			return errors.New("needs at least 2 options")
		}
		list, ok := q.Answer.([]any)
		if !ok || len(list) == 0 {
			return errors.New("answer must be a non-empty list of option indexes")
		}
		q.set = map[int]bool{}
		for _, v := range list {
			n, ok := toInt(v)
			if !ok || n < 0 || n >= len(q.Options) {
				return fmt.Errorf("answer %v is not an option index 0-%d", v, len(q.Options)-1)
			}
			q.set[n] = true
		}
	case "output":
		if q.Code == "" {
			return errors.New("code is required")
		}
		alts, err := toStrings(q.Answer)
		if err != nil {
			return err
		}
		q.accepted = [][]string{alts}
	case "fill":
		blanks := strings.Count(q.Code, BlankMarker)
		if blanks == 0 {
			return fmt.Errorf("code needs at least one %s gap", BlankMarker)
		}
		list, ok := q.Answer.([]any)
		if !ok || len(list) != blanks {
			return fmt.Errorf("answer must be a list with one entry per gap (%d)", blanks)
		}
		for _, v := range list {
			alts, err := toStrings(v)
			if err != nil {
				return err
			}
			q.accepted = append(q.accepted, alts)
		}
	case "bug":
		if q.Code == "" {
			return errors.New("code is required")
		}
		lines := strings.Count(strings.TrimRight(q.Code, "\n"), "\n") + 1
		n, ok := toInt(q.Answer)
		if !ok || n < 1 || n > lines {
			return fmt.Errorf("answer must be a line number 1-%d", lines)
		}
		q.index = n
	default:
		return fmt.Errorf("unknown type %q", q.Type)
	}
	return nil
}

func toInt(v any) (int, bool) {
	n, ok := v.(int)
	return n, ok
}

// toStrings accepts "text" or ["text", "alternative", ...].
func toStrings(v any) ([]string, error) {
	switch x := v.(type) {
	case string:
		return []string{x}, nil
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			s, ok := e.(string)
			if !ok {
				return nil, fmt.Errorf("accepted answer %v must be a string", e)
			}
			out = append(out, s)
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	return nil, errors.New("answer must be a string or a non-empty list of strings")
}

// shuffleOptions reorders the options of single and multi questions so the
// correct answers aren't always in the same slots. The order is derived from
// the question ID, so it is stable across restarts and the answer key stays
// consistent. Call it only after validation.
func (qz *Quiz) shuffleOptions() {
	for _, q := range qz.Questions {
		if q.Type != "single" && q.Type != "multi" {
			continue
		}
		h := fnv.New64a()
		h.Write([]byte(q.ID))
		perm := rand.New(rand.NewPCG(h.Sum64(), 0)).Perm(len(q.Options))

		opts := make([]string, len(q.Options))
		newPos := make([]int, len(q.Options)) // old index -> new index
		for i, old := range perm {
			opts[i] = q.Options[old]
			newPos[old] = i
		}
		q.Options = opts

		if q.Type == "single" {
			q.index = newPos[q.index]
			q.Answer = q.index
			continue
		}
		set := make(map[int]bool, len(q.set))
		for old := range q.set {
			set[newPos[old]] = true
		}
		q.set = set
		answer := make([]any, 0, len(set))
		for _, n := range q.CorrectAnswer().([]int) {
			answer = append(answer, n)
		}
		q.Answer = answer
	}
}
