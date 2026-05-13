package chat

import (
	"time"
)

// ChatRequest represents a chat completion request
// ref: open-sse/handlers/chatCore/requestDetail.js - extractRequestDetails
type ChatRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Stream           bool      `json:"stream,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	User             string    `json:"user,omitempty"`
	LogitBias        map[string]float64 `json:"logit_bias,omitempty"`
	LogProbs         bool      `json:"logprobs,omitempty"`
	TopLogProbs      int       `json:"top_logprobs,omitempty"`
	ResponseFormat   struct {
		Type string `json:"type,omitempty"`
	} `json:"response_format,omitempty"`
	Seed             int64     `json:"seed,omitempty"`
	Tools            []Tool    `json:"tools,omitempty"`
	ToolChoice       any       `json:"tool_choice,omitempty"`
	ParallelToolCalls bool     `json:"parallel_tool_calls,omitempty"`
	Store            bool      `json:"store,omitempty"`
}

// Message represents a single message in a chat conversation
// ref: open-sse/translator/formats.js - message structure
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []ContentPart
	Name    string `json:"name,omitempty"`
}

// ContentPart represents a part of message content (text or image)
// ref: open-sse/translator/formats.js - contentPart structure
type ContentPart struct {
	Type     string `json:"type"` // "text" or "image_url"
	Text     string `json:"text,omitempty"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

// Tool represents a tool that can be called by the model
// ref: open-sse/translator/formats.js - tool structure
type Tool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Parameters  any    `json:"parameters,omitempty"`
	} `json:"function,omitempty"`
}

// ChatResponse represents a non-streaming chat completion response
// ref: open-sse/handlers/chatCore/sseToJsonHandler.js - buildFinalResponse
type ChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	SystemFingerprint string `json:"system_fingerprint,omitempty"`
}

// Choice represents a non-streaming choice
// ref: open-sse/handlers/chatCore/sseToJsonHandler.js - choice structure
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	LogProbs     any     `json:"logprobs,omitempty"`
	FinishReason string  `json:"finish_reason"`
}

// StreamChoice represents a streaming choice
// ref: open-sse/utils/stream.js - streaming choice structure
type StreamChoice struct {
	Index        int            `json:"index"`
	Delta        Message        `json:"delta"`
	LogProbs     any            `json:"logprobs,omitempty"`
	FinishReason string         `json:"finish_reason,omitempty"`
}

// Usage represents token usage statistics
// ref: open-sse/utils/stream.js - extractUsageFromStream
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// RequestDetail represents detailed logging for a request
// ref: open-sse/handlers/chatCore/requestDetail.js - RequestDetail type
type RequestDetail struct {
	ID               string          `json:"id"`
	Timestamp        time.Time       `json:"timestamp"`
	Method           string          `json:"method"`
	Path             string          `json:"path"`
	Headers          map[string]string `json:"headers,omitempty"`
	Body             any             `json:"body,omitempty"`
	Response         any             `json:"response,omitempty"`
	StatusCode       int             `json:"status_code,omitempty"`
	Duration         time.Duration   `json:"duration_ms"`
	Error            string          `json:"error,omitempty"`
	ProviderID       int64           `json:"provider_id,omitempty"`
	AccountID        int64           `json:"account_id,omitempty"`
	Model            string          `json:"model,omitempty"`
	TokensPrompt     int             `json:"tokens_prompt,omitempty"`
	TokensCompletion int             `json:"tokens_completion,omitempty"`
}

// RequestDetailFilters represents filters for querying request details
// ref: open-sse/handlers/chatCore/requestDetail.js - filter structure
type RequestDetailFilters struct {
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	ProviderID  int64     `json:"provider_id,omitempty"`
	AccountID   int64     `json:"account_id,omitempty"`
	Model       string    `json:"model,omitempty"`
	StatusCode  int       `json:"status_code,omitempty"`
	HasError    bool      `json:"has_error,omitempty"`
	Limit       int       `json:"limit,omitempty"`
	Offset      int       `json:"offset,omitempty"`
}