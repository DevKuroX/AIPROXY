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

The backend is already implemented.

The frontend is currently being migrated from:
- hybrid fullstack architecture

into:
- thin UI-first architecture

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

# Final Architecture Goal

## Frontend

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

## Backend

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

# Frontend API Routes

Current frontend contains:
```txt
src/app/api/*
```

These routes are TEMPORARILY retained.

However:
They MUST become:
# thin proxy routes only

Allowed:
```ts
fetch(BACKEND_URL)
```

Forbidden:
- SQLite access
- local business logic
- provider execution
- analytics calculations
- local persistence

---

# Streaming Architecture

Streaming is critical infrastructure.

Frontend:
- renders chunks
- manages UI state

Backend:
- normalizes provider streams
- parses provider-specific formats
- handles retries
- handles provider adapters
- owns streaming business logic

Frontend MUST NOT parse provider-native stream formats.

---

# State Management

Allowed:
- React state
- Zustand
- TanStack Query

Forbidden:
- SQLite-backed state
- filesystem-backed state
- hidden persistence layers

---

# Persistence

Persistence belongs ONLY to backend.

Frontend must never:
- store critical state locally
- implement hidden DB layers
- recreate SQLite systems

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

# Shared Contracts

Frontend MUST follow backend API contracts exactly.

Do not:
- invent endpoints
- invent response formats
- create frontend-only schemas

---

# Migration Philosophy

Migration must be:
- incremental
- deterministic
- minimally invasive

Avoid:
- mass rewrites
- speculative redesigns
- architecture improvisation

---

# Known Legacy Areas

Potentially problematic areas:

```txt
src/lib/*
src/app/api/*
src/sse/*
src/shared/*
```

These areas may still contain:
- old persistence assumptions
- business logic leakage
- hybrid architecture remnants

Migration must remove these safely.

---

# Final Frontend Target

Frontend should ultimately contain ONLY:

- pages
- components
- hooks
- stores
- API wrappers
- stream renderers
- UI utilities

# Frontend Migration Status

## Current State

Frontend still contains:
- SQLite assumptions
- local persistence remnants
- business logic leakage
- hybrid backend/frontend patterns

---

## Current Refactor Goal

Convert frontend into:

- thin UI layer
- thin proxy routes
- backend-driven architecture

WITHOUT changing:
- user behavior
- interaction flow
- UX expectations

---

## Current Known Issues

- usageDb imports still unresolved
- localDb remnants
- shared constants path inconsistencies
- API route business logic leakage

These are migration leftovers.
Do NOT patch randomly.

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

Nothing else.