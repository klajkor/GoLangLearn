import { useState } from 'react'

export default function CopyButton({ text, className = '' }: { text: () => string; className?: string }) {
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text())
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // clipboard blocked; nothing useful to do
    }
  }

  return (
    <button type="button" onClick={copy} className={className} aria-label="Copy code">
      {copied ? '✓ Copied' : 'Copy'}
    </button>
  )
}
