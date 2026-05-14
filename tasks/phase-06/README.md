# Phase 6 — Legacy Module Deprecation

**Objective**: Remove all consumers of legacy business logic directories and delete the modules.

**Branch**: `phase/6-legacy-deprecation`

**Dependencies**: Phase 3, Phase 4, Phase 5

**Exit Gate**: `rg` returns zero imports of legacy dirs; modules deleted.

---

## Tasks

| ID | Task | Module | Risk |
|----|------|--------|------|
| [T6.1](T6.1-deprecate-oauth-module.md) | Deprecate oauth/ | `src/lib/oauth/` | 🟠 |
| [T6.2](T6.2-deprecate-tunnel-module.md) | Deprecate tunnel/ | `src/lib/tunnel/` | 🟠 |
| [T6.3](T6.3-deprecate-updater-module.md) | Deprecate updater/ | `src/lib/updater/` | 🟡 |
| [T6.4](T6.4-deprecate-usage-module.md) | Deprecate usage/ | `src/lib/usage/` | 🟠 |
| [T6.5](T6.5-deprecate-network-module.md) | Deprecate network/ | `src/lib/network/` | 🟡 |
| [T6.6](T6.6-deprecate-mitm-module.md) | Deprecate mitm/ | `src/lib/mitm/` + `src/mitm/` | 🟠 |
| [T6.7](T6.7-deprecate-provider-normalization.md) | Deprecate providerNormalization | `src/lib/providerNormalization.js` | 🟡 |
| [T6.8](T6.8-deprecate-init-cloud-sync.md) | Deprecate initCloudSync | `src/lib/initCloudSync.js` | 🟡 |
| [T6.9](T6.9-classify-console-log-buffer.md) | Classify consoleLogBuffer | `src/lib/consoleLogBuffer.js` | 🟢 |
| [T6.10](T6.10-build-smoke-per-batch.md) | Build + smoke per batch | per-batch | 🟡 |

---

## Legacy Modules

```
src/lib/oauth/
src/lib/tunnel/
src/lib/updater/
src/lib/usage/
src/lib/network/
src/lib/mitm/
src/mitm/
src/lib/providerNormalization.js
src/lib/initCloudSync.js
src/lib/consoleLogBuffer.js
```

## Per-Module Checklist

1. From T0.2 inventory: list consumers
2. Confirm replacement client exists (T2.5–T2.13)
3. Replace import in consumer
4. Run `tsc --noEmit`
5. Verify smoke pass
6. Verify `rg` returns zero
7. Delete module
8. Commit

## Forbidden

- Mass-deleting all in one commit
- Leaving stub files behind
- Recreating under different name

## Exit Criteria

- [ ] All 9 legacy module groups deleted or kept with PM justification
- [ ] Build + smoke green after each batch
