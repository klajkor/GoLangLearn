import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { api, type Curriculum, type Progress } from './api'

interface AppData {
  curriculum: Curriculum | null
  progress: Progress | null
  error: string | null
  reload: () => void
}

const AppDataContext = createContext<AppData | null>(null)

// AppDataProvider loads the curriculum and progress once and shares them.
export function AppDataProvider({ children }: { children: ReactNode }) {
  const [curriculum, setCurriculum] = useState<Curriculum | null>(null)
  const [progress, setProgress] = useState<Progress | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)

  useEffect(() => {
    let cancelled = false
    setError(null)
    Promise.all([api.curriculum(), api.progress()])
      .then(([c, p]) => {
        if (cancelled) return
        setCurriculum(c)
        setProgress(p)
      })
      .catch((e: Error) => !cancelled && setError(e.message))
    return () => {
      cancelled = true
    }
  }, [version])

  return (
    <AppDataContext.Provider value={{ curriculum, progress, error, reload: () => setVersion((v) => v + 1) }}>
      {children}
    </AppDataContext.Provider>
  )
}

export function useAppData(): AppData {
  const ctx = useContext(AppDataContext)
  if (!ctx) throw new Error('useAppData must be used inside AppDataProvider')
  return ctx
}
