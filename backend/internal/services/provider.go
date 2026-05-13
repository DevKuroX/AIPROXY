package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

// ProviderService handles provider detection and URL building.
// ref: 9router/open-sse/services/provider.js
type ProviderService struct {
	storage    *storage.DB
	httpClient *http.Client
	logger     Logger
}

// Logger interface for service logging
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// NewProviderService creates a new ProviderService instance.
func NewProviderService(db *storage.DB, logger Logger) *ProviderService {
	return &ProviderService{
		storage: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ProviderCredentials holds authentication credentials for a provider
type ProviderCredentials struct {
	APIKey       string
	AccessToken  string
	CopilotToken string
}

// ProviderConfig holds normalized provider configuration
// ref: 9router/open-sse/config/providers.js
type ProviderConfig struct {
	BaseURL      string
	Format       string
	Headers      map[string]string
	ClientID     string
	ClientSecret string
	TokenURL     string
	AuthURL      string
}

// RequestFormat represents the detected request format type
type RequestFormat string

const (
	FormatOpenAI         RequestFormat = "openai"
	FormatOpenAIResponses RequestFormat = "openai-responses"
	FormatClaude         RequestFormat = "claude"
	FormatGemini         RequestFormat = "gemini"
	FormatAntigravity    RequestFormat = "antigravity"
)

// DetectFormat detects the request format from the body structure.
// ref: 9router/open-sse/services/provider.js:48-126
func (s *ProviderService) DetectFormat(body map[string]any) RequestFormat {
	// OpenAI Responses API: has input (array or string) instead of messages[]
	// The Responses API accepts both input as array and input as a plain string
	// ref: 9router/open-sse/services/provider.js:50-54
	if input, ok := body["input"]; ok && !bodyHasMessages(body) {
		if _, isArray := input.([]any); isArray {
			return FormatOpenAIResponses
		}
		if _, isString := input.(string); isString {
			return FormatOpenAIResponses
		}
	}

	// Antigravity format: Gemini wrapped in body.request
	// ref: 9router/open-sse/services/provider.js:56-59
	if req, ok := body["request"].(map[string]any); ok {
		if _, hasContents := req["contents"]; hasContents {
			if ua, ok := body["userAgent"].(string); ok && ua == "antigravity" {
				return FormatAntigravity
			}
		}
	}

	// Gemini format: has contents array
	// ref: 9router/open-sse/services/provider.js:61-64
	if contents, ok := body["contents"].([]any); ok && len(contents) > 0 {
		return FormatGemini
	}

	// OpenAI-specific indicators (check BEFORE Claude)
	// These fields are OpenAI-specific and never appear in Claude format
	// ref: 9router/open-sse/services/provider.js:66-80
	if hasOpenAIIndicators(body) {
		return FormatOpenAI
	}

	// Claude format: messages with content as array of objects with type
	// Claude requires content to be array with specific structure
	// ref: 9router/open-sse/services/provider.js:82-122
	if messages, ok := body["messages"].([]any); ok && len(messages) > 0 {
		if isClaudeFormat(messages, body) {
			return FormatClaude
		}
	}

	// Default to OpenAI format
	// ref: 9router/open-sse/services/provider.js:124-125
	return FormatOpenAI
}

// bodyHasMessages checks if body has a messages field
func bodyHasMessages(body map[string]any) bool {
	_, ok := body["messages"]
	return ok
}

// hasOpenAIIndicators checks for OpenAI-specific fields
// ref: 9router/open-sse/services/provider.js:68-80
func hasOpenAIIndicators(body map[string]any) bool {
	// OpenAI-specific indicators
	if _, ok := body["stream_options"]; ok {
		return true
	}
	if _, ok := body["response_format"]; ok {
		return true
	}
	if _, ok := body["logprobs"]; ok {
		return true
	}
	if _, ok := body["top_logprobs"]; ok {
		return true
	}
	if _, ok := body["n"]; ok {
		return true
	}
	if _, ok := body["presence_penalty"]; ok {
		return true
	}
	if _, ok := body["frequency_penalty"]; ok {
		return true
	}
	if _, ok := body["logit_bias"]; ok {
		return true
	}
	if _, ok := body["user"]; ok {
		return true
	}
	return false
}

// isClaudeFormat checks if messages follow Claude structure
// ref: 9router/open-sse/services/provider.js:84-122
func isClaudeFormat(messages []any, body map[string]any) bool {
	firstMsg, ok := messages[0].(map[string]any)
	if !ok {
		return false
	}

	// Check for Claude-specific top-level fields
	if _, ok := body["system"]; ok {
		return true
	}
	if _, ok := body["anthropic_version"]; ok {
		return true
	}

	// Check content structure
	content, ok := firstMsg["content"]
	if !ok {
		return false
	}

	// If content is array, check for Claude-specific types
	if contentArr, ok := content.([]any); ok && len(contentArr) > 0 {
		firstContent, ok := contentArr[0].(map[string]any)
		if !ok {
			return false
		}

		contentType, _ := firstContent["type"].(string)
		if contentType != "text" {
			return false
		}

		// Check for Claude image format (source.type === "base64")
		for _, c := range contentArr {
			if cMap, ok := c.(map[string]any); ok {
				if cMap["type"] == "image" {
					if source, ok := cMap["source"].(map[string]any); ok {
						if source["type"] == "base64" {
							return true
						}
					}
				}
				// Check for Claude tool format
				if cMap["type"] == "tool_use" || cMap["type"] == "tool_result" {
					return true
				}
			}
		}
	}

	return false
}

// BuildUpstreamURL constructs the upstream URL for a provider.
// ref: 9router/open-sse/services/provider.js:155-209
func (s *ProviderService) BuildUpstreamURL(provider *models.Provider, baseURL string, apiType string, stream bool, model string) string {
	// Normalize base URL
	normalized := strings.TrimSuffix(baseURL, "")

	switch provider.Type {
	case "claude":
		// ref: 9router/open-sse/services/provider.js:168-169
		return fmt.Sprintf("%s?beta=true", normalized)

	case "gemini":
		// ref: 9router/open-sse/services/provider.js:171-174
		action := "generateContent"
		if stream {
			action = "streamGenerateContent?alt=sse"
		}
		return fmt.Sprintf("%s/%s:%s", normalized, model, action)

	case "gemini-cli":
		// ref: 9router/open-sse/services/provider.js:176-179
		action := "generateContent"
		if stream {
			action = "streamGenerateContent?alt=sse"
		}
		return fmt.Sprintf("%s:%s", normalized, action)

	case "antigravity":
		// ref: 9router/open-sse/services/provider.js:181-187
		path := "/v1internal:generateContent"
		if stream {
			path = "/v1internal:streamGenerateContent?alt=sse"
		}
		return fmt.Sprintf("%s%s", normalized, path)

	case "codex":
		// ref: 9router/open-sse/services/provider.js:189-190
		return normalized

	case "qwen":
		// ref: 9router/open-sse/services/provider.js:192-195
		return fmt.Sprintf("%s/chat/completions", normalized)

	case "github":
		// ref: 9router/open-sse/services/provider.js:197-198
		return normalized

	case "glm", "kimi", "minimax", "minimax-cn":
		// Claude-compatible providers
		// ref: 9router/open-sse/services/provider.js:200-204
		return fmt.Sprintf("%s?beta=true", normalized)

	case "openai", "openrouter", "iflow", "qoder":
		// OpenAI-compatible providers
		// ref: 9router/open-sse/services/provider.js:286-289
		if apiType == "responses" {
			return fmt.Sprintf("%s/responses", normalized)
		}
		return fmt.Sprintf("%s/chat/completions", normalized)

	default:
		// Default: use as-is or append /chat/completions
		// ref: 9router/open-sse/services/provider.js:206-208
		if strings.Contains(normalized, "/chat/completions") || strings.Contains(normalized, "/messages") {
			return normalized
		}
		return fmt.Sprintf("%s/chat/completions", normalized)
	}
}

// BuildHeaders builds request headers for a provider.
// ref: 9router/open-sse/services/provider.js:212-320
func (s *ProviderService) BuildHeaders(provider *models.Provider, creds *ProviderCredentials, stream bool) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Add provider-specific headers
	switch provider.Type {
	case "claude":
		// ref: 9router/open-sse/services/provider.js:249-256
		if creds.APIKey != "" {
			headers["x-api-key"] = creds.APIKey
		} else if creds.AccessToken != "" {
			headers["Authorization"] = fmt.Sprintf("Bearer %s", creds.AccessToken)
		}
		headers["anthropic-version"] = "2023-06-01"

	case "gemini":
		// ref: 9router/open-sse/services/provider.js:235-241
		if creds.APIKey != "" {
			headers["x-goog-api-key"] = creds.APIKey
		} else if creds.AccessToken != "" {
			headers["Authorization"] = fmt.Sprintf("Bearer %s", creds.AccessToken)
		}

	case "antigravity", "gemini-cli":
		// ref: 9router/open-sse/services/provider.js:243-247
		headers["Authorization"] = fmt.Sprintf("Bearer %s", creds.AccessToken)

	case "github":
		// ref: 9router/open-sse/services/provider.js:258-281
		token := creds.CopilotToken
		if token == "" {
			token = creds.AccessToken
		}
		headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
		headers["copilot-integration-id"] = "vscode-chat"
		headers["editor-version"] = "vscode/1.107.1"
		headers["editor-plugin-version"] = "copilot-chat/0.26.7"
		headers["user-agent"] = "GitHubCopilotChat/0.26.7"
		headers["openai-intent"] = "conversation-panel"
		headers["x-github-api-version"] = "2025-04-01"
		headers["x-request-id"] = generateUUID()
		headers["x-vscode-user-agent-library-version"] = "electron-fetch"
		headers["X-Initiator"] = "user"
		headers["Accept"] = "application/json"

	case "glm", "kimi", "minimax", "minimax-cn":
		// ref: 9router/open-sse/services/provider.js:295-300
		headers["x-api-key"] = creds.APIKey

	case "codex", "qwen", "openai", "openrouter", "iflow", "qoder":
		// ref: 9router/open-sse/services/provider.js:284-289
		token := creds.APIKey
		if token == "" {
			token = creds.AccessToken
		}
		headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

	default:
		// ref: 9router/open-sse/services/provider.js:308-310
		token := creds.APIKey
		if token == "" {
			token = creds.AccessToken
		}
		if token != "" {
			headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
		}
	}

	// Stream accept header
	// ref: 9router/open-sse/services/provider.js:314-317
	if stream {
		headers["Accept"] = "text/event-stream"
	}

	return headers
}

// generateUUID generates a UUID string for request IDs
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// TestConnection tests provider connectivity by making a test request.
func (s *ProviderService) TestConnection(ctx context.Context, providerID string) error {
	provider, err := s.storage.GetProviderByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Build test URL
	testURL := s.BuildUpstreamURL(provider, provider.BaseURL, "chat", false, "test-model")

	// Build headers
	creds := &ProviderCredentials{
		APIKey: provider.APIKey,
	}
	headers := s.BuildHeaders(provider, creds, false)

	// Create test request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 500 {
		return fmt.Errorf("provider returned server error: %d", resp.StatusCode)
	}

	return nil
}

