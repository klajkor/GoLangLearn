# Delivery Plan: Gamified Go Learning App

A local, single-user web app (opened in Chrome) that takes a learner from zero to advanced Go, with gamification and a quiz after each chapter.

## 1. Decisions

| Topic | Decision | Impact |
|---|---|---|
| Running learner code | Go Playground (`play.golang.org`) via a backend proxy | Needs internet to run code; snippets go to Google's public service. Lessons, quizzes and progress work offline. |
| Stack | Go backend + React + TypeScript + Tailwind (Vite) | Single Go binary with the built frontend embedded via `go:embed`. |
| Content | Written by Claude as Markdown/YAML files | About 18 chapters, editable without code changes. |
| Content review | Once, at the end (after Phase 6) | No review gates between tiers. |
| Gamification | XP, levels, streaks, badges, skill tree | Timed coding challenges are in the backlog. |
| Toolchain | Go 1.27 (installed via winget), Node 24 (frontend build only) | |
| Code editor | CodeMirror 6 instead of Monaco (changed in Phase 2) | Monaco handles touch input poorly, is several MB and needs web workers; CodeMirror is ~10× smaller, works on phones and has a Go mode. |
| Formatting | `gofmt` runs locally via Go's `go/format` | The Format button works offline. |

## 2. Architecture

```text
Browser (Chrome) ── http://localhost:8080
   │  React + Vite + TypeScript + Tailwind (built to static files)
   │  CodeMirror 6 editor · Framer Motion animations · responsive, dark/light
   ▼
Go server (single binary, frontend embedded with go:embed)
   ├─ /api/curriculum   → chapters/lessons loaded from /content (Markdown + YAML front-matter)
   ├─ /api/chapters/{c}/quiz → GET questions (no answers); POST answers, graded server-side
   ├─ /api/progress     → XP, level, streak, badges, unlocks
   ├─ /api/run          → proxy to Go Playground (avoids CORS, adds timeout and size limit)
   ├─ /api/format       → local gofmt (go/format)
   └─ SQLite (modernc.org/sqlite, pure Go, no CGO) → progress.db
```

- **Content is kept separate from code.** Each chapter is a folder with `chapter.yaml`, `lessons/NN-name.md` and (from Phase 3) `quiz.yaml`. Runnable examples live inline in lessons as ```` ```go run ```` fences; plain ```` ```go ```` fences are read-only snippets.
- **Quiz types:** multiple choice, multi-select, predict the output, fill-in-the-blank code, spot the bug. Each chapter has an optional `quiz.yaml`; it is validated when the server starts (and by `go test`), so a typo in a quiz can never break a learner mid-quiz.
- **Feedback timing:** explanations appear in a review after the whole quiz is submitted (not after each question), so grading stays on the server and the answers never reach the browser early.

## 3. Curriculum (about 18 chapters in 4 tiers)

| Tier | Chapters |
|---|---|
| **Beginner** | 1 Intro & toolchain · 2 Variables, types, constants · 3 Control flow · 4 Functions · 5 Arrays, slices, maps · 6 Strings, runes & formatting |
| **Intermediate** | 7 Structs & methods · 8 Interfaces · 9 Error handling · 10 Packages & modules · 11 Pointers & memory basics · 12 Testing & benchmarks |
| **Advanced** | 13 Goroutines & channels · 14 sync, select, context · 15 Generics · 16 Concurrency patterns (worker pools, pipelines) |
| **Expert** | 17 Reflection, `unsafe`, profiling (pprof) · 18 Building real services (net/http, JSON, project layout, idioms) |

Each chapter has 3–5 short lessons, runnable examples, 1–2 small practice tasks, and an end-of-chapter quiz of 8–12 questions.

## 4. Gamification design

- **XP:** lesson done +20, example run +5, quiz pass (≥70%) +50, perfect quiz +25 bonus, retry pass +25.
- **Levels:** a curved XP scale with Go-themed titles (Gopher Hatchling → Goroutine Wrangler → Channel Master → Go Guru), shown with a level-up animation.
- **Streaks:** daily activity counter with one streak freeze per week, plus a calendar heatmap.
- **Badges (about 20):** First Run, Perfect Quiz, 7-Day Streak, Concurrency Unlocked, Tier Complete ×4, Night Owl, and others.
- **Skill tree:** an interactive map of chapters. A node unlocks when the previous chapter's quiz is passed, and each node shows locked, in-progress or mastered.
- **Dashboard:** current level, XP bar, streak, next recommended lesson, recent badges.

