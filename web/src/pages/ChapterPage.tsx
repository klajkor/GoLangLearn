import { Link, useParams } from 'react-router-dom'
import type { QuizStat } from '../api'
import Spinner from '../components/Spinner'
import { useAppData } from '../data'
import { tierStyles } from '../tiers'
import NotFound from './NotFound'

export default function ChapterPage() {
  const { chapterId } = useParams()
  const { curriculum, progress } = useAppData()
  if (!curriculum || !progress) return <Spinner />

  const chapter = curriculum.chapters.find((c) => c.id === chapterId)
  if (!chapter) return <NotFound />

  const done = new Set(progress.completedLessons[chapter.id] ?? [])
  const style = tierStyles[chapter.tier]
  const next = curriculum.chapters.find((c) => c.order === chapter.order + 1)

  return (
    <div className="mx-auto max-w-3xl">
      <header>
        <span className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${style.badge}`}>
          {style.label} · Chapter {chapter.order}
        </span>
        <h1 className="mt-3 flex items-center gap-3 text-3xl font-bold tracking-tight">
          <span aria-hidden>{chapter.icon}</span>
          {chapter.title}
        </h1>
        <p className="mt-2 text-slate-500 dark:text-slate-400">{chapter.summary}</p>
      </header>

      {chapter.lessons.length === 0 ? (
        <div className="mt-8 rounded-2xl border border-dashed border-slate-300 p-10 text-center text-slate-500 dark:border-slate-700">
          🚧 Lessons for this chapter are coming soon.
        </div>
      ) : (
        <ol className="mt-8 flex flex-col gap-3">
          {chapter.lessons.map((l, i) => (
            <li key={l.id}>
              <Link
                to={`/chapters/${chapter.id}/lessons/${l.id}`}
                className="flex items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 transition hover:border-gopher-400 hover:shadow-sm dark:border-slate-800 dark:bg-slate-900 dark:hover:border-gopher-600"
              >
                <span
                  className={`grid size-9 shrink-0 place-items-center rounded-full text-sm font-bold ${
                    done.has(l.id)
                      ? 'bg-emerald-500 text-white'
                      : 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400'
                  }`}
                >
                  {done.has(l.id) ? '✓' : i + 1}
                </span>
                <span className="flex-1 font-medium">{l.title}</span>
                {l.minutes > 0 && <span className="text-xs text-slate-400">{l.minutes} min</span>}
              </Link>
            </li>
          ))}
        </ol>
      )}

      <QuizCard chapterId={chapter.id} questions={chapter.quizQuestions} stat={progress.quizzes[chapter.id]} />

      {next && (
        <Link
          to={`/chapters/${next.id}`}
          className="mt-8 inline-flex items-center gap-2 text-sm font-semibold text-gopher-600 hover:underline dark:text-gopher-300"
        >
          Next chapter: {next.title} →
        </Link>
      )}
    </div>
  )
}

function QuizCard({ chapterId, questions, stat }: { chapterId: string; questions: number; stat?: QuizStat }) {
  if (questions === 0) {
    return (
      <div className="mt-6 flex items-center gap-4 rounded-2xl bg-slate-100 p-4 text-sm text-slate-500 dark:bg-slate-900 dark:text-slate-400">
        <span className="text-2xl" aria-hidden>
          📝
        </span>
        Chapter quiz coming soon.
      </div>
    )
  }
  return (
    <Link
      to={`/chapters/${chapterId}/quiz`}
      className="group mt-6 flex items-center gap-4 rounded-2xl border-2 border-gopher-500/40 bg-gopher-500/5 p-4 transition hover:border-gopher-500 hover:bg-gopher-500/10"
    >
      <span className="text-3xl" aria-hidden>
        📝
      </span>
      <div className="flex-1">
        <div className="font-semibold">Chapter quiz</div>
        <div className="text-sm text-slate-500 dark:text-slate-400">
          {questions} questions
          {stat &&
            ` · best ${stat.bestPercent}% ${stat.passed ? "✓ passed" : "(not passed yet)"} · ${stat.attempts} ${stat.attempts === 1 ? "attempt" : "attempts"}`}
        </div>
      </div>
      <span className="font-semibold text-gopher-600 group-hover:translate-x-0.5 dark:text-gopher-300">
        {stat ? "Retake →" : "Start →"}
      </span>
    </Link>
  )
}
