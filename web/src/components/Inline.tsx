// Inline renders `backtick` spans in quiz text as inline code.
export default function Inline({ text }: { text: string }) {
  return (
    <>
      {text.split('`').map((part, i) =>
        i % 2 === 1 ? (
          <code key={i} className="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-[0.9em] dark:bg-slate-800">
            {part}
          </code>
        ) : (
          part
        ),
      )}
    </>
  )
}
