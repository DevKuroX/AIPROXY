package responses

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/stream"
	"github.com/DevKuroX/AIPROXY/internal/transformer/responses"
)

// Handler handles Responses API requests
// ref: open-sse/handlers/responsesHandler.js
type Handler struct {
	// Dependencies would be injected here
}

// NewHandler creates a new Responses API handler
func NewHandler() *Handler {
	return &Handler{}
}

// ResponsesRequest represents a Responses API request
// ref: open-sse/translator/helpers/responsesApiHelper.js - request body structure
type ResponsesRequest struct {
	Model        string          `json:"model"`
	Input        json.RawMessage `json:"input,omitempty"`        // Can be string or array
	Instructions string          `json:"instructions,omitempty"` // System instructions
	Stream       bool            `json:"stream,omitempty"`
	Tools        []Tool          `json:"tools,omitempty"`
	ToolChoice   any             `json:"tool_choice,omitempty"`
	MaxTokens    int             `json:"max_tokens,omitempty"`
	Temperature  float64         `json:"temperature,omitempty"`
	TopP         float64         `json:"top_p,omitempty"`
}

// Tool represents a tool definition in Responses API
// ref: open-sse/translator/helpers/responsesApiHelper.js
type Tool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Parameters  any    `json:"parameters,omitempty"`
	} `json:"function,omitempty"`
}

// HandleResponses handles /v1/responses endpoint
// ref: open-sse/handlers/responsesHandler.js - handleResponsesCore
func (h *Handler) HandleResponses(
	ctx context.Context,
	w http.ResponseWriter,
	upstreamResp *http.Response,
	config *stream.StreamConfig,
	clientRequestedStreaming bool,
) error {
	contentType := upstreamResp.Header.Get("Content-Type")

	// Case 1: Client wants non-streaming, but got SSE (provider forced it, e.g., Codex)
	// ref: open-sse/handlers/responsesHandler.js:55-79
	if !clientRequestedStreaming && strings.Contains(contentType, "text/event-stream") {
		return h.convertStreamToJSON(ctx, w, upstreamResp)
	}

	// Case 2: Client wants streaming, got SSE - transform it
	// ref: open-sse/handlers/responsesHandler.js:81-98
	if clientRequestedStreaming && strings.Contains(contentType, "text/event-stream") {
		return h.transformStream(ctx, w, upstreamResp, config)
	}

	// Case 3: Non-SSE response (error or non-streaming from provider) - return as-is
	// ref: open-sse/handlers/responsesHandler.js:100-101
	return h.handleDirectResponse(ctx, w, upstreamResp)
}

// convertStreamToJSON converts SSE stream to JSON response
// ref: open-sse/handlers/responsesHandler.js:55-79
func (h *Handler) convertStreamToJSON(
	ctx context.Context,
	w http.ResponseWriter,
	upstreamResp *http.Response,
) error {
	converter := responses.NewStreamToJSONConverter()

	jsonResponse, err := converter.Convert(ctx, upstreamResp.Body)
	if err != nil {
		http.Error(w, "Failed to convert streaming response to JSON", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	return err
}

// transformStream transforms SSE stream to Responses API format
// ref: open-sse/handlers/responsesHandler.js:81-98
func (h *Handler) transformStream(
	ctx context.Context,
	w http.ResponseWriter,
	upstreamResp *http.Response,
	config *stream.StreamConfig,
) error {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create SSE reader from upstream response
	sseReader := stream.NewSSEReader(upstreamResp.Body)
	defer sseReader.Close()

	// Create SSE writer for response
	sseWriter := stream.NewSSEWriter(w)

	// Create transformer
	transformer := responses.NewTransformer()

	// Process and transform the stream
	return transformer.Transform(ctx, sseReader, sseWriter, config)
}

// handleDirectResponse handles non-SSE responses directly
// ref: open-sse/handlers/responsesHandler.js:100-101
func (h *Handler) handleDirectResponse(
	ctx context.Context,
	w http.ResponseWriter,
	upstreamResp *http.Response,
) error {
	// Copy headers
	for key, values := range upstreamResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(upstreamResp.StatusCode)

	// Copy body
	_, err := io.Copy(w, upstreamResp.Body)
	return err
}

// normalizeResponsesInput normalizes Responses API input to array format.
// Accepts string or array, returns array of message items.
// An empty array is treated like an empty string — providers require at least one user
// message, so we inject a placeholder rather than forwarding an empty messages[].
// ref: open-sse/translator/helpers/responsesApiHelper.js - normalizeResponsesInput
func normalizeResponsesInput(input any) []map[string]any {
	// Handle string input
	if str, ok := input.(string); ok {
		text := strings.TrimSpace(str)
		if text == "" {
			text = "..."
		}
		return []map[string]any{
			{
				"type": "message",
				"role": "user",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": text,
					},
				},
			},
		}
	}

	// Handle array input
	if arr, ok := input.([]any); ok {
		// Empty input[] would produce messages:[] which all providers reject
		if len(arr) == 0 {
			return []map[string]any{
				{
					"type": "message",
					"role": "user",
					"content": []map[string]any{
						{
							"type": "input_text",
							"text": "...",
						},
					},
				},
			}
		}
		// Convert []any to []map[string]any
		result := make([]map[string]any, len(arr))
		for i, item := range arr {
			if m, ok := item.(map[string]any); ok {
				result[i] = m
			}
		}
		return result
	}

	return nil
}

// ConvertResponsesToChat converts Responses API request to Chat Completions format
// ref: open-sse/translator/helpers/responsesApiHelper.js - convertResponsesApiFormat
func ConvertResponsesToChat(body []byte) ([]byte, error) {
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	// Responses API has "input" field instead of "messages"
	if _, hasInput := req["input"]; !hasInput {
		return body, nil
	}

	// Use the local normalizeResponsesInput function
	normalized := normalizeResponsesInput(req["input"])
	if normalized == nil {
		return nil, fmt.Errorf("invalid input format")
	}

	// Build messages array
	messages := []map[string]any{}

	// Add instructions as system message if present
	if instructions, ok := req["instructions"].(string); ok && instructions != "" {
		messages = append(messages, map[string]any{
			"role":    "system",
			"content": instructions,
		})
	}

	// Add normalized input as user message
	messages = append(messages, map[string]any{
		"role":    "user",
		"content": normalized,
	})

	// Build result
	result := map[string]any{
		"model":    req["model"],
		"messages": messages,
	}

	// Copy other fields
	if v, ok := req["stream"]; ok {
		result["stream"] = v
	}
	if v, ok := req["temperature"]; ok {
		result["temperature"] = v
	}
	if v, ok := req["top_p"]; ok {
		result["top_p"] = v
	}
	if v, ok := req["max_tokens"]; ok {
		result["max_tokens"] = v
	}

	return json.Marshal(result)
}

// IsResponsesAPIRequest checks if a request body is in Responses API format
func IsResponsesAPIRequest(body []byte) bool {
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		return false
	}
	// Responses API has "input" field instead of "messages"
	_, hasInput := req["input"]
	return hasInput
}
