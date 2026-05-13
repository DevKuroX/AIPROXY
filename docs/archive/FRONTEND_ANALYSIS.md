# 9Router Frontend & API Analysis

## Executive Summary

9Router adalah **self-hosted AI gateway** dengan dashboard Next.js yang sangat berbeda dari implementasi kita saat ini. Perbedaan utama:

### 1. **Authentication Model**
- **9Router**: Cookie-based auth dengan `auth_token` httpOnly cookie
- **Current**: JWT Bearer token di localStorage
- **Impact**: Frontend kita tidak compatible dengan 9router auth flow

### 2. **API Structure**
- **9Router**: `/api/*` (bukan `/api/admin/*`)
- **Current**: `/api/admin/*` untuk semua admin endpoints
- **Impact**: Semua endpoint path berbeda

### 3. **Dashboard Pages**
9Router memiliki halaman yang sangat berbeda:
- `/dashboard` → Endpoint management (bukan welcome page)
- `/dashboard/providers` → Provider connections
- `/dashboard/combos` → Fallback chains
- `/dashboard/usage` → Analytics & logs
- `/dashboard/cli-tools` → CLI tool configs (Cursor, Cline, Codex, dll)
- `/dashboard/mitm` → MITM proxy settings
- `/dashboard/translator` → Request/response translator
- `/dashboard/quota` → Usage quotas
- `/dashboard/skills` → Custom skills
- `/dashboard/basic-chat` → Test chat interface
- `/dashboard/profile` → User profile & password

**Current frontend tidak memiliki halaman-halaman ini!**

---

## API Endpoints Comparison

### Auth Endpoints

| 9Router | Current | Status |
|---------|---------|--------|
| `POST /api/auth/login` | `POST /api/login` | ❌ Different path & response |
| `POST /api/auth/logout` | - | ❌ Missing |
| Returns cookie | Returns JWT token | ❌ Incompatible |

### Settings

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/settings` | - | ❌ Missing |
| `PATCH /api/settings` | - | ❌ Missing |

Returns:
```json
{
  "requireApiKey": false,
  "requireLogin": true,
  "hasPassword": true,
  "tunnelEnabled": false,
  "tunnelUrl": "",
  "rtkEnabled": true,
  "cavemanEnabled": false,
  "outboundProxyEnabled": false,
  ...
}
```

### API Keys

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/keys` | `GET /api/admin/keys` | ❌ Different path |
| `POST /api/keys` | `POST /api/admin/keys` | ❌ Different path |
| `DELETE /api/keys/:id` | `DELETE /api/admin/keys/:id` | ❌ Different path |

Response format:
```json
{
  "keys": [
    {
      "id": "uuid",
      "name": "My Key",
      "key": "sk-...",  // Only on creation
      "machineId": "...",
      "createdAt": "2024-01-01T00:00:00Z",
      "lastUsedAt": null
    }
  ]
}
```

### Providers

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/providers` | `GET /api/admin/providers` | ❌ Different path & format |
| `POST /api/providers` | `POST /api/admin/providers` | ❌ Different path & format |
| `PATCH /api/providers/:id` | `PATCH /api/admin/providers/:id` | ❌ Different path |
| `DELETE /api/providers/:id` | `DELETE /api/admin/providers/:id` | ❌ Different path |
| `POST /api/providers/:id/test` | - | ❌ Missing |
| `GET /api/providers/:id/models` | - | ❌ Missing |
| `GET /api/providers/client` | - | ❌ Missing |

Provider object structure:
```json
{
  "id": "uuid",
  "provider": "openai|anthropic|gemini|...",
  "name": "My 0penAI",
  "enabled": true,
  "apiKey": "sk-...",  // Hidden in GET response
  "baseUrl": "https://api.openai.com/v1",
  "models": ["gpt-4", "gpt-3.5-turbo"],
  "providerSpecificData": {},
  "connectionProxyEnabled": false,
  "proxyPoolId": null,
  "createdAt": "...",
  "updatedAt": "..."
}
```

### Combos (Fallback Chains)

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/combos` | `GET /api/admin/combos` | ❌ Different path |
| `POST /api/combos` | `POST /api/admin/combos` | ❌ Different path |
| `PATCH /api/combos/:id` | `PATCH /api/admin/combos/:id` | ❌ Different path |
| `DELETE /api/combos/:id` | `DELETE /api/admin/combos/:id` | ❌ Different path |

