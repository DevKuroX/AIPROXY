# Execution Manifests Index

Execution manifests provide **bounded execution boundaries** for AI agents.

---

# What is an Execution Manifest?

An execution manifest defines:

```
- Allowed files (can modify)
- Forbidden files (cannot touch)
- Required verification (must run)
- Stop conditions (when to stop)
```

---

# Available Manifests

| Task | Risk | Description |
|------|------|-------------|
| [T0.1](T0.1-manifest.md) | 🟢 Low | SQLite import inventory (read-only) |
| [T4.9](T4.9-manifest.md) | 🔴 Critical | v1 streaming proxy (highest blast radius) |
| [T7.11](T7.11-manifest.md) | 🔴 Critical | Delete open-sse directory |
| [T8.3](T8.3-manifest.md) | 🔴 Critical | Delete SQLite DB (PRIMARY GOAL) |

---

# How to Use

## For User

```
1. Open manifest for task
2. Review boundaries
3. Present to AI: "Execute T4.9 with manifest"
4. AI executes within boundaries
5. AI stops at condition
```

## For AI Executor

```
1. Read manifest
2. Read task file
3. Verify preconditions
4. Modify ONLY allowed files
5. Run required verification
6. Check stop conditions
7. Report result
8. STOP
```

---

# Creating New Manifests

Use template from `/TASK_EXECUTION_TEMPLATE.md`

For each critical task:
1. Copy template
2. Fill in task-specific values
3. Define file boundaries
4. Define verification requirements
5. Define stop conditions

---

# Enforcement Rules

During execution:

| Violation | Action |
|-----------|--------|
| Modify forbidden file | BLOCK, report, revert |
| Skip verification | BLOCK, force verification |
| Continue past stop | BLOCK, report |
| Missing manifest | CREATE before execution |

---

# Relationship to EXECUTION_LOOP.md

```
EXECUTION_LOOP.md      → Protocol (how to execute)
TASK_EXECUTION_TEMPLATE.md → Template (how to create manifest)
execution-manifests/   → Instances (specific task boundaries)
```

The manifest is the **contract** for each task execution.
