package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ref: open-sse/handlers/ttsProviders/index.js:13-21
type TTSProviderConfig struct {
	Name     string
	BaseURL  string
	APIKey   string
	Provider string
}

// ref: open-sse/handlers/ttsCore.js:62-63
type TTSResult struct {
	Audio  []byte
	Format string
}

// ref: open-sse/handlers/ttsProviders/index.js:23-25
type TTSAdapter interface {
	Synthesize(ctx context.Context, text string, model string, voice string, format string, config *TTSProviderConfig) (*TTSResult, error)
}

// ref: open-sse/handlers/ttsProviders/index.js:13-21
var ttsAdapters = make(map[string]TTSAdapter)

func RegisterTTSAdapter(provider string, adapter TTSAdapter) {
	ttsAdapters[provider] = adapter
}

func GetTTSAdapter(provider string) TTSAdapter {
	return ttsAdapters[provider]
}

// ref: open-sse/handlers/ttsCore.js:58-64
func SynthesizeTTS(ctx context.Context, config *TTSProviderConfig, input string, model string, voice string, format string) (*TTSResult, error) {
	adapter := GetTTSAdapter(config.Provider)
	if adapter != nil {
		return adapter.Synthesize(ctx, strings.TrimSpace(input), model, voice, format, config)
	}
	return nil, fmt.Errorf("provider '%s' does not support TTS via this route", config.Provider)
}

// ref: open-sse/handlers/ttsProviders/openai.js:1-30
type OpenAITTSAdapter struct{}

func init() {
	RegisterTTSAdapter("openai", &OpenAITTSAdapter{})
	RegisterTTSAdapter("0penai", &OpenAITTSAdapter{})
}

func (a *OpenAITTSAdapter) Synthesize(ctx context.Context, text string, model string, voice string, format string, config *TTSProviderConfig) (*TTSResult, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("no OpenAI API key configured")
	}

	ttsModel := model
	ttsVoice := voice

	if ttsModel == "" {
		ttsModel = "gpt-4o-mini-tts"
	}
	if ttsVoice == "" {
		ttsVoice = "alloy"
	}

	if strings.Contains(ttsModel, "/") {
		parts := strings.Split(ttsModel, "/")
		if len(parts) == 2 {
			ttsModel = parts[0]
			ttsVoice = parts[1]
		}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	reqBody := map[string]interface{}{
		"model": ttsModel,
		"voice": ttsVoice,
		"input": text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/audio/speech", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("OpenAI TTS failed: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI TTS failed: %d", resp.StatusCode)
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio response: %w", err)
	}

	audioFormat := "mp3"
	if format != "" {
		audioFormat = format
	}

	return &TTSResult{
		Audio:  audioData,
		Format: audioFormat,
	}, nil
}
