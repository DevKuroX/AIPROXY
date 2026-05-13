# docs/FALLBACK.md — Fallback, Combos, Multi-Account

> **Status:** scaffold. Port from `open-sse/services/accountFallback.js`,
> `open-sse/services/combo.js`.

## Two levels

### 1. Account-level (within one provider+model)
If a provider has multiple `provider_nodes` (multiple accounts/keys), the
router picks one, marks it `unavailable` on hard failure, retries with the
next. Refresh hook may bring it back to `available`.

State machine per node:
```
available ── 401/403/429 ─▶ unavailable
unavailable ── refresh ok ─▶ available
unavailable ── after cooldown ─▶ available  (probe)
```

Cooldown: 60s default, configurable in settings.

### 2. Combo-level (across multiple provider+model entries)
Models can be aliased to a "combo" like `combo:fast-cheap` which expands to
an ordered list `[anthropic/haiku, openai/gpt-4o-mini, ...]`. The router
iterates that list; for each entry it does the account-level dance above.

## Stop conditions
- Successful 2xx response → return immediately.
- 4xx that is NOT 401/403/429 → return error to client (don't retry).
- 5xx → next attempt.
- All attempts exhausted → return last error in OpenAI error envelope.

## Per-attempt logging
One `slog.Info` line per attempt with:
`request_id`, `attempt`, `provider_id`, `account_id`, `model`, `status`,
`duration_ms`, `decision` (`return|retry_account|retry_combo|fail`).

## Ports from 9router
| 9router file | Go target |
|---|---|
| `open-sse/services/accountFallback.js` | `internal/router/fallback.go` |
| `open-sse/services/combo.js` | `internal/router/combo.go` |
| `open-sse/services/tokenRefresh.js` | `internal/router/token_refresh.go` |
