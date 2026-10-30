import { highlightGo } from '../lib/highlight'
import CopyButton from './CopyButton'

// CodeBlock renders a read-only snippet; Go is syntax-highlighted.
export default function CodeBlock({ code, lang }: { code: string; lang: string }) {
  return (
    <div className="code-theme group relative my-6 overflow-hidden rounded-xl border border-slate-800 bg-[#1e2229] shadow-sm">
      {lang && (
        <span className="absolute left-4 top-2 text-[11px] font-semibold uppercase tracking-wider text-slate-500">
          {lang}
        </span>
      )}
      <CopyButton
        text={() => code}
        className="absolute right-2 top-2 rounded-md px-2 py-1 text-xs text-slate-400 opacity-0 transition hover:bg-white/10 hover:text-white focus:opacity-100 group-hover:opacity-100"
      />
      <pre className={`overflow-x-auto px-4 pb-4 font-mono text-[13.5px] leading-relaxed ${lang ? 'pt-8' : 'pt-4'}`}>
        <code>{lang === 'go' ? highlightGo(code) : code}</code>
      </pre>
    </div>
  )
}
