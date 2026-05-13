package providers

import (
	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
)

type openaiAdapter struct {
	image.BaseAdapter
	providerID string
	baseURL    string
}

var openaiEndpoints = map[string]string{
	"openai":     "https://api.openai.com/v1/images/generations",
	"minimax":    "https://api.minimaxi.com/v1/images/generations",
	"openrouter": "https://openrouter.ai/api/v1/images/generations",
	"recraft":    "https://external.api.recraft.ai/v1/images/generations",
}

func newOpenAIAdapter(providerID string) *openaiAdapter {
	return &openaiAdapter{
		providerID: providerID,
		baseURL:    openaiEndpoints[providerID],
		BaseAdapter: image.BaseAdapter{
			Async:     false,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func (a *openaiAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return a.baseURL, nil
}

func (a *openaiAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	headers := map[string]string{"Content-Type": "application/json"}
	key := creds.APIKey
	if key == "" {
		key = creds.AccessToken
	}
	if key != "" {
		headers["Authorization"] = "Bearer " + key
	}
	if a.providerID == "openrouter" {
		headers["HTTP-Referer"] = "https://endpoint-proxy.local"
		headers["X-Title"] = "Endpoint Proxy"
	}
	return headers, nil
}

func (a *openaiAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	n := body.N
	if n == 0 {
		n = 1
	}
	size := body.Size
	if size == "" {
		size = "1024x1024"
	}
	req := map[string]interface{}{
		"model":  model,
		"prompt": body.Prompt,
		"n":      n,
		"size":   size,
	}
	if body.Quality != "" {
		req["quality"] = body.Quality
	}
	if body.Style != "" {
		req["style"] = body.Style
	}
	if body.ResponseFormat != "" {
		req["response_format"] = body.ResponseFormat
	}
	return req, nil
}

func (a *openaiAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	case map[string]interface{}:
		created := int64(0)
		if c, ok := v["created"].(float64); ok {
			created = int64(c)
		}
		data := []image.ImageData{}
		if d, ok := v["data"].([]interface{}); ok {
			for _, item := range d {
				if m, ok := item.(map[string]interface{}); ok {
					img := image.ImageData{}
					if url, ok := m["url"].(string); ok {
						img.URL = url
					}
					if b64, ok := m["b64_json"].(string); ok {
						img.B64JSON = b64
					}
					if rp, ok := m["revised_prompt"].(string); ok {
						img.RevisedPrompt = rp
					}
					data = append(data, img)
				}
			}
		}
		return &image.ImageResponse{Created: created, Data: data}, nil
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	for providerID := range openaiEndpoints {
		image.RegisterAdapter(providerID, newOpenAIAdapter(providerID))
	}
}
