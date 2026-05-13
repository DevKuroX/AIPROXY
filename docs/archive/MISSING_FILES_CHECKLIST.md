# MISSING FILES CHECKLIST
## Complete File-by-File Status

**Last Updated:** 2026-05-12  
**Total Files:** 150+ (9router) vs 130 (ai_proxy)  
**Missing:** 20-30 files

---

## BACKEND FILES

### CONFIG (12 files in 9router, 3 in ai_proxy)

- [ ] `/open-sse/config/appConstants.js` → `/internal/config/constants.go`
- [ ] `/open-sse/config/codexInstructions.js` → `/internal/config/codex.go`
- [ ] `/open-sse/config/constants.js` → `/internal/config/constants.go`
- [ ] `/open-sse/config/defaultThinkingSignature.js` → `/internal/config/thinking.go`
- [ ] `/open-sse/config/errorConfig.js` → `/internal/errs/errors.go`
- [ ] `/open-sse/config/googleTtsLanguages.js` → `/internal/config/tts_languages.go`
- [x] `/open-sse/config/models.js` → `/internal/models/models.go` (Partial)
- [ ] `/open-sse/config/ollamaModels.js` → `/internal/config/ollama.go`
- [ ] `/open-sse/config/providerModels.js` → `/internal/config/provider_models.go`
- [ ] `/open-sse/config/providers.js` → `/internal/config/providers.go`
- [x] `/open-sse/config/runtimeConfig.js` → `/internal/config/config.go` (Partial)
- [ ] `/open-sse/config/ttsModels.js` → `/internal/config/tts_models.go`

### EXECUTORS (20 files - All Present)

- [x] `/open-sse/executors/antigravity.js` → `/internal/executor/antigravity.go`
- [x] `/open-sse/executors/azure.js` → `/internal/executor/azure.go`
- [x] `/open-sse/executors/base.js` → `/internal/executor/base.go`
- [x] `/open-sse/executors/codex.js` → `/internal/executor/codex.go`
- [x] `/open-sse/executors/commandcode.js` → `/internal/executor/commandcode.go`
- [x] `/open-sse/executors/cursor.js` → `/internal/executor/cursor.go`
- [x] `/open-sse/executors/default.js` → `/internal/executor/default.go`
- [x] `/open-sse/executors/gemini-cli.js` → `/internal/executor/gemini_cli.go`
- [x] `/open-sse/executors/github.js` → `/internal/executor/github.go`
- [x] `/open-sse/executors/grok-web.js` → `/internal/executor/grok_web.go`
- [x] `/open-sse/executors/iflow.js` → `/internal/executor/iflow.go`
- [x] `/open-sse/executors/kiro.js` → `/internal/executor/kiro.go`
- [x] `/open-sse/executors/ollama-local.js` → `/internal/executor/ollama_local.go`
- [x] `/open-sse/executors/opencode-go.js` → `/internal/executor/opencode_go.go`
- [x] `/open-sse/executors/opencode.js` → `/internal/executor/opencode.go`
- [x] `/open-sse/executors/perplexity-web.js` → `/internal/executor/perplexity_web.go`
- [x] `/open-sse/executors/qoder.js` → `/internal/executor/qoder.go`
- [x] `/open-sse/executors/qwen.js` → `/internal/executor/qwen.go`
- [x] `/open-sse/executors/vertex.js` → `/internal/executor/vertex.go`
- [x] `/open-sse/executors/index.js` → `/internal/executor/registry.go`

### HANDLERS (44 files in 9router, 4 in ai_proxy)

#### Chat Handlers
- [ ] `/open-sse/handlers/chatCore.js` → `/internal/handlers/chat.go`
- [ ] `/open-sse/handlers/chatCore/nonStreamingHandler.js` → `/internal/handlers/chat_non_streaming.go`
- [ ] `/open-sse/handlers/chatCore/requestDetail.js` → `/internal/handlers/request_detail.go`
- [ ] `/open-sse/handlers/chatCore/sseToJsonHandler.js` → `/internal/handlers/sse_to_json.go`
- [ ] `/open-sse/handlers/chatCore/streamingHandler.js` → `/internal/handlers/chat_streaming.go`

