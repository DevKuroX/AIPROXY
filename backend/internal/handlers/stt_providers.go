package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/router"
)

// ref: open-sse/handlers/sttCore.js:18-26
func resolveAudioContentType(filename string, mimeType string) string {
	if strings.HasPrefix(strings.ToLower(mimeType), "audio/") {
		return mimeType
	}
	name := strings.ToLower(filename)
	ext := ""
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		ext = name[idx+1:]
	}
	mimeMap := map[string]string{
		"mp3":  "audio/mpeg",
		"mp4":  "audio/mp4",
		"m4a":  "audio/mp4",
		"wav":  "audio/wav",
		"ogg":  "audio/ogg",
		"flac": "audio/flac",
		"webm": "audio/webm",
		"aac":  "audio/aac",
		"opus": "audio/opus",
	}
	if ct, ok := mimeMap[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// ref: open-sse/handlers/sttCore.js:7-16
func buildAuthHeaders(authHeader, token string) map[string]string {
	if token == "" {
		return map[string]string{}
	}
	switch authHeader {
	case "bearer":
		return map[string]string{"Authorization": "Bearer " + token}
	case "token":
		return map[string]string{"Authorization": "Token " + token}
	case "x-api-key":
		return map[string]string{"x-api-key": token}
	case "key":
		return map[string]string{"Authorization": "Key " + token}
	default:
		return map[string]string{"Authorization": "Bearer " + token}
	}
}

// ref: open-sse/handlers/sttCore.js:166-169
type STTProviderConfig struct {
	Name       string
	BaseURL    string
	APIKey     string
	AuthHeader string
	Format     string
}

// ref: open-sse/handlers/sttCore.js:156-164
type STTResult struct {
	Text string `json:"text"`
}

// ref: open-sse/handlers/sttCore.js:166-169
type STTAdapter interface {
	Transcribe(ctx context.Context, file io.Reader, filename string, mimeType string, model string, formData map[string]string, config *STTProviderConfig) (*STTResult, error)
}

// ref: open-sse/handlers/sttCore.js:183-190
var sttAdapters = make(map[string]STTAdapter)

func RegisterSTTAdapter(format string, adapter STTAdapter) {
	sttAdapters[format] = adapter
}

func GetSTTAdapter(format string) STTAdapter {
	return sttAdapters[format]
}

// ref: open-sse/handlers/sttCore.js:170-194
func TranscribeAudio(ctx context.Context, config *STTProviderConfig, file io.Reader, filename string, mimeType string, model string, formData map[string]string) (*STTResult, error) {
	adapter := GetSTTAdapter(config.Format)
	if adapter != nil {
		return adapter.Transcribe(ctx, file, filename, mimeType, model, formData, config)
	}
	return nil, fmt.Errorf("provider '%s' does not support STT via this route", config.Name)
}

// ref: open-sse/handlers/sttCore.js:141-154
type WhisperSTTAdapter struct{}

func init() {
	RegisterSTTAdapter("openai", &WhisperSTTAdapter{})
	RegisterSTTAdapter("0penai", &WhisperSTTAdapter{})
	RegisterSTTAdapter("whisper", &WhisperSTTAdapter{})
	RegisterSTTAdapter("", &WhisperSTTAdapter{})
}

// ref: open-sse/handlers/sttCore.js:141-154
func (a *WhisperSTTAdapter) Transcribe(ctx context.Context, file io.Reader, filename string, mimeType string, model string, formData map[string]string, config *STTProviderConfig) (*STTResult, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("no 0penAI API key configured")
	}

	sttModel := model
	if sttModel == "" {
		sttModel = "whisper-1"
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	err = writer.WriteField("model", sttModel)
	if err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	for _, k := range []string{"language", "prompt", "response_format", "temperature"} {
		if v, ok := formData[k]; ok && v != "" {
			err = writer.WriteField(k, v)
			if err != nil {
				return nil, fmt.Errorf("failed to write %s field: %w", k, err)
			}
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/audio/transcriptions", &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	authHeaders := buildAuthHeaders(config.AuthHeader, config.APIKey)
	for k, v := range authHeaders {
		req.Header.Set(k, v)
	}

	client := router.NewProxyClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("0penAI STT failed: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("0penAI STT failed: %d", resp.StatusCode)
	}

	var result STTResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
