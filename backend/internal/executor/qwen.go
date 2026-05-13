// Package executor provides provider-specific request/response handling.
// QwenExecutor handles Qwen API with specific headers and request transformations.
// ref: _ref/9router/open-sse/executors/qwen.js
package executor

import (
	"context"
	"encoding/json"
	"net/http"
)

// Qwen-specific constants.
// ref: open-sse/executors/qwen.js:5-19
const (
	QwenUserAgent = "QwenCode/0.12.3 (linux; x64)"
)

// QwenStainless holds the stainless headers for Qwen.
// ref: open-sse/executors/qwen.js:7-15
var QwenStainless = map[string]string{
	"os":              "Linux",
	"arch":            "x64",
	"lang":            "js",
	"runtime":         "node",
	"runtimeVersion":  "v18.19.1",
	"packageVersion":  "5.11.0",
	"retryCount":      "1",
}

// QwenDefaultSystemMessage is the system message prepended to Qwen requests.
// ref: open-sse/executors/qwen.js:16-19
var QwenDefaultSystemMessage = map[string]interface{}{
	"role": "system",
	"content": []map[string]interface{}{
		{
			"type": "text",
			"text": "",
			"remove": map[string]interface{}{
				"type": "ephemeral",
			},
		},
	},
}

// QwenExecutor implements the Executor interface for Qwen API.
// ref: open-sse/executors/qwen.js:71
type QwenExecutor struct {
	DefaultExecutor
}

// NewQwenExecutor creates a new Qwen executor.
// ref: open-sse/executors/qwen.js:72-74
func NewQwenExecutor() *QwenExecutor {
	return &QwenExecutor{
		DefaultExecutor: *NewDefaultExecutor("qwen"),
	}
}

// QwenCredentials holds Qwen-specific authentication data.
// ref: open-sse/executors/qwen.js:79
type QwenCredentials struct {
	AccessToken          string
	APIKey               string
	RefreshToken         string
	ProviderSpecificData map[string]interface{}
}

// PrepareRequest modifies the outgoing request with Qwen-specific headers.
// ref: open-sse/executors/qwen.js:47-69
func (q *QwenExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Transform request body if needed
	// ref: open-sse/executors/qwen.js:88-95
	if len(body) > 0 {
		var reqBody map[string]interface{}
		if err := json.Unmarshal(body, &reqBody); err == nil {
			// Apply transformations
			reqBody = q.transformRequest(reqBody)
			transformed, err := json.Marshal(reqBody)
			if err == nil {
				body = transformed
			}
		}
	}

	return nil
}

// transformRequest applies Qwen-specific request transformations.
// ref: open-sse/executors/qwen.js:88-95
func (q *QwenExecutor) transformRequest(body map[string]interface{}) map[string]interface{} {
	// Add stream_options if streaming without thinking enabled
	// ref: open-sse/executors/qwen.js:90-92
	if stream, ok := body["stream"].(bool); ok && stream {
		_, hasStreamOptions := body["stream_options"]
		_, hasThinking := body["thinking"]
		_, hasEnableThinking := body["enable_thinking"]
		
		if !hasStreamOptions && !hasThinking && !hasEnableThinking {
			body["stream_options"] = map[string]interface{}{
				"include_usage": true,
			}
		}
	}

	// Sanitize tool_choice when thinking is active
	// ref: open-sse/executors/qwen.js:39-45
	if q.isThinkingActive(body) {
		if tc, ok := body["tool_choice"]; ok {
			incompatible := tc == "required"
			if m, ok := tc.(map[string]interface{}); ok && len(m) > 0 {
				incompatible = true
			}
			if incompatible {
				body["tool_choice"] = "auto"
			}
		}
	}

	// Ensure system message is present
	// ref: open-sse/executors/qwen.js:21-29
	return q.ensureSystemMessage(body)
}

