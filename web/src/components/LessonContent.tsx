import { useMemo } from 'react'
import ReactMarkdown, { type Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import CodeBlock from './CodeBlock'
import type { LessonRef } from '../api'
import Playground from './Playground'

// Minimal structural type for the hast nodes react-markdown passes in.
interface HastNode {
  type: string
  tagName?: string
  value?: string
  properties?: { className?: unknown }
  data?: { meta?: string }
  children?: HastNode[]
}

const textOf = (n: HastNode): string => n.value ?? (n.children ?? []).map(textOf).join('')

// Fenced code: ```go run  → editable, runnable Playground
//              ```go      → highlighted read-only snippet
function makeComponents(lesson?: LessonRef): Components {
  return {
    pre({ node, children }) {
      const code = (node as HastNode | undefined)?.children?.[0]
      if (code?.tagName !== 'code') return <pre>{children}</pre>

      const classes = Array.isArray(code.properties?.className) ? (code.properties.className as string[]) : []
      const lang = classes.find((c) => c.startsWith('language-'))?.slice('language-'.length) ?? ''
      const meta = code.data?.meta ?? ''
      const text = textOf(code).replace(/\n$/, '')

      if (lang === 'go' && /\brun\b/.test(meta)) return <Playground code={text} lesson={lesson} />
      return (
        <div className="not-prose">
          <CodeBlock code={text} lang={lang} />
        </div>
      )
    },
    a({ href, children }) {
      const external = href?.startsWith('http')
      return (
        <a href={href} {...(external && { target: '_blank', rel: 'noreferrer' })}>
          {children}
        </a>
      )
    },
  }
}

export default function LessonContent({ markdown, lesson }: { markdown: string; lesson?: LessonRef }) {
  const components = useMemo(() => makeComponents(lesson), [lesson?.chapterId, lesson?.lessonId])
  return (
    <div className="prose prose-slate max-w-none dark:prose-invert prose-headings:tracking-tight prose-a:text-gopher-600 dark:prose-a:text-gopher-300 prose-code:rounded prose-code:bg-slate-100 prose-code:px-1.5 prose-code:py-0.5 prose-code:font-medium prose-code:before:content-none prose-code:after:content-none dark:prose-code:bg-slate-800">
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
        {markdown}
      </ReactMarkdown>
    </div>
  )
}
