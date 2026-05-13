// ref: _ref/9router/open-sse/handlers/embeddingProviders/openai.js
package providers

import (
	"strconv"
)

// OpenAICompatibleProvider handles embeddings for OpenAI-compatible providers.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/openai.js
type OpenAICompatibleProvider struct {
	endpoint   string
	providerID string
}

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

func (p *OpenAICompatibleProvider) BuildURL(model string, creds *Credentials, input Input) string {
	return p.endpoint
}

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

func (p *OpenAICompatibleProvider) Normalize(responseBody map[string]interface{}, model string) *EmbeddingResponse {
	if obj, ok := responseBody["object"].(string); ok && obj == "list" {
		return p.normalizeOpenAI(responseBody, model)
	}
	return &EmbeddingResponse{
		Object: "list",
		Data:   []EmbeddingData{},
		Model:  model,
		Usage:  Usage{},
	}
}

func (p *OpenAICompatibleProvider) normalizeOpenAI(responseBody map[string]interface{}, model string) *EmbeddingResponse {
	resp := &EmbeddingResponse{
		Object: "list",
		Model:  model,
	}
	if data, ok := responseBody["data"].([]interface{}); ok {
		resp.Data = make([]EmbeddingData, 0, len(data))
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				emb := EmbeddingData{Object: "embedding"}
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
