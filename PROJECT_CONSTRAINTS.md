# AIPROXY - ROOT AI RULES

THIS DOCUMENT IS AUTHORITATIVE.

All AI coding agents MUST read and follow this document before making ANY changes.

Applies to:
- Claude
- GPT
- GLM
- Qwen
- OpenCode
- Kiro
- RooCode
- Copilot
- Any autonomous coding agent

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

# CURRENT MIGRATION STATUS

Backend:
✅ Completed

Frontend:
⚠ Hybrid architecture still exists

Current frontend still contains:
- old API routes
- local DB assumptions
- usageDb references
- storage coupling
- business logic leakage

Migration is IN PROGRESS.

---

# MOST IMPORTANT RULES

## NEVER:

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

IMPORTANT:

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

NOT:

Browser
    ↓
Next.js
    ↓
SQLite/business logic

---

# REFACTOR STRATEGY

DO NOT:
- mass rewrite
- mass delete
- autonomous redesign

DO:
- incremental migration
- preserve existing UI contracts
- preserve component behavior
- preserve store shapes
- preserve response formats when possible

---

# BUILD RULES

NEVER:
- enter infinite build loops
- patch errors randomly
- run build after every tiny fix

Instead:
1. trace dependency graph
2. batch related fixes
3. migrate systematically
4. run build after grouped changes

---

# DELETE RULES

Before deleting ANY module:
1. locate ALL imports
2. replace ALL usages
3. verify migration path
4. THEN delete

Never delete first.

---

# STATE MANAGEMENT RULES

Allowed:
- React state
- Zustand
- TanStack Query

Forbidden:
- SQLite-backed state
- filesystem persistence
- hidden persistence layers

---

# STREAMING RULES

Streaming is core infrastructure.

Frontend:
- renders stream chunks

Backend:
- normalizes provider streams
- parses provider formats
- handles retries
- handles provider adapters

Frontend MUST NOT implement provider-specific stream parsing.

---

# AUTH RULES

OAuth/token logic belongs to backend.

Frontend may:
- render login UI
- redirect users
- store temporary auth state

Frontend may NOT:
- manage OAuth persistence
- implement token business logic
- become auth authority

---

# IMPORT RULES

Never create duplicate systems.

If a module already exists:
- extend carefully
- preserve exports
- preserve interfaces

Avoid:
- duplicate wrappers
- temporary adapters
- alternate stores

---

# MIGRATION PRIORITY

Priority order:

1. API thin layer
2. Shared types/contracts
3. Feature-by-feature migration
4. Streaming normalization
5. Dependency cleanup
6. SQLite removal
7. Final stabilization

---

# CURRENT KNOWN PROBLEMS

Known unresolved issues:
- usageDb imports
- localDb dependencies
- shared constants path inconsistencies
- API route business logic
- hybrid frontend/backend boundaries

These must be fixed SYSTEMATICALLY.

---

# WHEN UNSURE

STOP.

DO NOT:
- assume architecture
- improvise redesigns
- create temporary hacks
- invent systems

Ask for clarification instead.

---

# REQUIRED DOCUMENTS

AI agents MUST also read:

- STRUCTURE.md
- ARCHITECTURE.md
- FRONTEND_RULES.md
- API_CONTRACT.md
- MIGRATION_PLAN.md

These documents OVERRIDE assumptions.

---

# FINAL GOAL

Target architecture:

Frontend:
- pure UI
- thin proxy/API layer only
- no persistence
- no business logic

Backend:
- single source of truth
- all providers
- all logic
- all persistence
- all analytics

This architecture is FINAL.