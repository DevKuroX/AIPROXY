# AIPROXY — AI Gateway

OpenAI-compatible AI gateway with 55+ providers, streaming, account pool with state machine, proxy pools, RTK compression, and conversation compact.

**Port**: 1432  
**Base**: `http://localhost:1432/v1`  
**Auth**: Bearer token from `api_keys` table

---

## Quick Start

```bash
# Setup
export DATABASE_URL="postgresql://user:pass@localhost:5432/aiproxy"
cd backend && go build -o bin/server ./cmd/server && ./bin/server

# Test health
curl http://localhost:1432/health

# Chat with Kiro (free, needs account)
curl -X POST http://localhost:1432/v1/chat/completions \
  -H "Authorization: Bearer <your-api-key>" \
  -d '{"model":"kiro/claude-haiku-4.5","messages":[{"role":"user","content":"hi"}]}'

# Chat with OpenCode (free, no auth needed)
curl -X POST http://localhost:1432/v1/chat/completions \
  -H "Authorization: Bearer <your-api-key>" \
  -d '{"model":"oc/deepseek-v4-flash-free","messages":[{"role":"user","content":"hi"}]}'
```

---

## Architecture

```
Request → /v1/chat/completions
              ↓
         ParseModel("provider/model") → provider + model
              ↓
         GetProviderConfig → {BaseURL, AuthType, Format, Headers}
              ↓
         GetAccount from Pool (in-memory)
              ↓
         RTK Compress tool output (if enabled)
              ↓
         Inject Caveman prompt (if enabled)
              ↓
         Translate request format
              ↓
         Call Provider API
              ↓
         Translate response format → Return
```

### Flow Details

**Auth Types**:
- `AuthTypeNone` → virtual account `{accessToken: "public"}`, no DB needed
- `AuthTypeBearer` → API key stored in `api_keys` table
- `AuthTypeOAuth` → token from `provider_accounts` with auto-refresh

**Account State Machine**:
```
Active (>20% credit) → Depleting (<=20%) → RateLimited (429) → Exhausted (0%)
Selection priority: Active > Depleting > RateLimited > Exhausted
```

---

## Directory Map

```
backend/
├── cmd/server/main.go           # Entry point, wires everything
├── internal/
│   ├── api/
│   │   ├── routes.go             # All HTTP routes
│   │   ├── usage.go              # Kiro quota API
│   │   ├── proxy_api.go          # Proxy pool CRUD + scraper API
│   │   └── admin/                # Admin handlers (auth, settings)
│   ├── pool/
│   │   ├── account.go            # Account model + state machine
│   │   └── pool.go               # Pool manager (in-memory, state-based selection)
│   ├── providers/
│   │   └── config.go             # 55+ provider configurations
│   ├── router/
│   │   ├── handler.go            # Chat handler (streaming + non-streaming)
│   │   ├── model_parser.go       # Parse "provider/model" format
│   │   ├── anthropic_handler.go  # /v1/messages endpoint
│   │   ├── models_handler.go     # /v1/models discovery
│   │   ├── compact_handler.go    # /v1/responses/compact (self-summary)
│   │   └── rate_limit.go         # Exponential backoff
│   ├── proxy/
│   │   ├── model.go              # Proxy, ProxyPool structs + 7 scraper sources
│   │   ├── scraper.go            # Geonode, ProxyScrape, GitHub, WebShare scrapers
│   │   ├── tester.go             # Latency + region detection
│   │   ├── pool.go               # Proxy manager (selection, toggle, CRUD)
│   │   └── pgstore.go            # PostgreSQL store implementation
│   ├── rtk/                      # Response Token Keeper (compress tool output)
│   │   ├── autodetect.go         # Auto-detect content type
│   │   ├── caveman.go            # Terse prompt injector
│   │   ├── caveman_prompts.go    # Prompt templates
│   │   └── filters/              # git, grep, ls, tree, find, etc (10 filters)
│   ├── translator/               # Request/response format translators
│   │   ├── request/              # OpenAI → Gemini/Claude
│   │   └── response/             # Gemini/Claude → OpenAI
│   ├── auth/oauth/               # OAuth flows (Kiro, Claude refresh)
│   └── storage/
│       ├── migrations/           # DB migration files
│       └── *.go                  # DB operations (providers, accounts, keys)
```

---

## How To: Add a New Provider

### 1. Add config entry

Edit `backend/internal/providers/config.go`:

```go
// For API key provider (OpenAI-compatible)
"my-llm": {
    Name: "My LLM", Type: TypeOpenAI,
    BaseURL: "https://api.myprovider.com/v1", AuthType: AuthTypeBearer, Format: FormatOpenAI,
},

// For no-auth provider (free, public)
"my-free": {
    Name: "My Free LLM", Type: TypeOpenAI,
    BaseURL: "https://free-api.com/v1", AuthType: AuthTypeNone, Format: FormatOpenAI,
},
```

