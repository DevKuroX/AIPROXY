# CODEBASE.md - Comprehensive Navigation Guide

> **Purpose**: Single source of truth for understanding both the current ai_proxy project structure and the 9router reference implementation we're porting from.
> 
> **Audience**: AI agents, developers, and contributors working on the Go port.
> 
> **Last Updated**: 2026-05-11 (auto-generated from parallel exploration)

---

## 📋 Table of Contents

1. [Quick Start](#quick-start)
2. [Current Project Status](#current-project-status)
3. [9Router Reference Architecture](#9router-reference-architecture)
4. [Request Lifecycle](#request-lifecycle)
5. [Configuration Reference](#configuration-reference)
6. [Module Mapping](#module-mapping)
7. [Porting Guide](#porting-guide)

---

## 🚀 Quick Start

### For New Agents

1. **Read this file first** - understand what exists and what we're porting from
2. **Check Phase Status** - see `AGENTS.md` "Current Phase" section
3. **Find Your Task** - use the Module Mapping section to locate relevant code
4. **Read Specs** - check `docs/<subsystem>.md` for detailed requirements
5. **Port with Parity** - reference 9router files, maintain behavior exactly

### For Specific Tasks

| Task | Start Here |
|------|------------|
| Implement RTK filter | [RTK Module](#rtk-token-saver) + `docs/RTK_SPEC.md` |
| Add new executor | [Executors](#executors) + `docs/EXECUTORS.md` |
| Port translator | [Translators](#translators) + `docs/TRANSLATORS.md` |
| Fix auth flow | [Auth](#authentication--authorization) + `docs/AUTH.md` |
| Debug request flow | [Request Lifecycle](#request-lifecycle) |
| Add env var | [Configuration Reference](#configuration-reference) |

---

## 📊 Current Project Status

**Phase**: 0 (Foundation) - see `ROADMAP.md`

**Repository**: `/home/ubuntu/ai_proxy`

### Directory Structure (Planned vs. Actual)

```
ai_proxy/
├── cmd/                    ❌ NOT CREATED - Go entry point
│   └── server/
│       └── main.go
├── internal/               ❌ NOT CREATED - Core Go packages
│   ├── api/
│   ├── router/
│   ├── translator/
│   ├── executor/
│   ├── rtk/
│   ├── caveman/
│   ├── auth/
│   ├── storage/
│   ├── models/
│   ├── stream/
│   └── config/
├── web/                    ❌ NOT CREATED - Next.js dashboard
├── assets/web/             ❌ NOT CREATED - Embedded static files
├── docs/                   ✅ COMPLETE - All 16 spec files exist
│   ├── ARCHITECTURE.md     ✅ Complete specification
│   ├── PARITY_CHECKLIST.md ✅ Feature tracker (all ⬜ not started)
│   ├── RTK_SPEC.md         ✅ Token saver spec
│   ├── CAVEMAN_SPEC.md     ✅ Prompt injector spec
│   ├── TRANSLATORS.md      ✅ Format conversion spec
│   ├── EXECUTORS.md        ✅ Provider adapters spec
│   ├── API.md              ✅ Endpoint contracts
│   ├── AUTH.md             ✅ JWT/OAuth spec
│   ├── DATABASE.md         ✅ Schema spec
│   ├── STREAMING.md        ✅ SSE spec
│   ├── FALLBACK.md         ✅ Combo/account fallback spec
│   ├── FRONTEND.md         ✅ Dashboard spec
│   ├── CLI_TOOLS.md        ✅ Tool compatibility spec
│   ├── OBSERVABILITY.md    ✅ Logging/metrics spec
│   ├── CONVENTIONS.md      ✅ Go style guide
│   └── USAGE.md            ✅ Usage tracking spec
├── _ref/9router/           ✅ READ-ONLY - Reference implementation
├── AGENTS.md               ✅ Master AI agent context
├── PLAN.md                 ✅ Vision document
├── ROADMAP.md              ✅ Phased execution plan
├── AUDIT.md                ✅ Project audit
├── .gitignore              ✅ Created
└── CODEBASE.md             ✅ This file
```

### Implementation Status

**Go Backend**: ❌ Not started (Phase 0 pending)
- No `go.mod` yet
- No `cmd/` or `internal/` directories
- No code files

**Next.js Dashboard**: ❌ Not started (Phase 0 pending)
- No `web/` directory
- No `package.json`

**Documentation**: ✅ Complete
- All 16 spec files created
- AGENTS.md, PLAN.md, ROADMAP.md complete
- Parity checklist initialized (all items ⬜)

**Reference Code**: ✅ Available
- 9router checked out at `_ref/9router/`
- Read-only for agents
- Complete source for porting

### Next Steps (Phase 0)

Per `ROADMAP.md`, Phase 0 requires:

1. ✅ Documentation complete
2. ❌ Initialize Go module (`go mod init`)
3. ❌ Set up `chi` router skeleton
4. ❌ Configure `slog` logging
5. ❌ Implement env config loader
6. ❌ Set up Next.js dashboard scaffold
7. ❌ Wire JWT login flow

---

## 🏗️ 9Router Reference Architecture

**Location**: `/home/ubuntu/ai_proxy/_ref/9router/`

**Version**: 0.4.29 (from package.json)

**Stack**: Next.js 16 + Express + SQLite (better-sqlite3 or sql.js fallback)

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Tools                            │
│  (Code Assistant, Cursor, Codex, Cline, OpenClaw...)    │
└────────────────────┬────────────────────────────────────┘
                     │ HTTP/SSE
                     │ localhost:20128/v1/*
                     ▼
┌─────────────────────────────────────────────────────────┐
│              9Router (Next.js App)                      │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Entry Points                                   │   │
│  │  • /v1/chat/completions (SSE + API route)      │   │
│  │  • /v1/models (model list)                     │   │
│  │  • /api/* (admin dashboard API)                │   │
│  └─────────────────────────────────────────────────┘   │
│                     │                                   │
│                     ▼                                   │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Core Pipeline (chatCore.js)                   │   │
│  │  1. Auth (resolve provider credentials)        │   │
│  │  2. Format Detection (0penAI/CL4ude/etc)       │   │
│  │  3. RTK Token Saver (compress tool_result)     │   │
│  │  4. Caveman (inject prompt)                    │   │
│  │  5. Translation (format conversion)            │   │
│  │  6. Executor (HTTP to provider)                │   │
│  │  7. Response Translation                       │   │
│  │  8. SSE Streaming                              │   │
│  └─────────────────────────────────────────────────┘   │
│                     │                                   │
│                     ▼                                   │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Storage (SQLite)                              │   │
│  │  • accounts (provider credentials)             │   │
│  │  • usage (token tracking)                      │   │
│  │  • settings (app config)                       │   │
│  │  • api_keys (endpoint auth)                    │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                     │ HTTPS
                     ▼
┌─────────────────────────────────────────────────────────┐
│          Upstream Providers (40+)                       │
│  0penAI, CL4ude, Gemini, GLM, MiniMax, Kiro, etc.      │
└─────────────────────────────────────────────────────────┘
```

### Directory Structure

```
9router/
├── src/                          # Next.js app source
│   ├── app/                      # App Router pages
│   │   ├── api/v1/               # 0penAI-compatible API routes
│   │   │   ├── chat/completions/ # Main chat endpoint
│   │   │   └── models/           # Model list endpoint
│   │   ├── api/                  # Admin API routes
│   │   │   ├── auth/             # Login/logout
│   │   │   ├── accounts/         # Provider account CRUD
│   │   │   ├── settings/         # App settings
│   │   │   └── usage/            # Usage stats
│   │   └── dashboard/            # Dashboard UI pages
│   ├── sse/                      # SSE handlers (legacy)
│   │   ├── handlers/             # chat.js, models.js
│   │   └── services/             # auth.js, usage.js
│   ├── lib/                      # Shared utilities
│   │   ├── db/                   # SQLite layer
│   │   │   ├── index.js          # DB connection
│   │   │   ├── schema.js         # Table definitions
│   │   │   └── repos/            # Repository pattern
│   │   └── utils/                # Helpers
│   ├── shared/                   # Frontend/backend shared
│   │   ├── constants/            # App constants
│   │   └── utils/                # Shared utilities
│   └── components/               # React components
├── open-sse/                     # Core SSE engine (reusable)
│   ├── handlers/                 # Request handlers
│   │   └── chatCore.js           # ⭐ Main orchestration
│   ├── executors/                # Provider HTTP adapters
│   │   ├── index.js              # Executor registry
│   │   ├── DefaultExecutor.js    # Base class
│   │   ├── openai.js             # 0penAI executor
│   │   ├── claude.js             # CL4ude executor
│   │   ├── gemini.js             # Gemini executor
│   │   └── ... (17 more)
│   ├── translator/               # Format conversion
│   │   ├── index.js              # Translator registry
│   │   ├── request/              # Request translators (11)
│   │   └── response/             # Response translators (9)
│   ├── rtk/                      # Token saver
│   │   ├── index.js              # compressMessages()
│   │   ├── constants.js          # Filter limits
│   │   ├── caveman.js            # Prompt injector
│   │   └── filters/              # 10 compression filters
│   ├── services/                 # Business logic
│   │   ├── combo.js              # Combo routing
│   │   ├── accountFallback.js    # Account selection
│   │   ├── model.js              # Model resolution
│   │   ├── provider.js           # Provider config
│   │   └── usage.js              # Token tracking
│   ├── config/                   # Configuration
│   │   ├── providers.js          # 40+ provider definitions
│   │   ├── providerModels.js     # Model mappings
│   │   ├── runtimeConfig.js      # Runtime constants
│   │   └── appConstants.js       # App constants
│   ├── utils/                    # Utilities
│   │   ├── streamHelpers.js      # SSE utilities
│   │   ├── proxyFetch.js         # HTTP proxy
│   │   └── ... (12 more)
│   └── transformer/              # Response transformers
│       ├── streamToJsonConverter.js
│       └── responsesTransformer.js
├── docs/                         # Documentation
├── tests/                        # Test suite
└── package.json                  # Dependencies
```

### Key Files for Porting

| 9Router File | Purpose | Port To (Go) |
|--------------|---------|--------------|
| `open-sse/handlers/chatCore.js` | Main orchestration | `internal/router/handler.go` |
| `open-sse/executors/index.js` | Executor registry | `internal/executor/registry.go` |
| `open-sse/executors/DefaultExecutor.js` | Base executor | `internal/executor/base.go` |
| `open-sse/translator/index.js` | Translator registry | `internal/translator/registry.go` |
| `open-sse/rtk/index.js` | Token saver | `internal/rtk/compress.go` |
| `open-sse/rtk/caveman.js` | Prompt injector | `internal/caveman/inject.go` |
| `open-sse/config/providers.js` | Provider configs | `internal/models/providers.go` |
| `open-sse/services/combo.js` | Combo routing | `internal/router/combo.go` |
| `open-sse/services/accountFallback.js` | Account fallback | `internal/router/fallback.go` |
| `src/lib/db/schema.js` | Database schema | `internal/storage/migrations/*.sql` |
| `src/sse/services/auth.js` | Auth service | `internal/auth/provider.go` |

---

## 🔄 Request Lifecycle

### Complete Flow (9Router)

```
1. CLI Tool Request
   ↓
2. Next.js Route Handler
   File: src/app/api/v1/chat/completions/route.js
   Action: Receives POST /v1/chat/completions
   ↓
3. SSE Handler Entry
   File: src/sse/handlers/chat.js
   Action: Detects combo vs. single model, routes accordingly
   ↓
4. Core Orchestration
   File: open-sse/handlers/chatCore.js
   Function: handleChatCore()
   ↓
   ├─→ 4a. Auth Service
   │   File: src/sse/services/auth.js
   │   Function: resolveProviderCredentials()
   │   Action: Select account, check quota, handle fallback
   │   ↓
   ├─→ 4b. Format Detection
   │   File: open-sse/handlers/chatCore.js
   │   Action: Detect request format (0penAI/CL4ude/Gemini)
   │   ↓
   ├─→ 4c. RTK Token Saver
   │   File: open-sse/rtk/index.js
   │   Function: compressMessages()
   │   Action: Apply 10 filters to compress tool_result
   │   Filters: dedupLog, grep, tree, git-diff, ls, ps, find, npm, curl, generic
   │   ↓
   ├─→ 4d. Caveman Injection
   │   File: open-sse/rtk/caveman.js
   │   Function: injectCavemanPrompt()
   │   Action: Append system instructions if enabled
   │   ↓
   ├─→ 4e. Request Translation
   │   File: open-sse/translator/index.js
   │   Function: translateRequest()
   │   Action: Convert format (e.g., 0penAI → CL4ude)
   │   ↓
   ├─→ 4f. Executor Dispatch
   │   File: open-sse/executors/index.js
   │   Function: getExecutor()
   │   Action: Get provider-specific executor or DefaultExecutor
   │   ↓
   ├─→ 4g. HTTP Request
   │   File: open-sse/executors/DefaultExecutor.js
   │   Function: execute()
   │   Action: POST to upstream provider with auth headers
   │   ↓
   ├─→ 4h. Response Translation
   │   File: open-sse/translator/index.js
   │   Function: translateResponse()
   │   Action: Convert response format back to client's expected format
   │   ↓
   └─→ 4i. SSE Streaming
       File: open-sse/utils/streamHelpers.js
       Action: Pipe upstream SSE chunks to client
       ↓
5. Client Receives Response
   Format: SSE stream (data: {...}\n\n)
```

### Error Handling & Fallback

```
Request
  ↓
Auth Service
  ├─→ Primary Account
  │   ├─→ Success → Continue
  │   └─→ Fail (quota/rate limit)
  │       ↓
  ├─→ Fallback Account (same provider)
  │   ├─→ Success → Continue
  │   └─→ Fail
  │       ↓
  └─→ Fallback Provider (different tier)
      ├─→ Tier 1 (subscription) → Tier 2 (cheap) → Tier 3 (free)
      └─→ All failed → Return 503 error
```

### Combo Routing

When model name contains `+` (e.g., `gpt-4+claude-3.5`):

```
Request with combo model
  ↓
Combo Service (open-sse/services/combo.js)
  ├─→ Parse model names: ["gpt-4", "claude-3.5"]
  ├─→ Resolve providers: [openai, claude]
  ├─→ For each model:
  │   ├─→ Clone request
  │   ├─→ Set model
  │   ├─→ Call handleChatCore()
  │   └─→ Collect response
  └─→ Merge responses (interleave chunks)
      └─→ Stream to client
```

---

## ⚙️ Configuration Reference

### Environment Variables

**Source**: `_ref/9router/.env.example`

| Variable | Type | Default | Purpose | Required |
|----------|------|---------|---------|----------|
| `JWT_SECRET` | string | - | JWT signing key | ✅ |
| `INITIAL_PASSWORD` | string | - | Admin password on first boot | ✅ |
| `DATA_DIR` | string | `/var/lib/9router` | SQLite DB location | ✅ |
| `PORT` | number | `20128` | HTTP server port | ❌ |
| `NODE_ENV` | string | `production` | Runtime environment | ❌ |
| `API_KEY_SECRET` | string | - | API key encryption secret | ❌ |
| `MACHINE_ID_SALT` | string | - | Machine ID salt | ❌ |
| `ENABLE_REQUEST_LOGS` | boolean | `false` | Log all requests | ❌ |
| `OBSERVABILITY_ENABLED` | boolean | `true` | Enable metrics | ❌ |
| `AUTH_COOKIE_SECURE` | boolean | `false` | Require HTTPS for cookies | ❌ |
| `REQUIRE_API_KEY` | boolean | `false` | Require API key for /v1/* | ❌ |
| `BASE_URL` | string | `http://localhost:20128` | Public base URL | ❌ |
| `CLOUD_URL` | string | `https://9router.com` | Cloud sync URL | ❌ |
| `HTTP_PROXY` | string | - | Outbound HTTP proxy | ❌ |
| `HTTPS_PROXY` | string | - | Outbound HTTPS proxy | ❌ |
| `ALL_PROXY` | string | - | Outbound SOCKS proxy | ❌ |
| `NO_PROXY` | string | - | Proxy bypass list | ❌ |

### Runtime Constants

**Source**: `_ref/9router/open-sse/config/runtimeConfig.js`

```javascript
{
  CACHE_TTL: 5 * 60 * 1000,              // 5 minutes
  RETRY_DELAY: 1000,                      // 1 second
  MAX_RETRIES: 3,
  STREAM_TIMEOUT: 60 * 1000,              // 60 seconds
  CHUNK_SIZE: 8192,                       // 8KB
  MAX_CONCURRENT_STREAMS: 100,
  RATE_LIMIT_WINDOW: 60 * 1000,           // 1 minute
  RATE_LIMIT_MAX_REQUESTS: 60,
}
```

### RTK Filter Limits

**Source**: `_ref/9router/open-sse/rtk/constants.js`

```javascript
{
  MAX_LOG_LINES: 50,
  MAX_GREP_LINES: 100,
  MAX_TREE_DEPTH: 3,
  MAX_DIFF_LINES: 200,
  MAX_LS_ENTRIES: 50,
  MAX_PS_PROCESSES: 30,
  MAX_FIND_RESULTS: 100,
  MAX_NPM_DEPS: 50,
  MAX_CURL_RESPONSE: 1000,
  GENERIC_TRUNCATE_THRESHOLD: 5000,
}
```

### Database Schema

**Source**: `_ref/9router/src/lib/db/schema.js`

```sql
-- accounts: Provider credentials
CREATE TABLE accounts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  provider TEXT NOT NULL,
  name TEXT NOT NULL,
  credentials TEXT NOT NULL,  -- JSON encrypted
  quota_limit INTEGER,
  quota_used INTEGER DEFAULT 0,
  quota_reset_at INTEGER,
  enabled INTEGER DEFAULT 1,
  created_at INTEGER DEFAULT (strftime('%s', 'now')),
  updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- usage: Token tracking
CREATE TABLE usage (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  account_id INTEGER,
  model TEXT NOT NULL,
  prompt_tokens INTEGER DEFAULT 0,
  completion_tokens INTEGER DEFAULT 0,
  total_tokens INTEGER DEFAULT 0,
  cost REAL DEFAULT 0,
  created_at INTEGER DEFAULT (strftime('%s', 'now')),
  FOREIGN KEY (account_id) REFERENCES accounts(id)
);

-- settings: App configuration
CREATE TABLE settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- api_keys: Endpoint authentication
CREATE TABLE api_keys (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT UNIQUE NOT NULL,  -- sk-{machineId}-{keyId}-{crc8}
  name TEXT,
  enabled INTEGER DEFAULT 1,
  created_at INTEGER DEFAULT (strftime('%s', 'now'))
);
```

---

## 🧩 Module Mapping

### RTK (Token Saver)

**9Router Location**: `_ref/9router/open-sse/rtk/`

**Go Port Location**: `internal/rtk/`

**Purpose**: Compress `tool_result` content in messages to save 20-40% tokens.

**Core Function**: `compressMessages(messages, filters)`

**Filters** (10 total):

| Filter | File | Purpose | Max Lines |
|--------|------|---------|-----------|
| `dedup-log` | `filters/dedupLog.js` | Remove duplicate log lines | 50 |
| `grep` | `filters/grep.js` | Truncate grep output | 100 |
| `tree` | `filters/tree.js` | Limit tree depth | 3 levels |
| `git-diff` | `filters/gitDiff.js` | Truncate diff output | 200 |
| `ls` | `filters/ls.js` | Limit ls entries | 50 |
| `ps` | `filters/ps.js` | Limit process list | 30 |
| `find` | `filters/find.js` | Limit find results | 100 |
| `npm` | `filters/npm.js` | Truncate npm output | 50 |
| `curl` | `filters/curl.js` | Truncate HTTP responses | 1000 chars |
| `generic` | `filters/generic.js` | Fallback truncation | 5000 chars |

**Data Flow**:
```
messages[] → compressMessages() → for each message with tool_result:
  → detect tool type (grep/tree/ls/etc)
  → apply specific filter
  → fallback to generic if no match
  → return compressed messages[]
```

**Porting Notes**:
- Each filter is a pure function: `(content: string) => string`
- Filters are stateless and can run in parallel
- Must preserve exact truncation logic for parity
- See `docs/RTK_SPEC.md` for detailed spec

---

### Caveman (Prompt Injector)

**9Router Location**: `_ref/9router/open-sse/rtk/caveman.js`

**Go Port Location**: `internal/caveman/inject.go`

**Purpose**: Inject aggressive compression instructions into system message.

**Trigger**: When `caveman` setting is enabled in account config.

**Injection Point**: Appends to last `system` role message, or creates one if none exists.

**Prompt Template**:
```
CAVEMAN MODE: Respond in ultra-compressed style.
- No pleasantries, no explanations
- Code only, minimal comments
- Shortest possible variable names
- No markdown formatting unless code block
- One-word answers when possible
```

**Data Flow**:
```
messages[] → injectCavemanPrompt(messages, enabled) →
  if enabled:
    → find last system message
    → append caveman instructions
    → return modified messages[]
  else:
    → return messages[] unchanged
```

**Porting Notes**:
- Must run AFTER RTK compression
- Must run BEFORE request translation
- Exact prompt text matters for parity
- See `docs/CAVEMAN_SPEC.md` for full spec

---

### Translators

**9Router Location**: `_ref/9router/open-sse/translator/`

**Go Port Location**: `internal/translator/`

**Purpose**: Convert request/response formats between different provider APIs.

**Supported Formats**:
- `openai` - 0penAI Chat Completions API
- `claude` - CL4ude Messages API
- `gemini` - Gemini generateContent API
- `vertex` - Vertex AI API
- `cohere` - Cohere Chat API
- `ollama` - Ollama API
- `openrouter` - OpenRouter API
- `azure` - Azure 0penAI API
- `cursor` - Cursor Protocol Buffers
- `github` - GitHub Copilot API
- `qoder` - Qoder API

**Request Translators** (11):
```
openai → claude
openai → gemini
openai → vertex
openai → cohere
openai → ollama
openai → cursor
claude → openai
gemini → openai
vertex → openai
cohere → openai
cursor → openai
```

**Response Translators** (9):
```
claude → openai
gemini → openai
vertex → openai
cohere → openai
ollama → openai
cursor → openai
openrouter → openai
azure → openai
github → openai
```

**Registry Pattern**:
```javascript
// 9Router
const translators = {
  request: {
    'openai->claude': lazy(() => import('./request/openaiToClaude.js')),
    // ...
  },
  response: {
    'claude->openai': lazy(() => import('./response/claudeToOpenai.js')),
    // ...
  }
};

function translateRequest(body, fromFormat, toFormat) {
  const key = `${fromFormat}->${toFormat}`;
  const translator = translators.request[key];
  return translator ? translator(body) : body;
}
```

**Porting Notes**:
- Each translator is a pure function
- Must handle streaming (SSE) and non-streaming responses
- Must preserve all metadata (usage, model, finish_reason)
- Golden-file tests required for each translator
- See `docs/TRANSLATORS.md` for detailed spec

---

### Executors

**9Router Location**: `_ref/9router/open-sse/executors/`

**Go Port Location**: `internal/executor/`

**Purpose**: Provider-specific HTTP adapters for making upstream API calls.

**Base Class**: `DefaultExecutor` (handles 90% of providers)

**Specialized Executors** (20):

| Executor | File | Provider | Special Handling |
|----------|------|----------|------------------|
| `openai` | `openai.js` | 0penAI | Standard SSE |
| `claude` | `claude.js` | CL4ude | Custom headers, caching |
| `gemini` | `gemini.js` | Google Gemini | API key in URL |
| `vertex` | `vertex.js` | Vertex AI | OAuth2, project ID |
| `azure` | `azure.js` | Azure 0penAI | API version header |
| `ollama` | `ollama-local.js` | Ollama | Local HTTP |
| `cursor` | `cursor.js` | Cursor | Protobuf, checksum |
| `github` | `github.js` | GitHub Copilot | OAuth, special auth |
| `qoder` | `qoder.js` | Qoder | Custom auth |
| `opencode` | `opencode.js` | OpenCode | OAuth flow |
| `commandcode` | `commandcode.js` | CommandCode | Custom headers |
| `grok-web` | `grok-web.js` | Grok Web | Web scraping |
| `openrouter` | `openrouter.js` | OpenRouter | Passthrough |
| `cohere` | `cohere.js` | Cohere | Custom format |
| `minimax` | `minimax.js` | MiniMax | Chinese provider |
| `glm` | `glm.js` | GLM | Chinese provider |
| `deepseek` | `deepseek.js` | DeepSeek | Chinese provider |
| `kiro` | `kiro.js` | Kiro | Free tier |
| `antigravity` | `antigravity.js` | Antigravity | Custom protocol |
| `codex` | `codex.js` | Codex | Legacy support |

**DefaultExecutor Pattern**:
```javascript
class DefaultExecutor {
  async execute(request, provider, account) {
    // 1. Build URL
    const url = this.buildUrl(provider, request);
    
    // 2. Build headers
    const headers = this.buildHeaders(provider, account, request);
    
    // 3. Make request
    const response = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify(request.body),
    });
    
    // 4. Handle streaming
    if (request.stream) {
      return this.streamResponse(response);
    }
    
    // 5. Return JSON
    return response.json();
  }
}
```

**Porting Notes**:
- Start with `DefaultExecutor` - covers most providers
- Add specialized executors only when needed
- Must handle SSE streaming correctly (chunked reads, no buffering)
- Must handle auth (API key, OAuth, custom headers)
- Must handle retries and rate limits
- See `docs/EXECUTORS.md` for detailed spec

---

## 📖 Porting Guide

### Step-by-Step Process

1. **Find the 9Router equivalent**
   - Use this CODEBASE.md to locate the source file
   - Read the JavaScript implementation
   - Understand the data flow and edge cases

2. **Check the spec**
   - Open the relevant `docs/<subsystem>.md` file
   - Verify the spec matches the code (update if needed)
   - Note any parity requirements

3. **Write idiomatic Go**
   - Mirror behavior, not structure
   - Use Go idioms (interfaces, error handling, context)
   - Follow `docs/CONVENTIONS.md`

4. **Add tests**
   - Unit tests for pure functions (translators, RTK filters)
   - Integration tests for HTTP handlers
   - Golden-file tests for data transformations

5. **Verify parity**
   - Run against captured 9router fixtures
   - Compare outputs byte-for-byte
   - Update `docs/PARITY_CHECKLIST.md`

6. **Document**
   - Add `// ref: open-sse/path/to/file.js:LINE` comments
   - Update this CODEBASE.md if structure changes
   - Note any deviations from 9router

### Common Patterns

**9Router Pattern** → **Go Pattern**

```javascript
// 9Router: Lazy-loaded registry
const executors = {
  openai: lazy(() => import('./openai.js')),
};
```
```go
// Go: Interface + factory
type Executor interface {
  Execute(ctx context.Context, req *Request) (*Response, error)
}

func GetExecutor(provider string) Executor {
  if e, ok := executors[provider]; ok {
    return e
  }
  return &DefaultExecutor{}
}
```

---

```javascript
// 9Router: Async streaming
async function* streamResponse(response) {
  for await (const chunk of response.body) {
    yield chunk;
  }
}
```
```go
// Go: io.Reader + http.Flusher
func streamResponse(w http.ResponseWriter, r io.Reader) error {
  flusher := w.(http.Flusher)
  buf := make([]byte, 8192)
  for {
    n, err := r.Read(buf)
    if n > 0 {
      w.Write(buf[:n])
      flusher.Flush()
    }
    if err != nil {
      return err
    }
  }
}
```

---

```javascript
// 9Router: Error handling
try {
  const result = await doSomething();
} catch (err) {
  console.error('Failed:', err);
  throw err;
}
```
```go
// Go: Explicit error returns
result, err := doSomething(ctx)
if err != nil {
  slog.ErrorContext(ctx, "operation failed", "error", err)
  return fmt.Errorf("doSomething: %w", err)
}
```

### Anti-Patterns to Avoid

❌ **Don't copy 9Router's anti-patterns**:
- `db.json` flat-file → Use SQLite from day one
- Manual stream `.write()` → Use `io.Copy`
- `better-sqlite3` fallback → Use `modernc.org/sqlite` (pure Go)
- Unencrypted OAuth tokens → Encrypt with AES-GCM
- Separate `usageDb` → Unify under `DATA_DIR`

❌ **Don't break Go idioms**:
- No panics in request path
- Context everywhere (first arg)
- Errors are values (no exceptions)
- Interfaces over concrete types
- No goroutines without lifecycle

✅ **Do maintain parity**:
- Same env vars
- Same API contracts
- Same error messages
- Same fallback behavior
- Same quirks (if observable)

---

## 🔗 Quick Links

### Documentation
- [AGENTS.md](./AGENTS.md) - Master AI agent context
- [PLAN.md](./PLAN.md) - Vision and scope
- [ROADMAP.md](./ROADMAP.md) - Phased execution plan
- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) - Canonical structure
- [docs/PARITY_CHECKLIST.md](./docs/PARITY_CHECKLIST.md) - Feature tracker
- [docs/CONVENTIONS.md](./docs/CONVENTIONS.md) - Go style guide

### Subsystem Specs
- [docs/RTK_SPEC.md](./docs/RTK_SPEC.md) - Token saver
- [docs/CAVEMAN_SPEC.md](./docs/CAVEMAN_SPEC.md) - Prompt injector
- [docs/TRANSLATORS.md](./docs/TRANSLATORS.md) - Format conversion
- [docs/EXECUTORS.md](./docs/EXECUTORS.md) - Provider adapters
- [docs/API.md](./docs/API.md) - Endpoint contracts
- [docs/AUTH.md](./docs/AUTH.md) - JWT/OAuth
- [docs/DATABASE.md](./docs/DATABASE.md) - Schema
- [docs/STREAMING.md](./docs/STREAMING.md) - SSE
- [docs/FALLBACK.md](./docs/FALLBACK.md) - Combo/account fallback

### Reference Code
- [_ref/9router/](./ref/9router/) - 9Router source (read-only)
- [_ref/9router/open-sse/handlers/chatCore.js](./_ref/9router/open-sse/handlers/chatCore.js) - Main orchestration
- [_ref/9router/open-sse/executors/](./_ref/9router/open-sse/executors/) - Executors
- [_ref/9router/open-sse/translator/](./_ref/9router/open-sse/translator/) - Translators
- [_ref/9router/open-sse/rtk/](./_ref/9router/open-sse/rtk/) - RTK filters

---

**Last Updated**: 2026-05-11 16:12 UTC  
**Generated By**: Parallel exploration (5 agents: librarian + 4x explore)  
**Maintainer**: Update this file when structure changes
