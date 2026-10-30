export interface GoError {
  line: number
  col: number
  message: string
}

// parseGoErrors extracts positioned errors from compiler, vet and gofmt
// output, e.g. "./prog.go:5:2: undefined: x" or "3:1: expected ';'".
export function parseGoErrors(text: string): GoError[] {
  const out: GoError[] = []
  for (const raw of text.split('\n')) {
    const m = raw.match(/^(?:\.\/)?(?:prog\.go:)?(\d+):(\d+):\s*(.*)$/)
    if (m) out.push({ line: Number(m[1]), col: Number(m[2]), message: m[3] })
  }
  return out
}
