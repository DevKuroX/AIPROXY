# AIPROXY Migration Task Status

> **Last Updated**: 2026-05-14
> **Current Phase**: Not started
> **Active Task**: None

---

# Status Legend

| Status | Meaning |
|--------|---------|
| `[ ]` | PENDING — Not started |
| `[~]` | IN_PROGRESS — Currently executing |
| `[x]` | DONE — Completed and verified |
| `[!]` | BLOCKED — Cannot proceed |
| `[-]` | SKIPPED — Explicitly excluded |
| `[?]` | REGRESSION — Previously done, now broken |

---

# Phase 0 — Pre-Migration Inventory

| Task | Status | Notes |
|------|--------|-------|
| T0.1 | `[x]` | SQLite import inventory |
| T0.2 | `[x]` | Legacy import inventory |
| T0.3 | `[x]` | Stream import inventory |
| T0.4 | `[x]` | Filesystem import inventory |
| T0.5 | `[x]` | Route cohort classification |
| T0.6 | `[ ]` | Backend URL decision |
| T0.7 | `[ ]` | API Contract update |
| T0.8 | `[ ]` | Backend coverage verification |
| T0.9 | `[ ]` | PM decisions collection |
| T0.10 | `[ ]` | HAR fixtures collection |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 1 — Shared Contracts Stabilization

| Task | Status | Notes |
|------|--------|-------|
| T1.1 | `[ ]` | Constants import inventory |
| T1.2 | `[ ]` | Constants index normalization |
| T1.3 | `[ ]` | Constants source comments |
| T1.4 | `[ ]` | API contracts interfaces |
| T1.5 | `[ ]` | Migrate duplicate types |
| T1.6 | `[ ]` | TypeScript check |
| T1.7 | `[ ]` | Build check |
| T1.8 | `[ ]` | Phase 1 smoke test |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 2 — SQLite Dependency Isolation

| Task | Status | Notes |
|------|--------|-------|
| T2.1 | `[ ]` | Required clients inventory |
| T2.2 | `[ ]` | Create HTTP helper |
| T2.3 | `[ ]` | Refactor admin-api |
| T2.4 | `[ ]` | Refactor API clients |
| T2.5 | `[ ]` | Create usage API client |
| T2.6 | `[ ]` | Create settings API client |
| T2.7 | `[ ]` | Create providers API client |
| T2.8 | `[ ]` | Create proxy pools API client |
| T2.9 | `[ ]` | Create combos API client |
| T2.10 | `[ ]` | Create pricing API client |
| T2.11 | `[ ]` | Create models API client |
| T2.12 | `[ ]` | Create keys API client |
| T2.13 | `[ ]` | Create db export API client |
| T2.14 | `[ ]` | ESLint SQLite blocking |
| T2.15 | `[ ]` | Deprecation comments |
| T2.16 | `[ ]` | Backend coverage verification |
| T2.17 | `[ ]` | Build verification |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 3 — Cohort A Route Proxy

| Task | Status | Notes |
|------|--------|-------|
| T3.0 | `[ ]` | Create proxy helper |
| T3.1 | `[ ]` | Proxy helper tests |
| T3.2 | `[ ]` | Convert auth/health/version |
| T3.3 | `[ ]` | Convert admin CRUD set 1 |
| T3.4 | `[ ]` | Convert admin CRUD set 2 |
| T3.5 | `[ ]` | Convert settings/usage/oauth |
| T3.6 | `[ ]` | HAR replay validation |
| T3.7 | `[ ]` | Component consumers migration |
| T3.8 | `[ ]` | Build and smoke |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 4 — Cohort B Route Audit

| Task | Status | Notes |
|------|--------|-------|
| T4.1 | `[ ]` | init/ route |
| T4.2 | `[ ]` | shutdown/ route |
| T4.3 | `[ ]` | locale/ route |
| T4.4 | `[ ]` | cli-tools/ route |
| T4.5 | `[ ]` | cloud/ route |
| T4.6 | `[ ]` | tunnel/ route |
| T4.7 | `[ ]` | translator/ route |
| T4.8 | `[ ]` | v1beta proxy streaming |
| T4.9 | `[ ]` | v1 proxy streaming |
| T4.10 | `[ ]` | HAR replay verification |
| T4.11 | `[ ]` | Streaming verification matrix |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 5 — Auth + Token Unification

| Task | Status | Notes |
|------|--------|-------|
| T5.1 | `[ ]` | Token paths audit |
| T5.2 | `[ ]` | Token decision lock |
| T5.3 | `[ ]` | Single token accessor |
| T5.4 | `[ ]` | UserStore refactor |
| T5.5 | `[ ]` | Auth fetch unification |
| T5.6 | `[ ]` | Login flow audit |
| T5.7 | `[ ]` | OAuth callback audit |
| T5.8 | `[ ]` | Logout flow audit |
| T5.9 | `[ ]` | Auth regression manual test |
| T5.10 | `[ ]` | Direct token accessors |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 6 — Legacy Module Deprecation

