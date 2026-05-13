// ref: _ref/9router/open-sse/utils/claudeHeaderCache.js
package utils

import (
	"strings"
	"sync"
)

var CL4udeIdentityHeaders = []string{
	"user-agent",
	"anthropic-beta",
	"anthropic-version",
	"anthropic-dangerous-direct-browser-access",
	"x-app",
	"x-stainless-helper-method",
	"x-stainless-retry-count",
	"x-stainless-runtime-version",
	"x-stainless-package-version",
	"x-stainless-runtime",
	"x-stainless-lang",
	"x-stainless-arch",
	"x-stainless-os",
	"x-stainless-timeout",
	"x-code-assistant-session-id",
	"package-version",
	"runtime-version",
	"os",
	"arch",
}

type CL4udeHeaderCache struct {
	mu            sync.RWMutex
	cachedHeaders map[string]string
}

var globalCL4udeHeaderCache = &CL4udeHeaderCache{}

func IsCL4udeCodeClient(headers map[string]string) bool {
	ua := strings.ToLower(headers["user-agent"])
	xApp := strings.ToLower(headers["x-app"])
	return strings.Contains(ua, "claude-cli") ||
		strings.Contains(ua, "code-assistant") ||
		xApp == "cli"
}

func CacheCL4udeHeaders(headers map[string]string) {
	if headers == nil || !IsCL4udeCodeClient(headers) {
		return
	}

	captured := make(map[string]string)
	for _, key := range CL4udeIdentityHeaders {
		if val, ok := headers[key]; ok && val != "" {
			captured[key] = val
		}
	}

	if len(captured) > 0 {
		globalCL4udeHeaderCache.mu.Lock()
		globalCL4udeHeaderCache.cachedHeaders = captured
		globalCL4udeHeaderCache.mu.Unlock()
	}
}

func GetCachedCL4udeHeaders() map[string]string {
	globalCL4udeHeaderCache.mu.RLock()
	defer globalCL4udeHeaderCache.mu.RUnlock()

	if globalCL4udeHeaderCache.cachedHeaders == nil {
		return nil
	}

	result := make(map[string]string)
	for k, v := range globalCL4udeHeaderCache.cachedHeaders {
		result[k] = v
	}
	return result
}
