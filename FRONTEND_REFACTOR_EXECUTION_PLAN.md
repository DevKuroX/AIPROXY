# AIPROXY — Frontend Refactor Execution Plan

> **Status:** Planning document — execution-ready
> **Scope:** `frontend/` only. Backend (Go) is feature-complete and authoritative.
> **Generated against repo:** `DevKuroX/AIPROXY@main`
> **Reference docs (authoritative):** `PROJECT_CONSTRAINTS.md`, `ARCHITECTURE.md`, `API_CONTRACT.md`, `FEATURE_PARITY.md`, `FRONTEND_REFACTOR_TRACKER.md`, `STRUCTURE.md`
> **Behavioral reference:** 9router. **Architectural reference:** AIPROXY target (thin client).
> **Hard rule:** Behavior parity required. Architecture parity NOT required.

---

## 0. Observations From the Current Repository

Before planning, here are the empirical facts that drive every decision in this document. These are not assumptions; they were verified directly from the `main` branch.

### 0.1 Concrete frontend reality

| Area | Status |
|---|---|
| `frontend/src/app/page.js` + `page.tsx` | **Both exist** — dual app shell (JS legacy + TS new) |
| `frontend/src/app/layout.js` + `layout.tsx` | **Both exist** — dual root layout |
| `frontend/src/lib/usageDb.js` | **Shim** re-exporting from `@/lib/db/index.js` (SQLite still alive) |
| `frontend/src/lib/localDb.js` | **Shim** re-exporting from `@/lib/db/index.js` (SQLite still alive) |
| `frontend/src/lib/db/` | **EXISTS** — contains the actual SQLite DB layer |
| `frontend/src/lib/dataDir.js` | **Still references `APP_NAME = "9router"`** + filesystem `DATA_DIR` |
| `frontend/src/lib/api.ts` | NEW thin client. Targets `NEXT_PUBLIC_API_URL` (default `http://localhost:20128`). Token via `localStorage` |
| `frontend/src/lib/admin-api.ts` | NEW client for `/api/admin/*` on Go backend |
| `frontend/src/lib/analytics-api.ts` | NEW backend-driven analytics client |
| `frontend/src/lib/oauth-api.ts` | NEW backend-driven OAuth client |
| `frontend/src/lib/oauth/`, `tunnel/`, `updater/`, `usage/`, `network/` | Legacy logic dirs — **business logic leakage** |
| `frontend/src/app/api/` | **24 subroutes** — many still contain logic, not proxies |
| `frontend/open-sse/` | **9router-derived streaming pipeline LIVING IN FRONTEND** (config/, executors/, handlers/, rtk/, services/, transformer/, translator/, utils/) — **highest risk area** |
| `frontend/src/sse/` | Frontend stream handlers/services/utils |
| `frontend/src/mitm/` | MITM/filter layer — needs classification |
| `frontend/src/shared/constants/` | Known to have path inconsistencies (per tracker) |
| `frontend/src/store/` | Zustand stores: `userStore`, `themeStore`, `settingsStore`, `providerStore`, `notificationStore`, `headerSearchStore`, `index` |
| `frontend/test-*.js`, `serve-static.js`, `login-test.js` | Ad-hoc scripts at frontend root — to be classified or relocated |
| Backend authoritative base | `http://localhost:20128` (per `api.ts`). `STRUCTURE.md` mentions `:14322` and admin password endpoints. **Source of truth must be unified in `lib/api.ts`.** |

### 0.2 What this implies

1. The frontend is **mid-migration**, not pre-migration. Two architectures coexist (`.js` legacy + `.ts` new).
2. The "delete usageDb / localDb" task is **NOT** simple — they are shims, but the shimmed target (`@/lib/db/`) is the **real** SQLite layer that must be removed first.
3. `open-sse/` is a full provider-streaming engine embedded in the frontend. This is the single biggest architectural violation. It is also the **highest-risk** rewrite target because backend already owns stream normalization.
4. There is a **dual app shell** (`page.js` + `page.tsx`, `layout.js` + `layout.tsx`). Whichever Next.js picks is non-deterministic without explicit resolution — this is a latent regression bomb.
5. Backend port inconsistencies (`:20128` vs `:14322`) suggest config drift. This must be the **first** thing locked down before any migration step.

---

## 1. Current Frontend Architecture Analysis

### 1.1 Hybrid patterns (observed)

| Pattern | Evidence | Risk |
|---|---|---|
| Dual file extensions per route (`.js` + `.tsx`) | `app/page.js`+`page.tsx`, `app/layout.js`+`layout.tsx` | 🔴 Critical — non-deterministic resolution |
| Shim files masking SQLite | `lib/usageDb.js`, `lib/localDb.js` | 🟠 High — false sense of completion |
| Logic in Next API routes | 24 subroutes under `app/api/` | 🟠 High — needs route-by-route audit |
| Streaming engine in frontend | `open-sse/` (executors, translator, RTK, transformer) | 🔴 Critical — duplicates backend |
| Filesystem coupling | `lib/dataDir.js` references `APP_NAME = "9router"` and `process.platform === "win32"` paths | 🟠 High — frontend should never touch FS |
| Token in `localStorage` | `lib/api.ts` uses `localStorage` directly | 🟡 Medium — UX parity OK, but auth UX must match 9router cookie behavior per FEATURE_PARITY |
| Multiple API clients | `api.ts`, `admin-api.ts`, `analytics-api.ts`, `oauth-api.ts` + legacy shared/services | 🟡 Medium — fan-out, but contained |

### 1.2 Business logic leakage (suspected hotspots)

Verify and classify each (Phase 0 task in §4):
- `src/lib/oauth/` — OAuth flow logic
- `src/lib/tunnel/` — Tunnel handling
- `src/lib/updater/` + `lib/appUpdater.js` — App updater logic
- `src/lib/usage/` — Usage aggregation
- `src/lib/network/` — Network/proxy logic
- `src/lib/db/` — SQLite engine
- `src/lib/mitmAliasCache.js` + `src/mitm/` — MITM aliasing
- `src/lib/providerNormalization.js` — Provider normalization
- `src/lib/initCloudSync.js` — Cloud sync init
- `src/lib/consoleLogBuffer.js` — Console buffering (likely UI-only, classify)
- `src/lib/disabledModelsDb.js`, `requestDetailsDb.js` — Stub/shim files (177B/166B) — almost certainly re-exports of removed/db layer
- `src/app/api/v1/*`, `v1beta/*` — Likely real OpenAI/Gemini-compatible logic, not proxies
- `src/app/api/translator/*` — Format translation; backend already owns translator (`backend/internal/translator/`)
- `src/app/api/init/*` — App initialization (possibly DB seeding — forbidden in target)

### 1.3 Persistence coupling

