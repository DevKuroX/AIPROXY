package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
	"strings"
)

type runwaymlAdapter struct {
	image.BaseAdapter
}

func newRunwaymlAdapter() *runwaymlAdapter {
	return &runwaymlAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     true,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func (a *runwaymlAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	baseURL := "https://api.dev.runwayml.com/v1"
	if strings.Contains(model, "image") {
		return baseURL + "/text_to_image", nil
	}
	return baseURL + "/image_to_video", nil
}

func (a *runwaymlAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	key := creds.APIKey
	if key == "" {
		key = creds.AccessToken
	}
	return map[string]string{
		"Content-Type":    "application/json",
		"Authorization":   "Bearer " + key,
		"X-Runway-Version": "2024-11-06",
	}, nil
}

func (a *runwaymlAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	ratio := image.SizeToAspectRatio(body.Size)
	isVideo := !strings.Contains(model, "image")

	if isVideo {
		req := map[string]interface{}{
			"promptText": body.Prompt,
			"model":      model,
			"ratio":      ratio,
			"duration":   5,
		}
		if body.Image != "" {
			req["promptImage"] = body.Image
		}
		return req, nil
	}

	req := map[string]interface{}{
		"promptText": body.Prompt,
		"model":      model,
		"ratio":      ratio,
	}
	if body.Image != "" {
		req["referenceImages"] = []map[string]string{{"uri": body.Image}}
	}
	return req, nil
}

func (a *runwaymlAdapter) ParseResponse(ctx context.Context, response *http.Response, opts *image.ParseResponseOptions) (*image.ImageResponse, error) {
	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if data.ID == "" {
		return nil, fmt.Errorf("Runway: no task id returned")
	}

	taskURL := "https://api.dev.runwayml.com/v1/tasks/" + data.ID
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

		taskReq, err := http.NewRequestWithContext(ctx, "GET", taskURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create task request: %w", err)
		}
		for k, v := range headers {
			taskReq.Header.Set(k, v)
		}

		taskResp, err := http.DefaultClient.Do(taskReq)
		if err != nil {
			return nil, fmt.Errorf("task request failed: %w", err)
		}

		if taskResp.StatusCode != http.StatusOK {
			taskResp.Body.Close()
			return nil, fmt.Errorf("Runway status %d", taskResp.StatusCode)
		}

		taskBuf, err := io.ReadAll(taskResp.Body)
		taskResp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read task response: %w", err)
		}

		var taskData struct {
			Status  string   `json:"status"`
			Failure string   `json:"failure"`
			Output  []string `json:"output"`
		}
		if err := json.Unmarshal(taskBuf, &taskData); err != nil {
			return nil, fmt.Errorf("failed to parse task response: %w", err)
		}

		if taskData.Status == "SUCCEEDED" {
			var images []image.ImageData
			for _, url := range taskData.Output {
				images = append(images, image.ImageData{URL: url})
			}
			return &image.ImageResponse{Created: image.NowSec(), Data: images}, nil
		}

		if taskData.Status == "FAILED" || taskData.Status == "CANCELLED" {
			msg := taskData.Failure
			if msg == "" {
				msg = "Runway task failed"
			}
			return nil, fmt.Errorf("%s", msg)
		}
	}

	return nil, fmt.Errorf("Runway polling timeout")
}

func (a *runwaymlAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	case map[string]interface{}:
		var images []image.ImageData
		if output, ok := v["output"].([]interface{}); ok {
			for _, url := range output {
				if u, ok := url.(string); ok {
					images = append(images, image.ImageData{URL: u})
				}
			}
		}
		return &image.ImageResponse{Created: image.NowSec(), Data: images}, nil
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("runwayml", newRunwaymlAdapter())
}
