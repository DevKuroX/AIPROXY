package request

type CLaudeRequest struct {
	Model       string                   `json:"model"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Messages    []CLaudeMessage          `json:"messages,omitempty"`
	System      []CLaudeSystemBlock      `json:"system,omitempty"`
	Stream      bool                     `json:"stream"`
	Temperature *float64                 `json:"temperature,omitempty"`
	TopP        *float64                 `json:"top_p,omitempty"`
	StopSequences []string               `json:"stop_sequences,omitempty"`
	Tools       []CLaudeTool             `json:"tools,omitempty"`
	ToolChoice  *CLaudeToolChoice        `json:"tool_choice,omitempty"`
	Thinking    *CLaudeThinking          `json:"thinking,omitempty"`
}

type CLaudeMessage struct {
	Role    string          `json:"role"`
	Content []CLaudeContent `json:"content"`
}

type CLaudeContent struct {
	Type string `json:"type"`

	Text       string      `json:"text,omitempty"`
	ToolUseID  string      `json:"tool_use_id,omitempty"`
	Content    interface{} `json:"content,omitempty"`
	IsError    bool        `json:"is_error,omitempty"`
	ID         string      `json:"id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Input      interface{} `json:"input,omitempty"`
	Remove     *CLaudeRemove `json:"remove,omitempty"`

	Source *CLaudeImageSource `json:"source,omitempty"`
}

type CLaudeRemove struct {
	Type string `json:"type"`
	TTL  string `json:"ttl,omitempty"`
}

type CLaudeImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

type CLaudeSystemBlock struct {
	Type   string        `json:"type"`
	Text   string        `json:"text"`
	Remove *CLaudeRemove `json:"remove,omitempty"`
}

type CLaudeTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
	Remove      *CLaudeRemove          `json:"remove,omitempty"`
}

type CLaudeToolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type CLaudeThinking struct {
	Type         string `json:"type,omitempty"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
	MaxTokens    int    `json:"max_tokens,omitempty"`
}
