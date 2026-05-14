---
name: aiproxy-autonomous-loop
description: |
  REPLACES Codex built-in planning mode for the AIPROXY frontend migration.
  Use this skill when the user wants to run a FULL, UNATTENDED execution
  loop across multiple atomic tasks in `tasks/phase-XX/` until the entire
  phase (or selected batch) is finished. Combines brainstorming +
  create-plan up front, then executes the validated plan in a continuous
  loop with verification gates after every single task.

  IMPORTANT — UNATTENDED MODE:
  This skill is designed to be left running while the user is AWAY.
  The loop NEVER stops on a single task failure. A task that cannot pass
  verification is marked `[!]` BLOCKED in TASK_STATUS.md and the loop
  proceeds to the next task. The user reviews blockers when they return.
trigger_phrases:
  - "eksekusi semua plan"
  - "kerjakan sampai selesai"
  - "jangan pernah berhenti"
  - "full loop"
  - "autonomous run"
  - "idle mode"
  - "tinggal pergi"
authority:
  - PROJECT_CONSTRAINTS.md
  - AGENT_RULES.md
  - BUILD_RULES.md
  - EXECUTION_LOOP.md
  - REF_IMPLEMENTATION_RULES.md
  - API_CONTRACT.md
  - TASK_STATUS.md
loop_override:
  rule_overridden: "EXECUTE ONE TASK. STOP. AWAIT INSTRUCTION."
  replaced_with:  "EXECUTE ONE TASK. VERIFY. COMMIT OR MARK BLOCKED. CONTINUE."
  preserved:
    - One task = one commit (no chaining commits)
    - No creative refactor (BUILD_RULES intact)
    - Read FIRST root error only (BUILD_RULES.md Step 1)
    - No modification of files outside Allowed Files in manifest
    - 9router behavioral parity required
  stop_conditions:
    # Only these end the loop. Everything else = continue.
    - All tasks in the active plan are processed (either `[x]` or `[!]`)
    - User issues an explicit interrupt
    - Critical system failure (no disk space, no network, repo locked)
  NEVER_STOP_ON:
    - Single task verification failure → mark `[!]`, continue
    - Multiple consecutive task failures → mark all `[!]`, continue
    - Missing dependency between tasks → skip dependent, continue with independent
    - Build error that cannot be fixed → revert task changes, mark `[!]`, continue
    - Same root error after N attempts → revert, mark `[!]`, continue
---

# AIPROXY Autonomous Loop Skill (Unattended Mode)

## Project Context (read before doing anything)

AIPROXY is NOT a 1:1 port of 9router.

- **9router** = fullstack JS (Node + Next.js, SQLite, business logic in
  frontend layer). Behavioral reference only.
- **AIPROXY backend** = Go service on `http://localhost:1432`. Modular,
  feature-complete. **OUT OF SCOPE for this skill.**
- **AIPROXY frontend** = Next.js 16 + React 19, currently mid-refactor
  from fullstack-hybrid (9router-style) → thin client (calls Go backend
  via `frontend/src/app/api/*` proxy routes per `API_CONTRACT.md`).
  **THIS is the only scope this skill touches.**

What this means in practice for failures:

- Errors will overwhelmingly come from `npm run build` / `tsc --noEmit`
  during the frontend refactor, NOT from runtime / backend / network.
- The refactor is **destructive**: it deletes SQLite shims (`usageDb.js`,
  `localDb.js`, `lib/db/`), dual app shells (`page.js`+`page.tsx`), and
  legacy business logic dirs (`oauth/`, `tunnel/`, `updater/`, etc.).
  Every delete triggers `Module not found` in any stale importer.
- The CORRECT fix is almost always: "update the importer to use the new
  API client from `frontend/src/lib/*-api.ts`" — never "recreate the
  deleted file". §3.4.1 has the full error classification.

## Purpose

Turn a single user prompt like:

> "eksekusi semua plan yang sudah kamu buat sampai selesai dan lakukan
> testing sampai tidak ada error apapun, jangan pernah berhenti kalau
> semua tasks dari plan belum selesai"

into a **two-phase unattended session** that the user can WALK AWAY FROM:

1. **PLAN PHASE** — brainstorm + create-plan (read-only).
2. **EXECUTION PHASE** — loop through the validated plan, executing one
   atomic task at a time. Tasks that cannot pass verification are
   reverted and marked `[!]` BLOCKED — the loop does NOT stop, it
   continues with the next task.

