# AIPROXY Task Execution Template

> **Purpose**: Bounded execution manifest for AI agents
> **Usage**: Create one per task before execution
> **Status**: TEMPLATE

---

# TASK MANIFEST

## Current Task

| Field | Value |
|-------|-------|
| **Task ID** | `<T<phase>.<seq>>` |
| **Task Title** | `<title from task file>` |
| **Phase** | `<phase number>` |
| **Risk Level** | `<risk from task file>` |

---

# EXECUTION BOUNDARIES

## Allowed Files

Only modify files listed below:

```
<file pattern 1>
<file pattern 2>
<file pattern 3>
```

## Forbidden Files

DO NOT touch files matching these patterns:

```
<forbidden pattern 1>
<forbidden pattern 2>
<forbidden pattern 3>
```

## Allowed Imports

Only import from:

```
<allowed import path 1>
<allowed import path 2>
```

## Forbidden Imports

DO NOT import from:

```
<forbidden import 1>
<forbidden import 2>
```

---

# VERIFICATION REQUIREMENTS

## Required Commands

Run these commands after changes:

```bash
<command 1>  # <purpose>
<command 2>  # <purpose>
```

## Required Checks

- [ ] `<check 1>`
- [ ] `<check 2>`
- [ ] `<check 3>`

## Parity Verification

| Feature | Verify Against |
|---------|----------------|
| `<feature 1>` | 9router behavior |
| `<feature 2>` | 9router behavior |

---

# STOP CONDITIONS

## Success Condition

```
Task completed when:
- <condition 1>
- <condition 2>
- All verification passes
```

## Block Condition

```
Stop and report if:
- <block condition 1>
- <block condition 2>
```

---

# OUTPUT REQUIREMENTS

## Files Modified

| File | Change Type |
|------|-------------|
| `<file 1>` | `<created/modified/deleted>` |
| `<file 2>` | `<created/modified/deleted>` |

## Commit Message Format

```
<type>: <description> (T<phase>.<seq>)

<details if needed>
```

---

# CONTEXT LOAD LIMIT

Maximum files to read during execution:

```
1. EXECUTION_LOOP.md
2. BUILD_RULES.md
3. tasks/phase-XX/T<phase>.<seq>.md
4. <allowed files only>
```

Do NOT read:
- Entire codebase
- Unrelated modules
- Other task files
- Planning documents

---

# EXAMPLE FILLED TEMPLATE

## Current Task

| Field | Value |
|-------|-------|
| **Task ID** | T2.5 |
| **Task Title** | Create Usage API Client |
| **Phase** | 2 |
| **Risk Level** | 🟠 High |

## Allowed Files

```
src/lib/usage-api.ts
src/lib/http.ts
src/shared/contracts/*.ts
```

## Forbidden Files

```
src/lib/usageDb.js
src/lib/localDb.js
src/lib/db/*
src/app/api/*
src/components/*
```

## Allowed Imports

```
@/lib/http
@/shared/contracts
```

## Forbidden Imports

```
@/lib/usageDb
@/lib/localDb
@/lib/db
```

## Required Commands

```bash
tsc --noEmit           # Type check
npm run build          # Build verification
rg -n "usageDb" src/   # Confirm no usageDb imports in new file
```

## Required Checks

- [ ] File created at `src/lib/usage-api.ts`
- [ ] All functions from T2.1 inventory implemented
- [ ] No forbidden imports
- [ ] Types match backend contracts
- [ ] Build passes

## Parity Verification

| Feature | Verify Against |
|---------|----------------|
| Usage API functions | Backend endpoint shapes |
| Response types | API_CONTRACT.md |

## Success Condition

```
Task completed when:
- usage-api.ts created
- All functions implemented
- Build passes
- No forbidden imports
```

## Block Condition

```
Stop and report if:
- Backend endpoint missing (check T0.8)
- Type mismatch with API_CONTRACT.md
- Circular dependency detected
```

## Files Modified

| File | Change Type |
|------|-------------|
| `src/lib/usage-api.ts` | created |

## Commit Message Format

```
feat: add usage-api client (T2.5)

Backend-driven usage API client replacing usageDb.
Functions: getUsageHistory, getUsageStats, getChartData,
getActiveRequests, getRecentLogs, getRequestDetailById.
```

---

# HOW TO USE

1. Copy this template
2. Fill in task-specific values
3. Present to AI executor
4. AI executes within boundaries
5. AI stops when condition met
6. Review output
7. Give next task

This enforces:
- File-level boundaries
- Import-level boundaries
- Verification requirements
- Clear stop conditions
