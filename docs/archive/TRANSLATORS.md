# docs/TRANSLATORS.md — Format Conversion Layer

> **Status:** scaffold. Fill in per-pair sections as you port from
> `_ref/9router/open-sse/translator/`.

## Purpose

The translator layer converts a request from one provider format to another so
the same upstream executor can handle it. Translators are **pure functions**:
they take `[]byte` in, return `[]byte` out, and never do I/O.

## Detection (ref: open-sse/translator/index.js)

`translator.Detect(endpoint, body) (Format, error)` returns one of the
`FORMATS` enum values. Detection considers:
1. Endpoint path first (`/v1/messages` → claude, `/v1/chat/completions` → openai, etc.)
2. Body shape (presence of `system` field, `tools` schema, `inlineData`, etc.)

## Registry shape

```go
type Pair struct {
    Request   func(in []byte) ([]byte, error)
    Response  func(in []byte) ([]byte, error)
    Stream    func() stream.Transform // returns a fresh middleware per stream
}

var registry = map[FormatPair]Pair{
    {Src: FormatOpenAI, Dst: FormatClaude}: { Request: requestOpenAIToClaude, ... },
    ...
}
```

## Per-pair sections (TO BE WRITTEN)

For each pair below, document:
- Source 9router file(s)
- Quirks (e.g. Anthropic-Beta header handling, tool_use blocks)
- Stream-event mapping table

### OpenAI ↔ Claude
**ref:** `open-sse/translator/request/openaiToClaude.js`, `response/claudeToOpenai.js`

TODO:
- [ ] Message walk
- [ ] tool_calls ↔ tool_use mapping
- [ ] system prompt extraction
- [ ] image / multimodal content mapping
- [ ] stream chunk mapping (`message_start`, `content_block_delta`, `message_delta`, `message_stop`)

### OpenAI ↔ Gemini
**ref:** translator/request/openaiToGemini.js, response/geminiToOpenai.js
TODO.

### OpenAI ↔ Kiro
**ref:** translator/request/openaiToKiro.js, response/kiroToOpenai.js
TODO.

### OpenAI ↔ Cursor
**ref:** translator/request/openaiToCursor.js, response/cursorToOpenai.js
> Cursor uses **protobuf**, not JSON. See `open-sse/utils/cursorProtobuf.js`.
TODO.

### OpenAI ↔ CommandCode
**ref:** translator/request/openaiToCommandCode.js, response/commandcodeToOpenai.js
TODO.

### OpenAI ↔ Vertex
TODO.

### OpenAI ↔ Ollama
TODO.

### Antigravity → OpenAI
TODO.

### OpenAI Responses (`/v1/responses`)
**ref:** `open-sse/transformer/responsesTransformer.js`
TODO.

## Testing

Every translator must have golden fixtures in `testdata/translator/<pair>/`.
Capture via `scripts/capture_fixtures.sh` against a running 9router.
