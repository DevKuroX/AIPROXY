// ref: _ref/9router/open-sse/handlers/embeddingProviders/gemini.js
package providers

import (
	"fmt"
	"net/url"
	"strings"
)

const geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// GeminiProvider handles embeddings for Google Gemini.
// ref: _ref/9router/open-sse/handlers/embeddingProviders/gemini.js
type GeminiProvider struct{}

func modelPath(model string) string {
	if strings.HasPrefix(model, "models/") {
		return model
	}
	return "models/" + model
}

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

func (p *GeminiProvider) BuildHeaders(creds *Credentials) map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

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

func (p *GeminiProvider) Normalize(responseBody map[string]interface{}, model string) *EmbeddingResponse {
	if obj, ok := responseBody["object"].(string); ok && obj == "list" {
		return p.normalizeOpenAI(responseBody, model)
	}

	resp := &EmbeddingResponse{
		Object: "list",
		Model:  model,
		Usage:  Usage{PromptTokens: 0, TotalTokens: 0},
	}

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