- **SQLite chain:** `usageDb.js` → `db/index.js` → SQLite driver
- **SQLite chain:** `localDb.js` → `db/index.js` → SQLite driver
- **Filesystem chain:** `dataDir.js` → `os.homedir()` / `APPDATA` → `.9router/` folder
- **Implicit chain:** Any `src/app/api/*/route.{js,ts}` that imports `localDb` or `usageDb` indirectly touches SQLite

### 1.4 API route coupling

24 route directories. Already split into two cohorts:

**Cohort A — Almost certainly proxy-convertible** (low risk):
`auth/`, `health/`, `version/`, `combos/`, `keys/`, `providers/`, `provider-nodes/`, `proxy-pools/`, `settings/`, `pricing/`, `models/`, `media-providers/`, `tags/`, `usage/`, `oauth/`

**Cohort B — Requires audit before touching** (medium-high risk):
`v1/`, `v1beta/`, `translator/`, `init/`, `shutdown/`, `cloud/`, `cli-tools/`, `tunnel/`, `locale/`

The cohort B routes are the ones that historically host business logic or runtime-only behaviors.

### 1.5 Streaming coupling

- `frontend/src/sse/{handlers,services,utils}` — frontend SSE consumer
- `frontend/open-sse/{config,executors,handlers,rtk,services,transformer,translator,utils}` — full 9router streaming engine in frontend
- Backend: `backend/internal/stream/` owns normalization

**This is the largest violation in the project.** `open-sse/` duplicates what backend already does.

### 1.6 State management risks

- Stores in `src/store/` are healthy targets (Zustand, no hidden persistence apparent)
- Risk: any store calling `localDb`/`usageDb` for hydration would break on backend-driven mode
- `userStore` likely holds token — coordinate with `lib/api.ts` token storage to avoid two sources of truth

### 1.7 High-risk modules (ranked)

| Rank | Module | Reason |
|---|---|---|
| 1 | `frontend/open-sse/` | Full executor engine in frontend; parity-sensitive |
| 2 | `frontend/src/lib/db/` | SQLite root |
| 3 | `frontend/src/app/api/v1/`, `v1beta/` | Provider-format endpoints — must NOT contain logic |
| 4 | `frontend/src/lib/oauth/` | Auth flow leakage |
| 5 | Dual app shell (`page.js` + `page.tsx`) | Non-deterministic |
| 6 | `frontend/src/app/api/translator/` | Duplicates backend translator |
| 7 | `frontend/src/lib/usage/`, `usageDb.js` shim | Persistence leakage |

---

## 2. Dependency Graph Analysis

### 2.1 Master dependency chains

```
SQLITE ROOT CHAIN
  src/lib/db/index.js (SQLite driver)
    ←── src/lib/usageDb.js (shim)
    ←── src/lib/localDb.js (shim)
    ←── src/lib/disabledModelsDb.js (likely shim)
    ←── src/lib/requestDetailsDb.js (likely shim)
    ←── src/lib/usage/* (uses usageDb)
    ←── src/app/api/*/route.{js,ts} (any that imports the shims)
    ←── Components/pages calling those APIs

FILESYSTEM CHAIN
  src/lib/dataDir.js (DATA_DIR, APP_NAME="9router")
    ←── src/lib/db/* (DB file location)
    ←── any FS-using utility

API ROUTE CHAIN
  src/app/api/*/route.{js,ts}
    ←── src/lib/{localDb,usageDb,oauth,tunnel,usage}/*
    ←── src/shared/services/*
    ←── Direct DB or FS access

SHARED CONSTANTS COUPLING
  src/shared/constants/* (path inconsistencies per tracker)
    ←── src/lib/providerNormalization.js
    ←── src/app/api/providers/*
    ←── src/app/api/v1/*
    ←── open-sse/config/*

STREAMING CHAIN
  open-sse/{executors,translator,rtk,transformer}
    ←── src/sse/{handlers,services}
    ←── src/app/api/v1/chat/* (likely)
    ←── frontend stream renderer components

AUTH CHAIN
  src/lib/api.ts (login, getToken/setToken via localStorage)
    ←── src/store/userStore.js (likely duplicate token source)
    ←── src/lib/admin-api.ts (uses getToken)
    ←── src/app/login/*, src/app/callback/*
    ←── src/lib/oauth/* (legacy, must be neutralized)
```

### 2.2 Migration order (forced by graph)

1. **Backend contract lock** (port, base URL, env, auth header) — unblocks every API client
2. **Shared constants** — referenced by everything; must be stabilized before route migration
3. **`lib/api.ts` + token strategy unified** — every other client depends on it
4. **Thin clients (`admin-api`, `analytics-api`, `oauth-api`)** — replace API route logic
5. **API routes converted to proxies, cohort A first** — unblocks SQLite removal
6. **API routes converted to proxies, cohort B (audit-heavy)** — including `v1`, `v1beta`, `translator`
7. **`open-sse/` retirement** — backend stream normalization takes over
8. **`src/sse/` slimmed down to UI rendering only**
9. **`lib/oauth/`, `lib/tunnel/`, `lib/updater/`, `lib/usage/`, `lib/mitm`, `lib/network/`** — removed after callers migrated
10. **Shim files (`usageDb.js`, `localDb.js`, `disabledModelsDb.js`, `requestDetailsDb.js`)** — removed after zero imports
11. **`lib/db/` SQLite engine** — removed last, after all shims dead
12. **`lib/dataDir.js`** — removed (or stubbed to no-op) once no FS users remain
13. **Dual app shell collapsed** — `.js` siblings deleted after `.tsx` verified

### 2.3 Blockers and isolations

**Blockers (must migrate first, block everything downstream):**
- Backend base URL config (currently inconsistent: `20128` in code, `14322` in STRUCTURE.md)
- `src/shared/constants/*` path inconsistencies
- `src/lib/api.ts` token/auth shape

**Isolable (can migrate independently):**
- `src/store/themeStore.js`, `headerSearchStore.js`, `notificationStore.js` (pure UI)
- Most pages under `src/app/(dashboard)/*` once their API client is ready
- `src/i18n/*` (locale)

**Blocking-others (downstream is large):**
- `src/lib/db/` blocks ALL persistence removal
- `open-sse/` blocks streaming refactor
- `src/lib/oauth/` blocks auth route migration

---

## 3. Frontend Layer Classification

Every module in `frontend/src/` and `frontend/open-sse/` falls into exactly one of these seven categories. **Migration approach is determined by category, not by file.**

### Category A — Pure UI
**Definition:** React components, hooks, styles, layouts. No fetch logic, no business logic, no persistence.
**Examples (confirmed/likely):** `src/app/(dashboard)/*` components, `src/shared/components/*`, `src/store/themeStore.js`, `src/store/headerSearchStore.js`, `src/store/notificationStore.js`
**Approach:** Leave untouched. Only modify when a downstream API client changes shape.
**Risk:** 🟢 Low

