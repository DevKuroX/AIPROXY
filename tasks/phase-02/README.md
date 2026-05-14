# Phase 2 — SQLite Dependency Isolation

**Objective**: Build the backend-driven replacement clients so that Phase 3/4 route migration has a target. Do NOT delete SQLite yet.

**Branch**: `phase/2-sqlite-isolation`

**Dependencies**: Phase 1

**Exit Gate**: Replacement clients exist for every SQLite-backed feature; ESLint rule blocks new imports.

---

## Tasks

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T2.1](T2.1-required-clients-inventory.md) | Required client surface inventory | `inventory/required-clients.md` | 🟡 |
| [T2.2](T2.2-create-http-helper.md) | Create HTTP helper | `src/lib/http.ts` | 🟡 |
| [T2.3](T2.3-refactor-admin-api.md) | Refactor admin-api.ts | `src/lib/admin-api.ts` | 🟡 |
| [T2.4](T2.4-refactor-api-clients.md) | Refactor analytics/oauth APIs | `src/lib/*-api.ts` | 🟡 |
| [T2.5](T2.5-create-usage-api-client.md) | Create usage API client | `src/lib/usage-api.ts` | 🟠 |
| [T2.6](T2.6-create-settings-api-client.md) | Create settings API client | `src/lib/settings-api.ts` | 🟡 |
| [T2.7](T2.7-create-providers-api-client.md) | Create providers API client | `src/lib/providers-api.ts` | 🟠 |
| [T2.8](T2.8-create-proxy-pools-api-client.md) | Create proxy pools API client | `src/lib/proxy-pools-api.ts` | 🟡 |
| [T2.9](T2.9-create-combos-api-client.md) | Create combos API client | `src/lib/combos-api.ts` | 🟡 |
| [T2.10](T2.10-create-pricing-api-client.md) | Create pricing API client | `src/lib/pricing-api.ts` | 🟡 |
| [T2.11](T2.11-create-models-api-client.md) | Create models API client | `src/lib/models-api.ts` | 🟡 |
| [T2.12](T2.12-create-keys-api-client.md) | Create keys API client | `src/lib/keys-api.ts` | 🟡 |
| [T2.13](T2.13-create-db-export-api-client.md) | Create db export API client | conditional | 🟡 |
| [T2.14](T2.14-eslint-sqlite-blocking.md) | ESLint rule for SQLite blocking | `.eslintrc` / CI script | 🟡 |
| [T2.15](T2.15-deprecation-comments.md) | Shim deprecation comments | shim files | 🟢 |
| [T2.16](T2.16-backend-coverage-verification.md) | Backend coverage gap verification | tickets list | 🔴 |
| [T2.17](T2.17-build-verification.md) | Build verification | clean | 🟡 |

---

## Forbidden

- Modifying shim function signatures
- Calling new clients from existing code (that's Phase 3+)
- Inventing endpoints not in API_CONTRACT.md
- Creating fallback to SQLite

## Exit Criteria

- [ ] Replacement clients exist for every function in T2.1 inventory
- [ ] ESLint rule active and blocks new SQLite imports
- [ ] Build green
- [ ] No production code yet uses new clients
