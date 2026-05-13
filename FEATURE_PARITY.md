revisi gabungin sama yang lu buat

# AIPROXY vs 9router Feature Parity

> **Analysis Date:** 2026-05-13
> **AIPROXY Status:** v1.0 Complete (Phase 0-20)

---

## Overview

AIPROXY adalah implementasi ulang 9router dalam Go dengan PostgreSQL. Berikut perbandingan fitur yang sudah diimplementasi vs yang ada di referensi 9router.

---

## Feature Parity Matrix

### Core Modules

| Modul | 9router (JS) | AIPROXY (Go) | Status |
|-------|--------------|--------------|--------|
| **Config** | 12 files | 7 files | ✅ Parity |
| **Executors** | 20 files | 20 files | ✅ Parity |
| **Handlers** | 6 dirs | 6 dirs | ✅ Parity |
| **Services** | 8 files | 16 files | ✅ Parity (lebih lengkap) |
| **RTK** | 7 files | 6 files | ✅ Parity |
| **Translator** | 5 dirs | 5 dirs | ✅ Parity |
| **Transformer** | 2 files | 1 dir | ✅ Parity |

### Services

| Service | 9router | AIPROXY | Status |
|---------|---------|---------|--------|
| accountFallback | ✅ | ✅ `account_fallback.go` | ✅ |
| combo | ✅ | ✅ `combo.go` | ✅ |
| compact | ✅ | ✅ `compact.go` | ✅ |
| model | ✅ | ✅ `model.go` | ✅ |
| projectId | ✅ | ✅ `project_id.go` | ✅ |
| provider | ✅ | ✅ `provider.go` | ✅ |
| tokenRefresh | ✅ | ✅ `token_refresh.go` | ✅ |
| usage | ✅ | ✅ `usage.go` | ✅ |

### Executors

| Executor | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| antigravity | ✅ | ✅ | ✅ |
| azure | ✅ | ✅ | ✅ |
| codex | ✅ | ✅ | ✅ |
| commandcode | ✅ | ✅ | ✅ |
| cursor | ✅ | ✅ | ✅ |
| gemini-cli | ✅ | ✅ `gemini_cli` | ✅ |
| github | ✅ | ✅ | ✅ |
| grok-web | ✅ | ✅ `grok_web` | ✅ |
| iflow | ✅ | ✅ | ✅ |
| kiro | ✅ | ✅ | ✅ |
| ollama-local | ✅ | ✅ `ollama_local` | ✅ |
| opencode | ✅ | ✅ | ✅ |
| opencode-go | ✅ | ✅ `opencode_go` | ✅ |
| perplexity-web | ✅ | ✅ `perplexity_web` | ✅ |
| qoder | ✅ | ✅ | ✅ |
| qwen | ✅ | ✅ | ✅ |
| vertex | ✅ | ✅ | ✅ |
| base | ✅ | ✅ | ✅ |
| default | ✅ | ✅ | ✅ |
| registry | - | ✅ | ➕ Extra |

### Embedding Providers

| Provider | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| base | ✅ `_base.js` | ✅ `base.go` | ✅ |
| gemini | ✅ | ✅ | ✅ |
| openai | ✅ | ✅ | ✅ |
| openaiCompatNode | ✅ | ✅ `openaiCompatible` | ✅ |

### TTS Providers

| Provider | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| base | - | ✅ `base.go` | ➕ Extra |
| edgeTTS | ✅ | ✅ `edgeTTS.go` | ✅ |
| elevenlabs | ✅ | ✅ | ✅ |
| gemini | ✅ | ✅ | ✅ |
| googleTTS | ✅ | ✅ | ✅ |
| localDevice | ✅ | ✅ | ✅ |
| openai | ✅ | ✅ | ✅ |
| openrouter | ✅ | ✅ | ✅ |
| genericFormats | ✅ | - | ⚠️ Merged |

### Image Providers

