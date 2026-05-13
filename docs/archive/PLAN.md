# PLAN.md — 9Router Go Port

> **Owner reads first. Agents read second.**
> This document explains *what* and *why*. For *how-and-when*, see `ROADMAP.md`.
> For *where in the code*, see `docs/ARCHITECTURE.md`.

---

## 1. Vision

Build a **drop-in replacement** for 9router with:

- **Go backend** — single static binary, ~50–80 MB RAM at idle,
  ~150 MB at 1000 concurrent SSE streams.
- **Next.js dashboard** — same UX as 9router, modernized; statically exported
  and embedded into the Go binary.
- **100% functional parity** with 9router at the protocol/feature level.

Result: same product, ~5–10× cheaper to run, ~3× lower P99 latency, far easier
to operate and ship.

---

## 2. Scope (parity targets)

### 2.1 Public Gateway (`/v1/*` + `/v1beta/*`)

OpenAI-compatible surface used by CLI tools:

- `POST /v1/chat/completions` (OpenAI chat, streaming + non-streaming)
- `POST /v1/messages` (Anthropic Messages API)
- `POST /v1/responses` (OpenAI Responses API)
- `POST /v1/embeddings`
- `POST /v1/messages/count_tokens`
- `GET  /v1/models`
- `GET  /v1beta/models` (Gemini list)
- `POST /v1beta/models/{model}:generateContent`
- `POST /v1beta/models/{model}:streamGenerateContent`
- Image generation, TTS, STT, search, fetch (lower priority)

### 2.2 Admin / Dashboard API (`/api/*`)

- Auth: `/api/auth/login`, `/logout`, `/me`
- Settings: `/api/settings` (GET/PATCH)
- Providers: `/api/providers` CRUD + `/api/providers/:id/test`
- Provider nodes: `/api/provider-nodes` (custom OpenAI/Anthropic-compatible nodes)
- OAuth: `/api/oauth/:provider/{start|poll|callback}`
- API keys: `/api/keys` (keys clients use to call `/v1/*`)
- Model aliases: `/api/models/alias`
- Combos: `/api/combos`
- Pricing: `/api/pricing`
- Usage: `/api/usage/{summary|recent|by-provider|by-model}`
- Sync/cloud: `/api/sync/cloud` (optional cloud sync)
- CLI tools helpers: `/api/cli-tools/*` (write Claude Code/Codex configs)
- Health, version, shutdown, init

### 2.3 Core Engine (parity targets)

| Subsystem | 9router source | Spec doc |
|---|---|---|
| Format detection | `open-sse/translator/formats.js` | `docs/TRANSLATORS.md` |
| Request translators | `open-sse/translator/request/*` | `docs/TRANSLATORS.md` |
| Response translators | `open-sse/translator/response/*` | `docs/TRANSLATORS.md` |
| Per-provider executors | `open-sse/executors/*` (20+) | `docs/EXECUTORS.md` |
| Account fallback | `open-sse/services/accountFallback.js` | `docs/FALLBACK.md` |
| Combo fallback | `open-sse/services/combo.js` | `docs/FALLBACK.md` |
| Token refresh | `open-sse/services/tokenRefresh.js` | `docs/AUTH.md` |
| RTK token saver | `open-sse/rtk/*` (11 filters) | `docs/RTK_SPEC.md` |
| Caveman prompt injector | `open-sse/rtk/caveman.js`, `cavemanPrompts.js` | `docs/CAVEMAN_SPEC.md` |
| Native passthrough | `open-sse/handlers/chatCore.js` | `docs/TRANSLATORS.md` |
| Stream handler | `open-sse/utils/stream*.js` | `docs/STREAMING.md` |
| Usage tracking | `open-sse/utils/usageTracking.js` | `docs/USAGE.md` |
| Request logging | `open-sse/utils/requestLogger.js` | `docs/OBSERVABILITY.md` |

### 2.4 Provider Coverage (must match 9router)

**OAuth providers:** Claude, Codex (OpenAI ChatGPT-CLI), Gemini, Qwen, iFlow,
GitHub (Copilot), Kiro, Cursor, Antigravity, OpenCode.

