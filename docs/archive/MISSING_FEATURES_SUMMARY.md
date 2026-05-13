# Missing Features Summary - ai_proxy vs 9router

**Generated:** 2026-05-12  
**Status:** Preliminary analysis (detailed analysis in progress)

## Quick Stats

| Category | 9router | ai_proxy | Missing |
|----------|---------|----------|---------|
| Backend JS files (open-sse) | 150 | 130 Go files | ~20-30 files |
| Frontend files (src) | 382 JS/JSX | 15 TS/TSX | ~367 files |
| Dashboard pages | 25 pages | 10 pages | 15 pages |
| API routes | 24 routes | 10 routes | 14 routes |
| Shared components | 43 components | 0 components | 43 components |

---

## BACKEND - Missing Components

### 1. **HANDLERS** (Critical - Core Functionality)

#### Missing from `open-sse/handlers/`:

**Chat Handlers:**
- `chatCore/nonStreamingHandler.js` - Non-streaming chat responses
- `chatCore/streamingHandler.js` - SSE streaming chat responses
- `chatCore/sseToJsonHandler.js` - SSE to JSON conversion
- `chatCore/requestDetail.js` - Request detail logging

**Embedding Providers:**
- `embeddingProviders/_base.js` - Base embedding provider
- `embeddingProviders/gemini.js` - Gemini embeddings
- `embeddingProviders/openai.js` - OpenAI embeddings
- `embeddingProviders/openaiCompatNode.js` - OpenAI-compatible node embeddings

**Image Generation Providers (13 providers):**
- `imageProviders/_base.js` - Base image provider
- `imageProviders/blackForestLabs.js` - Black Forest Labs (Flux)
- `imageProviders/cloudflareAi.js` - Cloudflare AI
- `imageProviders/codex.js` - Codex image generation
- `imageProviders/comfyui.js` - ComfyUI integration
- `imageProviders/falAi.js` - Fal.ai
- `imageProviders/gemini.js` - Gemini image generation
- `imageProviders/huggingface.js` - Hugging Face
- `imageProviders/nanobanana.js` - Nanobanana
- `imageProviders/openai.js` - DALL-E
- `imageProviders/runwayml.js` - Runway ML
- `imageProviders/sdwebui.js` - Stable Diffusion WebUI
- `imageProviders/stabilityAi.js` - Stability AI

**TTS Providers (9 providers):**
- `ttsProviders/_base.js` - Base TTS provider
- `ttsProviders/edgeTts.js` - Edge TTS
- `ttsProviders/elevenlabs.js` - ElevenLabs
- `ttsProviders/gemini.js` - Gemini TTS
- `ttsProviders/genericFormats.js` - Generic format support
- `ttsProviders/googleTts.js` - Google TTS
- `ttsProviders/localDevice.js` - Local device TTS
- `ttsProviders/openai.js` - OpenAI TTS
- `ttsProviders/openrouter.js` - OpenRouter TTS

**Other Handlers:**
- `responsesHandler.js` - OpenAI Responses API handler
- `imageGenerationCore.js` - Core image generation logic
- `sttCore.js` - Speech-to-text core
- `ttsCore.js` - Text-to-speech core
- `fetch/index.js` - Web fetch handler
- `search/chatSearch.js` - Chat search integration
- `search/callers.js` - Search API callers
- `search/normalizers.js` - Search result normalizers

### 2. **SERVICES** (Critical - Business Logic)

Missing from `open-sse/services/`:
- `accountFallback.js` - Account-level fallback logic
- `combo.js` - Combo fallback orchestration
- `compact.js` - Response compaction service
- `model.js` - Model resolution and management
- `projectId.js` - Project ID handling (Gemini, etc.)
- `provider.js` - Provider management service
- `tokenRefresh.js` - Token refresh orchestration
- `usage.js` - Usage tracking and analytics

**Status in ai_proxy:** Partially implemented in `internal/router/` but not as separate services

### 3. **UTILITIES** (High Priority - Supporting Functions)

