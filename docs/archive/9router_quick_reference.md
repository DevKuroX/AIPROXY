# 9Router Quick Reference Guide

## Project Overview

**9Router** is a local AI routing gateway that connects CLI tools (Code Assistant, Cursor, Codex, Cline) to 40+ AI providers with intelligent fallback, token optimization, and format translation.

- **Repository**: [decolua/9router](https://github.com/decolua/9router)
- **Latest**: v0.4.31 (2026-05-12)
- **Language**: JavaScript (99.5%)
- **License**: MIT
- **Stars**: 8,920 | **Forks**: 1,387

---

## Architecture at a Glance

```
CLI Tools → 9Router (localhost:20128/v1) → 40+ Providers
                ↓
        ┌─────────────────────┐
        │ Format Translation  │
        │ RTK Compression     │
        │ Fallback Logic      │
        │ Token Refresh       │
        │ MITM Proxy          │
        └─────────────────────┘
```

---

## Core Modules

| Module | Purpose | Key Files |
|--------|---------|-----------|
| **open-sse/** | Provider-agnostic routing engine | config/, executors/, handlers/, services/, translator/, rtk/, utils/ |
| **src/app/api/** | 0penAI-compatible API endpoints | v1/, v1beta/, management APIs |
| **src/app/(dashboard)/** | React dashboard UI | providers/, combos/, cli-tools/, mitm/ |
| **src/lib/db/** | Database layer (4 adapters) | driver.js, repos/, adapters/, migrations/ |
| **src/lib/oauth/** | OAuth implementation (13 providers) | services/, utils/ |
| **src/mitm/** | MITM proxy system | server.js, handlers/, cert/, dns/ |

---

## Key Features

### 1. Multi-Provider Routing
- **40+ providers** (OAuth, API key, compatible nodes)
- **Intelligent fallback** (Tier 1: subscription → Tier 2: cheap → Tier 3: free)
- **Account-level fallback** (round-robin between accounts)

### 2. Format Translation
- **9 formats**: 0penAI, CL4ude, Gemini, Responses API, Cursor, Antigravity, Kiro, CommandCode, Ollama
- **Bidirectional translation** (request & response)
- **Tool call preservation** (function calling)

### 3. Token Optimization (RTK)
- **20-40% token savings** via compression
- **Caveman algorithm** for smart compression
- **10 filters** (grep, ls, git diff, tree, etc.)
- **Smart truncation** & deduplication

### 4. OAuth & Token Management
- **13 OAuth providers** with auto-refresh
- **Expiry detection** & buffer
- **Fallback on failure**
- **PKCE flow** support

### 5. MITM Proxy System
- **Transparent interception** (HTTP/HTTPS)
- **Self-signed certificates** (auto-generated)
- **Provider-specific handlers** (token rotation, request modification)
- **DNS configuration** support

### 6. Media Providers
- **Images**: 14 providers (Gemini, DALL-E, Codex, etc.)
- **TTS**: 10 providers (Google, 0penAI, Elevenlabs, etc.)
- **Embeddings**: 5 providers (0penAI, Gemini, etc.)
- **STT**: Speech-to-text support

### 7. Database Layer
- **4 adapters**: better-sqlite3, sql.js, Node.js, Bun
- **10 repositories**: connections, keys, combos, aliases, settings, pricing, logs, disabled models, proxy pools, nodes
- **Automatic migrations** & schema management

---

## Provider Support

### OAuth Providers (Auto Token Refresh)
CL4ude, Gemini, Codex, Cursor, GitHub Copilot, Antigravity, Kiro, iFlow, Qwen, OpenCode

### API Key Providers
0penAI, OpenRouter, GLM, Kimi, MiniMax, Azure 0penAI

### Compatible Nodes
Ollama (local), 0penAI-compatible endpoints

### Web Scrapers
Grok, Perplexity

---

## API Endpoints

### 0penAI-Compatible (v1/)
```
POST /v1/chat/completions          → Chat completion
POST /v1/embeddings                → Embeddings
POST /v1/images/generations        → Image generation
POST /v1/audio/speech              → Text-to-speech
POST /v1/audio/transcriptions      → Speech-to-text
GET  /v1/models                    → List models
```

### Management APIs
```
/api/providers/                    → Provider CRUD
/api/oauth/[provider]/[action]     → OAuth flow
/api/keys/                         → API key management
/api/combos/                       → Model combo management
/api/settings/                     → User settings
/api/usage/                        → Usage statistics
/api/models/                       → Model management
/api/media-providers/              → Image/TTS/embedding config
/api/proxy-pools/                  → Proxy pool management
/api/cloud/                        → Cloud sync
/api/auth/                         → Authentication
/api/health/                       → Health check
/api/version/                      → Version info
```

---

## File Organization

### open-sse/ (Provider-Agnostic)
```
open-sse/
├── config/              # Provider configs, models, constants
├── executors/           # 20 provider-specific executors
├── handlers/            # Chat, embeddings, images, TTS, STT, search
├── services/            # Fallback, token refresh, usage
├── translator/          # Format translation (9 formats)
├── rtk/                 # Token compression & filtering
├── utils/               # Stream, error, proxy, caching
└── index.js             # Public API exports
```

### src/ (Next.js Application)
```
src/
├── app/
│   ├── api/             # API routes (v1, management)
│   ├── (dashboard)/     # Dashboard UI pages
│   └── layout.js
├── lib/
│   ├── db/              # Database layer (4 adapters, 10 repos)
│   ├── oauth/           # OAuth (13 providers)
│   ├── network/         # Proxy utilities
│   └── ...
├── mitm/                # MITM proxy system
├── sse/                 # SSE handlers
├── shared/              # Shared utilities
└── store/               # Zustand state
```

---

## Database Schema

### Repositories (10 data types)
- **connections** - Provider OAuth/API key connections
- **apiKeys** - Stored API keys
- **combos** - Model fallback sequences
- **aliases** - Model name aliases
- **settings** - User settings & preferences
- **pricing** - Model pricing data
- **requestDetails** - Request logs & history
- **disabledModels** - Disabled model list
- **proxyPools** - Proxy pool configurations
- **nodes** - Compatible node endpoints

### Adapters (4 database backends)
- **better-sqlite3** - Primary (fastest, requires build tools)
- **sql.js** - Fallback (pure JS, no build required)
- **Node.js sqlite** - Alternative
- **Bun sqlite** - Bun runtime support

---

## Technology Stack

### Frontend
- Next.js 16.1.6
- React 19.2.4
- Tailwind CSS 4
- Monaco Editor
- Recharts
- Zustand (state)

### Backend
- Next.js API routes
- Express 5.2.1
- Node.js / Bun
- SQLite (multiple adapters)
- node-forge (certificates)
- http-proxy-middleware
- jose (JWT)
- bcryptjs (hashing)
- socks-proxy-agent

---

## Development

### Commands
```bash
npm run dev              # Dev server (port 20128)
npm run build            # Production build
npm run start            # Start production
npm run dev:bun         # Bun dev
npm run build:bun       # Bun build
npm run start:bun       # Bun start
```

### Docker
```bash
docker build -t 9router .
docker run -p 20128:20128 9router
```

### Testing
```bash
npm test                # Run tests (Vitest)
```

---

## Key Concepts

### Provider Format
Each provider has a format (claude, openai, gemini, etc.) that defines:
- Request structure
- Response structure
- Header requirements
- Authentication method

### Model Combo
A sequence of models to try in order:
1. Primary model (subscription)
2. Fallback model (cheap)
3. Free model (unlimited)

### Account Fallback
Multiple accounts per provider for:
- Quota rotation
- Rate limit avoidance
- Multi-user support

### Token Refresh (RTK)
Automatic OAuth token refresh with:
- Expiry detection
- Refresh logic per provider
- Fallback on failure

### MITM Proxy
Transparent interception for:
- Token rotation
- Request modification
- Response caching
- Provider-specific handling

---

## Common Tasks

### Add a New Provider
1. Create executor: `open-sse/executors/[provider].js`
2. Add config: `open-sse/config/providers.js`
3. Implement OAuth: `src/lib/oauth/services/[provider].js`
4. Add translators: `open-sse/translator/request/` & `response/`
5. Add tests

### Add RTK Compression Filter
1. Create filter: `open-sse/rtk/filters/[name].js`
2. Register: `open-sse/rtk/registry.js`
3. Add tests

### Add API Endpoint
1. Create route: `src/app/api/[path]/route.js`
2. Use database repos
3. Handle errors
4. Add tests

### Add Dashboard Page
1. Create page: `src/app/(dashboard)/dashboard/[page]/page.js`
2. Use React components
3. Call management APIs
4. Use Zustand for state

### Add Database Repository
1. Create repo: `src/lib/db/repos/[entity]Repo.js`
2. Implement CRUD
3. Support all 4 adapters
4. Add migration if needed

---

## Statistics

- **Total LOC**: ~17,767
- **JavaScript Files**: 82+
- **Providers**: 40+
- **Image Providers**: 14
- **TTS Providers**: 10
- **Embedding Providers**: 5
- **Test Files**: 25+
- **Database Adapters**: 4
- **OAuth Providers**: 13
- **Format Translators**: 9
- **RTK Filters**: 10
- **Contributors**: 90+

---

## Important Files

### Core Routing
- `open-sse/handlers/chatCore.js` - Main chat handler
- `open-sse/services/provider.js` - Provider config lookup
- `open-sse/services/accountFallback.js` - Fallback logic
- `open-sse/translator/index.js` - Format translation

### Providers
- `open-sse/executors/index.js` - Executor registry
- `open-sse/config/providers.js` - Provider definitions
- `open-sse/config/providerModels.js` - Model mapping

### Token Optimization
- `open-sse/rtk/index.js` - RTK orchestration
- `open-sse/rtk/caveman.js` - Compression algorithm
- `open-sse/rtk/registry.js` - Filter registry

### Database
- `src/lib/db/driver.js` - Database driver
- `src/lib/db/repos/` - Data repositories
- `src/lib/db/adapters/` - Database adapters

### OAuth
- `src/lib/oauth/services/` - OAuth implementations
- `open-sse/services/tokenRefresh.js` - Token refresh

### MITM
- `src/mitm/server.js` - MITM proxy server
- `src/mitm/handlers/` - Provider-specific handlers
- `src/mitm/cert/` - Certificate management

### API
- `src/app/api/v1/` - 0penAI-compatible endpoints
- `src/app/api/providers/` - Provider management
- `src/app/api/oauth/` - OAuth flow

### Dashboard
- `src/app/(dashboard)/dashboard/` - Dashboard pages
- `src/store/` - Zustand state

---

## Useful Links

- **GitHub**: https://github.com/decolua/9router
- **Website**: https://9router.com
- **Issues**: https://github.com/decolua/9router/issues
- **Architecture Docs**: https://github.com/decolua/9router/blob/master/docs/ARCHITECTURE.md

---

## Notes for Phase Prompts

1. **Always consider all 40+ providers** when implementing features
2. **Always consider all 9 formats** for translation
3. **Always respect fallback tiers** (subscription → cheap → free)
4. **Always apply RTK compression** before sending requests
5. **Always support all 4 database adapters**
6. **Always handle errors gracefully** with fallback
7. **Always test with multiple providers** (OAuth, API key, compatible node)
8. **Always support streaming** (SSE) and non-streaming (JSON)
9. **Always implement OAuth token refresh** for OAuth providers
10. **Always document API endpoints** with request/response format

