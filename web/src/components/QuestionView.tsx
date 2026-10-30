import { Fragment } from 'react'
import type { Answer, QuestionResult, QuestionType, QuizQuestion } from '../api'
import CodeBlock from './CodeBlock'
import Inline from './Inline'

export const typeLabels: Record<QuestionType, string> = {
  single: 'Multiple choice',
  multi: 'Select all that apply',
  output: 'Predict the output',
  fill: 'Fill in the gaps',
  bug: 'Spot the bug',
}

// isAnswered reports whether value is a complete answer for q.
export function isAnswered(q: QuizQuestion, value: Answer | undefined): boolean {
  switch (q.type) {
    case 'single':
    case 'bug':
      return typeof value === 'number'
    case 'multi':
      return Array.isArray(value) && value.length > 0
    case 'output':
      return typeof value === 'string' && value.trim() !== ''
    case 'fill': {
      const gaps = value as string[] | undefined
      return !!gaps && gaps.length === q.blanks && gaps.every((g) => g?.trim())
    }
  }
}

interface Props {
  question: QuizQuestion
  value: Answer | undefined
  onChange: (v: Answer) => void
  // When set, the question is read-only and the correct answer is highlighted.
  result?: QuestionResult
}

const toneClass = {
  idle: 'border-slate-200 hover:border-gopher-400 dark:border-slate-700 dark:hover:border-gopher-600',
  selected: 'border-gopher-500 bg-gopher-500/10 dark:border-gopher-400',
  right: 'border-emerald-500 bg-emerald-500/10',
  wrong: 'border-red-500 bg-red-500/10',
  muted: 'border-slate-200 opacity-70 dark:border-slate-800',
}

// tone picks the style for an option/line; in review mode the correct one is
// green, a wrong pick is red and the rest are dimmed.
function tone(selected: boolean, review: boolean, isCorrect: boolean): keyof typeof toneClass {
  if (!review) return selected ? 'selected' : 'idle'
  if (isCorrect) return 'right'
  return selected ? 'wrong' : 'muted'
}

const optionBase = 'flex w-full cursor-pointer items-start gap-3 rounded-xl border-2 p-3.5 text-left transition'

export default function QuestionView({ question: q, value, onChange, result }: Props) {
  const review = !!result
  const correct = result?.correctAnswer

  return (
    <div>
      <p className="text-lg font-semibold leading-snug">
        <Inline text={q.prompt} />
      </p>

      {q.code && q.type !== 'fill' && q.type !== 'bug' && (
        <div className="not-prose">
          <CodeBlock code={q.code.replace(/\n$/, '')} lang="go" />
        </div>
      )}

      {(q.type === 'single' || q.type === 'multi') && (
        <fieldset className="mt-4 flex flex-col gap-2.5" disabled={review}>
          <legend className="sr-only">{typeLabels[q.type]}</legend>
          {q.options?.map((opt, i) => {
            const multi = q.type === 'multi'
            const selected = multi ? ((value as number[] | undefined) ?? []).includes(i) : value === i
            const isCorrect = multi ? (correct as number[] | undefined)?.includes(i) === true : correct === i
            return (
              <label key={i} className={`${optionBase} ${toneClass[tone(selected, review, isCorrect)]}`}>
                <input
                  type={multi ? 'checkbox' : 'radio'}
                  name={q.id}
                  checked={selected}
                  className="mt-1 size-4 accent-gopher-500"
                  onChange={() => {
                    if (!multi) return onChange(i)
                    const cur = (value as number[] | undefined) ?? []
                    onChange(selected ? cur.filter((n) => n !== i) : [...cur, i].sort((a, b) => a - b))
                  }}
                />
                <span className="flex-1">
                  <Inline text={opt} />
                </span>
                {review && isCorrect && !selected && <span className="text-xs font-semibold text-emerald-600 dark:text-emerald-400">you missed this</span>}
                {review && isCorrect && <Mark>✓</Mark>}
                {review && selected && !isCorrect && <Mark>✗</Mark>}
              </label>
            )
          })}
        </fieldset>
      )}

      {q.type === 'output' && (
        <div className="mt-4">
          <label className="mb-1.5 block text-sm font-medium text-slate-500" htmlFor={`a-${q.id}`}>
            Program output
          </label>
          <textarea
            id={`a-${q.id}`}
            rows={3}
            disabled={review}
            spellCheck={false}
            autoCapitalize="off"
            autoCorrect="off"
            value={(value as string | undefined) ?? ''}
            onChange={(e) => onChange(e.target.value)}
            placeholder="Type exactly what is printed (one line per line)"
            className={`w-full resize-y rounded-xl border-2 bg-white p-3 font-mono text-sm outline-none focus:border-gopher-500 dark:bg-slate-900 ${
              review ? toneClass[result.correct ? 'right' : 'wrong'] : 'border-slate-200 dark:border-slate-700'
            }`}
          />
          {review && !result.correct && <Expected>{correct as string}</Expected>}
        </div>
      )}

      {q.type === 'fill' && q.code && (
        <FillCode
          question={q}
          values={(value as string[] | undefined) ?? []}
          disabled={review}
          rightAnswer={review ? (correct as string[]) : undefined}
          wrong={review && !result.correct}
          onChange={onChange}
        />
      )}

      {q.type === 'bug' && q.code && (
        <ol className="code-theme mt-4 overflow-x-auto rounded-xl border border-slate-800 bg-[#1e2229] py-2 font-mono text-[13.5px]">
          {q.code
            .replace(/\n$/, '')
            .split('\n')
            .map((line, i) => {
              const n = i + 1
              const selected = value === n
              const isCorrect = correct === n
              const style = review
                ? isCorrect
                  ? 'bg-emerald-500/20'
                  : selected
                    ? 'bg-red-500/25'
                    : ''
                : selected
                  ? 'bg-gopher-500/25'
                  : 'hover:bg-white/5'
              return (
                <li key={n}>
                  <button
                    type="button"
                    disabled={review}
                    aria-pressed={selected}
                    aria-label={`Line ${n}: ${line || 'blank'}`}
                    onClick={() => onChange(n)}
                    className={`flex w-full min-w-max items-start gap-4 px-3 py-0.5 text-left ${style} ${review ? 'cursor-default' : 'cursor-pointer'}`}
                  >
                    <span className="w-5 shrink-0 select-none text-right text-slate-500">{n}</span>
                    <span className="whitespace-pre">{line || ' '}</span>
                    {review && isCorrect && <span className="ml-auto text-emerald-400">✓ bug</span>}
                    {review && selected && !isCorrect && <span className="ml-auto text-red-400">✗ your pick</span>}
                  </button>
                </li>
              )
            })}
        </ol>
      )}
      {q.type === 'bug' && !review && (
        <p className="mt-2 text-sm text-slate-500">Click the line that contains the bug.</p>
      )}
    </div>
  )
}

