# docs/CONVENTIONS.md — Go Style, Naming, Anti-patterns

> Mandatory reading before writing any Go in this project. Builds on
> `AGENTS.md` → Hard Rules. If you find a rule here that contradicts
> `AGENTS.md`, **AGENTS.md wins**, and you must fix this file.

---

## 1. Package & file layout

- One package per directory. Package name = directory base name (lowercase, single word).
- File names are `snake_case.go`. One concept per file.
- `internal/` is for everything non-public. We only use `pkg/` for code we
  genuinely want consumed by external Go importers (none today).
- The `internal/router` package is the **only** place that knows about all
  other subsystems. Other packages must not import each other in cycles.

```
forbidden:
  internal/translator depending on internal/executor
  internal/executor   depending on internal/translator

allowed:
  internal/router → translator, executor, rtk, caveman, storage, auth
  everyone        → models, errs, stream, observability
```

---

## 2. Naming

| Thing | Rule | Example |
|---|---|---|
| Package | lowercase, no underscores | `translator`, `rtk`, not `gateway_router` |
| Receiver | 1-3 chars, type-derived | `r *Router`, `ex Executor` |
| Interface | suffix `-er` when single-method; descriptive noun when multi | `Executor`, `Flusher`, `TokenRefresher` |
| Exported | `CamelCase` | `func InjectCaveman(...)` |
| Constants | `CamelCase`. Use `SCREAMING_SNAKE` ONLY when mirroring a 9router constant verbatim | `MaxStreamBytes`, `FILTERS_GIT_DIFF` (mirror) |
| Errors | `Err`-prefix for sentinels; typed `*errs.Error` for envelopes | `var ErrModelNotFound = errors.New(...)` |
| Generated files | suffix `.gen.go`; never edit manually | `queries.sql.gen.go` |
| Test files | `*_test.go`, same package | `inject_test.go` |
| Golden fixtures | `testdata/<feature>/<case>.in.json` + `.out.json` | `testdata/translator/openai_to_claude/basic.in.json` |

---

## 3. Error handling

- Return errors, never panic in request path.
- Wrap with context: `fmt.Errorf("loading provider %q: %w", id, err)`.
- Convert to HTTP via `internal/errs/http.go` only at the boundary
  (`internal/api/*` handlers).
- For OpenAI-shape error responses, marshal via `internal/errs/openai.go` —
  do not hand-roll JSON.
- Sentinel `errors.Is` checks for control flow; type assertions
  (`var e *errs.Error`) for envelope-shaped errors.

```go
// good
if errors.Is(err, storage.ErrNotFound) {
    return errs.NotFound("provider", id)
}

// bad
if err.Error() == "not found" { ... }   // never string-match errors
```

---

## 4. Logging

- `slog` only. Use the package-level logger in `internal/observability`.
- Always `slog.InfoContext(ctx, ...)` — never `slog.Info(...)` in request code.
- Field naming: `snake_case`. Reserved keys:
  `request_id`, `provider_id`, `model`, `account_id`, `endpoint`, `duration_ms`.
- Levels:
  - `Debug` — verbose, off in prod
  - `Info` — request lifecycle (one per request)
  - `Warn` — recoverable / fallback fired
  - `Error` — request failed despite fallbacks
- NEVER log secrets, raw API keys, or full prompts. Truncate body samples
  to 256 bytes max.

---

## 5. Context discipline

- `context.Context` is **always the first parameter**, named `ctx`.
- Never store a context in a struct.
- Derive children with `context.WithTimeout`/`WithCancel`, never use
  `context.Background()` mid-request.
- For SSE streams: bind to `r.Context()`; on `<-ctx.Done()` close upstream,
  stop writes.

---

## 6. Concurrency

- Every `go func()` must satisfy one of:
  1. Has a `defer wg.Done()` and `wg.Wait()` somewhere upstream, OR
  2. Selects on `<-ctx.Done()` for shutdown, OR
  3. Has a written justification comment for fire-and-forget (rare).
- No global maps without `sync.Mutex` or `sync.Map`.
- Background workers (cloud sync, token refresh ticker) live in dedicated
  packages with `New(...) *Worker` + `Worker.Start(ctx) error` + `Worker.Stop()`.

---

## 7. Streaming (SSE)

