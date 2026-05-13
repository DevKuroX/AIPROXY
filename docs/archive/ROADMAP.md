# ROADMAP.md — Phased Execution Plan

> Read `AGENTS.md` and `PLAN.md` first. This document is the **time-ordered
> work plan**. Each phase has a clear **exit criterion** — do not move on
> until it's green.

Status legend: `[ ] todo` · `[~] in progress` · `[x] done` · `[!] blocked`

Current phase: **Phase 0 — Foundation**.

---

## Phase 0 — Foundation (Week 1)

**Goal:** Working skeleton. Server boots. Frontend boots. Login works. CI passes.

### Backend
- [ ] `go mod init github.com/<you>/9rgo`
- [ ] Folder skeleton from `docs/ARCHITECTURE.md`
- [ ] `cmd/server/main.go` boots `chi` on `:20128`
- [ ] `internal/config` loads env from `.env` and OS env
- [ ] `internal/storage` connects to PostgreSQL via `pgx/v5/pgxpool`
- [ ] `internal/storage/migrations/001_init.sql` — users, settings, providers, api_keys
- [ ] `internal/auth/jwt.go` — sign/verify HS256 JWT
- [ ] `internal/api/middleware/{logger,recover,auth}.go`
- [ ] `POST /api/auth/login` (password check, returns `httpOnly` cookie)
- [ ] `POST /api/auth/logout`
- [ ] `GET  /api/auth/me`
- [ ] `GET  /api/health`
- [ ] `GET  /api/version`
- [ ] `slog` JSON handler wired
- [ ] Graceful shutdown on SIGTERM

### Frontend
- [ ] `web/` Next.js 15 App Router scaffold
- [ ] Tailwind v4 + shadcn/ui installed
- [ ] Zustand store boilerplate (mirror 9router stores)
- [ ] TanStack Query provider
- [ ] `/login` page hits `/api/auth/login`
- [ ] `/dashboard` empty shell with auth-gate middleware
- [ ] Dark/light theme toggle (matches 9router)

### Tooling
- [ ] `Taskfile.yml` with `dev`, `build`, `test`, `lint`
- [ ] `golangci-lint` config
- [ ] `air` config for Go hot reload
- [ ] `.editorconfig`, `.gitignore`
- [ ] GitHub Actions: lint + test on PR
- [ ] `Dockerfile` (multi-stage skeleton)

### Embed
- [ ] `assets/web/` populated by `next build && next export`
- [ ] `internal/api/static.go` serves `embed.FS` under `/dashboard/*`

**Exit criteria:**
1. `task dev` runs Go and Next.js concurrently.
2. Can `curl -X POST /api/auth/login` with valid creds and receive cookie.
3. `task build` produces a single binary that, when run, serves both `/api/health` and `/dashboard/`.

---

## Phase 1 — MVP Proxy: Single Provider, Native Path (Week 2-3)

**Goal:** Prove end-to-end SSE proxy with OpenAI. No translation yet.

### Backend
- [ ] `internal/models/provider.go` — `ProviderConnection`, `APIKey`, `Settings` types
- [ ] `internal/storage/providers.go` — CRUD via sqlc
- [ ] `internal/storage/keys.go`
- [ ] `internal/auth/apikey.go` — HMAC-signed API key generation + verify
- [ ] `internal/api/middleware/apikey.go` — gate `/v1/*`
- [ ] `internal/api/admin/providers.go` — CRUD + `/:id/test`
- [ ] `internal/api/admin/keys.go` — CRUD
- [ ] `internal/executor/executor.go` — `Executor` interface
- [ ] `internal/executor/default.go` — generic OpenAI-compatible upstream call
- [ ] `internal/stream/proxy.go` — SSE pass-through with `http.Flusher`
- [ ] `internal/router/handler.go` — `HandleChat` entry; resolve provider+account
- [ ] `internal/router/credentials.go` — pick account, mark unavailable
- [ ] `internal/api/v1/chat.go` — `POST /v1/chat/completions`
- [ ] `internal/api/v1/models.go` — `GET /v1/models`
- [ ] Error envelope (`internal/errs/`) matching OpenAI shape

### Frontend
- [ ] `/dashboard/providers` page — list, add API-key provider, test, delete
- [ ] `/dashboard/keys` page — generate, list, revoke client API keys
- [ ] `/dashboard/endpoint` page — show base URL + sample config snippets

### Tests
- [ ] Unit: API key sign/verify
- [ ] Integration: `POST /v1/chat/completions` → mock upstream → stream to client
- [ ] Load test (k6): 100 concurrent streams, 60s, <100 MB RAM