#### Embeddings Handlers
- [ ] `/open-sse/handlers/embeddingsCore.js` → `/internal/handlers/embeddings.go`
- [ ] `/open-sse/handlers/embeddingProviders/_base.js` → `/internal/handlers/embeddings/base.go`
- [ ] `/open-sse/handlers/embeddingProviders/gemini.js` → `/internal/handlers/embeddings/gemini.go`
- [ ] `/open-sse/handlers/embeddingProviders/index.js` → `/internal/handlers/embeddings/registry.go`
- [ ] `/open-sse/handlers/embeddingProviders/openai.js` → `/internal/handlers/embeddings/openai.go`
- [ ] `/open-sse/handlers/embeddingProviders/openaiCompatNode.js` → `/internal/handlers/embeddings/openai_compat.go`

#### Image Handlers
- [ ] `/open-sse/handlers/imageGenerationCore.js` → `/internal/handlers/images.go`
- [ ] `/open-sse/handlers/imageProviders/_base.js` → `/internal/handlers/images/base.go`
- [ ] `/open-sse/handlers/imageProviders/blackForestLabs.js` → `/internal/handlers/images/flux.go`
- [ ] `/open-sse/handlers/imageProviders/cloudflareAi.js` → `/internal/handlers/images/cloudflare.go`
- [ ] `/open-sse/handlers/imageProviders/codex.js` → `/internal/handlers/images/codex.go`
- [ ] `/open-sse/handlers/imageProviders/comfyui.js` → `/internal/handlers/images/comfyui.go`
- [ ] `/open-sse/handlers/imageProviders/falAi.js` → `/internal/handlers/images/fal.go`
- [ ] `/open-sse/handlers/imageProviders/gemini.js` → `/internal/handlers/images/gemini.go`
- [ ] `/open-sse/handlers/imageProviders/huggingface.js` → `/internal/handlers/images/huggingface.go`
- [ ] `/open-sse/handlers/imageProviders/index.js` → `/internal/handlers/images/registry.go`
- [ ] `/open-sse/handlers/imageProviders/nanobanana.js` → `/internal/handlers/images/nanobanana.go`
- [ ] `/open-sse/handlers/imageProviders/openai.js` → `/internal/handlers/images/openai.go`
- [ ] `/open-sse/handlers/imageProviders/runwayml.js` → `/internal/handlers/images/runway.go`
- [ ] `/open-sse/handlers/imageProviders/sdwebui.js` → `/internal/handlers/images/sdwebui.go`
- [ ] `/open-sse/handlers/imageProviders/stabilityAi.js` → `/internal/handlers/images/stability.go`

#### TTS Handlers
- [ ] `/open-sse/handlers/ttsCore.js` → `/internal/handlers/tts.go`
- [ ] `/open-sse/handlers/ttsProviders/_base.js` → `/internal/handlers/tts/base.go`
- [ ] `/open-sse/handlers/ttsProviders/edgeTts.js` → `/internal/handlers/tts/edge.go`
- [ ] `/open-sse/handlers/ttsProviders/elevenlabs.js` → `/internal/handlers/tts/elevenlabs.go`
- [ ] `/open-sse/handlers/ttsProviders/gemini.js` → `/internal/handlers/tts/gemini.go`
- [ ] `/open-sse/handlers/ttsProviders/genericFormats.js` → `/internal/handlers/tts/generic.go`
- [ ] `/open-sse/handlers/ttsProviders/googleTts.js` → `/internal/handlers/tts/google.go`
- [ ] `/open-sse/handlers/ttsProviders/localDevice.js` → `/internal/handlers/tts/local.go`
- [ ] `/open-sse/handlers/ttsProviders/openai.js` → `/internal/handlers/tts/openai.go`
- [ ] `/open-sse/handlers/ttsProviders/openrouter.js` → `/internal/handlers/tts/openrouter.go`

#### Other Handlers
- [ ] `/open-sse/handlers/responsesHandler.js` → `/internal/handlers/responses.go`
- [ ] `/open-sse/handlers/sttCore.js` → `/internal/handlers/stt.go`
- [ ] `/open-sse/handlers/fetch/index.js` → `/internal/handlers/fetch.go`
- [ ] `/open-sse/handlers/search/index.js` → `/internal/handlers/search.go`
- [ ] `/open-sse/handlers/search/chatSearch.js` → `/internal/handlers/search_chat.go`
- [ ] `/open-sse/handlers/search/callers.js` → `/internal/handlers/search_callers.go`
- [ ] `/open-sse/handlers/search/normalizers.js` → `/internal/handlers/search_normalizers.go`

### SERVICES (8 files in 9router, 0 in ai_proxy)

