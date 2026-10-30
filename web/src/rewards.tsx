import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import type { LevelInfo, Reward } from './api'
import { useAppData } from './data'

interface Toast {
  id: number
  icon: string
  title: string
  detail?: string
}

interface Rewards {
  // celebrate shows toasts / the level-up screen for what an action earned
  // and refreshes progress. It accepts undefined so callers needn't check.
  celebrate: (reward: Reward | undefined) => void
}

const RewardsContext = createContext<Rewards | null>(null)

const TOAST_MS = 4500

export function RewardsProvider({ children }: { children: ReactNode }) {
  const { reload } = useAppData()
  const [toasts, setToasts] = useState<Toast[]>([])
  const [levelUp, setLevelUp] = useState<LevelInfo | null>(null)
  const nextId = useRef(0)

  const push = useCallback((t: Omit<Toast, 'id'>) => {
    const id = nextId.current++
    setToasts((all) => [...all, { ...t, id }])
    setTimeout(() => setToasts((all) => all.filter((x) => x.id !== id)), TOAST_MS)
  }, [])

  const celebrate = useCallback(
    (reward: Reward | undefined) => {
      if (!reward) return
      reload()
      if (reward.xpGained > 0) push({ icon: '⚡', title: `+${reward.xpGained} XP` })
      for (const b of reward.newBadges) push({ icon: b.icon, title: `Badge unlocked: ${b.name}`, detail: b.description })
      if (reward.leveledUp) setLevelUp(reward.level)
    },
    [push, reload],
  )

  const value = useMemo(() => ({ celebrate }), [celebrate])

  return (
    <RewardsContext.Provider value={value}>
      {children}
      <div
        aria-live="polite"
        className="pointer-events-none fixed inset-x-0 bottom-4 z-50 flex flex-col items-center gap-2 px-4 sm:items-end sm:px-6"
      >
        {toasts.map((t) => (
          <div
            key={t.id}
            className="toast-in pointer-events-auto flex max-w-sm items-center gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 shadow-lg dark:border-slate-700 dark:bg-slate-800"
          >
            <span className="text-2xl" aria-hidden>
              {t.icon}
            </span>
            <div>
              <div className="text-sm font-bold">{t.title}</div>
              {t.detail && <div className="text-xs text-slate-500 dark:text-slate-400">{t.detail}</div>}
            </div>
          </div>
        ))}
      </div>
      {levelUp && <LevelUpDialog level={levelUp} onClose={() => setLevelUp(null)} />}
    </RewardsContext.Provider>
  )
}

export function useRewards(): Rewards {
  const ctx = useContext(RewardsContext)
  if (!ctx) throw new Error('useRewards must be used inside RewardsProvider')
  return ctx
}

const confettiColors = ['#00add8', '#29c1e2', '#f59e0b', '#10b981', '#8b5cf6', '#ef4444']

function LevelUpDialog({ level, onClose }: { level: LevelInfo; onClose: () => void }) {
  const buttonRef = useRef<HTMLButtonElement>(null)
  // Pieces are generated once so the animation isn't restarted by re-renders.
  const pieces = useMemo(
    () =>
      Array.from({ length: 40 }, (_, i) => ({
        left: Math.random() * 100,
        delay: Math.random() * 0.8,
        duration: 2.2 + Math.random() * 1.6,
        color: confettiColors[i % confettiColors.length],
        rotate: Math.random() * 360,
      })),
    [],
  )

  useEffect(() => {
    buttonRef.current?.focus()
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  return (
    <div className="fixed inset-0 z-[60] grid place-items-center p-4" role="dialog" aria-modal="true" aria-label="Level up">
      <div className="absolute inset-0 bg-slate-950/60 backdrop-blur-sm" onClick={onClose} />
      <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden>
        {pieces.map((p, i) => (
          <span
            key={i}
            className="confetti absolute -top-4 block h-3 w-2 rounded-sm"
            style={{
              left: `${p.left}%`,
              backgroundColor: p.color,
              animationDelay: `${p.delay}s`,
              animationDuration: `${p.duration}s`,
              transform: `rotate(${p.rotate}deg)`,
            }}
          />
        ))}
      </div>
      <div className="pop-in relative w-full max-w-sm rounded-3xl bg-white p-8 text-center shadow-2xl dark:bg-slate-900">
        <div className="text-6xl" aria-hidden>
          🎉
        </div>
        <p className="mt-3 text-sm font-semibold uppercase tracking-wider text-gopher-600 dark:text-gopher-300">Level up!</p>
        <p className="mt-1 text-6xl font-extrabold tracking-tight">{level.level}</p>
        <p className="mt-2 text-xl font-bold">{level.title}</p>
        <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          {level.next - level.totalXP} XP to level {level.level + 1}
        </p>
        <button
          ref={buttonRef}
          type="button"
          onClick={onClose}
          className="mt-6 rounded-xl bg-gopher-500 px-6 py-2.5 text-sm font-semibold text-white transition hover:bg-gopher-600"
        >
          Keep going
        </button>
      </div>
    </div>
  )
}
