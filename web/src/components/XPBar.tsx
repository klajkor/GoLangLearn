import type { LevelInfo } from '../api'

// XPBar shows progress through the current level.
export default function XPBar({ level, tone = 'light' }: { level: LevelInfo; tone?: 'light' | 'default' }) {
  const span = level.next - level.floor
  const pct = Math.min(100, Math.round(((level.totalXP - level.floor) / span) * 100))
  const light = tone === 'light'
  return (
    <div>
      <div className={`mb-1 flex justify-between text-xs ${light ? 'text-white/80' : 'text-slate-500'}`}>
        <span>
          {level.totalXP - level.floor} / {span} XP
        </span>
        <span>Level {level.level + 1}</span>
      </div>
      <div
        role="progressbar"
        aria-label="Progress to next level"
        aria-valuenow={pct}
        aria-valuemin={0}
        aria-valuemax={100}
        className={`h-2.5 overflow-hidden rounded-full ${light ? 'bg-white/25' : 'bg-slate-200 dark:bg-slate-800'}`}
      >
        <div
          className={`h-full rounded-full transition-all duration-700 ${light ? 'bg-white' : 'bg-gopher-500'}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}
