package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OpenAIToCL4udeStream transforms OpenAI SSE events to CL4ude format line-by-line
// ref: open-sse/translator/response/openai-to-claude.js:29
type OpenAIToCL4udeStream struct {
	w                   http.ResponseWriter
	flush               http.Flusher
	messageStartSent    bool
	messageID           string
	model               string
	nextBlockIndex      int
	textBlockIndex      int
	textBlockStarted    bool
	textBlockClosed     bool
	thinkingBlockIndex  int
	thinkingBlockStarted bool
	toolCalls           map[int]*toolCallInfo
	mu                  sync.Mutex
}

type toolCallInfo struct {
	id         string
	name       string
	blockIndex int
}

// NewOpenAIToCL4udeStream creates a new stream transformer
func NewOpenAIToCL4udeStream(w io.Writer) *OpenAIToCL4udeStream {
	httpWriter, ok := w.(http.ResponseWriter)
	if !ok {
		panic("writer must be http.ResponseWriter for SSE streaming")
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("writer must support http.Flusher for SSE streaming")
	}
	return &OpenAIToCL4udeStream{
		w:         httpWriter,
		flush:     flusher,
		messageID: fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		toolCalls: make(map[int]*toolCallInfo),
	}
}

// Write processes a line of OpenAI SSE and transforms to CL4ude format
// ref: open-sse/translator/response/openai-to-claude.js:29
func (s *OpenAIToCL4udeStream) Write(line []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lineStr := strings.TrimSpace(string(line))

	if lineStr == "" {
		return nil
	}

	if !strings.HasPrefix(lineStr, "data: ") {
		return nil
	}

	data := strings.TrimPrefix(lineStr, "data: ")

	if data == "[DONE]" {
		return s.handleDone()
	}

	var chunk openAIStreamChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return nil
	}

	return s.transformChunk(&chunk)
}

// transformChunk converts OpenAI chunk to CL4ude events
// ref: open-sse/translator/response/openai-to-claude.js:29
func (s *OpenAIToCL4udeStream) transformChunk(chunk *openAIStreamChunk) error {
	if len(chunk.Choices) == 0 {
		return nil
	}

	choice := chunk.Choices[0]
	delta := choice.Delta

	// First chunk - send message_start
	// ref: open-sse/translator/response/openai-to-claude.js:71
	if !s.messageStartSent {
		s.messageStartSent = true
		if chunk.ID != "" {
			s.messageID = strings.Replace(chunk.ID, "chatcmpl-", "", 1)
		}
		if chunk.Model != "" {
			s.model = chunk.Model
		}
		s.nextBlockIndex = 0

		if err := s.writeMessageStart(); err != nil {
			return err
		}
	}

	// Handle content
	// ref: open-sse/translator/response/openai-to-claude.js:119
	if delta.Content != "" {
		if err := s.handleTextContent(delta.Content); err != nil {
			return err
		}
	}

	// Handle finish
	// ref: open-sse/translator/response/openai-to-claude.js:184
	if choice.FinishReason != "" {
		if err := s.handleFinish(choice.FinishReason); err != nil {
			return err
		}
	}

	return nil
}

// handleTextContent processes text content delta
// ref: open-sse/translator/response/openai-to-claude.js:119
func (s *OpenAIToCL4udeStream) handleTextContent(content string) error {
	// Stop thinking block if started
	if s.thinkingBlockStarted {
		if err := s.writeContentBlockStop(s.thinkingBlockIndex); err != nil {
			return err
		}
		s.thinkingBlockStarted = false
	}

	// Start text block if not started
	if !s.textBlockStarted {
		s.textBlockIndex = s.nextBlockIndex
		s.nextBlockIndex++
		s.textBlockStarted = true
		s.textBlockClosed = false

		if err := s.writeContentBlockStart(s.textBlockIndex, "text", nil); err != nil {
			return err
		}
	}

	// Write text delta
	// ref: open-sse/translator/response/openai-to-claude.js:133
	return s.writeContentBlockDelta(s.textBlockIndex, "text_delta", content, "")
}

// handleFinish processes finish_reason
// ref: open-sse/translator/response/openai-to-claude.js:184
func (s *OpenAIToCL4udeStream) handleFinish(finishReason string) error {
	// Stop thinking block
	if s.thinkingBlockStarted {
		if err := s.writeContentBlockStop(s.thinkingBlockIndex); err != nil {
			return err
		}
		s.thinkingBlockStarted = false
	}

	// Stop text block
	if s.textBlockStarted && !s.textBlockClosed {
		s.textBlockClosed = true
		if err := s.writeContentBlockStop(s.textBlockIndex); err != nil {
			return err
		}
		s.textBlockStarted = false
	}

	// Stop tool blocks
	for _, tc := range s.toolCalls {
		if err := s.writeContentBlockStop(tc.blockIndex); err != nil {
			return err
		}
	}

	// Write message_delta with stop_reason
	// ref: open-sse/translator/response/openai-to-claude.js:200
	stopReason := convertFinishReason(finishReason)
	if err := s.writeMessageDelta(stopReason); err != nil {
		return err
	}

	// Write message_stop
	// ref: open-sse/translator/response/openai-to-claude.js:205
	return s.writeMessageStop()
}

