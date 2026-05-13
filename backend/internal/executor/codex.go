// Package executor provides provider-specific request/response handling.
// ref: _ref/9router/open-sse/executors/codex.js
package executor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Default instructions injected when none provided
// ref: _ref/9router/open-sse/config/codexInstructions.js
const codexDefaultInstructions = `You are Codex, an AI coding assistant. Help the user with programming tasks, code analysis, debugging, and software development. Provide clear, accurate, and practical guidance.`

// Session management for conversation continuity
// ref: _ref/9router/open-sse/executors/codex.js:11-16
const sessionTTL = 60 * time.Minute

type sessionEntry struct {
	sessionID string
	lastUsed  time.Time
}

// CodexExecutor handles OpenAI Codex API (Responses API format).
// ref: _ref/9router/open-sse/executors/codex.js:76
type CodexExecutor struct {
	BaseExecutor
	mu                sync.RWMutex
	sessionMap        map[string]sessionEntry
	currentSessionID  string
	isCompact         bool
	machineID         string
}

// NewCodexExecutor creates a new Codex executor.
func NewCodexExecutor(machineID string) *CodexExecutor {
	return &CodexExecutor{
		BaseExecutor: NewBaseExecutor("codex"),
		sessionMap:   make(map[string]sessionEntry),
		machineID:    machineID,
	}
}

// hashContent creates a short hash of the input text.
// ref: _ref/9router/open-sse/executors/codex.js:18-20
func hashContent(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:])[:16]
}

// generateCodexSessionID creates a new session identifier.
// ref: _ref/9router/open-sse/executors/codex.js:22-24
func generateCodexSessionID() string {
	return fmt.Sprintf("sess_%s_%s",
		strconv.FormatInt(time.Now().Unix(), 36),
		strings.ToLower(strings.TrimPrefix(
			strings.ReplaceAll(fmt.Sprintf("%x", time.Now().UnixNano()), " ", ""),
			"0",
		))[:7])
}

// extractItemText extracts text from an input item.
// ref: _ref/9router/open-sse/executors/codex.js:27-34
func extractItemText(item map[string]interface{}) string {
	if item == nil {
		return ""
	}
	if content, ok := item["content"].(string); ok {
		return content
	}
	if content, ok := item["content"].([]interface{}); ok {
		var texts []string
		for _, c := range content {
			if cm, ok := c.(map[string]interface{}); ok {
				if t, ok := cm["text"].(string); ok && t != "" {
					texts = append(texts, t)
				}
				if o, ok := cm["output"].(string); ok && o != "" {
					texts = append(texts, o)
				}
			}
		}
		return strings.Join(texts, "")
	}
	return ""
}

// resolveConversationSessionId resolves a stable session ID for the conversation.
// ref: _ref/9router/open-sse/executors/codex.js:37-62
func (e *CodexExecutor) resolveConversationSessionId(input []interface{}) string {
	machineSessionID := "default"
	if e.machineID != "" {
		machineSessionID = "sess_" + hashContent(e.machineID)
	}

	if len(input) == 0 {
		return machineSessionID
	}

	// Find first assistant message with text content
	var text string
	for _, item := range input {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if role, ok := itemMap["role"].(string); ok && role == "assistant" {
				text = extractItemText(itemMap)
				if text != "" {
					break
				}
			}
		}
	}

	if text == "" {
		return machineSessionID
	}

	hash := hashContent(e.machineID + text)
	e.mu.RLock()
	entry, exists := e.sessionMap[hash]
	e.mu.RUnlock()

	if exists {
		e.mu.Lock()
		entry.lastUsed = time.Now()
		e.sessionMap[hash] = entry
		e.mu.Unlock()
		return entry.sessionID
	}

	sessionID := generateCodexSessionID()
	e.mu.Lock()
	e.sessionMap[hash] = sessionEntry{
		sessionID: sessionID,
		lastUsed:  time.Now(),
	}
	e.mu.Unlock()

	return sessionID
}

