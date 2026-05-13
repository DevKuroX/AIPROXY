package stream

import (
	"encoding/json"
	"fmt"
)

// StreamError represents an error in SSE format
// ref: open-sse/utils/error.js - createErrorResult
type StreamError struct {
	ErrorInfo struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

// NewStreamError creates a new stream error
// ref: open-sse/utils/error.js - createErrorResult
func NewStreamError(message, errType, code string) *StreamError {
	return &StreamError{
		ErrorInfo: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code,omitempty"`
		}{
			Message: message,
			Type:    errType,
			Code:    code,
		},
	}
}

// ToSSE formats the error as an SSE chunk
// ref: open-sse/utils/error.js - createErrorResult
func (e *StreamError) ToSSE() []byte {
	data, _ := json.Marshal(e)
	return []byte(fmt.Sprintf("event: error\ndata: %s\n\n", string(data)))
}

// ToJSON formats the error as a JSON response
// ref: open-sse/utils/error.js - createErrorResult
func (e *StreamError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// Error implements the error interface
func (e *StreamError) Error() string {
	return e.ErrorInfo.Message
}

// IsStreamError checks if an error is a stream error
func IsStreamError(err error) bool {
	_, ok := err.(*StreamError)
	return ok
}