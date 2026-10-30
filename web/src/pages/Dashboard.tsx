import { Link } from 'react-router-dom'
import Spinner from '../components/Spinner'
import { useAppData } from '../data'
import Heatmap from '../components/Heatmap'
import XPBar from '../components/XPBar'
import { nextStep } from '../lib/nextStep'
import { tierStyles } from '../tiers'
import type { Curriculum, Progress } from '../api'

export default function Dashboard() {
  const { curriculum, progress } = useAppData()
  if (!curriculum || !progress) return <Spinner />

  const totalLessons = curriculum.chapters.reduce((n, c) => n + c.lessons.length, 0)
  const doneLessons = Object.values(progress.completedLessons).reduce((n, l) => n + l.length, 0)

  return (
    <div className="mx-auto max-w-5xl">
      <section className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-gopher-500 to-sky-700 p-6 text-white shadow-lg sm:p-10">
        <div className="absolute -right-6 -top-6 text-[9rem] opacity-20 select-none sm:text-[12rem]" aria-hidden>
          🐹
        </div>
        <p className="text-sm font-medium uppercase tracking-wider text-white/80">Level {progress.level.level}</p>
        <h1 className="mt-1 max-w-xl text-3xl font-bold tracking-tight sm:text-4xl">{progress.level.title}</h1>
        <div className="mt-5 max-w-md">
          <XPBar level={progress.level} />
        </div>
        <div className="mt-6 grid max-w-lg grid-cols-4 gap-2 text-center sm:gap-3">
          <Stat label="XP" value={progress.totalXP} />
          <Stat label="Streak" value={`🔥 ${progress.streak.current}`} />
          <Stat label="Lessons" value={`${doneLessons}/${totalLessons}`} />
          <Stat label="Badges" value={progress.badges.length} />
        </div>
      </section>

      <div className="mt-6 grid gap-4 lg:grid-cols-2">
        <NextUp curriculum={curriculum} progress={progress} />
        <section className={card}>
          <div className="flex items-baseline justify-between">
            <h2 className="font-bold">Streak</h2>
            <span className="text-xs text-slate-500">
              best {progress.streak.longest} · {progress.streak.freezeReady ? '❄️ freeze ready' : 'freeze used this week'}
            </span>
          </div>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
            {progress.streak.activeToday
              ? 'You learned today. Nice!'
              : progress.streak.current > 0
                ? 'Learn something today to keep your streak alive.'
                : 'Complete a lesson today to start a streak.'}
          </p>
          <div className="mt-4">
            <Heatmap days={progress.activityDays} />
          </div>
        </section>
      </div>

      <RecentBadges progress={progress} />

      <div className="mt-4">
        <Link to="/map" className="text-sm font-semibold text-gopher-600 hover:underline dark:text-gopher-300">
          🗺️ Open the skill map →
        </Link>
      </div>

      {curriculum.tiers.map((tier) => {
        const chapters = curriculum.chapters.filter((c) => c.tier === tier)
        if (chapters.length === 0) return null
        return (
          <section key={tier} className="mt-10">
            <h2 className="mb-4 flex items-center gap-3 text-xl font-bold">
              {tierStyles[tier].label}
              <span className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${tierStyles[tier].badge}`}>
                {chapters.length} chapters
              </span>
            </h2>
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {chapters.map((ch) => (
                <Link
                  key={ch.id}
                  to={`/chapters/${ch.id}`}
                  className="group relative overflow-hidden rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md dark:border-slate-800 dark:bg-slate-900"
                >
                  <div className={`absolute inset-x-0 top-0 h-1 bg-gradient-to-r ${tierStyles[tier].accent}`} />
                  <div className="flex items-start gap-3">
                    <span className="text-3xl" aria-hidden>
                      {ch.icon}
                    </span>
                    <div className="min-w-0">
                      <p className="text-xs font-medium text-slate-400">Chapter {ch.order}</p>
                      <h3 className="font-semibold group-hover:text-gopher-600 dark:group-hover:text-gopher-300">
                        {ch.title}
                      </h3>
                    </div>
                  </div>
                  <p className="mt-3 line-clamp-2 text-sm text-slate-500 dark:text-slate-400">{ch.summary}</p>
                  <p className="mt-3 text-xs text-slate-400">
                    {ch.lessons.length
                      ? `${ch.lessons.length} ${ch.lessons.length === 1 ? 'lesson' : 'lessons'}`
                      : 'Coming soon'}
                  </p>
                </Link>
              ))}
            </div>
          </section>
        )
      })}
    </div>
  )
}

function Stat({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="rounded-xl bg-white/15 px-3 py-2 backdrop-blur">
      <div className="text-xl font-bold">{value}</div>
      <div className="text-xs text-white/80">{label}</div>
    </div>
  )
}

const card = 'rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900'

function NextUp({ curriculum, progress }: { curriculum: Curriculum; progress: Progress }) {
  const step = nextStep(curriculum, progress)
  return (
    <section className={card}>
      <h2 className="font-bold">Up next</h2>
      {step ? (
        <Link
          to={step.kind === 'lesson' ? `/chapters/${step.chapter.id}/lessons/${step.lessonId}` : `/chapters/${step.chapter.id}/quiz`}
          className="mt-3 flex items-center gap-4 rounded-xl bg-gopher-500/10 p-4 transition hover:bg-gopher-500/15"
        >
          <span className="text-3xl" aria-hidden>
            {step.kind === 'lesson' ? step.chapter.icon : '📝'}
          </span>
          <span className="min-w-0 flex-1">
            <span className="block text-xs text-slate-500">
              Chapter {step.chapter.order} · {step.chapter.title}
            </span>
            <span className="block font-semibold">{step.kind === 'lesson' ? step.title : 'Take the chapter quiz'}</span>
          </span>
          <span className="font-semibold text-gopher-600 dark:text-gopher-300">{progress.totalXP === 0 ? 'Start' : 'Continue'} →</span>
        </Link>
      ) : (
        <p className="mt-3 text-sm text-slate-500 dark:text-slate-400">
          🎉 You've mastered everything available so far. More chapters are on the way.
        </p>
      )}
    </section>
  )
}

function RecentBadges({ progress }: { progress: Progress }) {
  const byId = new Map(progress.badgeCatalog.map((b) => [b.id, b]))
  const recent = progress.badges.slice(-4).reverse()
  return (
    <section className={`${card} mt-4`}>
      <div className="flex items-baseline justify-between">
        <h2 className="font-bold">Recent badges</h2>
        <Link to="/badges" className="text-sm font-semibold text-gopher-600 hover:underline dark:text-gopher-300">
          All {progress.badges.length}/{progress.badgeCatalog.length} →
        </Link>
      </div>
      {recent.length === 0 ? (
        <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">Complete a lesson to earn your first badge.</p>
      ) : (
        <ul className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {recent.map((id) => {
            const b = byId.get(id)
            return (
              b && (
                <li key={id} className="flex items-center gap-3 rounded-xl bg-slate-100 p-3 dark:bg-slate-800/70">
                  <span className="text-2xl" aria-hidden>
                    {b.icon}
                  </span>
                  <span className="text-sm font-semibold">{b.name}</span>
                </li>
              )
            )
          })}
        </ul>
      )}
    </section>
  )
}
