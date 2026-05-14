# Phase 9 — Dual App Shell Collapse

**Objective**: One root layout, one root page. Choose TypeScript. Delete `.js` siblings.

**Branch**: `phase/9-shell-collapse`

**Dependencies**: Phase 1-8

**Exit Gate**: Only `.tsx` versions of `layout` and `page` exist at `src/app/`.

---

## Tasks

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T9.1](T9.1-diff-page-files.md) | Diff page.js vs page.tsx | `inventory/shell-diff.md` | 🟡 |
| [T9.2](T9.2-diff-layout-files.md) | Diff layout.js vs layout.tsx | `inventory/shell-diff.md` | 🟡 |
| [T9.3](T9.3-port-missing-behaviors.md) | Port missing behaviors | `src/app/*.tsx` | 🟠 |
| [T9.4](T9.4-delete-dual-shell-js.md) | Delete dual shell JS files | files | 🟠 |
| [T9.5](T9.5-config-audit.md) | jsconfig/tsconfig audit | configs | 🟡 |
| [T9.6](T9.6-visual-diff.md) | Visual diff | visual diff | 🟡 |
| [T9.7](T9.7-build-smoke.md) | Build + smoke | clean | 🟡 |

---

## Critical Context

Dual file extensions cause:
- Non-deterministic resolution by Next.js
- Different behavior on different builds
- Regression risk

## Files to Resolve

```
src/app/page.js + src/app/page.tsx
src/app/layout.js + src/app/layout.tsx
```

## Forbidden

- Restructuring layout tree
- Adding new routes/pages
- Updating Next.js version

## Exit Criteria

- [ ] Only `.tsx` versions exist
- [ ] Visual diff acceptable
- [ ] Build green
