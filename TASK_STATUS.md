# AIPROXY Migration Task Status

> **Last Updated**: 2026-05-14
> **Current Phase**: Phase 0 - Pre-Migration Inventory
> **Active Task**: T0.1

---

# Status Legend

| Status | Meaning |
|--------|---------|
| `[ ]` | PENDING - Not started |
| `[~]` | IN_PROGRESS - Currently executing |
| `[x]` | DONE - Completed and verified |
| `[!]` | BLOCKED - Cannot proceed |
| `[-]` | SKIPPED - Explicitly excluded |
| `[?]` | REGRESSION - Previously done, now broken |

---

# Phase 0 - Pre-Migration Inventory

| Task | Status | Notes |
|------|--------|-------|
| T0.1 | `[!]` | blocked - cascade_outside_scope: frontend/login-test.js:1:22 @typescript-eslint/no-require-imports |
| T0.2 | `[ ]` | Legacy import inventory |
| T0.3 | `[ ]` | Stream import inventory |
| T0.4 | `[ ]` | Filesystem import inventory |
| T0.5 | `[ ]` | Route cohort classification |
| T0.6 | `[ ]` | Backend URL decision |
| T0.7 | `[ ]` | API Contract update |
| T0.8 | `[ ]` | Backend coverage verification |
| T0.9 | `[ ]` | PM decisions collection |
| T0.10 | `[ ]` | HAR fixtures collection |
| T0.11 | `[ ]` | Preflight repair before execution resumes |

**Phase Status**: `[~]` IN PROGRESS

---

# Phase 1 - Shared Contracts Stabilization

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

# Phase 2 - SQLite Dependency Isolation

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

# Phase 3 - Cohort A Route Proxy

| Task | Status | Notes |
|------|--------|-------|
| T3.0 | `[ ]` | Create proxy helper |
| T3.1 | `[ ]` | Proxy helper tests |
| T3.2 | `[ ]` | Convert auth/health/version routes |
| T3.3 | `[ ]` | Convert providers routes |
| T3.4 | `[ ]` | Convert settings routes |
| T3.5 | `[ ]` | Convert usage routes |
| T3.6 | `[ ]` | Convert aliases routes |
| T3.7 | `[ ]` | Convert combos routes |
| T3.8 | `[ ]` | Phase 3 exit verification |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 4 - Cohort B Route Audit

| Task | Status | Notes |
|------|--------|-------|
| T4.1 | `[ ]` | Init routes decision |
| T4.2 | `[ ]` | Shutdown routes decision |
| T4.3 | `[ ]` | Locale routes decision |
| T4.4 | `[ ]` | CLI tools routes decision |
| T4.5 | `[ ]` | Cloud routes decision |
| T4.6 | `[ ]` | Tunnel routes decision |
| T4.7 | `[ ]` | Translator routes decision |
| T4.8 | `[ ]` | v1beta proxy streaming |
| T4.9 | `[ ]` | v1 proxy streaming |
| T4.10 | `[ ]` | HAR replay verification |
| T4.11 | `[ ]` | Streaming verification matrix |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 5 - Auth + Token Unification

| Task | Status | Notes |
|------|--------|-------|
| T5.1 | `[ ]` | Token source audit |
| T5.2 | `[ ]` | Unify token retrieval |
| T5.3 | `[ ]` | Update API clients token |
| T5.4 | `[ ]` | Remove token from userStore |
| T5.5 | `[ ]` | Update components token access |
| T5.6 | `[ ]` | Update login flow |
| T5.7 | `[ ]` | Update logout flow |
| T5.8 | `[ ]` | Update route auth integration |
| T5.9 | `[ ]` | Auth regression manual test |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 6 - Legacy Module Deprecation

| Task | Status | Notes |
|------|--------|-------|
| T6.1 | `[ ]` | Deprecate OAuth module |
| T6.2 | `[ ]` | Deprecate tunnel module |
| T6.3 | `[ ]` | Deprecate updater module |
| T6.4 | `[ ]` | Deprecate usage module |
| T6.5 | `[ ]` | Deprecate network module |
| T6.6 | `[ ]` | Deprecate MITM module |
| T6.7 | `[ ]` | Deprecate provider normalization |
| T6.8 | `[ ]` | Deprecate cloud sync module |
| T6.9 | `[ ]` | Classify console log buffer |
| T6.10 | `[ ]` | Phase 6 verification |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 7 - Streaming Normalization

