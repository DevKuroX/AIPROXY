package responses

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/stream"
)

// Transformer transforms Chat Completions SSE to Responses API SSE format
// ref: open-sse/transformer/responsesTransformer.js
type Transformer struct{}

// NewTransformer creates a new transformer
func NewTransformer() *Transformer {
	return &Transformer{}
}

// transformerState holds state during transformation
// ref: open-sse/transformer/responsesTransformer.js:55-77
type transformerState struct {
	seq                int
	responseID         string
	created            int64
	started            bool
	msgTextBuf         map[int]string
	msgItemAdded       map[int]bool
	msgContentAdded    map[int]bool
	msgItemDone        map[int]bool
	reasoningID        string
	reasoningIndex     int
	reasoningBuf       string
	reasoningPartAdded bool
	reasoningDone      bool
	inThinking         bool
	funcArgsBuf        map[int]string
	funcNames          map[int]string
	funcCallIds        map[int]string
	funcArgsDone       map[int]bool
	funcItemDone       map[int]bool
	buffer             string
	completedSent      bool
}

func newTransformerState() *transformerState {
	return &transformerState{
		seq:             0,
		responseID:      fmt.Sprintf("resp_%d", time.Now().UnixNano()/1000000),
		created:         time.Now().Unix(),
		started:         false,
		msgTextBuf:      make(map[int]string),
		msgItemAdded:    make(map[int]bool),
		msgContentAdded: make(map[int]bool),
		msgItemDone:     make(map[int]bool),
		reasoningIndex:  -1,
		funcArgsBuf:     make(map[int]string),
		funcNames:       make(map[int]string),
		funcCallIds:     make(map[int]string),
		funcArgsDone:    make(map[int]bool),
		funcItemDone:    make(map[int]bool),
	}
}

func (s *transformerState) nextSeq() int {
	s.seq++
	return s.seq
}

// Transform transforms SSE stream from Chat Completions to Responses API format
// ref: open-sse/transformer/responsesTransformer.js:54-438
func (t *Transformer) Transform(
	ctx context.Context,
	reader stream.StreamReader,
	writer stream.StreamWriter,
	config *stream.StreamConfig,
) error {
	state := newTransformerState()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chunk, err := reader.ReadChunk(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if stream.IsDoneChunk(chunk) {
			break
		}

		if err := t.processChunk(ctx, writer, chunk, state); err != nil {
			return err
		}
	}

	// Flush remaining items
	// ref: open-sse/transformer/responsesTransformer.js:427-436
	t.flushItems(ctx, writer, state)

	// Send [DONE]
	doneChunk := &stream.StreamChunk{
		Event: "message",
		Data:  []byte("[DONE]"),
	}
	return writer.WriteChunk(ctx, doneChunk)
}

// processChunk processes a single SSE chunk
// ref: open-sse/transformer/responsesTransformer.js:243-425
func (t *Transformer) processChunk(
	ctx context.Context,
	writer stream.StreamWriter,
	chunk *stream.StreamChunk,
	state *transformerState,
) error {
	// Parse the SSE data
	var data map[string]any
	if err := json.Unmarshal(chunk.Data, &data); err != nil {
		return nil // Skip malformed chunks
	}

	choices, ok := data["choices"].([]any)
	if !ok || len(choices) == 0 {
		return nil
	}

	choice := choices[0].(map[string]any)
	idx := 0
	if i, ok := choice["index"].(float64); ok {
		idx = int(i)
	}

	delta, ok := choice["delta"].(map[string]any)
	if !ok {
		delta = make(map[string]any)
	}

	// Emit initial events on first chunk
	// ref: open-sse/transformer/responsesTransformer.js:274-300
	if !state.started {
		state.started = true
		if id, ok := data["id"].(string); ok && id != "" {
			state.responseID = "resp_" + id
		}

		if err := t.emitCreated(ctx, writer, state); err != nil {
			return err
		}
		if err := t.emitInProgress(ctx, writer, state); err != nil {
			return err
		}
	}

	// Handle reasoning_content (OpenAI native format)
	// ref: open-sse/transformer/responsesTransformer.js:302-306
	if rc, ok := delta["reasoning_content"].(string); ok && rc != "" {
		t.startReasoning(ctx, writer, state, idx)
		t.emitReasoningDelta(ctx, writer, state, rc)
	}

	// Handle text content (may contain <think> tags)
	// ref: open-sse/transformer/responsesTransformer.js:308-371
	if content, ok := delta["content"].(string); ok && content != "" {
		t.handleContent(ctx, writer, state, idx, content)
	}

	// Handle tool_calls
	// ref: open-sse/transformer/responsesTransformer.js:373-415
	if toolCalls, ok := delta["tool_calls"].([]any); ok && len(toolCalls) > 0 {
		t.closeMessage(ctx, writer, state, idx)
		t.handleToolCalls(ctx, writer, state, toolCalls)
	}

	// Handle finish_reason
	// ref: open-sse/transformer/responsesTransformer.js:417-423
	if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
		t.closeAllItems(ctx, writer, state)
		t.sendCompleted(ctx, writer, state)
	}

	return nil
}

