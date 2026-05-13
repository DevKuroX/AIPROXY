# 9Router - Comprehensive Documentation Summary

**Current Year**: 2026  
**Latest Version**: v0.4.29 (May 2026)  
**Repository**: https://github.com/decolua/9router  
**Website**: https://9router.com  
**License**: MIT

---

## Executive Overview

9Router is a **local AI routing gateway and dashboard** built on Next.js that provides:

- Single 0penAI-compatible endpoint (`http://localhost:20128/v1`) for all CLI tools
- Intelligent 3-tier fallback routing (Subscription → Cheap → Free)
- RTK Token Saver: 20-40% input token compression for tool outputs
- Real-time quota tracking and usage monitoring
- Multi-provider support (40+ providers, 100+ models)
- Format translation across provider APIs
- OAuth + API key provider management
- Local persistence with optional cloud sync

**Key Value Proposition**: Never stop coding. Save 20-40% tokens with RTK + auto-fallback to FREE & cheap AI models.

---

## Architecture Overview

### System Context

```
┌─────────────────────────────────────────────────────────────┐
│ Developer Clients                                           │
│ • Code Assistant, Codex, Cursor, Cline, OpenClaw, Droid   │
│ • Browser Dashboard                                         │
└────────────────────┬────────────────────────────────────────┘
                     │ http://localhost:20128/v1
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 9Router Local Process (Next.js)                             │
│ ├─ V1 Compatibility API (/v1/*)                            │
│ ├─ Dashboard + Management API (/api/*)                     │
│ ├─ SSE + Translation Core (open-sse + src/sse)             │
│ ├─ Local DB (db.json, usage.json, log.txt)                 │
│ └─ RTK Token Saver + Caveman Mode                          │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
    [Tier 1]    [Tier 2]    [Tier 3]
    SUBSCRIPTION CHEAP      FREE
    • Code      • GLM       • Kiro
      Assistant • MiniMax   • OpenCode
    • Codex     • Kimi      • Vertex
    • Copilot
```

### Core Components

#### 1. **API & Routing Layer** (`src/app/api/*`)
- **Compatibility APIs**: `/v1/chat/completions`, `/v1/messages`, `/v1/models`, `/v1/count_tokens`
- **Management APIs**: `/api/providers/*`, `/api/keys/*`, `/api/combos/*`, `/api/usage/*`
- **Auth**: `/api/auth/*`, `/api/oauth/*`
- **Settings**: `/api/settings/*`, `/api/pricing`
- **Cloud Sync**: `/api/sync/*`, `/api/cloud/*`

#### 2. **SSE + Translation Core** (`open-sse/` + `src/sse/`)
- **Entry Point**: `src/sse/handlers/chat.js` (request parsing, combo handling)
- **Core Orchestration**: `open-sse/handlers/chatCore.js` (translation, executor dispatch, retry/refresh)
- **Provider Executors**: `open-sse/executors/*` (provider-specific network & format behavior)
- **Format Detection**: `open-sse/services/provider.js`
- **Model Resolution**: `src/sse/services/model.js`, `open-sse/services/model.js`
- **Account Fallback**: `open-sse/services/accountFallback.js`
- **Translation Registry**: `open-sse/translator/index.js`
- **Stream Utilities**: `open-sse/utils/stream.js`, `open-sse/utils/streamHandler.js`
- **Usage Tracking**: `open-sse/utils/usageTracking.js`

#### 3. **RTK Token Saver** (`open-sse/rtk/`)
- **Compression Engine**: `open-sse/rtk/index.js`
- **Auto-Detection**: `open-sse/rtk/autodetect.js` (detects tool output format)
- **Filters**: `open-sse/rtk/filters/` (git-diff, grep, find, ls, tree, dedup-log, smart-truncate, etc.)
- **Caveman Mode**: `open-sse/rtk/caveman.js` (aggressive output token compression)

