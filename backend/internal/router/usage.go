package router

import (
	"encoding/json"
	"strings"
)

// TokenUsage represents extracted token usage from an API response.
// ref: open-sse/services/usage.js:60
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	// RTK savings
	BytesBefore int `json:"bytes_before,omitempty"`
	BytesAfter  int `json:"bytes_after,omitempty"`
}

// ExtractUsage extracts token usage from a response body based on the format.
// ref: open-sse/services/usage.js:60
func ExtractUsage(body []byte, format string) (*TokenUsage, error) {
	switch strings.ToLower(format) {
	case "openai", "openai-format":
		return extractOpenAIUsage(body)
	case "claude", "anthropic":
		return extractClaudeUsage(body)
	case "gemini", "google":
		return extractGeminiUsage(body)
	default:
		// Try to auto-detect format
		return extractUsageAuto(body)
	}
}

// extractOpenAIUsage extracts usage from 0penAI format response.
// ref: open-sse/services/usage.js
func extractOpenAIUsage(body []byte) (*TokenUsage, error) {
	var resp struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	if resp.Usage.TotalTokens == 0 && resp.Usage.PromptTokens == 0 {
		return nil, nil
	}

	return &TokenUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}, nil
}

// extractClaudeUsage extracts usage from CLaude format response.
// ref: open-sse/services/usage.js
func extractClaudeUsage(body []byte) (*TokenUsage, error) {
	var resp struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Message struct {
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		} `json:"message"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	// Try top-level usage first, then message.usage
	inputTokens := resp.Usage.InputTokens
	outputTokens := resp.Usage.OutputTokens

	if inputTokens == 0 && resp.Message.Usage.InputTokens > 0 {
		inputTokens = resp.Message.Usage.InputTokens
		outputTokens = resp.Message.Usage.OutputTokens
	}

	if inputTokens == 0 && outputTokens == 0 {
		return nil, nil
	}

	return &TokenUsage{
		PromptTokens:     inputTokens,
		CompletionTokens: outputTokens,
		TotalTokens:      inputTokens + outputTokens,
	}, nil
}

// extractGeminiUsage extracts usage from Gemini format response.
// ref: open-sse/services/usage.js
func extractGeminiUsage(body []byte) (*TokenUsage, error) {
	var resp struct {
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	if resp.UsageMetadata.TotalTokenCount == 0 {
		return nil, nil
	}

	return &TokenUsage{
		PromptTokens:     resp.UsageMetadata.PromptTokenCount,
		CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      resp.UsageMetadata.TotalTokenCount,
	}, nil
}

// extractUsageAuto attempts to auto-detect the format and extract usage.
// ref: open-sse/services/usage.js
func extractUsageAuto(body []byte) (*TokenUsage, error) {
	// Try 0penAI format first (most common)
	if usage, err := extractOpenAIUsage(body); err == nil && usage != nil {
		return usage, nil
	}

	// Try CLaude format
	if usage, err := extractClaudeUsage(body); err == nil && usage != nil {
		return usage, nil
	}

	// Try Gemini format
	if usage, err := extractGeminiUsage(body); err == nil && usage != nil {
		return usage, nil
	}

	return nil, nil
}

// StreamUsageAccumulator accumulates token usage from SSE events.
// ref: open-sse/services/usage.js
type StreamUsageAccumulator struct {
	usage *TokenUsage
}

// NewStreamUsageAccumulator creates a new accumulator for streaming usage.
func NewStreamUsageAccumulator() *StreamUsageAccumulator {
	return &StreamUsageAccumulator{
		usage: &TokenUsage{},
	}
}

// ProcessEvent processes an SSE event line and extracts usage if present.
// ref: open-sse/services/usage.js
func (a *StreamUsageAccumulator) ProcessEvent(line []byte, format string) error {
	lineStr := strings.TrimSpace(string(line))
	if lineStr == "" || !strings.HasPrefix(lineStr, "data: ") {
		return nil
	}

	data := strings.TrimPrefix(lineStr, "data: ")
	if data == "[DONE]" {
		return nil
	}

	switch strings.ToLower(format) {
	case "openai", "openai-format":
		return a.processOpenAIEvent([]byte(data))
	case "claude", "anthropic":
		return a.processClaudeEvent([]byte(data))
	case "gemini", "google":
		return a.processGeminiEvent([]byte(data))
	default:
		// Try all formats
		_ = a.processOpenAIEvent([]byte(data))
		_ = a.processClaudeEvent([]byte(data))
		_ = a.processGeminiEvent([]byte(data))
		return nil
	}
}

// processOpenAIEvent extracts usage from 0penAI streaming events.
// ref: open-sse/services/usage.js
func (a *StreamUsageAccumulator) processOpenAIEvent(data []byte) error {
	var event struct {
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage,omitempty"`
	}

	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	if event.Usage != nil {
		a.usage.PromptTokens = event.Usage.PromptTokens
		a.usage.CompletionTokens = event.Usage.CompletionTokens
		a.usage.TotalTokens = event.Usage.TotalTokens
	}

	return nil
}

// processClaudeEvent extracts usage from CLaude streaming events.
// Usage comes from message_start and message_delta events.
// ref: open-sse/services/usage.js
func (a *StreamUsageAccumulator) processClaudeEvent(data []byte) error {
	var event struct {
		Type    string `json:"type"`
		Message *struct {
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		} `json:"message,omitempty"`
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage,omitempty"`
		Delta *struct {
			Type       string `json:"type"`
			StopReason string `json:"stop_reason"`
		} `json:"delta,omitempty"`
	}

	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	// message_start contains input tokens
	if event.Type == "message_start" && event.Message != nil {
		a.usage.PromptTokens = event.Message.Usage.InputTokens
	}

	// message_delta can contain output tokens
	if event.Type == "message_delta" && event.Usage != nil {
		a.usage.CompletionTokens = event.Usage.OutputTokens
		a.usage.TotalTokens = a.usage.PromptTokens + a.usage.CompletionTokens
	}

	return nil
}

// processGeminiEvent extracts usage from Gemini streaming events.
// ref: open-sse/services/usage.js
func (a *StreamUsageAccumulator) processGeminiEvent(data []byte) error {
	var event struct {
		UsageMetadata *struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata,omitempty"`
	}

	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	if event.UsageMetadata != nil {
		a.usage.PromptTokens = event.UsageMetadata.PromptTokenCount
		a.usage.CompletionTokens = event.UsageMetadata.CandidatesTokenCount
		a.usage.TotalTokens = event.UsageMetadata.TotalTokenCount
	}

	return nil
}

// GetUsage returns the accumulated usage.
func (a *StreamUsageAccumulator) GetUsage() *TokenUsage {
	if a.usage.TotalTokens == 0 && a.usage.PromptTokens == 0 && a.usage.CompletionTokens == 0 {
		return nil
	}
	return a.usage
}

// SetRTKSavings sets the RTK savings (bytes saved by compression).
func (u *TokenUsage) SetRTKSavings(before, after int) {
	u.BytesBefore = before
	u.BytesAfter = after
}

// RTKSavings returns the bytes saved by RTK compression.
func (u *TokenUsage) RTKSavings() int {
	if u.BytesBefore == 0 {
		return 0
	}
	return u.BytesBefore - u.BytesAfter
}
