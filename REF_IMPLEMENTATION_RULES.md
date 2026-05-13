# Reference Implementation Rules

9router is the behavioral reference implementation.

AIPROXY backend may differ internally.

BUT:
Frontend behavior must remain compatible unless explicitly changed.

---

# Mandatory Workflow

Before editing ANY feature:

1. Locate equivalent 9router implementation
2. Understand feature behavior
3. Compare with current AIPROXY implementation
4. Preserve UX behavior
5. THEN migrate

---

# Never Implement Blindly

DO NOT:
- rewrite from assumptions
- simplify unknown logic
- remove edge cases
- invent replacement flows

Always reference:
- original route
- original component
- original behavior

---

# Required Per-Task Reference

Every migration task must include:

- source 9router file(s)
- target AIPROXY file(s)
- preserved behaviors
- intentionally changed behaviors

---

# Example

GOOD:

Source:
```txt
9router/src/app/providers/page.tsx
```

Target:
```txt
AIPROXY/frontend/src/app/providers/page.tsx
```

Preserve:
- sorting
- filters
- badges
- loading state

Changed:
- data source now backend API

---

BAD:

"rewrite providers page cleaner"

# Migration Rules

AIPROXY is NOT a redesign of 9router.

It is:
- a backend rewrite in Go
- with frontend architecture decoupling

Behavioral parity with 9router must be preserved unless explicitly changed.

---

# Mandatory Workflow

Before modifying ANY feature:

1. Locate equivalent 9router implementation
2. Understand original behavior
3. Compare with AIPROXY implementation
4. Preserve UX and behavior
5. THEN migrate/refactor

Never implement blindly.

---

# Forbidden Actions

NEVER:
- simplify unknown flows
- remove edge-case handling
- redesign UX without approval
- rewrite features from assumptions
- replace systems "because cleaner"

9router is the behavioral reference implementation.

# Per-Task Reference Rules

Every migration/refactor task MUST include:

- Source 9router file(s)
- Target AIPROXY file(s)
- Behavior being preserved
- Internal implementation changes
- Known parity risks

---

# GOOD TASK

Source:
```txt
9router/src/app/providers/page.tsx
```

Target:
```txt
frontend/src/app/providers/page.tsx
```

Preserve:
- sorting
- badges
- filters
- loading state

Changed:
- data source moved to Go backend

---

# BAD TASK

"rewrite providers page cleaner"