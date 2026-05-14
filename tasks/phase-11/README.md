# Phase 11 — Final Stabilization & Architecture Lock

**Objective**: Lock the architecture in place via tooling and documentation. Make regressions impossible.

**Branch**: `phase/11-architecture-lock`

**Dependencies**: Phase 1-10

**Exit Gate**: CI enforces the rules; documentation reflects final state; three-party sign-off.

---

## Tasks

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T11.1](T11.1-create-frontend-architecture-md.md) | Create FRONTEND_ARCHITECTURE.md | new doc | 🟢 |
| [T11.2](T11.2-update-structure-md.md) | Update STRUCTURE.md | doc | 🟢 |
| [T11.3](T11.3-update-tracker.md) | Update FRONTEND_REFACTOR_TRACKER.md | doc | 🟢 |
| [T11.4](T11.4-ci-architecture-enforcement.md) | CI architecture enforcement | `.github/workflows/*` | 🟠 |
| [T11.5](T11.5-pre-commit-hook.md) | Pre-commit hook | `.husky/*` | 🟡 |
| [T11.6](T11.6-feature-parity-verification.md) | Feature parity verification | regression report | 🔴 |
| [T11.7](T11.7-har-regression-sweep.md) | HAR regression sweep | diff report | 🟠 |
| [T11.8](T11.8-streaming-parity-test.md) | Streaming parity test | report | 🔴 |
| [T11.9](T11.9-final-sign-off.md) | Final sign-off | tracker | 🟢 |

---

## Architecture Lock Enforcement

CI must block:

```
# SQLite
better-sqlite3, sqlite3
@/lib/db, @/lib/usageDb, @/lib/localDb

# Filesystem (in client code)
import from 'fs' in src/components, src/lib, etc.

# Legacy
open-sse
```

## Documentation Updates

- `FRONTEND_ARCHITECTURE.md` — thin-client model
- `STRUCTURE.md` — final layout (no open-sse, no db/)
- `FRONTEND_REFACTOR_TRACKER.md` — migration complete

## Final Verification

- FEATURE_PARITY.md items verified
- HAR regression green
- Streaming parity green

## Exit Criteria

- [ ] CI enforces architecture rules
- [ ] All FEATURE_PARITY.md items verified
- [ ] Full HAR + streaming regression green
- [ ] Three-party sign-off (PM + backend + frontend)
