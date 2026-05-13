// Package executor provides provider-specific request/response handling.
// KiroExecutor handles AWS CodeWhisperer/Kiro streaming API with AWS EventStream binary format.
// ref: _ref/9router/open-sse/executors/kiro.js
package executor

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// KiroExecutor handles Kiro AI (AWS CodeWhisperer) streaming responses.
// ref: open-sse/executors/kiro.js:12
type KiroExecutor struct {
	BaseExecutor
}

// NewKiroExecutor creates a new Kiro executor.
// ref: open-sse/executors/kiro.js:13
func NewKiroExecutor() *KiroExecutor {
	return &KiroExecutor{
		BaseExecutor: NewBaseExecutor("kiro"),
	}
}

// PrepareRequest modifies the outgoing request for Kiro-specific headers.
// ref: open-sse/executors/kiro.js:17
func (k *KiroExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Add AWS SDK specific headers
	// ref: open-sse/executors/kiro.js:20
	req.Header.Set("Amz-Sdk-Request", "attempt=1; max=3")
	req.Header.Set("Amz-Sdk-Invocation-Id", uuid.New().String())

	// Body is passed through unchanged for Kiro
	// ref: open-sse/executors/kiro.js:31
	return nil
}

// TransformResponse transforms AWS EventStream binary response to SSE text stream.
// ref: open-sse/executors/kiro.js:81
func (k *KiroExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return []byte("data: [DONE]\n\n"), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Transform binary EventStream to SSE format
	transformer := newKiroEventStreamTransformer()
	sseData, err := transformer.Transform(body)
	if err != nil {
		return nil, err
	}

	return sseData, nil
}

// HandleError processes Kiro-specific errors.
// ref: open-sse/executors/kiro.js:38
func (k *KiroExecutor) HandleError(ctx context.Context, err error) error {
	return err
}

// kiroEventStreamTransformer converts AWS EventStream binary to SSE text.
// ref: open-sse/executors/kiro.js:81
type kiroEventStreamTransformer struct {
	responseID string
	created    int64
	state      *kiroStreamState
}

// kiroStreamState tracks the state during EventStream parsing.
// ref: open-sse/executors/kiro.js:86
type kiroStreamState struct {
	chunkIndex           int
	endDetected          bool
	finishEmitted        bool
	hasToolCalls         bool
	toolCallIndex        int
	seenToolIDs          map[string]int
	totalContentLength   int
	contextUsagePercent  float64
	hasContextUsage      bool
	hasMeteringEvent     bool
	usage                *openAIUsage
}

// openAIUsage represents token usage in OpenAI format.
// ref: open-sse/executors/kiro.js:282
type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// openAIChunk represents a streaming chunk in OpenAI format.
// ref: open-sse/executors/kiro.js:129
type openAIChunk struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []openAIChoice   `json:"choices"`
	Usage   *openAIUsage     `json:"usage,omitempty"`
}

// openAIChoice represents a choice in OpenAI streaming response.
// ref: open-sse/executors/kiro.js:134
type openAIChoice struct {
	Index        int          `json:"index"`
	Delta        openAIDelta  `json:"delta"`
	FinishReason *string      `json:"finish_reason,omitempty"`
}

// openAIDelta represents delta content in OpenAI format.
// ref: open-sse/executors/kiro.js:136
type openAIDelta struct {
	Role      string              `json:"role,omitempty"`
	Content   string              `json:"content,omitempty"`
	ToolCalls []openAIToolCall    `json:"tool_calls,omitempty"`
}

// openAIToolCall represents a tool call in OpenAI format.
// ref: open-sse/executors/kiro.js:190
type openAIToolCall struct {
	Index    int                  `json:"index"`
	ID       string               `json:"id,omitempty"`
	Type     string               `json:"type,omitempty"`
	Function openAIToolFunction   `json:"function,omitempty"`
}

