import { useEffect, useRef, useState } from 'react'
import { basicSetup } from 'codemirror'
import { indentWithTab } from '@codemirror/commands'
import { go } from '@codemirror/lang-go'
import { indentUnit, syntaxHighlighting } from '@codemirror/language'
import { lintGutter, setDiagnostics, type Diagnostic } from '@codemirror/lint'
import { EditorState, Prec, type Text } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import { oneDarkTheme } from '@codemirror/theme-one-dark'
import { classHighlighter } from '@lezer/highlight'
import { api, type LessonRef, type RunEvent } from '../api'
import { parseGoErrors, type GoError } from '../lib/goErrors'
import { useRewards } from '../rewards'
import CopyButton from './CopyButton'

type Output =
  | { status: 'idle' }
  | { status: 'running'; events: RunEvent[]; startedAt: number }
  | { status: 'done'; events: RunEvent[]; exitCode: number; vetErrors?: string }
  | { status: 'compile-error'; text: string; errors: GoError[] }
  | { status: 'unavailable'; message: string }

// Longest pause replayed between output events (the Playground simulates
// time.Sleep; replaying it makes concurrency examples feel real).
const maxReplayDelayMs = 1000

const isMac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform)

const editorTheme = EditorView.theme(
  {
    '&': { fontSize: '13.5px', backgroundColor: '#1e2229', maxHeight: '480px' },
    '.cm-scroller': { overflow: 'auto', fontFamily: 'var(--font-mono)', lineHeight: '1.6' },
    '.cm-content': { padding: '12px 0' },
    '.cm-gutters': { backgroundColor: '#1e2229', border: 'none' },
    '.cm-gutters .cm-gutterElement': { color: '#9aa3b2' },
    '.cm-gutters .cm-activeLineGutter': { color: '#c5ccd8' },
    '.cm-activeLine, .cm-activeLineGutter': { backgroundColor: '#ffffff08' },
    '&.cm-focused': { outline: 'none' },
  },
  { dark: true },
)

function toDiagnostics(doc: Text, errors: GoError[], severity: Diagnostic['severity']): Diagnostic[] {
  return errors
    .filter((e) => e.line >= 1 && e.line <= doc.lines)
    .map((e) => {
      const line = doc.line(e.line)
      const from = Math.min(line.from + Math.max(e.col - 1, 0), line.to)
      return { from, to: line.to, severity, message: e.message }
    })
}

