# docs/USAGE.md — Usage Tracking & Pricing

> **Status:** scaffold.

## What we record
Every successful (and most failed) `/v1/*` request appends a row to `usage`:
- `request_id` (UUID v4)
- `api_key_id` (which client key)
- `provider_id`, `model` (resolved after combo expansion)
- `input_tokens`, `output_tokens` (parsed from upstream response)
- `cost_usd` (computed from `pricing` table)
- `duration_ms`
- `status` (HTTP code)

## Extraction
`internal/router/usage.go` reads token counts from:
- OpenAI: `usage.prompt_tokens` / `usage.completion_tokens`
- Claude: `usage.input_tokens` / `usage.output_tokens`
- Gemini: `usageMetadata.promptTokenCount` / `candidatesTokenCount`
- For streamed responses: aggregate from `[DONE]` frame or final usage event.

## Pricing
`pricing` table seeded from `open-sse/config/models.js`. Cost formula:
```
cost = (input_tokens / 1000) * input_per_1k + (output_tokens / 1000) * output_per_1k
```

## Dashboard endpoints
See `docs/API.md` → `/api/usage`.

## Retention
Default: keep raw rows 90 days, daily aggregates forever. Configurable in settings.

## Privacy
We do NOT store request/response bodies in `usage`. Bodies go to the
request log (`internal/observability/request_logger.go`) which is OFF by
default and writes to `${DATA_DIR}/requests.jsonl` with truncation.