// cleanupExpiredSessions removes old session entries.
// ref: _ref/9router/open-sse/executors/codex.js:65-70
func (e *CodexExecutor) cleanupExpiredSessions() {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	for key, entry := range e.sessionMap {
		if now.Sub(entry.lastUsed) > sessionTTL {
			delete(e.sessionMap, key)
		}
	}
}

// PrepareRequest modifies the outgoing request before sending to Codex.
// ref: _ref/9router/open-sse/executors/codex.js:154-228
func (e *CodexExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Parse the request body
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(body, &bodyMap); err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	// Check for compact mode
	// ref: _ref/9router/open-sse/executors/codex.js:155-156
	if compact, ok := bodyMap["_compact"].(bool); ok {
		e.isCompact = compact
		delete(bodyMap, "_compact")
	} else {
		e.isCompact = false
	}

	// Resolve session ID from conversation history
	// ref: _ref/9router/open-sse/executors/codex.js:158
	var input []interface{}
	if inp, ok := bodyMap["input"].([]interface{}); ok {
		input = inp
	}
	e.currentSessionID = e.resolveConversationSessionId(input)

	// Add session_id header
	// ref: _ref/9router/open-sse/executors/codex.js:88
	if req.Header.Get("session_id") == "" {
		sessionID := e.currentSessionID
		if sessionID == "" {
			sessionID = "default"
		}
		req.Header.Set("session_id", sessionID)
	}

	// Normalize input to array format if needed
	// ref: _ref/9router/open-sse/executors/codex.js:160-161
	if bodyMap["input"] == nil {
		bodyMap["input"] = []interface{}{
			map[string]interface{}{
				"type":    "message",
				"role":    "user",
				"content": []interface{}{map[string]interface{}{"type": "input_text", "text": "..."}},
			},
		}
	}

	// Ensure input is non-empty
	// ref: _ref/9router/open-sse/executors/codex.js:164-166
	if inputArr, ok := bodyMap["input"].([]interface{}); ok && len(inputArr) == 0 {
		bodyMap["input"] = []interface{}{
			map[string]interface{}{
				"type":    "message",
				"role":    "user",
				"content": []interface{}{map[string]interface{}{"type": "input_text", "text": "..."}},
			},
		}
	}

	// Enable streaming (Codex requires it)
	// ref: _ref/9router/open-sse/executors/codex.js:169
	bodyMap["stream"] = true

	// Inject default instructions if missing
	// ref: _ref/9router/open-sse/executors/codex.js:172-174
	if instructions, ok := bodyMap["instructions"].(string); !ok || strings.TrimSpace(instructions) == "" {
		bodyMap["instructions"] = codexDefaultInstructions
	}

	// Ensure store is false (Codex requirement)
	// ref: _ref/9router/open-sse/executors/codex.js:177
	bodyMap["store"] = false

	// Handle model and reasoning effort
	// ref: _ref/9router/open-sse/executors/codex.js:180-202
	modelName := ""
	if m, ok := bodyMap["model"].(string); ok {
		modelName = m
	}

	effortLevels := []string{"none", "low", "medium", "high", "xhigh"}
	var modelEffort string
	for _, level := range effortLevels {
		suffix := "-" + level
		if strings.HasSuffix(modelName, suffix) {
			modelEffort = level
			modelName = strings.TrimSuffix(modelName, suffix)
			break
		}
	}
	bodyMap["model"] = modelName

	// Setup reasoning config
	// ref: _ref/9router/open-sse/executors/codex.js:196-202
	if bodyMap["reasoning"] == nil {
		effort := "low"
		if re, ok := bodyMap["reasoning_effort"].(string); ok && re != "" {
			effort = re
		} else if modelEffort != "" {
			effort = modelEffort
		}
		bodyMap["reasoning"] = map[string]interface{}{
			"effort":  effort,
			"summary": "auto",
		}
	} else if reasoning, ok := bodyMap["reasoning"].(map[string]interface{}); ok {
		if reasoning["summary"] == nil {
			reasoning["summary"] = "auto"
		}
	}
	delete(bodyMap, "reasoning_effort")

	// Include reasoning encrypted content for reasoning models
	// ref: _ref/9router/open-sse/executors/codex.js:205-207
	if reasoning, ok := bodyMap["reasoning"].(map[string]interface{}); ok {
		if effort, ok := reasoning["effort"].(string); ok && effort != "" && effort != "none" {
			bodyMap["include"] = []interface{}{"reasoning.encrypted_content"}
		}
	}

	// Remove unsupported parameters
	// ref: _ref/9router/open-sse/executors/codex.js:209-225
	delete(bodyMap, "temperature")
	delete(bodyMap, "top_p")
	delete(bodyMap, "frequency_penalty")
	delete(bodyMap, "presence_penalty")
	delete(bodyMap, "logprobs")
	delete(bodyMap, "top_logprobs")
	delete(bodyMap, "n")
	delete(bodyMap, "seed")
	delete(bodyMap, "max_tokens")
	delete(bodyMap, "max_completion_tokens")
	delete(bodyMap, "max_output_tokens")
	delete(bodyMap, "user")
	delete(bodyMap, "prompt_cache_retention")
	delete(bodyMap, "metadata")
	delete(bodyMap, "stream_options")
	delete(bodyMap, "safety_identifier")

	// Modify URL for compact mode
	// ref: _ref/9router/open-sse/executors/codex.js:92-95
	if e.isCompact && req.URL != nil {
		req.URL.Path = req.URL.Path + "/compact"
	}

	// Re-serialize body
	newBody, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("failed to serialize request body: %w", err)
	}

	// Update request with new body
	req.Body = &bodyReader{data: newBody}
	req.ContentLength = int64(len(newBody))

	return nil
}

