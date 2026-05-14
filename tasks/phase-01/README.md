# Phase 1 — Shared Contracts Stabilization

**Objective**: Make `src/shared/constants/*`, `src/shared/utils/*`, `src/models/*` consistent and aligned with backend contracts. No behavior change.

**Branch**: `phase/1-shared-contracts`

**Dependencies**: Phase 0

**Exit Gate**: `tsc --noEmit` green; dashboard smoke-test renders unchanged.

---

## Tasks

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T1.1](T1.1-constants-import-inventory.md) | Constants import path inventory | `inventory/constants-paths.md` | 🟢 |
| [T1.2](T1.2-constants-index-normalization.md) | Constants index normalization | `src/shared/constants/index.ts` | 🟡 |
| [T1.3](T1.3-constants-source-comments.md) | Constants source comments | `src/shared/constants/*` | 🟢 |
| [T1.4](T1.4-api-contracts-interfaces.md) | API contracts TypeScript interfaces | `src/shared/contracts/*.ts` | 🟡 |
| [T1.5](T1.5-migrate-duplicate-types.md) | Migrate duplicate types to contracts | refactor only | 🟡 |
| [T1.6](T1.6-typescript-check.md) | TypeScript check | clean tsc | 🟡 |
| [T1.7](T1.7-build-check.md) | Build check | clean build | 🟡 |
| [T1.8](T1.8-smoke-test.md) | Phase 1 smoke test | recorded smoke pass | 🟢 |

---

## Forbidden

- Adding new constants
- Renaming existing constants
- Introducing new abstractions
- Changing any function/component signature

## Exit Criteria

- [ ] All consumers import constants from `@/shared/constants` only
- [ ] `src/shared/contracts/` exists and is the single source of API interfaces
- [ ] Build green, smoke green
