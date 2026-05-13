package stream

import (
	"context"
	"time"
)

// StreamChunk represents a single SSE chunk
// ref: open-sse/utils/stream.js - SSE chunk structure
type StreamChunk struct {
	Event string // "message", "error", "done", etc.
	Data  []byte // Raw data bytes
	ID    string // SSE id field
	Retry int    // SSE retry field
}

// StreamReader interface for reading SSE chunks
// ref: open-sse/utils/stream.js - createSSEReader
type StreamReader interface {
	ReadChunk(ctx context.Context) (*StreamChunk, error)
	Close() error
}

// StreamWriter interface for writing SSE chunks
// ref: open-sse/utils/stream.js - createSSEWriter
type StreamWriter interface {
	WriteChunk(ctx context.Context, chunk *StreamChunk) error
	Flush() error
	Close() error
}

// StreamFormat represents different streaming formats
// ref: open-sse/translator/formats.js - FORMATS enum
type StreamFormat string

const (
	FormatOpenAI          StreamFormat = "openai"
	FormatOpenAIResponses StreamFormat = "openai-responses"
	FormatCL4ude          StreamFormat = "claude"
	FormatCL4udeMessages  StreamFormat = "claude-messages"
	FormatGemini          StreamFormat = "gemini"
	FormatGeminiCLI       StreamFormat = "gemini-cli"
	FormatAntigravity     StreamFormat = "antigravity"
)

// Usage represents token usage statistics
// ref: open-sse/utils/stream.js - extractUsageFromStream
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamConfig holds configuration for stream processing
// ref: open-sse/handlers/chatCore/requestDetail.js - extractRequestConfig
type StreamConfig struct {
	Provider         string
	Model            string
	SourceFormat     StreamFormat
	TargetFormat     StreamFormat
	UserAgent        string
	ConnectionID     string
	APIKey           string
	RequestStartTime time.Time
	Stream           bool
}

// ChatRequest represents a chat completion request
// ref: open-sse/handlers/chatCore/requestDetail.js - extractRequestDetails
type ChatRequest struct {
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	Stream      bool                   `json:"stream,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	FrequencyPenalty float64           `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64           `json:"presence_penalty,omitempty"`
	Stop        []string               `json:"stop,omitempty"`
	User        string                 `json:"user,omitempty"`
	// ... all OpenAI fields
}

// Message represents a single message in a chat conversation
// ref: open-sse/translator/formats.js - message structure
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
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
}

// Choice represents a non-streaming choice
// ref: open-sse/handlers/chatCore/sseToJsonHandler.js - choice structure
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// StreamChoice represents a streaming choice
// ref: open-sse/utils/stream.js - streaming choice structure
type StreamChoice struct {
	Index        int            `json:"index"`
	Delta        Message        `json:"delta"`
	FinishReason string         `json:"finish_reason,omitempty"`
}
