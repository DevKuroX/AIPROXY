package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/handlers/chat"
	"github.com/DevKuroX/AIPROXY/internal/stream"
)

// EnhancedChatHandler provides enhanced chat handling with new streaming infrastructure
type EnhancedChatHandler struct {
	streamingHandler    *chat.StreamingHandler
	nonStreamingHandler *chat.NonStreamingHandler
	sseToJSONConverter  *chat.SSEToJSONConverter
}

// NewEnhancedChatHandler creates a new enhanced chat handler
func NewEnhancedChatHandler() *EnhancedChatHandler {
	return &EnhancedChatHandler{
		streamingHandler:    &chat.StreamingHandler{},
		nonStreamingHandler: &chat.NonStreamingHandler{},
		sseToJSONConverter:  &chat.SSEToJSONConverter{},
	}
}

// HandleChatCompletionsEnhanced handles chat completions with enhanced streaming
func (h *EnhancedChatHandler) HandleChatCompletionsEnhanced(
	w http.ResponseWriter,
	r *http.Request,
	upstreamResp *http.Response,
	provider string,
	model string,
	sourceFormat stream.StreamFormat,
	targetFormat stream.StreamFormat,
) error {
	ctx := r.Context()
	
	// Create stream config
	config := chat.StreamConfigFromRequest(r, provider, model, sourceFormat, targetFormat)
	
	// Check if request is streaming
	isStreaming := h.isStreamingRequest(r)
	
	if isStreaming {
		// Handle streaming response
		return h.streamingHandler.HandleStreaming(ctx, w, upstreamResp, config)
	} else {
		// Handle non-streaming response
		// Check if upstream response is SSE
		contentType := upstreamResp.Header.Get("Content-Type")
		if contentType == "text/event-stream" {
			// Convert SSE to JSON
			response, _, err := h.sseToJSONConverter.Convert(ctx, upstreamResp.Body, sourceFormat)
			if err != nil {
				return h.nonStreamingHandler.HandleNonStreamingError(w, err.Error(), "conversion_error", http.StatusInternalServerError)
			}
			
			// Write JSON response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(response)
			return err
		} else {
			// Direct non-streaming response
			return h.nonStreamingHandler.HandleNonStreaming(ctx, w, upstreamResp, config)
		}
	}
}

// isStreamingRequest checks if request is streaming
func (h *EnhancedChatHandler) isStreamingRequest(r *http.Request) bool {
	// Try to parse request body to check stream parameter
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// Can't read body, check headers
		r.Body = io.NopCloser(io.MultiReader(io.NopCloser(bytes.NewReader(body)), r.Body))
		return r.Header.Get("Accept") == "text/event-stream"
	}
	
	// Parse JSON to check stream field
	var req struct {
		Stream bool `json:"stream"`
	}
	if err := json.Unmarshal(body, &req); err == nil {
		// Restore request body
		r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body), r.Body))
		return req.Stream
	}
	
	// Restore request body
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body), r.Body))
	return r.Header.Get("Accept") == "text/event-stream"
}

// HandleStreamingError handles streaming errors
func (h *EnhancedChatHandler) HandleStreamingError(
	w http.ResponseWriter,
	message string,
	errorType string,
	code int,
) error {
	return h.streamingHandler.HandleStreamingError(w, message, errorType, code)
}

// HandleNonStreamingError handles non-streaming errors
func (h *EnhancedChatHandler) HandleNonStreamingError(
	w http.ResponseWriter,
	message string,
	errorType string,
	code int,
) error {
	return h.nonStreamingHandler.HandleNonStreamingError(w, message, errorType, code)
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
