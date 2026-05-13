# docs/API.md — Endpoint Contracts

> **Status:** scaffold. The **source of truth** is `docs/openapi.yaml` once
> generated. This file is human-readable summary.

## Conventions
- All `/v1/*` endpoints are OpenAI-compatible. Auth via `Authorization: Bearer <api_key>`.
- All `/v1beta/*` endpoints are Gemini-compatible. Auth via `x-goog-api-key` header **or** Bearer.
- All `/api/*` endpoints require a session cookie (`httpOnly`, set by `/api/auth/login`).
- All responses use either OpenAI or Claude error envelopes depending on the endpoint family.

## /api/auth

| Method | Path | Body | Response |
|---|---|---|---|
| POST | `/api/auth/login` | `{username, password}` | sets `Set-Cookie`; body `{ok:true}` |
| POST | `/api/auth/logout` | — | clears cookie |
| GET  | `/api/auth/me` | — | `{username, roles}` |

## /api/settings

| Method | Path | Notes |
|---|---|---|
| GET | `/api/settings` | returns full settings object |
| PUT | `/api/settings` | upsert; partial JSON allowed |

## /api/providers

CRUD. Plus `POST /api/providers/:id/test` to ping upstream.

## /api/keys

CRUD for API keys gated to `/v1/*`. Keys are HMAC-signed and stored
encrypted-at-rest.

## /api/combos, /api/models/alias, /api/pricing

CRUD. See `docs/DATABASE.md` for schemas.

## /api/usage

| Method | Path | Notes |
|---|---|---|
| GET | `/api/usage/summary?from=&to=` | aggregated |
| GET | `/api/usage/by-model?from=&to=` | grouped |
| GET | `/api/usage/by-key?from=&to=` | grouped |
| GET | `/api/usage/stream` | SSE — live tail |

## /api/oauth/:provider

| Method | Path | Notes |
|---|---|---|
| POST | `/api/oauth/:provider/start` | returns `{deviceCode, userCode, verificationUrl}` |
| POST | `/api/oauth/:provider/poll` | returns tokens or `pending` |
| POST | `/api/oauth/:provider/callback` | for OAuth 2.0 redirect flows |

Supported providers: `claude`, `gemini`, `codex`, `github`, `kiro`,
`cursor`, `antigravity`, `qwen`, `iflow`, `qoder`, `openai`.

## /api/sync/cloud

| Method | Path | Notes |
|---|---|---|
| POST | `/api/sync/cloud` | trigger immediate sync |
| GET  | `/api/sync/cloud/status` | last sync info |

## /api/cli-tools

| Method | Path | Notes |
|---|---|---|
| GET | `/api/cli-tools` | enumerate detected CLIs (Claude Code, Codex, Cursor, Cline, Continue) |
| POST | `/api/cli-tools/:name/install` | writes the CLI's config file pointing at this gateway |

## /api/health, /api/version, /api/shutdown
Self-explanatory.

## /v1/*
OpenAI-compatible. See PARITY_CHECKLIST §1.

## /v1beta/*
Gemini-compatible. See PARITY_CHECKLIST §1.

## Embedded UI
`GET /dashboard/*` serves the Next.js static export from `embed.FS`.
