# AGENTS.md — Master Context for AI Coding Agents

> This file is the **single source of truth** loaded by OpenCode, Claude Code,
> Cursor, and other AI coding agents working on this repository. Read it fully
> before doing any code change. If you cannot fit it all in context, prioritize
> the **Mission**, **Hard Rules**, and **Current Phase** sections.

---

> **Status note (auto-applied by setup-pack):**
> - `executor/claude.go` was a phantom — Claude is served by `DefaultExecutor`
>   after translation. Removed from layout.
> - `auth/oauth/opencode.go` was a phantom — `opencode` is an executor only,
>   not an OAuth flow. Removed.
> - OAuth flows for `qoder` and `openai` were missing from the spec — added.
> - RTK filter `git-log` is declared in `constants.js` but **not implemented**
>   in the 9router registry; do NOT port it.
> - All `docs/*.md` files now exist as stubs. Fill them in as you port code.

## 🎯 Mission

We are porting **9router** (`https://github.com/decolua/9router`) — a Next.js
AI gateway — to a new stack:

- **Backend:** Go (single binary, OpenAI-compatible gateway, SSE proxy)
- **Frontend:** Next.js (dashboard only, talks to Go via REST/SSE)

The new project must achieve **100% functional parity** with 9router's runtime
behavior. UI/UX can be modernized, but every observable feature, endpoint,
format, fallback rule, token-saver behavior, and quirk must match.

### Reference implementation
The 9router repo is the **golden source**. When in doubt:
1. Read the corresponding `.js` file in 9router.
2. Mirror its behavior in idiomatic Go.
3. Add a comment: `// ref: open-sse/path/to/file.js:LINE`.

The 9router source is checked out at `./_ref/9router/` (do not modify it).

---

## 🧱 Architecture Snapshot

```
┌───────────────────────────┐
│ CLI Tools (Claude Code,   │
│ Cursor, Codex, OpenClaw…) │
└─────────────┬─────────────┘
              │ /v1/* (OpenAI-compatible, SSE)
              ▼
┌─────────────────────────────────────────────┐
│  Go Backend (port 20128)                    │
│  ├── /v1/*          public gateway          │
│  ├── /api/*         admin API (JWT)         │
│  └── /dashboard/*   embedded Next.js export │
└─────────────┬───────────────────────────────┘
              │ HTTPS
              ▼
        Upstream LLM Providers (40+)

┌─────────────────────────┐
│ Next.js dashboard       │
│ (build → static export  │
│  → embed.FS in Go)      │
└─────────────────────────┘
```

**Both produced binaries: ONE.** Frontend is built and embedded into the Go
binary via `//go:embed`. Distribution is a single executable, just like
9router's npm package.

---

## 🚨 Hard Rules (NEVER break these)

1. **No CGO**. Use `modernc.org/sqlite` (pure-Go) for SQLite. This matters for
   cross-compilation and matches 9router's "no native build tools" goal.
2. **Stdlib first**. Reach for third-party only when stdlib costs >2x more
   code or measurable perf. Allowed core deps:
   `chi`, `slog`, `oauth2`, `jwt/v5`, `validator`, `modernc.org/sqlite`,
   `sqlc-generated`, `golang-migrate`, `sonic` (only in translator hot path).
3. **Context everywhere**. Every function that does I/O or might block takes
   `ctx context.Context` as first arg. Pass it through, do not store it.
4. **No goroutine without lifecycle**. Every `go func()` must be cancellable
   via context OR have a documented termination condition. No fire-and-forget.
5. **Streaming = `io.Reader`/`io.Writer`**. Never buffer a full SSE stream in
   memory. Use `http.Flusher` and chunked reads.
6. **Errors are values**. No panics in request path. Use typed errors
   (`internal/errs/`). Top-level recover middleware only catches bugs.
7. **Parity over elegance**. If 9router does something quirky (e.g., specific
   Anthropic-Beta header string), port it verbatim. We are not refactoring
   9router's semantics — only its implementation.
8. **Tests for translators**. Every request/response translator MUST have a
   golden-file test driven by fixtures captured from 9router.
9. **No secrets in code**. All secrets via env or DB (encrypted with AES-GCM).
10. **Single-binary deploy**. `go build` must produce ONE binary that boots,
    serves dashboard, and routes traffic. No external runtime deps.

---

## 📂 Repo Layout

See `docs/ARCHITECTURE.md` for the canonical tree. Quick view:

```
cmd/server/main.go            # entry point
internal/
  api/{v1,admin,middleware}   # HTTP routes
  router/                     # core orchestration (port of src/sse + open-sse/handlers)
  translator/                 # format conversion (port of open-sse/translator)
  executor/                   # per-provider adapters (port of open-sse/executors)
  rtk/                        # token-saver filters (port of open-sse/rtk)
  caveman/                    # caveman prompt injector (port of open-sse/rtk/caveman.js)
  auth/                       # JWT, API keys, OAuth
  storage/                    # SQLite + migrations
  models/                     # domain types
  stream/                     # SSE utilities
  config/                     # env loader
web/                          # Next.js dashboard (sub-project)
assets/web/                   # built static output, embedded via //go:embed
docs/                         # design specs (READ THESE)
_ref/9router/                 # reference 9router source (read-only)
```

---

## 🗺️ Current Phase

> **Update this section every time you finish a phase.** This tells the next
> agent where to pick up.

