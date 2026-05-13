// Package validation provides comprehensive validation and parity testing for ai_proxy
// against the 9router reference implementation.
package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Feature represents a single feature to validate for parity.
type Feature struct {
	Name        string `json:"name"`         // Feature name
	Category    string `json:"category"`     // Category: api, provider, rtk, etc.
	Description string `json:"description"`  // What this feature does
	NineRouter  bool   `json:"nine_router"`  // Exists in 9router
	AIProxy     bool   `json:"ai_proxy"`     // Exists in ai_proxy
	Tested      bool   `json:"tested"`       // Validation test exists
	Passing     bool   `json:"passing"`      // Test passes
	Notes       string `json:"notes"`        // Additional notes
}

// FeatureInventory tracks all features for validation.
type FeatureInventory struct {
	Features []Feature `json:"features"`
	Summary  Summary   `json:"summary"`
}

// Summary provides aggregate statistics.
type Summary struct {
	TotalFeatures      int `json:"total_features"`
	NineRouterFeatures int `json:"nine_router_features"`
	AIProxyFeatures    int `json:"ai_proxy_features"`
	TestedFeatures     int `json:"tested_features"`
	PassingFeatures    int `json:"passing_features"`
	ParityPercentage   int `json:"parity_percentage"`
}

// GenerateFeatureInventory creates a complete feature inventory comparing 9router and ai_proxy.
func GenerateFeatureInventory() (*FeatureInventory, error) {
	features := []Feature{}

	// API Endpoints
	features = append(features, getAPIEndpointFeatures()...)

	// Providers
	features = append(features, getProviderFeatures()...)

	// RTK Filters
	features = append(features, getRTKFeatures()...)

	// Translation
	features = append(features, getTranslationFeatures()...)

	// OAuth & Authentication
	features = append(features, getOAuthFeatures()...)

	// Usage & Analytics
	features = append(features, getUsageFeatures()...)

	// Fallback
	features = append(features, getFallbackFeatures()...)

	inventory := &FeatureInventory{
		Features: features,
		Summary:  calculateSummary(features),
	}

	return inventory, nil
}

