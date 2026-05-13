# Chat Handlers & Stream Core

## Overview

This document describes the chat handlers and streaming core implementation for AIPROXY, ported from 9router's `open-sse/handlers/chatCore/`.

## Architecture

### Core Components

1. **Stream Types** (`internal/stream/types.go`)
   - `StreamChunk`: Represents a single SSE chunk
   - `StreamReader`/`StreamWriter`: Interfaces for stream processing
   - `StreamFormat`: Enum for different streaming formats (0penAI, CL4ude, Gemini, etc.)

2. **SSE Reader** (`internal/stream/sse_reader.go`)
   - Parses SSE streams into `StreamChunk` objects
   - Handles multi-line data, events, IDs, and retry fields
   - Context-aware reading with cancellation support

3. **SSE Writer** (`internal/stream/writer.go`)
   - Writes `StreamChunk` objects as SSE streams
   - Implements `http.Flusher` for real-time streaming
   - Handles error formatting in SSE format

4. **Stream Relay** (`internal/stream/relay.go`)
   - Relays streams from readers to multiple writers
   - Supports tee functionality for logging/monitoring
   - Context cancellation and goroutine cleanup

5. **Stream Helpers** (`internal/stream/helpers.go`)
   - Format detection and normalization
   - Usage extraction from streams
   - Error chunk creation

### Handlers

1. **Streaming Handler** (`internal/handlers/chat/streaming.go`)
   - Handles real-time SSE streaming
   - Integrates with stream relay for efficient forwarding
   - Context cancellation on client disconnect

2. **Non-Streaming Handler** (`internal/handlers/chat/non_streaming.go`)
   - Handles buffered JSON responses
   - Converts SSE streams to single JSON responses when needed
   - Error handling in JSON format

3. **SSE-to-JSON Converter** (`internal/handlers/chat/sse_to_json.go`)
   - Converts SSE streams to complete JSON responses
   - Merges chunked content and extracts usage
   - Supports multiple streaming formats

### Storage

1. **Request Details** (`internal/storage/request_details.go`)
   - Stores detailed request logs for debugging
   - Tokens usage, response times, errors
   - Queryable with filters

2. **Database Migration** (`internal/storage/migrations/011_request_details.sql`)
   - PostgreSQL schema for request details
   - Indexed for performance

## Usage

### Streaming Request

```go
handler := chat.NewStreamingHandler()
config := chat.StreamConfigFromRequest(r, "openai", "gpt-4", 
    stream.Format0penAI, stream.Format0penAI)
err := handler.HandleStreaming(ctx, w, upstreamResp, config)
```

### Non-Streaming Request

```go
handler := chat.NewNonStreamingHandler()
config := chat.StreamConfigFromRequest(r, "openai", "gpt-4",
    stream.Format0penAI, stream.Format0penAI)
err := handler.HandleNonStreaming(ctx, w, upstreamResp, config)
```

### SSE-to-JSON Conversion

```go
converter := chat.NewSSEToJSONConverter()
jsonResponse, usage, err := converter.Convert(ctx, sseReader, stream.Format0penAI)
```

## Error Handling

### Streaming Errors
Streaming errors are formatted as SSE chunks:
```json
{"error": {"message": "...", "type": "...", "code": ...}}
```

### Non-Streaming Errors
Non-streaming errors are formatted as JSON responses:
```json
{"error": {"message": "...", "type": "...", "code": ...}}
```

## Performance Considerations

1. **No Buffering**: Streams are relayed chunk-by-chunk without buffering entire responses
2. **Context Cancellation**: Upstream requests are cancelled immediately on client disconnect
3. **Memory Efficiency**: Uses `io.Copy` and chunked reads for large streams
4. **Concurrent Relaying**: Multiple writers can receive the same stream concurrently

## Testing

Run tests with:
```bash
go test ./internal/stream/...
go test ./internal/handlers/chat/...
```

## Reference

- **9router Source**: `_ref/9router/open-sse/handlers/chatCore/`
- **Stream Utilities**: `_ref/9router/open-sse/utils/stream.js`
- **Error Formatting**: `_ref/9router/open-sse/utils/error.js`

## Status

✅ Core streaming infrastructure implemented
✅ SSE reader/writer with 9router parity
✅ Stream relay with tee functionality
✅ Streaming and non-streaming handlers
✅ SSE-to-JSON converter
✅ Request details storage
✅ Basic tests passing

🚧 Pending: Integration with translator layer
🚧 Pending: Format-specific normalization
🚧 Pending: Comprehensive integration tests
