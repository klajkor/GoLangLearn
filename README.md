# GoLearn 🐹

A local, gamified web app for learning Go from zero to advanced. See [game-delivery-plan.md](game-delivery-plan.md) for the roadmap.

## Requirements

- Go 1.23+ (`winget install GoLang.Go`); the lessons use a few newer library features, which the Go Playground supports
- Internet access for the in-lesson **Run** buttons (code runs on the Go Playground); everything else works offline
- Node.js 20+ (only to build the frontend)

## Run

```powershell
./run.ps1          # build and open http://127.0.0.1:8080
./run.ps1 dev      # hot-reload UI on http://localhost:5173 (API proxied to :8080)
./run.ps1 build    # one self-contained golearn.exe (UI and course built in)
./run.ps1 test     # Go tests + TypeScript type-check
./run.ps1 reset    # erase all progress and start from scratch (asks to confirm)
```

Progress is stored in `data/progress.db` (SQLite) in the folder you start the app from. To start over, stop the app and run `./run.ps1 reset` (it asks for confirmation; `-Force` skips the question). That deletes the database files, which are recreated empty on the next start. For a standalone `golearn.exe` started elsewhere, delete `data/progress.db` (plus `-wal` and `-shm`) in that folder, or start it with a different `-db` file.

### As a single file

```powershell
./run.ps1 build
.\golearn.exe -open     # serves http://127.0.0.1:8080 and opens your browser
```

`golearn.exe` contains the web UI and all lessons, so you can copy it anywhere. Flags: `-addr` (listen address, default `127.0.0.1:8080`), `-db` (progress file), `-open` (open the browser), `-content <dir>` (read lessons from a folder instead of the built-in copy, handy while editing them).

If port 8080 is busy the app says so; use `-addr 127.0.0.1:8090`.

## Layout

| Path | Purpose |
|---|---|
| `cmd/server` | Entrypoint: HTTP server, serves API and embedded UI |
| `internal/content` | Loads chapters/lessons/quizzes from `content/` |
| `internal/game` | Pure rules: XP amounts, level curve, streaks, badge catalogue |
| `internal/store` | SQLite schema and progress queries |
| `internal/api` | JSON API (`/api/curriculum`, `/api/chapters/{c}/lessons/{l}`, `/api/chapters/{c}/quiz`, `/api/progress`, `POST .../lessons/{l}/complete`, `/api/run`, `/api/format`) |
| `internal/playground` | Go Playground client used by `/api/run` |
| `web/` | React + TypeScript + Tailwind frontend (Vite); built into `web/dist` and embedded via `go:embed` |
| `content/` | Curriculum: `NN-chapter/chapter.yaml` + `lessons/NN-lesson.md` |

## Adding content

```text
content/03-control-flow/
  chapter.yaml            # title, tier (beginner|intermediate|advanced|expert), icon, summary
  lessons/01-if-else.md   # front matter: title, minutes; then Markdown
  quiz.yaml               # optional end-of-chapter quiz
```

The numeric prefix sets the order and is dropped from the ID (`03-control-flow` → `control-flow`). The front-matter `title` is shown as the page heading, so start the body at `##`. Restart the server to pick up changes.

Code fences in lessons:

- ```` ```go run ```` → editable example with Run / Format / Reset (runs on the [Go Playground](https://go.dev/play), needs internet)
- ```` ```go ```` → read-only, syntax-highlighted snippet
- any other language (```` ```bash ````) → plain read-only block

## Writing a quiz

`quiz.yaml` lists questions; every question needs `id`, `type`, `prompt` and `explanation`. `passPercent` defaults to 70. Wrap code words in `backticks` in prompts, options and explanations. See [content/01-intro/quiz.yaml](content/01-intro/quiz.yaml) for a complete example.

| `type` | Extra fields | `answer` |
|---|---|---|
| `single` | `options` | index of the correct option (from 0) |
| `multi` | `options` | list of correct option indexes |
| `output` | `code` | the printed text, or a list of accepted texts (trailing spaces and line endings are ignored) |
| `fill` | `code` with `___` gaps | one entry per gap: text or a list of accepted texts (case-sensitive) |
| `bug` | `code` | 1-based line number of the buggy line |

The server checks every quiz when it starts and `go test` checks the whole `content/` folder, so mistakes are reported with the question number and reason.

## Checking content

```bash
go run ./cmd/checkcontent              # every chapter (a few minutes)
go run ./cmd/checkcontent errors 05-collections   # only some chapters
```

The checker compiles, vets and runs every `go run` example, compares each quiz
`output` answer with what the program really prints, and compiles every `fill`
question with its gaps filled in. `go test ./...` also validates all quizzes.
Choice options are shuffled (stably, per question) when the server loads, so
write the answer key in any order.