| Task | Status | Notes |
|------|--------|-------|
| T6.1 | `[ ]` | Deprecate oauth |
| T6.2 | `[ ]` | Deprecate tunnel |
| T6.3 | `[ ]` | Deprecate updater |
| T6.4 | `[ ]` | Deprecate usage |
| T6.5 | `[ ]` | Deprecate network |
| T6.6 | `[ ]` | Deprecate mitm |
| T6.7 | `[ ]` | Deprecate providerNormalization |
| T6.8 | `[ ]` | Deprecate initCloudSync |
| T6.9 | `[ ]` | Classify consoleLogBuffer |
| T6.10 | `[ ]` | Build smoke per batch |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 7 — Streaming Normalization

| Task | Status | Notes |
|------|--------|-------|
| T7.1 | `[ ]` | Stream contract docs |
| T7.2 | `[ ]` | Event types verification |
| T7.3 | `[ ]` | Provider format leaks |
| T7.4 | `[ ]` | Create SSE consumer |
| T7.5 | `[ ]` | SSE consumer tests |
| T7.6 | `[ ]` | Slim SSE handlers |
| T7.7 | `[ ]` | Stream consumers inventory |
| T7.8 | `[ ]` | Chat UI migration |
| T7.9 | `[ ]` | Stream feature flag |
| T7.10 | `[ ]` | Streaming parity test |
| T7.11 | `[ ]` | Delete open-sse |
| T7.12 | `[ ]` | Zero imports verification |
| T7.13 | `[ ]` | Remove feature flag |
| T7.14 | `[ ]` | Lint rule open-sse |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 8 — SQLite Engine Removal

| Task | Status | Notes |
|------|--------|-------|
| T8.1 | `[ ]` | Zero importers verification |
| T8.2 | `[ ]` | Delete shim files |
| T8.3 | `[ ]` | Delete SQLite DB directory |
| T8.4 | `[ ]` | Delete dataDir.js |
| T8.5 | `[ ]` | Remove SQLite package |
| T8.6 | `[ ]` | SQLite driver search |
| T8.7 | `[ ]` | Build smoke SQLite-sensitive |
| T8.8 | `[ ]` | Lint rule SQLite |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 9 — Dual App Shell Collapse

| Task | Status | Notes |
|------|--------|-------|
| T9.1 | `[ ]` | Diff page files |
| T9.2 | `[ ]` | Diff layout files |
| T9.3 | `[ ]` | Port missing behaviors |
| T9.4 | `[ ]` | Delete dual shell JS |
| T9.5 | `[ ]` | Config audit |
| T9.6 | `[ ]` | Visual diff |
| T9.7 | `[ ]` | Build smoke |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 10 — Dead Code Cleanup

| Task | Status | Notes |
|------|--------|-------|
| T10.1 | `[ ]` | Dependency cleanup |
| T10.2 | `[ ]` | Unused exports cleanup |
| T10.3 | `[ ]` | Relocate test scripts |
| T10.4 | `[ ]` | Delete jsconfig |
| T10.5 | `[ ]` | i18n orphan cleanup |
| T10.6 | `[ ]` | Audit CLAUDE.md |
| T10.7 | `[ ]` | Bundle size measurement |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 11 — Architecture Lock

| Task | Status | Notes |
|------|--------|-------|
| T11.1 | `[ ]` | Create FRONTEND_ARCHITECTURE.md |
| T11.2 | `[ ]` | Update STRUCTURE.md |
| T11.3 | `[ ]` | Update tracker |
| T11.4 | `[ ]` | CI architecture enforcement |
| T11.5 | `[ ]` | Pre-commit hook |
| T11.6 | `[ ]` | Feature parity verification |
| T11.7 | `[ ]` | HAR regression sweep |
| T11.8 | `[ ]` | Streaming parity test |
| T11.9 | `[ ]` | Final sign-off |

**Phase Status**: `[ ]` NOT STARTED

---

# Summary

| Phase | Total | Done | Blocked | Progress |
|-------|-------|------|---------|----------|
| P0 | 10 | 0 | 0 | 0% |
| P1 | 8 | 0 | 0 | 0% |
| P2 | 17 | 0 | 0 | 0% |
| P3 | 9 | 0 | 0 | 0% |
| P4 | 11 | 0 | 0 | 0% |
| P5 | 10 | 0 | 0 | 0% |
| P6 | 10 | 0 | 0 | 0% |
| P7 | 14 | 0 | 0 | 0% |
| P8 | 8 | 0 | 0 | 0% |
| P9 | 7 | 0 | 0 | 0% |
| P10 | 7 | 0 | 0 | 0% |
| P11 | 9 | 0 | 0 | 0% |
| **TOTAL** | **120** | **0** | **0** | **0%** |
