# AGENTS.md — AI Agent Guide for AIPROXY

This file is for **AI coding agents** (Claude, GPT, OpenCode, etc.) when making changes to this project.

---

## First: Read These Files

```
README.md          → Project overview, architecture, how to add providers
backend/internal/
  providers/config.go  → All provider configs (add new ones here)
  pool/account.go      → Account state machine model
  pool/pool.go         → Pool selection logic
  router/handler.go    → Main chat handler (streaming + non-streaming)
  rtk/                 → RTK compression + caveman module
  proxy/               → Proxy scraper + pool module
  api/routes.go        → All API routes
```

---

## Conventions

### Model Format
```
"provider/model-name" — always. E.g., "kiro/claude-haiku-4.5", "oc/deepseek-v4-flash-free"
ParseModel() splits on "/" → provider + model
```

### Adding a Provider
1. Add to `internal/providers/config.go` (one entry)
2. No other file changes needed — routing is automatic
3. If special format (not OpenAI-compatible), add format handling in:
   - `router/handler.go` → `callProviderAPI()` switch case
   - `router/handler.go` → `buildProviderURL()` switch case

### Auth Types
```go
AuthTypeNone    → virtual account, no DB row needed
AuthTypeBearer  → requires api_keys table entry
AuthTypeOAuth   → requires provider_accounts + refresh token
AuthTypeCookie  → for cookie-based providers (grok-web, perplexity-web)
```

### DO NOT
- DO NOT modify `cmd/server/main.go` unless adding new global subsystem
- DO NOT hardcode credentials (use DB or config)
- DO NOT use `interface{}` for new code — use proper types
- DO NOT skip comments on complex logic (state machine, EventStream parsing)
- DO NOT add external dependencies without approval

### DO
- Add new providers to the BOTTOM of their section in config.go
- Follow existing filter pattern when adding RTK filters
- Use existing translator patterns for format conversion
- Test with `go build ./cmd/server` before committing

---

## Common Gotchas

**Kiro EventStream**: AWS returns binary EventStream, not JSON. Handler parses it chunk by chunk. See `callProviderAPI()` in handler.go for the binary parsing logic.

**Token Refresh**: Kiro tokens expire in 1 hour. Refresh via `POST https://prod.us-east-1.auth.desktop.kiro.dev/refreshToken`. Set `expires_at` in DB to enable proactive refresh.

**Account States**:
- `StateActive` → credit > 20%, selected first
- `StateDepleting` → credit <= 20%, selected if no active
- `StateRateLimited` → 429 received, cooldown then back to active
- `StateExhausted` → credit = 0%, skip until quota reset

