package response

import (
	"encoding/json"
	"fmt"
)

// ref: open-sse/translator/response/claude-to-openai.js:5-17
type OpenAIResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []OpenAIChoice     `json:"choices"`
	Usage   OpenAIUsage        `json:"usage,omitempty"`
}

// ref: open-sse/translator/response/claude-to-openai.js:11-15
type OpenAIChoice struct {
	Index        int            `json:"index"`
	Message      OpenAIMessage  `json:"message"`
	FinishReason string         `json:"finish_reason,omitempty"`
}

// ref: open-sse/translator/response/claude-to-openai.js:11-14
type OpenAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}

// ref: open-sse/translator/response/claude-to-openai.js:52-60
type OpenAIToolCall struct {
	Index    int                `json:"index"`
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function OpenAIFunctionCall `json:"function"`
}

// ref: open-sse/translator/response/claude-to-openai.js:55-59
type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ref: open-sse/translator/response/claude-to-openai.js:141-145
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ref: open-sse/translator/response/claude-to-openai.js:194-202
func convertStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	case "stop_sequence":
		return "stop"
	default:
		return "stop"
	}
}

// ref: open-sse/translator/response/claude-to-openai.js:20-191
func TranslateResponse(body *CLaudeResponse) (*OpenAIResponse, error) {
	if body == nil {
		return nil, nil
	}

	var content string
	var toolCalls []OpenAIToolCall
	toolCallIndex := 0

	// ref: open-sse/translator/response/claude-to-openai.js:36-63
	for _, block := range body.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "tool_use":
			// ref: open-sse/translator/response/claude-to-openai.js:48-62
			args := "{}"
			if len(block.Input) > 0 {
				args = string(block.Input)
			}
			toolCalls = append(toolCalls, OpenAIToolCall{
				Index: toolCallIndex,
				ID:    block.ID,
				Type:  "function",
				Function: OpenAIFunctionCall{
					Name:      block.Name,
					Arguments: args,
				},
			})
			toolCallIndex++
		}
	}

	// ref: open-sse/translator/response/claude-to-openai.js:109-128
	promptTokens := body.Usage.InputTokens + body.Usage.CacheReadInputTokens + body.Usage.CacheCreationInputTokens
	completionTokens := body.Usage.OutputTokens

	// ref: open-sse/translator/response/claude-to-openai.js:130-159
	finishReason := convertStopReason(body.StopReason)

	message := OpenAIMessage{
		Role:    "assistant",
		Content: content,
	}
	if len(toolCalls) > 0 {
		message.ToolCalls = toolCalls
	}

	// ref: open-sse/translator/response/claude-to-openai.js:5-17
	resp := &OpenAIResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", body.ID),
		Object:  "chat.completion",
		Model:   body.Model,
		Choices: []OpenAIChoice{
			{
				Index:        0,
				Message:      message,
				FinishReason: finishReason,
			},
		},
		Usage: OpenAIUsage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}

	return resp, nil
}

// TranslateResponseFromJSON parses JSON and translates to OpenAI format
func TranslateResponseFromJSON(data []byte) (*OpenAIResponse, error) {
	var claudeResp CLaudeResponse
	if err := json.Unmarshal(data, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CLaude response: %w", err)
	}
	return TranslateResponse(&claudeResp)
}