| Provider | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| blackForestLabs | ✅ | ✅ | ✅ |
| cloudflareAi | ✅ | ✅ | ✅ |
| codex | ✅ | ✅ | ✅ |
| comfyui | ✅ | ✅ | ✅ |
| falAi | ✅ | ✅ | ✅ |
| gemini | ✅ | ✅ | ✅ |
| huggingface | ✅ | ✅ | ✅ |
| nanobanana | ✅ | ✅ | ✅ |
| openai | ✅ | ✅ | ✅ |
| runwayml | ✅ | ✅ | ✅ |
| sdwebui | ✅ | ✅ | ✅ |
| stabilityAi | ✅ | ✅ | ✅ |

### RTK (Runtime Kit)

| Component | 9router | AIPROXY | Status |
|-----------|---------|---------|--------|
| apply | ✅ `applyFilter.js` | ✅ `apply.go` | ✅ |
| autodetect | ✅ | ✅ | ✅ |
| caveman | ✅ | ✅ | ✅ |
| cavemanPrompts | ✅ | ✅ `caveman_prompts` | ✅ |
| constants | ✅ | ✅ | ✅ |
| registry | ✅ | ✅ | ✅ |
| filters/ | ✅ | ✅ | ✅ |

### Config Files

| Config | 9router | AIPROXY | Status |
|--------|---------|---------|--------|
| providers | ✅ `providers.js` | ✅ `providers.go` | ✅ |
| models | ✅ `models.js` | ✅ `models.go` | ✅ |
| error_rules | ✅ `errorConfig.js` | ✅ `error_rules.go` | ✅ |
| endpoints | ✅ (inline) | ✅ `endpoints.go` | ✅ |
| pool | ✅ (inline) | ✅ `pool.go` | ✅ |
| appConstants | ✅ | - | ⚠️ Hardcoded |
| codexInstructions | ✅ | - | ⚠️ Hardcoded |
| defaultThinkingSignature | ✅ | - | ⚠️ Hardcoded |
| googleTtsLanguages | ✅ | - | ⚠️ Hardcoded |
| ollamaModels | ✅ | - | ⚠️ Hardcoded |
| providerModels | ✅ `36KB` | - | ⚠️ DB-driven |
| runtimeConfig | ✅ | - | ⚠️ Env vars |
| ttsModels | ✅ | - | ⚠️ DB-driven |

---

## Architecture Differences

### Database

| Aspect | 9router | AIPROXY |
|--------|---------|---------|
| **Engine** | SQLite (local) | PostgreSQL (server) |
| **Driver** | better-sqlite3 | pgx/v5 |
| **Migrations** | Custom JS | SQL files |
| **Schema** | `providerConnections` table | Same structure |

### OAuth Flow

| Aspect | 9router | AIPROXY |
|--------|---------|---------|
| **PKCE** | ✅ | ✅ |
| **Local Server** | ✅ Dynamic port | ✅ Dynamic port |
| **Token Storage** | SQLite JSON | PostgreSQL JSON |
| **Refresh** | ✅ Automatic | ✅ Automatic |

### Selection Strategy

| Strategy | 9router | AIPROXY |
|----------|---------|---------|
| fill-first | ✅ | ✅ |
| round-robin | ✅ | ✅ |
| Per-provider override | ✅ | ✅ |

---

## Missing / Different Implementation

### Hardcoded vs Config-driven

| Feature | 9router | AIPROXY | Impact |
|---------|---------|---------|--------|
| `providerModels.js` | Config file (36KB) | DB-driven | 🟡 Different approach |
| `codexInstructions.js` | Config file | Hardcoded | 🟡 Minor |
| `defaultThinkingSignature.js` | Config file | Hardcoded | 🟡 Minor |
| `googleTtsLanguages.js` | Config file | Hardcoded | 🟡 Minor |
| `ollamaModels.js` | Config file | Hardcoded | 🟡 Minor |
| `runtimeConfig.js` | Runtime config | Env vars | 🟢 Simpler |

### Potential Missing Features

| Feature | 9router | AIPROXY | Notes |
|---------|---------|---------|-------|
| **OpenAI format** | `openaiCompatNode.js` | ✅ `openaiCompatible.go` | Present |
| **genericFormats (TTS)** | ✅ Separate file | Merged into providers | 🟡 Minor |
| **Cloud mode** | ✅ `cloud/` dir | ❌ Not implemented | 🔴 Cloud deployment |
| **GitBook docs** | ✅ `gitbook/` dir | ❌ Not needed | 🟢 Different approach |

