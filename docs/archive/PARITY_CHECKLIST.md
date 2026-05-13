# docs/PARITY_CHECKLIST.md — Feature-by-Feature Parity Tracker

> Living document. Tick `[x]` only when a feature has:
> 1. Code in `internal/...`
> 2. Test (unit or integration) referencing a 9router fixture or behavior
> 3. A `// ref: <9router path>` comment in the Go source

---

## Legend
- ✅  done & verified against 9router
- 🟡  in progress
- ⬜  not started
- ⚠️  intentional divergence (document in `docs/DIVERGENCES.md`)

---

## 1. Public Gateway (`/v1/*` + `/v1beta/*`)

| Endpoint | 9router ref | Go file | Status |
|---|---|---|---|
| `POST /v1/chat/completions` | `open-sse/handlers/chatCore.js` | `internal/api/v1/chat.go` | ⬜ |
| `POST /v1/messages` | `open-sse/handlers/chatCore.js` | `internal/api/v1/messages.go` | ⬜ |
| `POST /v1/responses` | `open-sse/handlers/responsesHandler.js` | `internal/api/v1/responses.go` | ⬜ |
| `GET  /v1/models` | `open-sse/handlers/...` | `internal/api/v1/models.go` | ⬜ |
| `POST /v1/messages/count_tokens` | `open-sse/handlers/...` | `internal/api/v1/count_tokens.go` | ⬜ |
| `POST /v1/embeddings` | `open-sse/handlers/embeddingsCore.js` | `internal/api/v1/embeddings.go` | ⬜ |
| `POST /v1/images/generations` | `open-sse/handlers/imageGenerationCore.js` | `internal/api/v1/images.go` | ⬜ |
| `POST /v1/audio/speech` | `open-sse/handlers/ttsCore.js` | `internal/api/v1/audio.go` | ⬜ |
| `POST /v1/audio/transcriptions` | `open-sse/handlers/sttCore.js` | `internal/api/v1/audio.go` | ⬜ |
| `POST /v1beta/models/:model:generateContent` | gemini path | `internal/api/v1beta/generate.go` | ⬜ |
| `POST /v1beta/models/:model:streamGenerateContent` | gemini path | `internal/api/v1beta/generate.go` | ⬜ |
| `GET  /v1beta/models` | gemini list | `internal/api/v1beta/models.go` | ⬜ |

---

## 2. Admin / Dashboard API (`/api/*`)

| Endpoint | Status |
|---|---|
| `POST /api/auth/login` | ⬜ |
| `POST /api/auth/logout` | ⬜ |
| `GET  /api/auth/me` | ⬜ |
| `GET/PUT /api/settings` | ⬜ |
| `CRUD /api/providers` + `POST /:id/test` | ⬜ |
| `CRUD /api/providers/:id/nodes` | ⬜ |
| `CRUD /api/keys` | ⬜ |
| `CRUD /api/models/alias` | ⬜ |
| `CRUD /api/combos` | ⬜ |
| `CRUD /api/pricing` | ⬜ |
| `GET  /api/usage/*` | ⬜ |
| `POST /api/oauth/:provider/start` + callbacks | ⬜ |
| `POST /api/sync/cloud` | ⬜ |
| `GET  /api/cli-tools/*` | ⬜ |
| `GET  /api/health`, `/api/version` | ⬜ |
| `POST /api/shutdown` | ⬜ |

---

## 3. Translators

