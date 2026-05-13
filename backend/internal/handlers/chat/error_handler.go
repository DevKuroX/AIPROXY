package chat

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/stream"
)

// ErrorHandler handles chat errors and formats them appropriately
// ref: open-sse/handlers/chatCore/requestDetail.js - handleError
type ErrorHandler struct {
	logger *RequestDetailLogger
}

// NewErrorHandler creates a new error handler
// ref: open-sse/handlers/chatCore/requestDetail.js - createErrorHandler
func NewErrorHandler(logger *RequestDetailLogger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError handles an error and writes it to the response
// ref: open-sse/handlers/chatCore/requestDetail.js - handleError
func (h *ErrorHandler) HandleError(
	ctx context.Context,
	w http.ResponseWriter,
	err error,
	streaming bool,
) {
	// Log the error
	if h.logger != nil {
		detail := &RequestDetail{
			Error: err.Error(),
		}
		h.logger.LogRequest(ctx, detail)
	}

	// Format error based on streaming mode
	if streaming {
		h.writeStreamError(w, err)
	} else {
		h.writeJSONError(w, err)
	}
}

// writeStreamError writes an error as an SSE chunk
// ref: open-sse/handlers/chatCore/requestDetail.js - writeStreamError
func (h *ErrorHandler) writeStreamError(w http.ResponseWriter, err error) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create SSE error chunk
	sseError := stream.NewStreamError(err.Error(), "api_error", "")

	// Write SSE error chunk
	w.Write(sseError.ToSSE())
	w.(http.Flusher).Flush()
}

// writeJSONError writes an error as a JSON response
func (h *ErrorHandler) writeJSONError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	response := map[string]any{
		"error": map[string]string{
			"message": err.Error(),
			"type":    "api_error",
		},
	}
	json.NewEncoder(w).Encode(response)
}

// HandleStreamError handles an error during streaming
func (h *ErrorHandler) HandleStreamError(
	ctx context.Context,
	w http.ResponseWriter,
	err error,
) {
	sseError := stream.NewStreamError(err.Error(), "api_error", "")
	w.Write(sseError.ToSSE())
	w.(http.Flusher).Flush()
}

// HandleNonStreamingError handles an error for non-streaming response
// ref: open-sse/handlers/chatCore/requestDetail.js - handleNonStreamingError
func (h *ErrorHandler) HandleNonStreamingError(
	ctx context.Context,
	w http.ResponseWriter,
	err error,
) {
	// Set JSON headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	// Create OpenAI-compatible error response
	response := map[string]any{
		"error": map[string]string{
			"message": err.Error(),
			"type":    "api_error",
		},
	}

	json.NewEncoder(w).Encode(response)

	// Log the error
	if h.logger != nil {
		detail := &RequestDetail{
			StatusCode: http.StatusInternalServerError,
			Error:      err.Error(),
		}
		h.logger.LogRequest(ctx, detail)
	}
}