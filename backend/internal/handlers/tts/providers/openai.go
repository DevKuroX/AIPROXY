// ref: _ref/9router/open-sse/handlers/ttsProviders/openai.js
package providers

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

func init() {
	Register("openai", &OpenAIProvider{})
	Register("openai", &OpenAIProvider{})
}

// OpenAIProvider implements TTS for OpenAI.
// ref: _ref/9router/open-sse/handlers/ttsProviders/openai.js
type OpenAIProvider struct{}

func (p *OpenAIProvider) Synthesize(ctx context.Context, text string, model string, creds *Credentials) (*TTSResult, error) {
	if creds.APIKey == "" {
		return nil, fmt.Errorf("no OpenAI API key configured")
	}

	ttsModel := "gpt-4o-mini-tts"
	voice := "alloy"

	if model != "" && strings.Contains(model, "/") {
		parts := strings.Split(model, "/")
		if len(parts) == 2 {
			ttsModel = parts[0]
			voice = parts[1]
		}
	} else if model != "" {
		voice = model
	}

	baseURL := creds.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	reqBody := map[string]interface{}{
		"model": ttsModel,
		"voice": voice,
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
	req.Header.Set("Authorization", "Bearer "+creds.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if errMsg, ok := errResp["error"].(map[string]interface{}); ok {
				if msg, ok := errMsg["message"].(string); ok {
					return nil, fmt.Errorf("%s", msg)
				}
			}
		}
		return nil, fmt.Errorf("OpenAI TTS failed: %d", res.StatusCode)
	}

	audio, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &TTSResult{
		Base64: base64.StdEncoding.EncodeToString(audio),
		Format: "mp3",
	}, nil
}

var providerRegistry = map[string]Provider{}

func Register(name string, provider Provider) {
	providerRegistry[name] = provider
}
