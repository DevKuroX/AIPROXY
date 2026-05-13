# 9Router Phase Prompt Guidelines

## Overview

This document provides guidelines for creating accurate phase prompts for the 9Router project. Understanding the architecture is critical for generating contextually appropriate prompts.

---

## Key Architectural Principles

### 1. **Modular Separation**

- **`open-sse/`** - Provider-agnostic routing engine (can be published as npm package)
- **`src/`** - Next.js application (dashboard + API routes)
- **`src/mitm/`** - MITM proxy system (optional, for transparent interception)

**Implication**: Prompts should clarify which layer they're targeting.

### 2. **Provider Abstraction**

All providers follow a common pattern:
- **Executor** - Handles provider-specific request/response
- **Format** - Defines request/response structure
- **Translator** - Converts between formats
- **OAuth Service** - Handles authentication

**Implication**: Adding a new provider requires implementing these components.

### 3. **Format Translation**

9Router translates between 9 formats:
- 0penAI (standard)
- CL4ude ()
- Gemini (Google)
- Responses API (0penAI)
- Cursor
- Antigravity
- Kiro
- CommandCode
- Ollama

**Implication**: Any feature affecting request/response must consider all formats.

### 4. **Multi-Tier Fallback**

```
Tier 1: Subscription (quota tracking)
  ↓ quota exhausted
Tier 2: Cheap (budget limits)
  ↓ budget exceeded
Tier 3: Free (unlimited)
```

**Implication**: Features must respect tier boundaries and fallback logic.

### 5. **Token Optimization (RTK)**

RTK saves 20-40% tokens via:
- Caveman compression algorithm
- Tool output filtering (10 filters)
- Smart truncation
- Deduplication

**Implication**: RTK is applied at the request level, before sending to providers.

### 6. **Database Abstraction**

4 adapters support different environments:
- better-sqlite3 (primary, fastest)
- sql.js (fallback, pure JS)
- Node.js sqlite
- Bun sqlite

**Implication**: Database code must be adapter-agnostic.

### 7. **MITM Proxy System**

Optional transparent interception for:
- Token rotation
- Request modification
- Response caching
- Provider-specific handling

**Implication**: MITM handlers are provider-specific and optional.

---

## Phase Prompt Categories

### Category A: Core Routing Logic

**Scope**: `open-sse/` module

**Key Files**:
- `handlers/chatCore.js` - Main chat handler
- `services/provider.js` - Provider config lookup
- `services/accountFallback.js` - Fallback logic
- `translator/index.js` - Format translation

**Considerations**:
- Must support all 40+ providers
- Must handle all 9 formats
- Must respect fallback tiers
- Must apply RTK compression

**Example Prompt**:
> "Implement a new feature in the chat routing layer that [feature]. This must work across all 40+ providers and support format translation between 0penAI, CL4ude, and Gemini formats. Consider the multi-tier fallback system (subscription → cheap → free) and ensure RTK token compression is applied before sending requests."

---

### Category B: Provider Integration

**Scope**: `open-sse/executors/` + `open-sse/config/providers.js`

**Key Files**:
- `executors/[provider].js` - Provider executor
- `config/providers.js` - Provider config
- `src/lib/oauth/services/[provider].js` - OAuth implementation

**Considerations**:
- Each provider has unique request/response format
- OAuth providers need token refresh logic
- API key providers need key storage
- Compatible nodes need endpoint configuration

**Example Prompt**:
> "Add support for [new provider]. Implement the executor in `open-sse/executors/[provider].js`, add provider config to `open-sse/config/providers.js`, and implement OAuth in `src/lib/oauth/services/[provider].js`. Ensure format translation works with existing translators."

---

### Category C: Format Translation

**Scope**: `open-sse/translator/`

**Key Files**:
- `translator/formats.js` - Format definitions
- `translator/request/` - Request translators
- `translator/response/` - Response translators
- `translator/helpers/` - Translation utilities

**Considerations**:
- Each format has unique structure
- Translators must be bidirectional
- Tool calls must be preserved
- Streaming must be handled

**Example Prompt**:
> "Implement translation between [format1] and [format2]. Add request translator in `translator/request/[format1]-to-[format2].js` and response translator in `translator/response/[format2]-to-[format1].js`. Ensure tool calls are preserved and streaming is supported."

---

### Category D: Token Optimization (RTK)

**Scope**: `open-sse/rtk/`

**Key Files**:
- `rtk/index.js` - RTK orchestration
- `rtk/caveman.js` - Compression algorithm
- `rtk/filters/` - Compression filters

**Considerations**:
- RTK is applied before sending to providers
- Filters are applied based on tool output type
- Compression must preserve semantic meaning
- Must work with all formats

**Example Prompt**:
> "Implement a new RTK compression filter for [tool output type]. Add filter in `rtk/filters/[name].js`, register in `rtk/registry.js`, and ensure it preserves semantic meaning while reducing token count. Test with various output sizes."

---

### Category E: Database Operations

**Scope**: `src/lib/db/`

