package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CLaudeToOpenAIStream transforms CLaude SSE events to OpenAI format line-by-line
// ref: open-sse/translator/response/openai-to-claude.js
type CLaudeToOpenAIStream struct {
	w         http.ResponseWriter
	flush     http.Flusher
	messageID string
	model     string
}

// NewCLaudeToOpenAIStream creates a new stream transformer
func NewCLaudeToOpenAIStream(w io.Writer) (*CLaudeToOpenAIStream, error) {
	httpWriter, ok := w.(http.ResponseWriter)
	if !ok {
		return nil, fmt.Errorf("writer must be http.ResponseWriter for SSE streaming")
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("writer must support http.Flusher for SSE streaming")
	}
	return &CLaudeToOpenAIStream{
		w:         httpWriter,
		flush:     flusher,
		messageID: "chatcmpl-claude",
		model:     "claude",
	}, nil
}

// SetMessageID sets the message ID for the OpenAI response
func (s *CLaudeToOpenAIStream) SetMessageID(id string) {
	s.messageID = id
}

// SetModel sets the model name for the OpenAI response
func (s *CLaudeToOpenAIStream) SetModel(model string) {
	s.model = model
}

// Write processes a line of CLaude SSE and transforms to OpenAI format
// ref: open-sse/translator/response/openai-to-claude.js:29
func (s *CLaudeToOpenAIStream) Write(line []byte) error {
	lineStr := strings.TrimSpace(string(line))

	if lineStr == "" {
		return nil
	}

	if !strings.HasPrefix(lineStr, "data: ") {
		return nil
	}

	data := strings.TrimPrefix(lineStr, "data: ")

	if data == "[DONE]" {
		return s.writeDone()
	}

	var event CLaudeEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return nil
	}

	return s.transformEvent(&event)
}

// CLaudeEvent represents a CLaude SSE event
// ref: open-sse/translator/response/openai-to-claude.js:81
type CLaudeEvent struct {
	Type         string         `json:"type"`
	Index        int            `json:"index,omitempty"`
	Message      *CLaudeMessage `json:"message,omitempty"`
	Delta        *CLaudeDelta   `json:"delta,omitempty"`
	ContentBlock *CLaudeContent `json:"content_block,omitempty"`
	Usage        *CLaudeUsage   `json:"usage,omitempty"`
}

// CLaudeMessage represents the message object in message_start
// ref: open-sse/translator/response/openai-to-claude.js:83
type CLaudeMessage struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Role  string `json:"role"`
	Model string `json:"model"`
}

