package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
)

type nanobananaAdapter struct {
	image.BaseAdapter
}

func newNanobananaAdapter() *nanobananaAdapter {
	return &nanobananaAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     true,
			Streaming: false,
			SkipAuth:  false,
		},
	}
}

func (a *nanobananaAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return "https://api.nanobananaapi.ai/api/v1/nanobanana/generate", nil
}

func (a *nanobananaAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
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

func (a *nanobananaAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	ratio := image.SizeToAspectRatio(body.Size)

	isEdit := body.Image != "" || (len(body.Images) > 0)
	genType := "TEXTTOIAMGE"
	if isEdit {
		genType = "IMAGETOIAMGE"
	}

	n := body.N
	if n == 0 {
		n = 1
	}

	req := map[string]interface{}{
		"prompt":      body.Prompt,
		"type":        genType,
		"numImages":   n,
		"image_size":  ratio,
		"callBackUrl": "https://localhost/callback",
	}

	if isEdit {
		var urls []string
		for _, img := range body.Images {
			if img != "" {
				urls = append(urls, img)
			}
		}
		if body.Image != "" {
			urls = append(urls, body.Image)
		}
		req["imageUrls"] = urls
	}

	return req, nil
}

func (a *nanobananaAdapter) ParseResponse(ctx context.Context, response *http.Response, opts *image.ParseResponseOptions) (*image.ImageResponse, error) {
	buf, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var submitData struct {
		Code int `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			TaskID string `json:"taskId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &submitData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if submitData.Code != 200 {
		msg := submitData.Msg
		if msg == "" {
			msg = "NanoBanana submit failed"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	taskID := submitData.Data.TaskID
	if taskID == "" {
		return nil, fmt.Errorf("NanoBanana: no taskId returned")
	}

	pollURL := "https://api.nanobananaapi.ai/api/v1/nanobanana/record-info?taskId=" + url.QueryEscape(taskID)
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

		pollReq, err := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create poll request: %w", err)
		}
		for k, v := range headers {
			pollReq.Header.Set(k, v)
		}

		pollResp, err := http.DefaultClient.Do(pollReq)
		if err != nil {
			return nil, fmt.Errorf("poll request failed: %w", err)
		}

		if pollResp.StatusCode != http.StatusOK {
			pollResp.Body.Close()
			return nil, fmt.Errorf("NanoBanana status %d", pollResp.StatusCode)
		}

		pollBuf, err := io.ReadAll(pollResp.Body)
		pollResp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read poll response: %w", err)
		}

		var pollData struct {
			Data struct {
				SuccessFlag  int    `json:"successFlag"`
				ErrorMessage string `json:"errorMessage"`
				Response     struct {
					ResultImageUrl  string `json:"resultImageUrl"`
					OriginImageUrl  string `json:"originImageUrl"`
				} `json:"response"`
			} `json:"data"`
		}
		if err := json.Unmarshal(pollBuf, &pollData); err != nil {
			return nil, fmt.Errorf("failed to parse poll response: %w", err)
		}

		flag := pollData.Data.SuccessFlag
		if flag == 1 {
			imgURL := pollData.Data.Response.ResultImageUrl
			if imgURL == "" {
				imgURL = pollData.Data.Response.OriginImageUrl
			}
			if imgURL != "" {
				return &image.ImageResponse{
					Created: image.NowSec(),
					Data:    []image.ImageData{{URL: imgURL, RevisedPrompt: ""}},
				}, nil
			}
			return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
		}

		if flag == 2 || flag == 3 {
			msg := pollData.Data.ErrorMessage
			if msg == "" {
				msg = "NanoBanana generation failed"
			}
			return nil, fmt.Errorf("%s", msg)
		}
	}

	return nil, fmt.Errorf("NanoBanana polling timeout")
}

func (a *nanobananaAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	case map[string]interface{}:
		if resp, ok := v["response"].(map[string]interface{}); ok {
			imgURL, _ := resp["resultImageUrl"].(string)
			if imgURL == "" {
				imgURL, _ = resp["originImageUrl"].(string)
			}
			if imgURL != "" {
				return &image.ImageResponse{
					Created: image.NowSec(),
					Data:    []image.ImageData{{URL: imgURL, RevisedPrompt: prompt}},
				}, nil
			}
		}
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("nanobanana", newNanobananaAdapter())
}