// emit emits an SSE event
// ref: open-sse/transformer/responsesTransformer.js:82-87
func (t *Transformer) emit(ctx context.Context, writer stream.StreamWriter, state *transformerState, eventType string, data map[string]any) error {
	data["sequence_number"] = state.nextSeq()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	chunk := &stream.StreamChunk{
		Event: eventType,
		Data:  jsonData,
	}
	return writer.WriteChunk(ctx, chunk)
}

// emitCreated emits response.created event
// ref: open-sse/transformer/responsesTransformer.js:278-289
func (t *Transformer) emitCreated(ctx context.Context, writer stream.StreamWriter, state *transformerState) error {
	return t.emit(ctx, writer, state, "response.created", map[string]any{
		"type": "response.created",
		"response": map[string]any{
			"id":         state.responseID,
			"object":     "response",
			"created_at": state.created,
			"status":     "in_progress",
			"background": false,
			"error":      nil,
			"output":     []any{},
		},
	})
}

// emitInProgress emits response.in_progress event
// ref: open-sse/transformer/responsesTransformer.js:291-299
func (t *Transformer) emitInProgress(ctx context.Context, writer stream.StreamWriter, state *transformerState) error {
	return t.emit(ctx, writer, state, "response.in_progress", map[string]any{
		"type": "response.in_progress",
		"response": map[string]any{
			"id":         state.responseID,
			"object":     "response",
			"created_at": state.created,
			"status":     "in_progress",
		},
	})
}

// startReasoning starts reasoning output
// ref: open-sse/transformer/responsesTransformer.js:90-114
func (t *Transformer) startReasoning(ctx context.Context, writer stream.StreamWriter, state *transformerState, idx int) {
	if state.reasoningID == "" {
		state.reasoningID = fmt.Sprintf("rs_%s_%d", state.responseID, idx)
		state.reasoningIndex = idx

		t.emit(ctx, writer, state, "response.output_item.added", map[string]any{
			"type":         "response.output_item.added",
			"output_index": idx,
			"item": map[string]any{
				"id":      state.reasoningID,
				"type":    "reasoning",
				"summary": []any{},
			},
		})

		t.emit(ctx, writer, state, "response.reasoning_summary_part.added", map[string]any{
			"type":          "response.reasoning_summary_part.added",
			"item_id":       state.reasoningID,
			"output_index":  idx,
			"summary_index": 0,
			"part": map[string]any{
				"type": "summary_text",
				"text": "",
			},
		})
		state.reasoningPartAdded = true
	}
}

// emitReasoningDelta emits reasoning delta
// ref: open-sse/transformer/responsesTransformer.js:116-126
func (t *Transformer) emitReasoningDelta(ctx context.Context, writer stream.StreamWriter, state *transformerState, text string) {
	if text == "" {
		return
	}
	state.reasoningBuf += text

	t.emit(ctx, writer, state, "response.reasoning_summary_text.delta", map[string]any{
		"type":          "response.reasoning_summary_text.delta",
		"item_id":       state.reasoningID,
		"output_index":  state.reasoningIndex,
		"summary_index": 0,
		"delta":         text,
	})
}

// closeReasoning closes reasoning output
// ref: open-sse/transformer/responsesTransformer.js:128-158
func (t *Transformer) closeReasoning(ctx context.Context, writer stream.StreamWriter, state *transformerState) {
	if state.reasoningID != "" && !state.reasoningDone {
		state.reasoningDone = true

		t.emit(ctx, writer, state, "response.reasoning_summary_text.done", map[string]any{
			"type":          "response.reasoning_summary_text.done",
			"item_id":       state.reasoningID,
			"output_index":  state.reasoningIndex,
			"summary_index": 0,
			"text":          state.reasoningBuf,
		})

		t.emit(ctx, writer, state, "response.reasoning_summary_part.done", map[string]any{
			"type":          "response.reasoning_summary_part.done",
			"item_id":       state.reasoningID,
			"output_index":  state.reasoningIndex,
			"summary_index": 0,
			"part": map[string]any{
				"type": "summary_text",
				"text": state.reasoningBuf,
			},
		})

		t.emit(ctx, writer, state, "response.output_item.done", map[string]any{
			"type":         "response.output_item.done",
			"output_index": state.reasoningIndex,
			"item": map[string]any{
				"id":   state.reasoningID,
				"type": "reasoning",
				"summary": []map[string]any{
					{"type": "summary_text", "text": state.reasoningBuf},
				},
			},
		})
	}
}

