import { Suspense, useEffect, useState } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import { useAppData } from '../data'
import Sidebar from './Sidebar'
import Spinner from './Spinner'
import ThemeToggle from './ThemeToggle'

export default function Layout() {
  const { progress, error, reload } = useAppData()
  const [menuOpen, setMenuOpen] = useState(false)
  const location = useLocation()

  // Close the mobile drawer on navigation and scroll to top.
  useEffect(() => {
    setMenuOpen(false)
    window.scrollTo(0, 0)
  }, [location.pathname])

  return (
    <div className="min-h-dvh">
      <header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b border-slate-200 bg-white/80 px-4 backdrop-blur dark:border-slate-800 dark:bg-slate-950/80">
        <button
          type="button"
          className="grid size-9 place-items-center rounded-lg text-xl hover:bg-slate-200/70 lg:hidden dark:hover:bg-slate-800"
          aria-label="Open menu"
          aria-expanded={menuOpen}
          onClick={() => setMenuOpen(true)}
        >
          ☰
        </button>
        <Link to="/" className="flex items-center gap-2 font-bold tracking-tight">
          <span className="text-2xl" aria-hidden>
            🐹
          </span>
          <span className="text-lg">
            Go<span className="text-gopher-500 dark:text-gopher-300">Learn</span>
          </span>
        </Link>
        <div className="ml-auto flex items-center gap-2">
          {progress && progress.streak.current > 0 && (
            <span
              className="rounded-full bg-orange-500/10 px-2.5 py-1 text-sm font-semibold text-orange-700 dark:text-orange-300"
              title={`${progress.streak.current}-day streak`}
            >
              🔥 {progress.streak.current}
            </span>
          )}
          <Link
            to="/"
            className="rounded-full bg-gopher-500/10 px-3 py-1 text-sm font-semibold text-gopher-700 dark:text-gopher-300"
            title={progress ? `Level ${progress.level.level}: ${progress.level.title}` : 'Experience points'}
          >
            ⚡ {progress?.totalXP ?? 0}
            <span className="hidden sm:inline"> XP · Lv {progress?.level.level ?? 1}</span>
          </Link>
          <ThemeToggle />
        </div>
      </header>

      <div className="mx-auto flex max-w-7xl">
        {/* Desktop sidebar */}
        <aside className="sticky top-14 hidden h-[calc(100dvh-3.5rem)] w-72 shrink-0 overflow-y-auto border-r border-slate-200 lg:block dark:border-slate-800">
          <Sidebar />
        </aside>

        {/* Mobile drawer */}
        {menuOpen && (
          <div className="fixed inset-0 z-40 lg:hidden" role="dialog" aria-modal="true" aria-label="Menu">
            <div className="absolute inset-0 bg-slate-950/50 backdrop-blur-sm" onClick={() => setMenuOpen(false)} />
            <aside className="absolute inset-y-0 left-0 w-72 max-w-[85vw] overflow-y-auto bg-white shadow-xl dark:bg-slate-900">
              <div className="flex h-14 items-center justify-between border-b border-slate-200 px-4 dark:border-slate-800">
                <span className="font-bold">Chapters</span>
                <button
                  type="button"
                  className="grid size-9 place-items-center rounded-lg hover:bg-slate-200/70 dark:hover:bg-slate-800"
                  aria-label="Close menu"
                  onClick={() => setMenuOpen(false)}
                >
                  ✕
                </button>
              </div>
              <Sidebar onNavigate={() => setMenuOpen(false)} />
            </aside>
          </div>
        )}

        <main className="min-w-0 flex-1 px-4 py-6 sm:px-6 lg:px-10 lg:py-10">
          {error ? (
            <div className="mx-auto max-w-lg rounded-2xl border border-red-200 bg-red-50 p-6 text-center dark:border-red-900 dark:bg-red-950/40">
              <p className="font-semibold text-red-700 dark:text-red-300">Couldn't reach the GoLearn server.</p>
              <p className="mt-1 text-sm text-red-600/80 dark:text-red-400/80">{error}</p>
              <button
                type="button"
                onClick={reload}
                className="mt-4 rounded-lg bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-700"
              >
                Retry
              </button>
            </div>
          ) : (
            <Suspense fallback={<Spinner />}>
              <Outlet />
            </Suspense>
          )}
        </main>
      </div>
    </div>
  )
}