### Category B — Thin Proxy Routes
**Definition:** `src/app/api/*/route.{ts,js}` that ONLY does `fetch(BACKEND_URL + ...)` with auth header pass-through.
**Examples (target state for cohort A):** `auth/`, `health/`, `version/`, `keys/`, `combos/`, `providers/`, `provider-nodes/`, `settings/`, `pricing/`, `models/`, `usage/`, `tags/`
**Approach:** Standardize all to a single proxy helper. Strip every other concern.
**Risk:** 🟡 Medium (each conversion needs validation)

### Category C — Business Logic Leakage
**Definition:** Frontend code performing logic that belongs to backend (analytics, routing, model resolution, pricing).
**Examples:** `src/lib/usage/*` (aggregation), `src/lib/providerNormalization.js`, `src/lib/initCloudSync.js`, `src/app/api/init/*`, `src/app/api/translator/*`
**Approach:** Identify caller → switch caller to backend API → delete module.
**Risk:** 🟠 High (parity-sensitive)

### Category D — Persistence Leakage
**Definition:** Anything that reads/writes SQLite, filesystem, or hidden persistence.
**Examples:** `src/lib/db/*`, `src/lib/usageDb.js`, `src/lib/localDb.js`, `src/lib/disabledModelsDb.js`, `src/lib/requestDetailsDb.js`, `src/lib/dataDir.js`, `src/lib/mitmAliasCache.js`
**Approach:** SQLite Removal Strategy (§8). Trace → replace → delete.
**Risk:** 🔴 Critical

### Category E — Streaming Infrastructure
**Definition:** Stream parsing, executor chains, translator, transformer, RTK filters.
**Examples:** `frontend/open-sse/{config,executors,handlers,rtk,services,transformer,translator,utils}`, parts of `src/sse/handlers/`, `src/sse/services/`
**Approach:** Streaming Refactor Strategy (§6). Backend takes ownership; frontend renders only.
**Risk:** 🔴 Critical (highest regression surface)

### Category F — Shared Contracts
**Definition:** Types, constants, schemas shared between client and routes.
**Examples:** `src/shared/constants/*`, `src/shared/utils/*`, `src/models/*`
**Approach:** Stabilize early (Phase 1). Make them mirror backend `API_CONTRACT.md`. Do not invent.
**Risk:** 🟡 Medium (high blast radius if changed wrong)

### Category G — Legacy Hybrid Systems
**Definition:** Modules that were full-stack in 9router and still partially live in frontend.
**Examples:** `src/lib/oauth/*`, `src/lib/tunnel/*`, `src/lib/updater/*`, `src/lib/appUpdater.js`, `src/mitm/*`, `src/lib/network/*`, frontend root scripts (`serve-static.js`, `login-test.js`, `test-*.js`)
**Approach:** Per-module decision — backend-replace, deprecate, or relocate (test scripts → `frontend/scripts/`).
**Risk:** 🟠 High

### Dependency order across categories

```
F (Contracts)  →  B (Proxy routes)  →  C (Logic leakage)  →  D (Persistence)
                                    ↘  G (Legacy hybrid)
                  E (Streaming) is largely orthogonal; sequenced separately (§6)
A (Pure UI) follows whichever client it depends on.
```

---

## 4. Refactor Execution Phases

Each phase has a single objective and a single rollback boundary. **Do not interleave phases.** Phases 0–2 are unblockers; 3–9 are the migration body; 10–11 are stabilization.

---

### Phase 0 — Inventory & Lock-Down (no code change)
**Objective:** Produce an authoritative dependency map and lock the backend contract.

**Target files/folders:** Entire `frontend/`, `backend/internal/api/`.

**Migration strategy:**
1. Run `rg -n "from ['\"]@/lib/(usageDb|localDb|disabledModelsDb|requestDetailsDb|db/)" frontend/src` and save output as `IMPORT_GRAPH.md`.
2. Run `rg -n "from ['\"]@/lib/(oauth|tunnel|updater|usage|network|mitm)" frontend/src`.
3. Run `rg -n "open-sse" frontend/src` to map all consumers of the streaming engine.
4. For each `src/app/api/*/route.{ts,js}`, classify as **Cohort A** (already thin) or **Cohort B** (contains logic).
5. Verify backend base URL: pick ONE of `20128` or `14322`, update `API_CONTRACT.md` and `.env.example`. Fix `STRUCTURE.md` inconsistency.
6. Confirm Cohort A endpoints exist on Go backend (use `curl` against `backend/internal/api/`).

**Parity risks:** None (read-only).
**Validation:** Hand-review the produced inventory with backend owner.
**Rollback risk:** None.
**Dependencies:** None.
**Expected build break risks:** None.

---

### Phase 1 — Shared Contracts Stabilization
**Objective:** Make `src/shared/constants/*`, `src/shared/utils/*`, and `src/models/*` consistent with backend.

**Target:** `frontend/src/shared/constants/`, `frontend/src/shared/utils/`, `frontend/src/models/`.

**Migration strategy:**
1. List every import path inside `shared/constants/`. Normalize to a single index pattern.
2. For any constant that mirrors a backend value (provider list, error codes, model alias map): mark it `// SOURCE: backend (do not edit locally)` and replace hardcoded data with TypeScript types only, where possible.
3. Move any constant that is genuinely shared with backend into a single file: `src/shared/constants/api.ts` re-exporting from `API_CONTRACT.md` shapes.
4. Do NOT introduce new abstractions. Keep names identical to current ones to preserve component contracts.

**Parity risks:** 🟡 If a component imports a constant whose path silently changed.
**Validation:** `tsc --noEmit` clean; smoke-test dashboard renders.
**Rollback risk:** Low — git revert is sufficient.
**Dependencies:** Phase 0.
**Build break risks:** Path inconsistencies (the tracker already flags this). Batch all path fixes in one commit.

---

### Phase 2 — Auth + Token Source Unification
**Objective:** Single token source of truth via `src/lib/api.ts`; preserve 9router cookie/redirect UX.

**Target:** `src/lib/api.ts`, `src/store/userStore.js`, `src/app/login/*`, `src/app/callback/*`, any component using token.

**Migration strategy:**
1. Decide token storage: **keep `localStorage` for now** (matches current `api.ts`). Cookie migration is a separate, later decision approved by PM.
2. `userStore` must NOT cache the token independently. It may cache user profile only.
3. All authenticated clients (`admin-api.ts`, `analytics-api.ts`, `oauth-api.ts`) MUST go through a single `authFetch` helper. `admin-api.ts` already has one — promote it to `src/lib/http.ts` and import from the others.
4. Login flow: form submits → `lib/api.ts:login()` → store token → redirect. Preserve redirect target behavior verbatim.
5. Preserve password-only login UX (per `FEATURE_PARITY.md`).