function Mark({ children }: { children: string }) {
  return (
    <span aria-hidden className="font-bold">
      {children}
    </span>
  )
}

function Expected({ children }: { children: string }) {
  return (
    <div className="mt-2 text-sm">
      <span className="font-semibold text-emerald-600 dark:text-emerald-400">Expected output:</span>
      <pre className="mt-1 overflow-x-auto rounded-lg bg-emerald-500/10 p-3 font-mono text-sm">{children}</pre>
    </div>
  )
}

// FillCode shows the program with an input in place of each ___ gap.
function FillCode({
  question: q,
  values,
  disabled,
  rightAnswer,
  wrong,
  onChange,
}: {
  question: QuizQuestion
  values: string[]
  disabled: boolean
  rightAnswer?: string[]
  wrong: boolean
  onChange: (v: Answer) => void
}) {
  const parts = q.code!.replace(/\n$/, '').split('___')
  const set = (i: number, v: string) => {
    const next = Array.from({ length: parts.length - 1 }, (_, k) => values[k] ?? '')
    next[i] = v
    onChange(next)
  }
  return (
    <div className="mt-4">
      <pre className="code-theme overflow-x-auto rounded-xl border border-slate-800 bg-[#1e2229] p-4 font-mono text-[13.5px] leading-7">
        {parts.map((part, i) => (
          <Fragment key={i}>
            {part}
            {i < parts.length - 1 && (
              <input
                type="text"
                aria-label={`Gap ${i + 1}`}
                value={values[i] ?? ''}
                disabled={disabled}
                spellCheck={false}
                autoCapitalize="off"
                autoCorrect="off"
                size={Math.max(6, (rightAnswer?.[i] ?? values[i] ?? '').length + 2)}
                onChange={(e) => set(i, e.target.value)}
                className={`mx-0.5 rounded-md border-b-2 bg-white/10 px-2 py-0.5 text-center text-gopher-300 outline-none focus:bg-white/20 ${
                  wrong ? 'border-red-400' : disabled ? 'border-emerald-400' : 'border-gopher-400'
                }`}
              />
            )}
          </Fragment>
        ))}
      </pre>
      {wrong && rightAnswer && (
        <p className="mt-2 text-sm">
          <span className="font-semibold text-emerald-600 dark:text-emerald-400">Expected: </span>
          {rightAnswer.map((a, i) => (
            <code key={i} className="mr-2 rounded bg-emerald-500/10 px-1.5 py-0.5 font-mono">
              {a}
            </code>
          ))}
        </p>
      )}
    </div>
  )
}
