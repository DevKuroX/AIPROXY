package chat

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/stream"
)

// StreamingHandler handles streaming chat completions
// ref: open-sse/handlers/chatCore/streamingHandler.js
type StreamingHandler struct {
	// Dependencies would be injected here
}

// HandleStreaming handles streaming chat completion requests
func (h *StreamingHandler) HandleStreaming(
	ctx context.Context,
	w http.ResponseWriter,
	upstreamResp *http.Response,
	config *stream.StreamConfig,
) error {
	// Create SSE writer
	sseWriter := stream.NewSSEWriter(w)
	
	// Create SSE reader from upstream response
	sseReader := stream.NewSSEReader(upstreamResp.Body)
	defer sseReader.Close()
	
	// Create relay
	relay := stream.NewRelay(sseReader)
	relay.AddWriter(sseWriter)
	
	// Start relay with context
	return relay.Start(ctx)
}

// HandleStreamingWithTransform handles streaming with format transformation
// ref: open-sse/handlers/chatCore/streamingHandler.js - buildTransformStream
func (h *StreamingHandler) HandleStreamingWithTransform(
	ctx context.Context,
	w http.ResponseWriter,
	upstreamResp *http.Response,
	config *stream.StreamConfig,
) error {
	// Check if translation is needed
	needsTranslation := config.SourceFormat != config.TargetFormat
	
	if !needsTranslation {
		// Passthrough mode
		return h.HandleStreaming(ctx, w, upstreamResp, config)
	}
	
	// Translation mode - would integrate with translator package
	// For now, implement basic passthrough
	return h.HandleStreaming(ctx, w, upstreamResp, config)
}

// HandleStreamingError handles streaming errors
// ref: open-sse/utils/error.js - createErrorResult
func (h *StreamingHandler) HandleStreamingError(
	w http.ResponseWriter,
	message string,
	errorType string,
	code int,
) error {
	sseWriter := stream.NewSSEWriter(w)
	
	errorChunk := stream.CreateErrorChunk(message, errorType, code)
	return sseWriter.WriteChunk(context.Background(), errorChunk)
}

// StreamConfigFromRequest creates StreamConfig from HTTP request
// ref: open-sse/handlers/chatCore/requestDetail.js - extractRequestConfig
func StreamConfigFromRequest(
	r *http.Request,
	provider string,
	model string,
	sourceFormat stream.StreamFormat,
	targetFormat stream.StreamFormat,
) *stream.StreamConfig {
	return &stream.StreamConfig{
		Provider:         provider,
		Model:            model,
		SourceFormat:     sourceFormat,
		TargetFormat:     targetFormat,
		UserAgent:        r.UserAgent(),
		ConnectionID:     r.Header.Get("X-Connection-ID"),
		APIKey:           r.Header.Get("Authorization"),
		RequestStartTime: time.Now(),
		Stream:           true,
	}
}

// IsStreamingRequest checks if request is streaming
func IsStreamingRequest(r *http.Request) bool {
	// Check stream parameter in body
	// This would need to parse the request body
	// For now, check header
	return r.Header.Get("Accept") == "text/event-stream"
}

// WriteStreamingHeaders writes SSE headers
func WriteStreamingHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
}

// PipeStream pipes data from reader to writer with context
func PipeStream(ctx context.Context, dst io.Writer, src io.Reader) error {
	return stream.PipeWithContext(ctx, dst, src)
}
