# 9Router Codebase Map

**Reference**: https://github.com/decolua/9router  
**Local Copy**: `_ref/9router/`

---

## Directory Structure

### Root Level
```
9router/
├── src/                          # Next.js app (frontend + API routes)
├── open-sse/                     # Shared SSE/routing core (npm package)
├── docs/                         # Architecture documentation
├── gitbook/                      # Feature guides (multi-language)
├── public/                       # Static assets
├── tests/                        # Test fixtures
├── skills/                       # AI skills definitions
├── cloud/                        # Cloud sync helpers
├── package.json                  # Dependencies (Next.js 16, React 19)
├── next.config.mjs               # Next.js configuration
├── CHANGELOG.md                  # Version history
├── README.md                     # Main documentation
├── DOCKER.md                     # Docker deployment
└── Dockerfile                    # Multi-stage Docker build
```

---

## Key Directories

### `src/` - Next.js Application

#### `src/app/api/` - HTTP Routes
```
src/app/api/
├── v1/                           # 0penAI-compatible endpoints
│   ├── chat/completions/route.js
│   ├── messages/route.js
│   ├── responses/route.js
│   ├── models/route.js
│   └── count_tokens/route.js
├── v1beta/                       # Beta endpoints
│   └── models/[...path]/route.js
├── providers/                    # Provider CRUD
├── provider-nodes/               # Custom compatible nodes
├── oauth/                        # OAuth flows
├── keys/                         # API key management
├── models/                       # Model aliases & disabled
├── combos/                       # Fallback combo management
├── pricing/                      # Pricing overrides
├── usage/                        # Usage & analytics
├── auth/                         # Authentication
├── settings/                     # Configuration
├── sync/                         # Cloud sync
├── cloud/                        # Cloud helpers
└── cli-tools/                    # CLI tool helpers
```

#### `src/sse/` - SSE Handlers
```
src/sse/
├── handlers/
│   └── chat.js                   # Entry point for chat requests
├── services/
│   └── model.js                  # Model resolution
└── utils/
    └── ...
```

#### `src/lib/` - Utilities
```
src/lib/
├── localDb.js                    # Config persistence (db.json)
├── usageDb.js                    # Usage tracking (usage.json)
└── ...
```

### `open-sse/` - Core Routing Engine

#### `open-sse/handlers/`
```
open-sse/handlers/
└── chatCore.js                   # Core orchestration
    ├── Request translation
    ├── Executor dispatch
    ├── Retry/refresh handling
    ├── Stream setup
    └── Usage tracking
```

#### `open-sse/executors/` - Provider Adapters (17 providers)
```
open-sse/executors/
├── base.js                       # Base executor class
├── default.js                    # Generic 0penAI-compatible
├── antigravity.js                # Google Antigravity (17.7K)
├── azure.js                      # Azure 0penAI
├── codex.js                      # 0penAI Codex (8.6K)
├── commandcode.js                # CommandCode
├── cursor.js                     # Cursor IDE (22.4K)
├── gemini-cli.js                 # Gemini CLI
├── github.js                     # GitHub Copilot (14.3K)
├── grok-web.js                   # Grok Web (14.7K)
├── iflow.js                      # iFlow
├── kiro.js                       # Kiro AI (15.6K)
├── ollama-local.js               # Ollama Local
├── opencode-go.js                # OpenCode Go
├── opencode.js                   # OpenCode
└── perplexity-web.js             # Perplexity Web (19.0K)
```

**Executor Responsibilities**:
- Provider-specific authentication
- Request formatting
- Response parsing
- Error handling & retry logic
- Rate limit detection
- Token refresh

#### `open-sse/translator/` - Format Translation
```
open-sse/translator/
├── index.js                      # Translation registry
├── formats.js                    # Format definitions
├── request/                      # Request translators
│   ├── openai.js
│   ├── claude.js
│   ├── gemini.js
│   └── ...
└── response/                     # Response translators
    ├── openai.js
    ├── claude.js
    ├── gemini.js
    └── ...
```

**Supported Formats**:
- 0penAI (chat/completions, embeddings, TTS, STT, image, video)
- CL4ude (messages, responses)
- Gemini (generateContent)
- Cursor (proprietary)
- Kiro (proprietary)
- Vertex AI (proprietary)

#### `open-sse/rtk/` - Token Saver (20-40% compression)
```
open-sse/rtk/
├── index.js                      # Compression engine
├── autodetect.js                 # Format detection
├── applyFilter.js                # Filter application
├── caveman.js                    # Caveman mode (aggressive)
├── cavemanPrompts.js             # Caveman prompt templates
├── constants.js                  # RTK constants
├── registry.js                   # Filter registry
└── filters/                      # Compression filters
    ├── git-diff.js
    ├── git-status.js
    ├── grep.js
    ├── find.js
    ├── ls.js
    ├── tree.js
    ├── dedup-log.js
    ├── smart-truncate.js
    ├── read-numbered.js
    └── search-list.js
```

**Filters**: Auto-detect tool output format and apply lossless compression.

#### `open-sse/services/` - Business Logic
```
open-sse/services/
├── provider.js                   # Format detection & provider config
├── model.js                      # Model resolution
├── accountFallback.js            # Account-level fallback logic
└── tokenRefresh.js               # Token refresh orchestration
```

#### `open-sse/utils/` - Utilities
```
open-sse/utils/
├── stream.js                     # SSE stream utilities
├── streamHandler.js              # Stream handling
├── usageTracking.js              # Usage extraction & normalization
├── error.js                      # Error parsing & formatting
├── requestLogger.js              # Request logging
├── clientDetector.js             # Client tool detection
└── bypassHandler.js              # Bypass patterns (warmup, skip)
```

