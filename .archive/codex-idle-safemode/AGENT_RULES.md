# AIPROXY Agent Rules

> **Purpose**: Boundaries for AI agents during migration execution
> **Status**: OPERATIONAL
> **Applies to**: All AI coding agents (OpenCode, CL4ude, GPT, Qwen, etc.)

---

# EXECUTION PHILOSOPHY

```
Migration ≠ Refactoring
Migration ≠ Modernization  
Migration ≠ Optimization

Migration = Behavior-preserving architecture change
```

---

# CRITICAL RULES

## Rule 1: Task Boundary is Absolute

```
EXECUTE ONE TASK. STOP. AWAIT INSTRUCTION.
```

Never:
- Chain tasks
- "Continue" after completion
- Expand task scope
- "Improve" adjacent code

## Rule 2: Behavior Parity Required

```
Preserve 9router behavior EXACTLY.
Architecture parity NOT required.
```

Before changing ANY behavior:
1. Find 9router reference implementation
2. Document original behavior
3. Preserve it verbatim
4. Only change: data source (SQLite → Backend API)

## Rule 3: No Invention

```
Do not invent:
- New abstractions
- New patterns
- New architectures
- Compatibility shims
- Temporary hacks
```

## Rule 4: Architecture Lock

After Phase 11, these are PERMANENTLY FORBIDDEN:

```
- better-sqlite3, sqlite3
- @/lib/db, @/lib/usageDb, @/lib/localDb
- open-sse
- fs imports in client code
- localStorage persistence (except token)
```

---

# SPECIALIST DELEGATION RULES

## Allowed Delegation

Specialists are ALLOWED for isolated subproblems:

| Specialist Type | Allowed? | Scope |
|-----------------|----------|-------|
| Build/debug specialist | ✅ | Root-cause analysis only |
| Import tracing specialist | ✅ | Find imports only |
| TypeScript type specialist | ✅ | Fix type errors only |
| API contract verifier | ✅ | Verify shapes only |
| Streaming redesign | ❌ | FORBIDDEN |
| Architecture specialist | ❌ | FORBIDDEN |
| Refactor specialist | ❌ | FORBIDDEN |
| Optimization specialist | ❌ | FORBIDDEN |
| Modernization specialist | ❌ | FORBIDDEN |

## Delegation Constraints

When delegating to a specialist:

```
ALLOWED ONLY IF:
1. Delegation remains inside current task scope
2. Specialist cannot expand task scope
3. Specialist cannot modify unrelated systems
4. Specialist cannot continue to next tasks
5. Specialist follows all migration constraints
```

## Maximum Delegation Depth

```
MAX DEPTH: 1

Main executor → Specialist (STOP)

NEVER:
Main → Specialist → Sub-specialist → ...
```

Nested delegation = Context drift = Failure

## Main Executor Responsibility

Main executor ALWAYS remains responsible for:

```
- Task boundary enforcement
- Parity preservation
- Final verification
- Commit decision
- Status update
```

Specialist provides analysis.
Main executor makes decisions.

---

# SPECIALIST WORKFLOW

## Correct Usage

```
Main Executor:
  "I need to trace all imports of usageDb to understand dependency chain.
   Delegating to import-tracing specialist."

Specialist (bounded):
  - Runs: rg -n "usageDb" src/
  - Returns: List of files and line numbers
  - STOPS

Main Executor:
  - Receives list
  - Makes decisions about migration
  - Executes changes
  - Verifies
  - Commits
```

## Incorrect Usage

```
Main Executor:
  "This streaming code is complex. Handing to streaming specialist."

Specialist (unbounded):
  - Redesigns stream layer
  - Adds new abstractions
  - Changes behavior
  - 💀
```

---

# FORBIDDEN SPECIALIST ACTIONS

During delegation, specialists MUST NOT:

| Action | Why Forbidden |
|--------|---------------|
| Expand task scope | Drift |
| Modify unrelated systems | Regression |
| Add new abstractions | Complexity |
| Change architecture | Scope creep |
| "Improve" code | Endless refactor |
| Redesign systems | Parity break |
| Skip verification | Hidden bugs |

---

# DELEGATION DECISION TREE

```
Need help with:
│
├─ Build error root cause?
│  └─ YES → Delegate to build specialist (bounded to error analysis)
│  └─ NO → Continue
│
├─ Import tracing?
│  └─ YES → Delegate to tracing specialist (bounded to grep results)
│  └─ NO → Continue
│
├─ Type error?
│  └─ YES → Delegate to TS specialist (bounded to type fix only)
│  └─ NO → Continue
│
├─ Need to redesign/restructure?
│  └─ STOP → Not allowed without explicit user request
│
└─ Need to optimize/modernize?
   └─ STOP → Not allowed during migration
```

---

# VERIFICATION REQUIREMENTS

After ANY specialist involvement:

```
1. Main executor reviews ALL specialist output
2. Main executor verifies against task boundaries
3. Main executor runs verification
4. Main executor commits
5. Main executor updates status
```

Never let specialist:
- Commit directly
- Update status
- Continue to next task

---

# EXAMPLES

## Good Delegation

```
Task: T2.5 — Create usage-api.ts

Main Executor encounters complex type inference.

Delegates: "TS specialist, infer correct types for these 3 functions.
Boundaries: Only usage-api.ts, only type annotations, no logic changes."

Specialist returns:
- Fixed type annotations
- Explanation

Main Executor:
- Reviews
- Verifies build passes
- Commits
- Updates status
- STOPS
```

## Bad Delegation

```
Task: T4.9 — v1 proxy with streaming

Main Executor sees streaming code.

Delegates: "Streaming specialist, handle v1 streaming proxy."

Specialist:
- Redesigns stream consumer
- Adds new abstractions
- Changes behavior
- 💀 MIGRATION FAILS
```

---

# CONTEXT ISOLATION

When delegating, provide specialist with:

```
1. Current task file path
2. Specific question/subproblem
3. Relevant file paths (not entire codebase)
4. Constraints to follow
```

Do NOT provide:
- Entire codebase context
- Unrelated planning documents
- Other task files
- Freedom to expand scope

---

# ERROR HANDLING

If specialist:
- Expands scope → STOP, report, discard work
- Modifies forbidden files → STOP, report, discard work  
- Fails verification → Main executor fixes, not specialist

---

# SUMMARY: NEVER / ALWAYS

## NEVER

```
- Never continue to next task automatically
- Never refactor unrelated systems
- Never modernize architecture
- Never delete systems before isolation
- Never fix unrelated build errors
- Never chain specialists
- Never let specialist expand scope
- Never add type suppressions (as any, @ts-ignore)
```

## ALWAYS

```
- Read first root error only
- Stop after task completion
- Await human approval before next task
- Verify boundaries after specialist work
- Keep delegation depth ≤ 1
- Review all specialist output before committing
```

---

# SIGN-OFF

By executing any task, AI agent acknowledges:

```
I understand:
- Task boundaries are absolute
- Delegation must be constrained
- Behavior parity is required
- I am responsible for final verification
- I will STOP after one task
```

---

# ENFORCEMENT

Violations detected by:
- Code review
- Build failures
- Parity failures
- Architecture regressions

Response:
- Immediate halt
- Rollback to last checkpoint
- User notification
- No "continue" attempts