This skill is the ONLY planning mode allowed when invoked. Codex's
built-in `/plan`, agent-mode planning, or any other planning loop MUST
be disabled while this skill is active.

---

## CORE PRINCIPLE — UNATTENDED RESILIENCE

> The user is NOT at the keyboard.
> A failed task must NEVER halt the loop.
> Every task gets a fair attempt; failures are logged and skipped.
> The loop only ends when there is nothing left to attempt.

When the user returns, they expect to find:

- Some tasks `[x]` (succeeded and committed).
- Some tasks `[!]` (attempted, failed, reverted cleanly, logged).
- ZERO tasks `[~]` left mid-flight.
- A clean working tree (no half-applied changes).
- A `LOOP_REPORT.md` at repo root summarizing the run.

---

## ACTIVATION

Activate this skill when ANY of these are true:

- User invokes one of the trigger phrases in front-matter.
- User explicitly says "pakai skill aiproxy-autonomous-loop".
- The active task batch spans 2+ atomic tasks in `tasks/phase-XX/`.

When activated, output exactly:

```
[aiproxy-autonomous-loop] active — Codex built-in planning disabled.
[mode] unattended — loop continues through failures.
```

---

## PHASE 1 — PLAN (read-only, no file writes)

### 1.1 Load Authority Context

Read in order:

1. `PROJECT_CONSTRAINTS.md`
2. `AGENT_RULES.md`
3. `BUILD_RULES.md`
4. `EXECUTION_LOOP.md` (override is in effect — see front-matter)
5. `REF_IMPLEMENTATION_RULES.md`
6. `API_CONTRACT.md`
7. `TASK_STATUS.md` → identify current phase and ALL `[ ]` / `[~]` tasks
8. `tasks/phase-XX/README.md` for the current phase
9. Every `tasks/phase-XX/T*.md` that is `[ ]` or `[~]`

### 1.2 Brainstorm (compressed, no user questions in unattended mode)

In unattended mode, **DO NOT ASK QUESTIONS**. Make reasonable
assumptions and proceed. Log assumptions in `LOOP_REPORT.md`.

Internal summary (do not block on user):

- Current phase + pending task IDs
- Dependency graph (who waits on whom)
- Phase exit criteria

### 1.3 Create Plan (template MANDATORY)

```markdown
# Plan — Phase <N>: <phase title>

<1–3 sentence intent and approach for this loop.>

## Scope
- In:  Tasks T<N>.<a>, T<N>.<b>, … (every `[ ]` or `[~]`)
- Out: Anything outside the listed task IDs, any other phase

## Execution Order
1. T<N>.<a> — <one-line title>
2. T<N>.<b> — <one-line title>
…

## Per-Task Loop (every item above)
- [ ] 1. Read task file
- [ ] 2. Generate manifest from TASK_EXECUTION_TEMPLATE.md
- [ ] 3. Implement inside Allowed Files only
- [ ] 4. Run verification suite
- [ ] 5a. On green: commit + mark `[x]`
- [ ] 5b. On red after fix attempts: `git restore`, mark `[!]`, continue
- [ ] 6. Move to next task — NEVER stop

## Phase Exit Gate (run AFTER loop, even if some `[!]` exist)
- [ ] All tasks above are `[x]` or `[!]` (none left `[ ]` or `[~]`)
- [ ] Write LOOP_REPORT.md with full breakdown
- [ ] If all `[x]`: also run build/typecheck/lint phase verification

## Assumptions (logged, no questions asked)
- <list assumptions made instead of asking the user>
```

### 1.4 Validate Plan, then exit Phase 1

Internal checklist (fix and re-emit if any fails):

- [ ] Every task is in the **current phase** only.
- [ ] Each task listed individually.
- [ ] No invented task IDs.
- [ ] Verification commands exist in `frontend/package.json`.

Emit literally:

```
[plan validated] entering execution loop.
```

---

## PHASE 2 — EXECUTION LOOP (unattended, never stops)

### 2.1 Loop Skeleton

