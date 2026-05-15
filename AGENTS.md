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

## When Adding a Feature

1. Check if similar feature exists in `_ref/9router/skills/`
2. Check if similar code exists in `_ref/9router/open-sse/` or `_ref/9router/src/`
3. Match the provider/model pattern
4. Add config in `providers/config.go`
5. Add route in `api/routes.go`
6. Handle format in `router/handler.go` if needed
