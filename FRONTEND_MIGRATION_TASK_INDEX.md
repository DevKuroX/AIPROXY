# AIPROXY Frontend Migration — Master Execution Task Index

> **Status:** Execution-ready task breakdown
> **Source plan:** `FRONTEND_REFACTOR_EXECUTION_PLAN.md`
> **Repo:** `DevKuroX/AIPROXY@main`
> **Authority:** `PROJECT_CONSTRAINTS.md`, `ARCHITECTURE.md`, `API_CONTRACT.md`, `FEATURE_PARITY.md`, `FRONTEND_REFACTOR_TRACKER.md`, `STRUCTURE.md`
> **Behavior reference:** 9router. **Architecture target:** AIPROXY thin client.
> **Hard rule:** Behavior parity REQUIRED. Architecture parity NOT required.

---

## Conventions

- **Task ID format:** `T<phase>.<seq>` (e.g., `T3.4` = Phase 3, task 4).
- **Status legend:** `[ ]` not started · `[~]` in progress · `[x]` done · `[!]` blocked · `[-]` skipped (with reason).
- **Risk tags:** 🟢 Low · 🟡 Medium · 🟠 High · 🔴 Critical.
- **Commit policy:** One task = one commit unless explicitly noted "batch with Tn.x".
- **Branch policy:** One branch per phase (`phase/N-<slug>`); PRs squash-merged.
- **Gate:** A phase's exit criteria must pass before the next phase opens.

---

## 1. Master Migration Index

```
Phase 0 — Pre-Migration Inventory & Backend Contract Lock          [read-only]
Phase 1 — Shared Contracts Stabilization                           [unblocker]
Phase 2 — SQLite Dependency Isolation                              [unblocker]
Phase 3 — API Proxy Standardization (Cohort A — 15 routes)         [body]
Phase 4 — Cohort B Route Audit & Proxy (9 routes)                  [body]
Phase 5 — Auth + Token Source Unification                          [body]
Phase 6 — Legacy Module Deprecation                                [body]
Phase 7 — Streaming Normalization (open-sse retirement)            [body]
Phase 8 — SQLite Engine Removal                                    [terminal]
Phase 9 — Dual App Shell Collapse                                  [terminal]
Phase 10 — Dead Code & Dependency Cleanup                          [stabilization]
Phase 11 — Final Stabilization & Architecture Lock                 [stabilization]
```

### Dependency graph between phases

```
P0 ──► P1 ──► P2 ──► P3 ──► P4 ──┐
                  └──► P5 ────────┤
                                  ├──► P6 ──► P7 ──► P8 ──► P9 ──► P10 ──► P11
                                  │
                  P5 is also a prerequisite for P3 (auth header forwarding)
```

Read: P5 (Auth) must complete before P3 (Cohort A proxy) finalizes, because route migration depends on the unified `authFetch` helper. P5 is sequenced after P2 (which builds the replacement client scaffolding) and runs in parallel with P3 only if a feature flag isolates auth changes.

---

## 2. Phase Breakdowns

---

### Phase 0 — Pre-Migration Inventory & Backend Contract Lock

**Objective:** Produce a verifiable map of every coupling in `frontend/`, and lock the backend base URL contract. No code change.

**Branch:** `phase/0-inventory` (docs-only)
**Dependencies:** None
**Exit gate:** `IMPORT_GRAPH.md` committed; backend port unified; PM sign-off on Appendix-A questions from planning doc.

#### Tasks

| ID | Task | Output artifact | Risk |
|---|---|---|---|
| T0.1 | Run `rg -n "from ['\"]@/lib/(usageDb\|localDb\|disabledModelsDb\|requestDetailsDb\|db/)" frontend/src` | `inventory/sqlite-importers.txt` | 🟢 |
| T0.2 | Run `rg -n "from ['\"]@/lib/(oauth\|tunnel\|updater\|usage\|network\|mitm)" frontend/src` | `inventory/legacy-importers.txt` | 🟢 |
| T0.3 | Run `rg -n "open-sse" frontend/src` and `rg -n "open-sse" frontend/open-sse` | `inventory/stream-importers.txt` | 🟢 |
| T0.4 | Run `rg -n "process\.env\.(DATA_DIR\|APPDATA)" frontend/src` | `inventory/fs-importers.txt` | 🟢 |
| T0.5 | For each of the 24 `src/app/api/*/route.{ts,js}` files, read and classify as Cohort A (already thin proxy) or Cohort B (contains logic) | `inventory/route-cohorts.md` | 🟡 |
| T0.6 | Verify backend base URL: confirm with backend owner whether `:20128` (in `lib/api.ts`) or `:14322` (in `STRUCTURE.md`) is canonical. | Decision in tracker | 🟠 |
| T0.7 | Update `API_CONTRACT.md` to specify canonical `BACKEND_INTERNAL_URL` (server-only) and `NEXT_PUBLIC_API_URL` (browser-visible, if needed at all). | PR to docs | 🟡 |
| T0.8 | Confirm Cohort A endpoints exist on backend (`curl` each `/api/admin/*`, `/api/auth/*`, etc.). Document missing endpoints. | `inventory/backend-coverage.md` | 🟠 |
| T0.9 | Collect Appendix-A PM decisions from planning doc (7 questions: port, token storage, init/shutdown, cloud/tunnel/cli scope, locale, mitm, test scripts). | PM sign-off doc | 🔴 (blocks P3) |
| T0.10 | Capture HAR fixtures for top 10 user flows (login, dashboard load, providers CRUD, keys CRUD, settings save, usage chart, chat stream, OAuth start, OAuth callback, logout). | `fixtures/before/*.har` | 🟡 |
| T0.11 | Preflight repair before execution resumes: reconcile backend port references, task counts/status drift, missing authority docs, and final Next.js API route stance. | planning/status/doc reconciliation | 🔴 (blocks P1+) |