**API-key providers:** OpenAI, Anthropic, OpenRouter, GLM (Zhipu), Kimi (Moonshot),
MiniMax, Mistral, Groq, DeepSeek, Together, Fireworks, Cerebras, SambaNova,
Azure OpenAI, Vertex AI, Perplexity, xAI/Grok, MiMo (Xiaomi), Cloudflare Workers AI,
Ollama-local, and the rest of 9router's list (see `_ref/9router/open-sse/config/providers.js`).

**Specialized executors:** `antigravity`, `azure`, `codex`, `commandcode`,
`cursor`, `gemini-cli`, `github`, `grok-web`, `iflow`, `kiro`, `ollama-local`,
`opencode`, `opencode-go`, `perplexity-web`, `qoder`, `qwen`, `vertex`,
plus `default` for the rest.

### 2.5 Frontend Pages (parity)

- Login
- Dashboard (overview, charts, recent activity)
- Providers (list, add OAuth, add API key, test, multi-account)
- Provider Nodes (custom compatible endpoints)
- API Keys (for client CLI tools)
- Combos (fallback sequences)
- Aliases (model rename rules)
- Pricing (custom price overrides)
- Usage (charts, by provider/model/day)
- Endpoint (the `http://localhost:20128/v1` config + CLI tool helpers)
- Settings (cloud sync, RTK toggle, Caveman level, password, dark mode)
- Logs viewer (when `ENABLE_REQUEST_LOGS=true`)

### 2.6 Token Savers (the headline feature)

We support **two complementary** token savers:

1. **RTK (Request Tool-result Kompressor)** — losslessly compresses tool
   output content in the *request body* before it goes upstream. Filters
   detect `git diff`, `git status`, `git log`, `grep`, `find`, `ls`, `tree`,
   numbered-file reads, search lists, and generic deduplication / smart
   truncation. See `docs/RTK_SPEC.md`.
2. **Caveman** — injects a terse-style system prompt (`LITE` / `FULL` /
   `ULTRA`) so the model *generates* fewer output tokens. Three intensity
   levels. See `docs/CAVEMAN_SPEC.md`.

Both run server-side, both are toggleable per user/request, both run on
*every* request that flows through the gateway.

### 2.7 Non-functional Requirements

- **Cross-platform binary**: Linux (amd64, arm64), macOS (amd64, arm64),
  Windows (amd64).
- **Cold start**: <100 ms to ready.
- **Memory**: ≤80 MB at idle, ≤200 MB at 1000 concurrent SSE.
- **Stability**: handle 10k+ concurrent streams without OOM on a 2-vCPU/2 GB host.
- **Update**: `9rgo update` self-update flow (optional; mirror 9router npm update).

---

## 3. Out of Scope (initial release)

- Implementing the cloud-sync *server* side (we only call it).
- Mobile clients.
- Provider-specific UIs beyond what 9router has.
- An alternate "config-only" file mode (DB is source of truth).

---

## 4. Architectural Decisions (ADRs)

### ADR-001: Go for backend
**Decision:** Use Go.
**Why:** SSE proxy concurrency, single-binary deploy, low memory, predictable
under load. See chat transcript at project kickoff for full analysis.

### ADR-002: SQLite (pure-Go) for persistence
**Decision:** `modernc.org/sqlite` (no CGO).
**Why:** Cross-compile sanity, avoids 9router's `better-sqlite3` → `sql.js`
fallback pain. SQLite gives us proper transactions, indexes, migrations.

### ADR-003: chi over Gin/Fiber
**Decision:** `go-chi/chi/v5`.
**Why:** Closest to `net/http`. Middleware composition is idiomatic. No
custom context abstraction. Fiber's fasthttp doesn't play well with `http.Flusher`.

### ADR-004: slog (stdlib) for logging
**Decision:** `log/slog` since Go 1.21+.
**Why:** Structured logging without adding deps. Plays nicely with handlers
for JSON/text/OTel.

### ADR-005: sqlc for DB layer
**Decision:** `sqlc` generates typed Go from SQL.
**Why:** Raw SQL (we want full control of queries), no ORM lock-in, type-safe
at compile time.

