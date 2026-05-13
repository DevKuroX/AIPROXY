// Package router provides header building utilities for provider-specific authentication.
// ref: _ref/9router/open-sse/executors/default.js (buildHeaders method)
package router

import (
	"net/http"
)

// ProviderCredentials holds authentication data for a provider.
// Different providers use different credential fields.
// ref: _ref/9router/open-sse/executors/default.js:53-120
type ProviderCredentials struct {
	APIKey               string
	AccessToken          string
	ProviderType         string
	ProviderSpecificData map[string]interface{}
}

// BuildProviderHeaders builds HTTP headers for a given provider type and credentials.
// This centralizes header construction logic that was previously scattered across executors.
// ref: _ref/9router/open-sse/executors/default.js:53-120
func BuildProviderHeaders(providerType string, creds *ProviderCredentials, stream bool) http.Header {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	switch providerType {
	case "azure":
		// Azure uses "api-key" header instead of Authorization
		// ref: _ref/9router/open-sse/executors/azure.js:26-52
		if creds != nil && creds.APIKey != "" {
			headers.Set("api-key", creds.APIKey)
		}
		// Organization header if provided
		if creds != nil {
			if org, ok := creds.ProviderSpecificData["organization"].(string); ok && org != "" {
				headers.Set("OpenAI-Organization", org)
			}
		}

	case "claude":
		// CL4ude uses x-api-key or Bearer token
		// ref: _ref/9router/open-sse/executors/default.js:60-94
		if creds != nil {
			if creds.APIKey != "" {
				headers.Set("x-api-key", creds.APIKey)
			} else if creds.AccessToken != "" {
				headers.Set("Authorization", "Bearer "+creds.AccessToken)
			}
		}

	case "gemini":
		// Gemini uses x-goog-api-key or Bearer token
		// ref: _ref/9router/open-sse/executors/default.js:57-59
		if creds != nil {
			if creds.APIKey != "" {
				headers.Set("x-goog-api-key", creds.APIKey)
			} else if creds.AccessToken != "" {
				headers.Set("Authorization", "Bearer "+creds.AccessToken)
			}
		}

	case "gemini-cli":
		// Gemini CLI uses Bearer token with additional headers
		// ref: _ref/9router/open-sse/executors/gemini-cli.js:15-23
		if creds != nil && creds.AccessToken != "" {
			headers.Set("Authorization", "Bearer "+creds.AccessToken)
		}
		headers.Set("X-Goog-Api-Client", "google-genai-sdk/1.41.0 gl-node/v22.19.0")

	case "vertex", "vertex-partner":
		// Vertex uses Bearer token (from service account JWT assertion)
		// ref: _ref/9router/open-sse/executors/vertex.js
		if creds != nil && creds.AccessToken != "" {
			headers.Set("Authorization", "Bearer "+creds.AccessToken)
		}

	case "glm", "kimi", "minimax", "minimax-cn":
		// These providers use x-api-key
		// ref: _ref/9router/open-sse/executors/default.js:95-100
		if creds != nil {
			if creds.APIKey != "" {
				headers.Set("x-api-key", creds.APIKey)
			} else if creds.AccessToken != "" {
				headers.Set("x-api-key", creds.AccessToken)
			}
		}

	case "cursor":
		// Cursor uses complex headers with checksum
		// Full implementation in cursor.go:buildCursorHeaders
		// ref: _ref/9router/open-sse/utils/cursorChecksum.js:95-142
		if creds != nil && creds.AccessToken != "" {
			headers.Set("Authorization", "Bearer "+creds.AccessToken)
			headers.Set("connect-accept-encoding", "gzip")
			headers.Set("connect-protocol-version", "1")
			headers.Set("content-type", "application/connect+proto")
			headers.Set("user-agent", "connect-es/1.6.1")
		}

	case "codex":
		// Codex uses standard Bearer auth
		// ref: _ref/9router/open-sse/executors/codex.js
		if creds != nil && creds.APIKey != "" {
			headers.Set("Authorization", "Bearer "+creds.APIKey)
		}

	case "openai", "openai-compatible-chat", "openai-compatible-responses":
		// Standard OpenAI auth
		// ref: _ref/9router/open-sse/executors/default.js:105-119
		if creds != nil {
			if creds.APIKey != "" {
				headers.Set("Authorization", "Bearer "+creds.APIKey)
			} else if creds.AccessToken != "" {
				headers.Set("Authorization", "Bearer "+creds.AccessToken)
			}
		}

	case "anthr0pic-compatible-chat":
		// Anthropic-compatible endpoints
		// ref: _ref/9router/open-sse/executors/default.js:106-113
		if creds != nil {
			if creds.APIKey != "" {
				headers.Set("x-api-key", creds.APIKey)
			} else if creds.AccessToken != "" {
				headers.Set("Authorization", "Bearer "+creds.AccessToken)
			}
		}

	default:
		// Default to Bearer auth for unknown providers
		if creds != nil {
			if creds.APIKey != "" {
				headers.Set("Authorization", "Bearer "+creds.APIKey)
			} else if creds.AccessToken != "" {
				headers.Set("Authorization", "Bearer "+creds.AccessToken)
			}
		}
	}

	// Set Accept header for streaming
	if stream {
		headers.Set("Accept", "text/event-stream")
	} else {
		headers.Set("Accept", "application/json")
	}

	return headers
}

