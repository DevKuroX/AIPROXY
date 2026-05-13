// ref: _ref/9router/open-sse/handlers/ttsProviders/openrouter.js
package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func init() {
	Register("openrouter", &OpenRouterProvider{})
}

// OpenRouterProvider implements TTS for OpenRouter.
// ref: _ref/9router/open-sse/handlers/ttsProviders/openrouter.js
type OpenRouterProvider struct{}

func (p *OpenRouterProvider) Synthesize(ctx context.Context, text string, model string, creds *Credentials) (*TTSResult, error) {
	if creds.APIKey == "" {
		return nil, fmt.Errorf("no OpenRouter API key configured")
	}

	ttsModel := "openai/gpt-4o-mini-tts"
	voice := "alloy"

	if model != "" && strings.Contains(model, "/") {
		lastSlash := strings.LastIndex(model, "/")
		maybeModel := model[:lastSlash]
		maybeVoice := model[lastSlash+1:]
		if strings.Contains(maybeModel, "/") {
			ttsModel = maybeModel
			voice = maybeVoice
		} else {
			voice = model
		}
	} else if model != "" {
		voice = model
	}

	reqBody := map[string]interface{}{
		"model":      ttsModel,
		"modalities": []string{"text", "audio"},
		"audio":      map[string]string{"voice": voice, "format": "wav"},
		"stream":     true,
		"messages": []map[string]string{
			{"role": "user", "content": text},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creds.APIKey)
	req.Header.Set("HTTP-Referer", "https://endpoint-proxy.local")
	req.Header.Set("X-Title", "Endpoint Proxy")

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
		return nil, fmt.Errorf("OpenRouter TTS failed: %d", res.StatusCode)
	}

	var chunks []string
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") || line == "data: [DONE]" {
			continue
		}
		var jsonResp map[string]interface{}
		if err := json.Unmarshal([]byte(line[6:]), &jsonResp); err != nil {
			continue
		}
		choices, _ := jsonResp["choices"].([]interface{})
		if len(choices) > 0 {
			choice, _ := choices[0].(map[string]interface{})
			delta, _ := choice["delta"].(map[string]interface{})
			audio, _ := delta["audio"].(map[string]interface{})
			if data, ok := audio["data"].(string); ok {
				chunks = append(chunks, data)
			}
		}
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("OpenRouter TTS returned no audio data")
	}

	return &TTSResult{
		Base64: strings.Join(chunks, ""),
		Format: "wav",
	}, nil
}
