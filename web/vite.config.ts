import { writeFileSync } from 'node:fs'
import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// Vite empties dist/ on build; restore the placeholder that lets
// `go build` succeed (go:embed needs at least one file) on a fresh clone.
const keepGitkeep = (): Plugin => ({
  name: 'keep-gitkeep',
  closeBundle() {
    writeFileSync('dist/.gitkeep', '')
  },
})

export default defineConfig({
  plugins: [react(), tailwindcss(), keepGitkeep()],
  // The lazily loaded lesson chunk (Markdown + CodeMirror) is ~620 kB; fine for a local app.
  build: { chunkSizeWarningLimit: 700 },
  server: {
    port: 5173,
    proxy: { '/api': 'http://127.0.0.1:8080' },
  },
})
