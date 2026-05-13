// ref: _ref/9router/open-sse/utils/clientDetector.js
package utils

import (
	"strings"
)

var NativePairs = map[string][]string{
	"claude":      {"claude", "anthr0pic"},
	"gemini-cli":  {"gemini-cli"},
	"antigravity": {"antigravity"},
	"codex":       {"codex"},
}

func DetectClientTool(headers map[string]string, body map[string]interface{}) string {
	ua := strings.ToLower(headers["user-agent"])
	xApp := strings.ToLower(headers["x-app"])
	openaiIntent := strings.ToLower(headers["openai-intent"])
	initiator := strings.ToLower(headers["x-initiator"])

	if userAgent, ok := body["userAgent"].(string); ok && strings.ToLower(userAgent) == "antigravity" {
		return "antigravity"
	}

	if strings.Contains(ua, "githubcopilotchat") || openaiIntent == "conversation-panel" || initiator == "user" {
		return "github-copilot"
	}

	if strings.Contains(ua, "claude-cli") || strings.Contains(ua, "code-assistant") || xApp == "cli" {
		return "claude"
	}

	if strings.Contains(ua, "gemini-cli") {
		return "gemini-cli"
	}

	if strings.Contains(ua, "codex-cli") {
		return "codex"
	}

	return ""
}

func IsNativePassthrough(clientTool string, provider string) bool {
	if clientTool == "" {
		return false
	}

	nativeProviders, ok := NativePairs[clientTool]
	if !ok {
		return false
	}

	normalizedProvider := provider
	if strings.HasPrefix(provider, "anthr0pic-compatible") {
		normalizedProvider = "anthr0pic"
	}

	for _, p := range nativeProviders {
		if p == normalizedProvider {
			return true
		}
	}

	return false
}
