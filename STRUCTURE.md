# AIPROXY - Struktur Proyek Lengkap

> **Terakhir diperbarui:** 2026-05-13
> **Status:** v1.0 Complete

---

## Overview

**AIPROXY** adalah AI API proxy server yang mendukung multi-provider, OAuth authentication, dan real-time streaming. Arsitektur fullstack dengan backend Go dan frontend Next.js 16.

| Komponen | Teknologi | Status |
|----------|-----------|--------|
| Backend | Go 1.21+ | Production |
| Frontend | Next.js 16.2.6 + React 19 | Production |
| Database | PostgreSQL (pgx/v5) | Production |
| Auth | JWT + OAuth 2.0 | Production |

---

## Struktur Root

```
ai_proxy/
├── backend/                # Go backend server
├── frontend/               # Next.js frontend application
├── docs/                   # Documentation
├── .gitignore
├── .gitmodules
└── STRUCTURE.md
```

---

## Backend (`/backend`)

Backend adalah REST API server berbasis Go dengan arsitektur modular.

```
backend/
├── cmd/server/             # Application entrypoint
├── internal/               # Private application code
│   ├── api/               # HTTP routing & handlers
│   ├── auth/              # Authentication & authorization
│   ├── config/            # Configuration management
│   ├── executor/          # Provider executors (20+ providers)
│   ├── handlers/          # Request handlers
│   ├── models/            # Data models
│   ├── services/          # Business logic
│   ├── storage/           # Database operations
│   ├── translator/        # API format translation
│   └── rtk/               # Runtime Kit (Caveman)
└── _ref/9router/          # Reference implementation
```

---

## Frontend (`/frontend`)

Frontend adalah aplikasi Next.js 16 dengan React 19 dan TypeScript.

```
frontend/
├── src/
│   ├── app/               # Next.js App Router
│   ├── lib/               # Utility libraries
│   ├── models/            # Data models
│   ├── shared/            # Shared components
│   ├── sse/               # Server-Sent Events
│   └── store/             # State management
├── open-sse/              # Shared SSE utilities
└── package.json
```

---

## Provider Executors

AIPROXY mendukung 20+ provider AI dengan executor modular.

| Provider | Alias | Format | Auth Type |
|----------|-------|--------|-----------|
| CL4ude | `cc` | claude | OAuth 2.0 |
| 0penAI Codex | `cx` | openai-responses | OAuth 2.0 |
| Gemini CLI | `gc` | gemini-cli | OAuth 2.0 |
| GitHub Copilot | `gh` | openai | OAuth 2.0 |
| Qwen | `qw` | openai | OAuth 2.0 |
| iFlow | `if` | openai | OAuth 2.0 |
| Antigravity | `ag` | antigravity | OAuth 2.0 |
| Kiro | `kr` | kiro | OAuth 2.0 |
| Cursor | `cu` | cursor | Cookie Import |
| 0penAI | - | openai | API Key |
| Gemini API | - | gemini | API Key |
| Azure 0penAI | - | openai | API Key |
| Vertex AI | - | vertex | Service Account |

---

## OAuth Account Pool

OAuth endpoint digunakan untuk mengelola "akun pool" provider - bukan untuk login user.

**Konsep:**
```
Provider → Multiple Connections (akun berbeda)
                   ↓
          Round-robin / Fill-first selection
                   ↓
          Fallback jika rate-limited
```

---

## API Endpoints

### V1 Endpoints (0penAI-compatible)

| Endpoint | Description |
|----------|-------------|
| `/v1/chat/completions` | Chat completions |
| `/v1/embeddings` | Generate embeddings |
| `/v1/models` | List models |
| `/v1/images/generations` | Image generation |
| `/v1/audio/speech` | Text-to-speech |
| `/v1/audio/transcriptions` | Speech-to-text |

### Admin Endpoints

| Endpoint | Description |
|----------|-------------|
| `/admin/providers` | Provider configuration |
| `/admin/keys` | API key management |
| `/admin/combos` | Combo management |
| `/admin/oauth` | OAuth configuration |
| `/admin/usage` | Usage statistics |

---

## Development

### Environment Variables

```bash
# Backend (.env)
DATABASE_URL=postgres://user:pass@host:5432/db
JWT_SECRET=your-secret-key
ADMIN_PASSWORD=admin
PORT=20128
```

### Build & Run

```bash
# Backend
cd backend
go build ./cmd/server
./server

# Frontend
cd frontend
npm install
npm run dev
```

---

*Last updated: 2026-05-13*
