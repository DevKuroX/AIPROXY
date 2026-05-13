package request

type OpenAIRequest struct {
	Model            string                `json:"model"`
	Messages         []OpenAIMessage       `json:"messages"`
	MaxTokens        int                   `json:"max_tokens,omitempty"`
	Temperature      *float64              `json:"temperature,omitempty"`
	TopP             *float64              `json:"top_p,omitempty"`
	Stop             []string              `json:"stop,omitempty"`
	Stream           bool                  `json:"stream,omitempty"`
	Tools            []OpenAITool          `json:"tools,omitempty"`
	ToolChoice       interface{}           `json:"tool_choice,omitempty"`
	ResponseFormat   *OpenAIResponseFormat `json:"response_format,omitempty"`
	Thinking         *OpenAIThinking       `json:"thinking,omitempty"`
	ReasoningEffort  string                `json:"reasoning_effort,omitempty"`
	FrequencyPenalty *float64              `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64              `json:"presence_penalty,omitempty"`
	N                int                   `json:"n,omitempty"`
	User             string                `json:"user,omitempty"`
}

type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    interface{}      `json:"content,omitempty"`
	Name       string           `json:"name,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type OpenAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function OpenAIFunctionCall `json:"function"`
}

type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAITool struct {
	Type     string             `json:"type"`
	Function *OpenAIFunctionDef `json:"function,omitempty"`
}

type OpenAIFunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type OpenAIResponseFormat struct {
	Type       string            `json:"type"`
	JSONSchema *OpenAIJSONSchema `json:"json_schema,omitempty"`
}

type OpenAIJSONSchema struct {
	Name   string                 `json:"name"`
	Schema map[string]interface{} `json:"schema"`
}

type OpenAIThinking struct {
	Type         string `json:"type,omitempty"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
	MaxTokens    int    `json:"max_tokens,omitempty"`
}

type OpenAIContentPart struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ImageURL  *OpenAIImageURL        `json:"image_url,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
	IsError   bool                   `json:"is_error,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     interface{}            `json:"input,omitempty"`
	Source    map[string]interface{} `json:"source,omitempty"`
}

type OpenAIImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type OpenAIToolChoiceAuto struct {
	Type string `json:"type"`
}

type OpenAIToolChoiceNone struct {
	Type string `json:"type"`
}

type OpenAIToolChoiceFunction struct {
	Type     string `json:"type"`
	Function struct {
		Name string `json:"name"`
	} `json:"function"`
}
