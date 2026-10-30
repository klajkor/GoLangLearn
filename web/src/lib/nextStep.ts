import type { Chapter, Curriculum, Progress } from '../api'

export type NextStep =
  | { kind: 'lesson'; chapter: Chapter; lessonId: string; title: string }
  | { kind: 'quiz'; chapter: Chapter }

// nextStep recommends what to do next: the first unfinished lesson of the
// earliest chapter that isn't mastered, or its quiz once the lessons are done.
// It returns null when everything available is mastered.
export function nextStep(curriculum: Curriculum, progress: Progress): NextStep | null {
  for (const chapter of curriculum.chapters) {
    const status = progress.chapterStatus[chapter.id]
    if (status === 'soon' || status === 'mastered') continue
    const done = new Set(progress.completedLessons[chapter.id] ?? [])
    const lesson = chapter.lessons.find((l) => !done.has(l.id))
    if (lesson) return { kind: 'lesson', chapter, lessonId: lesson.id, title: lesson.title }
    if (chapter.quizQuestions > 0) return { kind: 'quiz', chapter }
  }
  return null
}
