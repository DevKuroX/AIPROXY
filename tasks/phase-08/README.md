# Phase 8 — SQLite Engine Removal

**Objective**: Delete `src/lib/db/`, all shims, `dataDir.js`, and the SQLite driver from `package.json`. Frontend has zero persistence.

**Branch**: `phase/8-sqlite-removal`

**Dependencies**: Phase 3, Phase 4, Phase 6

**Exit Gate**: `rg` confirms zero imports; build green; `better-sqlite3` not in lockfile.

---

## Tasks

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T8.1](T8.1-zero-importers-verification.md) | Zero importers verification | grep | 🔴 |
| [T8.2](T8.2-delete-shim-files.md) | Delete shim files | shim files | 🟠 |
| [T8.3](T8.3-delete-sqlite-db.md) | Delete SQLite DB directory | `src/lib/db/` | 🔴 |
| [T8.4](T8.4-delete-data-dir.md) | Delete dataDir.js | `src/lib/dataDir.js` | 🟡 |
| [T8.5](T8.5-remove-sqlite-package.md) | Remove SQLite from package.json | `package.json`, lockfile | 🟠 |
| [T8.6](T8.6-sqlite-driver-search.md) | SQLite driver search | grep | 🟠 |
| [T8.7](T8.7-build-smoke-sqlite-sensitive.md) | Build + smoke (SQLite-sensitive) | smoke | 🔴 |
| [T8.8](T8.8-lint-rule-sqlite.md) | Lint rule for SQLite | eslint | 🟠 |

---

## Removal Targets

```
src/lib/usageDb.js
src/lib/localDb.js
src/lib/disabledModelsDb.js
src/lib/requestDetailsDb.js
src/lib/dataDir.js
src/lib/db/
better-sqlite3 (package)
sqlite3 (package, if present)
```

## Forbidden

- Leaving no-op stubs
- Adding localStorage fallback
- Migrating SQLite to IndexedDB
- Recreating under different name

## Exit Criteria

- [ ] All 8 removal targets deleted
- [ ] SQLite driver gone from lockfile
- [ ] `rg` confirms zero residual imports
- [ ] Build + full smoke green

## Note

This is the **PRIMARY GOAL** of the migration. 🎯
