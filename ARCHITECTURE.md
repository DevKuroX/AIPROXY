# AIPROXY Architecture

## Overview

AIPROXY is a large-scale AI proxy/router platform inspired by 9router.

Current architecture:

- Go backend
- Next.js frontend
- PostgreSQL
- SSE streaming
- OAuth
- Multi-provider AI routing
- Thin frontend migration in progress

---

# Core Architecture

Browser
    ↓
Next.js Frontend
    ↓
Next.js API Thin Proxy Layer
    ↓
Go Backend API
    ↓
PostgreSQL / Redis / Providers

---

# Frontend Responsibilities

Frontend responsibilities:

- UI rendering
- client interaction
- local transient state
- stream rendering
- API communication
- route rendering
- loading/error handling

Frontend MUST NOT contain:

- SQLite
- embedded persistence
- business logic
- provider execution
- provider SDK logic
- analytics engines
- pricing logic
- filesystem logic
- routing engines
- auth authority
- model resolution

---

# Backend Responsibilities

Backend responsibilities:

- authentication
- authorization
- provider routing
- provider execution
- persistence
- analytics
- usage tracking
- OAuth/token management
- aliases
- combos
- pricing
- request logging
- stream normalization
- retry handling

Backend is the ONLY source of truth.

---

# Streaming Architecture

Frontend:
- renders chunks
- manages UI state

Backend:
- normalizes provider streams
- parses provider-specific formats
- handles retries
- handles provider adapters
- owns streaming business logic

---

# OAuth Architecture

Frontend:
- renders login flows
- redirects users
- handles temporary auth state

Backend:
- owns OAuth persistence
- manages tokens
- validates sessions
- manages auth business logic

---

# Critical Architectural Difference

9router:
- hybrid frontend/backend
- SQLite-based
- embedded persistence

AIPROXY:
- backend-authoritative
- PostgreSQL-based
- thin frontend target

This means:
- behavior parity is REQUIRED
- architecture parity is NOT required

Do NOT recreate 9router internal architecture.
Only preserve behavior and UX.
