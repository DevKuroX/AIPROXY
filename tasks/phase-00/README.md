# Phase 0 — Pre-Migration Inventory & Backend Contract Lock

**Objective**: Produce a verifiable map of every coupling in `frontend/`, and lock the backend base URL contract. No code change.

**Branch**: `phase/0-inventory` (docs-only)

**Dependencies**: None

**Exit Gate**: `IMPORT_GRAPH.md` committed; backend port unified; PM sign-off on Appendix-A questions.

---

## Tasks

| ID | Task | Output Artifact | Risk |
|----|------|-----------------|------|
| [T0.1](T0.1-sqlite-import-inventory.md) | SQLite import inventory | `inventory/sqlite-importers.txt` | 🟢 |
| [T0.2](T0.2-legacy-import-inventory.md) | Legacy module import inventory | `inventory/legacy-importers.txt` | 🟢 |
| [T0.3](T0.3-stream-import-inventory.md) | Streaming module import inventory | `inventory/stream-importers.txt` | 🟢 |
| [T0.4](T0.4-filesystem-import-inventory.md) | Filesystem import inventory | `inventory/fs-importers.txt` | 🟢 |
| [T0.5](T0.5-route-cohort-classification.md) | Route cohort classification | `inventory/route-cohorts.md` | 🟡 |
| [T0.6](T0.6-backend-url-decision.md) | Backend URL decision lock | Decision in tracker | 🟠 |
| [T0.7](T0.7-api-contract-update.md) | API Contract URL update | `API_CONTRACT.md` | 🟡 |
| [T0.8](T0.8-backend-coverage-verification.md) | Backend coverage verification | `inventory/backend-coverage.md` | 🟠 |
| [T0.9](T0.9-pm-decisions-collection.md) | PM decisions collection | PM sign-off doc | 🔴 |
| [T0.10](T0.10-har-fixtures-collection.md) | HAR fixtures collection | `fixtures/before/*.har` | 🟡 |

---

## Exit Criteria

- [ ] All inventory files committed to repo
- [ ] T0.6 decision documented and reflected in `lib/api.ts` and `STRUCTURE.md`
- [ ] T0.9 PM sign-off recorded
