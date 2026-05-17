package image

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	httpClient *http.Client
	logger     Logger
}

func NewHandler(logger Logger) *Handler {
	return &Handler{
		httpClient: &http.Client{},
		logger:     logger,
	}
}

type GenerateOptions struct {
	Body       *ImageRequest
	ModelInfo  ModelInfo
	Credentials *ImageCredentials
	BinaryOutput bool
	OnRequestSuccess func()
}

type ModelInfo struct {
	Provider string
	Model    string
}

type Result struct {
	Success  bool
	Response *http.Response
	Status   int
	Error    string
}

func (h *Handler) Generate(ctx context.Context, opts *GenerateOptions) *Result {
	provider := opts.ModelInfo.Provider
	model := opts.ModelInfo.Model

	if opts.Body.Prompt == "" {
		return &Result{Success: false, Status: http.StatusBadRequest, Error: "Missing required field: prompt"}
	}

	adapter := GetAdapter(provider)
	if adapter == nil {
		return &Result{
			Success: false,
			Status:  http.StatusBadRequest,
			Error:   fmt.Sprintf("Provider '%s' does not support image generation", provider),
		}
	}

	url, err := adapter.BuildURL(model, opts.Credentials)
	if err != nil {
		return &Result{Success: false, Status: http.StatusBadRequest, Error: err.Error()}
	}

	requestBody, err := adapter.BuildBody(model, opts.Body)
	if err != nil {
		return &Result{Success: false, Status: http.StatusBadRequest, Error: err.Error()}
	}

	headers, err := adapter.BuildHeaders(opts.Credentials, requestBody, model, opts.Body)
	if err != nil {
		return &Result{Success: false, Status: http.StatusBadRequest, Error: err.Error()}
	}

	if h.logger != nil {
		prompt := opts.Body.Prompt
		if len(prompt) > 50 {
			prompt = prompt[:50] + "..."
		}
		h.logger.Debug("IMAGE", fmt.Sprintf("%s | %s | prompt=\"%s\"", strings.ToUpper(provider), model, prompt))
	}

	var bodyReader io.Reader
	switch v := requestBody.(type) {
	case string:
		bodyReader = strings.NewReader(v)
	case []byte:
		bodyReader = bytes.NewReader(v)
	default:
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return &Result{Success: false, Status: http.StatusBadRequest, Error: "Failed to serialize request body"}
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
	if err != nil {
		return &Result{Success: false, Status: http.StatusBadGateway, Error: fmt.Sprintf("Failed to create request: %v", err)}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return &Result{Success: false, Status: http.StatusBadGateway, Error: fmt.Sprintf("Request failed: %v", err)}
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		if !adapter.NoAuth() {
			if h.logger != nil {
				h.logger.Warn("TOKEN", fmt.Sprintf("%s | auth failed for image generation", strings.ToUpper(provider)))
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := h.parseErrorResponse(resp)
		return &Result{Success: false, Status: resp.StatusCode, Error: errMsg}
	}

	var parsed interface{}
	if parseResp, err := adapter.ParseResponse(ctx, resp, &ParseResponseOptions{
		Headers:         headers,
		Log:             h.logger,
		OnRequestSuccess: opts.OnRequestSuccess,
		URL:             url,
		RequestBody:     requestBody,
		Model:           model,
		Body:            opts.Body,
	}); err != nil {
		return &Result{Success: false, Status: http.StatusBadGateway, Error: err.Error()}
	} else if parseResp != nil {
		parsed = parseResp
	} else {
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return &Result{Success: false, Status: http.StatusBadGateway, Error: "Failed to read response"}
		}
		if err := json.Unmarshal(buf, &parsed); err != nil {
			return &Result{Success: false, Status: http.StatusBadGateway, Error: fmt.Sprintf("Invalid response from %s", provider)}
		}
	}

	if opts.OnRequestSuccess != nil {
		opts.OnRequestSuccess()
	}

	normalized, err := adapter.Normalize(parsed, opts.Body.Prompt)
	if err != nil {
		return &Result{Success: false, Status: http.StatusBadGateway, Error: err.Error()}
	}

	if normalized == nil || normalized.Created == 0 || normalized.Data == nil {
		if m, ok := parsed.(map[string]interface{}); ok {
			normalized = h.parseImageResponse(m)
		}
	}

	if opts.BinaryOutput {
		var b64 string
		if len(normalized.Data) > 0 {
			b64 = normalized.Data[0].B64JSON
			if b64 == "" && normalized.Data[0].URL != "" {
				var err error
				b64, err = URLToBase64(ctx, normalized.Data[0].URL)
				if err != nil && h.logger != nil {
					h.logger.Debug("IMAGE", fmt.Sprintf("Failed to fetch image URL: %v", err))
				}
			}
		}
		if b64 != "" {
			buf, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return &Result{Success: false, Status: http.StatusBadGateway, Error: "Failed to decode image"}
			}
			outputFormat := opts.Body.OutputFormat
			if outputFormat == "" {
				outputFormat = "png"
			}
			outputFormat = strings.ToLower(outputFormat)

			mime := "image/png"
			ext := outputFormat
			switch outputFormat {
			case "jpeg", "jpg":
				mime = "image/jpeg"
				ext = "jpg"
			case "webp":
				mime = "image/webp"
			}

			return &Result{
				Success: true,
				Response: &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type":              []string{mime},
				"Content-Disposition":       []string{fmt.Sprintf("inline; filename=\"image.%s\"", ext)},
					},
					Body: io.NopCloser(bytes.NewReader(buf)),
				},
			}
		}
	}

	respBody, err := json.Marshal(normalized)
	if err != nil {
		return &Result{Success: false, Status: http.StatusInternalServerError, Error: "Failed to serialize response"}
	}

	return &Result{
		Success: true,
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewReader(respBody)),
		},
	}
}

func (h *Handler) parseErrorResponse(resp *http.Response) string {
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	var errData struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(buf, &errData); err == nil {
		if errData.Error.Message != "" {
			return errData.Error.Message
		}
		if errData.Message != "" {
			return errData.Message
		}
	}

	return string(buf)
}

func (h *Handler) parseImageResponse(m map[string]interface{}) *ImageResponse {
	created := int64(0)
	if c, ok := m["created"].(float64); ok {
		created = int64(c)
	}

	var data []ImageData
	if d, ok := m["data"].([]interface{}); ok {
		for _, item := range d {
			if dm, ok := item.(map[string]interface{}); ok {
				img := ImageData{}
				if url, ok := dm["url"].(string); ok {
					img.URL = url
				}
				if b64, ok := dm["b64_json"].(string); ok {
					img.B64JSON = b64
				}
				if rp, ok := dm["revised_prompt"].(string); ok {
					img.RevisedPrompt = rp
				}
				data = append(data, img)
			}
		}
	}

	return &ImageResponse{Created: created, Data: data}
}