**Exit criteria:**
- All inventory files committed to repo (in `docs/migration/inventory/` or similar).
- T0.6 decision documented and reflected in `lib/api.ts` and `STRUCTURE.md`.
- T0.9 PM sign-off recorded.
- T0.11 preflight repair complete: task counts match files, authority docs are not missing, and backend port/API-route stance is consistent.

---

### Phase 1 — Shared Contracts Stabilization

**Objective:** Make `src/shared/constants/*`, `src/shared/utils/*`, `src/models/*` consistent and aligned with backend contracts. No behavior change.

**Branch:** `phase/1-shared-contracts`
**Dependencies:** P0
**Exit gate:** `tsc --noEmit` green; dashboard smoke-test renders unchanged.

#### Tasks

| ID | Task | Target | Risk |
|---|---|---|---|
| T1.1 | Run `rg -n "from ['\"]@/shared/constants" frontend/src` → list every importer; document import-path variants in `inventory/constants-paths.md` | inventory | 🟢 |
| T1.2 | Normalize `src/shared/constants/` to a single index pattern: every consumer imports from `@/shared/constants` (no deep paths). Add re-export `index.ts` if missing. | `src/shared/constants/index.ts` | 🟡 |
| T1.3 | For each constant that mirrors a backend value (provider list, error codes, model alias keys), add header comment: `// SOURCE: backend (do not edit locally)`. | `src/shared/constants/*` | 🟢 |
| T1.4 | Create `src/shared/contracts/` (new dir) containing TypeScript interfaces mirroring `API_CONTRACT.md` shapes: `LoginResponse`, `MeResponse`, `Provider`, `ApiKey`, `Combo`, `Alias`, `UsageStats`, `StreamChunk`. | `src/shared/contracts/*.ts` | 🟡 |
| T1.5 | Migrate existing duplicate types in `src/lib/admin-api.ts`, `analytics-api.ts`, `oauth-api.ts` to import from `src/shared/contracts/`. Do NOT change shapes. | refactor only | 🟡 |
| T1.6 | Run `tsc --noEmit`. Fix only path-related errors. No semantic changes. | clean tsc | 🟡 |
| T1.7 | Run `next build`. Verify clean. | clean build | 🟡 |
| T1.8 | Manual smoke: dashboard renders, provider list displays, settings page opens. | recorded smoke pass | 🟢 |

**Forbidden in this phase:**
- Adding new constants.
- Renaming existing constants.
- Introducing new abstractions (no factories, no enums replacing strings).
- Changing any function/component signature that consumers depend on.

**Exit criteria:**
- All consumers import constants from `@/shared/constants` only.
- `src/shared/contracts/` exists and is the single source of API interfaces.
- Build green, smoke green.

---

### Phase 2 — SQLite Dependency Isolation

**Objective:** Build the backend-driven replacement clients so that Phase 3/4 route migration has a target. Do NOT delete SQLite yet. Goal: freeze the shim layer (no new imports allowed) and prepare replacement APIs.

**Branch:** `phase/2-sqlite-isolation`
**Dependencies:** P1
**Exit gate:** Replacement clients exist for every SQLite-backed feature; ESLint rule blocks new imports of `db/`/`usageDb`/`localDb`; existing imports still work (shims intact).

#### Tasks

| ID | Task | Target | Risk |
|---|---|---|---|
| T2.1 | Inventory required client surface: from `T0.1` shim-importer list, build a function signature catalog (every exported function from `usageDb.js` and `localDb.js` that is called somewhere). | `inventory/required-clients.md` | 🟡 |
| T2.2 | Create `src/lib/http.ts` — single `authFetch<T>(path, init)` helper. Move logic out of `admin-api.ts` (which already has one). | `src/lib/http.ts` | 🟡 |
| T2.3 | Refactor `admin-api.ts` to import `authFetch` from `src/lib/http.ts`. Verify all callers still pass. | `src/lib/admin-api.ts` | 🟡 |
| T2.4 | Refactor `analytics-api.ts` and `oauth-api.ts` to use `src/lib/http.ts`. | `src/lib/*-api.ts` | 🟡 |
| T2.5 | Create `src/lib/usage-api.ts` — covers backend equivalents of: `getUsageHistory`, `getUsageStats`, `getChartData`, `getActiveRequests`, `getRecentLogs`, `getRequestDetails`, `getRequestDetailById`. | `src/lib/usage-api.ts` | 🟠 |
| T2.6 | Create `src/lib/settings-api.ts` — covers: `getSettings`, `updateSettings`, `isCloudEnabled`, `getCloudUrl`. | `src/lib/settings-api.ts` | 🟡 |
| T2.7 | Create `src/lib/providers-api.ts` (or extend `admin-api.ts`) — covers: `getProviderConnections`, `createProviderConnection`, `updateProviderConnection`, `deleteProviderConnection`, `reorderProviderConnections`, `cleanupProviderConnections`, `getProviderNodes`, `createProviderNode`, `updateProviderNode`, `deleteProviderNode`. | `src/lib/providers-api.ts` | 🟠 |
| T2.8 | Create `src/lib/proxy-pools-api.ts` — covers: `getProxyPools`, `createProxyPool`, `updateProxyPool`, `deleteProxyPool`. | `src/lib/proxy-pools-api.ts` | 🟡 |
| T2.9 | Create `src/lib/combos-api.ts` (or extend `admin-api.ts`) — covers: `getCombos`, `createCombo`, `updateCombo`, `deleteCombo`, `getModelAliases`, `setModelAlias`, `deleteModelAlias`. | `src/lib/combos-api.ts` | 🟡 |
| T2.10 | Create `src/lib/pricing-api.ts` — covers: `getPricing`, `getPricingForModel`, `updatePricing`, `resetPricing`, `resetAllPricing`. | `src/lib/pricing-api.ts` | 🟡 |
| T2.11 | Create `src/lib/models-api.ts` — covers: `getCustomModels`, `addCustomModel`, `deleteCustomModel`, `getMitmAlias`, `setMitmAliasAll`. | `src/lib/models-api.ts` | 🟡 |
| T2.12 | Create `src/lib/keys-api.ts` (if not subsumed by admin-api) — covers: `getApiKeys`, `createApiKey`, `updateApiKey`, `deleteApiKey`, `validateApiKey`. | `src/lib/keys-api.ts` | 🟡 |
| T2.13 | Create `src/lib/db-export-api.ts` — covers `exportDb`, `importDb` IF still in product scope (else mark as dropped per Appendix-A scope decision). | conditional | 🟡 |
| T2.14 | Add ESLint custom rule (or simple grep CI step) that fails build if any NEW file imports from `@/lib/db/`, `@/lib/usageDb`, `@/lib/localDb`, `@/lib/disabledModelsDb`, `@/lib/requestDetailsDb`, `@/lib/dataDir`. Whitelist current importers (from T0.1). | `.eslintrc` / CI script | 🟡 |
| T2.15 | Add a deprecation comment to each shim file (`usageDb.js`, `localDb.js`, `disabledModelsDb.js`, `requestDetailsDb.js`): `// DEPRECATED: do not add new imports. Use src/lib/*-api.ts instead.` | shim files | 🟢 |
| T2.16 | Verify backend coverage gap from T0.8: if any client function in T2.5–T2.13 has NO backend equivalent, STOP and file a backend ticket. Do not invent endpoints. | tickets list | 🔴 |
| T2.17 | Run `tsc --noEmit` + `next build`. | clean | 🟡 |

