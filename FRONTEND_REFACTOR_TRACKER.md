# Frontend Refactor Tracker

## Project

AIPROXY Frontend Refactor

Goal:
Convert frontend from hybrid 9router-style architecture into:
# Thin UI + Go backend architecture

---

# Current Status

Backend:
✅ Completed

Frontend:
⚠ Refactor in progress

---

# Current Frontend Problems

Current frontend still contains:
- SQLite assumptions
- local persistence remnants
- business logic leakage
- provider coupling
- hybrid architecture remnants

---

# Architecture Decision

Decision:
Keep Next.js API routes TEMPORARILY.

BUT:
Convert them into:
# thin proxy-only routes

Status:
FINAL

---

# Completed

## Backend
- Go backend implemented
- PostgreSQL integrated
- OAuth implemented
- Streaming implemented
- Provider executors implemented
- Analytics implemented

---

# In Progress

## Frontend migration

Current tasks:
- API layer stabilization
- usageDb removal
- localDb removal
- route proxy conversion
- dependency tracing

---

# Known Problem Areas

Potentially problematic modules:

```txt
src/lib/db.js
src/lib/storage.js
src/lib/usage.js
src/lib/oauth.js
src/app/api/*
src/sse/*
src/shared/*
```

---

# Forbidden Regressions

NEVER:
- reintroduce SQLite
- recreate deleted DB modules
- add fallback persistence
- duplicate backend logic
- create frontend provider execution
- create temporary compatibility hacks

---

# Safe Migration Strategy

## Phase 1
Stabilize API layer

Target:
```txt
src/lib/api/*
```

---

## Phase 2
Trace dependency graph

Before deleting modules:
1. locate imports
2. classify usages
3. replace usages
4. THEN delete

---

## Phase 3
Convert API routes

OLD:
```txt
route.ts -> SQLite/business logic
```

NEW:
```txt
route.ts -> thin backend proxy
```

---

## Phase 4
Feature migration

Recommended order:

1. providers
2. settings
3. usage
4. aliases
5. combos
6. OAuth
7. streaming

---

## Phase 5
SQLite cleanup

ONLY after:
- imports removed
- usages migrated
- routes stabilized

---

## Phase 6
Final stabilization

- dead code cleanup
- dependency cleanup
- final build stabilization
- runtime verification

---

# Build Strategy

DO NOT:
- run endless build loops
- patch errors randomly

Correct workflow:
1. trace imports
2. batch fixes
3. migrate systematically
4. run build after grouped changes

---

# Current Known Errors

Current unresolved issues:
- usageDb imports
- localDb remnants
- shared/constants/providers path inconsistencies

These are migration leftovers.
Do NOT patch randomly.

---

# Streaming Warning

Streaming is critical infrastructure.

Files:
```txt
src/sse/*
open-sse/*
backend/internal/stream/*
```

Do NOT casually rewrite streaming architecture.

---

# Task Execution Strategy

Good task:
```txt
Replace usageDb usage in settings page with api usage client.
```

Bad task:
```txt
Refactor frontend architecture completely.
```

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

Frontend should NOT contain:

- SQLite
- business logic
- provider execution
- analytics engines
- routing engines
- persistence systems