**Gemini Web Response Parser** (`backend/internal/geminiweb/response.go`):

  The Gemini Web API uses a **length-prefixed frame protocol**, NOT newline-delimited JSON.
  Response format: `)]}'\n\n{digits}\n{json}\n` where `digits` = UTF-16 code unit count of `\n{json}\n`.

  **Bugs found and fixed:**

  1. **Frame boundary off-by-1** — Counting UTF-16 from `nl+1` (after the `\n` following digits) is WRONG. The leading `\n` IS counted in the length. Fix: count from `nl` (the `\n` after digits), then strip the leading `\n` from the extracted frame.

  2. **Security prefix** — Stripping `)]}'` via `\n\n` search is WRONG. Use exact 4-char prefix check: `strings.HasPrefix(content, ")]}'")` then `content = strings.TrimLeft(content[4:], " \t\r\n")`.

  3. **bufio.Scanner destroys frames** — Scanner strips `\n` delimiters, breaking length-precise frame parsing. Fix: use raw `io.Read` byte buffer + `strings.Builder` accumulator.

  4. **TrimSpace strips counted newline** — `strings.TrimSpace()` removes trailing `\n` that's counted in the frame length. Fix: use `TrimLeft` only.

  5. **`candData[8]==[1]` is NOT a Done indicator** — This is a standard flag present on ALL text frames. Checking it causes `Done=true` on the first text frame, terminating the stream prematurely. Fix: remove this check entirely. Use `"e"` frame and `"di"` frame for Done detection.

  6. **Streaming deadlock** — Without `return nil` after `chunk.Done==true` in ParseStream(), Gemini keeps the HTTP connection alive forever since there's no natural end-of-stream marker. Fix: `return nil` when `chunk.Done` is true.

  7. **Wrong tuple length check** — `len(tuple) < 4` rejects valid frames (real Gemini tuples have 3 elements). Fix: `len(tuple) < 1`.

  8. **Inner JSON marshaling** — The inner request must be marshaled to a JSON STRING before embedding in `f.req` payload. Python ref: `json.dumps(inner_req_list)` produces a string, not an object. Fix: `string(innerJSON)` instead of passing `innerReq` as object.

  **Frame types in response:**
  - `"wrb.fr"` or `"wra"` — Response wrapper. Index `[2]` contains inner JSON payload string. Parse with `parseInnerPayload()` to extract text from `payload[4]` (candidates).
  - `"di"` — Done indicator. Index `[1]` contains a count value. Check `v >= 2` (not `v == 2`, actual values are large like 4935).
  - `"e"` — End-of-stream. Format: `["e", 4, null, null, 137]`. Sets `Done=true`.

  **Image upload** (`backend/internal/geminiweb/image_test.go`):
  - Upload to `POST https://content-push.googleapis.com/upload` with multipart form-data
  - Headers: `X-Tenant-Id: bard-storage`, `Push-ID: feeds/mcudyrk2a4khkz`
  - Response body IS the uploaded URL (e.g. `/contrib_service/ttl_1d/...`)
  - Embed in `messageContent[3]` as `[[[uploadedURL], filename]]` (triple-nested array for single file)

**Proxy Selection**: ProxyManager.SelectProxy() picks lowest-latency alive proxy. Per-provider toggle in ProxySettings.

**RTK Settings**: `rtkEnabled=true` by default (compresses tool output). `cavemanEnabled=false` by default (terse prompts off). Toggle via `POST /api/admin/settings`.

---

## Migration Files

Located in `internal/storage/migrations/`. Named sequentially: `001_init.sql`, `002_...`, etc.
New migrations must follow the same naming convention and be idempotent (use `IF NOT EXISTS`).

---

## Testing

```bash
cd backend
go build ./cmd/server          # Check compilation
go build ./internal/...        # Check package compilation
# Start and test:
./bin/server &
curl localhost:1432/health
```

---

## AIPROXY Skills

Skills are bite-sized docs that tell agents how to use specific AIPROXY capabilities
without reading the whole codebase. Load a skill when the task matches its domain.

Available skills (in `backend/skills/`):

| Skill | When to load | Content |
|---|---|---|
| `aiproxy/SKILL.md` | Any AIPROXY task — start here | Setup, CLI, model format, features |
| `aiproxy-chat/SKILL.md` | Chat, code-gen, LLM completion | Endpoints, streaming, providers |
| `aiproxy-image/SKILL.md` | Image generation | Providers, prompt format |
| `aiproxy-tts/SKILL.md` | Text-to-speech | Endpoint, voices |
| `aiproxy-stt/SKILL.md` | Speech-to-text | Audio transcription |
| `aiproxy-embeddings/SKILL.md` | Text embeddings | Vector embeddings |
| `aiproxy-web/SKILL.md` | Web search or URL fetch | Search + fetch endpoints |

To use: `task(subagent_type="explore", ..., prompt="Read backend/skills/aiproxy/SKILL.md and ...")`

## When Adding a Feature

1. Check if similar feature exists in `_ref/9router/skills/`
2. Check if similar code exists in `_ref/9router/open-sse/` or `_ref/9router/src/`
3. Match the provider/model pattern
4. Add config in `providers/config.go`
5. Add route in `api/routes.go`
6. Handle format in `router/handler.go` if needed
7. Create/update skill in `backend/skills/` for the new capability