// TransformResponse processes the response from Codex.
// ref: _ref/9router/open-sse/executors/base.js (pass-through for streaming)
func (e *CodexExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	// For streaming responses, pass through as-is
	// Codex uses SSE format compatible with OpenAI
	return nil, nil // nil indicates pass-through
}

// codexError represents an error response from Codex.
type codexError struct {
	Type          string `json:"type"`
	Message       string `json:"message"`
	ResetsAt      int64  `json:"resets_at,omitempty"`
	ResetsInSecs  int64  `json:"resets_in_seconds,omitempty"`
}

// codexErrorResponse wraps the error structure.
type codexErrorResponse struct {
	Error *codexError `json:"error,omitempty"`
}

// HandleError processes errors from Codex requests.
// ref: _ref/9router/open-sse/executors/codex.js:126-148
func (e *CodexExecutor) HandleError(ctx context.Context, err error) error {
	// Return the error as-is for now
	// Rate limit parsing would be done in PrepareRequest by checking response status
	return err
}

// ParseRateLimitError parses Codex rate limit errors to extract reset time.
// ref: _ref/9router/open-sse/executors/codex.js:126-148
func ParseRateLimitError(statusCode int, body []byte) (resetsAtMs int64, parsed bool) {
	if statusCode != http.StatusTooManyRequests {
		return 0, false
	}

	var errResp codexErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return 0, false
	}

	if errResp.Error == nil || errResp.Error.Type != "usage_limit_reached" {
		return 0, false
	}

	now := time.Now().UnixMilli()

	// Check resets_at (Unix timestamp in seconds)
	if errResp.Error.ResetsAt > 0 {
		ms := errResp.Error.ResetsAt * 1000
		if ms > now {
			return ms, true
		}
	}

	// Check resets_in_seconds (relative duration)
	if errResp.Error.ResetsInSecs > 0 {
		return now + errResp.Error.ResetsInSecs*1000, true
	}

	return 0, false
}

// bodyReader implements io.ReadCloser for a byte slice.
type bodyReader struct {
	data   []byte
	offset int
}

func (r *bodyReader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, nil
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *bodyReader) Close() error {
	return nil
}