Combo structure:
```json
{
  "id": "uuid",
  "name": "Production Chain",
  "model": "gpt-4",
  "providers": [
    { "connectionId": "uuid1", "priority": 1 },
    { "connectionId": "uuid2", "priority": 2 }
  ],
  "enabled": true,
  "createdAt": "...",
  "updatedAt": "..."
}
```

### Usage & Analytics

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/usage/stats` | `GET /api/admin/analytics/stats` | ❌ Different path |
| `GET /api/usage/logs` | `GET /api/admin/analytics/logs` | ❌ Different path |
| `GET /api/usage/chart` | - | ❌ Missing |
| `GET /api/usage/history` | - | ❌ Missing |
| `GET /api/usage/providers` | - | ❌ Missing |
| `GET /api/usage/request-logs` | - | ❌ Missing |
| `GET /api/usage/stream` | - | ❌ Missing (SSE) |

### Models

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/models` | - | ❌ Missing |
| `POST /api/models/alias` | `GET /api/admin/aliases` | ❌ Different |
| `DELETE /api/models/alias` | `DELETE /api/admin/aliases/:id` | ❌ Different |
| `GET /api/models/availability` | - | ❌ Missing |
| `POST /api/models/test` | - | ❌ Missing |

### Provider Nodes (Custom 0penAI-compatible)

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/provider-nodes` | `GET /api/admin/nodes` | ❌ Different path |
| `POST /api/provider-nodes` | `POST /api/admin/nodes` | ❌ Different path |
| `PATCH /api/provider-nodes/:id` | `PATCH /api/admin/nodes/:id` | ❌ Different path |
| `DELETE /api/provider-nodes/:id` | `DELETE /api/admin/nodes/:id` | ❌ Different path |
| `POST /api/provider-nodes/validate` | - | ❌ Missing |

### Pricing

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/pricing` | `GET /api/admin/pricing` | ❌ Different path |
| `POST /api/pricing` | `POST /api/admin/pricing` | ❌ Different path |
| `PATCH /api/pricing/:id` | `PATCH /api/admin/pricing/:id` | ❌ Different path |
| `DELETE /api/pricing/:id` | `DELETE /api/admin/pricing/:id` | ❌ Different path |

### OAuth

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/oauth/:provider/:action` | - | ❌ Missing |
| `POST /api/oauth/cursor/import` | - | ❌ Missing |
| `POST /api/oauth/kiro/import` | - | ❌ Missing |

### Tunnel (Cloudflare/Tailscale)

| 9Router | Current | Status |
|---------|---------|--------|
| `POST /api/tunnel/enable` | - | ❌ Missing |
| `POST /api/tunnel/disable` | - | ❌ Missing |
| `GET /api/tunnel/status` | - | ❌ Missing |
| `POST /api/tunnel/tailscale-*` | - | ❌ Missing (8 endpoints) |

### CLI Tools Config

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/cli-tools/cursor-settings` | - | ❌ Missing |
| `GET /api/cli-tools/cline-settings` | - | ❌ Missing |
| `GET /api/cli-tools/codex-settings` | - | ❌ Missing |
| `GET /api/cli-tools/copilot-settings` | - | ❌ Missing |
| ... (15 endpoints total) | - | ❌ Missing |

### Translator

