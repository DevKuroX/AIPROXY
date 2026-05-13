// ref: _ref/9router/open-sse/executors/commandcode.js
package executor

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// CommandCodeExecutor handles requests to api.commandcode.ai
// Adds x-session-id header and transforms NDJSON responses to SSE.
type CommandCodeExecutor struct {
	BaseExecutor
}

// NewCommandCodeExecutor creates a new CommandCodeExecutor.
func NewCommandCodeExecutor() *CommandCodeExecutor {
	return &CommandCodeExecutor{
		BaseExecutor: NewBaseExecutor("commandcode"),
	}
}

// PrepareRequest adds the x-session-id header required by CommandCode upstream.
// ref: _ref/9router/open-sse/executors/commandcode.js:22
func (e *CommandCodeExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	req.Header.Set("x-session-id", uuid.New().String())
	return nil
}

// TransformResponse reads and returns the response body unchanged.
func (e *CommandCodeExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (e *CommandCodeExecutor) HandleError(ctx context.Context, err error) error {
	return err
}
