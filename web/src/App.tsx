import { lazy } from 'react'
import { Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import ChapterPage from './pages/ChapterPage'
import NotFound from './pages/NotFound'

// Pages that pull in heavy dependencies (CodeMirror highlighting, Markdown)
// load on demand, keeping the first paint small. Layout supplies the Suspense
// boundary so the header and sidebar stay put while a page loads.
const LessonPage = lazy(() => import('./pages/LessonPage'))
const QuizPage = lazy(() => import('./pages/QuizPage'))
const MapPage = lazy(() => import('./pages/MapPage'))
const BadgesPage = lazy(() => import('./pages/BadgesPage'))

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="chapters/:chapterId" element={<ChapterPage />} />
        <Route path="chapters/:chapterId/lessons/:lessonId" element={<LessonPage />} />
        <Route path="chapters/:chapterId/quiz" element={<QuizPage />} />
        <Route path="map" element={<MapPage />} />
        <Route path="badges" element={<BadgesPage />} />
        <Route path="*" element={<NotFound />} />
      </Route>
    </Routes>
  )
}

