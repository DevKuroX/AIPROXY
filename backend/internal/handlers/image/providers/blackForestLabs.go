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

type blackForestLabsAdapter struct {
	image.BaseAdapter
}

func newBlackForestLabsAdapter() *blackForestLabsAdapter {
	return &blackForestLabsAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     true,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func (a *blackForestLabsAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return "https://api.bfl.ai/v1/" + model, nil
}

func (a *blackForestLabsAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	key := creds.APIKey
	if key == "" {
		key = creds.AccessToken
	}
	return map[string]string{
		"Content-Type": "application/json",
		"x-key":        key,
	}, nil
}

func (a *blackForestLabsAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	req := map[string]interface{}{"prompt": body.Prompt}
	if body.Size != "" {
		w, h := image.ParseSizeDimensions(body.Size)
		if w > 0 {
			req["width"] = w
		}
		if h > 0 {
			req["height"] = h
		}
	}
	if body.Image != "" {
		req["image_prompt"] = body.Image
	}
	return req, nil
}

func (a *blackForestLabsAdapter) ParseResponse(ctx context.Context, response *http.Response, opts *image.ParseResponseOptions) (*image.ImageResponse, error) {
	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data struct {
		PollingURL string `json:"polling_url"`
	}
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if data.PollingURL == "" {
		return nil, fmt.Errorf("BFL: no polling_url returned")
	}

	deadline := time.Now().Add(time.Duration(image.PollTimeoutMs) * time.Millisecond)
	key := ""
	if opts != nil && opts.Headers != nil {
		key = opts.Headers["x-key"]
	}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		time.Sleep(time.Duration(image.PollIntervalMs) * time.Millisecond)

		pollReq, err := http.NewRequestWithContext(ctx, "GET", data.PollingURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create poll request: %w", err)
		}
		pollReq.Header.Set("x-key", key)
		pollReq.Header.Set("Accept", "application/json")

		pollResp, err := http.DefaultClient.Do(pollReq)
		if err != nil {
			return nil, fmt.Errorf("poll request failed: %w", err)
		}

		if pollResp.StatusCode != http.StatusOK {
			pollResp.Body.Close()
			return nil, fmt.Errorf("BFL status %d", pollResp.StatusCode)
		}

		pollBuf, err := io.ReadAll(pollResp.Body)
		pollResp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read poll response: %w", err)
		}

		var status struct {
			Status string `json:"status"`
			Error  string `json:"error"`
			Result struct {
				Sample string `json:"sample"`
			} `json:"result"`
		}
		if err := json.Unmarshal(pollBuf, &status); err != nil {
			return nil, fmt.Errorf("failed to parse poll response: %w", err)
		}

		if status.Status == "Ready" {
			if status.Result.Sample != "" {
				return &image.ImageResponse{
					Created: image.NowSec(),
					Data:    []image.ImageData{{URL: status.Result.Sample}},
				}, nil
			}
			return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
		}

		if status.Status == "Error" || status.Status == "Failed" {
			msg := status.Error
			if msg == "" {
				msg = "BFL generation failed"
			}
			return nil, fmt.Errorf("%s", msg)
		}
	}

	return nil, fmt.Errorf("BFL polling timeout")
}

func (a *blackForestLabsAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	case map[string]interface{}:
		if result, ok := v["result"].(map[string]interface{}); ok {
			if sample, ok := result["sample"].(string); ok {
				return &image.ImageResponse{
					Created: image.NowSec(),
					Data:    []image.ImageData{{URL: sample}},
				}, nil
			}
		}
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("black-forest-labs", newBlackForestLabsAdapter())
}
