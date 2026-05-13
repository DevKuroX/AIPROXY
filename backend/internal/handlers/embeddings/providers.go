// Package embeddings provides embedding provider implementations.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/index.js
package embeddings

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// EmbeddingProvider defines the interface for embedding providers.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/_base.js
type EmbeddingProvider interface {
	// BuildURL constructs the upstream API URL for the embedding request.
	BuildURL(model string, creds *Credentials, input Input) string
	// BuildHeaders constructs HTTP headers for the embedding request.
	BuildHeaders(creds *Credentials) map[string]string
	// BuildBody constructs the request body for the embedding request.
	BuildBody(model string, req *EmbeddingRequest) interface{}
	// Normalize transforms the provider response to OpenAI format.
	Normalize(responseBody map[string]interface{}, model string) *EmbeddingResponse
}

// Credentials holds authentication credentials.
type Credentials struct {
	APIKey       string
	AccessToken  string
	BaseURL      string // For custom nodes
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
	Dimensions     interface{} `json:"dimensions,omitempty"` // int or string
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

// OpenAICompatibleProvider handles embeddings for OpenAI-compatible providers.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/openai.js
type OpenAICompatibleProvider struct {
	endpoint   string
	providerID string
}

// OpenAI-compatible provider endpoints.
var openAIEndpoints = map[string]string{
	"openai":     "https://api.openai.com/v1/embeddings",
	"openrouter": "https://openrouter.ai/api/v1/embeddings",
	"mistral":    "https://api.mistral.ai/v1/embeddings",
	"voyage-ai":  "https://api.voyageai.com/v1/embeddings",
	"fireworks":  "https://api.fireworks.ai/inference/v1/embeddings",
	"together":   "https://api.together.xyz/v1/embeddings",
	"nebius":     "https://api.tokenfactory.nebius.com/v1/embeddings",
	"github":     "https://models.github.ai/inference/embeddings",
	"nvidia":     "https://integrate.api.nvidia.com/v1/embeddings",
	"jina-ai":    "https://api.jina.ai/v1/embeddings",
}

// NewOpenAICompatibleProvider creates a new OpenAI-compatible embedding provider.
func NewOpenAICompatibleProvider(providerID string) *OpenAICompatibleProvider {
	endpoint, ok := openAIEndpoints[providerID]
	if !ok {
		endpoint = "https://api.openai.com/v1/embeddings"
	}
	return &OpenAICompatibleProvider{
		endpoint:   endpoint,
		providerID: providerID,
	}
}

// BuildURL returns the endpoint URL for the provider.
func (p *OpenAICompatibleProvider) BuildURL(model string, creds *Credentials, input Input) string {
	return p.endpoint
}

// BuildHeaders returns headers for the request.
func (p *OpenAICompatibleProvider) BuildHeaders(creds *Credentials) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	for k, v := range BearerAuth(creds) {
		headers[k] = v
	}
	if p.providerID == "openrouter" {
		headers["HTTP-Referer"] = "https://endpoint-proxy.local"
		headers["X-Title"] = "Endpoint Proxy"
	}
	return headers
}

// BuildBody constructs the request body.
func (p *OpenAICompatibleProvider) BuildBody(model string, req *EmbeddingRequest) interface{} {
	body := map[string]interface{}{
		"model": model,
		"input": req.Input,
	}
	if req.EncodingFormat != "" {
		body["encoding_format"] = req.EncodingFormat
	}
	if req.Dimensions != nil {
		switch v := req.Dimensions.(type) {
		case float64:
			if v > 0 {
				body["dimensions"] = int(v)
			}
		case int:
			if v > 0 {
				body["dimensions"] = v
			}
		case string:
			if v != "" {
				if dim, err := strconv.Atoi(v); err == nil && dim > 0 {
					body["dimensions"] = dim
				}
			}
		}
	}
	return body
}

