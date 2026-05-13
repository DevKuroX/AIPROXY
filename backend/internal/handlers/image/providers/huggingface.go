package providers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
)

type huggingfaceAdapter struct {
	image.BaseAdapter
}

func newHuggingfaceAdapter() *huggingfaceAdapter {
	return &huggingfaceAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     false,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func (a *huggingfaceAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return "https://api-inference.huggingface.co/models/" + model, nil
}

func (a *huggingfaceAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	headers := map[string]string{"Content-Type": "application/json"}
	key := creds.APIKey
	if key == "" {
		key = creds.AccessToken
	}
	if key != "" {
		headers["Authorization"] = "Bearer " + key
	}
	return headers, nil
}

func (a *huggingfaceAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	return map[string]interface{}{"inputs": body.Prompt}, nil
}

func (a *huggingfaceAdapter) ParseResponse(ctx context.Context, response *http.Response, opts *image.ParseResponseOptions) (*image.ImageResponse, error) {
	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(buf)
	return &image.ImageResponse{
		Created: image.NowSec(),
		Data:    []image.ImageData{{B64JSON: b64}},
	}, nil
}

func (a *huggingfaceAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("huggingface", newHuggingfaceAdapter())
}
