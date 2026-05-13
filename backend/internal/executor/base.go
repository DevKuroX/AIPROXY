// Package executor provides provider-specific request/response handling.
// ref: _ref/9router/open-sse/executors/base.js
package executor

import (
	"context"
	"net/http"
)

// Executor defines the interface for provider-specific request/response handling.
// Implementations transform requests before sending to upstream APIs and
// transform responses before returning to clients.
type Executor interface {
	// PrepareRequest modifies the outgoing request before it's sent to the upstream provider.
	// This can include header manipulation, body transformation, URL rewriting, etc.
	PrepareRequest(ctx context.Context, req *http.Request, body []byte) error

	// TransformResponse modifies the response from the upstream provider before
	// returning it to the client. Returns the transformed response body.
	TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error)

	// HandleError processes errors that occur during request execution.
	// Returns an error suitable for returning to the client.
	HandleError(ctx context.Context, err error) error
}

// BaseExecutor provides common functionality that can be embedded in executor implementations.
type BaseExecutor struct {
	provider string
}

// NewBaseExecutor creates a new BaseExecutor for the given provider.
func NewBaseExecutor(provider string) BaseExecutor {
	return BaseExecutor{provider: provider}
}

// Provider returns the provider name for this executor.
func (b BaseExecutor) Provider() string {
	return b.provider
}