// Normalize returns the response as-is for OpenAI-compatible providers.
func (p *OpenAICompatibleProvider) Normalize(responseBody map[string]interface{}, model string) *EmbeddingResponse {
	// If already in OpenAI format, return as-is
	if obj, ok := responseBody["object"].(string); ok && obj == "list" {
		resp := &EmbeddingResponse{
			Object: "list",
			Model:  model,
		}
		if data, ok := responseBody["data"].([]interface{}); ok {
			for _, item := range data {
				if m, ok := item.(map[string]interface{}); ok {
					emb := EmbeddingData{}
					if idx, ok := m["index"].(float64); ok {
						emb.Index = int(idx)
					}
					if embVals, ok := m["embedding"].([]interface{}); ok {
						emb.Embedding = make([]float64, len(embVals))
						for i, v := range embVals {
							if f, ok := v.(float64); ok {
								emb.Embedding[i] = f
							}
						}
					}
					emb.Object = "embedding"
					resp.Data = append(resp.Data, emb)
				}
			}
		}
		if usage, ok := responseBody["usage"].(map[string]interface{}); ok {
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				resp.Usage.PromptTokens = int(pt)
			}
			if tt, ok := usage["total_tokens"].(float64); ok {
				resp.Usage.TotalTokens = int(tt)
			}
		}
		return resp
	}
	return &EmbeddingResponse{
		Object: "list",
		Data:   []EmbeddingData{},
		Model:  model,
		Usage:  Usage{},
	}
}

// GeminiProvider handles embeddings for Google Gemini.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/gemini.js
type GeminiProvider struct{}

const geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// modelPath normalizes the model path for Gemini API.
func modelPath(model string) string {
	if strings.HasPrefix(model, "models/") {
		return model
	}
	return "models/" + model
}

// BuildURL returns the Gemini API URL.
func (p *GeminiProvider) BuildURL(model string, creds *Credentials, input Input) string {
	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = creds.AccessToken
	}
	path := modelPath(model)
	var op string
	if _, ok := input.([]interface{}); ok {
		op = "batchEmbedContents"
	} else {
		op = "embedContent"
	}
	return fmt.Sprintf("%s/%s:%s?key=%s", geminiBaseURL, path, op, url.QueryEscape(apiKey))
}

// BuildHeaders returns headers for Gemini requests.
func (p *GeminiProvider) BuildHeaders(creds *Credentials) map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// BuildBody constructs the Gemini request body.
func (p *GeminiProvider) BuildBody(model string, req *EmbeddingRequest) interface{} {
	m := modelPath(model)
	if arr, ok := req.Input.([]interface{}); ok {
		requests := make([]map[string]interface{}, len(arr))
		for i, text := range arr {
			requests[i] = map[string]interface{}{
				"model": m,
				"content": map[string]interface{}{
					"parts": []interface{}{
						map[string]interface{}{"text": fmt.Sprintf("%v", text)},
					},
				},
			}
		}
		return map[string]interface{}{"requests": requests}
	}
	return map[string]interface{}{
		"model": m,
		"content": map[string]interface{}{
			"parts": []interface{}{
				map[string]interface{}{"text": fmt.Sprintf("%v", req.Input)},
			},
		},
	}
}

// Normalize transforms Gemini response to OpenAI format.
func (p *GeminiProvider) Normalize(responseBody map[string]interface{}, model string) *EmbeddingResponse {
	// Already OpenAI format
	if obj, ok := responseBody["object"].(string); ok && obj == "list" {
		return p.normalizeOpenAI(responseBody, model)
	}

	resp := &EmbeddingResponse{
		Object: "list",
		Model:  model,
		Usage:  Usage{PromptTokens: 0, TotalTokens: 0},
	}

	// Handle batch response
	if embeddings, ok := responseBody["embeddings"].([]interface{}); ok {
		resp.Data = make([]EmbeddingData, len(embeddings))
		for i, emb := range embeddings {
			if m, ok := emb.(map[string]interface{}); ok {
				if values, ok := m["values"].([]interface{}); ok {
					resp.Data[i] = EmbeddingData{
						Object:    "embedding",
						Index:     i,
						Embedding: make([]float64, len(values)),
					}
					for j, v := range values {
						if f, ok := v.(float64); ok {
							resp.Data[i].Embedding[j] = f
						}
					}
				}
			}
		}
		return resp
	}

	// Handle single response
	if emb, ok := responseBody["embedding"].(map[string]interface{}); ok {
		if values, ok := emb["values"].([]interface{}); ok {
			resp.Data = []EmbeddingData{{
				Object:    "embedding",
				Index:     0,
				Embedding: make([]float64, len(values)),
			}}
			for i, v := range values {
				if f, ok := v.(float64); ok {
					resp.Data[0].Embedding[i] = f
				}
			}
		}
	}

	return resp
}