**Parity risks:** 🟠 Login redirect, session restore on reload, logout flow.
**Validation:** Manual login/logout regression test. Compare redirect chain against 9router screenshots if available.
**Rollback risk:** Medium — auth touches every protected page.
**Dependencies:** Phase 1.
**Build break risks:** Type changes to `userStore` may ripple into pages.

---

### Phase 3 — API Route Proxy Conversion: Cohort A
**Objective:** Convert every clearly-proxyable route into a thin `fetch(BACKEND_URL + path, ...)` passthrough.

**Target routes (in order):**
1. `auth/` (low risk — already mostly proxy)
2. `health/`, `version/` (trivial)
3. `keys/`, `providers/`, `provider-nodes/`, `proxy-pools/` (admin CRUD — backend has `/api/admin/*`)
4. `combos/`, `pricing/`, `models/`, `media-providers/`, `tags/`
5. `settings/`, `usage/`, `oauth/`

**Migration strategy (per route):**
1. Read the existing `route.{ts,js}`. If it ONLY contains `fetch(BACKEND_URL...)`, mark done.
2. If it contains DB calls or logic, replace its body with the standard proxy template (§5.2).
3. Forward `Authorization`, `Content-Type`, request body, and method. Forward response body and status verbatim. Stream response if backend streams.
4. Delete any local import of `usageDb`/`localDb`/`db/` from the route file.
5. Do NOT delete the imported modules yet — only remove the route's dependency on them.

**Parity risks:** Response shape drift between old route and backend. Pin response format to `API_CONTRACT.md`.
**Validation:** For each converted route, diff response with previous behavior using saved HAR/Postman fixtures.
**Rollback risk:** Per-route — small.
**Dependencies:** Phase 2 (auth header forwarding) + Phase 1 (constants).
**Build break risks:** Type errors where consumers expected old shapes. Batch by sub-feature.

---

### Phase 4 — API Route Proxy Conversion: Cohort B (Audit-Heavy)
**Objective:** Convert routes that historically host logic or runtime behaviors.

**Target routes (in order, hardest last):**
1. `init/`, `shutdown/` — runtime lifecycle. Decide: delete from frontend entirely (backend manages lifecycle), or keep as no-op proxies.
2. `cli-tools/`, `cloud/`, `tunnel/`, `locale/` — feature-flag-dependent. Confirm with PM whether to keep, proxy, or remove.
3. `translator/` — backend already has translator. Either proxy to backend translator endpoint or remove entirely.
4. `v1beta/` — likely Gemini-compatible. Audit, proxy to `/v1beta/*` on backend.
5. `v1/` — OpenAI-compatible. Highest risk. Audit, proxy to `/v1/*` on backend. **Streaming routes here interact with Phase 7.**

**Migration strategy (per route):**
1. Snapshot current request/response (capture HAR).
2. Confirm backend equivalence exists and matches schema.
3. Replace with proxy. For streaming endpoints, pipe `ReadableStream` from backend to client without buffering.
4. Validate streaming chunk ordering and `done` signaling unchanged.

**Parity risks:** 🔴 v1/v1beta route behavior is end-user-visible API parity.
**Validation:** Replay HAR fixtures pre/post. Verify byte-for-byte streaming output where possible.
**Rollback risk:** Medium-high.
**Dependencies:** Phases 0–3.
**Build break risks:** Type drift for `v1` response interfaces.

---

### Phase 5 — Legacy Module Deprecation (G category, non-streaming)
**Objective:** Remove callers of `src/lib/{oauth,tunnel,updater,usage,network,mitm}/` from the frontend.

**Target:** Listed dirs. Caller list comes from Phase 0 inventory.

**Migration strategy:**
1. For each module: list every caller (Phase 0 output).
2. Switch each caller to the corresponding backend-driven client (`oauth-api.ts`, `admin-api.ts`, etc.).
3. Verify zero remaining imports with `rg`.
4. Delete the legacy module directory.
5. Run build after each module group.

**Parity risks:** 🟠 OAuth flow, MITM aliasing, tunnel behavior all have UX implications.
**Validation:** Manual OAuth login per provider; verify MITM aliasing still affects request routing identically.
**Rollback risk:** Medium.
**Dependencies:** Phases 2–4.
**Build break risks:** Likely many — batch deletions by module.

---

### Phase 6 — Pre-SQLite-Removal: Eliminate Shim Callers
**Objective:** Reach zero imports of `usageDb`, `localDb`, `disabledModelsDb`, `requestDetailsDb`.

**Target:** Every caller of those shims.

**Migration strategy:**
1. From Phase 0 inventory, list every caller.
2. Each caller must already be migrated to backend APIs by Phases 3–5. If any remain, that's a missed Phase-4 task — go back and fix.
3. Verify with `rg "from .@/lib/(usageDb|localDb|disabledModelsDb|requestDetailsDb)" frontend/src` returning zero matches.
4. **Do not delete the shim files yet.**

**Parity risks:** None at this step — it's a verification gate.
**Validation:** `rg` empty + build green.
**Rollback risk:** None.
**Dependencies:** Phases 3–5.
**Build break risks:** None.

---

### Phase 7 — Streaming Refactor
**Objective:** Retire `frontend/open-sse/`. Backend owns stream normalization. Frontend renders only.

**Target:** `frontend/open-sse/*`, `frontend/src/sse/handlers/`, `frontend/src/sse/services/`, any consumer of these.

See **§6 Streaming Refactor Strategy** for the detailed sub-plan. Summary:

1. Confirm backend `/api/chat/stream` (and `/v1/chat/completions` streaming) emits normalized chunks.
2. Replace frontend stream parsing with a single SSE consumer that emits rendering events.
3. Keep `src/sse/utils/` and `src/sse/handlers/` slimmed to **UI-side** chunk handling (deltas, abort UI, retry UI).
4. Delete `frontend/open-sse/` entirely.
5. Preserve: chunk render order, abort behavior, retry visibility, token streaming pacing.

**Parity risks:** 🔴 Highest in the project.
**Validation:** Side-by-side stream playback against 9router reference. Test cases: short response, long response, abort mid-stream, network-flap retry, provider error mid-stream.
**Rollback risk:** 🔴 High.
**Dependencies:** Phase 4 (v1 routes proxied first), Phase 0 (inventory).
**Build break risks:** Many imports from `open-sse/*` will break. Batch in one commit.

---

### Phase 8 — Drop SQLite Engine
**Objective:** Delete `src/lib/db/` and all shims. Frontend has zero persistence.

**Target:** `src/lib/db/*`, `src/lib/usageDb.js`, `src/lib/localDb.js`, `src/lib/disabledModelsDb.js`, `src/lib/requestDetailsDb.js`, `src/lib/dataDir.js`, `src/lib/mitmAliasCache.js`.

See **§8 SQLite Removal Strategy** for the detailed sub-plan. Summary:

1. Confirm Phase 6 gate is still green (zero shim imports).
2. Delete shim files first.
3. Delete `src/lib/db/` directory.
4. Delete `src/lib/dataDir.js` (or stub to throw).
5. Remove `better-sqlite3` (or equivalent) from `package.json` and lockfile.
6. Verify build green and runtime smoke pass.

**Parity risks:** 🟠 If any caller was missed, runtime crash.
**Validation:** Full smoke + integration test pass.
**Rollback risk:** Medium (revert is just `git revert` but caller bugs are runtime).
**Dependencies:** Phases 6–7.
**Build break risks:** Should be zero if Phase 6 was clean.

---

### Phase 9 — Dual App Shell Collapse
**Objective:** One root layout, one root page. Either fully TS or fully JS — choose TS.

**Target:** `src/app/page.js`, `src/app/layout.js`, `src/app/page.tsx`, `src/app/layout.tsx`.

**Migration strategy:**
1. Diff `page.js` vs `page.tsx`. Ensure `.tsx` is functionally equivalent.
2. Same for `layout.js` vs `layout.tsx`.
3. If `.tsx` is missing any behavior present in `.js`, port it.
4. Delete the `.js` siblings in a single commit.
5. Audit `frontend/jsconfig.json` and `frontend/tsconfig.json` for path aliases — keep both until all `.js` files are gone.

**Parity risks:** 🟡 The "wrong" file may currently be active.
**Validation:** Visual diff homepage, root layout, navigation.
**Rollback risk:** Low (small surface).
**Dependencies:** Phases 0–8 ideally; can run earlier if low risk validated.
**Build break risks:** Low.

---

### Phase 10 — Dead Code & Dependency Cleanup
**Objective:** Remove orphaned imports, unused dependencies, ad-hoc scripts.

**Target:**
- `frontend/login-test.js`, `frontend/test-api-me.js`, `frontend/test-dashboard.js`, `frontend/test-detailed.js`, `frontend/serve-static.js` → move to `frontend/scripts/` or delete
- Unused npm packages
- Empty directories

**Migration strategy:** Run `depcheck`, `ts-prune`, manual review. Move test scripts to a dedicated folder. Remove orphan packages from `package.json`.

**Parity risks:** None if scripts are unused.
**Validation:** Build green, full smoke.
**Rollback risk:** Low.
**Dependencies:** All prior phases.
**Build break risks:** Low.

---

### Phase 11 — Final Stabilization & Documentation
**Objective:** Lock the new architecture in place.

**Tasks:**
1. Update `STRUCTURE.md`, `ARCHITECTURE.md`, `FRONTEND_REFACTOR_TRACKER.md` to reflect final state.
2. Add a `frontend/ARCHITECTURE.md` describing the thin-client architecture and rules.
3. Add a lint rule (or pre-commit hook) that blocks imports of `src/lib/db/`, `src/lib/usageDb`, `src/lib/localDb`, `open-sse/` (in case any reappear via merge).
4. Add an ESLint plugin or simple `rg` CI step that fails the build if `better-sqlite3`, `sqlite3`, or `fs` (from non-Next-server contexts) is imported in `frontend/src`.
5. Final regression sweep against `FEATURE_PARITY.md`.

**Dependencies:** All.
**Risk:** None.

---

## 5. API Route Migration Strategy

### 5.1 Inventory (current `src/app/api/`)

24 subroutes confirmed: `auth/`, `cli-tools/`, `cloud/`, `combos/`, `health/`, `init/`, `keys/`, `locale/`, `media-providers/`, `models/`, `oauth/`, `pricing/`, `provider-nodes/`, `providers/`, `proxy-pools/`, `settings/`, `shutdown/`, `tags/`, `translator/`, `tunnel/`, `usage/`, `v1/`, `v1beta/`, `version/`.

### 5.2 Standard proxy pattern (the ONLY allowed body)

```ts
// frontend/src/app/api/<feature>/route.ts
import { NextRequest } from 'next/server';
import { proxyToBackend } from '@/lib/proxy';

export async function GET(req: NextRequest)    { return proxyToBackend(req); }
export async function POST(req: NextRequest)   { return proxyToBackend(req); }
export async function PATCH(req: NextRequest)  { return proxyToBackend(req); }
export async function DELETE(req: NextRequest) { return proxyToBackend(req); }
```

Where `proxyToBackend` (single helper in `src/lib/proxy.ts`):
- Reads `process.env.BACKEND_INTERNAL_URL` (server-side, NOT `NEXT_PUBLIC_API_URL`)
- Forwards method, path (after `/api/`), query, body, `Authorization`, `Cookie` headers
- Streams response body through if `Content-Type` is `text/event-stream` or `application/octet-stream`
- Forwards backend status code verbatim
- Never parses business payloads

### 5.3 Classification

| Route | Cohort | Action | Phase |
|---|---|---|---|
| `auth/` | A | Proxy | 3 |
| `health/` | A | Proxy (or delete and use backend `/health` directly) | 3 |
| `version/` | A | Proxy | 3 |
| `keys/` | A | Proxy to `/api/admin/keys` | 3 |
| `providers/` | A | Proxy to `/api/admin/providers` | 3 |
| `provider-nodes/` | A | Proxy to `/api/admin/nodes` | 3 |
| `proxy-pools/` | A | Proxy | 3 |
| `combos/` | A | Proxy to `/api/admin/combos` | 3 |
| `pricing/` | A | Proxy to `/api/admin/pricing` | 3 |
| `models/` | A | Proxy | 3 |
| `media-providers/` | A | Proxy | 3 |
| `tags/` | A | Proxy | 3 |
| `settings/` | A | Proxy | 3 |
| `usage/` | A | Proxy to `/api/admin/usage` | 3 |
| `oauth/` | A | Proxy to `/api/oauth/*` (backend owns flow) | 3 |
| `init/` | B | **Decide: delete or no-op proxy** | 4 |
| `shutdown/` | B | **Likely delete** (backend lifecycle) | 4 |
| `cli-tools/` | B | Audit → proxy or remove | 4 |
| `cloud/` | B | Audit → proxy or remove | 4 |
| `tunnel/` | B | Audit → proxy or remove | 4 |
| `locale/` | B | Likely keep as Next-side (UI locale) | 4 |
| `translator/` | B | Delete (backend has translator) | 4 |
| `v1beta/` | B | Proxy with streaming passthrough | 4 |
| `v1/` | B | Proxy with streaming passthrough | 4 + 7 |

### 5.4 Auth/session forwarding rule

- Client → Next route: `Authorization: Bearer <token>` (from `localStorage`)
- Next route → Backend: forward same header verbatim
- Cookies (if any used by 9router parity): forward `Cookie` header verbatim
- Next route NEVER inspects, decodes, or validates the token. That is backend's job.

### 5.5 Backend parity validation

For each migrated route:
1. Capture pre-migration HAR.
2. Replay against post-migration route.
3. Diff: status code, headers (allowlist: `content-type`, `content-length`, `set-cookie`), body.
4. Sign off in tracker.