Missing from `open-sse/utils/`:
- `bypassHandler.js` - Bypass logic for special cases
- `claudeCloaking.js` - Claude header cloaking
- `claudeHeaderCache.js` - Claude header caching
- `clientDetector.js` - Client detection (Cursor, Codex, etc.)
- `cursorChecksum.js` - Cursor request checksum
- `cursorProtobuf.js` - Cursor protobuf handling
- `error.js` - Error handling utilities
- `ollamaTransform.js` - Ollama request/response transformation
- `proxyFetch.js` - Proxy fetch utilities
- `reasoningContentInjector.js` - Reasoning content injection
- `requestLogger.js` - Request logging
- `sessionManager.js` - Session management
- `stream.js` - Stream utilities
- `streamHandler.js` - Stream handler
- `streamHelpers.js` - Stream helper functions
- `usageTracking.js` - Usage tracking utilities

### 4. **CONFIG** (Medium Priority - Configuration)

Missing from `open-sse/config/`:
- `appConstants.js` - Application constants
- `codexInstructions.js` - Codex-specific instructions
- `constants.js` - General constants
- `defaultThinkingSignature.js` - Default thinking signature
- `errorConfig.js` - Error configuration
- `googleTtsLanguages.js` - Google TTS language mappings
- `models.js` - Model definitions
- `ollamaModels.js` - Ollama model list
- `providerModels.js` - Provider-specific model mappings
- `providers.js` - Provider configurations
- `runtimeConfig.js` - Runtime configuration
- `ttsModels.js` - TTS model definitions

**Status in ai_proxy:** Basic config in `internal/config/config.go` - missing provider-specific configs

### 5. **TRANSFORMER** (Medium Priority)

Missing from `open-sse/transformer/`:
- `streamToJsonConverter.js` - SSE stream to JSON conversion
- `responsesTransformer.js` - Responses API transformation

---

## API ENDPOINTS - Missing Routes

### Missing from `src/app/api/`:

1. **`/api/cli-tools`** - CLI tool configuration helpers
2. **`/api/cloud`** - Cloud sync functionality
3. **`/api/locale`** - Internationalization
4. **`/api/media-providers`** - Media provider management
5. **`/api/models`** - Model management and listing
6. **`/api/pricing`** - Pricing information
7. **`/api/proxy-pools`** - Proxy pool management
8. **`/api/shutdown`** - Graceful shutdown
9. **`/api/tags`** - Tag management
10. **`/api/translator`** - Translator configuration
11. **`/api/tunnel`** - Tunnel management (Cloudflare, Tailscale)
12. **`/api/version`** - Version information
13. **`/api/v1/messages`** - Claude Messages API
14. **`/api/v1/responses`** - OpenAI Responses API
15. **`/api/v1/api/chat`** - Alternative chat endpoint
16. **`/api/v1/audio/voices`** - Voice listing
17. **`/api/v1/models/[kind]`** - Model listing by kind
18. **`/api/v1/models/info`** - Model information
19. **`/api/v1/responses/compact`** - Response compaction
20. **`/api/v1/web/fetch`** - Web fetch endpoint
21. **`/api/v1beta/models`** - Gemini model listing

**Status in ai_proxy:** Only 10 routes implemented (auth, providers, keys, combos, settings, oauth, embeddings, fetch, images, search, stt, tts)

---

## FRONTEND - Missing Components

### 1. **DASHBOARD PAGES** (15 missing pages)

Missing from `src/app/(dashboard)/dashboard/`:
- `basic-chat/` - Basic chat interface
- `cli-tools/` - CLI tools configuration
- `console-log/` - Console log viewer
- `endpoint/` - Endpoint configuration
- `media-providers/[kind]/` - Media provider management
- `media-providers/[kind]/[id]/` - Media provider detail
- `media-providers/combo/[id]/` - Media provider combo
- `media-providers/web/` - Web media providers
- `mitm/` - MITM proxy configuration
- `profile/` - User profile
- `providers/[id]/` - Provider detail page
- `providers/new/` - New provider page
- `proxy-pools/` - Proxy pool management
- `quota/` - Quota management
- `skills/` - Skills management
- `translator/` - Translator configuration

**Status in ai_proxy:** Only 7 pages (dashboard, providers, combos, keys, aliases, analytics, nodes, oauth)

### 2. **SHARED COMPONENTS** (43 components)

