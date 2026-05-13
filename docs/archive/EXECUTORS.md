# docs/EXECUTORS.md — Per-Provider HTTP Adapters

> **Status:** scaffold.

## Interface

```go
type Executor interface {
    // Execute sends the (already-translated, RTK-applied, Caveman-injected)
    // request to the upstream provider and returns either the full response
    // body (non-stream) or an io.ReadCloser (stream).
    Execute(ctx context.Context, req *Request) (*Response, error)
}
```

## Registry (ref: open-sse/executors/index.js)

```go
var registry = map[string]Executor{
    "antigravity":   &Antigravity{},
    "azure":         &Azure{},
    "gemini-cli":    &GeminiCLI{},
    "github":        &Github{},
    "iflow":         &IFlow{},
    "qoder":         &Qoder{},
    "kiro":          &Kiro{},
    "codex":         &Codex{},
    "cursor":        &Cursor{},
    "cu":            &Cursor{},  // alias (ref: index.js:26)
    "vertex":        NewVertex("vertex"),
    "vertex-partner": NewVertex("vertex-partner"),
    "qwen":          &Qwen{},
    "opencode":      &OpenCode{},
    "opencode-go":   &OpenCodeGo{},
    "grok-web":      &GrokWeb{},
    "perplexity-web": &PerplexityWeb{},
    "ollama-local":  &OllamaLocal{},
    "commandcode":   &CommandCode{},
}

func Get(providerID string) Executor {
    if ex, ok := registry[providerID]; ok { return ex }
    return NewDefault(providerID) // generic OpenAI-compat fallback
}
```

> ⚠️ There is **no** standalone `claude` executor in 9router. Claude is
> served via `DefaultExecutor` after the translator converts to OpenAI shape
> and back (or via direct passthrough when `/v1/messages` is used).

## Per-executor notes (TODO)

### default.go
generic OpenAI-compatible. Reads `baseURL`, `apiKey`, applies bearer auth.
ref: `open-sse/executors/default.js`

### cursor.go
**Special:** uses `cursorChecksum` + protobuf. Port:
- `open-sse/utils/cursorChecksum.js`
- `open-sse/utils/cursorProtobuf.js`

### codex.go
Uses Codex device-code OAuth tokens. Reads instructions from
`open-sse/config/codexInstructions.js`.

### antigravity.go
Has a header cache (`claudeHeaderCache.js`) and JWT-based session.

### kiro.go / iflow.go
OAuth flows live in `auth/oauth/`. Executor reads access token from storage.

(Continue for every executor as you port them.)

## Common helpers (`executor/base.go`)

Port of `open-sse/executors/base.js`:
- `BuildRequest(ctx, url, method, headers, body) *http.Request`
- `DoStream(ctx, req) (io.ReadCloser, error)` — sets keep-alive, no redirects, no buffer
- `DoJSON(ctx, req) ([]byte, error)`
- Error mapping helpers (429, 401 with `WWW-Authenticate`, etc.)