---

## 6. Streaming Refactor Strategy

### 6.1 Current state

- **Frontend (violates rules):**
  - `frontend/open-sse/{config,executors,handlers,rtk,services,transformer,translator,utils}` — full 9router streaming engine.
  - `frontend/src/sse/{handlers,services,utils}` — additional frontend stream layer.
- **Backend (correct owner):**
  - `backend/internal/stream/` — already normalizes provider streams (per `FEATURE_PARITY.md`).

### 6.2 Boundary definition

| Responsibility | Frontend | Backend |
|---|---|---|
| SSE connection management | ✅ | ❌ |
| Reading SSE chunks | ✅ | ❌ |
| Parsing provider-native deltas (OpenAI, Claude, Gemini formats) | ❌ | ✅ |
| Normalizing chunks to single AIPROXY chunk format | ❌ | ✅ |
| Retry on provider error | ❌ | ✅ |
| Fallback to next account | ❌ | ✅ |
| Rendering tokens incrementally | ✅ | ❌ |
| Abort signal forwarding | ✅ (client-side) + ✅ (backend stops upstream) | ✅ |
| Showing "retrying..." UI state | ✅ | ❌ (just sends event) |

### 6.3 Migration sub-phases (inside Phase 7)

**7a — Verify backend stream contract.** Inspect `backend/internal/stream/` and confirm the normalized chunk format. Document in `API_CONTRACT.md` under streaming section. Required event types (minimum): `delta`, `done`, `error`, `retry` (optional), `usage` (optional). Each event must have stable JSON shape.

**7b — Build a single frontend SSE consumer.** New file `src/sse/consumer.ts`. Owns `EventSource` (or `fetch` + `ReadableStream` for POST-SSE), parses the normalized chunk format only, exposes a typed event emitter (`onDelta`, `onDone`, `onError`, `onRetry`).

**7c — Migrate UI components.** Chat UI / stream renderers switch from importing `open-sse/*` to importing `src/sse/consumer.ts`.

**7d — Verify parity.** Test matrix:
| Case | Expected behavior |
|---|---|
| Short response (1 chunk) | Renders once, `done` fires |
| Long response (100+ chunks) | Renders progressively, order stable |
| User abort mid-stream | UI cancels, backend receives abort, no leaked tokens |
| Network blip | Retry visible if backend re-emits, else error UI |
| Provider 5xx mid-stream | Backend retries with fallback, frontend stays connected |
| Provider auth failure | Backend emits `error` event with message, frontend renders inline error |

**7e — Delete `frontend/open-sse/`.** Single commit. Verify `rg "open-sse"` returns zero matches in `frontend/src`.

**7f — Slim `frontend/src/sse/`.** Keep only UI-facing handlers. Anything that parses provider deltas (OpenAI/Claude/Gemini specific) gets deleted.

### 6.4 Parity-sensitive behaviors (do NOT change)

- **Chunk ordering:** Backend emits in order; frontend MUST render in order. No reordering, no batching that changes visible order.
- **Retry visibility:** If 9router showed "retrying with next account", AIPROXY must too. Backend emits `retry` event; frontend renders the same indicator.
- **Abort behavior:** Click "stop" → immediate UI stop → backend close connection → upstream provider request canceled. Latency to "stop" must match 9router.
- **Token pacing:** Do NOT throttle chunks in frontend. Render every chunk as it arrives.

### 6.5 Dangerous rewrite risks

| Risk | Mitigation |
|---|---|
| Chunk reordering bug | Strict in-order rendering; no `Promise.all` over chunks |
| Double-render on retry | Backend signals `retry` BEFORE re-emitting; frontend resets buffer on `retry` |
| Memory leak on abort | Consumer must close `ReadableStream` reader in `finally` |
| Lost final chunk | Always wait for explicit `done` event; do not infer completion from connection close |
| Provider format leak | Backend MUST normalize. If a provider-native chunk reaches the frontend, that is a backend bug — file ticket, do NOT patch in frontend. |

---

## 7. Authentication + OAuth Migration Safety

### 7.1 Current auth flow (observed)

- `src/lib/api.ts:login(username, password)` → POST `/api/login` → returns `{ token, user? }`
- Token stored in `localStorage.token`
- `src/lib/api.ts:getMe(token)` → GET `/api/me` with Bearer header
- `src/lib/admin-api.ts:authFetch` reads token via `getToken()` and forwards as Bearer
- `src/lib/oauth-api.ts` — separate OAuth client (likely backend-driven already)
- `src/lib/oauth/` — legacy directory, likely contains OAuth flow logic that must move to backend

### 7.2 What MUST remain frontend

- Login form (password-only, per `FEATURE_PARITY.md`)
- Token storage (`localStorage` for now — keeping 9router-like behavior)
- Redirect after successful login (preserve target route handling exactly)
- OAuth callback page (`src/app/callback/*`) — must only forward the callback code to backend
- "Login expired → redirect to login" UI logic
- Logout: clear `localStorage`, redirect

### 7.3 What MUST move (or already lives in) backend

- Password validation
- JWT signing/verification
- Session persistence (DB-backed in Postgres)
- OAuth provider flow (PKCE, code exchange, token refresh)
- OAuth pool management (multi-account)
- Token expiry rules

### 7.4 What must NOT be rewritten

- The login UX shape (form fields, button, error placement)
- The redirect target logic (where the user goes after login)
- The "remember me" / persistence behavior (if 9router persists across refresh, AIPROXY must too)
- The OAuth callback URL contract (path, query params)

### 7.5 OAuth migration sub-plan

1. Audit `src/lib/oauth/` and `src/app/api/oauth/` for any token-exchange logic.
2. Confirm backend handles: `/api/oauth/start`, `/api/oauth/callback`, `/api/oauth/refresh`, `/api/oauth/providers`.
3. Convert frontend OAuth flow to:
   - **Frontend:** redirect user to `/api/oauth/start?provider=X` (proxied to backend)
   - **Backend:** owns PKCE, redirect URL building, code exchange, token storage
   - **Frontend callback page:** receives backend success/error, redirects to dashboard
4. Delete `src/lib/oauth/` only after every page uses `oauth-api.ts`.

### 7.6 Cookie vs Bearer

Current frontend uses Bearer + `localStorage`. 9router uses cookies. **Do not silently switch.** This is a decision that requires explicit PM approval because:
- Cookies are more secure (HttpOnly, SameSite)
- But CORS/credentials handling differs
- And reload behavior differs

For this migration: **keep Bearer/localStorage**, log the decision in `FRONTEND_REFACTOR_TRACKER.md`, revisit later if security requires.

---

## 8. SQLite Removal Strategy

**Hard rule:** Trace → classify → replace → verify → delete. Never delete first.

### 8.1 Removal targets (in order)

