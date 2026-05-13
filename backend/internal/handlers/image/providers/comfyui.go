package providers

import (
	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
)

type comfyuiAdapter struct {
	image.BaseAdapter
}

func newComfyuiAdapter() *comfyuiAdapter {
	return &comfyuiAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     false,
			Streaming: false,
			SkipAuth:  true,
		},
	}
}

func (a *comfyuiAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return "http://localhost:8188", nil
}

func (a *comfyuiAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	return map[string]string{"Content-Type": "application/json"}, nil
}

func (a *comfyuiAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	return map[string]interface{}{"prompt": body.Prompt}, nil
}

func (a *comfyuiAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	case map[string]interface{}:
		if v["created"] != nil && v["data"] != nil {
			created := int64(0)
			if c, ok := v["created"].(float64); ok {
				created = int64(c)
			}
			var data []image.ImageData
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
						data = append(data, img)
					}
				}
			}
			return &image.ImageResponse{Created: created, Data: data}, nil
		}
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("comfyui", newComfyuiAdapter())
}