- Use `internal/stream/flusher.go`'s `FlushWriter`. Never call
  `w.(http.Flusher).Flush()` directly in handlers.
- Copy upstream → downstream via `io.Copy` with periodic flush; do not
  `ioutil.ReadAll`.
- Translators that mutate the stream operate as `io.Reader` middlewares
  (chunk-in, chunk-out). Never accumulate full payloads.

---

## 8. JSON & translation

- Default `encoding/json` is fine. Reach for `bytedance/sonic` ONLY in the
  translator hot path AND only after a benchmark proves the win.
- Use `json.RawMessage` to pass through fields you don't need to read.
  Do NOT round-trip-parse arbitrary provider blobs.
- Translators must be pure functions: `func(req []byte) ([]byte, error)`.
  No side effects, no I/O, no globals.

---

## 9. Database

- All queries go through `sqlc`-generated functions in `internal/storage/`.
- Migrations are forward-only in production; only test code uses `.down.sql`.
- Use `db.QueryRowContext` / `db.ExecContext` — context-less variants are banned by lint.
- One `*sql.DB` per process, shared. SQLite WAL mode is enabled in
  `internal/storage/db.go`.
- All OAuth tokens + refresh tokens stored encrypted via `internal/auth/crypto.go`
  (AES-256-GCM, key from env `MASTER_KEY`).

---

## 10. HTTP handlers

- Handlers are thin: parse input → call service → write output.
- Business logic lives in `internal/router/*` or sub-packages, not in
  `internal/api/*`.
- Decode with `json.NewDecoder(r.Body).DisallowUnknownFields()` for admin
  endpoints. Do NOT use `DisallowUnknownFields` on `/v1/*` — upstream
  clients send extra fields.
- Always set `Content-Type` before writing body.
- Always `defer r.Body.Close()`.

---

## 11. Tests

- Table-driven where possible.
- Golden files for translators, RTK, caveman. Update with
  `go test ./... -update`.
- Integration tests in `tests/integration/` spin up an `httptest.Server`
  and an in-memory `:memory:` SQLite.
- Mock upstream providers with `httptest.NewServer` + recorded fixtures.
- Coverage target: 80% for `translator`, `rtk`, `caveman`. 60% elsewhere.
- Race detector ON in CI (`go test -race`).

---

## 12. Linting

`.golangci.yml` enforces:
- `errcheck`, `govet`, `staticcheck`, `gosec`, `goimports`, `revive`,
  `gocritic`, `bodyclose`, `noctx`, `gosimple`, `unused`, `ineffassign`.

Run locally: `task lint` (or `golangci-lint run`).

---

## 13. Anti-patterns (auto-rejected in review)

1. ❌ Buffering an SSE stream in `bytes.Buffer` before forwarding.
2. ❌ `panic()` in request path (panics belong only in `init()` for fatal config).
3. ❌ `time.Sleep` in production code (use `time.NewTicker` + context).
4. ❌ Importing `internal/api/*` from `internal/router/*` (wrong direction).
5. ❌ Adding a new core dependency without an ADR entry in `PLAN.md`.
6. ❌ Hand-written JSON marshalling for types we could codegen from OpenAPI.
7. ❌ `interface{}` / `any` in public APIs — use concrete types or generics.
8. ❌ Storing context in a struct field.
9. ❌ Calling integration tools, network, or DB from a translator/RTK filter.
10. ❌ Adding a 9router quirk without a `// ref:` comment pointing to the
   exact `.js:LINE`.

---

## 14. Commit style

Conventional Commits, in present imperative:

```
feat(rtk): port grep filter

Ports open-sse/rtk/filters/grep.js → internal/rtk/filters/grep.go.

- Includes alias 'rg' (matches 9router registry.js:24).
- Adds golden tests for 4 captured fixtures.
- Updates PARITY_CHECKLIST.md.
```

Allowed prefixes: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`, `ci`.
Scopes: `api`, `router`, `translator`, `executor`, `rtk`, `caveman`, `auth`,
`storage`, `stream`, `web`, `infra`.

---

## 15. When you find a 9router quirk you disagree with

Port it anyway. Add a `// quirk:` comment explaining the surprise, then
file an issue tagged `parity-divergence-candidate`. Do NOT silently fix it.
