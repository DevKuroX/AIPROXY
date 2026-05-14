# Phase 10 — Dead Code & Dependency Cleanup

**Objective**: Remove orphan imports, unused npm packages, ad-hoc test scripts, empty directories.

**Branch**: `phase/10-cleanup`

**Dependencies**: Phase 1-9

**Exit Gate**: `depcheck` clean; `ts-prune` clean; bundle size measurably smaller.

---

## Tasks

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T10.1](T10.1-dependency-cleanup.md) | Dependency cleanup (depcheck) | `package.json` | 🟡 |
| [T10.2](T10.2-unused-exports-cleanup.md) | Unused exports cleanup (ts-prune) | source tree | 🟡 |
| [T10.3](T10.3-relocate-test-scripts.md) | Relocate test scripts | scripts | 🟢 |
| [T10.4](T10.4-delete-jsconfig.md) | Delete jsconfig.json | config | 🟡 |
| [T10.5](T10.5-i18n-orphan-cleanup.md) | i18n orphan cleanup | `src/i18n` | 🟢 |
| [T10.6](T10.6-audit-claude-md.md) | Audit CLAUDE.md | doc | 🟢 |
| [T10.7](T10.7-bundle-size-measurement.md) | Bundle size measurement | metrics | 🟢 |

---

## Cleanup Targets

- Unused npm dependencies
- Unused TypeScript exports
- Ad-hoc test scripts:
  - `login-test.js`
  - `test-api-me.js`
  - `test-dashboard.js`
  - `test-detailed.js`
  - `serve-static.js`
- Orphan `jsconfig.json`
- Empty directories

## Forbidden

- Removing dependencies you cannot trace
- Combining cleanup with refactor

## Exit Criteria

- [ ] `depcheck` clean (or exceptions documented)
- [ ] `ts-prune` clean (or exceptions documented)
- [ ] Test scripts relocated
- [ ] Bundle size delta documented
