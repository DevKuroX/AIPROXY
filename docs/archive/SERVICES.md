# Core Services Layer

The services layer contains pure business logic that orchestrates providers, accounts, fallback, and token management.

## Architecture

Services are stateless and do NOT handle HTTP. All state is stored in PostgreSQL or passed via context. Services are called by handlers and other services.

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Handler   │────▶│   Service   │────▶│   Storage   │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ HTTP Client │ (for external APIs)
                    └─────────────┘
```

## Services Overview

### 1. AccountFallbackService

Intelligent account-level fallback with exponential backoff.

**Location:** `internal/services/account_fallback.go`

**Key Functions:**
- `GetQuotaCooldown(backoffLevel)` - Exponential backoff: 2s, 4s, 8s... max 5min
- `CheckFallbackError(status, errorText, backoffLevel)` - Config-driven error matching
- `IsAccountUnavailable(unavailableUntil)` - Check if cooldown expired
- `GetUnavailableUntil(cooldownMs)` - Calculate ISO timestamp

**Error Rules:**
Text rules are checked first, then status codes:
- Text: "no credentials", "rate limit", "quota exceeded", "overloaded"
- Status: 401, 402, 403, 404, 429

**Ref:** `open-sse/services/accountFallback.js`

---

### 2. ComboService

Model combo handling with fallback and round-robin strategies.

**Location:** `internal/services/combo.go`

**Key Functions:**
- `GetRotatedModels(models, comboName, strategy, stickyLimit)` - Round-robin rotation
- `ResetComboRotation(comboName)` - Reset rotation state
- `GetComboModelsFromData(modelStr, combosData)` - Lookup combo models

**Strategies:**
- `fallback` - Try models in order until success
- `round-robin` - Rotate through models with sticky limit

**Ref:** `open-sse/services/combo.js`

---

### 3. CompactService

Response compaction for bandwidth optimization.

**Location:** `internal/services/compact.go`

**Key Functions:**
- `CompactResponse(response)` - Remove thinking, reduce verbosity
- `RemoveThinking(content)` - Strip thinking blocks
- `CompactUsage(usage)` - Reduce usage data
- `CompactStreamingLine(line)` - Handle SSE data lines

**Ref:** `open-sse/services/compact.js`

---

### 4. ModelService

Model resolution and provider alias mapping.

**Location:** `internal/services/model.go`

**Key Functions:**
- `ParseModel(modelStr)` - Parse "provider/model" or "alias/model" format
- `ResolveProviderAlias(aliasOrId)` - Resolve alias to provider ID
- `ResolveModelAlias(ctx, alias)` - Resolve model alias from database
- `GetModelInfo(ctx, modelStr)` - Full resolution with database lookup

**Provider Aliases:**
- `cc` → claude
- `cx` → codex
- `gc` → gemini-cli
- `qw` → qwen
- `gh` → github
- `kr` → kiro
- And 80+ more...

**Ref:** `open-sse/services/model.js`

---

### 5. ProjectIDService

Google Cloud Code project ID fetching with caching.

**Location:** `internal/services/project_id.go`

**Key Functions:**
- `GetProjectID(ctx, connectionID, accessToken)` - Fetch and cache project ID
- `CleanupNow()` - Evict stale cache entries
- `StartCacheCleanup()` / `StopCacheCleanup()` - Background cleanup

**Caching:**
- Cache TTL: 1 hour
- Pending fetch TTL: 2 minutes
- Cleanup interval: 10 minutes

**Ref:** `open-sse/services/projectId.js`

---

### 6. ProviderService

Provider detection and URL/header building.

**Location:** `internal/services/provider.go`

**Key Functions:**
- `DetectFormat(body)` - Detect request format (openai, claude, gemini, openai-responses)
- `BuildUpstreamURL(provider, baseURL, apiType)` - Construct provider URLs
- `BuildHeaders(provider, creds)` - Build auth headers
- `TestConnection(ctx, providerID)` - Test provider connectivity

**Format Detection:**
- `openai` - stream_options, response_format, logprobs
- `claude` - messages with content array
- `gemini` - contents array
- `openai-responses` - input field instead of messages

**Ref:** `open-sse/services/provider.js`

---

### 7. TokenRefreshService

OAuth token refresh orchestration with deduplication.

**Location:** `internal/services/token_refresh.go`

**Key Functions:**
- `RefreshIfNeeded(ctx, providerID, accountID)` - Refresh if expired
- `ForceRefresh(ctx, providerID, accountID)` - Force refresh
- `NeedsRefresh(ctx, providerID, accountID)` - Check if refresh needed
- `IsUnrecoverableRefreshError(result)` - Check for unrecoverable errors

**Provider-Specific Refresh:**
- GitHub, CL4ude, Gemini, Codex, Qwen, Kiro

**Deduplication:**
Concurrent refresh requests are deduplicated using sync.RWMutex and sync.WaitGroup.

**Token Expiry Buffer:** 5 minutes (provider-specific overrides available)

**Ref:** `open-sse/services/tokenRefresh.js`

---

### 8. UsageService

Enhanced usage tracking with provider quota fetching.

**Location:** `internal/services/usage.go`

**Key Functions:**
- `RecordUsage(ctx, usage)` - Record usage with cost calculation
- `GetUsageSummary(ctx, filters)` - Aggregated statistics
- `GetUsageByProvider(ctx, filters)` - Usage by provider
- `GetUsageByModel(ctx, filters)` - Usage by model
- `CalculateCost(ctx, usage)` - Cost calculation from pricing table
- `GetProviderQuota(ctx, provider, accessToken)` - Fetch provider quota

**Provider Quota APIs:**
- GitHub: `api.github.com/copilot_internal/user`
- Gemini: `cloudcode-pa.googleapis.com/v1internal:retrieveUserQuota`
- CL4ude: `api.anthr0pic.com/api/oauth/usage`
- Codex: `chatgpt.com/backend-api/wham/usage`

**Ref:** `open-sse/services/usage.js`

---

## Usage

### Initialization

```go
import (
    "github.com/DevKuroX/AIPROXY/internal/router"
    "github.com/DevKuroX/AIPROXY/internal/storage"
)

func main() {
    db := storage.NewDB(pool)
    logger := slog.Default()
    httpClient := &http.Client{}
    encryptionKey := []byte("your-encryption-key")
    
    services := router.InitServices(db, logger, httpClient, encryptionKey)
    
    // Start background tasks
    services.StartBackgroundTasks()
    defer services.StopBackgroundTasks()
}
```

### Using Services

```go
services := router.GetServices()

// Check fallback error
result := services.AccountFallback.CheckFallbackError(429, "rate limit exceeded", 0)
if result.ShouldFallback {
    cooldown := services.AccountFallback.GetUnavailableUntil(result.CooldownMs)
}

// Resolve model
modelInfo, err := services.Model.GetModelInfo(ctx, "cc/claude-3-opus")

// Compact response
compacted := services.Compact.CompactResponse(response)
```

## Database Schema

Migration `012_account_fallback.sql` adds:
- `account_fallback_state` - Per-account cooldown and backoff state
- `account_health_metrics` - Rolling success/failure rates
- `combo_rotation_state` - Round-robin rotation persistence

## Testing

```bash
go test ./internal/services/... -v
```

## 9router Parity

All services include `// ref:` comments pointing to the exact source location in 9router for verification and traceability.
