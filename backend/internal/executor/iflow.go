// ref: _ref/9router/open-sse/executors/iflow.js
package executor

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// IFlowExecutor handles requests to iFlow API with HMAC-SHA256 signature.
// ref: _ref/9router/open-sse/executors/iflow.js:8
type IFlowExecutor struct {
	BaseExecutor
}

// NewIFlowExecutor creates a new IFlowExecutor.
func NewIFlowExecutor() *IFlowExecutor {
	return &IFlowExecutor{
		BaseExecutor: NewBaseExecutor("iflow"),
	}
}

// createSignature creates an HMAC-SHA256 signature for iFlow API.
// Formula: HMAC-SHA256(key=apiKey, message="UserAgent:sessionID:timestamp")
// ref: _ref/9router/open-sse/executors/iflow.js:29
func createIFlowSignature(userAgent, sessionID string, timestamp int64, apiKey string) string {
	if apiKey == "" {
		return ""
	}
	payload := fmt.Sprintf("%s:%s:%d", userAgent, sessionID, timestamp)
	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// PrepareRequest adds iFlow-specific headers including signature.
// ref: _ref/9router/open-sse/executors/iflow.js:43
func (e *IFlowExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	sessionID := "session-" + uuid.New().String()
	timestamp := time.Now().UnixMilli()
	userAgent := "iFlow-Cli"

	// Extract API key from Authorization header if present
	apiKey := ""
	if auth := req.Header.Get("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
		apiKey = auth[7:]
	}

	signature := createIFlowSignature(userAgent, sessionID, timestamp, apiKey)

	req.Header.Set("session-id", sessionID)
	req.Header.Set("x-iflow-timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("x-iflow-signature", signature)

	return nil
}

// TransformResponse reads and returns the response body unchanged.
func (e *IFlowExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (e *IFlowExecutor) HandleError(ctx context.Context, err error) error {
	return err
}
