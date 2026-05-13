// ref: _ref/9router/open-sse/executors/qoder.js
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

// QoderExecutor handles requests to Qoder API with HMAC-SHA256 signature.
// Requires 3 custom headers: session-id, x-qoder-timestamp, x-qoder-signature.
// ref: _ref/9router/open-sse/executors/qoder.js:9
type QoderExecutor struct {
	BaseExecutor
}

// NewQoderExecutor creates a new QoderExecutor.
func NewQoderExecutor() *QoderExecutor {
	return &QoderExecutor{
		BaseExecutor: NewBaseExecutor("qoder"),
	}
}

// createSignature creates an HMAC-SHA256 signature for Qoder API.
// Formula: HMAC-SHA256(key=apiKey, message="UserAgent:sessionID:timestamp")
// ref: _ref/9router/open-sse/executors/qoder.js:18
func createQoderSignature(userAgent, sessionID string, timestamp int64, apiKey string) string {
	if apiKey == "" {
		return ""
	}
	payload := fmt.Sprintf("%s:%s:%d", userAgent, sessionID, timestamp)
	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// PrepareRequest adds Qoder-specific headers including signature.
// ref: _ref/9router/open-sse/executors/qoder.js:29
func (e *QoderExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	sessionID := "session-" + uuid.New().String()
	timestamp := time.Now().UnixMilli()
	userAgent := "Qoder-Cli"

	// Extract API key from Authorization header if present
	apiKey := ""
	if auth := req.Header.Get("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
		apiKey = auth[7:]
	}

	signature := createQoderSignature(userAgent, sessionID, timestamp, apiKey)

	req.Header.Set("session-id", sessionID)
	req.Header.Set("x-qoder-timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("x-qoder-signature", signature)

	return nil
}

// TransformResponse reads and returns the response body unchanged.
func (e *QoderExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (e *QoderExecutor) HandleError(ctx context.Context, err error) error {
	return err
}