## 5. Phased delivery

| Phase | Scope | Exit criteria |
|---|---|---|
| **0. Setup** (½ day) | Repo layout (`/cmd`, `/internal`, `/web`, `/content`), `run.ps1`, git init | `go run ./cmd/server` serves a page |
| **1. Core skeleton** (1–2 days) | Go server, content loader, SQLite schema, React shell with routing, responsive layout, theme | You can move between placeholder chapters on desktop and mobile widths |
| **2. Lessons + code runner** (1–2 days) | Markdown rendering with syntax highlighting, CodeMirror editor, Playground proxy with error handling and an offline state | You can edit and run an example and see its output or compile errors |
| **3. Quiz engine** (1–2 days) | All 5 question types, server-side grading, explanations in a review after submitting, retry | A full quiz works end to end and the score is saved |
| **4. Gamification** (2 days) | XP/levels, streaks, badges, skill-tree visualisation, dashboard, animations | Progress and unlocks survive a restart |
| **5. Content: Beginner + Intermediate** (2–3 days) | Chapters 1–12 with quizzes | Every example compiles on the Playground and every quiz has an answer key |
| **6. Content: Advanced + Expert** (2 days) | Chapters 13–18 | Same checks as Phase 5; then the full content review |
| **7. Polish & QA** (1 day) | Accessibility pass, Lighthouse check, Go unit tests for grading/XP, README, single-binary build | `go build` produces one `.exe`, and launching it opens the app in Chrome |

About 11–15 working sessions in total. **The MVP is the end of Phase 4**, shown with 2–3 sample chapters.

## 6. Risks and mitigations

- **Playground dependency** (needs internet, has rate limits, code leaves the machine): the proxy has timeouts and clear error messages. A later config flag could switch to running code with the local Go install.
- **Playground latency** (measured in Phase 2: up to ~10 s for a first compile, <1 s for programs it has already cached): the UI shows an elapsed-seconds counter while compiling.
- **Content accuracy:** a script compiles every example through the Playground. Quiz answers are checked during QA, and the full content is reviewed at the end.
- **Corporate network or proxy** may block `play.golang.org`: check this early.

## 7. Backlog (not in scope)

Timed coding challenges with hidden tests, spaced-repetition review of missed quiz questions, progress export/import, a mode that runs code with the local Go install.

## 8. Status

| Phase | Status |
|---|---|
| 0. Setup | ✅ Done (2026-10-30) |
| 1. Core skeleton | ✅ Done (2026-10-30): 18 chapter stubs, 3 placeholder lessons; verified at 1440px and 390px |
| 2. Lessons + code runner | ✅ Done (2026-10-30): Markdown, highlighted snippets, runnable editor with Run/Format/Reset/Copy, inline error underlines, vet warnings, replayed `time.Sleep`, offline state |
| 3. Quiz engine | ✅ Done (2026-10-30): 5 question types, load-time validation, server-side grading, saved attempts and best score, review with explanations; sample quizzes for chapters 1–2 |
| 4. Gamification | ✅ Done (2026-10-30): XP + 10 Go-themed levels, streaks with weekly freeze and calendar heatmap, 23 badges, skill map with chapter unlocks, dashboard with next step. XP: lesson +20, first example run per lesson +5, quiz pass +50 (+25 if passed on a retry instead), first 100% +25. Locks are advisory. |
| 5. Content ch. 1–12 | ✅ Written (2026-10-30): 57 lessons, 12 quizzes × 8 questions; `go run ./cmd/checkcontent` verifies every example and quiz output answer (175 programs) |
| 6. Content ch. 13–18 | ✅ Written (2026-10-30): all 18 chapters, 87 lessons (~60k words), 18 quizzes × 8 questions; `checkcontent` verifies 260 programs. **Awaiting your content review.** |
| 7. Polish & QA | ✅ Done (2026-10-30): single self-contained binary (`-open`, embedded course), Lighthouse accessibility 100 / best practices 100 on every page, main JS 699 → 273 KB with lazy pages, lazily created editors, long-term caching of fingerprinted assets, startup check for an incomplete UI build, friendly port-in-use error. Pending: your content review. |
