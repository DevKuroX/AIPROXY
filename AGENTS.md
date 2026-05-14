# AGENTS.md — AIPROXY Operating Manual for AI Coding Agents

> This file is auto-discovered by **OpenAI Codex CLI** (and similar agents)
> when started inside this repository. It defines the working agreement
> for ALL AI coding agents operating on AIPROXY.
>
> Discovery order (Codex CLI):
> `~/.codex/AGENTS.override.md` → `~/.codex/AGENTS.md` → **this file** →
> nested `AGENTS.override.md` if any. Closer-to-CWD wins.

---

## 1. Project context (1 paragraph)

AIPROXY is a multi-provider AI proxy/router platform inspired by 9router.
**Backend is Go (port 1432), production-complete, OUT OF SCOPE.**
**Frontend is Next.js 16 + React 19, currently mid-refactor** from
fullstack-hybrid (9router-style with embedded SQLite + business logic) to
thin-client (calls Go backend via `frontend/src/app/api/*` proxy routes).
ALL agent work in this repo targets the frontend refactor only.

---

## 2. Authority Documents (read these BEFORE any change)

In priority order (top wins on conflict):

1. `PROJECT_CONSTRAINTS.md` — root AI rules, authoritative
2. `AGENT_RULES.md` — task-boundary discipline, one task = one commit
3. `BUILD_RULES.md` — anti-panic-patching, first root error only
4. `REF_IMPLEMENTATION_RULES.md` — 9router = behavior reference only
5. `ARCHITECTURE.md` — target thin-client shape
6. `API_CONTRACT.md` — backend base URL `http://localhost:1432`
7. `STRUCTURE.md` — folder layout
8. `EXECUTION_LOOP.md` — bounded execution protocol
   (NOTE: §5 below overrides this for unattended mode)
9. `TASK_STATUS.md` — current phase + active task
10. `tasks/phase-<N>/README.md` + per-task `T<N>.<x>-*.md` files

If any rule above conflicts with this file, **the rule file wins** for
behavior; this file wins for **execution loop semantics only**.

---

## 3. Default Mode — Bounded (interactive)

When the user gives a task without unattended trigger phrases, follow
`EXECUTION_LOOP.md` as written:

```
EXECUTE ONE TASK. STOP. AWAIT INSTRUCTION.
```

- Use `DISPATCHER_PROMPT.md` template format.
- Use `TASK_EXECUTION_TEMPLATE.md` for the manifest.
- One task = one commit.
- BUILD_RULES.md Step 1: first root error only.

---

## 4. Activation — Unattended Loop Mode

This mode REPLACES Codex's built-in planning and EXECUTION_LOOP's
"STOP per task" rule. It activates when the user says ANY of:

- "eksekusi semua plan"
- "kerjakan sampai selesai"
- "jangan pernah berhenti"
- "full loop"
- "autonomous run"
- "idle mode"
- "tinggal pergi"
- "unattended"

When activated, emit this EXACT first line:

```
[autonomous-loop] active — Codex built-in planning disabled.
[mode] unattended — loop continues through failures.
```

Detailed playbook lives at:
`.codex/skills/aiproxy-autonomous-loop/SKILL.md` (read on activation).

---

## 5. Unattended Loop Protocol (overrides EXECUTION_LOOP.md)

```
EXECUTE ONE TASK. VERIFY. COMMIT OR MARK BLOCKED. CONTINUE.
```

**The loop NEVER stops on a single task failure.** A task that cannot pass
verification is reverted with `git restore . && git clean -fd`, marked
`[!]` in TASK_STATUS.md, then the loop proceeds to the NEXT task.

### 5.1 Stop conditions (ONLY these end the loop)

- All tasks in plan are processed (`[x]` OR `[!]`).
- User explicitly interrupts.
- Critical system failure (disk full, no network, repo locked).

### 5.2 NEVER stop on

- Single task verification failure → `[!]`, continue.
- Multiple consecutive failures → mark all `[!]`, continue.
- Same root error after 5 retries → revert, `[!]`, continue.
- Missing task dependency → skip dependent, continue with independent.
- Missing field in task file → `[!]` `task_file_malformed`, continue.
- Allowed Files too narrow → `[!]` `scope_outside_allowed_files`, continue.

