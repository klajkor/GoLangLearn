// Types and fetch helpers mirroring the Go API in internal/api.

export type Tier = 'beginner' | 'intermediate' | 'advanced' | 'expert'

export interface LessonMeta {
  id: string
  title: string
  minutes: number
}

export interface Lesson extends LessonMeta {
  body: string
}

export interface Chapter {
  id: string
  order: number
  title: string
  tier: Tier
  summary: string
  icon: string
  lessons: LessonMeta[]
  quizQuestions: number
}

export interface Curriculum {
  tiers: Tier[]
  chapters: Chapter[]
}

export interface QuizStat {
  attempts: number
  bestPercent: number
  passed: boolean
}

export interface LevelInfo {
  level: number
  title: string
  totalXP: number
  floor: number // XP at which this level started
  next: number // XP at which the next level starts
}

export interface Streak {
  current: number
  longest: number
  activeToday: boolean
  freezeReady: boolean
}

export interface Badge {
  id: string
  name: string
  description: string
  icon: string
}

// What an action earned; shown as toasts and a level-up celebration.
export interface Reward {
  xpGained: number
  totalXP: number
  level: LevelInfo
  leveledUp: boolean
  newBadges: Badge[]
}

export type ChapterStatus = 'soon' | 'locked' | 'available' | 'in-progress' | 'mastered'

export interface Progress {
  totalXP: number
  completedLessons: Record<string, string[]>
  quizzes: Record<string, QuizStat>
  badges: string[] // earned badge IDs, oldest first
  level: LevelInfo
  streak: Streak
  activityDays: string[] // YYYY-MM-DD, local time
  chapterStatus: Record<string, ChapterStatus>
  badgeCatalog: Badge[]
}

export interface LessonRef {
  chapterId: string
  lessonId: string
}

export type QuestionType = 'single' | 'multi' | 'output' | 'fill' | 'bug'

// Wire format of an answer: option index (single), line number (bug),
// option indexes (multi), text (output) or one text per gap (fill).
export type Answer = number | number[] | string | string[]

export interface QuizQuestion {
  id: string
  type: QuestionType
  prompt: string
  code?: string
  options?: string[]
  blanks?: number
}

export interface Quiz {
  chapterId: string
  passPercent: number
  questions: QuizQuestion[]
}

export interface QuestionResult {
  id: string
  correct: boolean
  correctAnswer: Answer
  explanation: string
}

export interface QuizResult {
  score: number
  total: number
  percent: number
  passPercent: number
  passed: boolean
  results: QuestionResult[]
  reward: Reward
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
  }
}

export interface RunEvent {
  message: string
  kind: 'stdout' | 'stderr'
  delayMs: number
}

export interface RunResult {
  events: RunEvent[]
  compileErrors?: string
  vetErrors?: string
  exitCode: number
  reward?: Reward // set when a lesson example earned XP
}

async function getJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init)
  const body = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new ApiError(res.status, body.error ?? res.statusText)
  }
  return body as T
}

export const api = {
  curriculum: () => getJSON<Curriculum>('/api/curriculum'),
  lesson: (chapterId: string, lessonId: string) =>
    getJSON<Lesson>(`/api/chapters/${encodeURIComponent(chapterId)}/lessons/${encodeURIComponent(lessonId)}`),
  progress: () => getJSON<Progress>('/api/progress'),
  quiz: (chapterId: string) => getJSON<Quiz>(`/api/chapters/${encodeURIComponent(chapterId)}/quiz`),
  submitQuiz: (chapterId: string, answers: Record<string, Answer>) =>
    getJSON<QuizResult>(`/api/chapters/${encodeURIComponent(chapterId)}/quiz`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ answers }),
    }),
  completeLesson: (chapterId: string, lessonId: string) =>
    getJSON<Reward>(
      `/api/chapters/${encodeURIComponent(chapterId)}/lessons/${encodeURIComponent(lessonId)}/complete`,
      { method: 'POST' },
    ),
  // lesson identifies the lesson an example belongs to, so a run can earn XP.
  run: (code: string, signal?: AbortSignal, lesson?: LessonRef) =>
    getJSON<RunResult>('/api/run', postCode(code, signal, lesson)),
  format: (code: string) => getJSON<{ code: string }>('/api/format', postCode(code)),
}

function postCode(code: string, signal?: AbortSignal, lesson?: LessonRef): RequestInit {
  return {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code, chapter: lesson?.chapterId, lesson: lesson?.lessonId }),
    signal,
  }
}
