// Package executor provides provider-specific request/response handling.
// OpenCodeExecutor handles OpenCode API with specific headers and URL routing.
// ref: _ref/9router/open-sse/executors/opencode.js
package executor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// OpenCode API constants.
// ref: open-sse/executors/opencode.js:5-6
var OpenCodeMessagesModels = map[string]bool{
	"big-pickle": true,
}

// OpenCodeExecutor implements the Executor interface for OpenCode API.
// ref: open-sse/executors/opencode.js:7
type OpenCodeExecutor struct {
	BaseExecutor
}

// NewOpenCodeExecutor creates a new OpenCode executor.
// ref: open-sse/executors/opencode.js:8-10
func NewOpenCodeExecutor() *OpenCodeExecutor {
	return &OpenCodeExecutor{
		BaseExecutor: NewBaseExecutor("opencode"),
	}
}

// PrepareRequest modifies the outgoing request with OpenCode-specific headers.
// ref: open-sse/executors/opencode.js:19-26
func (o *OpenCodeExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Set OpenCode-specific headers
	// ref: open-sse/executors/opencode.js:19-26
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer public")
	req.Header.Set("x-opencode-client", "desktop")
	req.Header.Set("Accept", "text/event-stream")

	// Set URL based on model
	// ref: open-sse/executors/opencode.js:12-17
	model := getModelFromBody(body)
	if OpenCodeMessagesModels[model] {
		req.URL.Path = "/zen/v1/messages"
	} else {
		req.URL.Path = "/zen/v1/chat/completions"
	}
	req.URL.Host = "opencode.ai"
	req.URL.Scheme = "https"

	return nil
}

// TransformResponse reads and returns the response body unchanged.
func (o *OpenCodeExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (o *OpenCodeExecutor) HandleError(ctx context.Context, err error) error {
	return err
}

// getModelFromBody extracts the model from the request body.
func getModelFromBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var reqBody struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		return ""
	}
	return reqBody.Model
}