// openAIToolFunction represents function details in a tool call.
// ref: open-sse/executors/kiro.js:194
type openAIToolFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// newKiroEventStreamTransformer creates a new transformer.
// ref: open-sse/executors/kiro.js:81
func newKiroEventStreamTransformer() *kiroEventStreamTransformer {
	return &kiroEventStreamTransformer{
		responseID: fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		created:    time.Now().Unix(),
		state: &kiroStreamState{
			seenToolIDs: make(map[string]int),
		},
	}
}

// Transform converts binary EventStream data to SSE format.
// ref: open-sse/executors/kiro.js:94
func (t *kiroEventStreamTransformer) Transform(data []byte) ([]byte, error) {
	var output bytes.Buffer
	buffer := data

	// Parse events from buffer
	// ref: open-sse/executors/kiro.js:105
	iterations := 0
	maxIterations := 1000

	for len(buffer) >= 16 && iterations < maxIterations {
		iterations++

		// Read total length (first 4 bytes, big-endian)
		// ref: open-sse/executors/kiro.js:108
		totalLength := binary.BigEndian.Uint32(buffer[0:4])

		if totalLength < 16 || int(totalLength) > len(buffer) {
			break
		}

		// Extract event frame
		eventData := buffer[:totalLength]
		buffer = buffer[totalLength:]

		// Parse the event frame
		// ref: open-sse/executors/kiro.js:115
		event := parseEventFrame(eventData)
		if event == nil {
			continue
		}

		eventType := event.Headers[":event-type"]

		// Handle assistantResponseEvent
		// ref: open-sse/executors/kiro.js:125
		if eventType == "assistantResponseEvent" && event.Payload != nil {
			if content, ok := event.Payload["content"].(string); ok && content != "" {
				t.state.totalContentLength += len(content)
				t.writeChunk(&output, content, "", false, nil)
			}
		}

		// Handle codeEvent
		// ref: open-sse/executors/kiro.js:147
		if eventType == "codeEvent" && event.Payload != nil {
			if content, ok := event.Payload["content"].(string); ok && content != "" {
				t.writeChunk(&output, content, "", false, nil)
			}
		}

		// Handle toolUseEvent
		// ref: open-sse/executors/kiro.js:164
		if eventType == "toolUseEvent" && event.Payload != nil {
			t.handleToolUseEvent(&output, event.Payload)
		}

		// Handle messageStopEvent
		// ref: open-sse/executors/kiro.js:245
		if eventType == "messageStopEvent" {
			t.writeFinishChunk(&output, false)
		}

		// Handle contextUsageEvent
		// ref: open-sse/executors/kiro.js:262
		if eventType == "contextUsageEvent" && event.Payload != nil {
			if percent, ok := event.Payload["contextUsagePercentage"].(float64); ok {
				t.state.contextUsagePercent = percent
				t.state.hasContextUsage = true
			}
		}

		// Handle meteringEvent
		// ref: open-sse/executors/kiro.js:269
		if eventType == "meteringEvent" {
			t.state.hasMeteringEvent = true
		}

		// Handle metricsEvent
		// ref: open-sse/executors/kiro.js:274
		if eventType == "metricsEvent" && event.Payload != nil {
			t.handleMetricsEvent(event.Payload)
		}

		// Emit final chunk after both metering and context events
		// ref: open-sse/executors/kiro.js:292
		if t.state.hasMeteringEvent && t.state.hasContextUsage && !t.state.finishEmitted {
			t.writeFinishChunk(&output, true)
		}
	}

	// Flush any remaining state
	// ref: open-sse/executors/kiro.js:341
	if !t.state.finishEmitted {
		t.writeFinishChunk(&output, false)
	}

	// Send final done message
	// ref: open-sse/executors/kiro.js:360
	output.WriteString("data: [DONE]\n\n")

	return output.Bytes(), nil
}

