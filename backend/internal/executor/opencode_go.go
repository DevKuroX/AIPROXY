// Package executor provides provider-specific request/response handling.
// OpenCodeGoExecutor handles OpenCode Go API with claude-format routing and auth.
// ref: _ref/9router/open-sse/executors/opencode-go.js
package executor

import (
	"context"
	"io"
	"net/http"
)

// OpenCode Go API constants.
// ref: open-sse/executors/opencode-go.js:6-8
var OpenCodeGoClaudeFormatModels = map[string]bool{
	"minimax-m2.5": true,
	"minimax-m2.7": true,
}

const openCodeGoBaseURL = "https://opencode.ai/zen/go/v1"

// OpenCodeGoExecutor implements the Executor interface for OpenCode Go API.
// ref: open-sse/executors/opencode-go.js:10
type OpenCodeGoExecutor struct {
	BaseExecutor
	lastModel string
}

// NewOpenCodeGoExecutor creates a new OpenCode Go executor.
// ref: open-sse/executors/opencode-go.js:11-13
func NewOpenCodeGoExecutor() *OpenCodeGoExecutor {
	return &OpenCodeGoExecutor{
		BaseExecutor: NewBaseExecutor("opencode-go"),
	}
}

// PrepareRequest modifies the outgoing request with OpenCode Go-specific headers.
// ref: open-sse/executors/opencode-go.js:16-36
func (o *OpenCodeGoExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	model := getModelFromBody(body)
	o.lastModel = model

	req.Header.Set("Content-Type", "application/json")

	if OpenCodeGoClaudeFormatModels[model] {
		req.Header.Set("x-api-key", "placeholder-api-key")
		req.Header.Set("anthr0pic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", "Bearer placeholder-api-key")
	}

	req.Header.Set("Accept", "text/event-stream")

	if OpenCodeGoClaudeFormatModels[model] {
		req.URL.Path = "/zen/go/v1/messages"
	} else {
		req.URL.Path = "/zen/go/v1/chat/completions"
	}
	req.URL.Host = "opencode.ai"
	req.URL.Scheme = "https"

	return nil
}

// TransformResponse reads and returns the response body unchanged.
func (o *OpenCodeGoExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (o *OpenCodeGoExecutor) HandleError(ctx context.Context, err error) error {
	return err
}

// injectReasoningContent injects reasoning_content placeholder for providers that require it.
// ref: open-sse/utils/reasoningContentInjector.js
func injectReasoningContent(provider, model string, body map[string]interface{}) map[string]interface{} {
	shouldInject := false
	scope := ""

	if provider == "deepseek" {
		shouldInject = true
		scope = "all"
	}

	if len(model) >= 5 && model[:5] == "kimi-" {
		shouldInject = true
		scope = "toolCalls"
	}

	if len(model) >= 9 && model[:9] == "deepseek-" {
		shouldInject = true
		scope = "all"
	}

	if !shouldInject {
		return body
	}

	messages, ok := body["messages"].([]interface{})
	if !ok {
		return body
	}

	newMessages := make([]interface{}, len(messages))
	for i, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			newMessages[i] = msg
			continue
		}

		role, _ := msgMap["role"].(string)
		if role != "assistant" {
			newMessages[i] = msg
			continue
		}

		rc, hasRC := msgMap["reasoning_content"].(string)
		if hasRC && len(rc) > 0 {
			newMessages[i] = msg
			continue
		}

		if scope == "toolCalls" {
			toolCalls, hasToolCalls := msgMap["tool_calls"].([]interface{})
			if !hasToolCalls || len(toolCalls) == 0 {
				newMessages[i] = msg
				continue
			}
		}

		newMsg := make(map[string]interface{})
		for k, v := range msgMap {
			newMsg[k] = v
		}
		newMsg["reasoning_content"] = " "
		newMessages[i] = newMsg
	}

	newBody := make(map[string]interface{})
	for k, v := range body {
		newBody[k] = v
	}
	newBody["messages"] = newMessages
	return newBody
}

// transformRequest applies reasoning content injection.
// ref: open-sse/executors/opencode-go.js:38-40
func (o *OpenCodeGoExecutor) transformRequest(body map[string]interface{}) map[string]interface{} {
	return injectReasoningContent("opencode-go", o.lastModel, body)
}
