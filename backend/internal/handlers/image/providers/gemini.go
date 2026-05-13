package providers

import (
	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
	"net/url"
	"strings"
)

type geminiAdapter struct {
	image.BaseAdapter
}

func newGeminiAdapter() *geminiAdapter {
	return &geminiAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     false,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func (a *geminiAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = creds.AccessToken
	}
	modelID := strings.TrimPrefix(model, "models/")
	return "https://generativelanguage.googleapis.com/v1beta/models/" + modelID + ":generateContent?key=" + url.QueryEscape(apiKey), nil
}

func (a *geminiAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	return map[string]string{"Content-Type": "application/json"}, nil
}

func (a *geminiAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	return map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{"text": body.Prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseModalities": []string{"TEXT", "IMAGE"},
		},
	}, nil
}

func (a *geminiAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case map[string]interface{}:
		candidates, _ := v["candidates"].([]interface{})
		if len(candidates) == 0 {
			return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{{B64JSON: "", RevisedPrompt: prompt}}}, nil
		}

		candidate, _ := candidates[0].(map[string]interface{})
		content, _ := candidate["content"].(map[string]interface{})
		parts, _ := content["parts"].([]interface{})

		var images []image.ImageData
		for _, part := range parts {
			pm, ok := part.(map[string]interface{})
			if !ok {
				continue
			}
			inlineData, _ := pm["inlineData"].(map[string]interface{})
			if inlineData != nil {
				if data, ok := inlineData["data"].(string); ok {
					images = append(images, image.ImageData{B64JSON: data})
				}
			}
		}

		if len(images) == 0 {
			return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{{B64JSON: "", RevisedPrompt: prompt}}}, nil
		}
		return &image.ImageResponse{Created: image.NowSec(), Data: images}, nil
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{{B64JSON: "", RevisedPrompt: prompt}}}, nil
}

func init() {
	image.RegisterAdapter("gemini", newGeminiAdapter())
}