- [ ] `/open-sse/services/accountFallback.js` → `/internal/services/account_fallback.go`
- [ ] `/open-sse/services/combo.js` → `/internal/services/combo.go`
- [ ] `/open-sse/services/compact.js` → `/internal/services/compact.go`
- [ ] `/open-sse/services/model.js` → `/internal/services/model.go`
- [ ] `/open-sse/services/projectId.js` → `/internal/services/project_id.go`
- [ ] `/open-sse/services/provider.js` → `/internal/services/provider.go`
- [ ] `/open-sse/services/tokenRefresh.js` → `/internal/services/token_refresh.go`
- [ ] `/open-sse/services/usage.js` → `/internal/services/usage.go`

### UTILITIES (16 files in 9router, 0 in ai_proxy)

- [ ] `/open-sse/utils/bypassHandler.js` → `/internal/utils/bypass.go`
- [ ] `/open-sse/utils/claudeCloaking.js` → `/internal/utils/claude_cloak.go`
- [ ] `/open-sse/utils/claudeHeaderCache.js` → `/internal/utils/claude_headers.go`
- [ ] `/open-sse/utils/clientDetector.js` → `/internal/utils/client_detect.go`
- [ ] `/open-sse/utils/cursorChecksum.js` → `/internal/utils/cursor_checksum.go`
- [ ] `/open-sse/utils/cursorProtobuf.js` → `/internal/utils/cursor_protobuf.go`
- [ ] `/open-sse/utils/error.js` → `/internal/utils/error.go`
- [ ] `/open-sse/utils/ollamaTransform.js` → `/internal/utils/ollama_transform.go`
- [ ] `/open-sse/utils/proxyFetch.js` → `/internal/utils/proxy_fetch.go`
- [ ] `/open-sse/utils/reasoningContentInjector.js` → `/internal/utils/reasoning_inject.go`
- [ ] `/open-sse/utils/requestLogger.js` → `/internal/utils/request_log.go`
- [ ] `/open-sse/utils/responseCache.js` → `/internal/utils/response_cache.go`
- [ ] `/open-sse/utils/sessionManager.js` → `/internal/utils/session.go`
- [ ] `/open-sse/utils/stream.js` → `/internal/utils/stream.go`
- [ ] `/open-sse/utils/streamHandler.js` → `/internal/utils/stream_handler.go`
- [ ] `/open-sse/utils/streamHelpers.js` → `/internal/utils/stream_helpers.go`
- [ ] `/open-sse/utils/usageTracking.js` → `/internal/utils/usage_track.go`

### TRANSLATOR (30 files in 9router, 18 in ai_proxy)

#### Request Translators
- [x] `/open-sse/translator/request/antigravity-to-openai.js` → Exists
- [x] `/open-sse/translator/request/claude-to-openai.js` → Exists
- [x] `/open-sse/translator/request/gemini-to-openai.js` → Exists
- [x] `/open-sse/translator/request/openai-to-claude.js` → Exists
- [x] `/open-sse/translator/request/openai-to-commandcode.js` → Exists
- [x] `/open-sse/translator/request/openai-to-cursor.js` → Exists
- [x] `/open-sse/translator/request/openai-to-gemini.js` → Exists
- [x] `/open-sse/translator/request/openai-to-kiro.js` → Exists
- [ ] `/open-sse/translator/request/openai-to-qwen.js` → Missing
- [ ] `/open-sse/translator/request/openai-to-perplexity.js` → Missing
- [ ] `/open-sse/translator/request/openai-to-iflow.js` → Missing
- [ ] `/open-sse/translator/request/openai-to-qoder.js` → Missing

#### Response Translators
- [x] `/open-sse/translator/response/claude-to-openai.js` → Exists
- [x] `/open-sse/translator/response/commandcode-to-openai.js` → Exists
- [x] `/open-sse/translator/response/cursor-to-openai.js` → Exists
- [x] `/open-sse/translator/response/gemini-to-openai.js` → Exists
- [x] `/open-sse/translator/response/kiro-to-openai.js` → Exists
- [x] `/open-sse/translator/response/ollama-to-openai.js` → Exists
- [x] `/open-sse/translator/response/openai-to-antigravity.js` → Exists
- [x] `/open-sse/translator/response/openai-to-claude.js` → Exists
- [ ] `/open-sse/translator/response/qwen-to-openai.js` → Missing
- [ ] `/open-sse/translator/response/perplexity-to-openai.js` → Missing
- [ ] `/open-sse/translator/response/iflow-to-openai.js` → Missing
- [ ] `/open-sse/translator/response/qoder-to-openai.js` → Missing