---

## Summary

### ✅ Feature Complete (Parity Achieved)

1. **Core Proxy** - Chat completions, embeddings, images, TTS, STT
2. **Provider Executors** - 20 executors matching 9router
3. **OAuth Flow** - PKCE, token refresh, multi-account pool
4. **RTK + Caveman** - Filter and transform responses
5. **Translator** - Format conversion (OpenAI ↔ Claude ↔ Gemini)
6. **Fallback** - Auto-retry with next account
7. **Usage Tracking** - Cost calculation, daily aggregation
8. **Combos** - Provider combinations for redundancy

### ⚠️ Different Implementation (Not Missing)

1. **Provider Models** - DB-driven instead of config file
2. **Runtime Config** - Environment variables instead of file
3. **Some constants** - Hardcoded in Go instead of config

### 🔴 Not Implemented (Intentional)

1. **Cloud Mode** - 9router has Cloudflare Workers deployment
2. **GitBook** - Documentation generation

### ➕ Extra in AIPROXY

1. **Registry executor** - Centralized executor registration
2. **Base TTS provider** - Cleaner abstraction
3. **Unit tests** - Comprehensive test coverage per module
4. **PostgreSQL** - More scalable than SQLite

---

## Recommendation

**AIPROXY sudah feature-complete** untuk penggunaan lokal/on-premise. Perbedaan utama adalah:

1. **Config files → Database**: Model definitions disimpan di database, bukan config file. Ini lebih fleksibel untuk admin panel.

2. **Cloud mode**: Tidak diimplementasi karena AIPROXY dirancang untuk deployment server tradisional, bukan Cloudflare Workers.

3. **Hardcoded constants**: Beberapa konstanta minor di-hardcode. Tidak critical karena jarang berubah.

**Tidak ada fitur penting yang hilang** dari 9router. Semua executor, handler, dan service sudah ada.

---

# FEATURE PARITY TRACKER

Goal:
AIPROXY frontend must preserve behavioral parity with 9router.

Do NOT redesign behavior unless explicitly approved.

---

# Authentication

9router behavior:
- password-only login
- no OAuth-first flow
- session persistence via cookies
- login redirect behavior preserved

AIPROXY target:
- preserve same login UX
- backend auth rewritten in Go
- frontend UI behavior unchanged

Status:
✅ backend implemented
⚠ frontend migration in progress

---

# Providers

9router behavior:
- provider CRUD
- provider health state
- enable/disable provider
- model mapping
- provider badges

Migration rule:
Frontend UI behavior must remain visually/functionally compatible.

---

# Streaming

9router behavior:
- live chunk rendering
- abort stream
- retry behavior
- token streaming order

Migration rule:
Preserve streaming UX exactly.

Backend implementation may differ internally.

---

# Settings

9router behavior:
- instant settings persistence
- optimistic updates
- theme persistence

Migration rule:
UI behavior preserved.
Persistence moved to Go backend.

---

# Forbidden Behavior Changes

NEVER:
- simplify flows
- remove UX edge cases
- redesign forms
- alter interaction flow
- change auth flow
- remove streaming features

without explicit approval.

# Frontend Behavioral Parity

## Authentication

9router behavior:
- password-only login
- cookie session persistence
- redirect after login
- no OAuth-first UX

AIPROXY requirement:
- preserve same frontend auth flow
- backend auth implementation may differ internally

Status:
⚠ frontend migration in progress

---

## Providers UI

9router behavior:
- provider grouping
- health badges
- enable/disable toggle
- filtering
- optimistic updates

AIPROXY requirement:
- preserve same interaction behavior
- backend source changed to Go API

---

## Streaming UX

9router behavior:
- live token streaming
- chunk ordering stable
- abort stream support
- retry behavior

AIPROXY requirement:
- preserve streaming UX parity

Backend implementation may differ internally.

---

## Settings UX

9router behavior:
- instant save
- optimistic updates
- persistent preferences

AIPROXY requirement:
- preserve UX
- persistence moved to backend

---

## Error Handling

9router behavior:
- inline errors
- retry visibility
- provider failure fallback visibility

AIPROXY requirement:
- preserve error visibility behavior

*Analysis generated: 2026-05-13*