// CLaudeDelta represents delta content
// ref: open-sse/translator/response/openai-to-claude.js:134
type CLaudeDelta struct {
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	Thinking   string `json:"thinking,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}

// CLaudeContent represents content_block in content_block_start
// ref: open-sse/translator/response/openai-to-claude.js:129
type CLaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// CLaudeUsage represents token usage
// ref: open-sse/translator/response/openai-to-claude.js:91
type CLaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// transformEvent converts a CLaude event to OpenAI format
func (s *CLaudeToOpenAIStream) transformEvent(event *CLaudeEvent) error {
	switch event.Type {
	case "message_start":
		if event.Message != nil {
			s.messageID = event.Message.ID
			if event.Message.Model != "" {
				s.model = event.Message.Model
			}
		}
		return nil

	case "content_block_start":
		return nil

	case "content_block_delta":
		return s.transformContentDelta(event)

	case "content_block_stop":
		return nil

	case "message_delta":
		return s.transformMessageDelta(event)

	case "message_stop":
		return s.writeDone()

	case "ping":
		return nil

	case "error":
		return s.transformError(event)

	default:
		return nil
	}
}

// transformContentDelta handles content_block_delta events
// ref: open-sse/translator/response/openai-to-claude.js:133
func (s *CLaudeToOpenAIStream) transformContentDelta(event *CLaudeEvent) error {
	if event.Delta == nil {
		return nil
	}

	switch event.Delta.Type {
	case "text_delta":
		if event.Delta.Text == "" {
			return nil
		}
		return s.writeContentChunk(event.Delta.Text)
	case "thinking_delta":
		return s.writeReasoningChunk(event.Delta.Thinking)
	default:
		return nil
	}
}

// transformMessageDelta handles message_delta events (contains usage/stop_reason)
// ref: open-sse/translator/response/openai-to-claude.js:200
func (s *CLaudeToOpenAIStream) transformMessageDelta(event *CLaudeEvent) error {
	if event.Delta != nil && event.Delta.StopReason != "" {
		finishReason := convertStopReason(event.Delta.StopReason)
		return s.writeFinishChunk(finishReason)
	}
	return nil
}

// transformError handles error events
func (s *CLaudeToOpenAIStream) transformError(event *CLaudeEvent) error {
	chunk := openAIStreamChunk{
		ID:      s.messageID,
		Object:  "chat.completion.chunk",
		Created: 0,
		Model:   s.model,
		Choices: []openAIStreamChoice{{
			Index:        0,
			Delta:        openAIStreamDelta{},
			FinishReason: "error",
		}},
	}
	return s.writeJSON(&chunk)
}

// writeContentChunk emits an OpenAI text content chunk
// ref: open-sse/translator/response/openai-to-claude.js:133
func (s *CLaudeToOpenAIStream) writeContentChunk(content string) error {
	chunk := openAIStreamChunk{
		ID:      s.messageID,
		Object:  "chat.completion.chunk",
		Created: 0,
		Model:   s.model,
		Choices: []openAIStreamChoice{{
			Index: 0,
			Delta: openAIStreamDelta{
				Content: content,
			},
		}},
	}
	return s.writeJSON(&chunk)
}

// writeReasoningChunk emits an OpenAI reasoning content chunk
func (s *CLaudeToOpenAIStream) writeReasoningChunk(content string) error {
	chunk := map[string]interface{}{
		"id":      s.messageID,
		"object":  "chat.completion.chunk",
		"created": 0,
		"model":   s.model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"delta": map[string]interface{}{
					"reasoning_content": content,
				},
			},
		},
	}
	return s.writeJSON(&chunk)
}

// writeFinishChunk emits an OpenAI finish chunk
// ref: open-sse/translator/response/openai-to-claude.js:184
func (s *CLaudeToOpenAIStream) writeFinishChunk(finishReason string) error {
	chunk := openAIStreamChunk{
		ID:      s.messageID,
		Object:  "chat.completion.chunk",
		Created: 0,
		Model:   s.model,
		Choices: []openAIStreamChoice{{
			Index:        0,
			Delta:        openAIStreamDelta{},
			FinishReason: finishReason,
		}},
	}
	return s.writeJSON(&chunk)
}

// writeDone emits the SSE termination signal
// ref: open-sse/translator/response/openai-to-claude.js:205
func (s *CLaudeToOpenAIStream) writeDone() error {
	s.writeHeaders()
	if _, err := s.w.Write([]byte("data: [DONE]\n\n")); err != nil {
		return err
	}
	s.flush.Flush()
	return nil
}

// writeHeaders ensures SSE headers are set
func (s *CLaudeToOpenAIStream) writeHeaders() {
	s.w.Header().Set("Content-Type", "text/event-stream")
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")
}

// writeJSON marshals and writes JSON with SSE formatting
func (s *CLaudeToOpenAIStream) writeJSON(v interface{}) error {
	s.writeHeaders()
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return err
	}
	s.flush.Flush()
	return nil
}

// convertStopReason converts CLaude stop_reason to OpenAI finish_reason
// ref: open-sse/translator/response/openai-to-claude.js:212
func convertStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	default:
		return "stop"
	}
}

// openAIStreamChunk represents an OpenAI streaming chunk
type openAIStreamChunk struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []openAIStreamChoice  `json:"choices"`
}

// openAIStreamChoice represents a choice in OpenAI response
type openAIStreamChoice struct {
	Index        int               `json:"index"`
	Delta        openAIStreamDelta `json:"delta"`
	FinishReason string            `json:"finish_reason,omitempty"`
}

// openAIStreamDelta represents delta content in OpenAI format
type openAIStreamDelta struct {
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"`
}