1. `src/lib/usageDb.js` (shim)
2. `src/lib/localDb.js` (shim)
3. `src/lib/disabledModelsDb.js` (shim, 177B)
4. `src/lib/requestDetailsDb.js` (shim, 166B)
5. `src/lib/mitmAliasCache.js` (may shim into db)
6. `src/lib/db/` (the actual SQLite engine)
7. `src/lib/dataDir.js` (filesystem coupling)
8. `better-sqlite3` (or driver in use) from `package.json`

### 8.2 Import tracing workflow

For each target above, in order:

```bash
# Find every importer in source tree
rg -n "from ['\"]@/lib/usageDb" frontend/src
rg -n "from ['\"]@/lib/localDb" frontend/src
rg -n "from ['\"]@/lib/db/" frontend/src
rg -n "from ['\"]@/lib/dataDir" frontend/src
rg -n "better-sqlite3" frontend/
```

### 8.3 Per-importer classification

Each importer falls into one of:

| Classification | Replacement |
|---|---|
| API route (`src/app/api/*/route.{ts,js}`) | Replace with proxy (Phase 3/4) |
| Lib utility (`src/lib/usage/*` etc.) | Replace caller chain; module eventually deletes |
| Store (`src/store/*`) | Replace with backend-driven fetch via TanStack Query (allowed) |
| Component (`src/app/(dashboard)/*`) | Switch to corresponding `lib/*-api.ts` client |
| Test/script | Delete or move to `frontend/scripts/` |

### 8.4 Replacement workflow (per importer)

1. Identify the backend endpoint that owns the equivalent data (consult `API_CONTRACT.md` + backend handlers).
2. If endpoint missing: STOP. File a ticket on backend. Do NOT invent endpoints.
3. Add the call to the appropriate thin client (`admin-api.ts`, `analytics-api.ts`, `usage-api.ts` if needed).
4. Replace the import in the caller. Adjust types to backend's response shape.
5. Run `tsc --noEmit` for that file.
6. Verify the UI still renders identical data.

### 8.5 Validation workflow

After all importers migrated:
1. `rg -n "from ['\"]@/lib/(usageDb|localDb|disabledModelsDb|requestDetailsDb|db/|dataDir)" frontend/src` → must return zero.
2. Build clean.
3. Smoke pass on: dashboard load, usage charts, settings save, provider toggle, model alias edit, key creation/deletion.
4. Run integration test suite if available.

### 8.6 Deletion (final step)

Single commit:
- Delete `src/lib/usageDb.js`, `localDb.js`, `disabledModelsDb.js`, `requestDetailsDb.js`, `mitmAliasCache.js`
- Delete `src/lib/db/` recursively
- Delete `src/lib/dataDir.js`
- Remove `better-sqlite3` from `package.json` + `package-lock.json`
- Verify `npm install` clean
- Verify `npm run build` clean

### 8.7 Forbidden shortcuts

- ❌ Delete files first then patch errors as they appear
- ❌ Leave `usageDb.js` as a no-op stub "in case"
- ❌ Recreate any SQLite module under a different name
- ❌ Add `localStorage`-based "fallback" persistence to replace SQLite
- ❌ Move SQLite to a different layer (e.g., a worker)

---

## 9. Build Stabilization Strategy

### 9.1 Anti-chaos rules

1. **Never run `npm run build` to "see what breaks."** Build only after a coherent batch of related changes.
2. **One phase, one rollback boundary.** Do not interleave phases across commits.
3. **Never patch a build error without tracing its root cause.** A missing import after deletion is fixed by **adding the import back temporarily**, not by stubbing the call site.
4. **Use `tsc --noEmit` for type checks**; reserve `next build` for full validation.

### 9.2 Safe workflow (per phase)

```
1. Pull main, branch off
2. Make the planned batch of changes (whole phase)
3. Run `tsc --noEmit` — fix type errors only, no behavior changes
4. Run `npm run lint`
5. Run `npm run build`
6. Read FIRST build error only; do not scroll
7. Fix root cause; do not patch symptoms
8. Re-run build
9. When green: run smoke tests
10. Commit, push, PR with phase number in title
```

### 9.3 Anti-regression workflow

For each phase:
1. Define explicit acceptance criteria upfront (in PR description).
2. Capture HAR fixtures of relevant endpoints BEFORE.
3. Capture HAR fixtures AFTER.
4. Diff. Any unexpected diff → block PR.

### 9.4 Detecting root-cause errors

When build fails after a change:
1. Read the **first** error in the output. Subsequent errors are usually cascades.
2. Identify whether it is: (a) missing import, (b) type mismatch, (c) module-not-found, (d) circular dep.
3. For each:
   - (a) Missing import → was the deleted module its only source? If yes, you skipped a replace step. Go back.
   - (b) Type mismatch → response shape drift. Pin to `API_CONTRACT.md`.
   - (c) Module not found → path inconsistency, likely in `shared/constants/`. Fix in Phase 1 if not done.
   - (d) Circular dep → restructure imports, do not paper over.

### 9.5 Batch sizes

- API route conversions: batch 3–5 routes per commit (same cohort)
- Store migrations: 1 store per commit
- Legacy module deletions: 1 module per commit
- SQLite removal: 1 single commit per Phase 8 sub-step (shims → engine → driver)
- Streaming refactor: large commits are unavoidable; gate behind feature flag if possible

### 9.6 Forbidden actions

- ❌ Commit with broken build "to keep moving"
- ❌ Disable TypeScript strict in a file to make it compile
- ❌ Use `@ts-ignore` to silence migration errors
- ❌ Use `any` to bypass type drift — that's how parity bugs hide
- ❌ Run `npm install` and commit the lockfile change as part of an unrelated phase

---

## 10. Final Target Architecture

### 10.1 Final frontend structure

```
frontend/
├── src/
│   ├── app/
│   │   ├── (dashboard)/              # Pure UI pages
│   │   ├── api/                      # THIN PROXIES ONLY
│   │   │   ├── auth/route.ts         # proxyToBackend
│   │   │   ├── admin/                # proxy to /api/admin/*
│   │   │   ├── v1/                   # proxy with streaming passthrough
│   │   │   ├── v1beta/               # proxy with streaming passthrough
│   │   │   ├── oauth/                # proxy
│   │   │   └── locale/               # Next-side i18n (allowed)
│   │   ├── callback/                 # OAuth callback UI (no logic)
│   │   ├── login/                    # Login UI
│   │   ├── layout.tsx                # Single root layout
│   │   └── page.tsx                  # Single root page
│   ├── lib/
│   │   ├── api.ts                    # Base client + token helpers
│   │   ├── http.ts                   # authFetch (single helper)
│   │   ├── proxy.ts                  # proxyToBackend (used by routes)
│   │   ├── admin-api.ts              # /api/admin/* client
│   │   ├── analytics-api.ts          # analytics client
│   │   ├── oauth-api.ts              # oauth client
│   │   └── usage-api.ts              # usage client
│   ├── sse/
│   │   ├── consumer.ts               # Single SSE consumer
│   │   └── ui/                       # UI-side render helpers
│   ├── store/                        # Zustand UI state only
│   ├── shared/
│   │   ├── components/               # UI primitives
│   │   ├── constants/                # Mirrors backend contract
│   │   ├── hooks/                    # UI hooks
│   │   └── utils/                    # UI utilities
│   ├── models/                       # TypeScript types only
│   └── i18n/                         # Locale resources
├── public/
├── next.config.ts
├── tsconfig.json
└── package.json                       # NO better-sqlite3, NO sqlite3, NO fs-heavy deps
```