**Forbidden in this phase:**
- Modifying any existing `usageDb`/`localDb` shim function signature.
- Calling new clients from any existing code (that's Phase 3+).
- Inventing endpoints not in `API_CONTRACT.md`.
- Creating a "fallback to SQLite" path in the new clients.

**Exit criteria:**
- Replacement clients exist for every function in T2.1 inventory.
- ESLint rule active and blocks new SQLite imports.
- Build green.
- No production code yet uses the new clients (changes are in `src/lib/` only).

---

### Phase 3 — API Proxy Standardization (Cohort A)

**Objective:** Convert the 15 Cohort A routes under `src/app/api/` into thin proxies using `proxyToBackend`. Migrate component consumers to call the new clients from P2 where needed.

**Branch:** `phase/3-cohort-a-proxy`
**Dependencies:** P2 (replacement clients) + P5 partial (single auth helper available — see cross-phase note)
**Exit gate:** All 15 Cohort A routes use `proxyToBackend`; HAR replay diff is empty for those endpoints.

#### Cross-phase note
P5 produces the unified `authFetch`. If P5 is not yet complete, T3.0 below produces a minimal `proxyToBackend` that reads auth headers from `req.headers` directly (server-side, no client token logic). This isolates P3 from P5's UX work.

#### Tasks — Infrastructure (do once)

| ID | Task | Target | Risk |
|---|---|---|---|
| T3.0 | Create `src/lib/proxy.ts` exporting `proxyToBackend(req: NextRequest): Promise<Response>`. Behavior: read `BACKEND_INTERNAL_URL` (server env), forward method/path/query/body/`Authorization`/`Cookie`, stream response body if `content-type` is `text/event-stream` or `application/octet-stream`, return backend status verbatim, never parse business payloads. | `src/lib/proxy.ts` | 🟠 |
| T3.1 | Write unit test for `proxyToBackend`: mock backend with `nock`/`msw`, verify method/header forwarding, streaming passthrough, status code forwarding, error passthrough. | `src/lib/proxy.test.ts` | 🟡 |

#### Tasks — Route conversions (one task per route group, batch 3–5 routes per commit)

| ID | Task | Routes | Backend target | Risk |
|---|---|---|---|---|
| T3.2 | Convert auth/health/version | `api/auth/*`, `api/health/*`, `api/version/*` | `/api/auth/*`, `/health`, `/version` | 🟡 |
| T3.3 | Convert admin CRUD (set 1) | `api/keys/*`, `api/providers/*`, `api/provider-nodes/*`, `api/proxy-pools/*` | `/api/admin/keys`, `/api/admin/providers`, `/api/admin/nodes`, `/api/admin/proxy-pools` | 🟠 |
| T3.4 | Convert admin CRUD (set 2) | `api/combos/*`, `api/pricing/*`, `api/models/*`, `api/media-providers/*`, `api/tags/*` | `/api/admin/combos`, `/api/admin/pricing`, `/api/admin/models`, `/api/admin/media-providers`, `/api/admin/tags` | 🟠 |
| T3.5 | Convert settings/usage/oauth | `api/settings/*`, `api/usage/*`, `api/oauth/*` | `/api/admin/settings`, `/api/admin/usage`, `/api/oauth/*` | 🟠 |
| T3.6 | Per-route validation: HAR replay vs P0 fixtures; status/body/header diff (allowlist `content-type`, `content-length`, `set-cookie`). | fixtures/after/*.har | 🟠 |
| T3.7 | Component consumers migration (if any directly call `usageDb`/`localDb` instead of the Next route): switch to corresponding `src/lib/*-api.ts` from P2. | scattered components | 🟠 |
| T3.8 | Run `tsc --noEmit`, `next build`, manual smoke per migrated feature. | clean | 🟡 |

**Per-route checklist (apply to T3.2–T3.5 individually):**
1. Read existing route file.
2. If it already calls `fetch(BACKEND_URL...)` → mark done.
3. Else: replace body with the proxy template (see §5.2 in planning doc).
4. Remove every import of `usageDb`/`localDb`/`db/` from the route.
5. Capture HAR before & after.
6. Diff status code + body + allowed headers.
7. Commit.

**Forbidden in this phase:**
- Touching Cohort B routes (`v1/`, `v1beta/`, `translator/`, etc. — those are P4).
- Deleting the imported `usageDb`/`localDb` modules (those are P8).
- Changing response shapes "to clean things up."
- Adding caching/middleware that was not present before.

**Exit criteria:**
- All 15 Cohort A routes use `proxyToBackend`.
- `rg -n "usageDb\|localDb\|@/lib/db/" frontend/src/app/api` returns matches ONLY in Cohort B routes.
- HAR replay diff empty for Cohort A endpoints.
- Build + smoke green.

---

### Phase 4 — Cohort B Route Audit & Proxy

**Objective:** Convert the 9 audit-heavy routes. Each requires individual PM/backend decision.

**Branch:** `phase/4-cohort-b-proxy` (or `phase/4-<route>` per route)
**Dependencies:** P3 + P0.9 PM decisions
**Exit gate:** Each Cohort B route is either proxied, deleted, or explicitly marked "out of scope, kept as-is" with PM sign-off.

#### Per-route decision tree

For each route below:
1. Capture HAR + read current `route.{ts,js}`.
2. Decide: **PROXY** (backend equivalence exists), **DELETE** (frontend should not handle this), or **KEEP-AS-IS** (out of migration scope; PM sign-off required).
3. If PROXY: apply standard `proxyToBackend` pattern.
4. If DELETE: remove route, update any consumer to call backend directly or to remove the feature gracefully.
5. Validate.

#### Tasks (hardest last)

| ID | Task | Route | Likely action | Risk |
|---|---|---|---|---|
| T4.1 | `init/` decision + execution | `api/init/*` | Likely DELETE (backend manages lifecycle) | 🟡 |
| T4.2 | `shutdown/` decision + execution | `api/shutdown/*` | Likely DELETE | 🟡 |
| T4.3 | `locale/` decision + execution | `api/locale/*` | Likely KEEP (Next-side i18n) — confirm with PM | 🟡 |
| T4.4 | `cli-tools/` decision + execution | `api/cli-tools/*` | Depends on Appendix-A; likely PROXY or DELETE | 🟠 |
| T4.5 | `cloud/` decision + execution | `api/cloud/*` | Depends on Appendix-A; if cloud mode not in scope → DELETE | 🟠 |
| T4.6 | `tunnel/` decision + execution | `api/tunnel/*` | Depends on Appendix-A; likely DELETE if MITM/tunnel out of scope | 🟠 |
| T4.7 | `translator/` decision + execution | `api/translator/*` | DELETE — backend has `internal/translator/` | 🟠 |
| T4.8 | `v1beta/` proxy with streaming passthrough | `api/v1beta/*` | PROXY (Gemini-compatible). All streaming endpoints must pipe `ReadableStream`. | 🔴 |
| T4.9 | `v1/` proxy with streaming passthrough | `api/v1/*` | PROXY (OpenAI-compatible). Highest end-user blast radius. **Touches P7 streaming.** | 🔴 |
| T4.10 | HAR replay verification for every converted route | fixtures | 🔴 |
| T4.11 | Streaming-specific verification for T4.8/T4.9: short response, long response, abort mid-stream, network blip, provider 5xx, provider auth fail. | test matrix | 🔴 |

**Per-route checklist (extends T3 checklist):**
- Confirm backend endpoint exists AND schema matches (`API_CONTRACT.md`).
- For streaming endpoints: verify `ReadableStream` is piped, NOT buffered. Verify chunk ordering preserved.
- For deletion: verify zero consumers in `frontend/src` via `rg`.

**Forbidden in this phase:**
- "Improving" the v1/v1beta response format.
- Buffering streaming responses "for ordering safety."
- Adding retry logic in the proxy (backend owns retries).
- Recreating the deleted `translator/` route's logic anywhere in frontend.

**Exit criteria:**
- All 9 Cohort B routes have a decision recorded in tracker.
- All PROXY routes pass HAR + streaming verification.
- All DELETE routes have zero remaining consumers.
- Build + smoke + streaming test matrix green.

---

### Phase 5 — Auth + Token Source Unification

**Objective:** Single token source of truth via `src/lib/api.ts`. Preserve 9router auth UX (password-only login, redirect behavior, session persistence on reload).

**Branch:** `phase/5-auth-unification`
**Dependencies:** P1 (contracts); ideally runs in parallel with early P3 if isolated behind component-level changes.
**Exit gate:** Manual auth regression suite passes; one token storage path; no duplicate token caches.

#### Tasks

| ID | Task | Target | Risk |
|---|---|---|---|
| T5.1 | Audit current token paths: `rg -n "localStorage.(getItem\|setItem).*token" frontend/src` and `rg -n "getToken\|setToken" frontend/src`. | `inventory/token-paths.md` | 🟢 |
| T5.2 | Decision lock (from P0.9): keep `localStorage` (current) per planning doc §7.6. Document in tracker. | doc only | 🟢 |
| T5.3 | Ensure `src/lib/api.ts:getToken/setToken` are the ONLY direct accessors. Refactor any other accessor to delegate. | `src/lib/api.ts` + callers | 🟠 |
| T5.4 | Refactor `src/store/userStore.js`: remove any token caching. Keep user profile cache only. Replace token-related actions with calls to `lib/api.ts`. | `src/store/userStore.js` | 🟠 |
| T5.5 | Verify `src/lib/admin-api.ts:authFetch` (now in `src/lib/http.ts` from P2.2) is the single auth helper. Replace any duplicate auth-fetch in `analytics-api.ts`, `oauth-api.ts`. | `src/lib/*-api.ts` | 🟡 |
| T5.6 | Audit login flow `src/app/login/*`: form → `lib/api.ts:login()` → `setToken()` → redirect. Preserve redirect target verbatim (parse query `?next=...` if 9router supported it). | `src/app/login/*` | 🟠 |
| T5.7 | Audit OAuth callback `src/app/callback/*`: must only forward the callback code to backend. Backend owns PKCE, code exchange, token storage. No logic in frontend. | `src/app/callback/*` | 🟠 |
| T5.8 | Audit logout flow: clear `localStorage.token`, clear `userStore`, redirect to `/login`. Verify no orphan state. | logout button(s) | 🟡 |
| T5.9 | Auth regression manual test: login with valid → dashboard loads; login with invalid → inline error preserved; reload while logged in → session persists; logout → cleared; expired token → redirected to login; OAuth start → callback → dashboard. | regression doc | 🔴 |
| T5.10 | If any component reads token directly from `localStorage`: refactor to use `getToken()`. | scattered components | 🟠 |

**Forbidden in this phase:**
- Switching to cookies without explicit PM approval (planning doc §7.6).
- Decoding JWT in the frontend.
- Storing tokens in Zustand `userStore`.
- Changing the login form fields/order/copy.
- Changing the OAuth callback URL path.

**Exit criteria:**
- Single source of truth: `src/lib/api.ts` token helpers.
- `userStore` has no token state.
- Manual auth regression suite passes for all 6 scenarios in T5.9.
- Build green.

---

### Phase 6 — Legacy Module Deprecation

**Objective:** Remove all consumers of `src/lib/{oauth,tunnel,updater,usage,network,mitm}/` and `src/mitm/`. Delete the modules.

**Branch:** `phase/6-legacy-deprecation` (one branch per module group)
**Dependencies:** P3, P4, P5 (most consumers should already be migrated through route conversion)
**Exit gate:** `rg` returns zero imports of legacy dirs; modules deleted.

#### Per-module sub-phases

| ID | Task | Module | Risk |
|---|---|---|---|
| T6.1 | `src/lib/oauth/` — list consumers from P0.2; replace each with `oauth-api.ts`; verify zero imports; delete dir. | oauth | 🟠 |
| T6.2 | `src/lib/tunnel/` — same workflow; if `tunnel/` is out-of-scope per Appendix-A, delete consumers too. | tunnel | 🟠 |
| T6.3 | `src/lib/updater/` + `src/lib/appUpdater.js` — same workflow. App-updater logic moves to backend or is dropped per PM. | updater | 🟡 |
| T6.4 | `src/lib/usage/` — replace consumers with `usage-api.ts` (P2.5); delete dir. | usage | 🟠 |
| T6.5 | `src/lib/network/` — same workflow. | network | 🟡 |
| T6.6 | `src/lib/mitmAliasCache.js` + `src/mitm/` — per Appendix-A scope. If MITM out of scope, delete consumers too. | mitm | 🟠 |
| T6.7 | `src/lib/providerNormalization.js` — backend-replace or delete. | providerNormalization | 🟡 |
| T6.8 | `src/lib/initCloudSync.js` — almost certainly delete (cloud sync is backend's responsibility). | initCloudSync | 🟡 |
| T6.9 | `src/lib/consoleLogBuffer.js` — classify: if pure UI logging buffer → keep; if persistence → delete. | consoleLogBuffer | 🟢 |
| T6.10 | After each module deletion: `tsc --noEmit` + `next build` + smoke. | per-batch | 🟡 |

**Per-module checklist:**
1. From P0.2 inventory: list every consumer.
2. For each consumer: confirm a replacement client exists (P2). If not, STOP.
3. Replace import in consumer; adjust types.
4. Run `tsc --noEmit`.
5. Verify smoke pass for that feature.
6. Verify `rg -n "from .@/lib/<module>" frontend/src` returns zero.
7. Delete the module directory.
8. Commit.

**Forbidden in this phase:**
- Mass-deleting all legacy dirs in one commit.
- Leaving stub files behind ("just in case").
- Recreating any module under a different name.
- Touching `src/lib/db/` or the SQLite shims (those are P8).

**Exit criteria:**
- All 9 legacy module groups (T6.1–T6.9) deleted or explicitly kept with PM justification.
- Build + smoke green after each batch.

---

### Phase 7 — Streaming Normalization (open-sse Retirement)

**Objective:** Retire `frontend/open-sse/` entirely. Backend owns stream normalization. Frontend renders only.

**Branch:** `phase/7-streaming-normalization`
**Dependencies:** P4 (v1/v1beta routes proxied with streaming passthrough)
**Exit gate:** `frontend/open-sse/` deleted; streaming test matrix passes 9router parity.

#### Sub-phases

##### 7a — Backend contract verification

| ID | Task | Target | Risk |
|---|---|---|---|
| T7.1 | Inspect `backend/internal/stream/` and document the normalized chunk format (event types, JSON schema). | `docs/STREAM_CONTRACT.md` | 🔴 |
| T7.2 | Confirm event types: `delta`, `done`, `error`, and optionally `retry`, `usage`. Each must have stable JSON shape. Update `API_CONTRACT.md` streaming section. | docs | 🔴 |
| T7.3 | If any provider-format leaks through backend stream (raw OpenAI/Claude/Gemini deltas reach frontend), file backend ticket. Do NOT patch in frontend. | tickets | 🔴 |

##### 7b — Frontend SSE consumer

| ID | Task | Target | Risk |
|---|---|---|---|
| T7.4 | Create `src/sse/consumer.ts`: owns `fetch` + `ReadableStream` (POST-SSE) and `EventSource` (GET-SSE) consumers. Parses only the normalized format from T7.1. Exposes typed callbacks: `onDelta`, `onDone`, `onError`, `onRetry`, `onUsage`. | `src/sse/consumer.ts` | 🔴 |
| T7.5 | Unit-test `consumer.ts`: chunk ordering, abort handling, error propagation, retry event handling. | `src/sse/consumer.test.ts` | 🟠 |
| T7.6 | Slim `src/sse/handlers/` and `src/sse/services/` to UI-only logic. Delete any provider-format parsing. | `src/sse/*` | 🟠 |

##### 7c — UI migration

| ID | Task | Target | Risk |
|---|---|---|---|
| T7.7 | Inventory every component that imports from `open-sse/*` or old `src/sse/handlers/*` provider-specific code. | `inventory/stream-consumers.md` | 🟡 |
| T7.8 | Migrate chat UI / stream renderer components to use `src/sse/consumer.ts`. Preserve incremental rendering, abort UI, retry indicator UI. | chat components | 🔴 |
| T7.9 | Feature flag `NEXT_PUBLIC_STREAM_V2`: gate the new consumer behind this flag during transition. Default OFF in dev until T7.10 passes; flip ON after. | env | 🟠 |
| T7.10 | Streaming parity test matrix (see planning §6.3 / 7d): execute each case against the new consumer. | regression report | 🔴 |

##### 7d — Cleanup

| ID | Task | Target | Risk |
|---|---|---|---|
| T7.11 | Delete `frontend/open-sse/` entirely. Single commit. | `open-sse/` | 🔴 |
| T7.12 | Verify `rg -n "open-sse" frontend/` returns zero matches. | grep | 🟡 |
| T7.13 | Remove `NEXT_PUBLIC_STREAM_V2` feature flag (now the only path). | env | 🟢 |
| T7.14 | Add lint rule blocking new imports referencing `open-sse` paths. | eslint | 🟡 |

**Forbidden in this phase:**
- Adding provider-format parsing in frontend (OpenAI/Claude/Gemini delta shapes).
- Reordering chunks in the consumer (no `Promise.all`, no batching that changes visible order).
- Inferring stream completion from connection close (always wait for explicit `done`).
- Adding frontend-side retry logic (backend owns retries).
- Throttling chunks for "smoother" rendering.

**Exit criteria:**
- `frontend/open-sse/` deleted.
- Streaming test matrix (planning §6.3 7d) passes 9router parity.
- Single frontend stream consumer in `src/sse/consumer.ts`.
- Build + smoke + streaming regression green.

---

### Phase 8 — SQLite Engine Removal

**Objective:** Delete `src/lib/db/`, all shims, `dataDir.js`, and the SQLite driver from `package.json`. Frontend has zero persistence.

**Branch:** `phase/8-sqlite-removal`
**Dependencies:** P3, P4, P6 (zero shim importers)
**Exit gate:** `rg` confirms zero imports; build green; smoke green; `better-sqlite3` not in lockfile.

#### Tasks

| ID | Task | Target | Risk |
|---|---|---|---|
| T8.1 | Re-verify zero shim importers: `rg -n "from ['\"]@/lib/(usageDb\|localDb\|disabledModelsDb\|requestDetailsDb\|db/\|dataDir\|mitmAliasCache)" frontend/src` MUST return zero matches. | grep | 🔴 |
| T8.2 | Delete shim files: `src/lib/usageDb.js`, `localDb.js`, `disabledModelsDb.js`, `requestDetailsDb.js`, `mitmAliasCache.js`. Single commit. | files | 🟠 |
| T8.3 | Delete `src/lib/db/` directory recursively. Single commit. | `src/lib/db/` | 🔴 |
| T8.4 | Delete `src/lib/dataDir.js`. Single commit. | file | 🟡 |
| T8.5 | Remove SQLite driver from `package.json`: `better-sqlite3`, `sqlite3`, any other SQLite-related package. Re-run `npm install`. Commit lockfile change. | `package.json`, lockfile | 🟠 |
| T8.6 | Search for any direct `require('better-sqlite3')` or `from 'better-sqlite3'`: `rg -n "better-sqlite3\|sqlite3" frontend/`. Must return zero. | grep | 🟠 |
| T8.7 | Build + smoke. Test usage charts, settings save, providers CRUD, keys CRUD specifically (the features most likely to have hidden SQLite calls). | smoke | 🔴 |
| T8.8 | Add lint rule: imports from `@/lib/db/`, `usageDb`, `localDb`, `dataDir`, `disabledModelsDb`, `requestDetailsDb`, `mitmAliasCache`, `better-sqlite3`, `sqlite3`, `fs/promises` (in client-side code) are BLOCKED. | eslint | 🟠 |

**Forbidden in this phase:**
- Leaving any file as a no-op stub ("just in case").
- Adding `localStorage`-based fallback persistence.
- Migrating SQLite to IndexedDB or any other client storage.
- Recreating any DB module under a different name.

**Exit criteria:**
- All 8 removal targets (planning §8.1) deleted.
- SQLite driver gone from lockfile.
- `rg` confirms zero residual imports.
- Build + full smoke green.

---

### Phase 9 — Dual App Shell Collapse

**Objective:** One root layout, one root page. Choose TypeScript. Delete `.js` siblings.

**Branch:** `phase/9-shell-collapse`
**Dependencies:** P1–P8 (low coupling, but safer last)
**Exit gate:** Only `.tsx` versions of `layout` and `page` exist at `src/app/`.

#### Tasks

| ID | Task | Target | Risk |
|---|---|---|---|
| T9.1 | Diff `src/app/page.js` vs `src/app/page.tsx`. Document any behavior in `.js` missing from `.tsx`. | `inventory/shell-diff.md` | 🟡 |
| T9.2 | Diff `src/app/layout.js` vs `src/app/layout.tsx`. Same. | `inventory/shell-diff.md` | 🟡 |
| T9.3 | Port missing behaviors from `.js` to `.tsx`. Preserve verbatim. | `src/app/*.tsx` | 🟠 |
| T9.4 | Delete `src/app/page.js` and `src/app/layout.js`. Single commit. | files | 🟠 |
| T9.5 | Audit `frontend/jsconfig.json` vs `tsconfig.json` for path alias drift. If both define `@/*`, ensure they agree. If all `.js` is gone from `src/`, delete `jsconfig.json` (T10 task). | configs | 🟡 |
| T9.6 | Visual diff homepage and root layout in browser vs pre-collapse screenshot. | visual diff | 🟡 |
| T9.7 | Build + full smoke. | clean | 🟡 |

**Forbidden in this phase:**
- Restructuring the layout tree.
- Adding new routes/pages.
- Updating Next.js version as part of this phase.

**Exit criteria:**
- `rg -n "page\.js\|layout\.js" frontend/src/app | head -5` shows no top-level `.js`.
- Visual diff acceptable.
- Build green.

---

### Phase 10 — Dead Code & Dependency Cleanup

**Objective:** Remove orphan imports, unused npm packages, ad-hoc test scripts, empty directories.

**Branch:** `phase/10-cleanup`
**Dependencies:** P1–P9
**Exit gate:** `depcheck` clean; `ts-prune` clean; build green; bundle size measurably smaller.

#### Tasks

| ID | Task | Target | Risk |
|---|---|---|---|
| T10.1 | Run `npx depcheck frontend/` and triage. Remove confirmed-unused packages. | `package.json` | 🟡 |
| T10.2 | Run `npx ts-prune` (or `knip`) and remove confirmed-unused exports. | source tree | 🟡 |
| T10.3 | Relocate ad-hoc scripts: `frontend/login-test.js`, `test-api-me.js`, `test-dashboard.js`, `test-detailed.js`, `serve-static.js` → `frontend/scripts/` OR delete OR convert to vitest/playwright (per Appendix-A decision). | scripts | 🟢 |
| T10.4 | Delete `frontend/jsconfig.json` if `rg -n "\.js$" frontend/src` returns zero (all files now `.ts`/`.tsx`). | config | 🟡 |
| T10.5 | Remove `i18n` orphan files if any (verify locale routes are wired). | `src/i18n` | 🟢 |
| T10.6 | Audit `frontend/CLAUDE.md` (currently 11B) — empty or stub? Delete or populate with current architecture rules. | doc | 🟢 |
| T10.7 | Run final `next build`; measure bundle size; compare to pre-migration. | metrics | 🟢 |

**Forbidden in this phase:**
- Removing dependencies that `depcheck` flags but you cannot trace (some may be runtime-loaded).
- Combining cleanup with refactor (no behavior changes here).

**Exit criteria:**
- `depcheck` clean (or all exceptions documented).
- `ts-prune` clean (or exceptions documented).
- Test scripts relocated.
- Bundle size delta documented in tracker.

---

### Phase 11 — Final Stabilization & Architecture Lock

**Objective:** Lock the architecture in place via tooling and documentation. Make regressions impossible without conscious override.

**Branch:** `phase/11-architecture-lock`
**Dependencies:** P1–P10
**Exit gate:** CI enforces the rules; documentation reflects the final state.

#### Tasks

| ID | Task | Target | Risk |
|---|---|---|---|
| T11.1 | Add `frontend/ARCHITECTURE.md` describing the thin-client model, allowed/forbidden patterns, and how new features should be wired. | new doc | 🟢 |
| T11.2 | Update `STRUCTURE.md` to reflect final repo layout (no `open-sse/`, no `src/lib/db/`, etc.). | doc | 🟢 |
| T11.3 | Update `FRONTEND_REFACTOR_TRACKER.md` to mark migration complete; record final decisions. | doc | 🟢 |
| T11.4 | Add CI step (lint or grep) that fails the build if any of these appear in `frontend/src` (excluding migration-marked files): `better-sqlite3`, `sqlite3`, `from '@/lib/db'`, `from '@/lib/usageDb'`, `from '@/lib/localDb'`, `import .* from 'fs'` (in non-server-route files), `open-sse`. | `.github/workflows/*` or `eslint.config.mjs` | 🟠 |
| T11.5 | Add a pre-commit hook (Husky or similar) that runs the same checks locally. | `.husky/*` | 🟡 |
| T11.6 | Run full regression sweep against `FEATURE_PARITY.md`: every listed behavior verified manually. | regression report | 🔴 |
| T11.7 | Run HAR replay across all 24 original Cohort A+B endpoints (those that still exist) against P0 fixtures. | diff report | 🟠 |
| T11.8 | Final streaming parity test (planning §6.3) against 9router reference. | report | 🔴 |
| T11.9 | Sign-off: PM + backend owner + frontend owner. | tracker | 🟢 |

**Forbidden in this phase:**
- Introducing new features.
- Changing UX.
- Skipping FEATURE_PARITY.md items "for speed."

**Exit criteria:**
- CI enforces architecture rules.
- All FEATURE_PARITY.md items verified.
- Full HAR + streaming regression green.
- Three-party sign-off in tracker.

---

## 3. Cross-Phase Concerns

### 3.1 Branching & PR strategy

- **One branch per phase.** Name: `phase/N-<slug>`.
- **PR title prefix:** `[P<N>]`. Example: `[P3] Convert auth/health/version routes to proxy`.
- **Squash-merge into `main`.**
- **No phase straddles two PRs** unless explicitly noted (P7 may span up to 3 PRs behind feature flag).

### 3.2 Commit discipline

- **One task per commit** as default.
- **Batch only when explicitly allowed** (e.g., T3.3 batches 4 routes).
- **Commit message format:** `[T<id>] <imperative summary>`. Example: `[T3.4] Proxy api/combos to /api/admin/combos`.
- **No "fix lint" or "wip" commits in shared branches.** Use `git commit --amend` or `--fixup` before push.

### 3.3 Build gating

- **`tsc --noEmit`** runs on every push.
- **`next build`** runs on every PR.
- **Smoke suite** runs on every PR (manual or automated).
- **HAR replay** runs on PRs touching `src/app/api/*`.
- **Streaming regression** runs on PRs touching `src/sse/*`, `open-sse/*`, or `src/app/api/v1*`.

### 3.4 Rollback boundaries

| Phase | Rollback unit | Risk if rolled back mid-way |
|---|---|---|
| 0 | Doc-only | None |
| 1 | Whole PR | Type errors only |
| 2 | Whole PR | None (new code, not consumed yet) |
| 3 | Per-route batch | Affected endpoints fall back to legacy behavior |
| 4 | Per-route | Same |
| 5 | Whole PR | Login UX may regress to dual-path |
| 6 | Per-module | Deleted module's callers break |
| 7 | Behind feature flag → per sub-phase | Streaming regression — keep flag for 1 release |
| 8 | Whole PR | SQLite reappears; previous shims must be re-added |
| 9 | Whole PR | `.js` shells must be re-added |
| 10 | Per-cleanup-group | Dead code reappears |
| 11 | Doc + CI only | Easy revert |

### 3.5 Parity verification matrix (run at end of P3, P4, P7, P11)

| Area | Verification |
|---|---|
| Auth UX | Login form fields, redirect chain, session reload, logout |
| Provider CRUD | Create, list, update, delete, enable/disable, optimistic UI |
| Streaming | Chunk order, abort, retry visibility, error inline rendering |
| Settings | Instant save, optimistic update, persistence on reload |
| OAuth | Per-provider login flow, callback, token refresh visibility |
| Usage charts | Time-range filter, chart rendering, request log table |
| Error handling | Inline errors, retry visibility, provider failure fallback |

---

## 4. Risk Register

| Risk | Phase | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| Backend port mismatch breaks all routes | P0 | High | Critical | Lock port in T0.6 before any code change |
| Streaming chunk reordering | P7 | Medium | Critical | Strict in-order consumer; tests in T7.5 |
| Auth regression on reload | P5 | Medium | Critical | T5.9 manual regression; preserve `localStorage` |
| Missing backend endpoint for migrated feature | P2/P3/P4 | Medium | High | T2.16 gate; never invent endpoints |
| Hidden SQLite consumer surfaces at runtime | P8 | Medium | High | T8.1 + T8.6 + T8.7 smoke matrix |
| Cohort B route v1/v1beta breaks paying users | P4 | Low | Critical | HAR replay + streaming matrix in T4.11 |
| `open-sse` removal breaks 9router UX parity | P7 | Medium | Critical | Feature flag in T7.9; 9router side-by-side test |
| Lint rule blocks legitimate FS access in server-side Next code | P11 | Medium | Medium | Scope rule to client-only files |

---

## 5. Open Questions (Block Phase 3 Entry)

These MUST be answered before P3 starts (carried from planning doc Appendix A):

1. ⏳ **Backend port:** `20128` vs `14322` — which is canonical?
2. ⏳ **Token storage:** keep `localStorage` or migrate to HttpOnly cookies?
3. ⏳ **`init/`, `shutdown/` routes:** delete or no-op proxy?
4. ⏳ **`cloud/`, `tunnel/`, `cli-tools/`:** in product scope?
5. ⏳ **`locale/` route:** keep Next-side or proxy?
6. ⏳ **`mitm/` directory:** product feature or 9router legacy?
7. ⏳ **Test scripts:** delete, move to `scripts/`, or convert to vitest?

---

## 6. Glossary

- **Cohort A:** API routes that are clearly proxy-convertible without business-logic changes (15 routes).
- **Cohort B:** API routes that historically host logic and require individual decisions (9 routes).
- **Shim:** A file whose only purpose is re-exporting from another module (e.g., `usageDb.js`).
- **`proxyToBackend`:** The single Next.js → Go proxy helper (`src/lib/proxy.ts`).
- **`authFetch`:** The single browser → Next-route authenticated fetch helper (`src/lib/http.ts`).
- **9router parity:** Behavior must match 9router as user-visible. Internal architecture may differ.
- **Streaming v2:** The new `src/sse/consumer.ts`-based streaming path (gated behind feature flag during P7).

---

## 7. Phase-by-Phase Checklist (Print-Friendly)

```
PHASE 0  [ ] Inventory + contract lock
PHASE 1  [ ] Shared contracts stabilized
PHASE 2  [ ] SQLite-replacement clients built; ESLint guard active
PHASE 3  [ ] Cohort A routes (15) proxied + HAR diff empty
PHASE 4  [ ] Cohort B routes (9) audited + proxied/deleted
PHASE 5  [ ] Auth/token unified; 6-scenario regression passes
PHASE 6  [ ] Legacy modules deleted (oauth, tunnel, updater, usage, network, mitm)
PHASE 7  [ ] open-sse/ retired; streaming matrix passes 9router parity
PHASE 8  [ ] SQLite engine + driver removed; lint guard active
PHASE 9  [ ] Dual app shell collapsed; only .tsx remains
PHASE 10 [ ] Dead code + deps cleaned; bundle size delta recorded
PHASE 11 [ ] Architecture locked in CI; PARITY sign-off complete
```

---

*End of task index. Total tasks: ~85 across 12 phases. Estimated rollback boundary: 1 PR per phase except P7 (up to 3 PRs behind feature flag).*
