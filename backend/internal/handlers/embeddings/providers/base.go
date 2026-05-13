// Package providers provides embedding provider implementations.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/_base.js
package providers

// EmbeddingProvider defines the interface for embedding providers.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/_base.js
type EmbeddingProvider interface {
	BuildURL(model string, creds *Credentials, input Input) string
	BuildHeaders(creds *Credentials) map[string]string
	BuildBody(model string, req *EmbeddingRequest) interface{}
	Normalize(responseBody map[string]interface{}, model string) *EmbeddingResponse
}

// Credentials holds authentication credentials.
type Credentials struct {
	APIKey       string
	AccessToken  string
	BaseURL      string
	ProviderData map[string]interface{}
}

// Input represents the input for embedding (string or array of strings).
type Input interface{}

// EmbeddingRequest represents an OpenAI-compatible embedding request.
// ref: _ref/9router/open-sse/handlers/embeddingsCore.js
type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          Input       `json:"input"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     interface{} `json:"dimensions,omitempty"`
}

// EmbeddingData represents a single embedding in the response.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// EmbeddingResponse represents an OpenAI-compatible embedding response.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  Usage           `json:"usage"`
}

// BearerAuth returns a bearer authorization header.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/_base.js
func BearerAuth(creds *Credentials) map[string]string {
	token := creds.APIKey
	if token == "" {
		token = creds.AccessToken
	}
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}