**Exit criteria:**
1. Add OpenAI provider via dashboard.
2. Generate an API key via dashboard.
3. Configure Claude Code (or any OpenAI-compatible CLI) against `http://localhost:20128/v1` with that key.
4. Run a query. SSE streams correctly. Usage row appears.

---

## Phase 2 — Translation Layer: OpenAI ↔ Claude (Week 4-5)

**Goal:** Client sends OpenAI format, upstream is Claude (Anthropic API). And vice versa.

### Backend
- [ ] `internal/translator/formats.go` — format constants + `DetectFormat`
- [ ] `internal/translator/detect.go` — body-shape detection
- [ ] `internal/translator/openai_to_claude.go` — request translator
- [ ] `internal/translator/claude_to_openai.go` — request translator (Claude client → OpenAI upstream)
- [ ] `internal/translator/stream/openai_to_claude.go` — SSE stream translator
- [ ] `internal/translator/stream/claude_to_openai.go` — SSE stream translator
- [ ] `internal/translator/index.go` — registry: source × target → translator pair
- [ ] `internal/router/handler.go` — wire translator into pipeline
- [ ] Native passthrough fast path (source == target)
- [ ] `internal/api/v1/messages.go` — `POST /v1/messages` (Claude endpoint)
- [ ] `internal/api/v1/count_tokens.go`

### Provider
- [ ] `internal/executor/claude.go` — Anthropic Messages API with headers (port `_ref/9router/open-sse/config/providers.js` Claude entry)

### Fixtures
- [ ] Capture request/response pairs from 9router → drop into `testdata/translator/`
- [ ] Golden tests for each translation direction

### Frontend
- [ ] Add API-key Claude provider via dashboard
- [ ] Settings: toggle native passthrough (debug)

**Exit criteria:**
1. Client sending OpenAI-shape request → routed to Anthropic → response translated back → indistinguishable from real OpenAI to the client.
2. All 4 directions (OpenAI↔Claude × stream/non-stream) pass golden tests.

---

## Phase 3 — Fallback, Combos, Multi-Account (Week 6)

**Goal:** 9router's headline reliability features.

### Backend
- [ ] `internal/models/combo.go` — combo schema
- [ ] `internal/storage/combos.go`
- [ ] `internal/router/fallback.go` — port `accountFallback.js` rules:
  - 401/403 → refresh, retry once
  - 429 / "rate limit" → mark account `rateLimitedUntil`
  - 5xx / network error → try next account
  - "exhausted quota" → mark account unavailable for N hours
- [ ] `internal/router/combo.go` — port `combo.js`: iterate models, propagate errors
- [ ] Round-robin & sticky round-robin strategy (per settings)
- [ ] Cooldown management with persistence
- [ ] Tests: simulate failures, assert next account/model called

### Frontend
- [ ] `/dashboard/combos` page — list, create, reorder, delete
- [ ] Multi-account UI on provider page
- [ ] Status badges: active / rate-limited / refresh-needed / exhausted

**Exit criteria:**
1. Create a combo `claude → glm → openai`.
2. Force Claude account to fail (revoke key). Request seamlessly falls to GLM.
3. Restore account → next request uses Claude again.

---

## Phase 4 — OAuth Providers: Claude + Gemini (Week 7-9)

**Goal:** First real OAuth flows. Token refresh stable under live traffic.

### Backend
- [ ] `internal/auth/oauth/` framework: `Flow` interface (`Start`, `Poll`, `Refresh`)
- [ ] `internal/auth/oauth/claude.go` — device code flow with claude.ai
- [ ] `internal/auth/oauth/gemini.go` — Google OAuth flow
- [ ] `internal/auth/oauth/storage.go` — encrypted token storage (AES-GCM)
- [ ] `internal/router/token_refresh.go` — pre-call refresh + 401 retry-after-refresh
- [ ] `internal/api/admin/oauth.go` — `/api/oauth/:provider/{start|poll}`

### Frontend
- [ ] OAuth flow UI: device code display, polling, success
- [ ] Token status display, manual refresh button

### Tests
- [ ] Refresh logic with mocked token server
- [ ] Concurrent-refresh dedup (only one refresh at a time per account)

**Exit criteria:**
1. Add Claude OAuth account via dashboard, sign in via device code.
2. Add Gemini OAuth account.
3. Long-running session: requests succeed across token expiry boundary (auto refresh).

### 🎯 **MVP MILESTONE** — at end of Phase 4 the product is daily-usable.

---

## Phase 5 — Specialized Executors (Week 10-12)

**Goal:** Port executors that have provider-specific quirks.

