export default function Spinner() {
  return (
    <div className="grid place-items-center py-20" role="status" aria-label="Loading">
      <div className="size-8 animate-spin rounded-full border-4 border-gopher-500/30 border-t-gopher-500" />
    </div>
  )
}
