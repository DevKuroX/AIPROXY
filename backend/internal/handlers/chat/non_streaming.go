package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/stream"
)

// NonStreamingHandler handles non-streaming chat completions
// ref: open-sse/handlers/chatCore/nonStreamingHandler.js
type NonStreamingHandler struct {
	// Dependencies would be injected here
}

// HandleNonStreaming handles non-streaming chat completion requests
func (h *NonStreamingHandler) HandleNonStreaming(
	ctx context.Context,
	w http.ResponseWriter,
	upstreamResp *http.Response,
	config *stream.StreamConfig,
) error {
	// Read entire response
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, upstreamResp.Body); err != nil {
		return err
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(upstreamResp.StatusCode)
	_, err := w.Write(buf.Bytes())
	return err
}

// HandleNonStreamingError handles non-streaming errors
func (h *NonStreamingHandler) HandleNonStreamingError(
	w http.ResponseWriter,
	message string,
	errorType string,
	code int,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    errorType,
			"code":    code,
		},
	}

	return json.NewEncoder(w).Encode(errorResp)
}

// ConvertSSEToJSON converts SSE stream to single JSON response
// ref: open-sse/handlers/chatCore/sseToJsonHandler.js
func (h *NonStreamingHandler) ConvertSSEToJSON(
	ctx context.Context,
	upstreamResp *http.Response,
	config *stream.StreamConfig,
) ([]byte, error) {
	// Check if response is SSE
	contentType := upstreamResp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		// Not SSE, return as-is
		return io.ReadAll(upstreamResp.Body)
	}

	// Create SSE reader
	sseReader := stream.NewSSEReader(upstreamResp.Body)
	defer sseReader.Close()

	// Buffer stream chunks
	chunks, err := stream.BufferStream(sseReader)
	if err != nil {
		return nil, err
	}

	// Extract usage
	usage, _ := stream.ExtractUsageFromStream(chunks)

	// Build final response based on format
	return h.buildFinalResponse(chunks, usage, config)
}

// buildFinalResponse builds final JSON response from chunks
func (h *NonStreamingHandler) buildFinalResponse(
	chunks []*stream.StreamChunk,
	usage *stream.Usage,
	config *stream.StreamConfig,
) ([]byte, error) {
	// This is a simplified implementation
	// In real implementation, would need to merge chunks based on format
	
	var response map[string]interface{}
	
	// Use last non-[DONE] chunk as base
	for i := len(chunks) - 1; i >= 0; i-- {
		chunk := chunks[i]
		if stream.IsDoneChunk(chunk) {
			continue
		}
		
		if len(chunk.Data) > 0 {
			if err := json.Unmarshal(chunk.Data, &response); err == nil {
				break
			}
		}
	}

	// Add usage if available
	if usage != nil && response != nil {
		response["usage"] = usage
	}

	return json.Marshal(response)
}

// IsNonStreamingRequest checks if request is non-streaming
func IsNonStreamingRequest(r *http.Request) bool {
	// Check stream parameter in body
	// For now, assume non-streaming if not explicitly streaming
	return !IsStreamingRequest(r)
}

// WriteNonStreamingHeaders writes JSON headers
func WriteNonStreamingHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// MergeResponses merges multiple responses (for fallback scenarios)
func MergeResponses(responses [][]byte) ([]byte, error) {
	if len(responses) == 0 {
		return nil, io.EOF
	}
	
	// For now, return first successful response
	// In real implementation, would need to merge based on format
	return responses[0], nil
}