#### 4. **Data Persistence** (`src/lib/`)
- **Config DB**: `src/lib/localDb.js` (providers, keys, aliases, combos, settings)
- **Usage DB**: `src/lib/usageDb.js` (usage history, request logs)
- **Storage**: `~/.9router/db.json`, `~/.9router/usage.json`, `~/.9router/log.txt`

---

## Key Features

### 1. RTK Token Saver (20-40% Compression)

**What It Does**: Automatically compresses tool outputs before sending to LLM.

**Supported Filters**:
- `git-diff` - Removes redundant hunk headers, collapses context
- `git-status` - Removes file paths, keeps only status
- `grep` - Removes redundant file paths, collapses matches
- `find` - Removes redundant directory prefixes
- `ls` - Removes redundant directory info
- `tree` - Collapses deep nesting
- `dedup-log` - Removes duplicate log lines
- `smart-truncate` - Intelligent truncation with context preservation
- `read-numbered` - Removes line numbers from code
- `search-list` - Deduplicates search results

**How It Works**:
1. Detects tool output format from first 1KB
2. Applies appropriate filter
3. If filter fails or makes output bigger, keeps original
4. Runs before format translation (works across all formats)
5. Default: ON (toggle in Dashboard → Endpoint Settings)

**Example**:
```
Without RTK: 47K tokens sent to LLM
With RTK:    28K tokens sent to LLM (40% saved, same context, same answer)
```

**Caveman Mode** (Optional, more aggressive):
- Rewrites output in simplified "caveman-speak"
- Can save up to 65% output tokens
- May reduce quality for strict system prompts
- Default: OFF

### 2. Smart 3-Tier Fallback Routing

**Tier 1: SUBSCRIPTION (Primary)**
- Code Assistant (Pro/Max)
- 0penAI Codex (Plus/Pro)
- Gemini CLI (FREE 180K/month)
- GitHub Copilot
- Antigravity (Google)

**Tier 2: CHEAP (Backup)**
- GLM-4.7 ($0.60/1M input)
- MiniMax M2.1 ($0.20/1M input)
- Kimi K2 ($9/month flat)

**Tier 3: FREE (Emergency)**
- iFlow (8 models)
- Qwen (3 models)
- Kiro (CL4ude FREE)
- OpenCode Free
- Vertex AI ($300 credits)

**Fallback Triggers**:
- 429 (quota exhausted)
- 401/403 (auth failed, retry after refresh)
- 503 (provider unavailable)
- Rate limiting detected
- Custom error heuristics per provider

**Decision Logic**:
1. Check quota availability
2. Prefer subscription → cheap → free
3. Consider reset timing
4. Skip unhealthy providers
5. Round-robin between accounts per provider

### 3. Combos - Custom Fallback Chains

**What Are Combos**: User-defined sequences of models that 9Router tries in order.

**Example**:
```
Combo: premium-coding
├─ cc/claude-opus-4-5-20251101 (try first)
├─ glm/glm-4.7 (if #1 quota exhausted)
└─ minimax/MiniMax-M2.1 (if #2 quota exhausted)
```

**Use Cases**:
- Maximize subscription value
- Minimize costs
- Ensure 24/7 availability
- Optimize for quality

**Creation**: Dashboard → Combos → Create New Combo → Select models → Save

### 4. Real-Time Quota Tracking

**Tracked Metrics**:
- Real-time token consumption per request
- Quota limits & remaining
- Reset countdown timers
- Cost estimation
- Monthly reports
- Usage analytics

**Dashboard Views**:
- Quota Overview (all providers)
- Per-request token breakdown
- Live usage graphs
- Cost trends
- Monthly summaries

### 5. Format Translation

**Supported Formats**:
- 0penAI (chat/completions, embeddings, TTS, STT, image, video)
- CL4ude (messages, responses)
- Gemini (generateContent)
- Cursor (proprietary)
- Kiro (proprietary)
- Vertex AI (proprietary)
- Custom 0penAI-compatible endpoints

