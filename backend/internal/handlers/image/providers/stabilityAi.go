package providers

import (
	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
	"strings"
)

type stabilityAdapter struct {
	image.BaseAdapter
}

func newStabilityAdapter() *stabilityAdapter {
	return &stabilityAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     false,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func modelToEndpoint(model string) string {
	if strings.Contains(model, "ultra") {
		return "ultra"
	}
	if strings.Contains(model, "sd3") {
		return "sd3"
	}
	return "core"
}

func (a *stabilityAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return "https://api.stability.ai/v2beta/stable-image/generate/" + modelToEndpoint(model), nil
}

func (a *stabilityAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	key := creds.APIKey
	if key == "" {
		key = creds.AccessToken
	}
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + key,
		"Accept":        "application/json",
	}, nil
}

func (a *stabilityAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	outputFormat := body.OutputFormat
	if outputFormat == "" {
		outputFormat = "png"
	}
	req := map[string]interface{}{
		"prompt":        body.Prompt,
		"output_format": strings.ToLower(outputFormat),
	}
	if body.Size != "" {
		req["aspect_ratio"] = image.SizeToAspectRatio(body.Size)
	}
	if body.Style != "" {
		req["style_preset"] = body.Style
	}
	if strings.Contains(model, "sd3") {
		req["model"] = model
	}
	return req, nil
}

func (a *stabilityAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case map[string]interface{}:
		if img, ok := v["image"].(string); ok {
			return &image.ImageResponse{
				Created: image.NowSec(),
				Data:    []image.ImageData{{B64JSON: img}},
			}, nil
		}
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("stability-ai", newStabilityAdapter())
}