**Key Files**:
- `db/driver.js` - Database driver
- `db/repos/` - Data repositories
- `db/adapters/` - Database adapters

**Considerations**:
- Must work with all 4 adapters
- Repositories abstract data access
- Migrations handle schema changes
- KV store for simple data

**Example Prompt**:
> "Add a new data repository for [entity]. Create `src/lib/db/repos/[entity]Repo.js` with CRUD operations. Ensure it works with all 4 database adapters (better-sqlite3, sql.js, Node.js, Bun). Add migration if schema changes are needed."

---

### Category F: API Endpoints

**Scope**: `src/app/api/`

**Key Files**:
- `api/v1/` - 0penAI-compatible endpoints
- `api/v1beta/` - Beta endpoints
- `api/providers/` - Provider management
- `api/oauth/` - OAuth flow

**Considerations**:
- v1 endpoints must be 0penAI-compatible
- Management endpoints use database repos
- OAuth endpoints handle provider auth
- All endpoints must handle errors gracefully

**Example Prompt**:
> "Implement a new API endpoint at `/api/[path]` that [functionality]. Use the appropriate database repository, handle errors with proper HTTP status codes, and document the request/response format."

---

### Category G: Dashboard UI

**Scope**: `src/app/(dashboard)/`

**Key Files**:
- `dashboard/providers/` - Provider management UI
- `dashboard/combos/` - Combo builder UI
- `dashboard/cli-tools/` - CLI tool configuration
- `dashboard/mitm/` - MITM settings

**Considerations**:
- Uses React 19 + Tailwind CSS
- Uses Zustand for state management
- Calls management APIs
- Must be responsive

**Example Prompt**:
> "Create a new dashboard page at `/dashboard/[page]` that [functionality]. Use React components, Tailwind CSS for styling, and Zustand for state. Call the appropriate management API endpoints."

---

### Category H: MITM Proxy System

**Scope**: `src/mitm/`

**Key Files**:
- `mitm/server.js` - MITM proxy server
- `mitm/handlers/` - Provider-specific handlers
- `mitm/cert/` - Certificate management

**Considerations**:
- MITM is optional (not required for basic routing)
- Handlers are provider-specific
- Certificates must be generated and installed
- DNS configuration may be needed

**Example Prompt**:
> "Implement MITM handler for [provider] in `src/mitm/handlers/[provider].js`. This handler should intercept requests to [provider] and [modification]. Ensure certificate generation and installation work on Windows, macOS, and Linux."

---

### Category I: OAuth & Token Management

**Scope**: `src/lib/oauth/` + `open-sse/services/tokenRefresh.js`

**Key Files**:
- `oauth/services/[provider].js` - OAuth implementation
- `services/tokenRefresh.js` - Token refresh logic
- `oauth/utils/` - OAuth utilities (PKCE, server, UI)

**Considerations**:
- Each provider has unique OAuth flow
- Token refresh must handle expiry
- PKCE flow for security
- Local OAuth server for callback

**Example Prompt**:
> "Implement OAuth for [provider]. Create `src/lib/oauth/services/[provider].js` with authorization and token exchange. Add token refresh logic to `open-sse/services/tokenRefresh.js`. Ensure PKCE flow is used and token expiry is detected."

---

### Category J: Media Providers (Images, TTS, Embeddings)

**Scope**: `open-sse/handlers/[imageProviders|ttsProviders|embeddingProviders]/`

**Key Files**:
- `handlers/imageGenerationCore.js` - Image handler
- `handlers/ttsCore.js` - TTS handler
- `handlers/embeddingsCore.js` - Embedding handler
- `handlers/[type]Providers/` - Provider implementations

**Considerations**:
- Each media type has different providers
- Providers have different request/response formats
- Must support streaming where applicable
- Format translation may be needed

**Example Prompt**:
> "Add support for [media provider] in `open-sse/handlers/[type]Providers/[provider].js`. Implement the base provider interface, handle request/response translation, and ensure it works with the [type]Core handler."

---

## Prompt Template

Use this template for creating phase prompts:

```markdown
# Phase [N]: [Feature Name]

## Objective
[Clear description of what this phase accomplishes]

## Scope
- **Primary Module**: [open-sse/ | src/ | src/mitm/ | etc.]
- **Key Files**: [List of files to modify/create]
- **Dependencies**: [Other phases or components this depends on]

## Requirements
- [ ] [Requirement 1]
- [ ] [Requirement 2]
- [ ] [Requirement 3]

## Architectural Considerations
- **Provider Support**: [How this affects provider support]
- **Format Translation**: [How this affects format translation]
- **Fallback Logic**: [How this affects fallback]
- **Token Optimization**: [How this affects RTK]
- **Database**: [How this affects database]

## Implementation Notes
- [Note 1]
- [Note 2]
- [Note 3]

## Testing
- [ ] Unit tests for [component]
- [ ] Integration tests with [provider]
- [ ] E2E tests for [feature]

## Success Criteria
- [Criterion 1]
- [Criterion 2]
- [Criterion 3]
```

