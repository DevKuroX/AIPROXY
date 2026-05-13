// Package providers provides TTS provider implementations.
// ref: _ref/9router/open-sse/handlers/ttsProviders/_base.js
package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// UA is the user agent string for TTS requests.
// ref: _ref/9router/open-sse/handlers/ttsProviders/_base.js:4
const UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"

// TTSResult represents the result of a TTS synthesis.
// ref: _ref/9router/open-sse/handlers/ttsCore.js:22-25
type TTSResult struct {
	Base64 string // Base64 encoded audio data
	Format string // Audio format (mp3, wav, ogg, etc.)
}

// Credentials holds authentication credentials for TTS providers.
type Credentials struct {
	APIKey      string
	AccessToken string
	BaseURL     string
}

// Voice represents a TTS voice.
type Voice struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Lang   string `json:"lang,omitempty"`
	Locale string `json:"locale,omitempty"`
	Gender string `json:"gender,omitempty"`
}

// Provider defines the interface for TTS providers.
// ref: _ref/9router/open-sse/handlers/ttsProviders/index.js:13-21
type Provider interface {
	// Synthesize converts text to speech.
	Synthesize(ctx context.Context, text string, model string, credentials *Credentials) (*TTSResult, error)
}

// VoiceFetcher defines an interface for providers that can list available voices.
type VoiceFetcher interface {
	FetchVoices(ctx context.Context, credentials *Credentials) ([]Voice, error)
}

// ResponseToBase64 converts an HTTP response with binary audio to base64.
// ref: _ref/9router/open-sse/handlers/ttsProviders/_base.js:7-16
func ResponseToBase64(res *http.Response, defaultFormat string) (*TTSResult, error) {
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if len(buf) < 100 {
		return nil, fmt.Errorf("upstream returned empty audio")
	}

	format := defaultFormat
	ctype := res.Header.Get("Content-Type")
	if strings.Contains(ctype, "wav") {
		format = "wav"
	} else if strings.Contains(ctype, "mpeg") || strings.Contains(ctype, "mp3") {
		format = "mp3"
	} else if strings.Contains(ctype, "ogg") {
		format = "ogg"
	}

	return &TTSResult{
		Base64: base64.StdEncoding.EncodeToString(buf),
		Format: format,
	}, nil
}

// ThrowUpstreamError parses an error response from an upstream API.
// ref: _ref/9router/open-sse/handlers/ttsProviders/_base.js:18-26
func ThrowUpstreamError(res *http.Response) error {
	body, _ := io.ReadAll(res.Body)
	msg := fmt.Sprintf("Upstream error (%d)", res.StatusCode)

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if m := getStringFromNestedMap(parsed, "error", "message"); m != "" {
			msg = m
		} else if m := getStringFromMap(parsed, "message"); m != "" {
			msg = m
		} else if m := getStringFromNestedMap(parsed, "detail", "message"); m != "" {
			msg = m
		} else if detail, ok := parsed["detail"].(string); ok && detail != "" {
			msg = detail
		} else if len(body) > 0 {
			msg = string(body)
		}
	} else if len(body) > 0 {
		msg = string(body)
	}

	return fmt.Errorf("%s", msg)
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	return ""
}

func getStringFromNestedMap(m map[string]interface{}, keys ...string) string {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			if s, ok := current[key].(string); ok {
				return s
			}
			return ""
		}
		if nested, ok := current[key].(map[string]interface{}); ok {
			current = nested
		} else {
			return ""
		}
	}
	return ""
}

// ParseModelVoice parses a model string as "modelId/voiceId".
// ref: _ref/9router/open-sse/handlers/ttsProviders/_base.js:29-39
func ParseModelVoice(model, defaultModel, defaultVoice string, knownModels []string) (modelId, voiceId string) {
	if model == "" {
		return defaultModel, defaultVoice
	}

	sortedModels := make([]string, len(knownModels))
	copy(sortedModels, knownModels)
	for i := 0; i < len(sortedModels)-1; i++ {
		for j := i + 1; j < len(sortedModels); j++ {
			if len(sortedModels[j]) > len(sortedModels[i]) {
				sortedModels[i], sortedModels[j] = sortedModels[j], sortedModels[i]
			}
		}
	}

	for _, id := range sortedModels {
		if model == id {
			return id, defaultVoice
		}
		if strings.HasPrefix(model, id+"/") {
			return id, model[len(id)+1:]
		}
	}

	idx := strings.LastIndex(model, "/")
	if idx > 0 {
		return model[:idx], model[idx+1:]
	}

	if defaultModel != "" {
		return defaultModel, defaultVoice
	}
	return model, defaultVoice
}
