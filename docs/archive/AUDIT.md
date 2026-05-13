# AUDIT.md — Setup Pack Audit Report

> Generated automatically while preparing this repo for opencode.
> Cross-checks every `.md` in `DevKuroX/AIPROXY` against the actual
> `decolua/9router` source.

## 0. Reference inventory

- 9router checked out from `https://github.com/decolua/9router` at the
  commit that was current when this audit ran.
- AIPROXY spec docs read from `https://github.com/DevKuroX/AIPROXY`.

---

## 1. File-existence checks (all PASS)

Every JS path referenced by `AGENTS.md`, `RTK_SPEC.md`, `CAVEMAN_SPEC.md`,
`ARCHITECTURE.md`, `PLAN.md`, `ROADMAP.md` was verified to exist in 9router:

| Spec reference | 9router path | Status |
|---|---|---|
| RTK entry | `open-sse/rtk/index.js` | ✅ |
| RTK constants | `open-sse/rtk/constants.js` | ✅ |
| RTK autodetect | `open-sse/rtk/autodetect.js` | ✅ |
| RTK apply | `open-sse/rtk/applyFilter.js` | ✅ |
| RTK registry | `open-sse/rtk/registry.js` | ✅ |
| Caveman | `open-sse/rtk/caveman.js`, `cavemanPrompts.js` | ✅ |
| Translator entry | `open-sse/translator/index.js` | ✅ |
| Translator formats | `open-sse/translator/formats.js` | ✅ |
| Chat handler | `open-sse/handlers/chatCore.js` | ✅ |
| Executor base | `open-sse/executors/base.js` | ✅ |
| Executor default | `open-sse/executors/default.js` | ✅ |
| Cursor checksum | `open-sse/utils/cursorChecksum.js` | ✅ |
| Cursor protobuf | `open-sse/utils/cursorProtobuf.js` | ✅ |
| Fallback service | `open-sse/services/accountFallback.js` | ✅ |
| Token refresh | `open-sse/services/tokenRefresh.js` | ✅ |
| Combo service | `open-sse/services/combo.js` | ✅ |
| Data dir | `src/lib/dataDir.js` | ✅ |
| Disabled models | `src/lib/disabledModelsDb.js` | ✅ |
| Cloud sync | `src/lib/initCloudSync.js` | ✅ |

---

## 2. Findings (issues fixed in this setup-pack)

### Finding 1 — Phantom file: `executor/claude.go`

- **Where:** `ARCHITECTURE.md` lists `internal/executor/claude.go`.
- **Reality:** `open-sse/executors/` has NO `claude.js`. The
  `executors/index.js` registry contains no `claude` entry.
  Claude is handled by `DefaultExecutor` after translation.
- **Fix applied:** removed from `docs/ARCHITECTURE.md`; clarified in
  `docs/EXECUTORS.md` and `docs/PARITY_CHECKLIST.md` §4.

### Finding 2 — Phantom file: `auth/oauth/opencode.go`

- **Where:** `ARCHITECTURE.md` lists `internal/auth/oauth/opencode.go`.
- **Reality:** `src/lib/oauth/services/` has NO `opencode.js`.
  `opencode` is an executor only (uses pre-issued credentials).
- **Fix applied:** removed from `docs/ARCHITECTURE.md`; clarified in
  `docs/AUTH.md`.

### Finding 3 — Missing OAuth flows: `qoder`, `openai`

- **Where:** `ARCHITECTURE.md` OAuth list is missing `qoder.go` and `openai.go`.
- **Reality:** Both exist in 9router: `src/lib/oauth/services/qoder.js` and
  `src/lib/oauth/services/openai.js`.
- **Fix applied:** added both to `docs/ARCHITECTURE.md`,
  `docs/AUTH.md`, and `docs/PARITY_CHECKLIST.md` §5.

### Finding 4 — RTK filter `git-log` declared but not implemented

- **Where:** `RTK_SPEC.md` §7 lists `git-log` as an implementable filter.
- **Reality:** `open-sse/rtk/constants.js:45` declares `GIT_LOG: "git-log"`,
  but there is NO `filters/gitLog.js` and `git-log` is NOT in the registry
  in `open-sse/rtk/registry.js`. It is effectively dead code in 9router.
- **Fix applied:** `docs/RTK_SPEC.md` now carries a warning at the top;
  `docs/PARITY_CHECKLIST.md` §6 documents the decision.

### Finding 5 — Doc paths inconsistent

- **Where:** `AGENTS.md` and `ARCHITECTURE.md` reference
  `docs/PARITY_CHECKLIST.md`, `docs/CONVENTIONS.md`, `docs/TRANSLATORS.md`,
  `docs/EXECUTORS.md`, `docs/API.md`, `docs/DATABASE.md`, `docs/AUTH.md`,
  `docs/FALLBACK.md`, `docs/STREAMING.md`, `docs/USAGE.md`,
  `docs/OBSERVABILITY.md`, `docs/CLI_TOOLS.md`, `docs/FRONTEND.md`,
  `docs/openapi.yaml`.
