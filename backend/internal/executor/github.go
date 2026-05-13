// Package executor provides provider-specific request/response handling.
// ref: _ref/9router/open-sse/executors/github.js
package executor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// GitHub Copilot constants
// ref: open-sse/executors/github.js:28-33
const (
	GitHubCopilotVSCodeVersion      = "1.100.0"
	GitHubCopilotCopilotChatVersion = "0.27.0"
	GitHubCopilotUserAgent          = "GitHubCopilotChat/" + GitHubCopilotCopilotChatVersion
	GitHubCopilotAPIVersion         = "2025-04-01"
	GitHubCopilotIntegrationID      = "vscode-chat"
)

// GitHubExecutor handles GitHub Copilot-specific request transformations.
// ref: open-sse/executors/github.js:12-16
type GitHubExecutor struct {
	BaseExecutor
	knownCodexModels map[string]bool
	codexModelsMu    sync.RWMutex
	baseURL          string
	responsesURL     string
	clientID         string
	clientSecret     string
}

// NewGitHubExecutor creates a new GitHub executor.
// ref: open-sse/executors/github.js:13-16
func NewGitHubExecutor() *GitHubExecutor {
	baseURL := os.Getenv("GITHUB_COPILOT_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.githubcopilot.com/chat/completions"
	}
	responsesURL := os.Getenv("GITHUB_COPILOT_RESPONSES_URL")
	if responsesURL == "" {
		responsesURL = "https://api.githubcopilot.com/responses"
	}
	return &GitHubExecutor{
		BaseExecutor:     NewBaseExecutor("github"),
		knownCodexModels: make(map[string]bool),
		baseURL:          baseURL,
		responsesURL:     responsesURL,
		clientID:         os.Getenv("GITHUB_CLIENT_ID"),
		clientSecret:     os.Getenv("GITHUB_CLIENT_SECRET"),
	}
}

// GitHubCredentials holds GitHub Copilot credential data.
type GitHubCredentials struct {
	AccessToken            string
	RefreshToken           string
	CopilotToken           string
	CopilotTokenExpiresAt  int64
	CopilotTokenExpiresStr string // ISO string format
}

// PrepareRequest transforms the request for GitHub Copilot API.
// ref: open-sse/executors/github.js:18-38
func (e *GitHubExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Extract credentials from context
	creds := extractGitHubCredentials(ctx)

	// Build URL - use base URL for chat completions
	// ref: open-sse/executors/github.js:18-20
	req.URL, _ = url.Parse(e.baseURL)

	// Set GitHub Copilot-specific headers
	// ref: open-sse/executors/github.js:22-38
	token := creds.CopilotToken
	if token == "" {
		token = creds.AccessToken
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("copilot-integration-id", GitHubCopilotIntegrationID)
	req.Header.Set("editor-version", "vscode/"+GitHubCopilotVSCodeVersion)
	req.Header.Set("editor-plugin-version", "copilot-chat/"+GitHubCopilotCopilotChatVersion)
	req.Header.Set("user-agent", GitHubCopilotUserAgent)
	req.Header.Set("openai-intent", "conversation-panel")
	req.Header.Set("x-github-api-version", GitHubCopilotAPIVersion)
	req.Header.Set("x-request-id", generateRequestID())
	req.Header.Set("x-vscode-user-agent-library-version", "electron-fetch")
	req.Header.Set("X-Initiator", "user")

	// Check if streaming based on Accept header or query param
	stream := req.Header.Get("Accept") == "text/event-stream"
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}

	// Transform request body if needed
	if len(body) > 0 {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			// Get model from body
			model, _ := bodyMap["model"].(string)

			// Transform request
			transformed := e.transformRequest(model, bodyMap)

			// Sanitize messages for chat completions
			sanitized := e.sanitizeMessagesForChatCompletions(transformed)

			// Re-encode body
			if newBody, err := json.Marshal(sanitized); err == nil {
				req.Body = newBodyReader(newBody)
				req.ContentLength = int64(len(newBody))
			}
		}
	}

	return nil
}

// TransformResponse modifies the response from GitHub Copilot.
func (e *GitHubExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	// For now, return response body as-is
	// The JavaScript version handles streaming transformation for /responses endpoint
	return readResponseBody(resp)
}

// HandleError processes errors from GitHub Copilot API.
func (e *GitHubExecutor) HandleError(ctx context.Context, err error) error {
	return fmt.Errorf("github copilot error: %w", err)
}

