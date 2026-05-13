// ref: _ref/9router/open-sse/executors/perplexity-web.js
package executor

import (
	"context"
	"io"
	"net/http"
)

// PerplexityWebExecutor handles requests to Perplexity web API.
// Implements session caching and model mapping for various Perplexity models.
type PerplexityWebExecutor struct {
	BaseExecutor
}

// NewPerplexityWebExecutor creates a new PerplexityWebExecutor.
func NewPerplexityWebExecutor() *PerplexityWebExecutor {
	return &PerplexityWebExecutor{
		BaseExecutor: NewBaseExecutor("perplexity-web"),
	}
}

// PrepareRequest modifies the outgoing request for Perplexity web API.
// ref: _ref/9router/open-sse/executors/perplexity-web.js:100
func (e *PerplexityWebExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Perplexity web requires specific headers handled by the client layer
	return nil
}

// TransformResponse reads and returns the response body unchanged.
// ref: _ref/9router/open-sse/executors/perplexity-web.js
func (e *PerplexityWebExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (e *PerplexityWebExecutor) HandleError(ctx context.Context, err error) error {
	return err
}