// handleDone handles stream termination
func (s *OpenAIToCL4udeStream) handleDone() error {
	return nil
}

// writeMessageStart emits CL4ude message_start event
// ref: open-sse/translator/response/openai-to-claude.js:81
func (s *OpenAIToCL4udeStream) writeMessageStart() error {
	event := map[string]interface{}{
		"type": "message_start",
		"message": map[string]interface{}{
			"id":            s.messageID,
			"type":          "message",
			"role":          "assistant",
			"model":         s.model,
			"content":       []interface{}{},
			"stop_reason":   nil,
			"stop_sequence": nil,
			"usage": map[string]int{
				"input_tokens":  0,
				"output_tokens": 0,
			},
		},
	}
	return s.writeJSON(&event)
}

// writeContentBlockStart emits CL4ude content_block_start event
// ref: open-sse/translator/response/openai-to-claude.js:104
func (s *OpenAIToCL4udeStream) writeContentBlockStart(index int, blockType string, extra map[string]interface{}) error {
	contentBlock := map[string]interface{}{
		"type": blockType,
	}
	if blockType == "text" {
		contentBlock["text"] = ""
	}
	for k, v := range extra {
		contentBlock[k] = v
	}

	event := map[string]interface{}{
		"type":          "content_block_start",
		"index":         index,
		"content_block": contentBlock,
	}
	return s.writeJSON(&event)
}

// writeContentBlockDelta emits CL4ude content_block_delta event
// ref: open-sse/translator/response/openai-to-claude.js:133
func (s *OpenAIToCL4udeStream) writeContentBlockDelta(index int, deltaType, text, thinking string) error {
	delta := map[string]interface{}{
		"type": deltaType,
	}
	if deltaType == "text_delta" {
		delta["text"] = text
	} else if deltaType == "thinking_delta" {
		delta["thinking"] = thinking
	}

	event := map[string]interface{}{
		"type":  "content_block_delta",
		"index": index,
		"delta": delta,
	}
	return s.writeJSON(&event)
}

// writeContentBlockStop emits CL4ude content_block_stop event
// ref: open-sse/translator/response/openai-to-claude.js:11
func (s *OpenAIToCL4udeStream) writeContentBlockStop(index int) error {
	event := map[string]interface{}{
		"type":  "content_block_stop",
		"index": index,
	}
	return s.writeJSON(&event)
}

// writeMessageDelta emits CL4ude message_delta event
// ref: open-sse/translator/response/openai-to-claude.js:200
func (s *OpenAIToCL4udeStream) writeMessageDelta(stopReason string) error {
	event := map[string]interface{}{
		"type": "message_delta",
		"delta": map[string]interface{}{
			"stop_reason": stopReason,
		},
		"usage": map[string]int{
			"input_tokens":  0,
			"output_tokens": 0,
		},
	}
	return s.writeJSON(&event)
}

// writeMessageStop emits CL4ude message_stop event
// ref: open-sse/translator/response/openai-to-claude.js:205
func (s *OpenAIToCL4udeStream) writeMessageStop() error {
	event := map[string]interface{}{
		"type": "message_stop",
	}
	return s.writeJSON(&event)
}

// writeHeaders ensures SSE headers are set
func (s *OpenAIToCL4udeStream) writeHeaders() {
	s.w.Header().Set("Content-Type", "text/event-stream")
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")
}

// writeJSON marshals and writes JSON with SSE formatting
func (s *OpenAIToCL4udeStream) writeJSON(v interface{}) error {
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

// convertFinishReason converts OpenAI finish_reason to CL4ude stop_reason
// ref: open-sse/translator/response/openai-to-claude.js:212
func convertFinishReason(reason string) string {
	switch reason {
	case "stop":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "tool_calls":
		return "tool_use"
	default:
		return "end_turn"
	}
}

// PassthroughWriter passes OpenAI SSE events through unchanged
type PassthroughWriter struct {
	w     http.ResponseWriter
	flush http.Flusher
}

// NewPassthroughWriter creates a passthrough writer
func NewPassthroughWriter(w io.Writer) *PassthroughWriter {
	httpWriter, ok := w.(http.ResponseWriter)
	if !ok {
		panic("writer must be http.ResponseWriter for SSE streaming")
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("writer must support http.Flusher for SSE streaming")
	}
	return &PassthroughWriter{
		w:     httpWriter,
		flush: flusher,
	}
}

// Write passes data through unchanged
func (p *PassthroughWriter) Write(line []byte) error {
	p.w.Header().Set("Content-Type", "text/event-stream")
	p.w.Header().Set("Cache-Control", "no-cache")
	p.w.Header().Set("Connection", "keep-alive")

	if _, err := p.w.Write(line); err != nil {
		return err
	}
	p.flush.Flush()
	return nil
}