**Phase: 0 (Foundation)** — see `ROADMAP.md` for the full plan.

**Done:**
- (none yet)

**Next:**
- Initialize Go module, set up `chi` router, slog, env config.
- Set up Next.js dashboard scaffold (App Router, Tailwind, shadcn/ui).
- Wire JWT login flow (admin login → cookie).

---

## 🤝 Conventions for AI Agents

### When asked to add a feature
1. **Find the 9router equivalent.** Use `grep`/`rg` in `_ref/9router/`.
2. **Read `docs/PARITY_CHECKLIST.md`** to see if it's already mapped.
3. **Open `docs/<spec>.md`** for the subsystem (RTK, executors, translators).
4. **Write code** that mirrors behavior, not structure (idiomatic Go is fine).
5. **Add a parity-test fixture** if the feature transforms data.
6. **Update `docs/PARITY_CHECKLIST.md`** with a checkmark and file ref.

### When writing Go
- Package names: lowercase, single word (`router`, not `chatRouter`).
- Receiver names: 1-2 chars (`r *Router`, not `router *Router`).
- Error wrapping: `fmt.Errorf("operation failed: %w", err)`.
- Logging: `slog` only. `slog.InfoContext(ctx, "msg", "key", val)`.
- Tests: same package, `_test.go` suffix. Use `t.Run` for subtests.
- No `init()` functions for app logic — only for registering things to packages.
- Constants in SCREAMING_SNAKE only if they mirror a 9router constant. Else `CamelCase`.

### When editing the dashboard
- Stick to **App Router** + Server Components only where they make sense.
- All data fetching via TanStack Query.
- Auth: `httpOnly` cookie from `/api/auth/login`. Use middleware to redirect.
- API types: **generated** from Go's OpenAPI spec (do not hand-write).
- Components: shadcn/ui, Tailwind v4. No CSS-in-JS.
- Charts: `recharts` (matches 9router stack).

### Commit conventions
- Conventional commits: `feat(router): add combo fallback`.
- Reference parity: in the body, link the 9router file:
  ```
  Ports open-sse/services/accountFallback.js → internal/router/fallback.go.
  ```

### When stuck
- Re-read the relevant `docs/*.md`.
- Search `_ref/9router/` for the same problem.
- If the spec is wrong, **update the spec** and note it in the PR.

---

## 🧪 Testing Philosophy

| Layer | Tool | What to test |
|---|---|---|
| Unit | `testing` + `testify/assert` | Pure functions: translators, RTK filters, caveman |
| Integration | `httptest` + sqlite memory DB | API routes, auth, fallback chains |
| E2E | k6 + recorded fixtures | Real CLI tool → Go → mock upstream |
| Load | k6 / vegeta | SSE concurrency, memory under 5k streams |

**Golden files:** translator + RTK have `testdata/<case>/in.json` and
`testdata/<case>/out.json`. To update: `go test ./... -update`.

---

## 📚 Documents You MUST Read Before Coding

In order:

1. `PLAN.md` — what we're building and why.
2. `ROADMAP.md` — phased plan, **find your current phase**.
3. `docs/ARCHITECTURE.md` — folder structure, layering rules.
4. `docs/PARITY_CHECKLIST.md` — feature-by-feature status.
5. `docs/CONVENTIONS.md` — Go style, naming, anti-patterns.
6. Subsystem doc relevant to your task:
   - `docs/RTK_SPEC.md` — token saver
   - `docs/CAVEMAN_SPEC.md` — prompt injector
   - `docs/TRANSLATORS.md` — format conversion
   - `docs/EXECUTORS.md` — per-provider adapters
   - `docs/API.md` — endpoint contracts
   - `docs/DATABASE.md` — schema and migrations
   - `docs/AUTH.md` — JWT, API keys, OAuth flows

---

## 🛟 Quick Commands

```bash
# Dev
task dev              # run Go + Next.js with hot reload
task test             # full test suite
task lint             # golangci-lint + eslint

# Build
task build            # single binary with embedded frontend
task docker           # multi-stage docker build

# Codegen
task gen:openapi      # regenerate Go and TS types from spec
task gen:sqlc         # regenerate DB layer from SQL

# Parity verification
task parity:rtk       # run RTK fixtures captured from 9router
task parity:trans     # run translator fixtures
```

---

## 🚫 Anti-patterns Observed in 9router (do not copy)

1. `db.json` flat-file mutex contention — we use SQLite from day one.
2. `usageDb` stored separately from main DB ignoring `DATA_DIR` — unify under `DATA_DIR`.
3. Manual stream `.write()` without drain handling — Go `io.Copy` makes this free.
4. `better-sqlite3` fallback to `sql.js` due to build tools — pure-Go SQLite kills this.
5. Encryption-at-rest gap for OAuth tokens in db.json — encrypt with AES-GCM by default.
6. Static `/api/v1/route.js` model list duplicating `/v1/models` — single source of truth.

---

## 🧭 If You Are A New Agent Joining

1. Read this file in full.
2. Read `PLAN.md` and `ROADMAP.md`.
3. Open the file your task touches; read its top-of-file comment.
4. Check `git log --oneline -20` to see recent work.
5. Run `task test` to confirm baseline is green.
6. Then code.

---

**Last updated:** see git log.
**Maintainer:** Project owner. Anyone editing this file must update the
"Last updated" note in their commit message.
