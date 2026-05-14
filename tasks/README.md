# AIPROXY Frontend Migration — Atomic Execution Tasks

> **Generated**: 2026-05-14
> **Source**: `FRONTEND_MIGRATION_TASK_INDEX.md`, `FRONTEND_REFACTOR_EXECUTION_PLAN.md`
> **Authority**: `PROJECT_CONSTRAINTS.md`, `ARCHITECTURE.md`, `API_CONTRACT.md`
> **Behavior Reference**: 9router
> **Architecture Target**: AIPROXY thin client

---

## Overview

This directory contains **atomic execution tasks** for the AIPROXY frontend migration.

Each task file is:
- **Deterministic** — specific inputs, specific outputs
- **Dependency-safe** — explicit prerequisites
- **Bounded** — single responsibility
- **Reviewable** — clear verification criteria
- **Execution-ready** — can be executed by an AI agent or human

---

## Phase Index

| Phase | Objective | Tasks | Branch |
|-------|-----------|-------|--------|
| [Phase 0](phase-00/README.md) | Pre-Migration Inventory | 10 | `phase/0-inventory` |
| [Phase 1](phase-01/README.md) | Shared Contracts Stabilization | 8 | `phase/1-shared-contracts` |
| [Phase 2](phase-02/README.md) | SQLite Dependency Isolation | 17 | `phase/2-sqlite-isolation` |
| [Phase 3](phase-03/README.md) | Cohort A Route Proxy (15 routes) | 9 | `phase/3-cohort-a-proxy` |
| [Phase 4](phase-04/README.md) | Cohort B Route Audit (9 routes) | 11 | `phase/4-cohort-b-proxy` |
| [Phase 5](phase-05/README.md) | Auth + Token Unification | 10 | `phase/5-auth-unification` |
| [Phase 6](phase-06/README.md) | Legacy Module Deprecation | 10 | `phase/6-legacy-deprecation` |
| [Phase 7](phase-07/README.md) | Streaming Normalization | 14 | `phase/7-streaming-normalization` |
| [Phase 8](phase-08/README.md) | SQLite Engine Removal | 8 | `phase/8-sqlite-removal` |
| [Phase 9](phase-09/README.md) | Dual App Shell Collapse | 7 | `phase/9-shell-collapse` |
| [Phase 10](phase-10/README.md) | Dead Code Cleanup | 7 | `phase/10-cleanup` |
| [Phase 11](phase-11/README.md) | Architecture Lock | 9 | `phase/11-architecture-lock` |

**Total Tasks**: 120

---

## Dependency Graph

```
P0 ──► P1 ──► P2 ──► P3 ──► P4 ──┐
                  └──► P5 ────────┤
                                  ├──► P6 ──► P7 ──► P8 ──► P9 ──► P10 ──► P11
```

---

## Critical Path

```
P0.9 (PM Decisions) ──► P3, P4, P5, P6
P0.6 (Backend URL) ──► P2.2, T3.0
P2 (Replacement Clients) ──► P3, P6
P4.9 (v1 Streaming) ──► P7
P8.3 (SQLite Deletion) ──► Architecture Goal
```

---

## Task File Format

Each task file follows this structure:

```markdown
# T<phase>.<seq> — Task Title

## Task Identity
- Task ID, Phase, Type, Risk level

## Objective
- What this task accomplishes

## Source Documents
- Reference to planning docs

## Prerequisites
- What must be complete before this task

## Execution
- Step-by-step instructions

## Output Artifacts
- Files produced/modified

## Verification
- Checklist for completion

## Dependencies
- What this task depends on

## Blocks
- What depends on this task

## Notes
- Additional context
```

---

## Conventions

- **Task ID format**: `T<phase>.<seq>` (e.g., `T3.4` = Phase 3, task 4)
- **Status legend**: `[ ]` not started · `[~]` in progress · `[x]` done · `[!]` blocked · `[-]` skipped
- **Risk tags**: 🟢 Low · 🟡 Medium · 🟠 High · 🔴 Critical
- **Commit policy**: One task = one commit unless explicitly noted
- **Branch policy**: One branch per phase; PRs squash-merged

---

## Risk Summary

| Phase | Critical Tasks | High Risk | Notes |
|-------|----------------|-----------|-------|
| P0 | T0.9 (PM Decisions) | T0.6, T0.8 | Blocks all downstream |
| P3 | None | T3.0, T3.3-T3.5 | Streaming critical |
| P4 | T4.8, T4.9, T4.10, T4.11 | Multiple | v1/v1beta highest blast radius |
| P5 | T5.9 | T5.3, T5.4, T5.6, T5.7 | Auth regression gate |
| P7 | T7.1-T7.4, T7.8, T7.10, T7.11 | Multiple | open-sse removal |
| P8 | T8.1, T8.3, T8.7 | T8.2, T8.5, T8.8 | SQLite deletion |
| P11 | T11.6, T11.8 | T11.4, T11.7 | Final lock |

---

## Execution Notes

1. **Read task file completely before starting**
2. **Verify prerequisites are met**
3. **Execute steps in order**
4. **Run verification checklist before marking complete**
5. **Commit immediately after verification passes**

---

## Related Documents

- `PROJECT_CONSTRAINTS.md` — Hard rules for AI agents
- `ARCHITECTURE.md` — Target architecture
- `API_CONTRACT.md` — Backend API contracts
- `FEATURE_PARITY.md` — Feature verification checklist
- `FRONTEND_REFACTOR_TRACKER.md` — Migration status tracker
- `REF_IMPLEMENTATION_RULES.md` — 9router reference rules