// NormalizeConfig normalizes provider configuration.
// ref: 9router/open-sse/services/provider.js:129-146
func (s *ProviderService) NormalizeConfig(provider *models.Provider, config map[string]any) *ProviderConfig {
	normalized := &ProviderConfig{
		BaseURL: provider.BaseURL,
		Format:  provider.Type,
		Headers: make(map[string]string),
	}

	// Set format based on provider type
	switch provider.Type {
	case "claude", "glm", "kimi", "minimax", "minimax-cn":
		normalized.Format = "claude"
	case "gemini":
		normalized.Format = "gemini"
	case "gemini-cli":
		normalized.Format = "gemini-cli"
	case "antigravity":
		normalized.Format = "antigravity"
	case "codex":
		normalized.Format = "openai-responses"
	default:
		normalized.Format = "openai"
	}

	// Apply any additional config from the map
	if baseURL, ok := config["base_url"].(string); ok && baseURL != "" {
		normalized.BaseURL = baseURL
	}
	if format, ok := config["format"].(string); ok && format != "" {
		normalized.Format = format
	}
	if headers, ok := config["headers"].(map[string]any); ok {
		for k, v := range headers {
			if str, ok := v.(string); ok {
				normalized.Headers[k] = str
			}
		}
	}
	if clientID, ok := config["client_id"].(string); ok {
		normalized.ClientID = clientID
	}
	if clientSecret, ok := config["client_secret"].(string); ok {
		normalized.ClientSecret = clientSecret
	}
	if tokenURL, ok := config["token_url"].(string); ok {
		normalized.TokenURL = tokenURL
	}
	if authURL, ok := config["auth_url"].(string); ok {
		normalized.AuthURL = authURL
	}

	return normalized
}

