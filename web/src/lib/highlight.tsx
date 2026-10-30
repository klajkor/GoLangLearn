import type { ReactNode } from 'react'
import { goLanguage } from '@codemirror/lang-go'
import { classHighlighter, highlightCode } from '@lezer/highlight'

// highlightGo returns syntax-highlighted Go as React nodes. Token classes
// (tok-*) are the same ones the editor uses, styled in index.css.
export function highlightGo(code: string): ReactNode[] {
  const out: ReactNode[] = []
  highlightCode(
    code,
    goLanguage.parser.parse(code),
    classHighlighter,
    (text, classes) => {
      out.push(
        classes ? (
          <span key={out.length} className={classes}>
            {text}
          </span>
        ) : (
          text
        ),
      )
    },
    () => out.push('\n'),
  )
  return out
}
