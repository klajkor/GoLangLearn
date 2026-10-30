import { NavLink } from 'react-router-dom'
import { useAppData } from '../data'
import { tierStyles } from '../tiers'

export default function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const { curriculum, progress } = useAppData()

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition ${
      isActive
        ? 'bg-gopher-500/10 font-semibold text-gopher-700 dark:text-gopher-300'
        : 'text-slate-600 hover:bg-slate-200/60 dark:text-slate-400 dark:hover:bg-slate-800/70'
    }`

  return (
    <nav aria-label="Curriculum" className="flex flex-col gap-6 p-4">
      <NavLink to="/" end className={linkClass} onClick={onNavigate}>
        <span aria-hidden>🏠</span> Dashboard
      </NavLink>
      <div className="-mt-4 flex flex-col gap-0.5">
        <NavLink to="/map" className={linkClass} onClick={onNavigate}>
          <span aria-hidden>🗺️</span> Skill map
        </NavLink>
        <NavLink to="/badges" className={linkClass} onClick={onNavigate}>
          <span aria-hidden>🏅</span> Badges
          {progress && (
            <span className="ml-auto text-xs text-slate-400">
              {progress.badges.length}/{progress.badgeCatalog.length}
            </span>
          )}
        </NavLink>
      </div>

      {curriculum?.tiers.map((tier) => {
        const chapters = curriculum.chapters.filter((c) => c.tier === tier)
        if (chapters.length === 0) return null
        return (
          <section key={tier}>
            <h2 className="mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500">
              {tierStyles[tier].label}
            </h2>
            <ul className="flex flex-col gap-0.5">
              {chapters.map((ch) => {
                const status = progress?.chapterStatus[ch.id]
                const complete = status === 'mastered'
                return (
                  <li key={ch.id}>
                    <NavLink to={`/chapters/${ch.id}`} className={linkClass} onClick={onNavigate}>
                      <span aria-hidden className="w-5 text-center">
                        {ch.icon}
                      </span>
                      <span className="flex-1 truncate">
                        {ch.order}. {ch.title}
                      </span>
                      {complete && <span aria-label="mastered">✅</span>}
                      {status === 'locked' && <span aria-label="locked">🔒</span>}
                    </NavLink>
                  </li>
                )
              })}
            </ul>
          </section>
        )
      })}
    </nav>
  )
}