// ListActiveProviders returns all enabled providers.
func (s *ProviderService) ListActiveProviders(ctx context.Context) ([]models.Provider, error) {
	providers, err := s.storage.ListProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	// Filter enabled providers
	var active []models.Provider
	for _, p := range providers {
		if p.Enabled {
			active = append(active, p)
		}
	}

	return active, nil
}

// GetTargetFormat returns the target format for a provider.
// ref: 9router/open-sse/services/provider.js:322-332
func (s *ProviderService) GetTargetFormat(providerType string) string {
	switch providerType {
	case "claude", "glm", "kimi", "minimax", "minimax-cn":
		return "claude"
	case "gemini":
		return "gemini"
	case "gemini-cli":
		return "gemini-cli"
	case "antigravity":
		return "antigravity"
	case "codex":
		return "openai-responses"
	default:
		return "openai"
	}
}

// IsLastMessageFromUser checks if the last message is from user.
// ref: 9router/open-sse/services/provider.js:335-340
func (s *ProviderService) IsLastMessageFromUser(body map[string]any) bool {
	var messages []any
	if m, ok := body["messages"].([]any); ok {
		messages = m
	} else if c, ok := body["contents"].([]any); ok {
		messages = c
	}

	if len(messages) == 0 {
		return true
	}

	lastMsg, ok := messages[len(messages)-1].(map[string]any)
	if !ok {
		return true
	}

	role, _ := lastMsg["role"].(string)
	return role == "user"
}

