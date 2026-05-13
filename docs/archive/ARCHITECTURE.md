
> **Setup-pack corrections applied to this file:**
> - Removed phantom `executor/claude.go` (Claude uses `DefaultExecutor` after translation).
> - Removed phantom `auth/oauth/opencode.go` (opencode has no OAuth flow in 9router).
> - Added missing `auth/oauth/qoder.go` and `auth/oauth/openai.go`.
> - RTK `git-log` is declared but not implemented; not part of the port.

# docs/ARCHITECTURE.md — Canonical Folder Tree & Layering

## Top-level layout

```
9rgo/                                  # rename to your project
├── cmd/
│   └── server/
│       └── main.go                    # entry point — wires everything
│
├── internal/                          # private; not importable externally
│   ├── api/                           # HTTP layer (handlers only — no business logic)
│   │   ├── v1/                        # OpenAI-compatible public API
│   │   │   ├── chat.go                # POST /v1/chat/completions
│   │   │   ├── messages.go            # POST /v1/messages (Claude)
│   │   │   ├── responses.go           # POST /v1/responses (OpenAI Responses)
│   │   │   ├── models.go              # GET  /v1/models
│   │   │   ├── count_tokens.go        # POST /v1/messages/count_tokens
│   │   │   ├── embeddings.go
│   │   │   ├── images.go              # phase 10
│   │   │   ├── audio.go               # tts / stt — phase 10
│   │   │   └── routes.go              # mount group on chi router
│   │   ├── v1beta/                    # Gemini-compatible surface
│   │   │   ├── models.go
│   │   │   ├── generate.go            # :generateContent + :streamGenerateContent
│   │   │   └── routes.go
│   │   ├── admin/                     # dashboard / management API
│   │   │   ├── auth.go                # /api/auth/{login,logout,me}
│   │   │   ├── settings.go            # /api/settings
│   │   │   ├── providers.go           # /api/providers + /:id/test
│   │   │   ├── provider_nodes.go
│   │   │   ├── oauth.go               # /api/oauth/:provider/...
│   │   │   ├── keys.go                # /api/keys
│   │   │   ├── aliases.go             # /api/models/alias
│   │   │   ├── combos.go              # /api/combos
│   │   │   ├── pricing.go
│   │   │   ├── usage.go               # /api/usage/*
│   │   │   ├── sync.go                # /api/sync/cloud
│   │   │   ├── cli_tools.go           # /api/cli-tools/*
│   │   │   ├── health.go
│   │   │   ├── version.go
│   │   │   ├── shutdown.go
│   │   │   └── routes.go
│   │   ├── middleware/
│   │   │   ├── recover.go
│   │   │   ├── logger.go              # slog request log
│   │   │   ├── cors.go
│   │   │   ├── auth_jwt.go            # dashboard cookie auth
│   │   │   ├── auth_apikey.go         # gates /v1/*
│   │   │   └── ratelimit.go           # optional
│   │   └── static.go                  # serves embedded Next.js export
│   │
│   ├── router/                        # core orchestration (port of src/sse/* + open-sse/handlers)
│   │   ├── handler.go                 # HandleChat — entry
│   │   ├── modelresolver.go           # parse "provider/model" or "combo:name"
│   │   ├── credentials.go             # pick account, mark unavailable, refresh hook
│   │   ├── fallback.go                # account-level retry rules
│   │   ├── combo.go                   # combo-level model iteration
│   │   ├── token_refresh.go           # pre-call + 401 retry-after-refresh
│   │   ├── passthrough.go             # native passthrough decision
│   │   ├── pipeline.go                # request → RTK → translate → Caveman → execute → translate response
│   │   └── usage.go                   # extract usage from response, persist
│   │
│   ├── translator/                    # format conversion (port of open-sse/translator)
│   │   ├── formats.go                 # FORMATS enum + DetectByEndpoint
│   │   ├── detect.go                  # body-shape detection
│   │   ├── index.go                   # registry: (src, dst) → translator pair
│   │   ├── request/
│   │   │   ├── openai_to_claude.go
│   │   │   ├── claude_to_openai.go
│   │   │   ├── openai_to_gemini.go
│   │   │   ├── gemini_to_openai.go
│   │   │   ├── openai_to_kiro.go
│   │   │   ├── openai_to_cursor.go
│   │   │   ├── openai_to_commandcode.go
│   │   │   ├── openai_to_vertex.go
│   │   │   ├── openai_to_ollama.go
│   │   │   ├── antigravity_to_openai.go
│   │   │   ├── openai_responses.go    # responses request handler
│   │   │   └── helpers.go
│   │   ├── response/
│   │   │   ├── claude_to_openai.go
│   │   │   ├── openai_to_claude.go
│   │   │   ├── gemini_to_openai.go
│   │   │   ├── kiro_to_openai.go
│   │   │   ├── cursor_to_openai.go
│   │   │   ├── commandcode_to_openai.go
│   │   │   ├── ollama_to_openai.go
│   │   │   ├── openai_to_antigravity.go
│   │   │   ├── openai_responses.go
│   │   │   └── helpers.go
│   │   └── stream/                    # SSE stream transformers
│   │       ├── openai_to_claude.go
│   │       ├── claude_to_openai.go
│   │       ├── ... (mirror request/response)
│   │
│   ├── executor/                      # per-provider HTTP adapters
│   │   ├── executor.go                # Executor interface
│   │   ├── base.go                    # shared helpers (port of executors/base.js)
│   │   ├── default.go                 # generic OpenAI-compatible
│   │   ├── codex.go
│   │   ├── cursor.go                  # uses cursorChecksum + protobuf
│   │   ├── kiro.go
│   │   ├── github.go                  # Copilot
│   │   ├── gemini_cli.go
│   │   ├── antigravity.go
│   │   ├── vertex.go
│   │   ├── azure.go
│   │   ├── qwen.go
│   │   ├── iflow.go
│   │   ├── ollama_local.go
│   │   ├── opencode.go
│   │   ├── opencode_go.go
│   │   ├── grok_web.go
│   │   ├── perplexity_web.go
│   │   ├── commandcode.go
│   │   ├── qoder.go
│   │   └── index.go                   # registry: providerID → Executor
│   │
│   ├── rtk/                           # RTK token saver (port of open-sse/rtk)
│   │   ├── constants.go               # caps, thresholds
│   │   ├── compress.go                # entry: CompressMessages
│   │   ├── autodetect.go              # pick filter by content shape
│   │   ├── apply.go                   # safeApply (try, fallback to identity on panic)
│   │   ├── shapes.go                  # message-shape helpers
│   │   ├── filters/
│   │   │   ├── gitdiff.go
│   │   │   ├── gitstatus.go
│   │   │   ├── gitlog.go
│   │   │   ├── grep.go
│   │   │   ├── find.go
│   │   │   ├── ls.go
│   │   │   ├── tree.go
│   │   │   ├── dedup_log.go
│   │   │   ├── smart_truncate.go
│   │   │   ├── read_numbered.go
│   │   │   └── search_list.go
│   │   └── registry.go                # name → filter func
│   │
│   ├── caveman/                       # caveman prompt injector (port of open-sse/rtk/caveman.js)
│   │   ├── prompts.go                 # LITE / FULL / ULTRA strings (verbatim)
│   │   ├── inject.go                  # InjectCaveman(body, format, level)
│   │   └── inject_test.go
│   │
│   ├── auth/                          # all auth concerns
│   │   ├── jwt.go                     # dashboard JWT (HS256)
│   │   ├── password.go                # bcrypt
│   │   ├── apikey.go                  # HMAC-signed API key for /v1
│   │   ├── crypto.go                  # AES-GCM for tokens-at-rest
│   │   ├── machine.go                 # machine ID (cloud sync identity)
│   │   └── oauth/
│   │       ├── flow.go                # Flow interface
│   │       ├── storage.go             # encrypted token helpers
│   │       ├── claude.go              # device code
│   │       ├── gemini.go              # OAuth 2.0
│   │       ├── codex.go
│   │       ├── github.go
│   │       ├── kiro.go
│   │       ├── cursor.go
│   │       ├── antigravity.go
│   │       ├── qwen.go
│   │       ├── iflow.go
│   │       ├── qoder.go
│       ├── openai.go
│       └── (opencode is an executor only — no OAuth flow in 9router)
│   │
│   ├── storage/                       # persistence layer
│   │   ├── db.go                      # *sql.DB setup, ping, migrate
│   │   ├── tx.go                      # transaction helper
│   │   ├── settings.go                # generated by sqlc
│   │   ├── providers.go
│   │   ├── provider_nodes.go
│   │   ├── keys.go
│   │   ├── aliases.go
│   │   ├── combos.go
│   │   ├── pricing.go
│   │   ├── usage.go
│   │   ├── queries/                   # raw SQL files for sqlc
│   │   │   ├── settings.sql
│   │   │   ├── providers.sql
│   │   │   └── ...
│   │   └── migrations/
│   │       ├── 001_init.up.sql
│   │       ├── 001_init.down.sql
│   │       └── ...
│   │
│   ├── models/                        # plain domain types (no behavior)
│   │   ├── provider.go
│   │   ├── provider_node.go
│   │   ├── key.go
│   │   ├── alias.go
│   │   ├── combo.go
│   │   ├── pricing.go
│   │   ├── usage.go
│   │   ├── settings.go
│   │   └── format.go
│   │
│   ├── stream/                        # SSE primitives
│   │   ├── flusher.go                 # FlushWriter wraps http.ResponseWriter
│   │   ├── scanner.go                 # SSE chunk scanner (data: ...\n\n)
│   │   ├── encoder.go                 # write SSE events
│   │   ├── proxy.go                   # io.Copy with flush
│   │   └── disconnect.go              # client-disconnect detector
│   │
│   ├── errs/                          # typed errors
│   │   ├── errs.go                    # Error type + categories
│   │   ├── http.go                    # Error → HTTP status
│   │   └── openai.go                  # marshalling to OpenAI error envelope
│   │
│   ├── observability/
│   │   ├── logger.go                  # slog setup
│   │   ├── request_logger.go          # port of utils/requestLogger.js
│   │   ├── metrics.go                 # Prometheus (phase 9)
│   │   └── pprof.go                   # phase 9
│   │
│   ├── cloud/
│   │   ├── sync.go                    # periodic upload/download
│   │   ├── scheduler.go               # uses time.Ticker, cancellable
│   │   └── client.go                  # HTTP client to cloud endpoint
│   │
│   ├── cli/                           # helpers that write CLI tool configs
│   │   ├── claude_code.go
│   │   ├── codex.go
│   │   ├── cursor.go
│   │   ├── cline.go
│   │   └── continue.go
│   │
│   ├── config/
│   │   ├── config.go                  # struct + env loader
│   │   └── defaults.go
│   │
│   └── version/
│       └── version.go                 # set via ldflags at build
│
├── pkg/                               # rarely used; only for actually-public APIs
│   └── (empty for now)
│
├── web/                               # Next.js dashboard (sub-project)
│   ├── package.json
│   ├── next.config.mjs                # output: 'export'
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   ├── app/
│   │   ├── (auth)/login/page.tsx
│   │   ├── (dashboard)/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx               # overview
│   │   │   ├── providers/page.tsx
│   │   │   ├── provider-nodes/page.tsx
│   │   │   ├── keys/page.tsx
│   │   │   ├── combos/page.tsx
│   │   │   ├── aliases/page.tsx
│   │   │   ├── pricing/page.tsx
│   │   │   ├── usage/page.tsx
│   │   │   ├── endpoint/page.tsx
│   │   │   ├── logs/page.tsx
│   │   │   └── settings/page.tsx
│   │   ├── api/                       # NEVER use Next.js API routes (Go is the API)
│   │   ├── globals.css
│   │   └── layout.tsx
│   ├── components/
│   │   ├── ui/                        # shadcn/ui generated
│   │   ├── charts/
│   │   ├── providers/
│   │   ├── auth/
│   │   └── shared/
│   ├── lib/
│   │   ├── api.ts                     # fetch wrapper (handles auth cookie)
│   │   ├── types.gen.ts               # GENERATED from openapi.yaml
│   │   └── utils.ts
│   ├── stores/                        # Zustand
│   │   ├── auth.ts
│   │   ├── providers.ts
│   │   ├── settings.ts
│   │   └── theme.ts
│   ├── hooks/                         # TanStack Query hooks
│   ├── public/
│   └── out/                           # build output (gitignored)
│
├── assets/
│   └── web/                           # copied from web/out/ at build time
│
├── testdata/                          # fixtures captured from 9router
│   ├── translator/
│   │   ├── openai_to_claude/
│   │   │   ├── basic.in.json
│   │   │   └── basic.out.json
│   │   └── ...
│   ├── rtk/
│   │   ├── gitdiff/
│   │   ├── grep/
│   │   └── ...
│   └── caveman/
│       ├── lite/
│       └── ...
│
├── tests/
│   ├── integration/
│   │   ├── chat_test.go
│   │   ├── auth_test.go
│   │   ├── fallback_test.go
│   │   └── helpers.go
│   ├── load/
│   │   ├── sse_concurrent.js          # k6
│   │   └── soak.js
│   └── e2e/
│       └── claude_code_test.go        # spawns real CLI tool against gateway
│
├── scripts/
│   ├── capture_fixtures.sh            # run 9router and capture request/response pairs
│   ├── build.sh
│   ├── dev.sh
│   ├── gen_openapi.sh
│   └── parity_diff.sh
│
├── docs/                              # YOU ARE READING THESE
│   ├── ARCHITECTURE.md                # ← you are here
│   ├── PARITY_CHECKLIST.md
│   ├── CONVENTIONS.md
│   ├── RTK_SPEC.md
│   ├── CAVEMAN_SPEC.md
│   ├── TRANSLATORS.md
│   ├── EXECUTORS.md
│   ├── API.md
│   ├── DATABASE.md
│   ├── AUTH.md
│   ├── FALLBACK.md
│   ├── STREAMING.md
│   ├── USAGE.md
│   ├── OBSERVABILITY.md
│   ├── CLI_TOOLS.md
│   ├── FRONTEND.md
│   └── openapi.yaml                   # API contract — source of truth
│
├── _ref/
│   └── 9router/                       # full 9router checkout — READ ONLY
│
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
│
├── .opencode/
│   └── rules.md                       # mirror of AGENTS.md (or symlink)
├── .claude/
│   └── CLAUDE.md                      # mirror of AGENTS.md (or symlink)
├── .cursorrules                       # mirror of AGENTS.md (or symlink)
│
├── go.mod
├── go.sum
├── Taskfile.yml
├── Dockerfile
├── .env.example
├── .editorconfig
├── .gitignore
├── .golangci.yml
├── AGENTS.md
├── PLAN.md
├── ROADMAP.md
├── LICENSE
└── README.md
```