// writeChunk writes a content chunk to the output.
// ref: open-sse/executors/kiro.js:129
func (t *kiroEventStreamTransformer) writeChunk(output *bytes.Buffer, content, role string, hasToolCalls bool, toolCalls []openAIToolCall) {
	delta := openAIDelta{}

	if t.state.chunkIndex == 0 && role != "" {
		delta.Role = role
	}
	if content != "" {
		delta.Content = content
	}
	if hasToolCalls {
		delta.ToolCalls = toolCalls
	}

	chunk := openAIChunk{
		ID:      t.responseID,
		Object:  "chat.completion.chunk",
		Created: t.created,
		Model:   "kiro",
		Choices: []openAIChoice{
			{
				Index: 0,
				Delta: delta,
			},
		},
	}

	t.writeSSE(output, chunk)
	t.state.chunkIndex++
}

// writeFinishChunk writes the finish chunk to output.
// ref: open-sse/executors/kiro.js:246
func (t *kiroEventStreamTransformer) writeFinishChunk(output *bytes.Buffer, withUsage bool) {
	if t.state.finishEmitted {
		return
	}

	finishReason := "stop"
	if t.state.hasToolCalls {
		finishReason = "tool_calls"
	}

	chunk := openAIChunk{
		ID:      t.responseID,
		Object:  "chat.completion.chunk",
		Created: t.created,
		Model:   "kiro",
		Choices: []openAIChoice{
			{
				Index:        0,
				Delta:        openAIDelta{},
				FinishReason: &finishReason,
			},
		},
	}

	// Include usage if requested and available
	// ref: open-sse/executors/kiro.js:328
	if withUsage && t.state.usage != nil {
		chunk.Usage = t.state.usage
	}

	t.writeSSE(output, chunk)
	t.state.finishEmitted = true
}

// handleToolUseEvent processes tool use events.
// ref: open-sse/executors/kiro.js:164
func (t *kiroEventStreamTransformer) handleToolUseEvent(output *bytes.Buffer, payload map[string]interface{}) {
	t.state.hasToolCalls = true

	// The payload is a map containing tool use data
	// In JS, it checks if payload is array, but in Go JSON parsing, arrays are separate types
	// We wrap single payload in array for consistent processing
	// ref: open-sse/executors/kiro.js:167
	toolUses := []map[string]interface{}{payload}

	for _, toolUse := range toolUses {
		toolCallID, _ := toolUse["toolUseId"].(string)
		if toolCallID == "" {
			toolCallID = fmt.Sprintf("call_%d", time.Now().UnixNano())
		}
		toolName, _ := toolUse["name"].(string)
		toolInput := toolUse["input"]

		// Check if this is a new tool
		// ref: open-sse/executors/kiro.js:175
		toolIndex, isNew := t.state.seenToolIDs[toolCallID]
		if isNew {
			toolIndex = t.state.toolCallIndex
			t.state.seenToolIDs[toolCallID] = toolIndex
			t.state.toolCallIndex++

			// Write tool call start chunk
			// ref: open-sse/executors/kiro.js:181
			role := ""
			if t.state.chunkIndex == 0 {
				role = "assistant"
			}
			t.writeChunk(output, "", role, true, []openAIToolCall{
				{
					Index: toolIndex,
					ID:    toolCallID,
					Type:  "function",
					Function: openAIToolFunction{
						Name:      toolName,
						Arguments: "",
					},
				},
			})
		}

		// Write arguments if present
		// ref: open-sse/executors/kiro.js:209
		if toolInput != nil {
			var argsStr string
			switch v := toolInput.(type) {
			case string:
				argsStr = v
			case map[string]interface{}, []interface{}:
				b, _ := json.Marshal(v)
				argsStr = string(b)
			default:
				continue
			}

			t.writeChunk(output, "", "", true, []openAIToolCall{
				{
					Index: toolIndex,
					Function: openAIToolFunction{
						Arguments: argsStr,
					},
				},
			})
		}
	}
}