// buildHeaders returns headers for GitHub Copilot requests.
// ref: open-sse/executors/github.js:22-38
func (e *GitHubExecutor) buildHeaders(creds *GitHubCredentials, stream bool) http.Header {
	headers := make(http.Header)

	token := creds.CopilotToken
	if token == "" {
		token = creds.AccessToken
	}

	headers.Set("Authorization", "Bearer "+token)
	headers.Set("Content-Type", "application/json")
	headers.Set("copilot-integration-id", GitHubCopilotIntegrationID)
	headers.Set("editor-version", "vscode/"+GitHubCopilotVSCodeVersion)
	headers.Set("editor-plugin-version", "copilot-chat/"+GitHubCopilotCopilotChatVersion)
	headers.Set("user-agent", GitHubCopilotUserAgent)
	headers.Set("openai-intent", "conversation-panel")
	headers.Set("x-github-api-version", GitHubCopilotAPIVersion)
	headers.Set("x-request-id", generateRequestID())
	headers.Set("x-vscode-user-agent-library-version", "electron-fetch")
	headers.Set("X-Initiator", "user")

	if stream {
		headers.Set("Accept", "text/event-stream")
	} else {
		headers.Set("Accept", "application/json")
	}

	return headers
}

// requiresMaxCompletionTokens checks if model requires max_completion_tokens.
// ref: open-sse/executors/github.js:107-109
func (e *GitHubExecutor) requiresMaxCompletionTokens(model string) bool {
	matched, _ := regexp.MatchString(`(?i)gpt-5|o[134]-`, model)
	return matched
}

// supportsTemperature checks if model supports temperature parameter.
// ref: open-sse/executors/github.js:112-115
func (e *GitHubExecutor) supportsTemperature(model string) bool {
	// gpt-5.4 and similar newer models don't support temperature
	matched, _ := regexp.MatchString(`(?i)gpt-5\.4`, model)
	return !matched
}

// supportsThinking checks if model supports thinking payload.
// ref: open-sse/executors/github.js:120-122
func (e *GitHubExecutor) supportsThinking(model string) bool {
	// Claude models reject thinking payload on Copilot
	return !strings.Contains(strings.ToLower(model), "claude")
}

// supportsReasoningEffort checks if model supports reasoning_effort parameter.
// ref: open-sse/executors/github.js:127-135
func (e *GitHubExecutor) supportsReasoningEffort(model string) bool {
	m := strings.ToLower(model)
	// Claude Opus 4.6 and Sonnet 4.6 DO support reasoning_effort
	if strings.Contains(m, "claude") && strings.Contains(m, "opus") && strings.Contains(m, "4.6") {
		return true
	}
	if strings.Contains(m, "claude") && strings.Contains(m, "sonnet") && strings.Contains(m, "4.6") {
		return true
	}
	// All other Claude models: strip
	if strings.Contains(m, "claude") {
		return false
	}
	// GPT-5 family, Gemini, etc.: keep
	return true
}

// transformRequest transforms the request body for GitHub Copilot.
// ref: open-sse/executors/github.js:137-160
func (e *GitHubExecutor) transformRequest(model string, body map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range body {
		result[k] = v
	}

	// Convert max_tokens to max_completion_tokens for newer models
	if e.requiresMaxCompletionTokens(model) {
		if maxTokens, ok := result["max_tokens"]; ok {
			result["max_completion_tokens"] = maxTokens
			delete(result, "max_tokens")
		}
	}

	// Strip temperature for models that don't support it
	if !e.supportsTemperature(model) {
		delete(result, "temperature")
	}

	// Strip Claude-style thinking payload (Copilot doesn't understand it)
	if !e.supportsThinking(model) {
		delete(result, "thinking")
	}

	// "none" means no thinking — strip it
	if effort, ok := result["reasoning_effort"].(string); ok && effort == "none" {
		delete(result, "reasoning_effort")
	}

	// Strip reasoning_effort for models that reject it
	if !e.supportsReasoningEffort(model) {
		delete(result, "reasoning_effort")
	}

	return result
}

