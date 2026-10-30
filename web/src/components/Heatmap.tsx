const WEEKS = 18

// localDay formats a date as YYYY-MM-DD in local time, matching the server.
function localDay(d: Date): string {
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

// Heatmap shows the last few weeks as a calendar grid, one column per week
// (Monday first), with active days filled in.
export default function Heatmap({ days }: { days: string[] }) {
  const active = new Set(days)
  const today = new Date()
  today.setHours(12, 0, 0, 0)
  const sinceMonday = (today.getDay() + 6) % 7
  const start = new Date(today)
  start.setDate(today.getDate() - sinceMonday - (WEEKS - 1) * 7)

  const cells = Array.from({ length: WEEKS * 7 }, (_, i) => {
    const d = new Date(start)
    d.setDate(start.getDate() + i)
    return { key: localDay(d), future: d > today }
  })

  return (
    <div className="overflow-x-auto" role="img" aria-label={`${active.size} active days in the last ${WEEKS} weeks`}>
      <div className="grid w-max grid-flow-col grid-rows-7 gap-1">
        {cells.map(({ key, future }) => (
          <span
            key={key}
            title={key}
            className={`size-3.5 rounded-[4px] ${
              future
                ? 'invisible'
                : active.has(key)
                  ? 'bg-gopher-500'
                  : 'bg-slate-200 dark:bg-slate-800'
            } ${key === localDay(today) ? 'ring-2 ring-gopher-500/50' : ''}`}
          />
        ))}
      </div>
    </div>
  )
}
