# AIPROXY Migration Build Rules

> **Purpose**: Prevent AI panic patching during build failures
> **Audience**: AI executors during verification phase
> **Status**: OPERATIONAL

---

# CRITICAL RULE

```
READ FIRST ROOT ERROR ONLY.
FIX THAT ERROR ONLY.
```

Do NOT:
- Fix cascade errors blindly
- Mass rename files
- Refactor while build unstable
- Delete systems prematurely
- Add type suppressions

---

# BUILD FAILURE PROTOCOL

## Step 1: Identify Root Error

When `npm run build` or `tsc --noEmit` fails:

1. Read ONLY the FIRST error
2. Identify the file and line number
3. Identify the error type

```
Example:
src/lib/api.ts(42,5): error TS2322: Type 'string' is not assignable to type 'number'.
                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                      This is the ROOT ERROR
```

## Step 2: Classify Error

| Error Type | Action |
|------------|--------|
| Type mismatch | Fix types in YOUR code |
| Missing import | Add import in YOUR code |
| Missing export | Export from YOUR code |
| Syntax error | Fix syntax in YOUR code |
| Circular dependency | Restructure YOUR code |

## Step 3: Verify Ownership

Ask: **"Did I modify this file in this task?"**

| Answer | Action |
|--------|--------|
| YES | Fix your change |
| NO | STOP — report unrelated error |

## Step 4: Fix Minimally

Fix ONLY:
- The specific error
- In the specific file
- With minimal change

Do NOT:
- Refactor surrounding code
- "Clean up" while fixing
- Rename variables
- Extract functions
- Add abstractions

## Step 5: Re-verify

Run the SAME build command.
If new error appears, repeat from Step 1.

---

# FORBIDDEN ACTIONS

## Type Suppression

NEVER use:
```typescript
// ❌ FORBIDDEN
as any
@ts-ignore
@ts-expect-error
// @ts-nocheck
```

If you think you need these:
- STOP
- Re-read the error
- Fix the root cause
- Or report blocker

## Mass Rename

NEVER rename multiple files during build failure.

```bash
# ❌ FORBIDDEN
mv src/lib/old*.ts src/lib/new*.ts
```

Mass rename causes:
- Import cascade errors
- Hidden breakage
- Context explosion

## Cascade Fixing

NEVER fix multiple errors at once.

```bash
# ❌ FORBIDDEN
# Fixing 10 type errors in one pass
```

Fix errors ONE AT A TIME.
Re-build after EACH fix.

## Premature Deletion

NEVER delete systems during build failure.

```bash
# ❌ FORBIDDEN
rm -rf src/lib/usage/
```

Deletion while unstable:
- Hides real errors
- Breaks dependencies
- Loses checkpoint

---

# IMPORT VERIFICATION

Before patching imports:

## Step 1: Verify Path

```typescript
// Check if path is correct
import { foo } from '@/lib/bar';
```

- Does `@/lib/bar` exist?
- Does it export `foo`?
- Is the path alias configured?

## Step 2: Check Circular

```bash
# Check for circular dependency
npx madge --circular src/
```

If circular:
- Report blocker
- Do NOT patch

## Step 3: Check Shim Status

```bash
# If importing from shims, verify shim still exists
rg -n "usageDb|localDb" src/lib/
```

If shim deleted:
- Use replacement API client
- Do NOT recreate shim

---

# SHIM SYSTEM RULES

## Current Shims (DO NOT DELETE YET)

```
src/lib/usageDb.js
src/lib/localDb.js
src/lib/disabledModelsDb.js
src/lib/requestDetailsDb.js
src/lib/dataDir.js
src/lib/db/
```

## When Shims Can Be Deleted

ONLY after:
1. All imports migrated (T8.1 verified)
2. Phase 8 reached
3. Explicit task to delete (T8.2, T8.3, T8.4)

## During Build Failure

If build fails due to shim import:
- Check if replacement client exists
- Migrate import to replacement
- Do NOT delete shim

---

# BUILD COMMAND REFERENCE

## Type Check (Fast)

```bash
cd /home/ubuntu/ai_proxy/frontend
npx tsc --noEmit
```

## Full Build

```bash
cd /home/ubuntu/ai_proxy/frontend
npm run build
```

## Lint Check

```bash
cd /home/ubuntu/ai_proxy/frontend
npm run lint
```

## Import Search

```bash
cd /home/ubuntu/ai_proxy/frontend
rg -n "<pattern>" src/
```

---

# ERROR TYPE GUIDE

## TypeScript Errors

| Error Code | Meaning | Fix |
|------------|---------|-----|
| TS2304 | Cannot find name | Add import or declaration |
| TS2307 | Cannot find module | Check path, add package |
| TS2322 | Type not assignable | Fix type or cast (rare) |
| TS2339 | Property not exist | Check interface, add property |
| TS2345 | Argument type wrong | Fix argument type |
| TS2403 | Variable conflict | Rename or scope |

## Next.js Errors

| Error | Meaning | Fix |
|-------|---------|-----|
| Module not found | Import path wrong | Fix import |
| Hydration mismatch | SSR/client differ | Check initial state |
| Chunk load error | Build artifact missing | Rebuild |

## ESLint Errors

| Error | Meaning | Fix |
|-------|---------|-----|
| no-unused-vars | Variable declared but not used | Remove or use |
| no-undef | Variable not defined | Add import or declaration |
| import/no-unresolved | Import path wrong | Fix import |

---

# STABLE CHECKPOINT

After successful build:

```bash
# Create checkpoint
git add -A
git commit -m "checkpoint: <task-id> verification pass"
```

If next task breaks build:
```bash
# Can always return here
git checkout HEAD
```

---

# EMERGENCY PROTOCOL

If build is completely broken:

1. STOP all changes
2. Return to last known good commit
   ```bash
   git log --oneline | head -20
   git checkout <last-good-commit>
   ```
3. Report blocker
4. Do NOT try to "fix forward"

---

# VERIFICATION CHECKLIST

Before declaring task done:

- [ ] `tsc --noEmit` passes
- [ ] `npm run build` passes (if modified source)
- [ ] `npm run lint` passes (if configured)
- [ ] No `@ts-ignore` or `as any` added
- [ ] Checkpoint committed
- [ ] TASK_STATUS.md updated
