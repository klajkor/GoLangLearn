import Spinner from '../components/Spinner'
import { useAppData } from '../data'

export default function BadgesPage() {
  const { progress } = useAppData()
  if (!progress) return <Spinner />

  const earned = new Set(progress.badges)

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="text-3xl font-bold tracking-tight">🏅 Badges</h1>
      <p className="mt-2 text-slate-500 dark:text-slate-400">
        {earned.size} of {progress.badgeCatalog.length} earned. Locked badges show what to aim for.
      </p>

      <ul className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {progress.badgeCatalog.map((b) => {
          const has = earned.has(b.id)
          return (
            <li
              key={b.id}
              className={`flex items-start gap-4 rounded-2xl border p-4 ${
                has
                  ? 'border-gopher-500/40 bg-gopher-500/5'
                  : 'border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900'
              }`}
            >
              <span
                className={`grid size-12 shrink-0 place-items-center rounded-xl text-2xl ${
                  has ? 'bg-gopher-500/15' : 'bg-slate-100 grayscale dark:bg-slate-800'
                }`}
                aria-hidden
              >
                {has ? b.icon : '🔒'}
              </span>
              <div>
                <h2 className="font-semibold">{b.name}</h2>
                <p className="mt-0.5 text-sm text-slate-500 dark:text-slate-400">{b.description}</p>
                <p className="sr-only">{has ? 'Earned' : 'Not earned yet'}</p>
              </div>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