// sanitizeMessagesForChatCompletions sanitizes messages for the chat completions endpoint.
// ref: open-sse/executors/github.js:43-104
func (e *GitHubExecutor) sanitizeMessagesForChatCompletions(body map[string]interface{}) map[string]interface{} {
	messages, ok := body["messages"].([]interface{})
	if !ok {
		return body
	}

	result := make(map[string]interface{})
	for k, v := range body {
		result[k] = v
	}

	// Handle response_format for Claude models via GitHub
	// ref: open-sse/executors/github.js:48-76
	if responseFormat, ok := result["response_format"].(map[string]interface{}); ok {
		if model, ok := result["model"].(string); ok && strings.Contains(strings.ToLower(model), "claude") {
			var systemInstruction string
			if rfType, ok := responseFormat["type"].(string); ok {
				if rfType == "json_schema" {
					systemInstruction = "CRITICAL: You must ONLY output raw JSON. Never use markdown code blocks. Never use backticks. Never wrap JSON in triple backticks. Output ONLY the raw JSON object."
				} else if rfType == "json_object" {
					systemInstruction = "CRITICAL: You must ONLY output raw JSON. Never use markdown code blocks. Never use backticks."
				}
			}

			if systemInstruction != "" {
				// Add to system message
				systemIdx := -1
				for i, msg := range messages {
					if m, ok := msg.(map[string]interface{}); ok {
						if role, ok := m["role"].(string); ok && role == "system" {
							systemIdx = i
							break
						}
					}
				}

				if systemIdx >= 0 {
					if m, ok := messages[systemIdx].(map[string]interface{}); ok {
						if content, ok := m["content"].(string); ok {
							m["content"] = systemInstruction + "\n\n" + content
						}
					}
				} else {
					// Prepend system message
					messages = append([]interface{}{
						map[string]interface{}{"role": "system", "content": systemInstruction},
					}, messages...)
				}

				// Prepend to last user message
				lastUserIdx := -1
				for i := len(messages) - 1; i >= 0; i-- {
					if m, ok := messages[i].(map[string]interface{}); ok {
						if role, ok := m["role"].(string); ok && role == "user" {
							lastUserIdx = i
							break
						}
					}
				}

				if lastUserIdx >= 0 {
					if m, ok := messages[lastUserIdx].(map[string]interface{}); ok {
						var userContent string
						switch v := m["content"].(type) {
						case string:
							userContent = v
						default:
							b, _ := json.Marshal(v)
							userContent = string(b)
						}
						m["content"] = "Respond with ONLY raw JSON (no markdown, no backticks, no code blocks): " + userContent
					}
				}
			}
		}
	}

	// Sanitize message content
	// ref: open-sse/executors/github.js:77-101
	sanitizedMessages := make([]interface{}, len(messages))
	for i, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			sanitizedMessages[i] = msg
			continue
		}

		// Copy message
		newMsg := make(map[string]interface{})
		for k, v := range m {
			newMsg[k] = v
		}

		// Check content
		content := newMsg["content"]
		if content == nil {
			sanitizedMessages[i] = newMsg
			continue
		}

		// String content is always fine
		if _, ok := content.(string); ok {
			sanitizedMessages[i] = newMsg
			continue
		}

		// Array content: filter/convert unsupported part types
		if contentArr, ok := content.([]interface{}); ok {
			cleanContent := make([]interface{}, 0)
			for _, part := range contentArr {
				p, ok := part.(map[string]interface{})
				if !ok {
					cleanContent = append(cleanContent, part)
					continue
				}

				partType, _ := p["type"].(string)
				if partType == "text" || partType == "image_url" {
					cleanContent = append(cleanContent, part)
					continue
				}

				// Serialize tool_use, tool_result, thinking, etc. as text
				var text string
				if t, ok := p["text"].(string); ok {
					text = t
				} else if c, ok := p["content"]; ok {
					switch v := c.(type) {
					case string:
						text = v
					default:
						b, _ := json.Marshal(v)
						text = string(b)
					}
				} else {
					b, _ := json.Marshal(p)
					text = string(b)
				}

				if text != "" {
					cleanContent = append(cleanContent, map[string]interface{}{
						"type": "text",
						"text": text,
					})
				}
			}

			if len(cleanContent) > 0 {
				newMsg["content"] = cleanContent
			} else {
				newMsg["content"] = nil
			}
		}

		sanitizedMessages[i] = newMsg
	}

	result["messages"] = sanitizedMessages
	return result
}

// needsRefresh checks if credentials need to be refreshed.
// ref: open-sse/executors/github.js:347-362
func (e *GitHubExecutor) needsRefresh(creds *GitHubCredentials) bool {
	// Always refresh if no copilotToken
	if creds.CopilotToken == "" {
		return true
	}

	if creds.CopilotTokenExpiresAt > 0 {
		// Unix timestamp in seconds
		expiresAtMs := creds.CopilotTokenExpiresAt * 1000
		if time.Now().UnixMilli() > expiresAtMs-5*60*1000 {
			return true
		}
	}

	if creds.CopilotTokenExpiresStr != "" {
		if t, err := time.Parse(time.RFC3339, creds.CopilotTokenExpiresStr); err == nil {
			if time.Now().After(t.Add(-5 * time.Minute)) {
				return true
			}
		}
	}

	return false
}

// extractGitHubCredentials extracts GitHub credentials from context.
func extractGitHubCredentials(ctx context.Context) *GitHubCredentials {
	if v := ctx.Value("githubCredentials"); v != nil {
		if creds, ok := v.(*GitHubCredentials); ok {
			return creds
		}
	}
	return &GitHubCredentials{}
}

// generateRequestID generates a unique request ID.
// ref: open-sse/executors/github.js:33
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), randomString(8))
	}
	return hex.EncodeToString(b)
}

// randomString generates a random string of given length.
func randomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// newBodyReader creates an io.ReadCloser from a byte slice.
func newBodyReader(b []byte) io.ReadCloser {
	return &bodyReader{data: b, offset: 0}
}

// readResponseBody reads the response body.
func readResponseBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()

	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return buf, nil
}