---

## Layering rules

Strictly enforced — see `internal/` layering:

```
api/v1, api/v1beta, api/admin
        │
        ▼
     router/         ← orchestration: pipeline, fallback, combo
        │
        ├──► translator/    (request shape → upstream shape)
        ├──► rtk/           (mutates request body in place)
        ├──► caveman/       (injects into system prompt)
        ├──► executor/      (network call)
        ├──► auth/          (creds + refresh)
        └──► storage/       (read providers/accounts; write usage)

storage/   ◄──── models/
   │
   └─► sqlite db file
```

**Allowed imports:**
- `api/*` may import `router`, `models`, `errs`, `auth`, `config`, `observability`.
- `router/*` may import `translator`, `rtk`, `caveman`, `executor`, `auth`, `storage`, `stream`, `models`, `errs`.
- `executor/*` may import `models`, `stream`, `auth`, `errs`.
- `translator/*` may import `models`, `errs`. **No HTTP, no DB.**
- `rtk/*` and `caveman/*`: **pure functions only**. No imports from elsewhere except stdlib.
- `models/*`: stdlib only.

**Disallowed:**
- `internal/*` may not import from `cmd/*` or `web/*`.
- `models/*` may not import from anywhere else inside the repo.
- Circular imports: forbidden (enforced by `golangci-lint`).