// placeholderHeight reserves the editor's eventual height (13.5px font at
// 1.6 line height, 12px padding top and bottom, capped like the editor) so
// the page doesn't jump when the editor appears.
function placeholderHeight(code: string): number {
  return Math.min(480, code.split('\n').length * 21.6 + 24)
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

export default function Playground({ code: initialCode, lesson }: { code: string; lesson?: LessonRef }) {
  const { celebrate } = useRewards()
  const rootRef = useRef<HTMLDivElement>(null)
  const hostRef = useRef<HTMLDivElement>(null)
  // CodeMirror is costly to start, so lessons with many examples only create
  // an editor once it is near the viewport.
  const [near, setNear] = useState(() => typeof IntersectionObserver === 'undefined')
  const viewRef = useRef<EditorView | null>(null)
  const abortRef = useRef<AbortController | null>(null)
  const runRef = useRef<() => void>(() => {})
  const [output, setOutput] = useState<Output>({ status: 'idle' })
  const [dirty, setDirty] = useState(false)
  const [formatError, setFormatError] = useState<string | null>(null)

  useEffect(() => {
    if (near) return
    const io = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) setNear(true)
      },
      { rootMargin: '600px' },
    )
    io.observe(rootRef.current!)
    return () => io.disconnect()
  }, [near])

  useEffect(() => {
    if (!near) return
    const view = new EditorView({
      parent: hostRef.current!,
      state: EditorState.create({
        doc: initialCode,
        extensions: [
          Prec.highest(keymap.of([{ key: 'Mod-Enter', run: () => (runRef.current(), true) }])),
          basicSetup,
          keymap.of([indentWithTab]),
          go(),
          indentUnit.of('\t'),
          EditorState.tabSize.of(4),
          syntaxHighlighting(classHighlighter),
          oneDarkTheme,
          editorTheme,
          lintGutter(),
          EditorView.contentAttributes.of({ 'aria-label': 'Go code editor' }),
          EditorView.updateListener.of((u) => {
            if (u.docChanged) setDirty(u.state.doc.toString() !== initialCode)
          }),
        ],
      }),
    })
    viewRef.current = view
    return () => {
      abortRef.current?.abort()
      view.destroy()
    }
  }, [initialCode, near])

  const getCode = () => viewRef.current?.state.doc.toString() ?? initialCode

  const showDiagnostics = (errors: GoError[], severity: Diagnostic['severity']) => {
    const view = viewRef.current
    if (view) view.dispatch(setDiagnostics(view.state, toDiagnostics(view.state.doc, errors, severity)))
  }

  const run = async () => {
    abortRef.current?.abort()
    const ac = new AbortController()
    abortRef.current = ac
    setFormatError(null)
    showDiagnostics([], 'error')
    const startedAt = Date.now()
    setOutput({ status: 'running', events: [], startedAt })

    try {
      const res = await api.run(getCode(), ac.signal, lesson)
      if (res.compileErrors) {
        const errors = parseGoErrors(res.compileErrors)
        showDiagnostics(errors, 'error')
        setOutput({ status: 'compile-error', text: res.compileErrors.trim(), errors })
        return
      }
      if (res.vetErrors) showDiagnostics(parseGoErrors(res.vetErrors), 'warning')

      const shown: RunEvent[] = []
      for (const ev of res.events) {
        if (ev.delayMs > 0) await sleep(Math.min(ev.delayMs, maxReplayDelayMs))
        if (ac.signal.aborted) return
        shown.push(ev)
        setOutput({ status: 'running', events: [...shown], startedAt })
      }
      setOutput({ status: 'done', events: shown, exitCode: res.exitCode, vetErrors: res.vetErrors })
      celebrate(res.reward)
    } catch (e) {
      if (ac.signal.aborted) return
      setOutput({ status: 'unavailable', message: (e as Error).message })
    }
  }
  runRef.current = run

  const format = async () => {
    const view = viewRef.current
    if (!view) return
    try {
      const { code } = await api.format(getCode())
      setFormatError(null)
      showDiagnostics([], 'error')
      if (code !== getCode()) view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: code } })
    } catch (e) {
      const msg = (e as Error).message
      setFormatError(msg)
      showDiagnostics(parseGoErrors(msg), 'error')
    }
  }

  const reset = () => {
    abortRef.current?.abort()
    const view = viewRef.current
    if (view) {
      view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: initialCode } })
      showDiagnostics([], 'error')
    }
    setFormatError(null)
    setOutput({ status: 'idle' })
  }

  const jumpTo = (e: GoError) => {
    const view = viewRef.current
    if (!view || e.line > view.state.doc.lines) return
    const line = view.state.doc.line(e.line)
    view.dispatch({ selection: { anchor: Math.min(line.from + e.col - 1, line.to) }, scrollIntoView: true })
    view.focus()
  }

  const running = output.status === 'running'

  return (
    <div ref={rootRef} className="not-prose code-theme my-6 overflow-hidden rounded-xl border border-slate-800 bg-[#1e2229] shadow-md">
      <div className="flex items-center gap-1 border-b border-white/10 bg-[#181b21] px-2 py-2 sm:gap-2 sm:px-3">
        <span className="mr-auto flex min-w-0 items-center gap-2 text-xs font-semibold text-slate-400">
          <span className="size-2 shrink-0 rounded-full bg-gopher-400" aria-hidden />
          <span className="hidden sm:inline">Try it</span>
          {dirty && <span className="truncate font-normal text-amber-400/90">edited</span>}
        </span>
        <ToolbarButton onClick={reset} disabled={!dirty && output.status === 'idle'} title="Restore the original code">
          Reset
        </ToolbarButton>
        <ToolbarButton onClick={format} title="Format with gofmt">
          Format
        </ToolbarButton>
        <CopyButton text={getCode} className={toolbarBtn} />
        <button
          type="button"
          onClick={run}
          disabled={running}
          title={`Run (${isMac ? '⌘' : 'Ctrl'}+Enter)`}
          className="inline-flex shrink-0 items-center gap-1.5 rounded-lg bg-gopher-500 px-3.5 py-1.5 text-xs font-bold text-white shadow-sm transition hover:bg-gopher-400 disabled:opacity-60"
        >
          {running ? <span className="size-3 animate-spin rounded-full border-2 border-white/40 border-t-white" /> : '▶'}
          Run
        </button>
      </div>

      <div ref={hostRef} style={near ? undefined : { minHeight: placeholderHeight(initialCode) }} />

      {formatError && (
        <div className="border-t border-white/10 bg-red-500/10 px-4 py-2 font-mono text-xs text-red-300">
          gofmt: {formatError}
        </div>
      )}

      {output.status !== 'idle' && (
        <div className="pg-output border-t border-white/10 bg-[#14161b]" data-status={output.status} aria-live="polite">
          <div className="flex items-center justify-between px-4 pt-2 text-[11px] font-semibold uppercase tracking-wider text-slate-500">
            Output
            {output.status === 'done' && (
              <span className={output.exitCode === 0 ? 'text-emerald-400/80' : 'text-red-400/80'}>
                {output.exitCode === 0 ? 'Program exited' : `Exit status ${output.exitCode}`}
              </span>
            )}
          </div>
          <OutputBody output={output} onJump={jumpTo} onRetry={run} />
        </div>
      )}
    </div>
  )
}