```
FOR task IN plan.execution_order:
    branch_clean_check()                    # working tree clean?
    IF NOT clean:
        git stash drop || git restore .     # nuke uncommitted junk
    
    manifest = build_manifest(task)
    apply_changes(manifest)
    
    verification = run_verification_with_retries(task)
    
    IF verification == GREEN:
        commit(task)
        mark_done(task, TASK_STATUS.md)
        commit_status_update()
        log("✓ T<N>.<x> done")
    ELSE:
        # CRITICAL: do NOT stop. Revert and continue.
        git restore .                       # undo partial work
        git clean -fd                       # nuke untracked files
        mark_blocked(task, TASK_STATUS.md, reason=last_root_error)
        commit_status_update()              # commit the [!] marker
        log("✗ T<N>.<x> blocked: <reason>")
    
    CONTINUE                                # NEVER break out of the loop

# After the FOR loop completes:
write_loop_report()
run_phase_exit_gate_if_all_green()
emit("[loop] FINISHED — see LOOP_REPORT.md")
```

### 2.2 Per-Task Procedure

#### Step A — Pre-flight Cleanup

Before every task:

```bash
git status --porcelain         # must be empty
# if not empty:
git restore .
git clean -fd
```

This guarantees task N never inherits task N-1's mess.

#### Step B — Manifest

Generate manifest in memory from `TASK_EXECUTION_TEMPLATE.md`:
- Task ID, title, phase, risk
- Allowed Files (verbatim from task `.md`)
- Verification commands
- Commit message format: `T<N>.<x>: <title>`

#### Step C — Implement

Modify ONLY files in Allowed Files. If a needed file is outside scope,
do NOT expand — that task becomes `[!]` BLOCKED with reason
`scope_outside_allowed_files`.

#### Step D — Verify (see §3.3)

Run the verification suite. The agent may retry a fix UP TO 5 times
PER TASK using BUILD_RULES (first root error only). After 5 attempts:

- `git restore .` and `git clean -fd` to nuke partial work
- Mark `[!]` in TASK_STATUS.md with reason = last root error message
- Commit only the TASK_STATUS update
- Move to next task — DO NOT STOP

#### Step E — Commit (only on green)

```bash
git add <Allowed Files only>
git commit -m "T<N>.<x>: <task title>

Phase: <N>
Risk:  <risk>
Verification: typecheck ✓ build ✓ lint ✓
"
```

Then update `TASK_STATUS.md`:
- `[ ]` or `[~]` → `[x]` for this task
- Separate commit: `chore(status): T<N>.<x> done`

#### Step F — Mark Blocked (on red, after retries exhausted)

```bash
git restore .
git clean -fd
# now update TASK_STATUS.md to flip this task to [!]
# add a one-line note: T<N>.<x> [!] — <first root error short message>
git add TASK_STATUS.md
git commit -m "chore(status): T<N>.<x> blocked — <short reason>"
```

#### Step G — Continue

Loop back to next task IMMEDIATELY. No user prompt. No confirmation.
The loop is the contract.

### 2.3 Concurrency

Strictly sequential. Never parallelize. TASK_STATUS.md is a single
source of truth and concurrent writes break the audit trail.

---

## PHASE 3 — REUSABLE PROCEDURES

### 3.1 Reading a Task File

Path: `tasks/phase-<N>/T<N>.<x>-<slug>.md`

Required fields:
- Goal
- Allowed Files
- Verification
- Risk
- Dependencies

If any field is MISSING from the task file:
- Mark `[!]` with reason `task_file_malformed`
- Continue loop (do NOT stop the run)

### 3.2 Building the Manifest

Use `TASK_EXECUTION_TEMPLATE.md` verbatim. Fill in:
1. Current Task from §3.1
2. Allowed Files from §3.1
3. Forbidden:
   - `@ts-ignore`, `eslint-disable` to silence errors
   - Mass renames
   - Cross-task module deletion
   - Backend Go code changes
4. Verification from §3.1

### 3.3 Verification Suite

CONTEXT: This is a frontend-only refactor. Backend Go is feature-complete
and runs on `http://localhost:1432`. Errors will overwhelmingly come from
`npm run build` and `tsc --noEmit` because we are DELETING SQLite shims,
dual shells (`.js` + `.tsx`), and fullstack business-logic leakage. Most
failures will be:

- `Module not found: Can't resolve '@/lib/db/...'` (deleted target)
- `Cannot find name 'usageDb'` (removed shim)
- TS2307 / TS2305 / TS2339 (missing import / member / type)
- `npm ERR! peer dep` after `package.json` mutation
- Next.js build cache corruption after `.tsx`/`.js` shell removal

Verification commands (run in order, in `frontend/`):

