import { lazy, Suspense, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, ApiError, type Lesson } from '../api'
import Spinner from '../components/Spinner'
import { useAppData } from '../data'
import { useRewards } from '../rewards'
import NotFound from './NotFound'

// Markdown + CodeMirror are only needed on lesson pages; keep them out of the main bundle.
const LessonContent = lazy(() => import('../components/LessonContent'))

const LESSON_XP = 20 // mirrors game.XPLesson on the server

export default function LessonPage() {
  const { chapterId = '', lessonId = '' } = useParams()
  const { curriculum, progress } = useAppData()
  const { celebrate } = useRewards()
  const [completing, setCompleting] = useState(false)
  const [lesson, setLesson] = useState<Lesson | null>(null)
  const [status, setStatus] = useState<'loading' | 'ok' | 'notfound' | 'error'>('loading')

  useEffect(() => {
    let cancelled = false
    setStatus('loading')
    api
      .lesson(chapterId, lessonId)
      .then((l) => {
        if (cancelled) return
        setLesson(l)
        setStatus('ok')
      })
      .catch((e) => !cancelled && setStatus(e instanceof ApiError && e.status === 404 ? 'notfound' : 'error'))
    return () => {
      cancelled = true
    }
  }, [chapterId, lessonId])

  if (status === 'notfound') return <NotFound />
  if (status === 'error') return <p className="text-center text-red-500">Could not load this lesson.</p>
  if (status === 'loading' || !lesson || !curriculum) return <Spinner />

  const chapter = curriculum.chapters.find((c) => c.id === chapterId)!
  const idx = chapter.lessons.findIndex((l) => l.id === lessonId)
  const prev = chapter.lessons[idx - 1]
  const next = chapter.lessons[idx + 1]
  const isDone = progress?.completedLessons[chapterId]?.includes(lessonId) ?? false

  const complete = async () => {
    setCompleting(true)
    try {
      celebrate(await api.completeLesson(chapterId, lessonId))
    } catch {
      // Leave the button enabled so the learner can try again.
    } finally {
      setCompleting(false)
    }
  }

  return (
    <article className="mx-auto max-w-3xl">
      <nav className="mb-6 text-sm text-slate-500" aria-label="Breadcrumb">
        <Link to={`/chapters/${chapter.id}`} className="hover:text-gopher-600 hover:underline">
          {chapter.icon} {chapter.title}
        </Link>
        <span className="mx-2">/</span>
        <span>
          Lesson {idx + 1} of {chapter.lessons.length}
        </span>
      </nav>

      <header className="mb-6">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">{lesson.title}</h1>
        {lesson.minutes > 0 && <p className="mt-2 text-sm text-slate-500">⏱ {lesson.minutes} min read</p>}
      </header>

      <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm sm:p-8 dark:border-slate-800 dark:bg-slate-900">
        <Suspense fallback={<Spinner />}>
          <LessonContent markdown={lesson.body} lesson={{ chapterId, lessonId }} />
        </Suspense>
      </div>

      <div className="mt-6 flex justify-center">
        {isDone ? (
          <p className="rounded-xl bg-emerald-500/10 px-5 py-2.5 text-sm font-semibold text-emerald-700 dark:text-emerald-300">
            ✓ Lesson completed
          </p>
        ) : (
          <button
            type="button"
            disabled={completing}
            onClick={complete}
            className="rounded-xl bg-emerald-700 px-6 py-3 text-sm font-bold text-white shadow-sm transition hover:bg-emerald-800 disabled:opacity-60"
          >
            {completing ? 'Saving…' : `✓ Mark complete  +${LESSON_XP} XP`}
          </button>
        )}
      </div>

      <div className="mt-8 flex flex-wrap justify-between gap-3">
        {prev ? (
          <Link to={`/chapters/${chapter.id}/lessons/${prev.id}`} className={navBtn}>
            ← {prev.title}
          </Link>
        ) : (
          <span />
        )}
        {next ? (
          <Link to={`/chapters/${chapter.id}/lessons/${next.id}`} className={`${navBtn} ${primary}`}>
            {next.title} →
          </Link>
        ) : (
          <Link
            to={chapter.quizQuestions > 0 ? `/chapters/${chapter.id}/quiz` : `/chapters/${chapter.id}`}
            className={`${navBtn} ${primary}`}
          >
            {chapter.quizQuestions > 0 ? "📝 Take the quiz →" : "Back to chapter →"}
          </Link>
        )}
      </div>
    </article>
  )
}

const navBtn =
  'rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-semibold transition hover:bg-slate-100 dark:border-slate-700 dark:hover:bg-slate-800'
const primary = 'border-transparent bg-gopher-500 text-white hover:bg-gopher-600 dark:hover:bg-gopher-600'
