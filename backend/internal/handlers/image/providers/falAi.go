package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
)

type falAiAdapter struct {
	image.BaseAdapter
}

func newFalAiAdapter() *falAiAdapter {
	return &falAiAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     true,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func (a *falAiAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return "https://queue.fal.run/" + model, nil
}

func (a *falAiAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	key := creds.APIKey
	if key == "" {
		key = creds.AccessToken
	}
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Key " + key,
	}, nil
}

func (a *falAiAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	n := body.N
	if n == 0 {
		n = 1
	}
	req := map[string]interface{}{
		"prompt":     body.Prompt,
		"num_images": n,
	}
	if body.Size != "" {
		req["image_size"] = image.SizeToAspectRatio(body.Size)
	}
	if body.Image != "" {
		req["image_url"] = body.Image
	}
	return req, nil
}

func (a *falAiAdapter) ParseResponse(ctx context.Context, response *http.Response, opts *image.ParseResponseOptions) (*image.ImageResponse, error) {
	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data struct {
		StatusURL   string `json:"status_url"`
		ResponseURL string `json:"response_url"`
	}
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if data.StatusURL == "" {
		return nil, fmt.Errorf("Fal: no status_url returned")
	}

	deadline := time.Now().Add(time.Duration(image.PollTimeoutMs) * time.Millisecond)
	headers := map[string]string{}
	if opts != nil && opts.Headers != nil {
		for k, v := range opts.Headers {
			headers[k] = v
		}
	}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		time.Sleep(time.Duration(image.PollIntervalMs) * time.Millisecond)

		statusReq, err := http.NewRequestWithContext(ctx, "GET", data.StatusURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create status request: %w", err)
		}
		for k, v := range headers {
			statusReq.Header.Set(k, v)
		}

		statusResp, err := http.DefaultClient.Do(statusReq)
		if err != nil {
			return nil, fmt.Errorf("status request failed: %w", err)
		}

		if statusResp.StatusCode != http.StatusOK {
			statusResp.Body.Close()
			return nil, fmt.Errorf("Fal status %d", statusResp.StatusCode)
		}

		statusBuf, err := io.ReadAll(statusResp.Body)
		statusResp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read status response: %w", err)
		}

		var status struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal(statusBuf, &status); err != nil {
			return nil, fmt.Errorf("failed to parse status response: %w", err)
		}

		if status.Status == "COMPLETED" {
			if data.ResponseURL == "" {
				return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
			}

			respReq, err := http.NewRequestWithContext(ctx, "GET", data.ResponseURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create response request: %w", err)
			}
			for k, v := range headers {
				respReq.Header.Set(k, v)
			}

			respResp, err := http.DefaultClient.Do(respReq)
			if err != nil {
				return nil, fmt.Errorf("response request failed: %w", err)
			}
			defer respResp.Body.Close()

			respBuf, err := io.ReadAll(respResp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read final response: %w", err)
			}

			var finalData map[string]interface{}
			if err := json.Unmarshal(respBuf, &finalData); err != nil {
				return nil, fmt.Errorf("failed to parse final response: %w", err)
			}
			return a.Normalize(finalData, "")
		}

		if status.Status == "FAILED" {
			msg := status.Error
			if msg == "" {
				msg = "Fal generation failed"
			}
			return nil, fmt.Errorf("%s", msg)
		}
	}

	return nil, fmt.Errorf("Fal polling timeout")
}

func (a *falAiAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	case map[string]interface{}:
		var images []image.ImageData

		if imgs, ok := v["images"].([]interface{}); ok {
			for _, img := range imgs {
				switch iv := img.(type) {
				case string:
					images = append(images, image.ImageData{URL: iv})
				case map[string]interface{}:
					if url, ok := iv["url"].(string); ok {
						images = append(images, image.ImageData{URL: url})
					}
				}
			}
		} else if img, ok := v["image"].(interface{}); ok {
			switch iv := img.(type) {
			case string:
				images = append(images, image.ImageData{URL: iv})
			case map[string]interface{}:
				if url, ok := iv["url"].(string); ok {
					images = append(images, image.ImageData{URL: url})
				}
			}
		}

		return &image.ImageResponse{Created: image.NowSec(), Data: images}, nil
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("fal-ai", newFalAiAdapter())
}