- **Reality:** none of those files existed in the AIPROXY repo. Also,
  `ARCHITECTURE.md`, `RTK_SPEC.md`, `CAVEMAN_SPEC.md` sat at the repo root,
  not under `docs/`.
- **Fix applied:** `ARCHITECTURE.md`, `RTK_SPEC.md`, `CAVEMAN_SPEC.md` moved
  to `docs/`. All 11 missing doc stubs created with section scaffolding
  and TODO markers (PARITY_CHECKLIST, CONVENTIONS, TRANSLATORS, EXECUTORS,
  API, DATABASE, AUTH, FALLBACK, STREAMING, USAGE, OBSERVABILITY, CLI_TOOLS,
  FRONTEND).
  `docs/openapi.yaml` is the only file NOT scaffolded — it should be
  generated as you build the API.

### Finding 6 — No agent-config mirrors

- **Where:** `ARCHITECTURE.md` mentions `.opencode/rules.md`, `.claude/CLAUDE.md`,
  `.cursorrules` should exist as mirrors of `AGENTS.md`.
- **Reality:** none of those existed.
- **Fix applied:** `.opencode/rules.md`, `.claude/CLAUDE.md`, `.cursorrules`
  scaffolded, pointing at `AGENTS.md` as the source of truth.

---

## 3. Sanity checks that passed

- Port number is consistent: 9router uses `:20128` (`package.json` dev
  script). `ROADMAP.md` Phase 0 also says `chi` on `:20128`. ✅
- Executor inventory (excluding the phantom `claude.go`) matches
  `executors/index.js` exactly. ✅
- RTK filter inventory (excluding `git-log`) matches `registry.js` exactly. ✅
- Tech-stack decisions (ADR-001..010 in `PLAN.md`) are mutually consistent
  with the Hard Rules in `AGENTS.md`. ✅
- 9router has multiple SQLite paths (`localDb.js` + `disabledModelsDb.js` +
  a separate usage DB). The spec ADR-002 + Hard Rule 1 already note we
  consolidate under a single `${DATA_DIR}/9rgo.db`. ✅

---

## 4. Items left for you to decide

1. **OpenAPI spec.** `docs/openapi.yaml` is the official contract. Decide:
   - Hand-write it now and codegen everything else, OR
   - Generate it from Go handler annotations (e.g. `swaggo/swag`).
   The Hard Rules suggest the former (single source of truth, codegen
   on both sides). Pick before starting Phase 1.
2. **Claude Code's `CLAUDE.md` location.** Some users prefer it at the
   repo root rather than `.claude/CLAUDE.md`. Both work — pick one or
   symlink.
3. **Frontend codegen tool.** `openapi-typescript` is the simplest. If you
   need full SDK generation including hooks, consider `openapi-fetch` or
   `orval`. Decide before Phase 0 finishes.
4. **CI provider.** `ROADMAP.md` Phase 0 says GitHub Actions. If you target
   a different host, swap and update `.github/` references.
5. **Module path.** Replace `github.com/<you>/9rgo` in `go.mod`,
   `Taskfile.yml`, etc., with your actual repo path.

---

## 5. Files in this setup-pack

```
AGENTS.md                      # patched with status callout
PLAN.md                        # unchanged
ROADMAP.md                     # unchanged
AUDIT.md                       # this file
.opencode/rules.md             # NEW — mirror for OpenCode
.claude/CLAUDE.md              # NEW — mirror for Claude Code
.cursorrules                   # NEW — mirror for Cursor
docs/
  ARCHITECTURE.md              # moved from root + patched
  RTK_SPEC.md                  # moved from root + patched
  CAVEMAN_SPEC.md              # moved from root
  PARITY_CHECKLIST.md          # NEW
  CONVENTIONS.md               # NEW
  TRANSLATORS.md               # NEW (scaffold)
  EXECUTORS.md                 # NEW (scaffold)
  API.md                       # NEW (scaffold)
  DATABASE.md                  # NEW (scaffold)
  AUTH.md                      # NEW (scaffold)
  FALLBACK.md                  # NEW (scaffold)
  STREAMING.md                 # NEW (scaffold)
  USAGE.md                     # NEW (scaffold)
  OBSERVABILITY.md             # NEW (scaffold)
  CLI_TOOLS.md                 # NEW (scaffold)
  FRONTEND.md                  # NEW (scaffold)
```

---

## 6. How to apply

```bash
# from your new Go project root (which currently mirrors AIPROXY layout):
cd your-go-project

# wipe the old root-level specs that we moved into docs/
rm -f ARCHITECTURE.md RTK_SPEC.md CAVEMAN_SPEC.md

# unzip / copy this setup-pack on top
unzip ../aiproxy-setup-pack.zip

# add 9router as the read-only reference
mkdir -p _ref
git clone --depth 1 https://github.com/decolua/9router.git _ref/9router
echo "/_ref/" >> .gitignore   # don't track the reference checkout

# verify
ls docs/                      # should list all 14 .md files
cat AGENTS.md | head -20      # should show "Status note" callout
```

Then open the project in opencode / Claude Code / Cursor. Each tool will
auto-pick up its rules file. Start at `ROADMAP.md` → Phase 0.