// isThinkingActive checks if thinking mode is enabled.
// ref: open-sse/executors/qwen.js:32-36
func (q *QwenExecutor) isThinkingActive(body map[string]interface{}) bool {
	if thinking, ok := body["thinking"]; ok {
		if thinking == true {
			return true
		}
		if enableThinking, ok := body["enable_thinking"].(bool); ok && enableThinking {
			return true
		}
		if m, ok := thinking.(map[string]interface{}); ok {
			if t, ok := m["type"].(string); ok && t == "enabled" {
				return true
			}
		}
	}
	if enableThinking, ok := body["enable_thinking"].(bool); ok && enableThinking {
		return true
	}
	return false
}

// ensureSystemMessage prepends the default system message to the messages array.
// ref: open-sse/executors/qwen.js:21-29
func (q *QwenExecutor) ensureSystemMessage(body map[string]interface{}) map[string]interface{} {
	messages, ok := body["messages"].([]interface{})
	if !ok {
		// No messages array, create one with just the system message
		body["messages"] = []interface{}{QwenDefaultSystemMessage}
		return body
	}

	// Prepend the system message
	newMessages := make([]interface{}, 0, len(messages)+1)
	newMessages = append(newMessages, QwenDefaultSystemMessage)
	newMessages = append(newMessages, messages...)
	body["messages"] = newMessages

	return body
}

// BuildHeaders creates Qwen-specific upstream headers.
// ref: open-sse/executors/qwen.js:47-69
func (q *QwenExecutor) BuildHeaders(credentials *QwenCredentials, stream bool) http.Header {
	headers := make(http.Header)

	// Determine token
	token := ""
	if credentials != nil {
		if credentials.APIKey != "" {
			token = credentials.APIKey
		} else if credentials.AccessToken != "" {
			token = credentials.AccessToken
		}
	}

	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", "Bearer "+token)
	headers.Set("User-Agent", QwenUserAgent)
	headers.Set("X-DashScope-AuthType", "qwen-oauth")
	headers.Set("X-DashScope-CacheControl", "enable")
	headers.Set("X-DashScope-UserAgent", QwenUserAgent)
	headers.Set("X-Stainless-Arch", QwenStainless["arch"])
	headers.Set("X-Stainless-Lang", QwenStainless["lang"])
	headers.Set("X-Stainless-Os", QwenStainless["os"])
	headers.Set("X-Stainless-Package-Version", QwenStainless["packageVersion"])
	headers.Set("X-Stainless-Retry-Count", QwenStainless["retryCount"])
	headers.Set("X-Stainless-Runtime", QwenStainless["runtime"])
	headers.Set("X-Stainless-Runtime-Version", QwenStainless["runtimeVersion"])
	headers.Set("Connection", "keep-alive")
	headers.Set("Accept-Language", "*")
	headers.Set("Sec-Fetch-Mode", "cors")

	if stream {
		headers.Set("Accept", "text/event-stream")
	} else {
		headers.Set("Accept", "application/json")
	}

	return headers
}

// BuildURL constructs the Qwen API URL.
// Qwen tokens are bound to a resource_url returned at OAuth time.
// Using portal.qwen.ai when the token is issued for another shard returns 401/403.
// ref: open-sse/executors/qwen.js:76-82
func (q *QwenExecutor) BuildURL(credentials *QwenCredentials) string {
	host := "portal.qwen.ai"
	
	if credentials != nil && credentials.ProviderSpecificData != nil {
		if resourceURL, ok := credentials.ProviderSpecificData["resourceUrl"].(string); ok && resourceURL != "" {
			// Remove protocol and trailing slash
			host = resourceURL
			// Strip https:// prefix
			if len(host) > 8 && host[:8] == "https://" {
				host = host[8:]
			} else if len(host) > 7 && host[:7] == "http://" {
				host = host[7:]
			}
			// Strip trailing slash
			if len(host) > 0 && host[len(host)-1] == '/' {
				host = host[:len(host)-1]
			}
		}
	}

	return "https://" + host + "/v1/chat/completions"
}

// TransformResponse reads and returns the response body unchanged (pass-through).
// Qwen uses standard SSE format.
func (q *QwenExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return q.DefaultExecutor.TransformResponse(ctx, resp)
}

// HandleError returns the error unchanged (pass-through).
func (q *QwenExecutor) HandleError(ctx context.Context, err error) error {
	return err
}