| Task | Status | Notes |
|------|--------|-------|
| T7.1 | `[ ]` | Inspect backend stream format |
| T7.2 | `[ ]` | Confirm event types |
| T7.3 | `[ ]` | Check provider format leaks |
| T7.4 | `[ ]` | Create SSE consumer |
| T7.5 | `[ ]` | Unit test SSE consumer |
| T7.6 | `[ ]` | Slim SSE handlers |
| T7.7 | `[ ]` | Inventory stream consumers |
| T7.8 | `[ ]` | Migrate chat UI |
| T7.9 | `[ ]` | Add stream feature flag |
| T7.10 | `[ ]` | Streaming parity test |
| T7.11 | `[ ]` | Delete open-sse directory |
| T7.12 | `[ ]` | Verify zero open-sse imports |
| T7.13 | `[ ]` | Remove stream feature flag |
| T7.14 | `[ ]` | Add open-sse lint rule |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 8 - SQLite Engine Removal

| Task | Status | Notes |
|------|--------|-------|
| T8.1 | `[ ]` | Re-verify zero shim importers |
| T8.2 | `[ ]` | Delete shim files |
| T8.3 | `[ ]` | Delete SQLite DB directory |
| T8.4 | `[ ]` | Delete dataDir.js |
| T8.5 | `[ ]` | Remove SQLite from package.json |
| T8.6 | `[ ]` | Search direct SQLite imports |
| T8.7 | `[ ]` | Build and smoke test |
| T8.8 | `[ ]` | Add architecture enforcement lint |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 9 - Dual App Shell Collapse

| Task | Status | Notes |
|------|--------|-------|
| T9.1 | `[ ]` | Delete layout.js |
| T9.2 | `[ ]` | Delete page.js |
| T9.3 | `[ ]` | Verify single app shell |
| T9.4 | `[ ]` | Delete dual shell JS |
| T9.5 | `[ ]` | Remove .js imports for layout |
| T9.6 | `[ ]` | Verify TypeScript config |
| T9.7 | `[ ]` | Phase 9 exit verification |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 10 - Dead Code Cleanup

| Task | Status | Notes |
|------|--------|-------|
| T10.1 | `[ ]` | Dependency cleanup |
| T10.2 | `[ ]` | Run depcheck |
| T10.3 | `[ ]` | Run ts-prune |
| T10.4 | `[ ]` | Remove unused files |
| T10.5 | `[ ]` | Relocate test files |
| T10.6 | `[ ]` | Document bundle size |
| T10.7 | `[ ]` | Phase 10 exit verification |

**Phase Status**: `[ ]` NOT STARTED

---

# Phase 11 - Architecture Lock

| Task | Status | Notes |
|------|--------|-------|
| T11.1 | `[ ]` | Create frontend ARCHITECTURE.md |
| T11.2 | `[ ]` | Update STRUCTURE.md |
| T11.3 | `[ ]` | Update FRONTEND_REFACTOR_TRACKER.md |
| T11.4 | `[ ]` | CI architecture enforcement |
| T11.5 | `[ ]` | Add pre-commit hook |
| T11.6 | `[ ]` | Feature parity verification |
| T11.7 | `[ ]` | Run HAR replay verification |
| T11.8 | `[ ]` | Final streaming parity test |
| T11.9 | `[ ]` | Sign-off |

**Phase Status**: `[ ]` NOT STARTED

---

# Summary

| Phase | Total | Done | Progress |
|-------|-------|------|----------|
| P0 | 11 | 0 | 0% |
| P1 | 8 | 0 | 0% |
| P2 | 17 | 0 | 0% |
| P3 | 9 | 0 | 0% |
| P4 | 11 | 0 | 0% |
| P5 | 9 | 0 | 0% |
| P6 | 10 | 0 | 0% |
| P7 | 14 | 0 | 0% |
| P8 | 8 | 0 | 0% |
| P9 | 7 | 0 | 0% |
| P10 | 7 | 0 | 0% |
| P11 | 9 | 0 | 0% |
| **TOTAL** | **120** | **0** | **0%** |

---

# Next Tasks

1. **T0.1** - SQLite import inventory
2. **T0.2** - Legacy module import inventory
3. **T0.3** - Streaming module import inventory
4. **T0.4** - Filesystem import inventory
5. **T0.5** - Route cohort classification