// HasThinkingConfig checks if request has thinking config.
// ref: 9router/open-sse/services/provider.js:343-345
func (s *ProviderService) HasThinkingConfig(body map[string]any) bool {
	if _, ok := body["reasoning_effort"]; ok {
		return true
	}
	if thinking, ok := body["thinking"].(map[string]any); ok {
		if t, ok := thinking["type"].(string); ok && t == "enabled" {
			return true
		}
	}
	return false
}

// NormalizeThinkingConfig normalizes thinking config based on last message role.
// ref: 9router/open-sse/services/provider.js:348-356
func (s *ProviderService) NormalizeThinkingConfig(body map[string]any) map[string]any {
	if !s.IsLastMessageFromUser(body) {
		delete(body, "reasoning_effort")
		delete(body, "thinking")
	}
	return body
}

// GetProviderFallbackCount returns the number of fallback URLs for a provider.
// ref: 9router/open-sse/services/provider.js:149-152
func (s *ProviderService) GetProviderFallbackCount(providerType string) int {
	// For now, return 1 as most providers have single URL
	// Antigravity has multiple fallback URLs
	if providerType == "antigravity" {
		return 2
	}
	return 1
}

// ParseRequestBody parses a JSON request body into a map.
func (s *ProviderService) ParseRequestBody(data []byte) (map[string]any, error) {
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse request body: %w", err)
	}
	return body, nil
}

// SetHTTPClient allows setting a custom HTTP client (useful for testing).
func (s *ProviderService) SetHTTPClient(client *http.Client) {
	s.httpClient = client
}