For each executor: read `_ref/9router/open-sse/executors/<name>.js`, port,
add request/response fixture tests.

- [ ] `codex.go` — OpenAI ChatGPT-CLI
- [ ] `cursor.go` — Cursor proprietary protocol (cursorChecksum, cursorProtobuf)
- [ ] `kiro.go` — Kiro headers + body format
- [ ] `github.go` — GitHub Copilot
- [ ] `gemini-cli.go` — Gemini CLI protocol
- [ ] `antigravity.go` — Antigravity request wrap
- [ ] `vertex.go` — Google Vertex AI
- [ ] `azure.go` — Azure OpenAI deployment routing
- [ ] `qwen.go`
- [ ] `iflow.go`
- [ ] `ollama-local.go`
- [ ] `opencode.go`, `opencode-go.go`
- [ ] `grok-web.go`, `perplexity-web.go` (web-scraping providers)
- [ ] `commandcode.go`
- [ ] `qoder.go`

### Translators added
- [ ] `openai_to_gemini` + `gemini_to_openai` (request + stream)
- [ ] `openai_to_kiro` + `kiro_to_openai`
- [ ] `openai_to_cursor` + `cursor_to_openai`
- [ ] `openai_to_commandcode` + `commandcode_to_openai`
- [ ] `openai_to_vertex`
- [ ] `openai_to_ollama` + `ollama_to_openai`
- [ ] `antigravity_to_openai` + `openai_to_antigravity`

**Exit criteria:**
1. At least 12 of the above executors callable from dashboard test.
2. Each has at least one golden fixture from 9router.

---

## Phase 6 — Token Savers: RTK + Caveman (Week 13)

**Goal:** Both savers at byte-for-byte parity with 9router.

### RTK
- [ ] `internal/rtk/constants.go` — port `constants.js`
- [ ] `internal/rtk/filters/` — one file per filter:
  - `gitdiff.go`, `gitstatus.go`, `gitlog.go`
  - `grep.go`, `find.go`, `ls.go`, `tree.go`
  - `dedup_log.go`, `smart_truncate.go`, `read_numbered.go`, `search_list.go`
- [ ] `internal/rtk/autodetect.go` — port `autodetect.js`
- [ ] `internal/rtk/apply.go` — port `applyFilter.js` (`safeApply`)
- [ ] `internal/rtk/compress.go` — port `index.js`: handles 4 message shapes:
  - OpenAI tool message (string)
  - OpenAI tool message (array of `text` parts)
  - Claude `tool_result` block (string content)
  - Claude `tool_result` block (array content)
  - OpenAI Responses `function_call_output` (string or array)
- [ ] Hook into translator: run **before** any format conversion
- [ ] Settings toggle + per-request override header
- [ ] Stats reporting (bytesBefore, bytesAfter, hits) → logs + usage row

### Caveman
- [ ] `internal/caveman/prompts.go` — port `cavemanPrompts.js` verbatim
- [ ] `internal/caveman/inject.go` — port `caveman.js`:
  - Dispatch by format (OpenAI / Claude / Gemini-family)
  - Append to existing system / developer message
  - Or prepend new system message
  - Handles `body.instructions` (OpenAI Responses string form)
- [ ] Three levels: `LITE`, `FULL`, `ULTRA`
- [ ] Settings toggle + level selector + per-request override

### Tests
- [ ] Fixture per filter: real-world `git diff`, `grep`, etc., before & after
- [ ] Caveman: golden fixtures for each level × each format

### Frontend
- [ ] Settings page section: RTK toggle, Caveman level dropdown
- [ ] Usage dashboard: show "tokens saved" metric per request

**Exit criteria:**
1. RTK fixture suite green (≥30 fixtures across all filters).
2. Caveman fixture suite green (3 levels × 4+ formats).
3. End-to-end test: Claude Code session with RTK+Caveman ON, output indistinguishable from OFF, but token usage measurably lower.

---

## Phase 7 — Usage Tracking, Pricing, Analytics (Week 14)

### Backend
- [ ] `internal/storage/usage.go` — append-only usage log
- [ ] `internal/router/usage.go` — extract usage from each response (each provider differs)
- [ ] `internal/storage/pricing.go` — pricing table per model
- [ ] `internal/api/admin/pricing.go` — CRUD overrides
- [ ] `internal/api/admin/usage.go`:
  - `/api/usage/summary?range=24h|7d|30d`
  - `/api/usage/recent?limit=...`
  - `/api/usage/by-provider`
  - `/api/usage/by-model`
