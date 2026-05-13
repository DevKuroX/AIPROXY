package stream

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// SSEReader implements StreamReader for parsing SSE streams
// ref: open-sse/utils/stream.js - parseSSELine, createSSEReader
type SSEReader struct {
	scanner *bufio.Scanner
	closed  bool
}

// NewSSEReader creates a new SSE reader from an io.Reader
func NewSSEReader(r io.Reader) *SSEReader {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 1024*1024) // 1MB max line
	return &SSEReader{
		scanner: scanner,
		closed:  false,
	}
}

// ReadChunk reads the next SSE chunk from the stream
// ref: open-sse/utils/stream.js - parseSSELine
func (r *SSEReader) ReadChunk(ctx context.Context) (*StreamChunk, error) {
	if r.closed {
		return nil, io.EOF
	}

	var chunk StreamChunk
	var dataBuilder strings.Builder
	var inDataBlock bool

	for r.scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := r.scanner.Text()
		if line == "" {
			// Empty line indicates end of SSE event
			if inDataBlock && dataBuilder.Len() > 0 {
				chunk.Data = []byte(dataBuilder.String())
				return &chunk, nil
			}
			continue
		}

		// Parse SSE line
		if strings.HasPrefix(line, "data: ") {
			inDataBlock = true
			data := strings.TrimPrefix(line, "data: ")
			if dataBuilder.Len() > 0 {
				dataBuilder.WriteString("\n")
			}
			dataBuilder.WriteString(data)
		} else if strings.HasPrefix(line, "event: ") {
			chunk.Event = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "id: ") {
			chunk.ID = strings.TrimPrefix(line, "id: ")
		} else if strings.HasPrefix(line, "retry: ") {
			retryStr := strings.TrimPrefix(line, "retry: ")
			// Parse retry as integer, ignore errors
			var retry int
			_, _ = fmt.Sscanf(retryStr, "%d", &retry)
			chunk.Retry = retry
		}
		// Ignore other lines (comments, etc.)
	}

	// Handle scanner error or EOF
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	// Return last chunk if we have data
	if inDataBlock && dataBuilder.Len() > 0 {
		chunk.Data = []byte(dataBuilder.String())
		return &chunk, nil
	}

	return nil, io.EOF
}

// Close closes the reader
func (r *SSEReader) Close() error {
	r.closed = true
	return nil
}

// IsDoneChunk checks if a chunk is a [DONE] marker
// ref: open-sse/utils/streamHelpers.js - hasValuableContent
func IsDoneChunk(chunk *StreamChunk) bool {
	if chunk == nil || len(chunk.Data) == 0 {
		return false
	}
	data := string(chunk.Data)
	return strings.TrimSpace(data) == "[DONE]"
}

// ParseSSELine parses a single SSE line (helper function)
// ref: open-sse/utils/streamHelpers.js - parseSSELine
func ParseSSELine(line string) (field, value string, ok bool) {
	if line == "" {
		return "", "", false
	}

	// Skip comments
	if strings.HasPrefix(line, ":") {
		return "", "", false
	}

	// Find first colon
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		// No colon, treat as field with empty value
		return strings.TrimSpace(line), "", true
	}

	field = strings.TrimSpace(line[:colonIndex])
	value = strings.TrimSpace(line[colonIndex+1:])
	
	// Handle single space after colon (SSE spec)
	if len(line) > colonIndex+1 && line[colonIndex+1] == ' ' {
		value = strings.TrimSpace(line[colonIndex+2:])
	}

	return field, value, true
}

// DetectStreamFormat detects the streaming format from response headers and data
// ref: open-sse/utils/streamHelpers.js - detectFormat
func DetectStreamFormat(contentType string, firstChunk *StreamChunk) StreamFormat {
	contentType = strings.ToLower(contentType)
	
	// Check content type first
	if strings.Contains(contentType, "text/event-stream") {
		// Check chunk data for format hints
		if firstChunk != nil && len(firstChunk.Data) > 0 {
			data := string(firstChunk.Data)
			if strings.Contains(data, "claude") {
				return FormatCL4ude
			} else if strings.Contains(data, "gemini") {
				return FormatGemini
			} else if strings.Contains(data, "antigravity") {
				return FormatAntigravity
			}
		}
		return FormatOpenAI
	}
	
	// Non-SSE responses
	if strings.Contains(contentType, "application/json") {
		// Would need to parse JSON to determine format
		return FormatOpenAI
	}
	
	return FormatOpenAI // Default
}