// handleContent handles text content
// ref: open-sse/transformer/responsesTransformer.js:308-371
func (t *Transformer) handleContent(ctx context.Context, writer stream.StreamWriter, state *transformerState, idx int, content string) {
	// Handle <think> tags
	if strings.Contains(content, "<think>") {
		state.inThinking = true
		content = strings.Replace(content, "<think>", "", -1)
		t.startReasoning(ctx, writer, state, idx)
	}

	if strings.Contains(content, "</think>") {
		parts := strings.SplitN(content, "</think>", 2)
		thinkPart := parts[0]
		textPart := ""
		if len(parts) > 1 {
			textPart = parts[1]
		}

		if thinkPart != "" {
			t.emitReasoningDelta(ctx, writer, state, thinkPart)
		}
		t.closeReasoning(ctx, writer, state)
		state.inThinking = false
		content = textPart
	}

	if state.inThinking && content != "" {
		t.emitReasoningDelta(ctx, writer, state, content)
		return
	}

	// Regular text content
	if content != "" {
		t.emitTextContent(ctx, writer, state, idx, content)
	}
}

// emitTextContent emits text content events
// ref: open-sse/transformer/responsesTransformer.js:335-370
func (t *Transformer) emitTextContent(ctx context.Context, writer stream.StreamWriter, state *transformerState, idx int, content string) {
	msgID := fmt.Sprintf("msg_%s_%d", state.responseID, idx)

	if !state.msgItemAdded[idx] {
		state.msgItemAdded[idx] = true

		t.emit(ctx, writer, state, "response.output_item.added", map[string]any{
			"type":         "response.output_item.added",
			"output_index": idx,
			"item": map[string]any{
				"id":      msgID,
				"type":    "message",
				"content": []any{},
				"role":    "assistant",
			},
		})
	}

	if !state.msgContentAdded[idx] {
		state.msgContentAdded[idx] = true

		t.emit(ctx, writer, state, "response.content_part.added", map[string]any{
			"type":          "response.content_part.added",
			"item_id":       msgID,
			"output_index":  idx,
			"content_index": 0,
			"part": map[string]any{
				"type":        "output_text",
				"annotations": []any{},
				"logprobs":    []any{},
				"text":        "",
			},
		})
	}

	t.emit(ctx, writer, state, "response.output_text.delta", map[string]any{
		"type":          "response.output_text.delta",
		"item_id":       msgID,
		"output_index":  idx,
		"content_index": 0,
		"delta":         content,
		"logprobs":      []any{},
	})

	if state.msgTextBuf[idx] == "" {
		state.msgTextBuf[idx] = ""
	}
	state.msgTextBuf[idx] += content
}

// closeMessage closes a message item
// ref: open-sse/transformer/responsesTransformer.js:160-194
func (t *Transformer) closeMessage(ctx context.Context, writer stream.StreamWriter, state *transformerState, idx int) {
	if state.msgItemAdded[idx] && !state.msgItemDone[idx] {
		state.msgItemDone[idx] = true
		fullText := state.msgTextBuf[idx]
		msgID := fmt.Sprintf("msg_%s_%d", state.responseID, idx)

		t.emit(ctx, writer, state, "response.output_text.done", map[string]any{
			"type":          "response.output_text.done",
			"item_id":       msgID,
			"output_index":  idx,
			"content_index": 0,
			"text":          fullText,
			"logprobs":      []any{},
		})

		t.emit(ctx, writer, state, "response.content_part.done", map[string]any{
			"type":          "response.content_part.done",
			"item_id":       msgID,
			"output_index":  idx,
			"content_index": 0,
			"part": map[string]any{
				"type":        "output_text",
				"annotations": []any{},
				"logprobs":    []any{},
				"text":        fullText,
			},
		})

		t.emit(ctx, writer, state, "response.output_item.done", map[string]any{
			"type":         "response.output_item.done",
			"output_index": idx,
			"item": map[string]any{
				"id":   msgID,
				"type": "message",
				"content": []map[string]any{
					{
						"type":        "output_text",
						"annotations": []any{},
						"logprobs":    []any{},
						"text":        fullText,
					},
				},
				"role": "assistant",
			},
		})
	}
}