**Translation Flow**:
1. Detect source format from request
2. Translate to provider's native format
3. Execute request
4. Translate response back to source format
5. Stream to client

### 6. OAuth + Token Refresh

**Supported OAuth Providers**:
- CL4ude-Code
- Antigravity
- Codex
- GitHub
- Cursor

**Token Refresh**:
- Automatic refresh before expiry
- In-flight request caching (prevent race conditions)
- Unrecoverable error handling
- Per-provider refresh strategies

### 7. Multi-Account Support

**Features**:
- Multiple accounts per provider
- Round-robin account rotation
- Per-account quota tracking
- Account-level fallback
- Priority ordering

### 8. Provider Executors (17 Providers)

**Implemented Executors**:
- `antigravity.js` (17.7K) - Google Antigravity
- `azure.js` - Azure 0penAI
- `codex.js` (8.6K) - 0penAI Codex
- `commandcode.js` - CommandCode
- `cursor.js` (22.4K) - Cursor IDE
- `default.js` (13.6K) - Generic 0penAI-compatible
- `gemini-cli.js` - Gemini CLI
- `github.js` (14.3K) - GitHub Copilot
- `grok-web.js` (14.7K) - Grok Web
- `iflow.js` - iFlow
- `kiro.js` (15.6K) - Kiro AI
- `ollama-local.js` - Ollama Local
- `opencode-go.js` - OpenCode Go
- `opencode.js` - OpenCode
- `perplexity-web.js` (19.0K) - Perplexity Web

**Executor Responsibilities**:
- Provider-specific authentication
- Request formatting
- Response parsing
- Error handling
- Rate limit detection
- Token refresh logic

---

## Data Model & Storage

### Connection Schema
```javascript
CONNECTION {
  id: string,
  provider: string,
  name: string,
  priority: number,
  isActive: boolean,
  apiKey: string,
  accessToken: string,
  refreshToken: string,
  expiresAt: string,
  testStatus: string,
  lastError: string,
  rateLimitedUntil: string,
  providerSpecificData: json
}
```

### Combo Schema
```javascript
COMBO {
  id: string,
  name: string,
  models: [string],
  description: string,
  createdAt: timestamp
}
```

### Usage Record Schema
```javascript
USAGE_RECORD {
  id: string,
  timestamp: timestamp,
  provider: string,
  model: string,
  inputTokens: number,
  outputTokens: number,
  totalTokens: number,
  cost: number,
  duration: number,
  status: string
}
```

### Storage Locations
- **Main Config**: `${DATA_DIR}/db.json` or `~/.9router/db.json`
- **Usage Stats**: `~/.9router/usage.json`
- **Request Logs**: `~/.9router/log.txt`
- **Translator Debug**: `~/.9router/logs/...`

---

## API Endpoints

### Compatibility APIs (`/v1/*`)

#### Chat Completions
```
POST /v1/chat/completions
Content-Type: application/json

{
  "model": "cc/claude-opus-4-5",
  "messages": [...],
  "stream": true,
  "temperature": 0.7
}
```

#### Messages (CL4ude format)
```
POST /v1/messages
Content-Type: application/json

{
  "model": "cc/claude-opus-4-5",
  "messages": [...],
  "max_tokens": 2048
}
```

#### Models List
```
GET /v1/models
```

Returns all available models across all connected providers.

#### Token Counting
```
POST /v1/messages/count_tokens
Content-Type: application/json

{
  "model": "cc/claude-opus-4-5",
  "messages": [...]
}
```

### Management APIs (`/api/*`)

#### Providers
```
GET /api/providers                    # List all providers
POST /api/providers                   # Create new provider connection
GET /api/providers/:id                # Get provider details
PUT /api/providers/:id                # Update provider
DELETE /api/providers/:id             # Delete provider
POST /api/providers/:id/test          # Test provider connection
```

