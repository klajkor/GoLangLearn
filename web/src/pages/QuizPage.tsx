import { useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, ApiError, type Answer, type Quiz, type QuizResult } from '../api'
import Inline from '../components/Inline'
import QuestionView, { isAnswered, typeLabels } from '../components/QuestionView'
import Spinner from '../components/Spinner'
import { useAppData } from '../data'
import { useRewards } from '../rewards'
import NotFound from './NotFound'

type Load = { status: 'loading' } | { status: 'notfound' } | { status: 'error' } | { status: 'ready'; quiz: Quiz }

export default function QuizPage() {
  const { chapterId = '' } = useParams()
  const { curriculum } = useAppData()
  const { celebrate } = useRewards()
  const [load, setLoad] = useState<Load>({ status: 'loading' })
  const [answers, setAnswers] = useState<Record<string, Answer>>({})
  const [step, setStep] = useState(0)
  const [result, setResult] = useState<QuizResult | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)
  const topRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    let cancelled = false
    setLoad({ status: 'loading' })
    setAnswers({})
    setStep(0)
    setResult(null)
    api
      .quiz(chapterId)
      .then((quiz) => !cancelled && setLoad({ status: 'ready', quiz }))
      .catch((e) => !cancelled && setLoad({ status: e instanceof ApiError && e.status === 404 ? 'notfound' : 'error' }))
    return () => {
      cancelled = true
    }
  }, [chapterId])

  const chapter = curriculum?.chapters.find((c) => c.id === chapterId)
  if (load.status === 'notfound' || (curriculum && !chapter)) return <NotFound />
  if (load.status === 'error') return <p className="text-center text-red-500">Could not load this quiz.</p>
  if (load.status === 'loading' || !chapter || !curriculum) return <Spinner />

  const { quiz } = load
  const next = curriculum.chapters.find((c) => c.order === chapter.order + 1)

  const retry = () => {
    setAnswers({})
    setStep(0)
    setResult(null)
    setSubmitError(null)
    topRef.current?.scrollIntoView({ block: 'start' })
  }

  const goTo = (i: number) => {
    setStep(i)
    topRef.current?.scrollIntoView({ block: 'start', behavior: 'smooth' })
  }

  const submit = async () => {
    setSubmitting(true)
    setSubmitError(null)
    try {
      const res = await api.submitQuiz(chapter.id, answers)
      setResult(res)
      celebrate(res.reward) // also refreshes the best score shown on the chapter page
      topRef.current?.scrollIntoView({ block: 'start' })
    } catch (e) {
      setSubmitError((e as Error).message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div ref={topRef} className="mx-auto max-w-3xl scroll-mt-20">
      <nav className="mb-6 text-sm text-slate-500" aria-label="Breadcrumb">
        <Link to={`/chapters/${chapter.id}`} className="hover:text-gopher-600 hover:underline">
          {chapter.icon} {chapter.title}
        </Link>
        <span className="mx-2">/</span>
        <span>Quiz</span>
      </nav>

      {result ? (
        <Results
          quiz={quiz}
          result={result}
          answers={answers}
          chapterPath={`/chapters/${chapter.id}`}
          nextChapter={next && { path: `/chapters/${next.id}`, title: next.title }}
          onRetry={retry}
        />
      ) : (
        <Taking
          quiz={quiz}
          answers={answers}
          step={step}
          submitting={submitting}
          submitError={submitError}
          onAnswer={(id, v) => setAnswers((a) => ({ ...a, [id]: v }))}
          onStep={goTo}
          onSubmit={submit}
        />
      )}
    </div>
  )
}

function Taking({
  quiz,
  answers,
  step,
  submitting,
  submitError,
  onAnswer,
  onStep,
  onSubmit,
}: {
  quiz: Quiz
  answers: Record<string, Answer>
  step: number
  submitting: boolean
  submitError: string | null
  onAnswer: (id: string, v: Answer) => void
  onStep: (i: number) => void
  onSubmit: () => void
}) {
  const q = quiz.questions[step]
  const last = step === quiz.questions.length - 1
  const done = quiz.questions.map((x) => isAnswered(x, answers[x.id]))
  const remaining = done.filter((d) => !d).length

  return (
    <>
      <div className="mb-6">
        <div className="mb-2 flex items-center justify-between text-sm">
          <span className="font-semibold">
            Question {step + 1} of {quiz.questions.length}
          </span>
          <span className="rounded-full bg-gopher-500/10 px-2.5 py-0.5 text-xs font-semibold text-gopher-700 dark:text-gopher-300">
            {typeLabels[q.type]}
          </span>
        </div>
        <ol className="flex gap-1.5" aria-label="Questions">
          {quiz.questions.map((x, i) => (
            <li key={x.id} className="flex-1">
              <button
                type="button"
                onClick={() => onStep(i)}
                aria-label={`Question ${i + 1}${done[i] ? ' (answered)' : ''}`}
                aria-current={i === step}
                className={`h-2.5 w-full rounded-full transition ${
                  i === step
                    ? 'bg-gopher-500 ring-2 ring-gopher-500/40'
                    : done[i]
                      ? 'bg-gopher-300 dark:bg-gopher-700'
                      : 'bg-slate-200 dark:bg-slate-800'
                }`}
              />
            </li>
          ))}
        </ol>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm sm:p-8 dark:border-slate-800 dark:bg-slate-900">
        <QuestionView key={q.id} question={q} value={answers[q.id]} onChange={(v) => onAnswer(q.id, v)} />
      </div>

      {submitError && (
        <p role="alert" className="mt-4 rounded-xl bg-red-50 p-3 text-sm text-red-700 dark:bg-red-950/40 dark:text-red-300">
          {submitError}
        </p>
      )}

      <div className="mt-6 flex items-center justify-between gap-3">
        <button type="button" disabled={step === 0} onClick={() => onStep(step - 1)} className={`${btn} ${btnGhost}`}>
          ← Back
        </button>
        {last ? (
          <button
            type="button"
            disabled={remaining > 0 || submitting}
            onClick={onSubmit}
            className={`${btn} ${btnPrimary}`}
            title={remaining > 0 ? `${remaining} unanswered` : undefined}
          >
            {submitting ? 'Grading…' : remaining > 0 ? `${remaining} left to answer` : 'Submit quiz'}
          </button>
        ) : (
          <button
            type="button"
            disabled={!done[step]}
            onClick={() => onStep(step + 1)}
            className={`${btn} ${btnPrimary}`}
          >
            Next →
          </button>
        )}
      </div>
    </>
  )
}

function Results({
  quiz,
  result,
  answers,
  chapterPath,
  nextChapter,
  onRetry,
}: {
  quiz: Quiz
  result: QuizResult
  answers: Record<string, Answer>
  chapterPath: string
  nextChapter?: { path: string; title: string }
  onRetry: () => void
}) {
  const [onlyMistakes, setOnlyMistakes] = useState(false)
  const byId = new Map(result.results.map((r) => [r.id, r]))
  const mistakes = result.results.filter((r) => !r.correct).length

  const R = 52
  const C = 2 * Math.PI * R
  const color = result.passed ? 'text-emerald-500' : 'text-amber-500'

  return (
    <>
      <section
        className="rounded-3xl border border-slate-200 bg-white p-6 text-center shadow-sm sm:p-10 dark:border-slate-800 dark:bg-slate-900"
        aria-live="polite"
      >
        <div className="relative mx-auto size-32">
          <svg viewBox="0 0 120 120" className="size-full -rotate-90" aria-hidden>
            <circle cx="60" cy="60" r={R} fill="none" strokeWidth="10" className="stroke-slate-200 dark:stroke-slate-800" />
            <circle
              cx="60"
              cy="60"
              r={R}
              fill="none"
              strokeWidth="10"
              strokeLinecap="round"
              strokeDasharray={C}
              strokeDashoffset={C * (1 - result.percent / 100)}
              className={`stroke-current transition-all duration-700 ${color}`}
            />
          </svg>
          <div className="absolute inset-0 grid place-items-center">
            <div>
              <div className="text-3xl font-extrabold">{result.percent}%</div>
              <div className="text-xs text-slate-500">
                {result.score}/{result.total}
              </div>
            </div>
          </div>
        </div>

        <h1 className="mt-5 text-2xl font-bold tracking-tight sm:text-3xl">
          {result.passed ? (result.percent === 100 ? '🏆 Perfect score!' : '🎉 Quiz passed!') : '💪 Almost there'}
        </h1>
        <p className="mt-2 text-slate-500 dark:text-slate-400">
          {result.passed
            ? 'Nice work. You can retake the quiz any time to improve your score.'
            : `You need ${result.passPercent}% to pass. Review the explanations below and try again.`}
        </p>

        {result.reward.xpGained > 0 && (
          <p className="mx-auto mt-4 w-fit rounded-full bg-gopher-500/10 px-4 py-1.5 text-sm font-bold text-gopher-700 dark:text-gopher-300">
            ⚡ +{result.reward.xpGained} XP earned
          </p>
        )}

        <div className="mt-6 flex flex-wrap justify-center gap-3">
          {result.passed && nextChapter && (
            <Link to={nextChapter.path} className={`${btn} ${btnPrimary}`}>
              Next: {nextChapter.title} →
            </Link>
          )}
          <button type="button" onClick={onRetry} className={`${btn} ${result.passed ? btnGhost : btnPrimary}`}>
            Try again
          </button>
          <Link to={chapterPath} className={`${btn} ${btnGhost}`}>
            Back to chapter
          </Link>
        </div>
      </section>

      <div className="mb-4 mt-10 flex items-center justify-between gap-3">
        <h2 className="text-xl font-bold">Review</h2>
        {mistakes > 0 && (
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={onlyMistakes}
              onChange={(e) => setOnlyMistakes(e.target.checked)}
              className="size-4 accent-gopher-500"
            />
            Only mistakes ({mistakes})
          </label>
        )}
      </div>

      <ol className="flex flex-col gap-5">
        {quiz.questions.map((q, i) => {
          const r = byId.get(q.id)!
          if (onlyMistakes && r.correct) return null
          return (
            <li
              key={q.id}
              className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm sm:p-6 dark:border-slate-800 dark:bg-slate-900"
            >
              <div className="mb-3 flex items-center gap-2 text-sm font-semibold">
                <span
                  className={`grid size-6 place-items-center rounded-full text-xs text-white ${r.correct ? 'bg-emerald-500' : 'bg-red-500'}`}
                  aria-hidden
                >
                  {r.correct ? '✓' : '✗'}
                </span>
                <span>
                  Question {i + 1} · {r.correct ? 'Correct' : 'Incorrect'}
                </span>
              </div>
              <QuestionView question={q} value={answers[q.id]} onChange={() => {}} result={r} />
              <div className="mt-4 rounded-xl bg-slate-100 p-4 text-sm leading-relaxed dark:bg-slate-800/70">
                <span className="font-semibold">💡 Explanation: </span>
                <Inline text={r.explanation} />
              </div>
            </li>
          )
        })}
      </ol>
    </>
  )
}

const btn =
  'inline-flex items-center justify-center rounded-xl px-5 py-2.5 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-50'
const btnPrimary = 'bg-gopher-500 text-white hover:bg-gopher-600 disabled:hover:bg-gopher-500'
const btnGhost =
  'border border-slate-200 hover:bg-slate-100 disabled:hover:bg-transparent dark:border-slate-700 dark:hover:bg-slate-800'