// handleToolCalls handles tool calls
// ref: open-sse/transformer/responsesTransformer.js:373-415
func (t *Transformer) handleToolCalls(ctx context.Context, writer stream.StreamWriter, state *transformerState, toolCalls []any) {
	for _, tc := range toolCalls {
		tcMap, ok := tc.(map[string]any)
		if !ok {
			continue
		}

		tcIdx := 0
		if i, ok := tcMap["index"].(float64); ok {
			tcIdx = int(i)
		}

		newCallId, _ := tcMap["id"].(string)
		funcName := ""
		if fn, ok := tcMap["function"].(map[string]any); ok {
			funcName, _ = fn["name"].(string)
		}

		if funcName != "" {
			state.funcNames[tcIdx] = funcName
		}

		if state.funcCallIds[tcIdx] == "" && newCallId != "" {
			state.funcCallIds[tcIdx] = newCallId

			t.emit(ctx, writer, state, "response.output_item.added", map[string]any{
				"type":         "response.output_item.added",
				"output_index": tcIdx,
				"item": map[string]any{
					"id":        fmt.Sprintf("fc_%s", newCallId),
					"type":      "function_call",
					"arguments": "",
					"call_id":   newCallId,
					"name":      state.funcNames[tcIdx],
				},
			})
		}

		if state.funcArgsBuf[tcIdx] == "" {
			state.funcArgsBuf[tcIdx] = ""
		}

		if fn, ok := tcMap["function"].(map[string]any); ok {
			if args, ok := fn["arguments"].(string); ok && args != "" {
				refCallId := state.funcCallIds[tcIdx]
				if refCallId == "" {
					refCallId = newCallId
				}
				if refCallId != "" {
					t.emit(ctx, writer, state, "response.function_call_arguments.delta", map[string]any{
						"type":         "response.function_call_arguments.delta",
						"item_id":      fmt.Sprintf("fc_%s", refCallId),
						"output_index": tcIdx,
						"delta":        args,
					})
				}
				state.funcArgsBuf[tcIdx] += args
			}
		}
	}
}

// closeToolCall closes a tool call
// ref: open-sse/transformer/responsesTransformer.js:196-223
func (t *Transformer) closeToolCall(ctx context.Context, writer stream.StreamWriter, state *transformerState, idx int) {
	callId := state.funcCallIds[idx]
	if callId != "" && !state.funcItemDone[idx] {
		args := state.funcArgsBuf[idx]
		if args == "" {
			args = "{}"
		}

		t.emit(ctx, writer, state, "response.function_call_arguments.done", map[string]any{
			"type":         "response.function_call_arguments.done",
			"item_id":      fmt.Sprintf("fc_%s", callId),
			"output_index": idx,
			"arguments":    args,
		})

		t.emit(ctx, writer, state, "response.output_item.done", map[string]any{
			"type":         "response.output_item.done",
			"output_index": idx,
			"item": map[string]any{
				"id":        fmt.Sprintf("fc_%s", callId),
				"type":      "function_call",
				"arguments": args,
				"call_id":   callId,
				"name":      state.funcNames[idx],
			},
		})

		state.funcItemDone[idx] = true
		state.funcArgsDone[idx] = true
	}
}

// sendCompleted sends response.completed event
// ref: open-sse/transformer/responsesTransformer.js:225-240
func (t *Transformer) sendCompleted(ctx context.Context, writer stream.StreamWriter, state *transformerState) {
	if !state.completedSent {
		state.completedSent = true

		t.emit(ctx, writer, state, "response.completed", map[string]any{
			"type": "response.completed",
			"response": map[string]any{
				"id":         state.responseID,
				"object":     "response",
				"created_at": state.created,
				"status":     "completed",
				"background": false,
				"error":      nil,
			},
		})
	}
}

// flushItems flushes all remaining items
// ref: open-sse/transformer/responsesTransformer.js:427-436
func (t *Transformer) flushItems(ctx context.Context, writer stream.StreamWriter, state *transformerState) {
	for idx := range state.msgItemAdded {
		t.closeMessage(ctx, writer, state, idx)
	}
	t.closeReasoning(ctx, writer, state)
	for idx := range state.funcCallIds {
		t.closeToolCall(ctx, writer, state, idx)
	}
	t.sendCompleted(ctx, writer, state)
}

// closeAllItems closes all items for a given index
func (t *Transformer) closeAllItems(ctx context.Context, writer stream.StreamWriter, state *transformerState) {
	for idx := range state.msgItemAdded {
		t.closeMessage(ctx, writer, state, idx)
	}
	t.closeReasoning(ctx, writer, state)
	for idx := range state.funcCallIds {
		t.closeToolCall(ctx, writer, state, idx)
	}
}
