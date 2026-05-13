# AIPROXY - ROOT AI RULES

THIS DOCUMENT IS AUTHORITATIVE.

All AI coding agents MUST read and follow this document before making ANY changes.

---

# PROJECT CONTEXT

AIPROXY is a large-scale AI proxy platform inspired by 9router.

Architecture:
- Go backend
- Next.js frontend
- PostgreSQL
- OAuth
- SSE streaming
- Multi-provider executors

The backend is already implemented and production-grade.

Current work focuses on:
# FRONTEND ARCHITECTURE REFACTOR

Goal:
Convert frontend from hybrid fullstack architecture into:
# UI-FIRST THIN CLIENT ARCHITECTURE

---

# CRITICAL ARCHITECTURE RULE

Frontend is NOT allowed to contain:
- SQLite
- embedded persistence
- provider business logic
- provider SDK logic
- filesystem logic
- analytics calculation
- request routing logic
- authentication core logic

Frontend responsibilities:
- UI rendering
- API calls
- transient client state
- streaming rendering
- user interaction

Backend responsibilities:
- ALL persistence
- ALL provider execution
- ALL business logic
- OAuth/token management
- analytics
- routing
- pricing
- model resolution

---

# NEVER

- recreate deleted SQLite modules
- recreate usageDb.js
- recreate localDb.js
- add compatibility shims
- add fallback persistence systems
- redesign architecture
- invent new abstractions
- move logic back into frontend
- create temporary hacks

---

# API ROUTE RULE

Next.js API routes are TEMPORARILY ALLOWED.

BUT:
They MUST become:
# THIN PROXY LAYERS ONLY

Allowed:
```ts
fetch(BACKEND_URL)
```

Forbidden:
- database access
- local business logic
- SQLite
- analytics calculation
- provider execution

---

# FRONTEND DATA FLOW

Correct architecture:

Browser
    ↓
Next.js UI/API thin layer
    ↓
Go backend
    ↓
PostgreSQL/providers

---

# REFERENCE IMPLEMENTATION

9router is the behavioral reference implementation.

AIPROXY backend may differ internally.

BUT:
Frontend behavior must remain compatible unless explicitly changed.

---

# MIGRATION WORKFLOW

Before modifying ANY feature:

1. Locate equivalent 9router implementation
2. Understand feature behavior
3. Compare with current AIPROXY implementation
4. Preserve UX behavior
5. THEN migrate

---

# WHEN UNSURE

STOP.

DO NOT:
- assume architecture
- improvise redesigns
- create temporary hacks
- invent systems

Ask for clarification instead.
