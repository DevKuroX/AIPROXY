# docs/AUTH.md — JWT, API Keys, OAuth Flows

> **Status:** scaffold.

## Three auth surfaces

| Surface | Used by | Mechanism |
|---|---|---|
| Dashboard `/api/*` | browser (Next.js) | `httpOnly` cookie JWT (HS256) |
| Gateway `/v1/*` and `/v1beta/*` | CLI tools, scripts | API key (HMAC-validated) |
| Outbound (to providers) | the gateway itself | per-provider OAuth or static keys |

## Dashboard JWT

- HS256, key = env `JWT_SECRET` (auto-generated to DB on first boot).
- Claims: `sub` (user id), `iat`, `exp` (24h default), `roles`.
- Stored in cookie: `Path=/; HttpOnly; SameSite=Strict; Secure` (when HTTPS).
- Middleware: `internal/api/middleware/auth_jwt.go`.

## API keys (`/v1/*` gate)

- Format: `9rk_<base64url(32 random bytes)>`.
- Stored as `HMAC-SHA256(plaintext, MASTER_KEY)` in `api_keys.key_hash`.
- Middleware `auth_apikey.go` reads `Authorization: Bearer <key>`, hashes,
  looks up. Falls through 401 on miss.
- Optional rate limit per key in `middleware/ratelimit.go`.

## Outbound OAuth — overview

Implemented in `internal/auth/oauth/`. One file per provider, all conforming to:

```go
type Flow interface {
    Start(ctx context.Context) (Challenge, error)         // returns user-facing code / URL
    Poll(ctx context.Context, ch Challenge) (Tokens, error) // for device flows
    Callback(ctx context.Context, q url.Values) (Tokens, error) // for redirect flows
    Refresh(ctx context.Context, refreshToken string) (Tokens, error)
}
```

Supported providers (ref: `_ref/9router/src/lib/oauth/services/`):
- `claude` — device code (Anthropic)
- `gemini` — OAuth 2.0 (Google)
- `codex` — device code
- `github` — device code (Copilot)
- `kiro` — OAuth 2.0
- `cursor` — custom (use `cursorChecksum`)
- `antigravity` — JWT session
- `qwen` — OAuth 2.0
- `iflow` — cookie/redirect
- `qoder` — OAuth 2.0
- `openai` — ChatGPT login

> NOTE: `opencode` is NOT an OAuth flow. It is an executor that consumes a
> pre-issued token. Earlier drafts of ARCHITECTURE.md listed it under
> auth/oauth/ — that was wrong and has been corrected.

## Token storage

- `provider_nodes.credentials` BLOB column.
- AES-256-GCM via `auth/crypto.go`.
- Refresh hook runs in `internal/router/token_refresh.go`:
  - Pre-call: refresh if `expiresAt - now < 60s`.
  - On 401 with `WWW-Authenticate` or known body shape: refresh once + retry.

## Logout & revocation

- `POST /api/auth/logout` clears cookie.
- `POST /api/keys/:id/revoke` sets `revoked_at` (key never matches again).
- No JWT denylist (short expiry + revocable refresh).
