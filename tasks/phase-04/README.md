# Phase 4 — Cohort B Route Audit & Proxy

**Objective**: Convert the 9 audit-heavy routes. Each requires individual PM/backend decision.

**Branch**: `phase/4-cohort-b-proxy`

**Dependencies**: Phase 3 + T0.9 PM decisions

**Exit Gate**: Each Cohort B route is either proxied, deleted, or explicitly marked out-of-scope.

---

## Tasks

| ID | Task | Route | Risk |
|----|------|-------|------|
| [T4.1](T4.1-init-route.md) | init/ decision + execution | `api/init/*` | 🟡 |
| [T4.2](T4.2-shutdown-route.md) | shutdown/ decision + execution | `api/shutdown/*` | 🟡 |
| [T4.3](T4.3-locale-route.md) | locale/ decision + execution | `api/locale/*` | 🟡 |
| [T4.4](T4.4-cli-tools-route.md) | cli-tools/ decision + execution | `api/cli-tools/*` | 🟠 |
| [T4.5](T4.5-cloud-route.md) | cloud/ decision + execution | `api/cloud/*` | 🟠 |
| [T4.6](T4.6-tunnel-route.md) | tunnel/ decision + execution | `api/tunnel/*` | 🟠 |
| [T4.7](T4.7-translator-route.md) | translator/ decision + execution | `api/translator/*` | 🟠 |
| [T4.8](T4.8-v1beta-proxy-streaming.md) | v1beta/ proxy with streaming | `api/v1beta/*` | 🔴 |
| [T4.9](T4.9-v1-proxy-streaming.md) | v1/ proxy with streaming | `api/v1/*` | 🔴 |
| [T4.10](T4.10-har-replay-verification.md) | HAR replay verification | fixtures | 🔴 |
| [T4.11](T4.11-streaming-verification-matrix.md) | Streaming verification matrix | test matrix | 🔴 |

---

## Cohort B Routes

```
v1/          → PROXY (0penAI-compatible, highest blast radius)
v1beta/      → PROXY (Gemini-compatible)
translator/  → DELETE (backend has internal/translator/)
init/        → DELETE (backend manages lifecycle)
shutdown/    → DELETE (backend manages lifecycle)
cloud/       → TBD (depends on PM decision)
cli-tools/   → TBD (depends on PM decision)
tunnel/      → TBD (depends on PM decision)
locale/      → KEEP-AS-IS (frontend-only i18n) or DELETE
```

## Decision Tree

For each route:
1. Capture HAR + read route file
2. Decide: **PROXY** / **DELETE** / **KEEP-AS-IS**
3. If PROXY: apply proxy pattern
4. If DELETE: verify zero consumers, remove
5. Validate

## Forbidden

- Improving response formats
- Buffering streaming responses
- Adding retry logic

## Exit Criteria

- [ ] All 9 Cohort B routes have decision recorded
- [ ] All PROXY routes pass HAR + streaming verification
- [ ] All DELETE routes have zero consumers
- [ ] Build + smoke + streaming green