### 5.3 Per-task procedure

For each task in the plan, strict sequential:

1. **Pre-flight**: working tree must be clean. If dirty, `git restore .`
   then `git clean -fd`.
2. **Read** `tasks/phase-<N>/T<N>.<x>-*.md`.
3. **Build manifest** from `TASK_EXECUTION_TEMPLATE.md` (Task ID, title,
   Allowed Files, verification commands, risk).
4. **Implement** changes ONLY in Allowed Files declared by the task.
5. **Verify** (see §6).
6. **On GREEN**:
   - `git add <Allowed Files only>`
   - `git commit -m "T<N>.<x>: <title>"`
   - Update `TASK_STATUS.md` → flip `[ ]`/`[~]` to `[x]`.
   - `git commit -m "chore(status): T<N>.<x> done"` (separate commit).
   - Log: `[loop] T<N>.<x> ✓ ( done/total · ok=n blocked=m )`
7. **On RED** (after retry policy in §6 exhausted):
   - `git restore .`
   - `git clean -fd`
   - Update `TASK_STATUS.md` → `[!]` with reason = first root error msg.
   - `git commit -m "chore(status): T<N>.<x> blocked — <reason>"`
   - Log: `[loop] T<N>.<x> ✗ blocked ( ... )`
8. **Continue** to next task. No confirmation request. Ever.

### 5.4 Concurrency

Strictly sequential. NEVER parallelize. TASK_STATUS.md is the single
source of truth and concurrent writes corrupt the audit trail.

---

## 6. Verification suite (frontend-only refactor)

Errors will overwhelmingly come from `npm run build` / `tsc --noEmit`
because the refactor is **destructive**: it deletes SQLite shims
(`usageDb.js`, `localDb.js`, `lib/db/`), dual app shells (`page.js`+`page.tsx`),
and legacy business-logic dirs (`oauth/`, `tunnel/`, `updater/`, ...).

### 6.1 Commands (run in order, in `frontend/`)

```bash
cd frontend

# 0. Sync deps if package.json or lockfile changed
if git diff --name-only HEAD~1..HEAD | grep -qE '^frontend/(package\.json|package-lock\.json)$'; then
    npm install --no-audit --no-fund --prefer-offline
fi

# 1. Typecheck (fastest, highest signal)
npm run typecheck 2>&1 || npx tsc --noEmit 2>&1

# 2. Lint
npm run lint 2>&1

# 3. Build (purge stale cache first)
rm -rf .next
npm run build 2>&1

# 4. Task-specific verification (from the task file's "Verification" section)
<commands listed in the task file>

# 5. Tests if configured (do not fail loop if no test runner)
npm test -- --run 2>/dev/null || true
```

### 6.2 Retry policy

- Up to **5 retries** per task. Each retry: apply BUILD_RULES.md Step 1
  (read first root error, fix that root only).
- After 5 retries → §5.3 step 7 (revert, mark `[!]`, continue loop).

### 6.3 Error classification (refactor patterns)

| First error pattern                                | Correct minimal fix (inside Allowed Files only)            |
|----------------------------------------------------|-----------------------------------------------------------|
| `Module not found: '@/lib/db/...'`                 | Update importer to use API client per `API_CONTRACT.md`. **NEVER** recreate deleted SQLite layer. |
| `Module not found: '@/lib/usageDb'` / `'localDb'`  | Same — switch importer to API client.                     |
| TS2307 `Cannot find module 'X'`                    | Same.                                                      |
| TS2305 `Module has no exported member 'Y'`         | Update importer to new export from API client.            |
| TS2339 `Property 'X' does not exist on 'Y'`        | Use new type from `frontend/src/types/` if in scope.      |
| `npm ERR! peer dep` / `ERESOLVE`                   | `rm -rf node_modules package-lock.json && npm install`     |
| `JavaScript heap out of memory`                    | Retry once with `NODE_OPTIONS=--max-old-space-size=4096`. |
| `.next` parse / chunk error                        | `rm -rf frontend/.next` then retry (§6.1 already does).   |
| `ENOENT ... page.js`                               | Verify only `.tsx` shell exists. If both missing → `[!]`. |
| TS error in file OUTSIDE Allowed Files             | DO NOT touch. `[!]` `cascade_outside_scope`.              |

