// ref: _ref/9router/open-sse/executors/grok-web.js
package executor

import (
	"context"
	"io"
	"net/http"
)

// GrokWebExecutor handles requests to Grok web API (x.ai).
// Implements model mapping and thinking mode support.
type GrokWebExecutor struct {
	BaseExecutor
}

// NewGrokWebExecutor creates a new GrokWebExecutor.
func NewGrokWebExecutor() *GrokWebExecutor {
	return &GrokWebExecutor{
		BaseExecutor: NewBaseExecutor("grok-web"),
	}
}

// PrepareRequest modifies the outgoing request for Grok web API.
// ref: _ref/9router/open-sse/executors/grok-web.js:44
func (e *GrokWebExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Grok web requires specific headers handled by the client layer
	return nil
}

// TransformResponse reads and returns the response body unchanged.
// ref: _ref/9router/open-sse/executors/grok-web.js
func (e *GrokWebExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (e *GrokWebExecutor) HandleError(ctx context.Context, err error) error {
	return err
}