---

## Common Pitfalls to Avoid

### 1. **Ignoring Format Translation**
❌ "Add support for [feature]"
✅ "Add support for [feature] with translation for all 9 formats"

### 2. **Forgetting Fallback Logic**
❌ "Route requests to provider"
✅ "Route requests to provider with multi-tier fallback (subscription → cheap → free)"

### 3. **Not Considering RTK**
❌ "Send request to provider"
✅ "Apply RTK compression, then send request to provider"

### 4. **Ignoring Database Adapters**
❌ "Use SQLite to store data"
✅ "Use database abstraction layer to support all 4 adapters"

### 5. **Forgetting OAuth Token Refresh**
❌ "Use OAuth provider"
✅ "Use OAuth provider with automatic token refresh and expiry detection"

### 6. **Not Testing All Providers**
❌ "Test with one provider"
✅ "Test with multiple providers (OAuth, API key, compatible node)"

### 7. **Ignoring Streaming**
❌ "Handle responses"
✅ "Handle both streaming (SSE) and non-streaming (JSON) responses"

### 8. **Forgetting Error Handling**
❌ "Call provider API"
✅ "Call provider API with error handling, fallback, and user-friendly error messages"

---

## Phase Prompt Examples

### Example 1: Adding a New Provider

```markdown
# Phase X: Add Support for [Provider Name]

## Objective
Enable 9Router to route requests to [Provider Name] with full format translation and fallback support.

## Scope
- **Primary Module**: open-sse/
- **Key Files**:
  - open-sse/executors/[provider].js (new)
  - open-sse/config/providers.js (modify)
  - src/lib/oauth/services/[provider].js (new)
  - open-sse/translator/request/openai-to-[provider].js (new)
  - open-sse/translator/response/[provider]-to-openai.js (new)

## Requirements
- [ ] Implement executor for [Provider Name]
- [ ] Add provider config with baseUrl, format, headers
- [ ] Implement OAuth flow (if applicable)
- [ ] Add request/response translators
- [ ] Support all models from [Provider Name]
- [ ] Handle provider-specific errors
- [ ] Add unit tests

## Architectural Considerations
- **Provider Support**: [Provider Name] is [OAuth|API Key|Compatible Node]
- **Format Translation**: Translate between 0penAI and [Provider Format]
- **Fallback Logic**: [Provider Name] is Tier [1|2|3]
- **Token Optimization**: RTK applies before sending to [Provider Name]
- **Database**: Store [Provider Name] credentials in connections repo

## Implementation Notes
- [Provider Name] uses [authentication method]
- [Provider Name] supports models: [list]
- [Provider Name] has rate limits: [details]
- [Provider Name] requires headers: [list]

## Testing
- [ ] Unit tests for executor
- [ ] Unit tests for translators
- [ ] Integration test with [Provider Name]
- [ ] E2E test with CLI tool

## Success Criteria
- Requests route successfully to [Provider Name]
- Format translation works correctly
- Token refresh works (if OAuth)
- Fallback works when [Provider Name] is unavailable
- All models are available
```

### Example 2: Adding RTK Compression Filter

```markdown
# Phase Y: Add RTK Compression Filter for [Tool Output]

## Objective
Reduce token usage for [tool output type] by 30-40% using RTK compression.

## Scope
- **Primary Module**: open-sse/rtk/
- **Key Files**:
  - open-sse/rtk/filters/[name].js (new)
  - open-sse/rtk/registry.js (modify)
  - open-sse/rtk/index.js (modify)

## Requirements
- [ ] Implement compression filter for [tool output]
- [ ] Register filter in registry
- [ ] Preserve semantic meaning
- [ ] Achieve 30-40% token reduction
- [ ] Add unit tests
- [ ] Add integration tests

## Architectural Considerations
- **Token Optimization**: Filter reduces tokens for [tool output]
- **Format Translation**: Filter works with all formats
- **Fallback Logic**: Filter applies before fallback
- **Database**: Store filter statistics in usage repo

## Implementation Notes
- [Tool output] typically contains [structure]
- Compression strategy: [description]
- Semantic preservation: [how meaning is preserved]
- Edge cases: [list]

## Testing
- [ ] Unit tests with various output sizes
- [ ] Integration tests with all formats
- [ ] Semantic preservation tests
- [ ] Token reduction measurement

## Success Criteria
- Token reduction: 30-40%
- Semantic meaning preserved
- Works with all formats
- No errors on edge cases
```

---

## Checklist for Phase Prompt Creation

- [ ] Clearly define objective
- [ ] Specify primary module(s)
- [ ] List all files to modify/create
- [ ] Include architectural considerations
- [ ] Consider all 40+ providers
- [ ] Consider all 9 formats
- [ ] Consider fallback logic
- [ ] Consider RTK compression
- [ ] Consider database layer
- [ ] Include testing requirements
- [ ] Define success criteria
- [ ] Avoid common pitfalls
- [ ] Use consistent terminology
- [ ] Include implementation notes
- [ ] Consider error handling

