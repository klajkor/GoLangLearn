import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <div className="py-20 text-center">
      <div className="text-6xl" aria-hidden>
        🕳️
      </div>
      <h1 className="mt-4 text-2xl font-bold">This gopher hole is empty</h1>
      <p className="mt-2 text-slate-500">The page you're looking for doesn't exist.</p>
      <Link to="/" className="mt-6 inline-block font-semibold text-gopher-600 hover:underline dark:text-gopher-300">
        Back to dashboard
      </Link>
    </div>
  )
}