- [ ] Request log file `~/.9rgo/log.txt` toggleable via `ENABLE_REQUEST_LOGS`

### Frontend
- [ ] `/dashboard` overview with recharts (tokens/day, by-provider pie)
- [ ] `/dashboard/usage` detailed table + filters
- [ ] `/dashboard/pricing` editor

**Exit criteria:**
1. Every request shows in usage table with prompt/completion tokens + cost.
2. Cost matches manual calc within 1% (rounding).

---

## Phase 8 — Provider Nodes, Aliases, CLI Helpers, Cloud Sync (Week 15-16)

### Provider nodes (custom compatible endpoints)
- [ ] Schema, CRUD, test
- [ ] Dashboard page

### Aliases
- [ ] `model_aliases` table
- [ ] Resolver in router: `myalias` → `claude-sonnet-4.5`
- [ ] Dashboard page

### CLI Helpers
- [ ] `/api/cli-tools/claude-code/config` — write `~/.claude/settings.json`
- [ ] `/api/cli-tools/codex/config`, `cursor`, `cline`, `continue`
- [ ] Dashboard buttons: "Configure Claude Code" → one click

### Cloud Sync
- [ ] `internal/lib/cloud_sync.go` — periodic sync via `NEXT_PUBLIC_CLOUD_URL`
- [ ] `/api/sync/cloud` enable/disable/sync actions
- [ ] Machine-ID handling (`node-machine-id` equivalent in Go)

**Exit criteria:**
1. Custom OpenAI-compatible node added and routable.
2. Aliases work in `/v1/chat/completions` model field.
3. "Configure Claude Code" button writes correct config to home dir.
4. Cloud sync toggles successfully (against a stub server if real one not available).

---

## Phase 9 — Hardening, Packaging, Docs (Week 17-18)

### Hardening
- [ ] `pprof` endpoint at `/debug/pprof` (auth-gated)
- [ ] Prometheus metrics at `/metrics` (auth-gated)
- [ ] Soak test: 5k concurrent streams for 30 min
- [ ] Memory profile: identify any leak / unbounded growth
- [ ] Panic recovery middleware coverage check

### Packaging
- [ ] Cross-platform binaries via `goreleaser`
- [ ] Multi-arch Docker image (linux/amd64 + linux/arm64)
- [ ] Homebrew tap (optional)
- [ ] `9rgo update` self-update (optional)

### Docs
- [ ] README with quickstart
- [ ] `docs/INSTALL.md`
- [ ] `docs/CONFIG.md` (env vars)
- [ ] `docs/PROVIDERS.md` (how to add each provider)
- [ ] `docs/MIGRATION_FROM_9ROUTER.md`

**Exit criteria:**
1. Binary <40 MB.
2. Docker image <60 MB.
3. 5k concurrent SSE for 30 min: no leaks, RAM <500 MB.
4. README quickstart works on a fresh Linux VM.

---

## Phase 10 — Optional Extras (post-1.0)

- [ ] Image generation endpoints
- [ ] TTS / STT endpoints
- [ ] Search endpoint
- [ ] Web fetch endpoint
- [ ] MITM proxy mode (port `src/mitm/`)
- [ ] Tunnel mode (cloudflared, port `src/lib/tunnel/`)
- [ ] Plugin system for community executors

---

## Phase Tracker

Use this checklist at the top of every PR description:

```
Phase: [N]
This PR completes:
- [ ] item
- [ ] item

Parity with 9router:
- [ ] feature X matches behavior of _ref/9router/path/to/file.js
- [ ] fixture added: testdata/foo/bar.json

Exit criterion(s) covered:
- [ ] criterion 1
```

---

## Time Estimates (solo dev, AI-assisted)

| Phase | Estimate | Hard? |
|---|---|---|
| 0 — Foundation | 1 week | Easy |
| 1 — MVP Proxy | 2 weeks | Medium |
| 2 — Translation | 2 weeks | Medium-Hard |
| 3 — Fallback | 1 week | Medium |
| 4 — OAuth | 3 weeks | Hard |
| 5 — Executors | 3 weeks | Medium-Hard (volume) |
| 6 — RTK + Caveman | 1 week | Medium |
| 7 — Usage | 1 week | Easy-Medium |
| 8 — Nodes/Aliases/CLI/Cloud | 2 weeks | Medium |
| 9 — Hardening | 2 weeks | Medium |
| **Total to v1.0** | **~18 weeks** | — |

MVP usable: **end of Phase 4 (~9 weeks)**.

---

**When you finish a phase:** update the "Current phase" line at the top of
this file AND update `AGENTS.md` § "Current Phase" with what's done and what's next.
