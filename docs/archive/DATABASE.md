# docs/DATABASE.md — Schema & Migrations

> **Status:** scaffold.

## Engine

- **SQLite** via `modernc.org/sqlite` (pure-Go, no CGO).
- WAL mode + `PRAGMA synchronous=NORMAL` + `foreign_keys=ON`.
- Single DB file at `${DATA_DIR}/9rgo.db`. **No separate `usageDb`** — fixes a
  9router anti-pattern.

## Migrations

- `internal/storage/migrations/NNN_<name>.up.sql` / `.down.sql`.
- Forward-only in production; `.down.sql` exists only for tests.
- Migration runner: `golang-migrate` (file source, sqlite driver).

## Schema (initial)

```sql
-- 001_init.up.sql

CREATE TABLE users (
    id            INTEGER PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL          -- JSON
);

CREATE TABLE providers (
    id            TEXT PRIMARY KEY,    -- e.g. "anthropic", "openai", "cursor"
    name          TEXT NOT NULL,
    base_url      TEXT,
    is_specialized INTEGER NOT NULL DEFAULT 0,
    enabled       INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE provider_nodes (
    id           INTEGER PRIMARY KEY,
    provider_id  TEXT NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    label        TEXT NOT NULL,
    credentials  BLOB NOT NULL,        -- AES-256-GCM ciphertext (auth/crypto.go)
    available    INTEGER NOT NULL DEFAULT 1,
    last_used_at DATETIME
);

CREATE TABLE api_keys (
    id            INTEGER PRIMARY KEY,
    key_hash      TEXT NOT NULL UNIQUE,  -- HMAC-SHA256(api_key, MASTER_KEY)
    name          TEXT NOT NULL,
    scopes        TEXT NOT NULL DEFAULT '*',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at    DATETIME
);

CREATE TABLE model_aliases (
    alias       TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL REFERENCES providers(id),
    model       TEXT NOT NULL
);

CREATE TABLE combos (
    name        TEXT PRIMARY KEY,
    members     TEXT NOT NULL          -- JSON array [{provider, model}, ...]
);

CREATE TABLE pricing (
    provider_id TEXT NOT NULL,
    model       TEXT NOT NULL,
    input_per_1k  REAL NOT NULL,
    output_per_1k REAL NOT NULL,
    PRIMARY KEY (provider_id, model)
);

CREATE TABLE usage (
    id            INTEGER PRIMARY KEY,
    ts            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    request_id    TEXT NOT NULL,
    api_key_id    INTEGER REFERENCES api_keys(id),
    provider_id   TEXT NOT NULL,
    model         TEXT NOT NULL,
    input_tokens  INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,
    cost_usd      REAL NOT NULL,
    duration_ms   INTEGER NOT NULL,
    status        INTEGER NOT NULL
);

CREATE INDEX idx_usage_ts            ON usage(ts);
CREATE INDEX idx_usage_provider      ON usage(provider_id);
CREATE INDEX idx_usage_key           ON usage(api_key_id);
CREATE INDEX idx_provider_nodes_pid  ON provider_nodes(provider_id);
```

## sqlc

- `internal/storage/queries/*.sql` — raw queries grouped by domain.
- `sqlc.yaml` at repo root generates `*.sql.gen.go` next to each `.sql`.
- Run `task gen` to regenerate.

## Encryption-at-rest

- AES-256-GCM. Key = first 32 bytes of `SHA256(env MASTER_KEY)`.
- Nonce: random 12 bytes prepended to ciphertext.
- Helpers in `internal/auth/crypto.go`.
- Applied to: `provider_nodes.credentials`, OAuth refresh tokens.
