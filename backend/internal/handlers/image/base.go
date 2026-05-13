// Package image provides image generation handlers for various providers.
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js
package image

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Polling constants
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js:3-4
const (
	PollIntervalMs = 1500
	PollTimeoutMs  = 120000
)

// Sleep helper for polling
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js:6
func sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// NowSec returns current Unix timestamp in seconds.
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js:29-31
func NowSec() int64 {
	return time.Now().Unix()
}

// SizeToAspectRatio maps OpenAI size to provider-specific aspect ratio.
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js:9-19
func SizeToAspectRatio(size string) string {
	if size == "" {
		return "1:1"
	}
	aspectMap := map[string]string{
		"1024x1024":  "1:1",
		"1024x1792":  "9:16",
		"1792x1024":  "16:9",
		"1024x1536":  "2:3",
		"1536x1024":  "3:2",
	}
	if ratio, ok := aspectMap[size]; ok {
		return ratio
	}
	return "1:1"
}

// URLToBase64 fetches a URL and returns base64 encoded content.
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js:22-27
func URLToBase64(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

// ImageCredentials holds authentication credentials for image providers.
// ref: _ref/9router/open-sse/handlers/imageGenerationCore.js:31-32
type ImageCredentials struct {
	APIKey               string
	AccessToken          string
	IDToken              string
	ProviderSpecificData map[string]interface{}
}

// ImageRequest represents the normalized image generation request.
// ref: _ref/9router/open-sse/handlers/imageGenerationCore.js:19
type ImageRequest struct {
	Prompt         string                 `json:"prompt"`
	Model          string                 `json:"model,omitempty"`
	N              int                    `json:"n,omitempty"`
	Size           string                 `json:"size,omitempty"`
	Quality        string                 `json:"quality,omitempty"`
	Style          string                 `json:"style,omitempty"`
	ResponseFormat string                 `json:"response_format,omitempty"`
	OutputFormat   string                 `json:"output_format,omitempty"`
	Image          string                 `json:"image,omitempty"`
	Images         []string               `json:"images,omitempty"`
	MaskImage      string                 `json:"mask_image,omitempty"`
	ImageDetail    string                 `json:"image_detail,omitempty"`
	Background     string                 `json:"background,omitempty"`
	Width          int                    `json:"width,omitempty"`
	Height         int                    `json:"height,omitempty"`
	Extra          map[string]interface{} `json:"-"` // Additional provider-specific fields
}

// GetExtraString returns a string extra field value.
func (r *ImageRequest) GetExtraString(key string) string {
	if r.Extra == nil {
		return ""
	}
	if v, ok := r.Extra[key].(string); ok {
		return v
	}
	return ""
}

// GetExtraInt returns an int extra field value.
func (r *ImageRequest) GetExtraInt(key string) int {
	if r.Extra == nil {
		return 0
	}
	switch v := r.Extra[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

// ImageData represents a single generated image.
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js (implicit in normalize functions)
type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageResponse represents the normalized image generation response.
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js (implicit in normalize functions)
type ImageResponse struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

// ParseResponseOptions provides context for custom response parsing.
// ref: _ref/9router/open-sse/handlers/imageGenerationCore.js:126-136
type ParseResponseOptions struct {
	Headers         map[string]string
	Log             Logger
	StreamToClient  bool
	OnRequestSuccess func()
	URL             string
	RequestBody     interface{}
	Model           string
	Body            *ImageRequest
}

// Logger interface for image handlers.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// ImageAdapter defines the interface for provider-specific image generation.
// ref: _ref/9router/open-sse/handlers/imageProviders/index.js
type ImageAdapter interface {
	// BuildURL constructs the API endpoint URL.
	BuildURL(model string, creds *ImageCredentials) (string, error)

	// BuildHeaders constructs the HTTP headers for the request.
	BuildHeaders(creds *ImageCredentials, requestBody interface{}, model string, body *ImageRequest) (map[string]string, error)

	// BuildBody constructs the request body for the provider.
	BuildBody(model string, body *ImageRequest) (interface{}, error)

	// ParseResponse parses the provider response (optional - for async/binary/SSE responses).
	// Returns nil to use default JSON parsing.
	ParseResponse(ctx context.Context, response *http.Response, opts *ParseResponseOptions) (*ImageResponse, error)

	// Normalize converts provider response to OpenAI-compatible format.
	Normalize(responseBody interface{}, prompt string) (*ImageResponse, error)

	// IsAsync returns true if the provider uses async polling.
	IsAsync() bool

	// IsStreaming returns true if the provider uses SSE streaming.
	IsStreaming() bool

	// NoAuth returns true if the provider doesn't require authentication.
	NoAuth() bool
}

// BaseAdapter provides default implementations for optional methods.
// ref: _ref/9router/open-sse/handlers/imageProviders/_base.js (implicit patterns)
type BaseAdapter struct {
	Async     bool
	Streaming bool
	SkipAuth  bool
}

// IsAsync returns whether this adapter uses async polling.
func (b *BaseAdapter) IsAsync() bool {
	return b.Async
}

// IsStreaming returns whether this adapter uses SSE streaming.
func (b *BaseAdapter) IsStreaming() bool {
	return b.Streaming
}

// NoAuth returns whether authentication is skipped.
func (b *BaseAdapter) NoAuth() bool {
	return b.SkipAuth
}

// ParseResponse provides default JSON parsing.
func (b *BaseAdapter) ParseResponse(ctx context.Context, response *http.Response, opts *ParseResponseOptions) (*ImageResponse, error) {
	return nil, nil // Use default JSON parsing
}

// adapterRegistry holds registered image adapters.
// ref: _ref/9router/open-sse/handlers/imageProviders/index.js:15-31
var adapterRegistry = make(map[string]ImageAdapter)

// RegisterAdapter registers an image adapter for a provider.
func RegisterAdapter(provider string, adapter ImageAdapter) {
	adapterRegistry[provider] = adapter
}

// GetAdapter returns the adapter for a provider.
// ref: _ref/9router/open-sse/handlers/imageProviders/index.js:33-35
func GetAdapter(provider string) ImageAdapter {
	return adapterRegistry[provider]
}

// IsImageProvider checks if a provider supports image generation.
// ref: _ref/9router/open-sse/handlers/imageProviders/index.js:37-39
func IsImageProvider(provider string) bool {
	_, ok := adapterRegistry[provider]
	return ok
}

// ParseSizeDimensions extracts width and height from size string.
func ParseSizeDimensions(size string) (width, height int) {
	if size == "" {
		return 0, 0
	}
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return 0, 0
	}
	var w, h int
	fmt.Sscanf(parts[0], "%d", &w)
	fmt.Sscanf(parts[1], "%d", &h)
	return w, h
}