---

## Build modes

### Dev mode (`task dev`)
- Go runs on `:20128` via `air`
- Next.js runs on `:3000` via `next dev`
- Next config has `rewrites` to proxy `/api/*` and `/v1/*` to `:20128`
- Frontend NOT embedded; Go serves a placeholder "dev mode" page at `/dashboard/*`

### Build mode (`task build`)
1. `cd web && npm run build` → produces `web/out/`
2. `cp -r web/out/* assets/web/`
3. `go build -o bin/9rgo ./cmd/server`
4. Single binary serves both `/v1/*`, `/api/*`, `/dashboard/*`

### Docker mode (`task docker`)
- Multi-stage:
  - Stage 1: `node:20-alpine` → build frontend
  - Stage 2: `golang:1.23-alpine` → build Go with embedded frontend
  - Stage 3: `gcr.io/distroless/static-debian12` → copy binary
- Final image: ~30-40 MB

---

## Why this structure works

- **`internal/`** keeps everything private — nobody can import internals from outside.
- **`models/` standalone** prevents accidental DB/HTTP imports in domain logic.
- **`rtk/` and `caveman/` are leaf packages** — pure, testable, swappable.
- **`router/` orchestrates** but contains no provider-specific logic.
- **`executor/` is provider-specific** — additions/removals don't touch router.
- **`translator/` is format-specific** — additions don't touch router or executor.
- **API handlers are thin** — they parse, call router, write response. No logic.

This is the **opposite of 9router's structure**, which mixes concerns across
`src/sse/` and `open-sse/`. We split the concerns cleanly so each subsystem
can evolve independently.
