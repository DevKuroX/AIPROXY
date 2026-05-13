// Package translator provides format detection and translation between AI API formats.
// ref: open-sse/translator/formats.js
package translator

import (
	"encoding/json"
	"strings"
)

// Format constants for different AI API request/response formats.
// ref: open-sse/translator/formats.js:2-16
type Format string

const (
	// FormatOpenAI is the standard OpenAI chat completions format.
	FormatOpenAI Format = "openai"
	// FormatOpenAIResponses is the OpenAI Responses API format.
	FormatOpenAIResponses Format = "openai-responses"
	// FormatCL4ude is the Anthropic Claude messages format.
	FormatCL4ude Format = "claude"
	// FormatGemini is the Google Gemini generateContent format.
	FormatGemini Format = "gemini"
	// FormatGeminiCLI is the Gemini CLI format.
	FormatGeminiCLI Format = "gemini-cli"
	// FormatVertex is the Google Vertex AI format.
	FormatVertex Format = "vertex"
	// FormatAntigravity is the Antigravity format (Gemini wrapped).
	FormatAntigravity Format = "antigravity"
	// FormatKiro is the Kiro format.
	FormatKiro Format = "kiro"
	// FormatCursor is the Cursor format.
	FormatCursor Format = "cursor"
	// FormatOllama is the Ollama format.
	FormatOllama Format = "ollama"
	// FormatCommandCode is the CommandCode format.
	FormatCommandCode Format = "commandcode"
)

// detectRequest represents the parsed JSON structure for format detection.
type detectRequest struct {
	Messages         []rawMessage `json:"messages"`
	Contents         []any        `json:"contents"`
	Input            any          `json:"input"`
	System           any          `json:"system"`
	AnthropicVersion string       `json:"anthropic_version"`
	Model            string       `json:"model"`
	UserAgent        string       `json:"userAgent"`
	Request          *struct {
		Contents []any `json:"contents"`
	} `json:"request"`
	// OpenAI-specific fields
	StreamOptions    any `json:"stream_options"`
	ResponseFormat   any `json:"response_format"`
	Logprobs         any `json:"logprobs"`
	TopLogprobs      any `json:"top_logprobs"`
	N                any `json:"n"`
	PresencePenalty  any `json:"presence_penalty"`
	FrequencyPenalty any `json:"frequency_penalty"`
	LogitBias        any `json:"logit_bias"`
	User             any `json:"user"`
}

// rawMessage represents a message in either OpenAI or Claude format.
type rawMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // can be string or []contentBlock
}

// contentBlock represents a content block in Claude or OpenAI multimodal format.
type contentBlock struct {
	Type    string `json:"type"`
	Source  *struct {
		Type string `json:"type"`
	} `json:"source"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

// DetectFormat detects the request format by inspecting the body structure.
// ref: open-sse/services/provider.js:49-126
func DetectFormat(body []byte) Format {
	var req detectRequest
	// If we can't parse, default to OpenAI
	if err := unmarshalDetect(body, &req); err != nil {
		return FormatOpenAI
	}

	// OpenAI Responses API: has input (array or string) instead of messages[]
	// ref: open-sse/services/provider.js:50-54
	if req.Input != nil && req.Messages == nil {
		switch req.Input.(type) {
		case []any, string:
			return FormatOpenAIResponses
		}
	}

	// Antigravity format: Gemini wrapped in body.request
	// ref: open-sse/services/provider.js:56-59
	if req.Request != nil && len(req.Request.Contents) > 0 && req.UserAgent == "antigravity" {
		return FormatAntigravity
	}

	// Gemini format: has contents array
	// ref: open-sse/services/provider.js:61-64
	if len(req.Contents) > 0 {
		return FormatGemini
	}

	// OpenAI-specific indicators (check BEFORE Claude)
	// ref: open-sse/services/provider.js:66-80
	if req.StreamOptions != nil ||
		req.ResponseFormat != nil ||
		req.Logprobs != nil ||
		req.TopLogprobs != nil ||
		req.N != nil ||
		req.PresencePenalty != nil ||
		req.FrequencyPenalty != nil ||
		req.LogitBias != nil ||
		req.User != nil {
		return FormatOpenAI
	}

	// Claude format: messages with content as array of objects with type
	// ref: open-sse/services/provider.js:82-122
	if len(req.Messages) > 0 {
		firstMsg := req.Messages[0]

		// If content is array, check if it follows Claude structure
		if contentArr, ok := firstMsg.Content.([]any); ok && len(contentArr) > 0 {
			if firstBlock, ok := contentArr[0].(map[string]any); ok {
				if blockType, _ := firstBlock["type"].(string); blockType == "text" {
					// Check for Claude-specific fields
					if req.System != nil || req.AnthropicVersion != "" {
						return FormatCL4ude
					}

					// Check image format: Claude (source.type) vs OpenAI (image_url.url)
					hasClaudeImage := false
					hasOpenAIImage := false
					for _, c := range contentArr {
						if cb, ok := c.(map[string]any); ok {
							if t, _ := cb["type"].(string); t == "image" {
								if src, ok := cb["source"].(map[string]any); ok {
									if srcType, _ := src["type"].(string); srcType == "base64" {
										hasClaudeImage = true
									}
								}
							}
							if t, _ := cb["type"].(string); t == "image_url" {
								if _, ok := cb["image_url"].(map[string]any); ok {
									hasOpenAIImage = true
								}
							}
							// Check for Claude tool format
							if t, _ := cb["type"].(string); t == "tool_use" || t == "tool_result" {
								return FormatCL4ude
							}
						}
					}

					if hasClaudeImage {
						return FormatCL4ude
					}
					if hasOpenAIImage {
						return FormatOpenAI
					}

					// Model path check: Claude models don't have "/" in name
					if strings.Contains(req.Model, "/") {
						return FormatOpenAI
					}
				}
			}
		}

		// If content is string or other indicators
		if req.System != nil || req.AnthropicVersion != "" {
			return FormatCL4ude
		}
	}

	// Default to OpenAI format
	// ref: open-sse/services/provider.js:124-125
	return FormatOpenAI
}

// DetectFormatByEndpoint detects format from URL pathname + body.
// Returns empty string to fall back to body-based detection.
// ref: open-sse/translator/formats.js:22-35
func DetectFormatByEndpoint(path string, body []byte) Format {
	// /v1/responses is always openai-responses
	if strings.Contains(path, "/v1/responses") {
		return FormatOpenAIResponses
	}

	// /v1/messages is always Claude
	if strings.Contains(path, "/v1/messages") {
		return FormatCL4ude
	}

	// /v1/chat/completions + input[] → treat as openai (Cursor CLI sends Responses body via chat endpoint)
	if strings.Contains(path, "/v1/chat/completions") {
		var req detectRequest
		if err := unmarshalDetect(body, &req); err == nil {
			if _, ok := req.Input.([]any); ok && len(req.Messages) == 0 {
				return FormatOpenAI
			}
		}
	}

	// Return empty to indicate fallback to body-based detection
	return ""
}

// unmarshalDetect is a helper to unmarshal JSON for detection purposes.
func unmarshalDetect(body []byte, v *detectRequest) error {
	return json.Unmarshal(body, v)
}