Missing from `src/shared/components/`:
- `AddCustomEmbeddingModal.js`
- `Avatar.js`
- `Badge.js`
- `Button.js`
- `Card.js`
- `ChangelogModal.js`
- `ComboFormModal.js`
- `CursorAuthModal.js`
- `Drawer.js`
- `EditConnectionModal.js`
- `Footer.js`
- `GitLabAuthModal.js`
- `Header.js`
- `HeaderMenu.js`
- `IFlowCookieModal.js`
- `Input.js`
- `KiroAuthModal.js`
- `KiroOAuthWrapper.js`
- `KiroSocialOAuthModal.js`
- `LanguageSwitcher.js`
- `Loading.js`
- `ManualConfigModal.js`
- `McpMarketplaceModal.js`
- `Modal.js`
- `ModelSelectModal.js`
- `NineRemoteButton.js`
- `NineRemotePromoModal.js`
- `NoAuthProxyCard.js`
- `OAuthModal.js`
- `Pagination.js`
- `PricingModal.js`
- `ProviderIcon.js`
- `ProviderInfoCard.js`
- `RequestLogger.js`
- `SegmentedControl.js`
- `Select.js`
- `Sidebar.js`
- `ThemeProvider.js`
- `ThemeToggle.js`
- `Toggle.js`
- `Tooltip.js`
- `UsageStats.js`
- `layouts/AuthLayout.js`
- `layouts/DashboardLayout.js`

**Status in ai_proxy:** 0 shared components - everything is inline

### 3. **STATE MANAGEMENT** (7 stores)

Missing from `src/store/`:
- `headerSearchStore.js` - Header search state
- `notificationStore.js` - Notification state
- `providerStore.js` - Provider state
- `settingsStore.js` - Settings state
- `themeStore.js` - Theme state
- `userStore.js` - User state
- `index.js` - Store exports

**Status in ai_proxy:** No state management - using React state only

### 4. **LIBRARY UTILITIES** (16 files)

Missing from `src/lib/`:
- `appUpdater.js` - Application updater
- `consoleLogBuffer.js` - Console log buffering
- `dataDir.js` - Data directory management
- `disabledModelsDb.js` - Disabled models database
- `initCloudSync.js` - Cloud sync initialization
- `localDb.js` - Local database wrapper
- `mitmAliasCache.js` - MITM alias caching
- `providerNormalization.js` - Provider normalization
- `requestDetailsDb.js` - Request details database
- `usageDb.js` - Usage database
- `db/migrate.js` - Database migrations
- `db/schema.js` - Database schema
- `db/driver.js` - Database driver
- `db/backup.js` - Database backup
- `network/initOutboundProxy.js` - Outbound proxy init
- `network/proxyTest.js` - Proxy testing
- `network/outboundProxy.js` - Outbound proxy
- `network/connectionProxy.js` - Connection proxy
- `tunnel/state.js` - Tunnel state
- `tunnel/networkProbe.js` - Network probing
- `tunnel/cloudflared.js` - Cloudflare tunnel
- `tunnel/tunnelManager.js` - Tunnel manager
- `tunnel/tunnelConfig.js` - Tunnel config
- `tunnel/tailscale.js` - Tailscale integration
- `usage/fetcher.js` - Usage fetcher
- `updater/updater.js` - Updater logic

**Status in ai_proxy:** Only 4 API wrapper files (admin-api.ts, analytics-api.ts, api.ts, oauth-api.ts)

### 5. **SHARED UTILITIES** (9 files)

Missing from `src/shared/utils/`:
- `clineAuth.js` - Cline authentication
- `cn.js` - Class name utilities
- `cloud.js` - Cloud utilities
- `api.js` - API utilities
- `machine.js` - Machine utilities
- `providerModelsFetcher.js` - Provider models fetcher
- `machineId.js` - Machine ID generation
- `apiKey.js` - API key utilities

### 6. **SHARED CONSTANTS** (11 files)

Missing from `src/shared/constants/`:
- `cliTools.js` - CLI tools constants
- `colors.js` - Color constants
- `config.js` - Configuration constants
- `coworkPlugins.js` - Cowork plugins
- `mitmToolHosts.js` - MITM tool hosts
- `models.js` - Model constants
- `pricing.js` - Pricing constants
- `providers.js` - Provider constants
- `skills.js` - Skills constants
- `ttsProviders.js` - TTS provider constants

### 7. **SHARED HOOKS** (3 files)

Missing from `src/shared/hooks/`:
- `useCopyToClipboard.js` - Copy to clipboard hook
- `useTheme.js` - Theme hook

### 8. **SHARED SERVICES** (3 files)

