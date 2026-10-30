import { Link } from 'react-router-dom'
import type { Chapter, ChapterStatus, Progress } from '../api'
import Spinner from '../components/Spinner'
import { useAppData } from '../data'
import { tierStyles } from '../tiers'

const nodeStyle: Record<ChapterStatus, string> = {
  mastered: 'border-emerald-500 bg-emerald-500 text-white shadow-emerald-500/30',
  'in-progress': 'border-gopher-500 bg-gopher-500 text-white shadow-gopher-500/40 ring-4 ring-gopher-500/25',
  available: 'border-gopher-500 bg-white text-slate-900 dark:bg-slate-900 dark:text-white',
  locked: 'border-slate-300 bg-slate-200 text-slate-500 grayscale dark:border-slate-700 dark:bg-slate-800',
  soon: 'border-dashed border-slate-300 bg-transparent text-slate-400 dark:border-slate-700',
}

const statusLabel: Record<ChapterStatus, string> = {
  mastered: 'Mastered',
  'in-progress': 'In progress',
  available: 'Ready to start',
  locked: 'Locked',
  soon: 'Coming soon',
}

export default function MapPage() {
  const { curriculum, progress } = useAppData()
  if (!curriculum || !progress) return <Spinner />

  return (
    <div className="mx-auto max-w-3xl">
      <h1 className="text-3xl font-bold tracking-tight">🗺️ Skill map</h1>
      <p className="mt-2 text-slate-500 dark:text-slate-400">
        Master a chapter (finish every lesson and pass its quiz) to unlock the next one. Locks are a guide: you can still
        open any chapter.
      </p>
      <Legend />

      {curriculum.tiers.map((tier) => {
        const chapters = curriculum.chapters.filter((c) => c.tier === tier)
        if (chapters.length === 0) return null
        const mastered = chapters.filter((c) => progress.chapterStatus[c.id] === 'mastered').length
        return (
          <section key={tier} className="mt-10">
            <h2 className="flex items-center gap-3 text-xl font-bold">
              {tierStyles[tier].label}
              <span className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${tierStyles[tier].badge}`}>
                {mastered}/{chapters.length} mastered
              </span>
            </h2>
            <ol className="relative mt-6">
              {/* the trail the nodes sit on */}
              <span
                aria-hidden
                className={`absolute bottom-6 left-8 top-6 w-1 rounded-full bg-gradient-to-b sm:left-1/2 sm:-translate-x-1/2 ${tierStyles[tier].accent} opacity-30`}
              />
              {chapters.map((ch, i) => (
                <Node key={ch.id} chapter={ch} status={progress.chapterStatus[ch.id]} progress={progress} flip={i % 2 === 1} />
              ))}
            </ol>
          </section>
        )
      })}
    </div>
  )
}

function Legend() {
  const items: ChapterStatus[] = ['mastered', 'in-progress', 'available', 'locked', 'soon']
  return (
    <ul className="mt-4 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500 dark:text-slate-400">
      {items.map((s) => (
        <li key={s} className="flex items-center gap-1.5">
          <span className={`size-3 rounded-full border-2 ${nodeStyle[s].split(' ').slice(0, 2).join(' ')}`} aria-hidden />
          {statusLabel[s]}
        </li>
      ))}
    </ul>
  )
}

function Node({
  chapter,
  status,
  progress,
  flip,
}: {
  chapter: Chapter
  status: ChapterStatus
  progress: Progress
  flip: boolean
}) {
  const doneCount = progress.completedLessons[chapter.id]?.length ?? 0
  const quiz = progress.quizzes[chapter.id]
  const detail =
    status === 'soon'
      ? 'Coming soon'
      : [
          `${doneCount}/${chapter.lessons.length} lessons`,
          chapter.quizQuestions > 0 ? (quiz?.passed ? `quiz ${quiz.bestPercent}% ✓` : 'quiz not passed') : null,
        ]
          .filter(Boolean)
          .join(' · ')

  return (
    <li className="relative flex items-center gap-4 py-3 pl-0 sm:grid sm:grid-cols-2 sm:gap-0">
      <Link
        to={`/chapters/${chapter.id}`}
        className={`group relative z-10 flex w-full items-center gap-4 sm:w-auto ${flip ? 'sm:col-start-2 sm:flex-row sm:pl-10' : 'sm:col-start-1 sm:flex-row-reverse sm:pr-10 sm:text-right'}`}
      >
        <span
          className={`relative grid size-16 shrink-0 place-items-center rounded-full border-4 text-3xl shadow-lg transition group-hover:scale-105 ${nodeStyle[status]}`}
          aria-hidden
        >
          {chapter.icon}
          {status === 'mastered' && (
            <span className="absolute -right-1 -top-1 grid size-6 place-items-center rounded-full bg-white text-xs font-bold text-emerald-600 shadow">
              ✓
            </span>
          )}
          {status === 'locked' && (
            <span className="absolute -right-1 -top-1 grid size-6 place-items-center rounded-full bg-white text-xs shadow dark:bg-slate-700">
              🔒
            </span>
          )}
        </span>
        <span className="min-w-0 flex-1 sm:flex-none">
          <span className="block text-xs font-medium text-slate-400">Chapter {chapter.order}</span>
          <span className={`block font-semibold group-hover:text-gopher-600 dark:group-hover:text-gopher-300 ${status === 'locked' || status === 'soon' ? 'text-slate-500' : ''}`}>
            {chapter.title}
          </span>
          <span className="block text-xs text-slate-500 dark:text-slate-400">{detail}</span>
          <span className="sr-only">{statusLabel[status]}</span>
        </span>
      </Link>
    </li>
  )
}
