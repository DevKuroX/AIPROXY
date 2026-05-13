// ref: _ref/9router/open-sse/utils/error.js
package utils

import (
	"encoding/json"
	"fmt"
	"io"
)

type ErrorType struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

var ErrorTypes = map[int]ErrorType{
	400: {Type: "invalid_request_error", Code: "bad_request"},
	401: {Type: "authentication_error", Code: "invalid_api_key"},
	402: {Type: "billing_error", Code: "payment_required"},
	403: {Type: "permission_error", Code: "insufficient_quota"},
	404: {Type: "invalid_request_error", Code: "model_not_found"},
	406: {Type: "invalid_request_error", Code: "model_not_supported"},
	429: {Type: "rate_limit_error", Code: "rate_limit_exceeded"},
	500: {Type: "server_error", Code: "internal_server_error"},
	502: {Type: "server_error", Code: "bad_gateway"},
	503: {Type: "server_error", Code: "service_unavailable"},
	504: {Type: "server_error", Code: "gateway_timeout"},
}

var DefaultErrorMessages = map[int]string{
	400: "Bad request",
	401: "Invalid API key provided",
	402: "Payment required",
	403: "You exceeded your current quota",
	404: "Model not found",
	406: "Model not supported",
	429: "Rate limit exceeded",
	500: "Internal server error",
	502: "Bad gateway - upstream provider error",
	503: "Service temporarily unavailable",
	504: "Gateway timeout",
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func BuildErrorBody(statusCode int, message string) ErrorResponse {
	errorInfo, ok := ErrorTypes[statusCode]
	if !ok {
		if statusCode >= 500 {
			errorInfo = ErrorType{Type: "server_error", Code: "internal_server_error"}
		} else {
			errorInfo = ErrorType{Type: "invalid_request_error", Code: ""}
		}
	}

	msg := message
	if msg == "" {
		msg, ok = DefaultErrorMessages[statusCode]
		if !ok {
			msg = "An error occurred"
		}
	}

	return ErrorResponse{
		Error: ErrorDetail{
			Message: msg,
			Type:    errorInfo.Type,
			Code:    errorInfo.Code,
		},
	}
}

func ErrorJSON(statusCode int, message string) []byte {
	body := BuildErrorBody(statusCode, message)
	data, _ := json.Marshal(body)
	return data
}

type ParsedUpstreamError struct {
	StatusCode  int    `json:"status_code"`
	Message     string `json:"message"`
	ResetsAtMs  int64  `json:"resets_at_ms,omitempty"`
}

func ParseUpstreamError(statusCode int, bodyText string) ParsedUpstreamError {
	msg, ok := DefaultErrorMessages[statusCode]
	if !ok {
		msg = fmt.Sprintf("Upstream error: %d", statusCode)
	}

	if bodyText != "" {
		var body map[string]interface{}
		if err := json.Unmarshal([]byte(bodyText), &body); err == nil {
			if errObj, ok := body["error"].(map[string]interface{}); ok {
				if m, ok := errObj["message"].(string); ok && m != "" {
					msg = m
				}
			}
			if errMsg, ok := body["error_message"].(string); ok && errMsg != "" {
				msg = errMsg
			}
			if detail, ok := body["detail"].(string); ok && detail != "" {
				msg = detail
			}
		}
	}

	return ParsedUpstreamError{
		StatusCode: statusCode,
		Message:    msg,
	}
}

func WriteStreamError(writer io.Writer, statusCode int, message string) error {
	errorBody := BuildErrorBody(statusCode, message)
	data, err := json.Marshal(errorBody)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "data: %s\n\n", string(data))
	return err
}