| Pair | Direction | 9router ref | Status |
|---|---|---|---|
| OpenAI ↔ Claude | request | `open-sse/translator/request/openaiToClaude.js` | ⬜ |
| OpenAI ↔ Claude | response | `open-sse/translator/response/claudeToOpenai.js` | ⬜ |
| OpenAI ↔ Claude | stream | `open-sse/translator/stream/*` | ⬜ |
| OpenAI ↔ Gemini | request | translator | ⬜ |
| OpenAI ↔ Gemini | response | translator | ⬜ |
| OpenAI ↔ Gemini | stream | translator | ⬜ |
| OpenAI → Kiro | request | translator | ⬜ |
| Kiro → OpenAI | response | translator | ⬜ |
| OpenAI → Cursor | request | translator | ⬜ |
| Cursor → OpenAI | response | translator | ⬜ |
| OpenAI → CommandCode | request | translator | ⬜ |
| CommandCode → OpenAI | response | translator | ⬜ |
| OpenAI → Vertex | request | translator | ⬜ |
| OpenAI → Ollama | request | translator | ⬜ |
| Antigravity → OpenAI | response | translator | ⬜ |
| OpenAI Responses | request/response | `open-sse/transformer/responsesTransformer.js` | ⬜ |

---

## 4. Executors (per-provider HTTP adapters)

Verified against `open-sse/executors/index.js` registry:

| Provider key | Go file | Specialized? | Status |
|---|---|---|---|
| `default` (generic OpenAI-compat) | `executor/default.go` | base | ⬜ |
| `antigravity` | `executor/antigravity.go` | yes | ⬜ |
| `azure` | `executor/azure.go` | yes | ⬜ |
| `codex` | `executor/codex.go` | yes | ⬜ |
| `commandcode` | `executor/commandcode.go` | yes | ⬜ |
| `cursor` (+ alias `cu`) | `executor/cursor.go` | yes (cursorChecksum + protobuf) | ⬜ |
| `gemini-cli` | `executor/gemini_cli.go` | yes | ⬜ |
| `github` (Copilot) | `executor/github.go` | yes | ⬜ |
| `grok-web` | `executor/grok_web.go` | yes | ⬜ |
| `iflow` | `executor/iflow.go` | yes | ⬜ |
| `kiro` | `executor/kiro.go` | yes | ⬜ |
| `ollama-local` | `executor/ollama_local.go` | yes | ⬜ |
| `opencode` | `executor/opencode.go` | yes | ⬜ |
| `opencode-go` | `executor/opencode_go.go` | yes | ⬜ |
| `perplexity-web` | `executor/perplexity_web.go` | yes | ⬜ |
| `qoder` | `executor/qoder.go` | yes | ⬜ |
| `qwen` | `executor/qwen.go` | yes | ⬜ |
| `vertex` (+ `vertex-partner`) | `executor/vertex.go` | yes | ⬜ |

> NOTE: `claude.go` is NOT a standalone executor in 9router — Claude is served
> by `DefaultExecutor` after the translator converts. `ARCHITECTURE.md` was
> corrected to remove the phantom `executor/claude.go`.

---

## 5. OAuth Flows

Verified against `src/lib/oauth/services/`:

| Provider | Go file | Status |
|---|---|---|
| `claude` | `auth/oauth/claude.go` | ⬜ |
| `gemini` | `auth/oauth/gemini.go` | ⬜ |
| `codex` | `auth/oauth/codex.go` | ⬜ |
| `github` | `auth/oauth/github.go` | ⬜ |
| `kiro` | `auth/oauth/kiro.go` | ⬜ |
| `cursor` | `auth/oauth/cursor.go` | ⬜ |
| `antigravity` | `auth/oauth/antigravity.go` | ⬜ |
| `qwen` | `auth/oauth/qwen.go` | ⬜ |
| `iflow` | `auth/oauth/iflow.go` | ⬜ |
| `qoder` | `auth/oauth/qoder.go` | ⬜ |
| `openai` (ChatGPT login) | `auth/oauth/openai.go` | ⬜ |

> NOTE: `opencode` is NOT an OAuth flow in 9router — it's an executor only
> (uses static credentials). `ARCHITECTURE.md` was corrected.

---

## 6. RTK Filters

Verified against `open-sse/rtk/registry.js`:

