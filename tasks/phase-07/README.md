# Phase 7 — Streaming Normalization (open-sse Retirement)

**Objective**: Retire `frontend/open-sse/` entirely. Backend owns stream normalization. Frontend renders only.

**Branch**: `phase/7-streaming-normalization`

**Dependencies**: Phase 4

**Exit Gate**: `frontend/open-sse/` deleted; streaming test matrix passes 9router parity.

---

## Tasks

### Sub-phase 7a — Backend Contract Verification

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T7.1](T7.1-stream-contract-docs.md) | Stream contract documentation | `docs/STREAM_CONTRACT.md` | 🔴 |
| [T7.2](T7.2-event-types-verification.md) | Event types verification | docs | 🔴 |
| [T7.3](T7.3-provider-format-leaks.md) | Provider format leak check | tickets | 🔴 |

### Sub-phase 7b — Frontend SSE Consumer

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T7.4](T7.4-create-sse-consumer.md) | Create SSE consumer | `src/sse/consumer.ts` | 🔴 |
| [T7.5](T7.5-sse-consumer-tests.md) | SSE consumer unit tests | `src/sse/consumer.test.ts` | 🟠 |
| [T7.6](T7.6-slim-sse-handlers.md) | Slim SSE handlers | `src/sse/*` | 🟠 |

### Sub-phase 7c — UI Migration

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T7.7](T7.7-stream-consumers-inventory.md) | Stream consumers inventory | `inventory/stream-consumers.md` | 🟡 |
| [T7.8](T7.8-chat-ui-migration.md) | Chat UI migration | chat components | 🔴 |
| [T7.9](T7.9-stream-feature-flag.md) | Stream feature flag | env | 🟠 |
| [T7.10](T7.10-streaming-parity-test.md) | Streaming parity test | regression report | 🔴 |

### Sub-phase 7d — Cleanup

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T7.11](T7.11-delete-open-sse.md) | Delete open-sse directory | `open-sse/` | 🔴 |
| [T7.12](T7.12-zero-imports-verification.md) | Zero imports verification | grep | 🟡 |
| [T7.13](T7.13-remove-feature-flag.md) | Remove stream feature flag | env | 🟢 |
| [T7.14](T7.14-lint-rule-open-sse.md) | Lint rule for open-sse | eslint | 🟡 |

---

## Critical Context

`frontend/open-sse/` contains:
- Full 9router streaming engine
- Executors, translator, RTK, transformer
- **Highest-risk architectural violation**

## Forbidden

- Adding provider-format parsing in frontend
- Reordering chunks
- Adding frontend retry logic
- Throttling chunks

## Exit Criteria

- [ ] `frontend/open-sse/` deleted
- [ ] Streaming test matrix passes 9router parity
- [ ] Single frontend stream consumer in `src/sse/consumer.ts`
- [ ] Build + smoke + streaming green
