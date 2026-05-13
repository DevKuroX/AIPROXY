// Package tts provides TTS handler and provider implementations.
// ref: _ref/9router/open-sse/handlers/ttsCore.js
package tts

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/handlers/tts/providers"
)

// ProviderConfig holds configuration for a TTS provider.
// ref: _ref/9router/open-sse/handlers/ttsProviders/index.js:13-21
type ProviderConfig struct {
	Name        string
	Provider    string
	BaseURL     string
	APIKey      string
	AccessToken string
}

// SynthesizeResult represents the result of TTS synthesis.
type SynthesizeResult struct {
	Audio  []byte
	Format string
}

// CreateTTSError creates an error result.
type TTSError struct {
	Status  int
	Message string
}

func (e *TTSError) Error() string {
	return e.Message
}

var specialAdapters = map[string]providers.Provider{}

func RegisterAdapter(name string, adapter providers.Provider) {
	specialAdapters[name] = adapter
}

func GetAdapter(provider string) providers.Provider {
	return specialAdapters[provider]
}

// Synthesize synthesizes text to speech.
// ref: _ref/9router/open-sse/handlers/ttsCore.js:51-70
func Synthesize(ctx context.Context, config *ProviderConfig, input, model, voice, responseFormat string) (*SynthesizeResult, *TTSError) {
	if strings.TrimSpace(input) == "" {
		return nil, &TTSError{Status: http.StatusBadRequest, Message: "Missing required field: input"}
	}

	adapter := GetAdapter(config.Provider)
	if adapter != nil {
		creds := &providers.Credentials{
			APIKey:      config.APIKey,
			AccessToken: config.AccessToken,
			BaseURL:     config.BaseURL,
		}
		result, err := adapter.Synthesize(ctx, strings.TrimSpace(input), model, creds)
		if err != nil {
			return nil, &TTSError{Status: http.StatusBadGateway, Message: err.Error()}
		}
		return createTTSResponse(result, responseFormat)
	}

	return nil, &TTSError{
		Status:  http.StatusBadRequest,
		Message: fmt.Sprintf("Provider '%s' does not support TTS via this route", config.Provider),
	}
}

func createTTSResponse(result *providers.TTSResult, responseFormat string) (*SynthesizeResult, *TTSError) {
	audio, err := base64.StdEncoding.DecodeString(result.Base64)
	if err != nil {
		return nil, &TTSError{Status: http.StatusInternalServerError, Message: "Failed to decode audio"}
	}
	return &SynthesizeResult{
		Audio:  audio,
		Format: result.Format,
	}, nil
}
