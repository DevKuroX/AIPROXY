# Dispatcher Prompt Template

> **Purpose**: Short prompt for task dispatching
> **Usage**: One prompt = one task
> **For**: OpenCode, Qwen, DeepSeek, any AI executor

---

# DISPATCHER PROMPT

```
Read and follow strictly:

- EXECUTION_LOOP.md
- BUILD_RULES.md
- AGENT_RULES.md
- PROJECT_CONSTRAINTS.md
- TASK_STATUS.md

Then execute ONLY this task:

{{TASK_PATH}}

Rules:
- Do NOT continue to another task
- Do NOT modify unrelated files
- Do NOT redesign architecture
- Do NOT fix unrelated build errors
- Read FIRST root build error only
- Preserve 9router behavior parity
- Stop immediately after task completion
- Maximum delegation depth: 1

Required output:
1. Files modified
2. Changes made
3. Verification results
4. TASK_STATUS.md update
5. Any blockers

STOP after task execution.
```

---

# USAGE

## Correct

```
Execute:

tasks/phase-00/T0.1-sqlite-import-inventory.md
```

```
Execute:

tasks/phase-02/T2.5-create-usage-api-client.md
```

```
Execute:

tasks/phase-04/T4.9-v1-proxy-streaming.md
```

## Wrong

```
Execute phase 2                    ❌ Too broad
Execute all inventory tasks        ❌ Chaining
Continue to next task              ❌ Forbidden
```

---

# EXAMPLE SESSION

## User Input

```
Read and follow strictly:

- EXECUTION_LOOP.md
- BUILD_RULES.md
- AGENT_RULES.md
- PROJECT_CONSTRAINTS.md
- TASK_STATUS.md

Then execute ONLY this task:

tasks/phase-00/T0.1-sqlite-import-inventory.md

Rules:
- Do NOT continue to another task
- Do NOT modify unrelated files
- Do NOT redesign architecture
- Do NOT fix unrelated build errors
- Read FIRST root build error only
- Preserve 9router behavior parity
- Stop immediately after task completion
- Maximum delegation depth: 1

Required output:
1. Files modified
2. Changes made
3. Verification results
4. TASK_STATUS.md update
5. Any blockers

STOP after task execution.
```

## AI Output

```
✓ COMPLETED: T0.1 — SQLite Import Inventory

FILES MODIFIED:
- docs/migration/inventory/sqlite-importers.txt (created)

CHANGES MADE:
- Created inventory directory
- Ran: rg -n "from ['\"]@/lib/(usageDb|localDb|disabledModelsDb|requestDetailsDb|db/)" frontend/src/
- Captured output to sqlite-importers.txt

VERIFICATION:
- File exists: PASS
- Contains grep output: PASS
- No source modified: PASS

TASK_STATUS.md UPDATE:
- T0.1: [ ] → [x] DONE

BLOCKERS: None

STATUS: READY FOR NEXT TASK
```

---

# QUICK REFERENCE

## Phase 0 — Inventory (10 tasks)

```
tasks/phase-00/T0.1-sqlite-import-inventory.md
tasks/phase-00/T0.2-legacy-import-inventory.md
tasks/phase-00/T0.3-stream-import-inventory.md
tasks/phase-00/T0.4-filesystem-import-inventory.md
tasks/phase-00/T0.5-route-cohort-classification.md
tasks/phase-00/T0.6-backend-url-decision.md
tasks/phase-00/T0.7-api-contract-update.md
tasks/phase-00/T0.8-backend-coverage-verification.md
tasks/phase-00/T0.9-pm-decisions-collection.md
tasks/phase-00/T0.10-har-fixtures-collection.md
```

## Phase 2 — Critical Path

```
tasks/phase-02/T2.2-create-http-helper.md
tasks/phase-02/T2.5-create-usage-api-client.md
tasks/phase-02/T2.16-backend-coverage-verification.md
```

## Phase 4 — Streaming (Critical)

```
tasks/phase-04/T4.8-v1beta-proxy-streaming.md
tasks/phase-04/T4.9-v1-proxy-streaming.md
tasks/phase-04/T4.11-streaming-verification-matrix.md
```

## Phase 8 — SQLite Deletion (PRIMARY GOAL)

```
tasks/phase-08/T8.3-delete-sqlite-db.md
```

---

# COPY-PASTE READY

## Start Migration (Phase 0)

```
Read and follow strictly:

- EXECUTION_LOOP.md
- BUILD_RULES.md
- AGENT_RULES.md
- PROJECT_CONSTRAINTS.md
- TASK_STATUS.md

Then execute ONLY this task:

tasks/phase-00/T0.1-sqlite-import-inventory.md

Rules:
- Do NOT continue to another task
- Do NOT modify unrelated files
- Do NOT redesign architecture
- Do NOT fix unrelated build errors
- Read FIRST root build error only
- Preserve 9router behavior parity
- Stop immediately after task completion
- Maximum delegation depth: 1

Required output:
1. Files modified
2. Changes made
3. Verification results
4. TASK_STATUS.md update
5. Any blockers

STOP after task execution.
```
