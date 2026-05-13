package stream

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// SSEWriter implements StreamWriter for writing SSE streams
// ref: open-sse/utils/streamHandler.js - createSSEWriter
type SSEWriter struct {
	writer  io.Writer
	flusher http.Flusher
	closed  bool
}

// NewSSEWriter creates a new SSE writer from an http.ResponseWriter
// ref: open-sse/utils/streamHandler.js - createSSEWriter
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	var flusher http.Flusher
	if f, ok := w.(http.Flusher); ok {
		flusher = f
	}

	return &SSEWriter{
		writer:  w,
		flusher: flusher,
		closed:  false,
	}
}

// WriteChunk writes an SSE chunk to the stream
// ref: open-sse/utils/streamHandler.js - writeChunk
func (w *SSEWriter) WriteChunk(ctx context.Context, chunk *StreamChunk) error {
	if w.closed {
		return io.ErrClosedPipe
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Write event field if present
	if chunk.Event != "" {
		if _, err := fmt.Fprintf(w.writer, "event: %s\n", chunk.Event); err != nil {
			return err
		}
	}

	// Write id field if present
	if chunk.ID != "" {
		if _, err := fmt.Fprintf(w.writer, "id: %s\n", chunk.ID); err != nil {
			return err
		}
	}

	// Write retry field if present
	if chunk.Retry > 0 {
		if _, err := fmt.Fprintf(w.writer, "retry: %d\n", chunk.Retry); err != nil {
			return err
		}
	}

	// Write data field (can be multi-line)
	data := string(chunk.Data)
	lines := splitLines(data)
	for _, line := range lines {
		if _, err := fmt.Fprintf(w.writer, "data: %s\n", line); err != nil {
			return err
		}
	}

	// End of event
	if _, err := fmt.Fprint(w.writer, "\n"); err != nil {
		return err
	}

	// Flush if we have a flusher
	if w.flusher != nil {
		w.flusher.Flush()
	}

	return nil
}

// Flush flushes any buffered data
// ref: open-sse/utils/streamHandler.js - flush
func (w *SSEWriter) Flush() error {
	if w.closed {
		return io.ErrClosedPipe
	}

	if w.flusher != nil {
		w.flusher.Flush()
	}
	return nil
}

// WriteDone writes the [DONE] marker for OpenAI compatibility
func (w *SSEWriter) WriteDone() error {
	if w.closed {
		return io.ErrClosedPipe
	}
	if _, err := fmt.Fprint(w.writer, "data: [DONE]\n\n"); err != nil {
		return err
	}
	if w.flusher != nil {
		w.flusher.Flush()
	}
	return nil
}

// Close closes the writer
func (w *SSEWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	if w.flusher != nil {
		w.flusher.Flush()
	}
	return nil
}

// WriteError writes an error as an SSE chunk
// ref: open-sse/utils/error.js - createErrorResult
func (w *SSEWriter) WriteError(ctx context.Context, err error) error {
	errorChunk := &StreamChunk{
		Event: "error",
		Data:  []byte(`{"error":{"message":"` + err.Error() + `","type":"api_error"}}`),
	}
	return w.WriteChunk(ctx, errorChunk)
}

// WriteRaw writes raw SSE data (for passthrough)
// ref: open-sse/utils/streamHandler.js - writeRaw
func (w *SSEWriter) WriteRaw(ctx context.Context, data []byte) error {
	if w.closed {
		return io.ErrClosedPipe
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, err := w.writer.Write(data); err != nil {
		return err
	}

	if w.flusher != nil {
		w.flusher.Flush()
	}
	return nil
}

// splitLines splits data into lines for SSE formatting
// ref: open-sse/utils/stream.js - splitLines
func splitLines(data string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}