| Filter | 9router file | Go file | Status |
|---|---|---|---|
| `git-diff` | `filters/gitDiff.js` | `rtk/filters/gitdiff.go` | ⬜ |
| `git-status` | `filters/gitStatus.js` | `rtk/filters/gitstatus.go` | ⬜ |
| `grep` (alias: `rg`) | `filters/grep.js` | `rtk/filters/grep.go` | ⬜ |
| `find` (alias: `fd`) | `filters/find.js` | `rtk/filters/find.go` | ⬜ |
| `ls` | `filters/ls.js` | `rtk/filters/ls.go` | ⬜ |
| `tree` | `filters/tree.js` | `rtk/filters/tree.go` | ⬜ |
| `dedup-log` | `filters/dedupLog.js` | `rtk/filters/dedup_log.go` | ⬜ |
| `smart-truncate` | `filters/smartTruncate.js` | `rtk/filters/smart_truncate.go` | ⬜ |
| `read-numbered` | `filters/readNumbered.js` | `rtk/filters/read_numbered.go` | ⬜ |
| `search-list` | `filters/searchList.js` | `rtk/filters/search_list.go` | ⬜ |

> NOTE: `git-log` is declared as `FILTERS.GIT_LOG = "git-log"` in
> `open-sse/rtk/constants.js:45` but has **no implementation** and is **not in
> the registry**. RTK_SPEC.md §7 was corrected to mark this as
> "declared but not implemented in golden source — do NOT port".

---

## 7. Caveman

| Mode | 9router ref | Go file | Status |
|---|---|---|---|
| LITE prompt (verbatim) | `open-sse/rtk/cavemanPrompts.js` | `caveman/prompts.go` | ⬜ |
| FULL prompt (verbatim) | same | same | ⬜ |
| ULTRA prompt (verbatim) | same | same | ⬜ |
| OpenAI injection | `caveman.js:injectMessagesSystem` | `caveman/inject.go` | ⬜ |
| Claude injection | `caveman.js:injectClaudeSystem` | `caveman/inject.go` | ⬜ |
| Gemini injection | `caveman.js:injectGeminiSystem` | `caveman/inject.go` | ⬜ |

---

## 8. Fallback / Combo / Multi-Account

| Feature | 9router ref | Status |
|---|---|---|
| Account-level fallback | `open-sse/services/accountFallback.js` | ⬜ |
| Combo iteration | `open-sse/services/combo.js` | ⬜ |
| Token refresh (pre + 401) | `open-sse/services/tokenRefresh.js` | ⬜ |
| Provider normalization | `src/lib/providerNormalization.js` | ⬜ |
| Disabled models cache | `src/lib/disabledModelsDb.js` | ⬜ |

---

## 9. Frontend Pages (dashboard parity)

| Page | 9router source | Next.js path | Status |
|---|---|---|---|
| Overview | `src/app/page.js` | `web/app/(dashboard)/page.tsx` | ⬜ |
| Providers | `src/app/providers/*` | `web/app/(dashboard)/providers/page.tsx` | ⬜ |
| Provider Nodes | matching | matching | ⬜ |
| Keys | matching | matching | ⬜ |
| Combos | matching | matching | ⬜ |
| Aliases | matching | matching | ⬜ |
| Pricing | matching | matching | ⬜ |
| Usage | matching | matching | ⬜ |
| Endpoint | matching | matching | ⬜ |
| Logs | matching | matching | ⬜ |
| Settings | matching | matching | ⬜ |
| Login | `src/app/login/*` | `web/app/(auth)/login/page.tsx` | ⬜ |

---

## 10. Cross-cutting

| Concern | Status |
|---|---|
| SSE flush correctness (no buffer >64KB) | ⬜ |
| Encrypted-at-rest OAuth tokens (AES-GCM) | ⬜ |
| Single `DATA_DIR` for ALL DBs (no usageDb split) | ⬜ |
| `//go:embed` static frontend | ⬜ |
| Single binary output (no CGO) | ⬜ |
| OpenAPI → `web/lib/types.gen.ts` codegen | ⬜ |
| Graceful shutdown (SIGTERM, drain SSEs) | ⬜ |

---

**Maintenance:** Every PR that ports a feature MUST tick the relevant row(s) and link the PR in the commit body.
