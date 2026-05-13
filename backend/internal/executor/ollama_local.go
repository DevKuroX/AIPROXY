// Package executor provides provider-specific request/response handling.
// OllamaLocalExecutor handles local Ollama API with host resolution.
// ref: _ref/9router/open-sse/executors/ollama-local.js
package executor

import (
	"context"
	"io"
	"net/http"
	"os"
)

// OllamaLocalExecutor implements the Executor interface for local Ollama API.
// ref: open-sse/executors/ollama-local.js:4
type OllamaLocalExecutor struct {
	DefaultExecutor
}

// NewOllamaLocalExecutor creates a new Ollama local executor.
// ref: open-sse/executors/ollama-local.js:5-7
func NewOllamaLocalExecutor() *OllamaLocalExecutor {
	return &OllamaLocalExecutor{
		DefaultExecutor: *NewDefaultExecutor("ollama-local"),
	}
}

// PrepareRequest modifies the outgoing request for local Ollama.
// ref: open-sse/executors/ollama-local.js:9-11
func (o *OllamaLocalExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	host := resolveOllamaLocalHost()
	req.URL.Scheme = "http"
	req.URL.Host = host
	req.URL.Path = "/api/chat"
	return nil
}

// resolveOllamaLocalHost returns the Ollama local host address.
// ref: open-sse/config/providers.js (resolveOllamaLocalHost)
func resolveOllamaLocalHost() string {
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		return host
	}
	return "localhost:11434"
}

// TransformResponse reads and returns the response body unchanged.
func (o *OllamaLocalExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (o *OllamaLocalExecutor) HandleError(ctx context.Context, err error) error {
	return err
}