### 2. Add account to DB

```sql
-- For bearer auth
INSERT INTO api_keys (key, key_hash, name, is_active) VALUES ('sk-...', 'hash...', 'My Key', true);

-- For OAuth  
INSERT INTO providers (name, type, base_url) VALUES ('My LLM', 'my-llm', 'https://api.myprovider.com/v1');
INSERT INTO provider_accounts (provider_id, name, api_key) VALUES (1, 'account@email.com', 'sk-...');
```

### 3. Done! Use it:

```bash
curl -d '{"model":"my-llm/gpt-4","messages":[...]}'
```

---

## How To: Add a New Filter (RTK)

Edit `backend/internal/rtk/filters/`:

1. Create `myfilter.go` with `Filter` function
2. Register in `registry.go`'s `init()` function
3. Add detection pattern in `autodetect.go`

---

## API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/health` | Health check | None |
| POST | `/v1/chat/completions` | Chat (OpenAI format) | API Key |
| POST | `/v1/messages` | Chat (Anthropic format) | API Key |
| POST | `/v1/responses/compact` | Compact conversation | API Key |
| GET | `/v1/models` | List LLM models | API Key |
| GET | `/v1/models/image` | List image models | API Key |
| GET | `/v1/models/tts` | List TTS models | API Key |
| GET | `/v1/models/embedding` | List embedding models | API Key |
| GET | `/v1/api/usage/kiro` | Kiro credit quota | API Key |
| GET | `/api/proxy/settings` | Proxy settings | API Key |
| POST | `/api/proxy/settings` | Update proxy settings | API Key |
| GET | `/api/proxies` | List proxies | API Key |
| POST | `/api/proxies` | Add proxy | API Key |
| GET | `/api/proxy-pools` | List proxy pools | API Key |
| POST | `/api/proxy-pools` | Create pool | API Key |
| POST | `/api/scraper/start` | Start proxy scraper | API Key |
| GET | `/api/admin/settings` | Get RTK/Caveman settings | JWT |
| POST | `/api/admin/settings` | Update settings | JWT |

---

## Key Features

| Feature | File | Status |
|---------|------|--------|
| 55+ Providers | `internal/providers/config.go` | ✅ |
| Streaming SSE | `internal/router/handler.go` | ✅ |
| Anthropic API | `internal/router/anthropic_handler.go` | ✅ |
| Model Discovery | `internal/router/models_handler.go` | ✅ |
| Account Pool + State Machine | `internal/pool/` | ✅ |
| Auto Token Refresh | `internal/router/handler.go` | ✅ |
| Rate Limit Backoff | `internal/router/rate_limit.go` | ✅ |
| RTK Tool Compression | `internal/rtk/` | ✅ |
| Caveman Terse Prompts | `internal/rtk/caveman.go` | ✅ |
| Proxy Scraper (7 sources) | `internal/proxy/` | ✅ |
| Proxy Pool Manager | `internal/proxy/pool.go` | ✅ |
| Per-Provider Proxy Toggle | `internal/proxy/model.go` | ✅ |
| Conversation Compact | `internal/router/compact_handler.go` | ✅ |
| Kiro Quota API | `internal/api/usage.go` | ✅ |
| WebSearch (stub) | `internal/api/v1/search.go` | ✅ |
| WebFetch (stub) | `internal/api/v1/fetch.go` | ✅ |
| Image Gen (stub) | `internal/api/v1/images.go` | ✅ |

---

## Architecture Decisions

**Why model format is "provider/model"**?
- Routing key = provider prefix, model name = what provider receives
- 
**Why memory-first for account pool?**
- Zero DB latency per request
- DB only hit on startup + periodic sync (planned)

**Why separate proxy package?**
- Proxy is cross-cutting (any provider can use it)
- Scraper + tester + pool = complex domain, better isolated

**Why RTK + Caveman?**
- RTK: compress tool output (60-90% savings) — no LLM cost
- Caveman: terse prompts reduce output tokens — no LLM cost
- 

---

## Environment

```env
DATABASE_URL=postgresql://user:pass@localhost:5432/aiproxy
JWT_SECRET=your-secret
ADMIN_PASSWORD=admin123
PORT=1432
```

---

## Dependencies

- Go 1.25+
- PostgreSQL 15+
- `github.com/jackc/pgx/v5` — PostgreSQL driver
- `github.com/golang-jwt/jwt/v5` — JWT auth
- `golang.org/x/crypto` — password hashing
