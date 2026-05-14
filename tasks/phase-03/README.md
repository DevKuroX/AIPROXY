# Phase 3 — API Proxy Standardization (Cohort A)

**Objective**: Convert the 15 Cohort A routes under `src/app/api/` into thin proxies using `proxyToBackend`.

**Branch**: `phase/3-cohort-a-proxy`

**Dependencies**: Phase 2 + Phase 5 partial

**Exit Gate**: All 15 Cohort A routes use `proxyToBackend`; HAR replay diff is empty.

---

## Tasks

| ID | Task | Routes | Risk |
|----|------|--------|------|
| [T3.0](T3.0-create-proxy-helper.md) | Create proxy helper | `src/lib/proxy.ts` | 🟠 |
| [T3.1](T3.1-proxy-helper-tests.md) | Proxy helper unit tests | `src/lib/proxy.test.ts` | 🟡 |
| [T3.2](T3.2-convert-auth-health-version.md) | Convert auth/health/version | 3 routes | 🟡 |
| [T3.3](T3.3-convert-admin-crud-set1.md) | Convert admin CRUD set 1 | 4 routes | 🟠 |
| [T3.4](T3.4-convert-admin-crud-set2.md) | Convert admin CRUD set 2 | 5 routes | 🟠 |
| [T3.5](T3.5-convert-settings-usage-oauth.md) | Convert settings/usage/oauth | 3 routes | 🟠 |
| [T3.6](T3.6-har-replay-validation.md) | HAR replay validation | fixtures | 🟠 |
| [T3.7](T3.7-component-consumers-migration.md) | Component consumers migration | scattered | 🟠 |
| [T3.8](T3.8-build-smoke.md) | Build and smoke verification | clean | 🟡 |

---

## Cohort A Routes

```
auth/
health/
version/
combos/
keys/
providers/
provider-nodes/
proxy-pools/
settings/
pricing/
models/
media-providers/
tags/
usage/
oauth/
```

## Per-Route Checklist

1. Read existing route file
2. If already calls `fetch(BACKEND_URL)` → mark done
3. Else: replace with proxy template
4. Remove `usageDb`/`localDb`/`db/` imports
5. Capture HAR before & after
6. Diff status + body + headers
7. Commit

## Forbidden

- Touching Cohort B routes
- Deleting SQLite modules
- Changing response shapes

## Exit Criteria

- [ ] All 15 Cohort A routes use `proxyToBackend`
- [ ] `rg` confirms zero SQLite imports in Cohort A routes
- [ ] HAR replay diff empty
- [ ] Build + smoke green