func (p *GeminiProvider) normalizeOpenAI(responseBody map[string]interface{}, model string) *EmbeddingResponse {
	resp := &EmbeddingResponse{
		Object: "list",
		Model:  model,
	}
	if data, ok := responseBody["data"].([]interface{}); ok {
		resp.Data = make([]EmbeddingData, len(data))
		for i, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				emb := EmbeddingData{Object: "embedding", Index: i}
				if idx, ok := m["index"].(float64); ok {
					emb.Index = int(idx)
				}
				if embVals, ok := m["embedding"].([]interface{}); ok {
					emb.Embedding = make([]float64, len(embVals))
					for j, v := range embVals {
						if f, ok := v.(float64); ok {
							emb.Embedding[j] = f
						}
					}
				}
				resp.Data[i] = emb
			}
		}
	}
	if usage, ok := responseBody["usage"].(map[string]interface{}); ok {
		if pt, ok := usage["prompt_tokens"].(float64); ok {
			resp.Usage.PromptTokens = int(pt)
		}
		if tt, ok := usage["total_tokens"].(float64); ok {
			resp.Usage.TotalTokens = int(tt)
		}
	}
	return resp
}

// OpenAICompatNodeProvider handles custom OpenAI-compatible nodes.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/openaiCompatNode.js
type OpenAICompatNodeProvider struct {
	base *OpenAICompatibleProvider
}

// NewOpenAICompatNodeProvider creates a new OpenAI-compatible node provider.
func NewOpenAICompatNodeProvider() *OpenAICompatNodeProvider {
	return &OpenAICompatNodeProvider{
		base: NewOpenAICompatibleProvider("openai"),
	}
}

// BuildURL returns the base URL from credentials.
func (p *OpenAICompatNodeProvider) BuildURL(model string, creds *Credentials, input Input) string {
	rawBaseURL := creds.BaseURL
	if rawBaseURL == "" {
		if data, ok := creds.ProviderData["baseUrl"].(string); ok {
			rawBaseURL = data
		}
	}
	if rawBaseURL == "" {
		rawBaseURL = "https://api.openai.com/v1"
	}
	baseURL := strings.TrimSuffix(rawBaseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/embeddings")
	return baseURL + "/embeddings"
}

// BuildHeaders delegates to base provider.
func (p *OpenAICompatNodeProvider) BuildHeaders(creds *Credentials) map[string]string {
	return p.base.BuildHeaders(creds)
}

// BuildBody delegates to base provider.
func (p *OpenAICompatNodeProvider) BuildBody(model string, req *EmbeddingRequest) interface{} {
	return p.base.BuildBody(model, req)
}

// Normalize delegates to base provider.
func (p *OpenAICompatNodeProvider) Normalize(responseBody map[string]interface{}, model string) *EmbeddingResponse {
	return p.base.Normalize(responseBody, model)
}

// Provider registry
var providers = make(map[string]EmbeddingProvider)

func init() {
	// Register OpenAI-compatible providers
	for id := range openAIEndpoints {
		providers[id] = NewOpenAICompatibleProvider(id)
	}
	// Register Gemini providers
	gemini := &GeminiProvider{}
	providers["gemini"] = gemini
	providers["google_ai_studio"] = gemini
}

// GetProvider returns the embedding provider for the given provider ID.
// Returns nil if the provider doesn't support embeddings.
func GetProvider(providerID string) EmbeddingProvider {
	if p, ok := providers[providerID]; ok {
		return p
	}
	// Check for custom node providers
	if strings.HasPrefix(providerID, "openai-compatible-") || strings.HasPrefix(providerID, "custom-embedding-") {
		return NewOpenAICompatNodeProvider()
	}
	return nil
}

// RegisterProvider allows registering custom embedding providers.
func RegisterProvider(providerID string, provider EmbeddingProvider) {
	providers[providerID] = provider
}

// MarshalEmbeddingResponse converts EmbeddingResponse to JSON bytes.
func MarshalEmbeddingResponse(resp *EmbeddingResponse) ([]byte, error) {
	return json.Marshal(resp)
}