#### `open-sse/config/` - Configuration
```
open-sse/config/
├── providerModels.js             # Provider & model mappings
└── runtimeConfig.js              # Runtime constants
```

---

## Data Flow Diagrams

### Request Processing Pipeline

```
Client Request
    ↓
src/sse/handlers/chat.js
├─ Parse request body
├─ Detect source format
├─ Extract model/combo
├─ Resolve to provider + model
    ↓
open-sse/rtk/index.js
├─ Auto-detect tool output format
├─ Apply compression filter
├─ Log compression stats
    ↓
open-sse/translator/index.js
├─ Translate request to provider format
├─ Inject provider-specific headers
    ↓
open-sse/executors/[provider].js
├─ Get credentials (refresh if needed)
├─ Execute HTTP request
├─ Parse response
    ↓
open-sse/translator/response/
├─ Translate response back to source format
├─ Extract usage metrics
    ↓
open-sse/utils/usageTracking.js
├─ Record tokens used
├─ Calculate cost
├─ Update quota
├─ Log request details
    ↓
Stream to Client
```

### Fallback Flow

```
Request fails (429, 401, 503, etc.)
    ↓
open-sse/utils/error.js
├─ Parse error type
├─ Extract retry-after
    ↓
open-sse/services/accountFallback.js
├─ Check if retryable
├─ Select next provider/account
├─ Check quota availability
    ↓
Retry with new provider
    ├─ Translate to new format
    ├─ Execute with new credentials
    ↓
Success or Continue Fallback
```

---

## Critical Files for Porting

### Must-Read Files
1. **`docs/ARCHITECTURE.md`** - System design & boundaries
2. **`open-sse/handlers/chatCore.js`** - Core orchestration logic
3. **`open-sse/rtk/index.js`** - Token compression engine
4. **`open-sse/translator/index.js`** - Format translation registry
5. **`open-sse/services/accountFallback.js`** - Fallback decision logic
6. **`src/sse/handlers/chat.js`** - Request entry point

### Provider Executors (Reference)
- **`open-sse/executors/default.js`** - Generic 0penAI-compatible (template)
- **`open-sse/executors/cursor.js`** - Complex executor example (22.4K)
- **`open-sse/executors/github.js`** - OAuth provider example (14.3K)

### RTK Filters (Reference)
- **`open-sse/rtk/filters/git-diff.js`** - Most common filter
- **`open-sse/rtk/autodetect.js`** - Format detection logic
- **`open-sse/rtk/caveman.js`** - Aggressive compression

---

## Configuration & Constants

### Provider Mappings
**File**: `open-sse/config/providerModels.js`

Maps provider IDs to:
- Aliases (e.g., `cc/` → Code Assistant)
- Model lists
- Format overrides
- Strip lists (fields to remove)
- Thinking config

### Runtime Constants
**File**: `open-sse/config/runtimeConfig.js`

- HTTP status codes
- Error messages
- Timeout values
- Retry policies

### RTK Constants
**File**: `open-sse/rtk/constants.js`

- `RAW_CAP` - Max raw output size
- `MIN_COMPRESS_SIZE` - Minimum size to compress
- Filter thresholds

---

## Testing & Fixtures

### Test Structure
```
tests/
└── README.md                     # Test documentation
```

### Golden Files (Translator Tests)
```
testdata/
├── [case]/
│   ├── in.json                   # Input request
│   └── out.json                  # Expected output
```

**Update**: `go test ./... -update` (for Go port)

---

## Build & Deployment

### Development
```bash
npm run dev                        # Hot reload on port 20128
```

### Production Build
```bash
npm run build                      # Next.js build
npm run start                      # Start server
```

### Docker
```bash
docker build -t 9router .
docker run -p 20128:20128 9router
```

### Environment Variables
```bash
PORT=20128
HOSTNAME=0.0.0.0
NEXT_PUBLIC_BASE_URL=http://localhost:20128
DATA_DIR=~/.9router
```

---

## Key Patterns & Conventions

### Error Handling
- **Typed Errors**: `open-sse/utils/error.js` defines error types
- **Retry Logic**: Built into `open-sse/services/accountFallback.js`
- **Fallback**: Automatic provider switching on error

### Streaming
- **SSE Format**: Server-Sent Events for chat completions
- **Chunked Reads**: Never buffer full stream in memory
- **Drain Handling**: Use `http.Flusher` for backpressure

### Format Translation
- **Bidirectional**: Request → Provider → Response → Client
- **Auto-Detection**: Detect source format from request
- **Fallback**: If translation fails, pass through original

### Token Tracking
- **Per-Request**: Track input/output tokens
- **Per-Provider**: Aggregate usage by provider
- **Cost Calculation**: Multiply tokens by provider pricing

---

## Known Quirks (Important for Porting)

1. **usageDb Location**: Stores under `~/.9router`, not `DATA_DIR`
2. **Static Model List**: `/api/v1/route.js` is static, not main source
3. **Request Logging**: Full headers/body logged; treat as sensitive
4. **Cloud Sync**: Requires correct `NEXT_PUBLIC_BASE_URL`
5. **RTK Markdown**: May misalign column formatting
6. **Caveman Mode**: Can reduce quality; test before enabling

---

## References

- **GitHub**: https://github.com/decolua/9router
- **NPM**: https://www.npmjs.com/package/9router
- **Website**: https://9router.com
- **Latest Release**: v0.4.29 (May 2026)