```bash
cd frontend

# 0. Sync deps if package.json or lockfile changed in this task
if git diff --name-only HEAD~1..HEAD | grep -qE '^frontend/(package\.json|package-lock\.json)$'; then
    npm install --no-audit --no-fund --prefer-offline
fi

# 1. Typecheck (fastest, highest signal during refactor)
npm run typecheck 2>&1 || npx tsc --noEmit 2>&1

# 2. Lint
npm run lint 2>&1

# 3. Build (Next.js production build)
rm -rf .next                     # purge stale cache; cheap insurance
npm run build 2>&1

# 4. Task-specific verification (from the task file's "Verification" section)
<commands>

# 5. Tests if configured (don't fail loop if no test runner)
npm test -- --run 2>/dev/null || true
```

`npm run dev` is NOT used in per-task verification because it's a
long-running process. It's run ONCE as a smoke test at end of loop
(see §3.7).

### 3.4 Build Failure Protocol (BUILD_RULES.md, with retry cap)

1. Read **first error in build output** only.
2. Identify root cause of THAT error.
3. Classify the error (see §3.4.1 below).
4. Apply minimal fix inside Allowed Files.
5. Re-run §3.3 from top.
6. If green → continue Step E in §2.2.
7. If red AND attempt_count < 5 → goto step 1.
8. If attempt_count == 5 → §2.2 Step F (mark blocked, continue loop).

Forbidden under any circumstance:
- `as any`, `as unknown as X`
- `@ts-ignore`, `@ts-expect-error`
- `eslint-disable-next-line` / disable comments
- Commenting out failing tests
- Deleting failing tests
- Skipping verification steps
- Re-creating files the task explicitly deleted (would defeat the refactor)

### 3.4.1 Error Classification (frontend refactor patterns)

| Pattern in first error                                    | Likely cause                                  | Allowed minimal fix                                       |
|----------------------------------------------------------|----------------------------------------------|-----------------------------------------------------------|
| `Module not found: Can't resolve '@/lib/db/...'`         | Importer still references deleted SQLite layer | Update importer to use new API client (per `API_CONTRACT.md`) IF importer is in Allowed Files. Else `[!]` `importer_outside_scope` |
| `Module not found: Can't resolve '@/lib/usageDb'` / `localDb` | Importer using removed shim | Same as above — switch importer to API client |
| TS2307 `Cannot find module 'X'`                          | Same as Module not found, TS layer            | Same rule                                                  |
| TS2305 `Module has no exported member 'Y'`               | Re-export removed during refactor             | Update importer to use new export name from API client    |
| TS2339 `Property 'X' does not exist on type 'Y'`         | Type contract changed (often during T1.x)     | Use new type from `frontend/src/types/` if in Allowed Files. Else `[!]` |
| `npm ERR! peer dep` / `ERESOLVE`                         | Dep tree out of sync                          | `rm -rf node_modules package-lock.json && npm install`     |
| `EACCES` / `EPERM`                                       | Filesystem perms                              | `[!]` `system_perm_error`                                  |
| `JavaScript heap out of memory`                          | Next.js OOM after large delete                | Retry build once with `NODE_OPTIONS=--max-old-space-size=4096`. If still red → `[!]` |
| `.next` parse error / `Failed to compile cache`          | Stale build cache                              | `rm -rf frontend/.next` (already in §3.3) then retry      |
| `Error: Cannot find module './chunks/...'`               | Same — stale `.next`                          | Same — purge `.next`                                       |
| `ENOENT: no such file or directory, open '...page.js'`   | Next.js confused by dual shell removal        | Verify only `.tsx` shell exists; if both deleted, `[!]` `shell_missing` |
| TS error in file NOT in Allowed Files                    | Cascade from this task's deletion             | DO NOT touch — `[!]` `cascade_outside_scope`, log the cascade target for the next task to handle |

If the first error doesn't match any pattern above, fall back to
generic BUILD_RULES.md Step 1: fix the first root only.

### 3.5 Progress Logging

After each task:

```
[loop] T<N>.<x> ✓  ( <done>/<total> · ok=<n_ok> blocked=<n_blocked> )
```

or

```
[loop] T<N>.<x> ✗ blocked  ( <done>/<total> · ok=<n_ok> blocked=<n_blocked> )
       reason: <first root error one-liner>
```

At the very end:

```
[loop] FINISHED — <n_ok> ok · <n_blocked> blocked · see LOOP_REPORT.md
```

### 3.7 End-of-Loop Smoke Test (npm run dev, bounded)

ONLY after the FOR loop in §2.1 completes, run a single bounded dev-server
smoke test to confirm the app boots. Skip if zero tasks were `[x]` (nothing
to smoke-test).