// BuildAzureHeaders builds Azure-specific headers.
// Azure uses "api-key" header instead of "Authorization: Bearer".
// ref: _ref/9router/open-sse/executors/azure.js:26-52
func BuildAzureHeaders(apiKey, organization string, stream bool) http.Header {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	if apiKey != "" {
		headers.Set("api-key", apiKey)
	}

	if organization != "" {
		headers.Set("OpenAI-Organization", organization)
	}

	if stream {
		headers.Set("Accept", "text/event-stream")
	}

	return headers
}

// BuildClaudeHeaders builds CL4ude-specific headers.
// CL4ude requires x-api-key or Bearer token and anthr0pic-version header.
// ref: _ref/9router/open-sse/executors/default.js:60-94
func BuildClaudeHeaders(apiKey, accessToken, version string, stream bool) http.Header {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	if apiKey != "" {
		headers.Set("x-api-key", apiKey)
	} else if accessToken != "" {
		headers.Set("Authorization", "Bearer "+accessToken)
	}

	if version != "" {
		headers.Set("anthropic-version", version)
	}

	if stream {
		headers.Set("Accept", "text/event-stream")
	}

	return headers
}

// BuildGeminiHeaders builds Gemini API headers.
// Gemini uses x-goog-api-key or Bearer token.
// ref: _ref/9router/open-sse/executors/default.js:57-59
func BuildGeminiHeaders(apiKey, accessToken string, stream bool) http.Header {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	if apiKey != "" {
		headers.Set("x-goog-api-key", apiKey)
	} else if accessToken != "" {
		headers.Set("Authorization", "Bearer "+accessToken)
	}

	if stream {
		headers.Set("Accept", "text/event-stream")
	}

	return headers
}

// BuildVertexHeaders builds Vertex AI headers.
// Vertex uses Bearer token from service account JWT assertion.
// ref: _ref/9router/open-sse/executors/vertex.js
func BuildVertexHeaders(accessToken string, stream bool) http.Header {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	if accessToken != "" {
		headers.Set("Authorization", "Bearer "+accessToken)
	}

	if stream {
		headers.Set("Accept", "text/event-stream")
	}

	return headers
}

// BuildOpenAIHeaders builds standard OpenAI-compatible headers.
// Uses Authorization: Bearer for API key.
// ref: _ref/9router/open-sse/executors/default.js:105-119
func BuildOpenAIHeaders(apiKey string, stream bool) http.Header {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	if apiKey != "" {
		headers.Set("Authorization", "Bearer "+apiKey)
	}

	if stream {
		headers.Set("Accept", "text/event-stream")
	}

	return headers
}
