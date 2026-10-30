package game

// State is the snapshot of learner progress that badge rules look at.
type State struct {
	Level            int
	LessonsDone      int
	RunLessons       int // lessons whose examples have been run successfully
	QuizzesPassed    int // chapters whose quiz is passed
	PerfectQuizzes   int // chapters with a 100% attempt
	RetryPassed      bool
	ChaptersMastered int
	MasteredTiers    map[string]bool // tier -> every chapter in it is mastered
	LongestStreak    int
}

// Badge is an achievement. Earned decides if the state qualifies.
type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`

	earned func(State) bool
}

func atLeast(get func(State) int, n int) func(State) bool {
	return func(s State) bool { return get(s) >= n }
}

// Badges is the full catalogue, in display order.
var Badges = []Badge{
	{ID: "first-steps", Name: "First Steps", Description: "Complete your first lesson.", Icon: "👣",
		earned: atLeast(func(s State) int { return s.LessonsDone }, 1)},
	{ID: "bookworm", Name: "Bookworm", Description: "Complete 5 lessons.", Icon: "📖",
		earned: atLeast(func(s State) int { return s.LessonsDone }, 5)},
	{ID: "scholar", Name: "Scholar", Description: "Complete 15 lessons.", Icon: "🎓",
		earned: atLeast(func(s State) int { return s.LessonsDone }, 15)},
	{ID: "polymath", Name: "Polymath", Description: "Complete 30 lessons.", Icon: "🧠",
		earned: atLeast(func(s State) int { return s.LessonsDone }, 30)},

	{ID: "hello-gopher", Name: "Hello, Gopher", Description: "Run an example from a lesson.", Icon: "▶️",
		earned: atLeast(func(s State) int { return s.RunLessons }, 1)},
	{ID: "tinkerer", Name: "Tinkerer", Description: "Run the examples in 5 different lessons.", Icon: "🔧",
		earned: atLeast(func(s State) int { return s.RunLessons }, 5)},
	{ID: "code-machine", Name: "Code Machine", Description: "Run the examples in 15 different lessons.", Icon: "⚙️",
		earned: atLeast(func(s State) int { return s.RunLessons }, 15)},

	{ID: "quiz-taker", Name: "Quiz Taker", Description: "Pass your first chapter quiz.", Icon: "📝",
		earned: atLeast(func(s State) int { return s.QuizzesPassed }, 1)},
	{ID: "perfectionist", Name: "Perfectionist", Description: "Score 100% on a chapter quiz.", Icon: "💯",
		earned: atLeast(func(s State) int { return s.PerfectQuizzes }, 1)},
	{ID: "hat-trick", Name: "Hat Trick", Description: "Score 100% on three chapter quizzes.", Icon: "🎩",
		earned: atLeast(func(s State) int { return s.PerfectQuizzes }, 3)},
	{ID: "comeback-kid", Name: "Comeback Kid", Description: "Pass a quiz after failing it.", Icon: "💪",
		earned: func(s State) bool { return s.RetryPassed }},
	{ID: "quiz-master", Name: "Quiz Master", Description: "Pass 5 chapter quizzes.", Icon: "🧩",
		earned: atLeast(func(s State) int { return s.QuizzesPassed }, 5)},

	{ID: "chapter-champion", Name: "Chapter Champion", Description: "Master a chapter: every lesson done and its quiz passed.", Icon: "🏅",
		earned: atLeast(func(s State) int { return s.ChaptersMastered }, 1)},
	{ID: "trailblazer", Name: "Trailblazer", Description: "Master 3 chapters.", Icon: "🧭",
		earned: atLeast(func(s State) int { return s.ChaptersMastered }, 3)},

	{ID: "beginner-graduate", Name: "Beginner Graduate", Description: "Master every Beginner chapter.", Icon: "🌱",
		earned: func(s State) bool { return s.MasteredTiers["beginner"] }},
	{ID: "intermediate-graduate", Name: "Intermediate Graduate", Description: "Master every Intermediate chapter.", Icon: "🌿",
		earned: func(s State) bool { return s.MasteredTiers["intermediate"] }},
	{ID: "advanced-graduate", Name: "Advanced Graduate", Description: "Master every Advanced chapter.", Icon: "🌳",
		earned: func(s State) bool { return s.MasteredTiers["advanced"] }},
	{ID: "go-guru", Name: "Go Guru", Description: "Master every Expert chapter. The whole course is behind you.", Icon: "👑",
		earned: func(s State) bool { return s.MasteredTiers["expert"] }},

	{ID: "on-a-roll", Name: "On a Roll", Description: "Learn 3 days in a row.", Icon: "🔥",
		earned: atLeast(func(s State) int { return s.LongestStreak }, 3)},
	{ID: "week-warrior", Name: "Week Warrior", Description: "Learn 7 days in a row.", Icon: "🗓️",
		earned: atLeast(func(s State) int { return s.LongestStreak }, 7)},
	{ID: "unstoppable", Name: "Unstoppable", Description: "Learn 30 days in a row.", Icon: "🚀",
		earned: atLeast(func(s State) int { return s.LongestStreak }, 30)},

	{ID: "level-5", Name: "Goroutine Wrangler", Description: "Reach level 5.", Icon: "🧵",
		earned: atLeast(func(s State) int { return s.Level }, 5)},
	{ID: "level-10", Name: "Level 10", Description: "Reach level 10.", Icon: "⭐",
		earned: atLeast(func(s State) int { return s.Level }, 10)},
}

// BadgeByID returns the catalogue entry with the given ID.
func BadgeByID(id string) (Badge, bool) {
	for _, b := range Badges {
		if b.ID == id {
			return b, true
		}
	}
	return Badge{}, false
}

// NewlyEarned returns the badges s qualifies for that aren't in have.
func NewlyEarned(s State, have map[string]bool) []Badge {
	var out []Badge
	for _, b := range Badges {
		if !have[b.ID] && b.earned(s) {
			out = append(out, b)
		}
	}
	return out
}