If first error doesn't match the table → fall back to BUILD_RULES.md
Step 1 generic protocol.

### 6.4 ABSOLUTELY FORBIDDEN

Under any circumstance, do NOT:

- Use `@ts-ignore`, `@ts-expect-error`, `as any`, `as unknown as X`.
- Add `eslint-disable-next-line` or any disable comment.
- Comment out failing tests or delete failing tests.
- Skip verification steps.
- Re-create files the task explicitly deleted (would defeat the refactor).
- Modify `backend/`, `tasks/`, or any `*_RULES.md`/`*_CONSTRAINTS.md` doc.
- Push directly to `main`. Only push to `phase/<N>-<slug>` branches.

---

## 7. End-of-loop artifacts (mandatory)

After the FOR loop in §5 completes (regardless of how many `[!]`):

### 7.1 Smoke test (bounded)

```bash
cd frontend
timeout 45s npm run dev > /tmp/dev.log 2>&1 &
DEV_PID=$!
sleep 30
if grep -qE '(ready|started server|listening|Local:)' /tmp/dev.log; then
    SMOKE=ok
else
    SMOKE=fail
fi
kill $DEV_PID 2>/dev/null
wait $DEV_PID 2>/dev/null
```

A failing smoke test does NOT undo committed tasks. It only gets logged.

### 7.2 LOOP_REPORT.md at repo root

```markdown
# Loop Report — <ISO timestamp>

Phase: <N>
Branch: phase/<N>-<slug>
Duration: <Hh Mm>

## Summary
- Total attempted: <n>
- ✓ Completed: <n_ok>
- ✗ Blocked: <n_blocked>

## Completed
- T<N>.<x> — <title>
...

## Blocked (require human review)
### T<N>.<y> — <title>
- First root error: <message>
- File: <path:line>
- Attempts: <count>
- Last action before revert: <one line>

## Smoke Test
<ok | fail with reason>

## Assumptions Made
- <list>
```

Commit: `chore: loop report <ISO timestamp>`.

### 7.3 Push branch (NEVER to main)

```bash
git push origin phase/<N>-<slug>
```

### 7.4 Final log line

```
[loop] FINISHED — <n_ok> ok · <n_blocked> blocked · see LOOP_REPORT.md
```

---

## 8. Interaction contract (unattended mode)

While in unattended mode:

- The agent **NEVER** asks "should I continue?".
- The agent **NEVER** asks clarifying questions mid-loop — assume,
  log under "Assumptions Made" in LOOP_REPORT.md.
- The agent **NEVER** stops on a single task failure.
- The agent **NEVER** says "this looks risky" and halts — risk is
  per-task, not per-loop.
- The agent **ONLY** stops on: user interrupt, plan empty, system error.

On user interrupt (Ctrl+C, "stop", etc.):

1. Finish current task's verification step only.
2. If green, commit. If red, `git restore .` first.
3. Write partial LOOP_REPORT.md.
4. Emit `[loop] STOPPED by user — partial report written`.

---

## 9. Resume

User re-invokes any trigger phrase from §4. The loop:

1. Reads `TASK_STATUS.md` for current phase.
2. Plans over `[ ]` and `[~]` only (skips `[x]` AND `[!]` unless user
   says "retry blocked tasks").
3. Continues from there.

---

## 10. Reference detail

For the extended playbook (manifests, additional appendices), read:

- `.codex/skills/aiproxy-autonomous-loop/SKILL.md`
- `.codex/skills/aiproxy-autonomous-loop/PROMPT_TEMPLATES.md`
- `.codex/skills/aiproxy-autonomous-loop/INSTALL_PROMPT.md`

These are reference documents, not auto-loaded. This file (AGENTS.md)
is sufficient to run the loop unattended.

— END OF AGENTS.md —