#### Combos
```
GET /api/combos                       # List all combos
POST /api/combos                      # Create new combo
GET /api/combos/:id                   # Get combo details
PUT /api/combos/:id                   # Update combo
DELETE /api/combos/:id                # Delete combo
```

#### Usage & Analytics
```
GET /api/usage/summary                # Overall usage summary
GET /api/usage/daily                  # Daily breakdown
GET /api/usage/monthly                # Monthly breakdown
GET /api/usage/by-provider            # Usage by provider
GET /api/usage/by-model               # Usage by model
GET /api/usage/requests               # Recent requests
```

#### Settings
```
GET /api/settings                     # Get all settings
PUT /api/settings                     # Update settings
GET /api/settings/endpoint            # Endpoint-specific settings
```

#### Auth
```
POST /api/auth/login                  # Login with password
POST /api/auth/logout                 # Logout
GET /api/auth/status                  # Check auth status
POST /api/auth/password               # Change password
```

---

## Request/Response Flow

### Typical Request Lifecycle

```
1. Client Request
   ↓
2. Parse & Validate (src/sse/handlers/chat.js)
   ├─ Detect source format
   ├─ Extract model/combo
   ├─ Resolve to provider + model
   ↓
3. RTK Compression (open-sse/rtk/index.js)
   ├─ Auto-detect tool output format
   ├─ Apply appropriate filter
   ├─ Log compression stats
   ↓
4. Format Translation (open-sse/translator/index.js)
   ├─ Translate request to provider format
   ├─ Inject provider-specific headers
   ↓
5. Executor Dispatch (open-sse/executors/*)
   ├─ Get provider executor
   ├─ Check credentials (refresh if needed)
   ├─ Execute request
   ↓
6. Response Handling
   ├─ Parse provider response
   ├─ Extract usage metrics
   ├─ Translate back to source format
   ├─ Stream to client
   ↓
7. Usage Tracking (open-sse/utils/usageTracking.js)
   ├─ Record tokens used
   ├─ Calculate cost
   ├─ Update quota
   ├─ Log request details
```

### Fallback Flow (on Error)

```
1. Request fails (429, 401, 503, etc.)
   ↓
2. Parse error (open-sse/utils/error.js)
   ├─ Determine error type
   ├─ Extract retry-after if available
   ↓
3. Fallback Decision (open-sse/services/accountFallback.js)
   ├─ Check if retryable
   ├─ Select next provider/account
   ├─ Check quota availability
   ↓
4. Retry Request
   ├─ Translate to new provider format
   ├─ Execute with new credentials
   ↓
5. Success or Continue Fallback
```

---

## Configuration & Environment

### Environment Variables

```bash
# Server
PORT=20128                              # Server port
HOSTNAME=0.0.0.0                        # Bind address
NODE_ENV=production                     # Environment

# Dashboard
NEXT_PUBLIC_BASE_URL=http://localhost:20128

# Cloud Sync (optional)
NEXT_PUBLIC_CLOUD_URL=https://cloud.9router.com

# Data Directory
DATA_DIR=~/.9router                     # Config storage location

# Logging
DEBUG=*                                 # Enable debug logging
LOG_LEVEL=info                          # Log level
```

### Configuration Files

**`.env.example`**:
```bash
PORT=20128
HOSTNAME=0.0.0.0
NEXT_PUBLIC_BASE_URL=http://localhost:20128
```

**`db.json`** (auto-created):
```json
{
  "connections": [...],
  "combos": [...],
  "aliases": [...],
  "settings": {...},
  "pricing": {...}
}
```

---

## Deployment Options

### Local Development
```bash
npm install
PORT=20128 NEXT_PUBLIC_BASE_URL=http://localhost:20128 npm run dev
```

### Production (npm)
```bash
npm install -g 9router
9router
```

### Production (from source)
```bash
npm run build
PORT=20128 HOSTNAME=0.0.0.0 NEXT_PUBLIC_BASE_URL=http://localhost:20128 npm run start
```