```bash
cd frontend
timeout 45s npm run dev > /tmp/dev.log 2>&1 &
DEV_PID=$!
sleep 30
# success criteria: a "ready" line appears in log within 30s
if grep -qE '(ready|started server|listening|Local:)' /tmp/dev.log; then
    SMOKE=ok
else
    SMOKE=fail
fi
kill $DEV_PID 2>/dev/null
wait $DEV_PID 2>/dev/null
```

Log result to LOOP_REPORT.md under a `## Smoke Test` section. A failing
smoke test does NOT undo any committed task — it just gets reported.
Reason: refactor may legitimately leave dev broken between phases (e.g.
Phase 2 isolates SQLite but Phase 3 is what actually routes through Go).

### 3.6 LOOP_REPORT.md (mandatory output at end)

Write to repo root. Format:

```markdown
# Loop Report — <ISO timestamp>

Phase: <N>
Branch: phase/<N>-<slug>
Duration: <Hh Mm>

## Summary
- Total tasks attempted: <n>
- ✓ Completed: <n_ok>
- ✗ Blocked: <n_blocked>

## Completed
- T<N>.<x> — <title>
- ...

## Blocked (require human review)
### T<N>.<y> — <title>
- First root error: <message>
- File: <path:line>
- Attempts: <count>
- Last action before revert: <one line>

## Assumptions Made
- <list>
```

Commit it: `chore: loop report <timestamp>`.

---

## PHASE 4 — INTERACTION CONTRACT

While this skill is active:

- The agent **NEVER** asks "should I continue?" — answer is always YES.
- The agent **NEVER** asks for clarification mid-loop — log assumption, continue.
- The agent **NEVER** stops on a single task failure.
- The agent **NEVER** stops because "this looks risky" — risk is per-task, not per-loop.
- The agent **DOES** stop only on: user interrupt, plan empty, critical system error.

On user interrupt:
1. Finish current task's verification.
2. If green, commit. If red, `git restore .` first.
3. Write partial LOOP_REPORT.md.
4. Emit `[loop] STOPPED by user — partial report written`.

---

## PHASE 5 — RESUME

User re-invokes any trigger phrase. The skill:

1. Reads `TASK_STATUS.md` for current phase.
2. Filters to `[ ]` and `[~]` (skips `[x]` AND `[!]` — blockers need human).
3. If user explicitly says "retry blocked tasks", also include `[!]`.
4. Re-emits plan covering remaining tasks.
5. Continues loop.

---

## APPENDIX A — Files this skill writes

During Phase 2 only:

- `frontend/**` (subject to per-task Allowed Files)
- `TASK_STATUS.md` (status flips: `[x]`, `[!]`)
- `LOOP_REPORT.md` at repo root (created/overwritten per run)
- Git commits and branch moves

Never written:

- `tasks/**` (read-only execution surface)
- Authority docs (`*_RULES.md`, `PROJECT_CONSTRAINTS.md`, `ARCHITECTURE.md`,
  `EXECUTION_LOOP.md`, `API_CONTRACT.md`, `FEATURE_PARITY.md`,
  `FEATURE_AUDIT.md`, `PARITY_CHECKLIST.md`,
  `FRONTEND_REFACTOR_EXECUTION_PLAN.md`,
  `FRONTEND_MIGRATION_TASK_INDEX.md`, `STRUCTURE.md`,
  `REF_IMPLEMENTATION_RULES.md`, `DISPATCHER_PROMPT.md`,
  `TASK_EXECUTION_TEMPLATE.md`, `FRONTEND_REFACTOR_TRACKER.md`)
- `backend/**`

## APPENDIX B — Failure / Resume Matrix

| Situation                          | Action                                       |
|------------------------------------|----------------------------------------------|
| Verification fails (red)           | Retry up to 5x, then `[!]` and continue      |
| Same root error 5x                 | `git restore`, `[!]`, continue               |
| Task file missing required field   | `[!]` task_file_malformed, continue          |
| Allowed Files insufficient         | `[!]` scope_outside_allowed_files, continue  |
| Backend code change required       | `[!]` out_of_scope_backend, continue         |
| Working tree dirty before task     | Auto `git restore` + `git clean -fd`         |
| User issues "stop"                 | Finish verify only, write report, exit       |
| Plan list exhausted                | Write LOOP_REPORT.md, exit cleanly           |

— END OF SKILL —
