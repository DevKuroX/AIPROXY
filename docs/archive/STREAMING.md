# docs/STREAMING.md — SSE Internals

> **Status:** scaffold.

## Wire format
Server-Sent Events. Each event:
```
data: <utf-8 chunk>\n\n
```
Some upstreams use `event: <name>\ndata: ...` — preserve event names in
passthrough mode.

## Flusher
`internal/stream/flusher.go` wraps `http.ResponseWriter`:
- Sets `Content-Type: text/event-stream`, `Cache-Control: no-cache`,
  `Connection: keep-alive`, `X-Accel-Buffering: no`.
- Disables compression.
- `Write` writes + `Flush` on each call.

## Proxy loop
`internal/stream/proxy.go`:
```go
func Proxy(ctx context.Context, dst io.Writer, src io.Reader) error {
    buf := make([]byte, 4096)
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        n, err := src.Read(buf)
        if n > 0 {
            if _, werr := dst.Write(buf[:n]); werr != nil { return werr }
            if f, ok := dst.(http.Flusher); ok { f.Flush() }
        }
        if err != nil { return err }
    }
}
```

## Translator middlewares
Translators that mutate streams expose `stream.Transform`:
```go
type Transform interface {
    // Process is called once per upstream chunk; may emit 0..N downstream chunks.
    Process(chunk []byte, emit func([]byte) error) error
    Flush(emit func([]byte) error) error
}
```

## Disconnect detection
If client closes the connection, `r.Context().Done()` fires. The proxy loop
returns; we propagate cancellation to the upstream request via the
HTTP request context.

## Tests
- Unit: feed scripted byte sequences through `Process` / `Flush`, assert outputs.
- Integration: spawn an `httptest.Server` that streams chunks with delays,
  assert end-to-end timing tolerance + correctness.
- Load: `tests/load/sse_concurrent.js` (k6) — 5k concurrent streams under
  1GB memory.
