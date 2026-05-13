# docs/OBSERVABILITY.md

> **Status:** scaffold.

## Logging
- `slog` JSON handler to stdout.
- Level via env `LOG_LEVEL` (`debug|info|warn|error`).
- One request log line at INFO per request:
  ```
  {ts, level, msg:"request", request_id, method, path, status, duration_ms,
   provider_id, model, attempt, account_id, input_tokens, output_tokens}
  ```

## Request body logging
- Disabled by default.
- Toggle via `REQUEST_LOG_ENABLED=true`.
- Writes to `${DATA_DIR}/requests.jsonl`, gzip-rotated daily.
- Truncates body samples to 1KB.
- Never logs `Authorization` headers; redacts known secret fields.

## Metrics (phase 9)
- Prometheus at `/metrics` (admin-auth gated).
- Counters: `requests_total{provider,status}`, `tokens_total{kind,provider}`.
- Histogram: `request_duration_seconds{provider}`.
- Gauges: `active_streams`, `provider_nodes_unavailable`.

## Tracing (optional, phase 9)
- OpenTelemetry SDK; OTLP exporter via env.
- Spans: HTTP request → translator → executor → upstream.

## pprof
- `/debug/pprof/*` mounted when `PPROF=true` env set. Admin-auth gated.