const toolbarBtn =
  'shrink-0 rounded-lg px-2 py-1.5 sm:px-2.5 text-xs font-medium text-slate-300 transition hover:bg-white/10 hover:text-white disabled:pointer-events-none disabled:opacity-40'

function ToolbarButton(props: { onClick: () => void; disabled?: boolean; title: string; children: string }) {
  return (
    <button type="button" className={toolbarBtn} {...props}>
      {props.children}
    </button>
  )
}

// Elapsed shows whole seconds since a timestamp; the first compile of a
// program on the Playground can take ~10s, so visible progress matters.
function Elapsed({ since }: { since: number }) {
  const [now, setNow] = useState(Date.now)
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(t)
  }, [])
  const s = Math.floor((now - since) / 1000)
  return s >= 1 ? <span className="tabular-nums">{s}s</span> : null
}

function OutputBody({ output, onJump, onRetry }: { output: Output; onJump: (e: GoError) => void; onRetry: () => void }) {
  const pre = 'max-h-72 overflow-auto px-4 pb-3 pt-1 font-mono text-[13px] leading-relaxed whitespace-pre-wrap'

  switch (output.status) {
    case 'running':
    case 'done':
      return (
        <>
          <pre className={pre}>
            {output.events.map((ev, i) => (
              <span key={i} className={ev.kind === 'stderr' ? 'text-red-300' : 'text-slate-100'}>
                {ev.message}
              </span>
            ))}
            {output.status === 'running' && output.events.length === 0 && (
              <span className="text-slate-500">
                Compiling on the Go Playground… <Elapsed since={output.startedAt} />
              </span>
            )}
            {output.status === 'done' && output.events.length === 0 && (
              <span className="text-slate-500">(no output)</span>
            )}
          </pre>
          {output.status === 'done' && output.vetErrors && (
            <pre className={`${pre} border-t border-white/5 text-amber-300`}>go vet: {output.vetErrors.trim()}</pre>
          )}
        </>
      )
    case 'compile-error':
      return (
        <div className="px-4 pb-3 pt-1">
          <p className="mb-1 text-xs font-semibold text-red-400">Compile error</p>
          {output.errors.length > 0 ? (
            <ul className="font-mono text-[13px] text-red-300">
              {output.errors.map((e, i) => (
                <li key={i}>
                  <button type="button" onClick={() => onJump(e)} className="text-left hover:underline">
                    <span className="text-red-400/70">
                      line {e.line}:{e.col}
                    </span>{' '}
                    {e.message}
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <pre className="font-mono text-[13px] whitespace-pre-wrap text-red-300">{output.text}</pre>
          )}
        </div>
      )
    case 'unavailable':
      return (
        <div className="flex flex-wrap items-center gap-3 px-4 pb-3 pt-1 text-sm text-amber-300">
          <span aria-hidden>📡</span>
          <span className="flex-1">{output.message}</span>
          <button
            type="button"
            onClick={onRetry}
            className="rounded-lg border border-amber-400/40 px-3 py-1 text-xs font-semibold hover:bg-amber-400/10"
          >
            Retry
          </button>
        </div>
      )
    default:
      return null
  }
}
