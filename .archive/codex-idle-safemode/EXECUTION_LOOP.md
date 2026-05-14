# AIPROXY Migration Execution Loop

> **Purpose**: Bounded execution protocol for AI agents
> **Audience**: AI executors, not humans
> **Status**: OPERATIONAL

---

# CRITICAL RULE

```
EXECUTE ONE TASK. STOP. AWAIT INSTRUCTION.
```

Do NOT:
- Chain multiple tasks
- "Continue" to next task
- Fix unrelated issues
- Refactor adjacent code
- Improve architecture
- Modernize patterns

---

# EXECUTION PROTOCOL

## Step 1: LOAD CONTEXT

Read files in this order:
1. `/tasks/README.md` — Phase index
2. `/tasks/phase-XX/README.md` — Current phase overview
3. `/tasks/phase-XX/T<phase>.<seq>.md` — ONE task file only

Maximum context load: 3 files

---

## Step 2: VERIFY PREREQUISITES

Check task file section: `## Prerequisites`

If ANY prerequisite not met:
- STOP
- Report blocker
- Await instruction

---

## Step 3: EXECUTE TASK

Follow steps in: `## Execution`

Do NOT:
- Skip steps
- Add steps
- Modify steps
- Interpret steps loosely

---

## Step 4: RUN TARGETED VERIFICATION

Run ONLY what the task specifies in `## Verification`

Typical verification:
```bash
tsc --noEmit           # Type check
npm run build          # Build check
rg -n "<pattern>" src/ # Import check
```

If build fails:
1. Read FIRST root error only
2. Fix ONLY that error
3. Re-run verification
4. Do NOT fix cascade errors blindly

---

## Step 5: UPDATE STATUS

Update `/tasks/TASK_STATUS.md`:
```
[T<phase>.<seq>] DONE
```

---

## Step 6: COMMIT

```bash
git add -A
git commit -m "<type>: <message> (T<phase>.<seq>)"
```

Commit type:
- `feat:` — New functionality
- `refactor:` — Code restructure
- `fix:` — Bug fix
- `docs:` — Documentation
- `chore:` — Maintenance
- `test:` — Test changes

---

## Step 7: REPORT

Output format:
```
✓ COMPLETED: T<phase>.<seq> — <task title>

CHANGES:
- <file 1>
- <file 2>

VERIFICATION:
- tsc: PASS
- build: PASS
- <specific check>: PASS

STATUS: READY FOR NEXT TASK
```

---

## Step 8: STOP

Do NOT continue to next task.
Do NOT "clean up" surrounding code.
Do NOT suggest improvements.

AWAIT USER INSTRUCTION.

---

# FORBIDDEN ACTIONS

During execution, NEVER:

| Action | Why Forbidden |
|--------|---------------|
| Chain tasks | Context drift, error cascade |
| Fix unrelated errors | Scope creep |
| Refactor adjacent code | Regression risk |
| Modernize patterns | Architecture drift |
| Delete shim prematurely | Dependency break |
| Mass rename files | Import cascade |
| Skip verification | Hidden regression |
| Skip commit | Lost checkpoint |
| "Improve" code | Endless refactor syndrome |

---

# ERROR HANDLING

## Build Error

1. Read FIRST error only
2. Check if error is in files you modified
3. If YES: Fix YOUR change
4. If NO: STOP, report unrelated error
5. Do NOT fix errors in unrelated files

## Type Error

1. Verify imports are correct
2. Check if interface changed
3. Update ONLY affected code
4. Do NOT add `as any`, `@ts-ignore`, `@ts-expect-error`

## Test Failure

1. Check if test covers your change
2. If YES: Fix YOUR code
3. If NO: STOP, report unrelated failure

## Blocked

If task cannot proceed:
```
⚠ BLOCKED: T<phase>.<seq>

BLOCKER: <specific reason>
REQUIRED: <what needs to happen>

AWAITING: <user decision / backend fix / external dependency>
```

---

# VERIFICATION GATES

After EVERY task:

```bash
# Type check (fast)
tsc --noEmit

# If task modified source files
npm run build

# If task modified imports
rg -n "<forbidden-pattern>" src/
```

---

# CONTEXT RESET

If you feel:
- Unclear about current task
- Tempted to "improve" code
- Unsure about scope

STOP. Re-read:
1. Task file
2. This execution loop
3. Project constraints

---

# PHASE EXIT CRITERIA

Before declaring phase complete:

1. All tasks in phase README marked DONE
2. Exit criteria in phase README verified
3. No blocking issues
4. Build green
5. Smoke test green

Report:
```
✓ PHASE <N> COMPLETE

VERIFIED:
- [x] All tasks done
- [x] Exit criteria met
- [x] Build green
- [x] Smoke green

READY FOR PHASE <N+1>
```

---

# SPECIALIST DELEGATION

Delegation is ALLOWED for bounded subproblems. See `AGENT_RULES.md` for full rules.

## Allowed Specialists

| Type | Scope |
|------|-------|
| Build/debug | Root-cause analysis only |
| Import tracing | Grep results only |
| TypeScript type | Type fixes only |
| API contract | Shape verification only |

## Forbidden Specialists

| Type | Reason |
|------|--------|
| Streaming redesign | Architecture change |
| Architecture | Scope creep |
| Refactor | Drift |
| Optimization | Drift |
| Modernization | Drift |

## Delegation Rules

```
MAX DEPTH: 1 (no nested delegation)

SPECIALIST MUST:
- Stay inside task scope
- Not modify unrelated systems
- Not expand task scope
- Not continue to next task

MAIN EXECUTOR MUST:
- Review all specialist output
- Make final decisions
- Run verification
- Commit
```

See `AGENT_RULES.md` for complete delegation protocol.

---

# ARCHITECTURE LOCK

After Phase 11, these patterns are FROZEN:

```
FORBIDDEN:
- better-sqlite3, sqlite3
- @/lib/db, @/lib/usageDb, @/lib/localDb
- open-sse
- fs imports in client code
```

Any violation must be:
1. BLOCKED
2. REPORTED
3. REVERTED if already committed
