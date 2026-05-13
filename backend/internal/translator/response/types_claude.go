package response

import "encoding/json"

// ref: open-sse/translator/response/claude-to-openai.js:27-33
type CLaudeMessage struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Role         string               `json:"role"`
	Model        string               `json:"model"`
	Content      []CLaudeContentBlock `json:"content"`
	StopReason   string               `json:"stop_reason"`
	StopSequence string               `json:"stop_sequence,omitempty"`
	Usage        CLaudeUsage          `json:"usage"`
}

// ref: open-sse/translator/response/claude-to-openai.js:36-63
type CLaudeContentBlock struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	Thinking string          `json:"thinking,omitempty"`
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name,omitempty"`
	Input    json.RawMessage `json:"input,omitempty"`
}

// ref: open-sse/translator/response/claude-to-openai.js:109-128
type CLaudeUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
}

// ref: open-sse/translator/response/claude-to-openai.js:27-33
type CLaudeResponse struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Role         string               `json:"role"`
	Model        string               `json:"model"`
	Content      []CLaudeContentBlock `json:"content"`
	StopReason   string               `json:"stop_reason"`
	StopSequence string               `json:"stop_sequence,omitempty"`
	Usage        CLaudeUsage          `json:"usage"`
}
