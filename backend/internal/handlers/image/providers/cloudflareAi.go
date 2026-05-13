package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
)

type cloudflareAiAdapter struct {
	image.BaseAdapter
}

var multipartModels = map[string]bool{
	"@cf/black-forest-labs/flux-2-dev":       true,
	"@cf/black-forest-labs/flux-2-klein-4b":  true,
	"@cf/black-forest-labs/flux-2-klein-9b":  true,
}

var optionalFields = []string{
	"negative_prompt",
	"guidance",
	"seed",
	"num_steps",
	"steps",
	"strength",
}

func newCloudflareAiAdapter() *cloudflareAiAdapter {
	return &cloudflareAiAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     false,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func (a *cloudflareAiAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	accountID := ""
	if creds.ProviderSpecificData != nil {
		if id, ok := creds.ProviderSpecificData["accountId"].(string); ok {
			accountID = id
		}
	}
	if accountID == "" {
		return "", fmt.Errorf("cloudflare-ai requires accountId in providerSpecificData")
	}
	return "https://api.cloudflare.com/client/v4/accounts/" + accountID + "/ai/run/" + model, nil
}

func (a *cloudflareAiAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	headers := map[string]string{}
	key := creds.APIKey
	if key == "" {
		key = creds.AccessToken
	}
	if key != "" {
		headers["Authorization"] = "Bearer " + key
	}

	_, isMultipart := multipartModels[model]
	if !isMultipart {
		headers["Content-Type"] = "application/json"
	}
	return headers, nil
}

func (a *cloudflareAiAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	req := map[string]interface{}{"prompt": body.Prompt}

	w, h := image.ParseSizeDimensions(body.Size)
	if body.Width > 0 {
		w = body.Width
	}
	if body.Height > 0 {
		h = body.Height
	}
	if w > 0 {
		req["width"] = w
	}
	if h > 0 {
		req["height"] = h
	}

	for _, key := range optionalFields {
		if v := body.GetExtraString(key); v != "" {
			req[key] = v
		}
		if v := body.GetExtraInt(key); v != 0 {
			req[key] = v
		}
	}

	if body.Image != "" {
		b64, err := a.resolveImageB64(body.Image)
		if err == nil && b64 != "" {
			req["image_b64"] = b64
		}
	}

	return req, nil
}

func (a *cloudflareAiAdapter) resolveImageB64(input string) (string, error) {
	if strings.HasPrefix(input, "https://") || strings.HasPrefix(input, "http://") {
		resp, err := http.Get(input)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(buf), nil
	}
	if strings.HasPrefix(input, "data:image/") {
		parts := strings.SplitN(input, ",", 2)
		if len(parts) == 2 {
			return parts[1], nil
		}
	}
	return input, nil
}

func (a *cloudflareAiAdapter) ParseResponse(ctx context.Context, response *http.Response, opts *image.ParseResponseOptions) (*image.ImageResponse, error) {
	contentType := response.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "image/") {
		buf, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read image: %w", err)
		}
		return &image.ImageResponse{
			Created: image.NowSec(),
			Data:    []image.ImageData{{B64JSON: base64.StdEncoding.EncodeToString(buf)}},
		}, nil
	}

	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var jsonResp map[string]interface{}
	if err := json.Unmarshal(buf, &jsonResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return a.Normalize(jsonResp, "")
}

func (a *cloudflareAiAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	case map[string]interface{}:
		if v["created"] != nil && v["data"] != nil {
			return parseImageResponse(v)
		}

		result := v
		if r, ok := v["result"]; ok {
			if rm, ok := r.(map[string]interface{}); ok {
				result = rm
			}
		}

		if responses, ok := result["responses"].([]interface{}); ok {
			for _, resp := range responses {
				if rm, ok := resp.(map[string]interface{}); ok {
					if success, ok := rm["success"].(bool); !ok || success {
						if res, ok := rm["result"]; ok {
							if resMap, ok := res.(map[string]interface{}); ok {
								return a.Normalize(resMap, prompt)
							}
						}
					}
				}
			}
		}

		var imageData string
		if s, ok := result["image"].(string); ok {
			imageData = s
		} else if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
			if dm, ok := data[0].(map[string]interface{}); ok {
				if b64, ok := dm["b64_json"].(string); ok {
					imageData = b64
				} else if url, ok := dm["url"].(string); ok {
					return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{{URL: url}}}, nil
				}
			}
		}

		if imageData != "" {
			item := imageItemFromString(imageData)
			if item != nil {
				return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{*item}}, nil
			}
		}

		return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func parseImageResponse(v map[string]interface{}) (*image.ImageResponse, error) {
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

func imageItemFromString(value string) *image.ImageData {
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "data:image/") {
		parts := strings.SplitN(value, ",", 2)
		if len(parts) == 2 {
			return &image.ImageData{B64JSON: parts[1]}
		}
	}
	if strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://") {
		return &image.ImageData{URL: value}
	}
	return &image.ImageData{B64JSON: value}
}

func init() {
	image.RegisterAdapter("cloudflare-ai", newCloudflareAiAdapter())
}
