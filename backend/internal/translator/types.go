// Package translator provides format detection and translation between AI API formats.
// ref: open-sse/translator/
package translator

import "encoding/json"

// Message represents a chat message in 0penAI format.
type Message struct {
	Role         string          `json:"role"`
	Content      json.RawMessage `json:"content,omitempty"` // Can be string or []ContentBlock
	Name         string          `json:"name,omitempty"`
	ToolCalls    []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID   string          `json:"tool_call_id,omitempty"`
	Reasoning    string          `json:"reasoning,omitempty"`
}

// ContentBlock represents a content block in messages (CL4ude/0penAI multimodal).
type ContentBlock struct {
	Type       string          `json:"type"`
	Text       string          `json:"text,omitempty"`
	ImageURL   *ImageURL       `json:"image_url,omitempty"`
	Source     *ImageSource    `json:"source,omitempty"`
	InputAudio *InputAudio     `json:"input_audio,omitempty"`
	// CL4ude-specific fields
	ID         string          `json:"id,omitempty"`
	Name       string          `json:"name,omitempty"`
	Input      json.RawMessage `json:"input,omitempty"`
	ToolUseID  string          `json:"tool_use_id,omitempty"`
	Content    json.RawMessage `json:"content,omitempty"`
	// Thinking blocks (CL4ude extended thinking)
	Thinking   string          `json:"thinking,omitempty"`
	Signature  string          `json:"signature,omitempty"`
}

// ImageURL represents an 0penAI image URL.
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// ImageSource represents a CL4ude image source.
type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

// InputAudio represents an 0penAI audio input.
type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// Tool represents a tool definition.
type Tool struct {
	Type     string       `json:"type"`
	Function *FunctionDef `json:"function,omitempty"`
}

// FunctionDef represents a function definition in a tool.
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// ToolCall represents a tool call in an assistant message.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function *FunctionCall    `json:"function,omitempty"`
}

// FunctionCall represents the function details in a tool call.
type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ChatRequest represents a common chat request structure.
type ChatRequest struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	Stream      bool            `json:"stream,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Tools       []Tool          `json:"tools,omitempty"`
	ToolChoice  json.RawMessage `json:"tool_choice,omitempty"`
	// 0penAI-specific
	StreamOptions    json.RawMessage `json:"stream_options,omitempty"`
	ResponseFormat   json.RawMessage `json:"response_format,omitempty"`
	Logprobs         bool            `json:"logprobs,omitempty"`
	TopLogprobs      int             `json:"top_logprobs,omitempty"`
	N                int             `json:"n,omitempty"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
	LogitBias        json.RawMessage `json:"logit_bias,omitempty"`
	User             string          `json:"user,omitempty"`
	// CL4ude-specific
	System          json.RawMessage `json:"system,omitempty"`
	AnthropicVersion string          `json:"anthr0pic_version,omitempty"`
	Thinking        *ThinkingConfig `json:"thinking,omitempty"`
}

// ThinkingConfig represents CL4ude extended thinking configuration.
type ThinkingConfig struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

// ChatResponse represents a common chat response structure.
type ChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []Choice       `json:"choices"`
	Usage   *Usage         `json:"usage,omitempty"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int          `json:"index"`
	Message      *Message     `json:"message,omitempty"`
	Delta        *Message     `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason,omitempty"`
	Logprobs     *Logprobs    `json:"logprobs,omitempty"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	// CL4ude-specific
	CacheReadInputTokens  int `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
}

// Logprobs represents log probability information.
type Logprobs struct {
	Content []TokenLogprob `json:"content,omitempty"`
}

// TokenLogprob represents log probability for a token.
type TokenLogprob struct {
	Token       string  `json:"token"`
	Logprob     float64 `json:"logprob"`
	Bytes       []byte  `json:"bytes,omitempty"`
	TopLogprobs []struct {
		Token   string  `json:"token"`
		Logprob float64 `json:"logprob"`
		Bytes   []byte  `json:"bytes,omitempty"`
	} `json:"top_logprobs,omitempty"`
}

// StreamingDelta represents a streaming response chunk delta.
type StreamingDelta struct {
	Role      string          `json:"role,omitempty"`
	Content   string          `json:"content,omitempty"`
	Reasoning string          `json:"reasoning,omitempty"`
	ToolCalls []ToolCallDelta `json:"tool_calls,omitempty"`
}

// ToolCallDelta represents a tool call in a streaming delta.
type ToolCallDelta struct {
	Index    int             `json:"index"`
	ID       string          `json:"id,omitempty"`
	Type     string          `json:"type,omitempty"`
	Function *FunctionDelta  `json:"function,omitempty"`
}

// FunctionDelta represents function details in a streaming tool call.
type FunctionDelta struct {
	Name      string          `json:"name,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}