// handleMetricsEvent extracts token usage from metrics event.
// ref: open-sse/executors/kiro.js:274
func (t *kiroEventStreamTransformer) handleMetricsEvent(payload map[string]interface{}) {
	// Extract from nested metricsEvent or direct payload
	// ref: open-sse/executors/kiro.js:276
	var metrics map[string]interface{}
	if m, ok := payload["metricsEvent"].(map[string]interface{}); ok {
		metrics = m
	} else {
		metrics = payload
	}

	inputTokens, _ := metrics["inputTokens"].(float64)
	outputTokens, _ := metrics["outputTokens"].(float64)

	if inputTokens > 0 || outputTokens > 0 {
		t.state.usage = &openAIUsage{
			PromptTokens:     int(inputTokens),
			CompletionTokens: int(outputTokens),
			TotalTokens:      int(inputTokens + outputTokens),
		}
	}
}

// writeSSE writes a chunk as SSE data.
func (t *kiroEventStreamTransformer) writeSSE(output *bytes.Buffer, chunk openAIChunk) {
	data, _ := json.Marshal(chunk)
	output.WriteString("data: ")
	output.Write(data)
	output.WriteString("\n\n")
}

// eventFrame represents a parsed AWS EventStream frame.
// ref: open-sse/executors/kiro.js:460
type eventFrame struct {
	Headers  map[string]string
	Payload  map[string]interface{}
}

// parseEventFrame parses an AWS EventStream binary frame.
// ref: open-sse/executors/kiro.js:404
func parseEventFrame(data []byte) *eventFrame {
	if len(data) < 16 {
		return nil
	}

	// Read headers length (bytes 4-8)
	// ref: open-sse/executors/kiro.js:407
	headersLength := binary.BigEndian.Uint32(data[4:8])

	// Parse headers
	// ref: open-sse/executors/kiro.js:409
	headers := make(map[string]string)
	offset := 12 // After prelude (8 bytes) + prelude CRC (4 bytes)
	headerEnd := 12 + int(headersLength)

	for offset < headerEnd && offset < len(data) {
		// Read header name length
		// ref: open-sse/executors/kiro.js:415
		nameLen := int(data[offset])
		offset++
		if offset+nameLen > len(data) {
			break
		}

		// Read header name
		// ref: open-sse/executors/kiro.js:419
		name := string(data[offset : offset+nameLen])
		offset += nameLen

		// Read header type
		// ref: open-sse/executors/kiro.js:422
		headerType := data[offset]
		offset++

		// Type 7 = string
		// ref: open-sse/executors/kiro.js:425
		if headerType == 7 {
			// Read value length (2 bytes)
			// ref: open-sse/executors/kiro.js:426
			if offset+2 > len(data) {
				break
			}
			valueLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
			offset += 2

			if offset+valueLen > len(data) {
				break
			}

			// Read value
			// ref: open-sse/executors/kiro.js:430
			value := string(data[offset : offset+valueLen])
			offset += valueLen
			headers[name] = value
		} else {
			// Skip unknown header types
			break
		}
	}

	// Parse payload
	// ref: open-sse/executors/kiro.js:438
	payloadStart := 12 + int(headersLength)
	payloadEnd := len(data) - 4 // Exclude message CRC

	var payload map[string]interface{}
	if payloadEnd > payloadStart {
		payloadStr := string(data[payloadStart:payloadEnd])

		// Skip empty or whitespace-only payloads
		// ref: open-sse/executors/kiro.js:447
		if payloadStr != "" && len(payloadStr) > 0 {
			// Trim whitespace
			for len(payloadStr) > 0 && (payloadStr[0] == ' ' || payloadStr[0] == '\n' || payloadStr[0] == '\r' || payloadStr[0] == '\t') {
				payloadStr = payloadStr[1:]
			}

			if payloadStr != "" {
				if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
					// Store raw payload on parse error
					// ref: open-sse/executors/kiro.js:456
					payload = map[string]interface{}{
						"raw": payloadStr,
					}
				}
			}
		}
	}

	return &eventFrame{
		Headers: headers,
		Payload: payload,
	}
}