#### Helpers
- [x] `/open-sse/translator/helpers/claudeHelper.js` → Exists
- [x] `/open-sse/translator/helpers/geminiHelper.js` → Exists
- [x] `/open-sse/translator/helpers/imageHelper.js` → Exists
- [x] `/open-sse/translator/helpers/maxTokensHelper.js` → Exists
- [x] `/open-sse/translator/helpers/openaiHelper.js` → Exists
- [x] `/open-sse/translator/helpers/responsesApiHelper.js` → Exists
- [x] `/open-sse/translator/helpers/toolCallHelper.js` → Exists
- [ ] `/open-sse/translator/helpers/visionHelper.js` → Missing
- [ ] `/open-sse/translator/helpers/functionCallingHelper.js` → Missing
- [ ] `/open-sse/translator/helpers/streamingHelper.js` → Missing
- [ ] `/open-sse/translator/helpers/errorHelper.js` → Missing

### TRANSFORMER (2 files - Both Present)

- [x] `/open-sse/transformer/responsesTransformer.js` → Exists
- [x] `/open-sse/transformer/streamToJsonConverter.js` → Exists

### RTK (17 files - All Present)

- [x] All RTK files present in ai_proxy

---

## FRONTEND FILES

### Dashboard Pages (15 Missing)

- [ ] `/src/app/(dashboard)/dashboard/providers/page.js`
- [ ] `/src/app/(dashboard)/dashboard/settings/page.js`
- [ ] `/src/app/(dashboard)/dashboard/keys/page.js`
- [ ] `/src/app/(dashboard)/dashboard/usage/page.js`
- [ ] `/src/app/(dashboard)/dashboard/combos/page.js`
- [ ] `/src/app/(dashboard)/dashboard/cli-tools/page.js`
- [ ] `/src/app/(dashboard)/dashboard/endpoint/page.js`
- [ ] `/src/app/(dashboard)/dashboard/media-providers/page.js`
- [ ] `/src/app/(dashboard)/dashboard/proxy-pools/page.js`
- [ ] `/src/app/(dashboard)/dashboard/models/page.js`
- [ ] `/src/app/(dashboard)/dashboard/oauth/page.js`
- [ ] `/src/app/(dashboard)/dashboard/basic-chat/page.js`
- [ ] `/src/app/(dashboard)/dashboard/console-log/page.js`
- [ ] `/src/app/(dashboard)/dashboard/mitm/page.js`
- [ ] `/src/app/(dashboard)/dashboard/profile/page.js`

### Shared Components (43 Missing)

- [ ] `/src/shared/components/ProviderForm.js`
- [ ] `/src/shared/components/ModelSelectModal.js`
- [ ] `/src/shared/components/EditConnectionModal.js`
- [ ] `/src/shared/components/ComboFormModal.js`
- [ ] `/src/shared/components/ProviderCard.js`
- [ ] `/src/shared/components/SettingsPanel.js`
- [ ] `/src/shared/components/Button.js`
- [ ] `/src/shared/components/Card.js`
- [ ] `/src/shared/components/Modal.js`
- [ ] `/src/shared/components/Drawer.js`
- [ ] `/src/shared/components/Badge.js`
- [ ] `/src/shared/components/Tabs.js`
- [ ] `/src/shared/components/Toast.js`
- [ ] `/src/shared/components/Tooltip.js`
- [ ] `/src/shared/components/Table.js`
- [ ] `/src/shared/components/FilterBar.js`
- [ ] `/src/shared/components/SearchBar.js`
- [ ] `/src/shared/components/Loading.js`
- [ ] `/src/shared/components/ErrorBoundary.js`
- [ ] `/src/shared/components/NotificationCenter.js`
- [ ] `/src/shared/components/StatusIndicator.js`
- [ ] `/src/shared/components/Avatar.js`
- [ ] `/src/shared/components/LanguageSwitcher.js`
- [ ] `/src/shared/components/Header.js`
- [ ] `/src/shared/components/Sidebar.js`
- [ ] `/src/shared/components/FormField.js`
- [ ] `/src/shared/components/FileUploadModal.js`
- [ ] `/src/shared/components/ImportModal.js`
- [ ] `/src/shared/components/ExportModal.js`
- [ ] `/src/shared/components/ChangelogModal.js`
- [ ] `/src/shared/components/ManualConfigModal.js`
- [ ] `/src/shared/components/CursorAuthModal.js`
- [ ] `/src/shared/components/KiroSocialOAuthModal.js`
- [ ] `/src/shared/components/NineRemoteButton.js`
- [ ] `/src/shared/components/NineRemotePromoModal.js`
- [ ] `/src/shared/components/NoAuthProxyCard.js`
- [ ] `/src/shared/components/FeatureCard.js`
- [ ] `/src/shared/components/SkillCard.js`
- [ ] `/src/shared/components/RequestLogger.js`
- [ ] `/src/shared/components/ResponseViewer.js`
- [ ] `/src/shared/components/AddCustomEmbeddingModal.js`
- [ ] `/src/shared/components/McpMarketplaceModal.js`
- [ ] `/src/shared/components/layouts/DashboardLayout.js`
- [ ] `/src/shared/components/layouts/AuthLayout.js`