| 9Router | Current | Status |
|---------|---------|--------|
| `POST /api/translator/translate` | - | ❌ Missing |
| `POST /api/translator/send` | - | ❌ Missing |
| `GET /api/translator/console-logs` | - | ❌ Missing |
| `GET /api/translator/console-logs/stream` | - | ❌ Missing (SSE) |

### Proxy Pools

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/proxy-pools` | - | ❌ Missing |
| `POST /api/proxy-pools` | - | ❌ Missing |
| `PATCH /api/proxy-pools/:id` | - | ❌ Missing |
| `DELETE /api/proxy-pools/:id` | - | ❌ Missing |
| `POST /api/proxy-pools/:id/test` | - | ❌ Missing |

### Other Endpoints

| 9Router | Current | Status |
|---------|---------|--------|
| `GET /api/health` | `GET /health` | ❌ Different path |
| `GET /api/init` | - | ❌ Missing |
| `POST /api/shutdown` | - | ❌ Missing |
| `GET /api/version` | - | ❌ Missing |

---

## Frontend Component Analysis

### Current Frontend Issues

1. **Wrong Dashboard Home**
   - Current: Generic welcome page
   - 9Router: Endpoint management with API keys, tunnel setup, RTK settings

2. **Missing Pages**
   - CLI Tools configuration
   - MITM proxy settings
   - Translator interface
   - Quota management
   - Skills management
   - Basic chat test interface
   - Profile/password management

3. **Wrong API Paths**
   - All `/api/admin/*` should be `/api/*`
   - Response formats don't match

4. **Missing Features**
   - Cloudflare Tunnel integration
   - Tailscale integration
   - RTK (Request Token Kompressor) toggle
   - Caveman prompt compression
   - Connection proxy settings
   - Proxy pool management

---

## Data Storage

9Router uses **SQLite** via `better-sqlite3`:
- Location: `~/.9router/data.db`
- Tables: `settings`, `api_keys`, `provider_connections`, `combos`, `usage_logs`, `provider_nodes`, `pricing`, `proxy_pools`

Current implementation uses **PostgreSQL** with different schema.

---

## Recommendations

### Option 1: Port 9Router Frontend (High Effort)
- Copy entire `/src/app/(dashboard)` from 9router
- Copy `/src/shared/components` 
- Copy `/src/lib` utilities
- Adapt to Go backend
- **Effort**: 2-3 weeks
- **Result**: 100% feature parity

### Option 2: Build Minimal Compatible Frontend (Medium Effort)
- Keep current React/Next.js structure
- Implement only core pages: endpoint, providers, combos, usage
- Match 9router API contracts
- **Effort**: 1 week
- **Result**: 60% feature parity

### Option 3: Use 9Router Frontend As-Is (Low Effort)
- Serve 9router frontend directly
- Implement Go backend to match 9router API exactly
- **Effort**: 3-4 days
- **Result**: 100% feature parity, but tied to 9router frontend

---

## Critical Missing Features in Current Implementation

1. **Settings Management** - No `/api/settings` endpoint
2. **Tunnel Integration** - No Cloudflare/Tailscale support
3. **CLI Tools Config** - No Cursor/Cline/Codex config generation
4. **Translator** - No request/response translation UI
5. **MITM Proxy** - No MITM proxy for debugging
6. **Proxy Pools** - No rotating proxy support
7. **Model Testing** - No test endpoint for providers
8. **Usage Streaming** - No SSE for real-time usage
9. **Skills** - No custom skills management
10. **Basic Chat** - No test chat interface

---

## Next Steps

**Decision Required:**
1. Which option to pursue? (Port frontend, build minimal, or use as-is)
2. Priority features to implement first?
3. Keep PostgreSQL or switch to SQLite for compatibility?

**Immediate Actions:**
1. Fix API paths (`/api/admin/*` → `/api/*`)
2. Implement `/api/settings` endpoint
3. Fix auth to use cookies instead of JWT in localStorage
4. Implement core dashboard pages (endpoint, providers, usage)