**Removed:**
- `frontend/open-sse/` ← deleted entirely
- `frontend/src/lib/db/` ← deleted entirely
- `frontend/src/lib/{oauth,tunnel,updater,usage,network,mitm}/` ← deleted
- `frontend/src/lib/{usageDb,localDb,disabledModelsDb,requestDetailsDb,dataDir,mitmAliasCache}.js` ← deleted
- `frontend/src/mitm/` ← deleted (logic moves to backend)
- `frontend/jsconfig.json` ← deleted once all `.js` gone (keep `tsconfig.json`)
- `page.js`, `layout.js` ← deleted (kept `.tsx`)

### 10.2 Final API flow

```
Browser
   │ fetch('/api/admin/keys', { Authorization: 'Bearer <token>' })
   ▼
Next.js (Node runtime)
   │ src/app/api/admin/keys/route.ts → proxyToBackend(req)
   │ Forwards: method, path, query, body, Authorization, Cookie
   ▼
Go Backend (http://backend-internal:20128)
   │ /api/admin/keys handler
   │ Validates JWT, authorizes, executes business logic
   ▼
PostgreSQL / Providers
```

For streaming:
```
Browser EventSource / fetch+ReadableStream
   ▼
Next.js (Node) — streaming passthrough, no parsing
   ▼
Go Backend — normalizes provider stream
   ▼
Provider (OpenAI / Claude / Gemini / ...)
```

### 10.3 Final state ownership

| State | Owner |
|---|---|
| Auth token | Frontend (`localStorage`) — short-lived, refreshed by backend |
| User profile | Frontend `userStore` (in-memory) — fetched from `/api/me` |
| UI theme | Frontend `themeStore` (localStorage allowed for UI prefs only) |
| Header search query | Frontend `headerSearchStore` (in-memory) |
| Notifications | Frontend `notificationStore` (in-memory) |
| Provider list | Backend (frontend caches via TanStack Query) |
| API keys | Backend |
| Combos | Backend |
| Aliases | Backend |
| Pricing | Backend |
| Usage stats | Backend |
| OAuth tokens | Backend |
| Provider routing decisions | Backend |
| Stream chunks | Backend produces, Frontend renders |

### 10.4 Final streaming ownership

- **Backend:** stream connection to provider, format parsing, normalization, retry, fallback, account rotation.
- **Frontend:** SSE consumer for normalized format, UI rendering, abort signal, retry-state UI.
- **Forbidden in frontend:** provider-format parsing, retry decisions, account selection.

### 10.5 Final auth ownership

- **Backend:** password hashing, JWT signing/verification, session DB, OAuth flow, token refresh, token revocation.
- **Frontend:** login form, token transit (Bearer header), redirect, logout (local clear + redirect), OAuth callback page (forwards code to backend).
- **Forbidden in frontend:** JWT decoding for authorization decisions, password validation, OAuth code exchange.

### 10.6 Clear separation summary

| Concern | Frontend | Backend |
|---|---|---|
| Rendering | ✅ | ❌ |
| User input | ✅ | ❌ |
| Routing (Next.js pages) | ✅ | ❌ |
| Routing (provider/model selection) | ❌ | ✅ |
| State (transient/UI) | ✅ | ❌ |
| State (persistent) | ❌ | ✅ |
| Streaming parse | ❌ | ✅ |
| Streaming render | ✅ | ❌ |
| Auth UX | ✅ | ❌ |
| Auth authority | ❌ | ✅ |
| Analytics calc | ❌ | ✅ |
| Pricing | ❌ | ✅ |
| Model aliasing | ❌ | ✅ |
| OAuth pool | ❌ | ✅ |
| Provider SDK calls | ❌ | ✅ |

---

## Appendix A — Open Questions Requiring PM Decision

These were surfaced during analysis. **Do not assume defaults.** Confirm before starting Phase 3.

1. **Backend port:** `20128` (in `lib/api.ts`) vs `14322` (in `STRUCTURE.md`). Which is canonical?
2. **Token storage:** keep `localStorage` (current) or migrate to HttpOnly cookies (more secure, matches 9router)?
3. **`init/` and `shutdown/` routes:** delete entirely or keep as no-op proxies?
4. **`cloud/`, `tunnel/`, `cli-tools/` routes:** these are 9router-specific. Are they in AIPROXY product scope?
5. **`locale/` route:** keep Next-side (UI locale) or proxy to backend?
6. **`mitm/` directory:** is MITM still a product feature, or 9router-only legacy?
7. **Test scripts at frontend root:** delete, move to `scripts/`, or convert to vitest/playwright?

---

## Appendix B — Phase Checklist (One-Line Per Phase)

- [ ] Phase 0: Inventory + backend contract lock
- [ ] Phase 1: Shared constants stabilized
- [ ] Phase 2: Auth + token unified
- [ ] Phase 3: Cohort A API routes proxied
- [ ] Phase 4: Cohort B API routes proxied (audit-heavy)
- [ ] Phase 5: Legacy module deprecation (oauth/tunnel/updater/usage/network/mitm)
- [ ] Phase 6: Zero shim imports verified
- [ ] Phase 7: `open-sse/` retired; backend streams normalized
- [ ] Phase 8: SQLite engine + driver removed
- [ ] Phase 9: Dual app shell collapsed (`.js` siblings deleted)
- [ ] Phase 10: Dead code + dependency cleanup
- [ ] Phase 11: Documentation + lint rules locking architecture

---

## Appendix C — Hard Prohibitions (Re-statement from PROJECT_CONSTRAINTS.md)

NEVER, under any circumstance during this migration:
- Recreate `usageDb.js` or `localDb.js` with new internals
- Reintroduce SQLite or any embedded persistence
- Add `localStorage`-based "fallback" persistence for backend-owned data
- Add compatibility shims or temporary adapters
- Move backend logic back to the frontend
- Parse provider-native stream formats in the frontend
- Invent endpoints not in `API_CONTRACT.md`
- Change user-facing behavior without explicit PM approval
- Delete a module before tracing its imports
- Patch build errors without root-cause analysis

---

*End of plan. Total phases: 12 (0–11). Estimated rollback boundary per phase: 1 commit (with the exception of Phase 7, which may span 3 commits behind a feature flag).*