### Hooks (2 Missing)

- [ ] `/src/shared/hooks/useCopyToClipboard.js`
- [ ] `/src/shared/hooks/useLocalStorage.js`

### Utilities (12 Missing)

- [ ] `/src/shared/utils/api.js`
- [ ] `/src/shared/utils/auth.js`
- [ ] `/src/shared/utils/format.js`
- [ ] `/src/shared/utils/storage.js`
- [ ] `/src/shared/utils/validation.js`
- [ ] `/src/shared/constants/index.js`
- [ ] `/src/shared/constants/colors.js`
- [ ] `/src/shared/constants/models.js`
- [ ] `/src/shared/constants/providers.js`
- [ ] `/src/shared/constants/pricing.js`
- [ ] `/src/shared/constants/skills.js`
- [ ] `/src/shared/constants/cliTools.js`

### State Management (5 Missing)

- [ ] `/src/store/authStore.js`
- [ ] `/src/store/providerStore.js`
- [ ] `/src/store/settingsStore.js`
- [ ] `/src/store/uiStore.js`
- [ ] `/src/store/usageStore.js`

---

## API ENDPOINTS

### Missing Routes (14+)

- [ ] `POST /api/v1/messages` - CL4ude Messages API
- [ ] `POST /api/v1/responses` - 0penAI Responses API
- [ ] `GET /api/models` - List models
- [ ] `GET /api/models/{kind}` - Models by kind
- [ ] `GET /api/models/info` - Model info
- [ ] `GET /api/providers` - List providers
- [ ] `POST /api/providers` - Create provider
- [ ] `PATCH /api/providers/{id}` - Update provider
- [ ] `DELETE /api/providers/{id}` - Delete provider
- [ ] `POST /api/providers/{id}/test` - Test provider
- [ ] `GET /api/settings` - Get settings
- [ ] `PATCH /api/settings` - Update settings
- [ ] `GET /api/proxy-pools` - List proxy pools
- [ ] `POST /api/proxy-pools` - Create pool
- [ ] `GET /api/cli-tools` - CLI tools config
- [ ] `GET /api/media-providers` - Media providers
- [ ] `GET /api/pricing` - Pricing info
- [ ] `GET /api/version` - Version info
- [ ] `POST /api/shutdown` - Graceful shutdown

---

## SUMMARY

| Category | Total | Implemented | Missing | % Complete |
|----------|-------|-------------|---------|-----------|
| Config | 12 | 3 | 9 | 25% |
| Executors | 20 | 20 | 0 | 100% |
| Handlers | 44 | 4 | 40 | 9% |
| Services | 8 | 0 | 8 | 0% |
| Utilities | 16 | 0 | 16 | 0% |
| Translators | 30 | 18 | 12 | 60% |
| Transformer | 2 | 2 | 0 | 100% |
| RTK | 17 | 17 | 0 | 100% |
| **Backend Total** | **149** | **64** | **85** | **43%** |
| Frontend Pages | 15 | 0 | 15 | 0% |
| Frontend Components | 43 | 0 | 43 | 0% |
| Frontend Hooks | 2 | 0 | 2 | 0% |
| Frontend Utils | 12 | 0 | 12 | 0% |
| Frontend Stores | 5 | 0 | 5 | 0% |
| **Frontend Total** | **77** | **0** | **77** | **0%** |
| **API Endpoints** | **24+** | **10** | **14+** | **42%** |
| **GRAND TOTAL** | **250+** | **74** | **176+** | **30%** |