Missing from `src/shared/services/`:
- `initializeCloudSync.js` - Cloud sync initialization
- `initializeApp.js` - App initialization
- `cloudSyncScheduler.js` - Cloud sync scheduler

---

## MITM PROXY - Completely Missing

Missing from `src/mitm/`:
- `config.js` - MITM configuration
- `logger.js` - MITM logger
- `server.js` - MITM server
- `manager.js` - MITM manager
- `winElevated.js` - Windows elevation
- `handlers/cursor.js` - Cursor MITM handler
- `handlers/copilot.js` - Copilot MITM handler
- `handlers/kiro.js` - Kiro MITM handler
- `handlers/base.js` - Base MITM handler
- `handlers/antigravity.js` - Antigravity MITM handler

**Status in ai_proxy:** MITM proxy not implemented at all

---

## SSE SERVICES - Partially Missing

Missing from `src/sse/`:
- `services/tokenRefresh.js` - Token refresh service
- `services/auth.js` - Auth service
- `services/model.js` - Model service
- `utils/logger.js` - SSE logger
- `handlers/imageGeneration.js` - Image generation handler
- `handlers/embeddings.js` - Embeddings handler
- `handlers/tts.js` - TTS handler
- `handlers/stt.js` - STT handler
- `handlers/fetch.js` - Fetch handler
- `handlers/search.js` - Search handler

**Status in ai_proxy:** Some handlers exist in `internal/api/v1/` but not complete

---

## PRIORITY CLASSIFICATION

### **P0 - Critical (Blocks Core Functionality)**
1. Chat handlers (streaming, non-streaming, SSE to JSON)
2. Services (accountFallback, combo, tokenRefresh, usage)
3. Stream utilities (stream.js, streamHandler.js, streamHelpers.js)
4. Client detector (for Cursor, Codex, etc.)
5. Error handling utilities

### **P1 - High (Major Features)**
1. Image generation providers (13 providers)
2. TTS providers (9 providers)
3. Embedding providers (4 providers)
4. Responses API handler
5. Transformer (streamToJsonConverter, responsesTransformer)
6. Provider-specific utilities (claudeCloaking, cursorProtobuf, etc.)

### **P2 - Medium (Enhanced Features)**
1. MITM proxy (entire subsystem)
2. Tunnel management (Cloudflare, Tailscale)
3. Cloud sync
4. Console log viewer
5. Request logger
6. Config system (provider configs, model configs)

### **P3 - Low (Nice to Have)**
1. Frontend components (43 components)
2. Dashboard pages (15 pages)
3. State management (7 stores)
4. Shared utilities and constants
5. App updater
6. i18n/locale support

---

## ESTIMATED EFFORT

| Category | Files | Estimated LOC | Effort (days) |
|----------|-------|---------------|---------------|
| Backend Handlers | ~40 files | ~8,000 LOC | 15-20 days |
| Backend Services | ~8 files | ~2,000 LOC | 5-7 days |
| Backend Utils | ~16 files | ~3,000 LOC | 7-10 days |
| Backend Config | ~12 files | ~1,500 LOC | 3-5 days |
| API Endpoints | ~21 routes | ~4,000 LOC | 10-15 days |
| Frontend Components | ~43 files | ~6,000 LOC | 15-20 days |
| Frontend Pages | ~15 pages | ~3,000 LOC | 7-10 days |
| Frontend Utils/Stores | ~30 files | ~4,000 LOC | 10-12 days |
| MITM Proxy | ~10 files | ~2,000 LOC | 5-7 days |
| **TOTAL** | **~195 files** | **~33,500 LOC** | **77-106 days** |

**Note:** This is for a single developer. With parallel work and proper task decomposition, this can be reduced significantly.

---

## NEXT STEPS

1. **Prioritize P0 items** - Get core chat functionality to 100% parity
2. **Port handlers systematically** - Start with chatCore, then embeddings, images, TTS
3. **Implement missing services** - accountFallback, combo, tokenRefresh
4. **Add utilities** - Stream handling, client detection, error handling
5. **Complete API endpoints** - Messages API, Responses API, model listing
6. **Frontend modernization** - Decide which features to port vs. redesign
7. **MITM proxy** - Evaluate if needed for Go version

---

**Status:** This is a preliminary summary. Detailed file-by-file analysis is in progress.
