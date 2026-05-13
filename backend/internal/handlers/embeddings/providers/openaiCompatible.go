// ref: _ref/9router/open-sse/handlers/embeddingProviders/openaiCompatNode.js
package providers

import (
	"strings"
)

// OpenAICompatibleNode handles custom node providers with configurable base URL.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/openaiCompatNode.js
type OpenAICompatibleNode struct {
	base *OpenAICompatibleProvider
}

func NewOpenAICompatibleNode() *OpenAICompatibleNode {
	return &OpenAICompatibleNode{
		base: NewOpenAICompatibleProvider("openai"),
	}
}

func (p *OpenAICompatibleNode) BuildURL(model string, creds *Credentials, input Input) string {
	rawBaseURL := ""
	if creds.ProviderData != nil {
		if url, ok := creds.ProviderData["baseUrl"].(string); ok {
			rawBaseURL = url
		}
	}
	if rawBaseURL == "" {
		rawBaseURL = "https://api.openai.com/v1"
	}
	baseURL := strings.TrimSuffix(rawBaseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/embeddings")
	return baseURL + "/embeddings"
}

func (p *OpenAICompatibleNode) BuildHeaders(creds *Credentials) map[string]string {
	return p.base.BuildHeaders(creds)
}

func (p *OpenAICompatibleNode) BuildBody(model string, req *EmbeddingRequest) interface{} {
	return p.base.BuildBody(model, req)
}

func (p *OpenAICompatibleNode) Normalize(responseBody map[string]interface{}, model string) *EmbeddingResponse {
	return p.base.Normalize(responseBody, model)
}
