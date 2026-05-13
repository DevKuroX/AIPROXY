package providers

import (
	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
)

type sdwebuiAdapter struct {
	image.BaseAdapter
}

func newSdwebuiAdapter() *sdwebuiAdapter {
	return &sdwebuiAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     false,
			Streaming: false,
			SkipAuth:  true,
		},
	}
}

func (a *sdwebuiAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return "http://localhost:7860/sdapi/v1/txt2img", nil
}

func (a *sdwebuiAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	return map[string]string{"Content-Type": "application/json"}, nil
}

func (a *sdwebuiAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	n := body.N
	if n == 0 {
		n = 1
	}
	size := body.Size
	if size == "" {
		size = "1024x1024"
	}

	w, h := image.ParseSizeDimensions(size)
	if w == 0 {
		w = 512
	}
	if h == 0 {
		h = 512
	}

	return map[string]interface{}{
		"prompt":     body.Prompt,
		"width":      w,
		"height":     h,
		"steps":      20,
		"batch_size": n,
	}, nil
}

func (a *sdwebuiAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case map[string]interface{}:
		if imgs, ok := v["images"].([]interface{}); ok {
			var data []image.ImageData
			for _, img := range imgs {
				if b64, ok := img.(string); ok {
					data = append(data, image.ImageData{B64JSON: b64})
				}
			}
			return &image.ImageResponse{Created: image.NowSec(), Data: data}, nil
		}
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("sdwebui", newSdwebuiAdapter())
}
