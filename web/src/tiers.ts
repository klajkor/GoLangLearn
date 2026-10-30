import type { Tier } from './api'

interface TierStyle {
  label: string
  badge: string // pill classes
  accent: string // gradient classes for cards/headers
}

export const tierStyles: Record<Tier, TierStyle> = {
  beginner: {
    label: 'Beginner',
    badge: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300',
    accent: 'from-emerald-400 to-teal-500',
  },
  intermediate: {
    label: 'Intermediate',
    badge: 'bg-sky-100 text-sky-800 dark:bg-sky-500/15 dark:text-sky-300',
    accent: 'from-sky-400 to-gopher-600',
  },
  advanced: {
    label: 'Advanced',
    badge: 'bg-violet-100 text-violet-800 dark:bg-violet-500/15 dark:text-violet-300',
    accent: 'from-violet-400 to-fuchsia-500',
  },
  expert: {
    label: 'Expert',
    badge: 'bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300',
    accent: 'from-amber-400 to-orange-500',
  },
}
