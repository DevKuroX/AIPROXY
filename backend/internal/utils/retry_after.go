// ref: _ref/9router/open-sse/config/errorConfig.js (retry parsing)
package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ParseRetryAfter(headerValue string, now time.Time) time.Time {
	if headerValue == "" {
		return time.Time{}
	}

	if seconds, err := strconv.Atoi(strings.TrimSpace(headerValue)); err == nil {
		if seconds < 0 {
			seconds = 0
		}
		return now.Add(time.Duration(seconds) * time.Second)
	}

	if t, err := http.ParseTime(headerValue); err == nil {
		return t
	}

	return time.Time{}
}

func ParseRetryAfterFromHeader(headers http.Header) time.Time {
	retryAfter := headers.Get("Retry-After")
	if retryAfter == "" {
		retryAfter = headers.Get("retry-after")
	}
	return ParseRetryAfter(retryAfter, time.Now())
}

func ParseRetryAfterMs(headers http.Header) int64 {
	retryAfter := headers.Get("Retry-After")
	if retryAfter == "" {
		retryAfter = headers.Get("retry-after")
	}

	if retryAfter == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil {
		return int64(seconds * 1000)
	}

	if t, err := http.ParseTime(retryAfter); err == nil {
		return t.UnixMilli() - time.Now().UnixMilli()
	}

	return 0
}

func FormatRetryAfter(seconds int) string {
	return strconv.Itoa(seconds)
}

func FormatRetryAfterDate(t time.Time) string {
	return t.Format(http.TimeFormat)
}

func CalculateExponentialBackoff(attempt int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return baseDelay
	}

	delay := baseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay > maxDelay {
			return maxDelay
		}
	}

	return delay
}

const (
	BackoffBase    = 2000 * time.Millisecond
	BackoffMax     = 5 * time.Minute
	BackoffMaxLevel = 15
)

func GetBackoffDelay(attempt int) time.Duration {
	return CalculateExponentialBackoff(attempt, BackoffBase, BackoffMax)
}

func ParseResetsAtFromError(bodyText string) int64 {
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(bodyText), &body); err != nil {
		return 0
	}

	if errorObj, ok := body["error"].(map[string]interface{}); ok {
		if resetsAt, ok := errorObj["resets_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, resetsAt); err == nil {
				return t.UnixMilli()
			}
		}
		if resetsAtMs, ok := errorObj["resets_at_ms"].(float64); ok {
			return int64(resetsAtMs)
		}
	}

	if resetsAt, ok := body["resets_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, resetsAt); err == nil {
			return t.UnixMilli()
		}
	}

	if resetsAtMs, ok := body["resets_at_ms"].(float64); ok {
		return int64(resetsAtMs)
	}

	return 0
}

func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