### Docker
```bash
docker build -t 9router .
docker run -p 20128:20128 -v ~/.9router:/root/.9router 9router
```

### Cloudflare Workers
- Supported via `open-sse` module
- Requires custom deployment wrapper

### Tunneling
- Tailscale integration
- Custom tunnel support
- MITM proxy for IDE subscriptions

---

## Known Quirks & Limitations

### Documented Issues

1. **usageDb Storage**: Currently stores under `~/.9router` and does not follow `DATA_DIR` env var
2. **Static Model List**: `/api/v1/route.js` returns static model list, not the main source used by `/v1/models`
3. **Request Logging**: Full headers/body logged when enabled; treat log directory as sensitive
4. **Cloud Sync**: Depends on correct `NEXT_PUBLIC_BASE_URL` and cloud endpoint reachability
5. **RTK Markdown Tables**: May misalign column formatting in Markdown tables and ASCII art
6. **Caveman Mode**: Can reduce quality for strict system prompts; test before enabling

### Provider Discontinuations (2026)

- ❌ **iFlow**: Was free unlimited, now changed to paid
- ❌ **Qwen**: Free tier discontinued
- ❌ **Gemini CLI**: Free tier discontinued

**Recommended Alternatives**: Kiro, OpenCode Free, Vertex AI

---

## Technology Stack

### Backend
- **Runtime**: Node.js
- **Framework**: Next.js 16.1.6 (App Router)
- **Database**: LowDB (JSON) + optional better-sqlite3
- **Auth**: JWT (jose), bcryptjs
- **Streaming**: Native Node.js streams, http-proxy-middleware

### Frontend
- **Framework**: React 19.2.4
- **Styling**: Tailwind CSS 4
- **Charts**: Recharts 3.7.0
- **State**: Zustand 5.0.10
- **Editor**: Monaco Editor 0.55.1
- **UI Flow**: XYFlow 12.10.1

### Key Dependencies
- `express` - HTTP server
- `http-proxy-middleware` - Request proxying
- `node-machine-id` - Machine identification
- `socks-proxy-agent` - Proxy support
- `sql.js` - SQLite fallback
- `undici` - HTTP client

---

## Community & Support

- **Website**: https://9router.com
- **GitHub**: https://github.com/decolua/9router
- **Issues**: https://github.com/decolua/9router/issues
- **Discussions**: https://github.com/decolua/9router/discussions
- **NPM**: https://www.npmjs.com/package/9router

### Version History

| Version Range | Period | Releases | Focus |
|---|---|---|---|
| 0.4.1 - 0.4.29 | Apr - May 2026 | 23 | SQLite migration, MCP, Cloudflare AI |
| 0.3.3 - 0.3.99 | Feb - Apr 2026 | 78 | Provider expansion, MITM, OAuth |
| 0.2.13 - 0.2.98 | Jan - Feb 2026 | 79 | Foundation, RTK, Caveman |

---

## References

### Official Documentation
- **README**: https://github.com/decolua/9router/blob/master/README.md
- **Architecture**: https://github.com/decolua/9router/blob/master/docs/ARCHITECTURE.md
- **Changelog**: https://github.com/decolua/9router/blob/master/CHANGELOG.md
- **Docker**: https://github.com/decolua/9router/blob/master/DOCKER.md

### Feature Guides (GitBook)
- **Quota Tracking**: gitbook/content/en/features/quota-tracking.md
- **Smart Routing**: gitbook/content/en/features/smart-routing.md
- **Combos**: gitbook/content/en/features/combos.md

### Source Code Structure
- **API Routes**: `src/app/api/*`
- **SSE Core**: `src/sse/handlers/chat.js`
- **Translation**: `open-sse/translator/index.js`
- **Executors**: `open-sse/executors/*`
- **RTK**: `open-sse/rtk/index.js`
- **Utilities**: `open-sse/utils/*`

---

**Document Generated**: May 2026  
**9Router Version**: v0.4.29  
**Status**: Active Development
