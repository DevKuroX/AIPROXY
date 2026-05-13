// ref: _ref/9router/open-sse/executors/default.js
package executor

import (
	"context"
	"io"
	"net/http"
)

// DefaultExecutor is a pass-through executor that makes no modifications
// to requests or responses.
type DefaultExecutor struct {
	BaseExecutor
}

// NewDefaultExecutor creates a new DefaultExecutor for the given provider.
func NewDefaultExecutor(provider string) *DefaultExecutor {
	return &DefaultExecutor{
		BaseExecutor: NewBaseExecutor(provider),
	}
}

// PrepareRequest returns the body unchanged (pass-through).
func (d *DefaultExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	return nil
}

// TransformResponse reads and returns the response body unchanged (pass-through).
func (d *DefaultExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged (pass-through).
func (d *DefaultExecutor) HandleError(ctx context.Context, err error) error {
	return err
}
