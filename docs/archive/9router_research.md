# 9Router Project Structure & Architecture Research

**Repository**: [decolua/9router](https://github.com/decolua/9router)  
**Latest Release**: v0.4.31 (2026-05-12)  
**Language**: JavaScript (99.5%)  
**License**: MIT  
**Stars**: 8,920 | **Forks**: 1,387  
**Total LOC**: ~17,767 lines

---

## Executive Summary

9Router is a **local AI routing gateway and dashboard** built on Next.js that provides:
- Single OpenAI-compatible endpoint (`/v1/*`) for CLI tools
- Multi-provider routing with intelligent fallback
- Token refresh (RTK) for 20-40% token savings
- MITM proxy system for transparent provider interception
- Web dashboard for management and monitoring

**Core Architecture**:
- **Frontend**: Next.js 16 React dashboard
- **Backend**: Next.js API routes + Express middleware
- **Routing Core**: `open-sse/` module (provider-agnostic SSE handling)
- **Proxy System**: `src/mitm/` (MITM certificate generation & interception)
- **Database**: SQLite (better-sqlite3 or sql.js fallback)

---

## Directory Structure

```
9router/
├── open-sse/                    # Core routing & translation engine (npm-publishable)
│   ├── config/                  # Provider configs, models, constants
│   ├── executors/               # Provider-specific request executors (20 providers)
│   ├── handlers/                # Core handlers (chat, embeddings, images, TTS, STT)
│   ├── services/                # Business logic (fallback, token refresh, usage)
│   ├── translator/              # Request/response format translation
│   ├── rtk/                      # Token refresh & compression logic
│   ├── utils/                   # Utilities (stream, error, proxy, caching)
│   └── index.js                 # Public API exports
│
├── src/                         # Next.js application
│   ├── app/                     # Next.js app router
│   │   ├── api/                 # API routes (v1, v1beta, management)
│   │   ├── (dashboard)/         # Dashboard UI pages
│   │   └── layout.js
│   ├── lib/                     # Shared libraries
│   │   ├── db/                  # Database layer (SQLite adapters, repos)
│   │   ├── oauth/               # OAuth provider implementations
│   │   ├── network/             # Proxy & network utilities
│   │   └── ...
│   ├── mitm/                    # MITM proxy system
│   │   ├── cert/                # Certificate generation & installation
│   │   ├── handlers/            # Provider-specific MITM handlers
│   │   ├── dns/                 # DNS configuration
│   │   ├── server.js            # MITM proxy server
│   │   └── manager.js           # MITM lifecycle management
│   ├── sse/                     # SSE-specific handlers
│   ├── shared/                  # Shared utilities
│   └── store/                   # Zustand state management
│
├── cloud/                       # Cloud sync module (optional)
├── tests/                       # Unit & E2E tests
├── docs/                        # Documentation
├── scripts/                     # Build & utility scripts
└── public/                      # Static assets
```

---

## Core Modules

### 1. **open-sse/** - Routing Engine (Provider-Agnostic)

**Purpose**: Handles all provider communication, format translation, and streaming.

#### **config/**
- `providers.js` - Provider definitions (baseUrl, format, headers, OAuth config)
- `providerModels.js` - Model availability per provider
- `models.js` - Model metadata
- `constants.js` - Runtime constants
- `appConstants.js` - OAuth endpoints, system prompts
- `runtimeConfig.js` - Cache TTL, backoff, cooldown settings
- `errorConfig.js` - Error handling rules
- `ttsModels.js` - TTS model configurations
- `googleTtsLanguages.js` - Language mappings

#### **executors/** (20 provider implementations)
Each executor handles provider-specific request/response logic:
- `base.js` - Abstract base executor
- `default.js` - OpenAI-compatible fallback
- `claude.js` - Anthropic Claude
- `gemini.js` - Google Gemini
- `codex.js` - OpenAI Codex
- `cursor.js` - Cursor IDE
- `antigravity.js` - Antigravity provider
- `kiro.js` - Kiro provider
- `qwen.js` - Alibaba Qwen
- `vertex.js` - Google Vertex AI
- `azure.js` - Azure OpenAI
- `ollama-local.js` - Local Ollama
- `opencode.js` - OpenCode provider
- `github.js` - GitHub provider
- `iflow.js` - iFlow provider
- `grok-web.js` - Grok web scraper
- `perplexity-web.js` - Perplexity web scraper
- `commandcode.js` - CommandCode provider
- `qoder.js` - Qoder provider

#### **handlers/**
- `chatCore.js` - Main chat request handler
  - `chatCore/streamingHandler.js` - SSE streaming
  - `chatCore/nonStreamingHandler.js` - JSON responses
  - `chatCore/sseToJsonHandler.js` - SSE→JSON conversion
  - `chatCore/requestDetail.js` - Request metadata
- `embeddingsCore.js` - Embedding generation
  - `embeddingProviders/` - Provider implementations (OpenAI, Gemini, etc.)
- `imageGenerationCore.js` - Image generation
  - `imageProviders/` - 14 image provider implementations
- `ttsCore.js` - Text-to-speech
  - `ttsProviders/` - 10 TTS provider implementations
- `sttCore.js` - Speech-to-text
- `responsesHandler.js` - OpenAI Responses API
- `search/` - Web search integration

#### **translator/**
Format translation between providers:
- `formats.js` - Format definitions (claude, openai, gemini, etc.)
- `index.js` - Translation registry & orchestration
- `request/` - Request translators (12 formats)
  - `openai-to-claude.js`, `claude-to-openai.js`, etc.
- `response/` - Response translators (9 formats)
- `helpers/` - Translation utilities
  - `claudeHelper.js`, `geminiHelper.js`, `openaiHelper.js`, etc.

#### **services/**
- `provider.js` - Provider config lookup, URL building, header generation
- `model.js` - Model parsing and resolution
- `accountFallback.js` - Multi-account fallback logic
- `tokenRefresh.js` - OAuth token refresh for all providers
- `combo.js` - Model combo (fallback sequence) management
- `usage.js` - Usage tracking and cost calculation
- `projectId.js` - Project ID management
- `compact.js` - Compression utilities

#### **rtk/** - Token Refresh & Compression
- `index.js` - RTK orchestration
- `autodetect.js` - Auto-detect tool output patterns
- `caveman.js` - Caveman compression algorithm
- `cavemanPrompts.js` - Compression prompts
- `applyFilter.js` - Filter application
- `registry.js` - Filter registry
- `filters/` - 10 compression filters (grep, ls, git diff, etc.)

#### **utils/**
- `proxyFetch.js` - Fetch with proxy support
- `streamHandler.js` - Stream creation & management
- `stream.js` - SSE stream transformation
- `error.js` - Error formatting
- `clientDetector.js` - Client type detection
- `claudeHeaderCache.js` - Claude header caching
- `cursorProtobuf.js` - Cursor protobuf handling
- `sessionManager.js` - Session management
- `usageTracking.js` - Usage tracking
- `reasoningContentInjector.js` - Reasoning content injection
- `bypassHandler.js` - Bypass logic
- `ollamaTransform.js` - Ollama response transformation

---

### 2. **src/app/api/** - API Endpoints

#### **v1/** - OpenAI-Compatible API
- `chat/completions` - Chat completion endpoint
- `embeddings` - Embedding endpoint
- `images/generations` - Image generation
- `audio/speech` - TTS endpoint
- `audio/transcriptions` - STT endpoint
- `models` - List available models

#### **v1beta/** - Beta endpoints
- Extended/experimental features

#### **Management APIs**
- `providers/` - Provider CRUD
  - `[id]/` - Provider details
  - `[id]/models/` - Provider models
  - `[id]/test/` - Test provider connection
- `oauth/[provider]/[action]` - OAuth flow
- `keys/` - API key management
- `combos/` - Model combo management
- `settings/` - User settings
- `usage/` - Usage statistics
- `models/` - Model management
  - `alias/` - Model aliases
  - `disabled/` - Disabled models
  - `availability/` - Model availability
- `proxy-pools/` - Proxy pool management
- `media-providers/` - Image/TTS/embedding provider config
- `cloud/` - Cloud sync
- `auth/` - Authentication
- `health/` - Health check
- `version/` - Version info

---

### 3. **src/lib/** - Shared Libraries

#### **db/** - Database Layer
- `driver.js` - Database driver abstraction
- `index.js` - Database initialization
- `schema.js` - Database schema
- `migrate.js` - Migration runner
- `version.js` - Version tracking
- `adapters/` - Database adapters
  - `betterSqliteAdapter.js` - better-sqlite3 (primary)
  - `nodeSqliteAdapter.js` - Node.js sqlite
  - `bunSqliteAdapter.js` - Bun runtime
  - `sqljsAdapter.js` - sql.js (fallback)
- `repos/` - Data repositories (10 repos)
  - `connectionsRepo.js` - Provider connections
  - `apiKeysRepo.js` - API keys
  - `combosRepo.js` - Model combos
  - `aliasRepo.js` - Model aliases
  - `settingsRepo.js` - Settings
  - `pricingRepo.js` - Pricing data
  - `requestDetailsRepo.js` - Request logs
  - `disabledModelsRepo.js` - Disabled models
  - `proxyPoolsRepo.js` - Proxy pools
  - `nodesRepo.js` - Compatible nodes
- `helpers/` - Database helpers
  - `kvStore.js` - Key-value store
  - `jsonCol.js` - JSON column handling
  - `metaStore.js` - Metadata store
- `migrations/` - Database migrations

#### **oauth/** - OAuth Implementation
- `providers.js` - OAuth provider registry
- `services/` - Provider-specific OAuth (13 providers)
  - `claude.js`, `gemini.js`, `codex.js`, `cursor.js`, etc.
- `constants/oauth.js` - OAuth constants
- `utils/` - OAuth utilities
  - `pkce.js` - PKCE flow
  - `server.js` - Local OAuth server
  - `ui.js` - OAuth UI
  - `banner.js` - Terminal banner

#### **network/** - Network & Proxy
- `outboundProxy.js` - Outbound proxy configuration
- `connectionProxy.js` - Connection proxy
- `initOutboundProxy.js` - Proxy initialization
- `proxyTest.js` - Proxy testing

---

### 4. **src/mitm/** - MITM Proxy System

**Purpose**: Transparent interception of provider requests for token rotation, caching, etc.

- `server.js` - MITM proxy server (HTTP/HTTPS)
- `manager.js` - MITM lifecycle management
- `config.js` - MITM configuration
- `logger.js` - MITM logging
- `paths.js` - Certificate paths
- `dbReader.js` - Database reader for MITM
- `winElevated.js` - Windows elevation handling
- `cert/` - Certificate management
  - `generate.js` - Generate self-signed certs
  - `install.js` - Install root CA
  - `rootCA.js` - Root CA management
- `dns/` - DNS configuration
  - `dnsConfig.js` - DNS settings
- `handlers/` - Provider-specific MITM handlers
  - `base.js` - Base handler
  - `antigravity.js` - Antigravity token rotation
  - `cursor.js` - Cursor interception
  - `copilot.js` - Copilot interception
  - `kiro.js` - Kiro interception
- `dev/` - Development utilities

---

### 5. **src/app/(dashboard)/** - Dashboard UI

**Pages**:
- `dashboard/` - Main dashboard
- `dashboard/cli-tools/` - CLI tool configuration
- `dashboard/providers/[id]/` - Provider details & management
- `dashboard/combos/` - Model combo management
- `dashboard/endpoint/` - Endpoint configuration
- `dashboard/mitm/` - MITM proxy settings
- `dashboard/media-providers/` - Image/TTS/embedding providers
- `dashboard/console-log/` - Request logs
- `dashboard/basic-chat/` - Chat interface
- `dashboard/profile/` - User profile

**Components**:
- CLI tool cards (CL4ude, Codex, Cursor, Cline, etc.)
- Provider connection UI
- Model selection
- Combo builder
- Settings panels

---

### 6. **src/sse/** - SSE-Specific Handlers

Specialized SSE (Server-Sent Events) handling for streaming responses.

---

### 7. **src/shared/** - Shared Utilities

Common utilities used across the application.

---

### 8. **src/store/** - State Management

Zustand-based state management for the dashboard.

---

## Key Features & Systems

### 1. **Provider Support (40+ Providers)**

**OAuth Providers** (auto-token refresh):
- Claude (Anthropic)
- Gemini (Google)
- Codex (OpenAI)
- Cursor
- GitHub Copilot
- Antigravity
- Kiro
- iFlow
- Qwen (Alibaba)
- OpenCode

**API Key Providers**:
- OpenAI
- OpenRouter
- GLM (Zhipu)
- Kimi
- MiniMax
- Azure OpenAI

**Compatible Nodes**:
- Ollama (local)
- OpenAI-compatible endpoints
- Custom nodes

---

### 2. **Format Translation**

Supports translation between:
- **OpenAI** - Standard format
- **Claude** - Anthropic format
- **Gemini** - Google format
- **Responses API** - OpenAI Responses format
- **Cursor** - Cursor IDE format
- **Antigravity** - Antigravity format
- **Kiro** - Kiro format
- **CommandCode** - CommandCode format
- **Ollama** - Ollama format

---

### 3. **RTK (Token Refresh & Compression)**

**Token Saving**: 20-40% reduction via:
- Caveman compression algorithm
- Tool output filtering (git diff, grep, ls, etc.)
- Smart truncation
- Deduplication

**Filters**:
- `grep.js` - Grep output compression
- `ls.js` - Directory listing compression
- `gitDiff.js` - Git diff compression
- `gitStatus.js` - Git status compression
- `tree.js` - Tree output compression
- `readNumbered.js` - Numbered content compression
- `searchList.js` - Search result compression
- `smartTruncate.js` - Smart truncation
- `find.js` - Find command compression
- `dedupLog.js` - Deduplication

---

### 4. **Multi-Account Fallback**

- **Tier 1**: Subscription providers (quota tracking)
- **Tier 2**: Cheap providers (budget limits)
- **Tier 3**: Free providers (unlimited)
- **Account-level fallback**: Round-robin between accounts per provider

---

### 5. **MITM Proxy System**

**Capabilities**:
- Transparent HTTP/HTTPS interception
- Self-signed certificate generation & installation
- Provider-specific request/response modification
- Token rotation at proxy level
- DNS configuration
- Windows elevation support

**Handlers**:
- Antigravity token rotation
- Cursor request modification
- Copilot interception
- Kiro interception

---

### 6. **Database Layer**

**Adapters**:
- better-sqlite3 (primary, fastest)
- sql.js (fallback, pure JS)
- Node.js sqlite
- Bun sqlite

**Repositories** (10 data types):
- Connections (provider auth)
- API Keys
- Model Combos
- Model Aliases
- Settings
- Pricing
- Request Details (logs)
- Disabled Models
- Proxy Pools
- Compatible Nodes

---

### 7. **Media Providers**

#### **Image Generation** (14 providers)
- Gemini
- OpenAI DALL-E
- Codex
- Black Forest Labs
- Cloudflare AI
- FAL AI
- Hugging Face
- Nanobanana
- Replicate
- Stability AI
- ComfyUI
- Pollinations
- Segmind
- Together AI

#### **Text-to-Speech** (10 providers)
- Google TTS
- OpenAI TTS
- Elevenlabs
- Edge TTS
- Gemini TTS
- OpenRouter TTS
- Generic formats
- Local device audio

#### **Embeddings** (5 providers)
- OpenAI
- Gemini
- OpenAI-compatible nodes
- Custom implementations

---

### 8. **Responses API**

Handles OpenAI Responses API format with:
- Streaming support
- JSON conversion
- Provider translation

---

## Configuration Files

### **open-sse/config/providers.js**
Defines all provider configurations:
```javascript
{
  baseUrl: "https://api.example.com/v1",
  format: "openai",
  headers: { /* provider-specific headers */ },
  clientId: "oauth-client-id",
  tokenUrl: "https://oauth.example.com/token"
}
```

### **open-sse/config/providerModels.js**
Maps models to providers:
```javascript
{
  "gpt-4": { provider: "openai", ... },
  "claude-3-opus": { provider: "claude", ... }
}
```

### **open-sse/config/runtimeConfig.js**
Runtime settings:
- Cache TTL
- Default max tokens
- Cooldown periods
- Backoff configuration

---

## API Endpoints

### **Chat Completion**
```
POST /v1/chat/completions
```
- Streaming & non-streaming
- Multi-provider routing
- Format translation
- Token refresh

### **Embeddings**
```
POST /v1/embeddings
```
- Multiple embedding providers
- Batch processing

### **Image Generation**
```
POST /v1/images/generations
```
- 14 image providers
- Format translation

### **Text-to-Speech**
```
POST /v1/audio/speech
```
- 10 TTS providers
- Audio format support

### **Speech-to-Text**
```
POST /v1/audio/transcriptions
```
- STT support

### **Models List**
```
GET /v1/models
```
- Available models
- Provider info

---

## Utilities & Helpers

### **Stream Handling**
- `streamHandler.js` - Stream creation & management
- `stream.js` - SSE transformation
- `pipeWithDisconnect()` - Graceful disconnection

### **Error Handling**
- `error.js` - Error formatting
- `errorConfig.js` - Error rules
- Provider-specific error mapping

### **Client Detection**
- `clientDetector.js` - Detect CLI tool type
- Cursor, Codex, Code Assistant, etc.

### **Caching**
- `claudeHeaderCache.js` - Claude header caching
- TTL-based cache invalidation

### **Proxy Support**
- `proxyFetch.js` - Fetch with proxy
- SOCKS5 support
- HTTP/HTTPS proxy

---

## Testing

**Test Framework**: Vitest

**Test Coverage**:
- Unit tests (25+ test files)
- E2E tests
- Provider integration tests
- RTK compression tests
- Database tests
- OAuth flow tests

**Key Test Files**:
- `combo-routing.test.js` - Fallback routing
- `rtk.test.js` - Token compression
- `embeddings.test.js` - Embedding providers
- `codex-image-fetch.test.js` - Image generation
- `antigravity-cache.test.js` - Caching
- `db-driver-chain.test.js` - Database

---

## Build & Deployment

### **Development**
```bash
npm run dev              # Next.js dev server (port 20128)
npm run dev:bun         # Bun runtime
```

### **Production Build**
```bash
npm run build            # Next.js build
npm run start            # Start production server
npm run build:bun       # Bun build
npm run start:bun       # Bun start
```

### **Docker**
```bash
docker build -t 9router .
docker run -p 20128:20128 9router
```

---

## Dependencies

**Core**:
- Next.js 16.1.6
- React 19.2.4
- Express 5.2.1
- Zustand 5.0.10

**Database**:
- better-sqlite3 (optional)
- sql.js (fallback)

**Utilities**:
- node-forge (certificate generation)
- http-proxy-middleware
- socks-proxy-agent
- undici (HTTP client)
- jose (JWT)
- bcryptjs (password hashing)

**UI**:
- Tailwind CSS 4
- Monaco Editor
- Recharts (charts)
- Material Symbols

---

## Key Concepts

### **Provider Format**
Each provider has a format (claude, openai, gemini, etc.) that defines:
- Request structure
- Response structure
- Header requirements
- Authentication method

### **Model Combo**
A sequence of models to try in order:
1. Primary model (subscription)
2. Fallback model (cheap)
3. Free model (unlimited)

### **Account Fallback**
Multiple accounts per provider for:
- Quota rotation
- Rate limit avoidance
- Multi-user support

### **Token Refresh (RTK)**
Automatic OAuth token refresh with:
- Expiry detection
- Refresh logic per provider
- Fallback on failure

### **MITM Proxy**
Transparent interception for:
- Token rotation
- Request modification
- Response caching
- Provider-specific handling

---

## File Statistics

- **Total Lines**: ~17,767
- **JavaScript Files**: 82+ files
- **Providers**: 40+
- **Image Providers**: 14
- **TTS Providers**: 10
- **Embedding Providers**: 5
- **Test Files**: 25+
- **Database Adapters**: 4
- **OAuth Providers**: 13

---

## Recent Updates (2026)

- v0.4.31 (2026-05-12) - Latest
- v0.4.29 (2026-05-10)
- Multiple releases per month
- Active development & bug fixes
- Community contributions (90+ contributors)

---

## Key Takeaways for Phase Prompts

1. **Modular Architecture**: `open-sse/` is provider-agnostic, `src/` is Next.js-specific
2. **Format Translation**: Central to multi-provider support
3. **Token Optimization**: RTK compression is a core feature
4. **MITM System**: Enables transparent provider interception
5. **Database Abstraction**: Multiple adapter support for portability
6. **OAuth Management**: Automatic token refresh for 13+ providers
7. **Streaming**: SSE-based streaming for all providers
8. **Fallback Logic**: Multi-tier (subscription → cheap → free)
9. **Dashboard**: React-based management UI
10. **Extensibility**: Easy to add new providers via executors