### ADR-006: Next.js (App Router) for frontend
**Decision:** Next.js 15 with `output: 'export'`.
**Why:** Mirrors 9router's React + Zustand + Tailwind stack, lets us reuse
patterns. Static export → embed in Go binary.

### ADR-007: Frontend ↔ Backend contract via OpenAPI
**Decision:** Maintain `docs/openapi.yaml` as the source of truth.
**Why:** Generate Go server stubs (optional) and TS types automatically.
Prevents drift between client and server expectations.

### ADR-008: RTK filters as pure functions
**Decision:** Each RTK filter is a `func(string) string` with no side effects.
**Why:** Trivially testable with golden files. Mirrors the 9router
implementation faithfully.

### ADR-009: Caveman as a pluggable injector
**Decision:** Caveman injection is a middleware-style step inside the
translator pipeline, dispatched by output format.
**Why:** Same 9router pattern. Allows future "personas" beyond Caveman
without touching translators.

### ADR-010: Auth tokens encrypted at rest
**Decision:** AES-GCM with key derived from `API_KEY_SECRET` env.
**Why:** 9router stores OAuth tokens in plaintext `db.json` — a known gap.
We fix it from day one.

---

## 5. Major Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| 9router translator edge cases missed | High | High | Capture **golden fixtures** from a running 9router instance; CI fails on drift. |
| OAuth flow per-provider drift | High | Med | Implement 2–3 providers fully (Claude, Gemini) before tackling rest. Centralize device-code template. |
| CLI tools (Claude Code, Cursor) change formats | Med | Med | Pin tool versions in `docs/CLI_TOOLS.md`. Subscribe to release notes. |
| SQLite migration mistakes | Med | High | All migrations are forward-only, numbered, run on startup, with backup-before-apply. |
| Cross-platform build breaks | Low | Med | CI matrix builds linux/mac/windows on every PR. |
| Solo-dev burnout | Med | High | Strict phase gates (see `ROADMAP.md`). Skip non-essential providers until MVP usable. |
| RTK behavioral parity gaps | Med | Med | Per-filter fixtures; soak-test with real `git diff`/`grep` outputs from this repo. |
| Caveman prompt mutation breaks tools | Low | Med | Toggle on user, levels off by default. Document side effects. |

---

## 6. Success Criteria

### MVP (end of Phase 4):
- [ ] Claude Code can be configured against the Go gateway and complete a coding session end-to-end (OpenAI ↔ Claude translation working).
- [ ] At least **3 providers** working: OpenAI, Anthropic (Claude OAuth), Gemini.
- [ ] Dashboard can: log in, add provider, generate API key, see usage.
- [ ] Streaming p99 latency overhead ≤ 25 ms vs. direct provider call.
- [ ] 1000 concurrent SSE streams stable for 10 minutes under <200 MB RAM.

### v1.0 (end of Phase 8):
- [ ] All 20+ executors ported.
- [ ] All RTK filters at parity (CI passes against golden fixtures).
- [ ] Caveman 3 levels working across all output formats.
- [ ] All admin endpoints reachable from dashboard.
- [ ] Docker image <60 MB.
- [ ] Cross-platform binaries published.

---

## 7. Team & Process

- **Solo dev** running via OpenCode / Claude Code AI assistance.
- AI agents follow `AGENTS.md` rules.
- One feature branch per phase: `phase/01-foundation`, `phase/02-mvp-proxy`, …
- PRs merged into `main` only when phase exit criteria met (see ROADMAP).
- Tag versions: `v0.<phase>.<patch>`.

---

## 8. Glossary

- **Executor** — per-provider adapter that knows how to make the upstream call.
- **Translator** — converts request/response shape between formats (OpenAI ↔ Claude ↔ Gemini ↔ …).
- **Combo** — an ordered sequence of model names that fall back on failure.
- **Account** — a single credential (OAuth account or API key) within a provider.
- **Provider connection** — an account record in DB with provider + credentials + state.
- **Native passthrough** — special case where source and target format match, so no translation runs.
- **RTK** — token saver that compresses tool-result content in the request.
- **Caveman** — token saver that injects a brevity instruction into the system prompt.
- **9router** — the upstream reference project we are porting.
- **9rgo** — working name of this port (rename in your fork).
