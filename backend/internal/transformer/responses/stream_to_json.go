package responses

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/stream"
)

// StreamToJSONConverter converts Responses API SSE streams to a single JSON response
// ref: open-sse/transformer/streamToJsonConverter.js
type StreamToJSONConverter struct{}

// NewStreamToJSONConverter creates a new converter
func NewStreamToJSONConverter() *StreamToJSONConverter {
	return &StreamToJSONConverter{}
}

// convertState holds state during SSE-to-JSON conversion
// ref: open-sse/transformer/streamToJsonConverter.js:58-64
type convertState struct {
	ResponseID string             `json:"id"`
	Created    int64              `json:"created_at"`
	Status     string             `json:"status"`
	Usage      responsesAPIUsage  `json:"usage"`
	Items      map[int]OutputItem `json:"-"`
}

type responsesAPIUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// OutputItem represents an output item in Responses API
// ref: open-sse/transformer/streamToJsonConverter.js:29
type OutputItem struct {
	ID        string        `json:"id,omitempty"`
	Type      string        `json:"type"`
	Content   []ContentPart `json:"content,omitempty"`
	Role      string        `json:"role,omitempty"`
	Name      string        `json:"name,omitempty"`
	Arguments string        `json:"arguments,omitempty"`
	CallID    string        `json:"call_id,omitempty"`
	Summary   []SummaryPart `json:"summary,omitempty"`
}

type ContentPart struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	Annotations []any  `json:"annotations,omitempty"`
	Logprobs    []any  `json:"logprobs,omitempty"`
}

type SummaryPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Convert converts an SSE stream to a JSON response
// ref: open-sse/transformer/streamToJsonConverter.js:49-103
func (c *StreamToJSONConverter) Convert(ctx context.Context, reader io.Reader) ([]byte, error) {
	sseReader := stream.NewSSEReader(reader)
	defer sseReader.Close()

	state := &convertState{
		ResponseID: "",
		Created:    0,
		Status:     "in_progress",
		Usage:      responsesAPIUsage{},
		Items:      make(map[int]OutputItem),
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		chunk, err := sseReader.ReadChunk(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if stream.IsDoneChunk(chunk) {
			break
		}

		c.processSSEMessage(chunk, state)
	}

	return c.buildFinalResponse(state)
}

// processSSEMessage processes a single SSE message
// ref: open-sse/transformer/streamToJsonConverter.js:10-40
func (c *StreamToJSONConverter) processSSEMessage(chunk *stream.StreamChunk, state *convertState) {
	if len(chunk.Data) == 0 {
		return
	}

	var data map[string]any
	if err := json.Unmarshal(chunk.Data, &data); err != nil {
		return
	}

	eventType := chunk.Event

	switch eventType {
	case "response.created":
		c.handleResponseCreated(data, state)
	case "response.output_item.done":
		c.handleOutputItemDone(data, state)
	case "response.completed":
		c.handleResponseCompleted(data, state)
	case "response.failed":
		state.Status = "failed"
	}
}

// handleResponseCreated handles response.created event
// ref: open-sse/transformer/streamToJsonConverter.js:25-27
func (c *StreamToJSONConverter) handleResponseCreated(data map[string]any, state *convertState) {
	if resp, ok := data["response"].(map[string]any); ok {
		if id, ok := resp["id"].(string); ok {
			state.ResponseID = id
		}
		if created, ok := resp["created_at"].(float64); ok {
			state.Created = int64(created)
		}
	}
}

// handleOutputItemDone handles response.output_item.done event
// ref: open-sse/transformer/streamToJsonConverter.js:28-29
func (c *StreamToJSONConverter) handleOutputItemDone(data map[string]any, state *convertState) {
	outputIndex := 0
	if idx, ok := data["output_index"].(float64); ok {
		outputIndex = int(idx)
	}

	if item, ok := data["item"].(map[string]any); ok {
		state.Items[outputIndex] = c.parseOutputItem(item)
	}
}

// handleResponseCompleted handles response.completed event
// ref: open-sse/transformer/streamToJsonConverter.js:30-36
func (c *StreamToJSONConverter) handleResponseCompleted(data map[string]any, state *convertState) {
	state.Status = "completed"

	if resp, ok := data["response"].(map[string]any); ok {
		if usage, ok := resp["usage"].(map[string]any); ok {
			if input, ok := usage["input_tokens"].(float64); ok {
				state.Usage.InputTokens = int(input)
			}
			if output, ok := usage["output_tokens"].(float64); ok {
				state.Usage.OutputTokens = int(output)
			}
			if total, ok := usage["total_tokens"].(float64); ok {
				state.Usage.TotalTokens = int(total)
			}
		}
	}
}

// parseOutputItem parses an output item from the SSE data
func (c *StreamToJSONConverter) parseOutputItem(item map[string]any) OutputItem {
	result := OutputItem{}

	if id, ok := item["id"].(string); ok {
		result.ID = id
	}
	if typ, ok := item["type"].(string); ok {
		result.Type = typ
	}
	if role, ok := item["role"].(string); ok {
		result.Role = role
	}
	if name, ok := item["name"].(string); ok {
		result.Name = name
	}
	if args, ok := item["arguments"].(string); ok {
		result.Arguments = args
	}
	if callID, ok := item["call_id"].(string); ok {
		result.CallID = callID
	}

	if content, ok := item["content"].([]any); ok {
		result.Content = c.parseContentArray(content)
	}

	if summary, ok := item["summary"].([]any); ok {
		result.Summary = c.parseSummaryArray(summary)
	}

	return result
}

// parseContentArray parses content array
func (c *StreamToJSONConverter) parseContentArray(content []any) []ContentPart {
	result := make([]ContentPart, 0, len(content))
	for _, item := range content {
		if cp, ok := item.(map[string]any); ok {
			part := ContentPart{}
			if typ, ok := cp["type"].(string); ok {
				part.Type = typ
			}
			if text, ok := cp["text"].(string); ok {
				part.Text = text
			}
			part.Annotations = []any{}
			part.Logprobs = []any{}
			result = append(result, part)
		}
	}
	return result
}

// parseSummaryArray parses summary array
func (c *StreamToJSONConverter) parseSummaryArray(summary []any) []SummaryPart {
	result := make([]SummaryPart, 0, len(summary))
	for _, item := range summary {
		if sp, ok := item.(map[string]any); ok {
			part := SummaryPart{}
			if typ, ok := sp["type"].(string); ok {
				part.Type = typ
			}
			if text, ok := sp["text"].(string); ok {
				part.Text = text
			}
			result = append(result, part)
		}
	}
	return result
}

// buildFinalResponse builds the final JSON response
// ref: open-sse/transformer/streamToJsonConverter.js:88-102
func (c *StreamToJSONConverter) buildFinalResponse(state *convertState) ([]byte, error) {
	// Build output array from accumulated items (ordered by index)
	output := c.buildOutputArray(state.Items)

	if state.ResponseID == "" {
		state.ResponseID = "resp_" + strconv.FormatInt(state.Created, 10)
	}

	response := map[string]any{
		"id":         state.ResponseID,
		"object":     "response",
		"created_at": state.Created,
		"status":     state.Status,
		"output":     output,
		"usage":      state.Usage,
	}

	return json.Marshal(response)
}

// buildOutputArray builds output array ordered by index
// ref: open-sse/transformer/streamToJsonConverter.js:88-93
func (c *StreamToJSONConverter) buildOutputArray(items map[int]OutputItem) []OutputItem {
	if len(items) == 0 {
		return []OutputItem{}
	}

	// Find max index
	maxIndex := 0
	for idx := range items {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	// Build ordered output
	output := make([]OutputItem, 0, maxIndex+1)
	for i := 0; i <= maxIndex; i++ {
		if item, ok := items[i]; ok {
			output = append(output, item)
		} else {
			// Default message item
			output = append(output, OutputItem{
				Type:    "message",
				Content: []ContentPart{},
				Role:    "assistant",
			})
		}
	}

	return output
}

// ConvertResponsesStreamToJson converts SSE stream to JSON (convenience function)
// ref: open-sse/transformer/streamToJsonConverter.js:49
func ConvertResponsesStreamToJson(ctx context.Context, reader io.Reader) ([]byte, error) {
	converter := NewStreamToJSONConverter()
	return converter.Convert(ctx, reader)
}

// IsSSEContentType checks if content type is SSE
func IsSSEContentType(contentType string) bool {
	return strings.Contains(contentType, "text/event-stream")
}