// getAPIEndpointFeatures returns all API endpoint features.
func getAPIEndpointFeatures() []Feature {
	return []Feature{
		// Chat Completions
		{Name: "POST /v1/chat/completions", Category: "api", Description: "Chat completions endpoint", NineRouter: true, AIProxy: true},
		{Name: "Chat streaming support", Category: "api", Description: "SSE streaming for chat completions", NineRouter: true, AIProxy: true},
		{Name: "Chat non-streaming support", Category: "api", Description: "Non-streaming chat responses", NineRouter: true, AIProxy: true},

		// Completions (Legacy)
		{Name: "POST /v1/completions", Category: "api", Description: "Legacy completions endpoint", NineRouter: true, AIProxy: true},

		// Models
		{Name: "GET /v1/models", Category: "api", Description: "List all models", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/models/{kind}", Category: "api", Description: "List models by kind", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/models/info", Category: "api", Description: "Get model info", NineRouter: true, AIProxy: true},

		// Embeddings
		{Name: "POST /v1/embeddings", Category: "api", Description: "Generate embeddings", NineRouter: true, AIProxy: true},

		// Images
		{Name: "POST /v1/images/generations", Category: "api", Description: "Generate images", NineRouter: true, AIProxy: true},

		// Audio/TTS
		{Name: "POST /v1/audio/speech", Category: "api", Description: "Text-to-speech", NineRouter: true, AIProxy: true},
		{Name: "POST /v1/audio/transcriptions", Category: "api", Description: "Speech-to-text", NineRouter: true, AIProxy: true},

		// Search
		{Name: "POST /v1/search", Category: "api", Description: "Search endpoint", NineRouter: true, AIProxy: true},

		// Fetch
		{Name: "POST /v1/fetch", Category: "api", Description: "Web fetch endpoint", NineRouter: true, AIProxy: true},

		// Files API
		{Name: "GET /v1/files", Category: "api", Description: "List files", NineRouter: true, AIProxy: true},
		{Name: "POST /v1/files", Category: "api", Description: "Upload file", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/files/{id}", Category: "api", Description: "Get file info", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/files/{id}/content", Category: "api", Description: "Get file content", NineRouter: true, AIProxy: true},
		{Name: "DELETE /v1/files/{id}", Category: "api", Description: "Delete file", NineRouter: true, AIProxy: true},

		// Fine-tuning API
		{Name: "POST /v1/fine_tuning/jobs", Category: "api", Description: "Create fine-tuning job", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/fine_tuning/jobs", Category: "api", Description: "List fine-tuning jobs", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/fine_tuning/jobs/{id}", Category: "api", Description: "Get fine-tuning job", NineRouter: true, AIProxy: true},
		{Name: "POST /v1/fine_tuning/jobs/{id}/cancel", Category: "api", Description: "Cancel fine-tuning job", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/fine_tuning/jobs/{id}/events", Category: "api", Description: "List fine-tuning events", NineRouter: true, AIProxy: true},

		// Batch API
		{Name: "POST /v1/batches", Category: "api", Description: "Create batch", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/batches", Category: "api", Description: "List batches", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/batches/{id}", Category: "api", Description: "Get batch", NineRouter: true, AIProxy: true},
		{Name: "POST /v1/batches/{id}/cancel", Category: "api", Description: "Cancel batch", NineRouter: true, AIProxy: true},

		// Assistants API
		{Name: "POST /v1/assistants", Category: "api", Description: "Create assistant", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/assistants", Category: "api", Description: "List assistants", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/assistants/{id}", Category: "api", Description: "Get assistant", NineRouter: true, AIProxy: true},
		{Name: "POST /v1/assistants/{id}", Category: "api", Description: "Update assistant", NineRouter: true, AIProxy: true},
		{Name: "DELETE /v1/assistants/{id}", Category: "api", Description: "Delete assistant", NineRouter: true, AIProxy: true},

		// Threads API
		{Name: "POST /v1/threads", Category: "api", Description: "Create thread", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/threads/{id}", Category: "api", Description: "Get thread", NineRouter: true, AIProxy: true},
		{Name: "POST /v1/threads/{id}", Category: "api", Description: "Update thread", NineRouter: true, AIProxy: true},
		{Name: "DELETE /v1/threads/{id}", Category: "api", Description: "Delete thread", NineRouter: true, AIProxy: true},

		// Messages API
		{Name: "POST /v1/threads/{id}/messages", Category: "api", Description: "Create message", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/threads/{id}/messages", Category: "api", Description: "List messages", NineRouter: true, AIProxy: true},
		{Name: "GET /v1/threads/{id}/messages/{mid}", Category: "api", Description: "Get message", NineRouter: true, AIProxy: true},
		{Name: "POST /v1/threads/{id}/messages/{mid}", Category: "api", Description: "Update message", NineRouter: true, AIProxy: true},
		{Name: "DELETE /v1/threads/{id}/messages/{mid}", Category: "api", Description: "Delete message", NineRouter: true, AIProxy: true},
		{Name: "POST /v1/messages/count_tokens", Category: "api", Description: "Count message tokens", NineRouter: true, AIProxy: true},

		// Admin API
		{Name: "POST /api/login", Category: "api", Description: "User login", NineRouter: true, AIProxy: true},
		{Name: "POST /api/logout", Category: "api", Description: "User logout", NineRouter: true, AIProxy: true},
		{Name: "GET /api/me", Category: "api", Description: "Get current user", NineRouter: true, AIProxy: true},
		{Name: "GET /api/admin/usage", Category: "api", Description: "List usage records", NineRouter: true, AIProxy: true},
		{Name: "GET /api/admin/usage/stats", Category: "api", Description: "Get usage stats", NineRouter: true, AIProxy: true},
		{Name: "GET /api/admin/pricing", Category: "api", Description: "List pricing", NineRouter: true, AIProxy: true},
		{Name: "POST /api/admin/pricing", Category: "api", Description: "Create pricing", NineRouter: true, AIProxy: true},
		{Name: "GET /api/provider-nodes", Category: "api", Description: "List provider nodes", NineRouter: true, AIProxy: true},
		{Name: "POST /api/provider-nodes", Category: "api", Description: "Create provider node", NineRouter: true, AIProxy: true},
		{Name: "PATCH /api/provider-nodes/{id}", Category: "api", Description: "Update provider node", NineRouter: true, AIProxy: true},
		{Name: "DELETE /api/provider-nodes/{id}", Category: "api", Description: "Delete provider node", NineRouter: true, AIProxy: true},
		{Name: "POST /api/provider-nodes/{id}/test", Category: "api", Description: "Test provider node", NineRouter: true, AIProxy: true},

		// Responses API (Anthropic-style)
		{Name: "POST /v1/responses", Category: "api", Description: "Responses API endpoint", NineRouter: true, AIProxy: true},
	}
}

// getProviderFeatures returns all provider features.
func getProviderFeatures() []Feature {
	return []Feature{
		// Major LLM Providers
		{Name: "OpenAI provider", Category: "provider", Description: "OpenAI GPT models", NineRouter: true, AIProxy: true},
		{Name: "Claude provider", Category: "provider", Description: "Anthropic Claude models", NineRouter: true, AIProxy: true},
		{Name: "Gemini provider", Category: "provider", Description: "Google Gemini models", NineRouter: true, AIProxy: true},
		{Name: "GitHub Models provider", Category: "provider", Description: "GitHub Models marketplace", NineRouter: true, AIProxy: true},
		{Name: "Grok provider", Category: "provider", Description: "xAI Grok models", NineRouter: true, AIProxy: true},
		{Name: "Codex provider", Category: "provider", Description: "OpenAI Codex models", NineRouter: true, AIProxy: true},
		{Name: "Cursor provider", Category: "provider", Description: "Cursor AI models", NineRouter: true, AIProxy: true},
		{Name: "Kiro provider", Category: "provider", Description: "Kiro AI models", NineRouter: true, AIProxy: true},
		{Name: "Qoder provider", Category: "provider", Description: "Qoder models", NineRouter: true, AIProxy: true},
		{Name: "Ollama provider", Category: "provider", Description: "Local Ollama models", NineRouter: true, AIProxy: true},

		// Provider capabilities
		{Name: "Provider chat streaming", Category: "provider", Description: "Streaming chat for all providers", NineRouter: true, AIProxy: true},
		{Name: "Provider embeddings", Category: "provider", Description: "Embeddings generation", NineRouter: true, AIProxy: true},
		{Name: "Provider image generation", Category: "provider", Description: "Image generation support", NineRouter: true, AIProxy: true},
		{Name: "Provider TTS", Category: "provider", Description: "Text-to-speech support", NineRouter: true, AIProxy: true},
		{Name: "Provider STT", Category: "provider", Description: "Speech-to-text support", NineRouter: true, AIProxy: true},
	}
}

// getRTKFeatures returns RTK filter features.
func getRTKFeatures() []Feature {
	return []Feature{
		{Name: "RTK dedup_log filter", Category: "rtk", Description: "Deduplicate log lines", NineRouter: true, AIProxy: true},
		{Name: "RTK find filter", Category: "rtk", Description: "Find text in content", NineRouter: true, AIProxy: true},
		{Name: "RTK gitdiff filter", Category: "rtk", Description: "Format git diff output", NineRouter: true, AIProxy: true},
		{Name: "RTK gitstatus filter", Category: "rtk", Description: "Format git status output", NineRouter: true, AIProxy: true},
		{Name: "RTK grep filter", Category: "rtk", Description: "Grep-like search", NineRouter: true, AIProxy: true},
		{Name: "RTK ls filter", Category: "rtk", Description: "List directory contents", NineRouter: true, AIProxy: true},
		{Name: "RTK read_numbered filter", Category: "rtk", Description: "Numbered file reading", NineRouter: true, AIProxy: true},
		{Name: "RTK search_list filter", Category: "rtk", Description: "Search list generation", NineRouter: true, AIProxy: true},
		{Name: "RTK smart_truncate filter", Category: "rtk", Description: "Intelligent truncation", NineRouter: true, AIProxy: true},
		{Name: "RTK tree filter", Category: "rtk", Description: "Directory tree display", NineRouter: true, AIProxy: true},
		{Name: "Caveman prompts", Category: "rtk", Description: "Token optimization prompts", NineRouter: true, AIProxy: true},
		{Name: "RTK auto-detection", Category: "rtk", Description: "Auto-detect RTK in messages", NineRouter: true, AIProxy: true},
	}
}

// getTranslationFeatures returns translation features.
func getTranslationFeatures() []Feature {
	return []Feature{
		{Name: "OpenAI to Claude translation", Category: "translation", Description: "Convert OpenAI format to Claude", NineRouter: true, AIProxy: true},
		{Name: "Claude to OpenAI translation", Category: "translation", Description: "Convert Claude format to OpenAI", NineRouter: true, AIProxy: true},
		{Name: "OpenAI to Gemini translation", Category: "translation", Description: "Convert OpenAI format to Gemini", NineRouter: true, AIProxy: true},
		{Name: "Gemini to OpenAI translation", Category: "translation", Description: "Convert Gemini format to OpenAI", NineRouter: true, AIProxy: true},
		{Name: "Streaming translation", Category: "translation", Description: "Translate streaming responses", NineRouter: true, AIProxy: true},
		{Name: "Error format translation", Category: "translation", Description: "Translate error responses", NineRouter: true, AIProxy: true},
	}
}

// getOAuthFeatures returns OAuth features.
func getOAuthFeatures() []Feature {
	return []Feature{
		{Name: "Claude OAuth", Category: "oauth", Description: "Claude device code flow", NineRouter: true, AIProxy: true},
		{Name: "Gemini OAuth", Category: "oauth", Description: "Gemini OAuth flow", NineRouter: true, AIProxy: true},
		{Name: "GitHub OAuth", Category: "oauth", Description: "GitHub OAuth flow", NineRouter: true, AIProxy: true},
		{Name: "Cursor OAuth", Category: "oauth", Description: "Cursor OAuth flow", NineRouter: true, AIProxy: true},
		{Name: "Kiro OAuth", Category: "oauth", Description: "Kiro OAuth flow", NineRouter: true, AIProxy: true},
		{Name: "Token refresh", Category: "oauth", Description: "Automatic token refresh", NineRouter: true, AIProxy: true},
		{Name: "Concurrent refresh deduplication", Category: "oauth", Description: "Deduplicate concurrent refreshes", NineRouter: true, AIProxy: true},
		{Name: "Token expiration handling", Category: "oauth", Description: "Handle expired tokens", NineRouter: true, AIProxy: true},
	}
}

// getUsageFeatures returns usage tracking features.
func getUsageFeatures() []Feature {
	return []Feature{
		{Name: "Token counting", Category: "usage", Description: "Count input/output tokens", NineRouter: true, AIProxy: true},
		{Name: "Cost calculation", Category: "usage", Description: "Calculate request costs", NineRouter: true, AIProxy: true},
		{Name: "Usage aggregation", Category: "usage", Description: "Aggregate usage statistics", NineRouter: true, AIProxy: true},
		{Name: "Analytics queries", Category: "usage", Description: "Query usage analytics", NineRouter: true, AIProxy: true},
		{Name: "Usage persistence", Category: "usage", Description: "Store usage records", NineRouter: true, AIProxy: true},
	}
}

// getFallbackFeatures returns fallback features.
func getFallbackFeatures() []Feature {
	return []Feature{
		{Name: "Account fallback", Category: "fallback", Description: "Fallback to backup accounts", NineRouter: true, AIProxy: true},
		{Name: "Combo fallback", Category: "fallback", Description: "Fallback via combo definitions", NineRouter: true, AIProxy: true},
		{Name: "Exponential backoff", Category: "fallback", Description: "Exponential backoff on errors", NineRouter: true, AIProxy: true},
		{Name: "Fallback order priority", Category: "fallback", Description: "Subscription > pay-per-use > free", NineRouter: true, AIProxy: true},
		{Name: "Error aggregation", Category: "fallback", Description: "Aggregate errors from all attempts", NineRouter: true, AIProxy: true},
	}
}

// calculateSummary computes aggregate statistics from features.
func calculateSummary(features []Feature) Summary {
	s := Summary{TotalFeatures: len(features)}
	for _, f := range features {
		if f.NineRouter {
			s.NineRouterFeatures++
		}
		if f.AIProxy {
			s.AIProxyFeatures++
		}
		if f.Tested {
			s.TestedFeatures++
		}
		if f.Passing {
			s.PassingFeatures++
		}
	}
	if s.NineRouterFeatures > 0 {
		s.ParityPercentage = (s.AIProxyFeatures * 100) / s.NineRouterFeatures
	}
	return s
}

// ValidateFeatureInventory validates all features and returns errors.
func ValidateFeatureInventory(inventory *FeatureInventory) []error {
	var errors []error
	for _, f := range inventory.Features {
		if f.NineRouter && !f.AIProxy {
			errors = append(errors, fmt.Errorf("missing feature: %s (%s)", f.Name, f.Category))
		}
		if f.Tested && !f.Passing {
			errors = append(errors, fmt.Errorf("failing test: %s (%s)", f.Name, f.Category))
		}
	}
	return errors
}

// ToJSON returns the inventory as formatted JSON.
func (i *FeatureInventory) ToJSON() (string, error) {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveToFile saves the inventory to a JSON file.
func (i *FeatureInventory) SaveToFile(path string) error {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadFromFile loads inventory from a JSON file.
func LoadFromFile(path string) (*FeatureInventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var inventory FeatureInventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil, err
	}
	return &inventory, nil
}

// GetFeaturesByCategory returns features filtered by category.
func (i *FeatureInventory) GetFeaturesByCategory(category string) []Feature {
	var result []Feature
	for _, f := range i.Features {
		if strings.EqualFold(f.Category, category) {
			result = append(result, f)
		}
	}
	return result
}

// GetMissingFeatures returns features that exist in 9router but not in ai_proxy.
func (i *FeatureInventory) GetMissingFeatures() []Feature {
	var missing []Feature
	for _, f := range i.Features {
		if f.NineRouter && !f.AIProxy {
			missing = append(missing, f)
		}
	}
	return missing
}

// GetFailingTests returns features with failing tests.
func (i *FeatureInventory) GetFailingTests() []Feature {
	var failing []Feature
	for _, f := range i.Features {
		if f.Tested && !f.Passing {
			failing = append(failing, f)
		}
	}
	return failing
}

// MarkTested marks a feature as tested.
func (i *FeatureInventory) MarkTested(name string, passing bool) {
	for idx := range i.Features {
		if i.Features[idx].Name == name {
			i.Features[idx].Tested = true
			i.Features[idx].Passing = passing
			break
		}
	}
	i.Summary = calculateSummary(i.Features)
}
