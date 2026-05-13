// ref: _ref/9router/open-sse/config/errorConfig.js
package config

import (
	"strings"
	"time"
)

type ErrorTypeInfo struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

var ErrorTypes = map[int]ErrorTypeInfo{
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

type BackoffConfig struct {
	Base     time.Duration
	Max      time.Duration
	MaxLevel int
}

var BackoffConfigDefault = BackoffConfig{
	Base:     2000 * time.Millisecond,
	Max:      5 * time.Minute,
	MaxLevel: 15,
}

var TransientCooldownMs = 30 * time.Second
var MaxRateLimitCooldownMs = 30 * time.Minute

type CooldownDurations struct {
	Long  time.Duration
	Short time.Duration
}

var Cooldown = CooldownDurations{
	Long:  2 * time.Minute,
	Short: 5 * time.Second,
}

type ErrorRule struct {
	Text       string        `json:"text,omitempty"`
	Status     int           `json:"status,omitempty"`
	CooldownMs time.Duration `json:"cooldownMs,omitempty"`
	Backoff    bool          `json:"backoff,omitempty"`
}

var ErrorRules = []ErrorRule{
	{Text: "no credentials", CooldownMs: Cooldown.Long},
	{Text: "request not allowed", CooldownMs: Cooldown.Short},
	{Text: "improperly formed request", CooldownMs: Cooldown.Long},
	{Text: "rate limit", Backoff: true},
	{Text: "too many requests", Backoff: true},
	{Text: "quota exceeded", Backoff: true},
	{Text: "capacity", Backoff: true},
	{Text: "overloaded", Backoff: true},
	{Text: "temporarily unavailable", Backoff: true},
	{Text: "insufficient_quota", Backoff: true},
	{Text: "resource_exhausted", Backoff: true},
	{Text: "try again later", Backoff: true},
	{Status: 429, Backoff: true},
	{Status: 503, CooldownMs: Cooldown.Short},
	{Status: 502, CooldownMs: Cooldown.Short},
}

func ClassifyError(statusCode int, message string) ErrorRule {
	lowerMsg := strings.ToLower(message)

	for _, rule := range ErrorRules {
		if rule.Text != "" && strings.Contains(lowerMsg, strings.ToLower(rule.Text)) {
			return rule
		}
	}

	for _, rule := range ErrorRules {
		if rule.Status != 0 && rule.Status == statusCode {
			return rule
		}
	}

	return ErrorRule{}
}

func ShouldUseBackoff(statusCode int, message string) bool {
	rule := ClassifyError(statusCode, message)
	return rule.Backoff
}

func GetErrorCooldown(statusCode int, message string) time.Duration {
	rule := ClassifyError(statusCode, message)
	if rule.Backoff {
		return BackoffConfigDefault.Base
	}
	if rule.CooldownMs > 0 {
		return rule.CooldownMs
	}
	return TransientCooldownMs
}

func GetErrorType(statusCode int) ErrorTypeInfo {
	if et, ok := ErrorTypes[statusCode]; ok {
		return et
	}
	if statusCode >= 500 {
		return ErrorTypeInfo{Type: "server_error", Code: "internal_server_error"}
	}
	return ErrorTypeInfo{Type: "invalid_request_error", Code: ""}
}

func GetDefaultErrorMessage(statusCode int) string {
	if msg, ok := DefaultErrorMessages[statusCode]; ok {
		return msg
	}
	return "An error occurred"
}

func IsRetryableError(statusCode int) bool {
	return statusCode == 429 || statusCode == 502 || statusCode == 503 || statusCode == 504
}

func IsAuthError(statusCode int) bool {
	return statusCode == 401 || statusCode == 403
}

func IsClientError(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}

func IsServerError(statusCode int) bool {
	return statusCode >= 500
}
