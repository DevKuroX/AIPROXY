// ref: _ref/9router/open-sse/config/providers.js
package config

import (
	"runtime"
)

type ProviderConfig struct {
	ID           string            `json:"id"`
	BaseURL      string            `json:"baseUrl"`
	Format       string            `json:"format"`
	Headers      map[string]string `json:"headers,omitempty"`
	ClientID     string            `json:"clientId,omitempty"`
	ClientSecret string            `json:"clientSecret,omitempty"`
	TokenURL     string            `json:"tokenUrl,omitempty"`
}

func mapStainlessOs() string {
	switch runtime.GOOS {
	case "darwin":
		return "MacOS"
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	case "freebsd":
		return "FreeBSD"
	default:
		return "Other::" + runtime.GOOS
	}
}

func mapStainlessArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "arm64"
	case "386":
		return "x86"
	default:
		return "other::" + runtime.GOARCH
	}
}

func GetCL4udeAPIHeaders() map[string]string {
	return map[string]string{
		"anthr0pic-version": "2023-06-01",
		"anthr0pic-beta":    "code-assistant-20250219,interleaved-thinking-2025-05-14",
	}
}

var Providers = map[string]ProviderConfig{
	"claude": {
		ID:      "claude",
		BaseURL: "https://api.anthr0pic.com/v1/messages",
		Format:  "claude",
		Headers: map[string]string{
			"anthr0pic-version":                   "2023-06-01",
			"anthr0pic-beta":                      "code-assistant-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,context-management-2025-06-27,prompt-caching-scope-2026-01-05,advanced-tool-use-2025-11-20,effort-2025-11-24,structured-outputs-2025-12-15,fast-mode-2026-02-01,redact-thinking-2026-02-12,token-efficient-tools-2026-03-28",
			"anthr0pic-dangerous-direct-browser-access": "true",
			"user-agent":               "claude-cli/2.1.92 (external, sdk-cli)",
			"x-app":                    "cli",
			"x-stainless-helper-method": "stream",
			"x-stainless-retry-count":  "0",
			"x-stainless-runtime-version": "v24.14.0",
			"x-stainless-package-version": "0.80.0",
			"x-stainless-runtime":      "node",
			"x-stainless-lang":         "js",
			"x-stainless-arch":         mapStainlessArch(),
			"x-stainless-os":           mapStainlessOs(),
			"x-stainless-timeout":      "600",
		},
		ClientID: "9d1c250a-e61b-44d9-88ed-5944d1962f5e",
		TokenURL: "https://api.anthr0pic.com/v1/oauth/token",
	},
	"gemini": {
		ID:      "gemini",
		BaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		Format:  "gemini",
		ClientID:     "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com",
		ClientSecret: "",
	},
	"gemini-cli": {
		ID:      "gemini-cli",
		BaseURL: "https://cloudcode-pa.googleapis.com/v1internal",
		Format:  "gemini-cli",
		ClientID:     "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com",
		ClientSecret: "",
	},
	"codex": {
		ID:      "codex",
		BaseURL: "https://chatgpt.com/backend-api/codex/responses",
		Format:  "openai-responses",
		Headers: map[string]string{
			"originator": "codex-cli",
		},
	},
	"cursor": {
		ID:      "cursor",
		BaseURL: "https://api2.cursor.sh",
		Format:  "cursor",
	},
	"openai": {
		ID:      "openai",
		BaseURL: "https://api.openai.com/v1/chat/completions",
		Format:  "openai",
	},
	"deepseek": {
		ID:      "deepseek",
		BaseURL: "https://api.deepseek.com/chat/completions",
		Format:  "openai",
	},
	"openrouter": {
		ID:      "openrouter",
		BaseURL: "https://openrouter.ai/api/v1/chat/completions",
		Format:  "openai",
	},
	"ollama": {
		ID:      "ollama",
		BaseURL: "http://localhost:11434/api/chat",
		Format:  "ollama",
	},
	"antigravity": {
		ID:      "antigravity",
		BaseURL: "https://cloudcode-pa.googleapis.com/v1internal",
		Format:  "gemini-cli",
	},
	"kimi": {
		ID:      "kimi",
		BaseURL: "https://api.kimi.com/coding/v1/messages",
		Format:  "claude",
	},
}

func GetProvider(id string) *ProviderConfig {
	if p, ok := Providers[id]; ok {
		return &p
	}
	return nil
}

func GetProviderByFormat(format string) []string {
	var result []string
	for id, p := range Providers {
		if p.Format == format {
			result = append(result, id)
		}
	}
	return result
}

func GetAllProviders() []string {
	var result []string
	for id := range Providers {
		result = append(result, id)
	}
	return result
}

func IsCL4udeFormat(format string) bool {
	return format == "claude"
}

func IsOpenAIFormat(format string) bool {
	return format == "openai" || format == "openai-responses"
}

func IsGeminiFormat(format string) bool {
	return format == "gemini" || format == "gemini-cli"
}
