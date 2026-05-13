// ref: _ref/9router/open-sse/handlers/ttsProviders/elevenlabs.js
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
	"sync"
	"time"
)

func init() {
	Register("elevenlabs", &ElevenLabsProvider{})
}

// ElevenLabsProvider implements TTS for ElevenLabs.
// ref: _ref/9router/open-sse/handlers/ttsProviders/elevenlabs.js
type ElevenLabsProvider struct {
	mu          sync.Mutex
	voicesCache map[string]cachedVoices
}

type cachedVoices struct {
	voices []Voice
	time   time.Time
}

const voicesTTL11 = 24 * time.Hour

func (p *ElevenLabsProvider) Synthesize(ctx context.Context, text string, model string, creds *Credentials) (*TTSResult, error) {
	if creds.APIKey == "" {
		return nil, fmt.Errorf("ElevenLabs API key required")
	}

	modelId := "eleven_flash_v2_5"
	voiceId := model

	if model != "" && strings.Contains(model, "/") {
		parts := strings.Split(model, "/")
		if len(parts) == 2 {
			modelId = parts[0]
			voiceId = parts[1]
		}
	}

	reqBody := map[string]interface{}{
		"text":    text,
		"model_id": modelId,
		"voice_settings": map[string]float64{
			"stability":       0.5,
			"similarity_boost": 0.75,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceId)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", creds.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if detail, ok := errResp["detail"].(map[string]interface{}); ok {
				if msg, ok := detail["message"].(string); ok {
					return nil, fmt.Errorf("%s", msg)
				}
			}
		}
		return nil, fmt.Errorf("ElevenLabs TTS failed: %d", res.StatusCode)
	}

	audio, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if len(audio) < 1024 {
		return nil, fmt.Errorf("ElevenLabs TTS returned empty audio")
	}

	return &TTSResult{
		Base64: base64.StdEncoding.EncodeToString(audio),
		Format: "mp3",
	}, nil
}

func (p *ElevenLabsProvider) FetchVoices(ctx context.Context, creds *Credentials) ([]Voice, error) {
	if creds.APIKey == "" {
		return nil, fmt.Errorf("ElevenLabs API key required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if p.voicesCache == nil {
		p.voicesCache = make(map[string]cachedVoices)
	}

	if cached, ok := p.voicesCache[creds.APIKey]; ok && now.Sub(cached.time) < voicesTTL11 {
		return cached.voices, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.elevenlabs.io/v1/voices", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("xi-api-key", creds.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ElevenLabs voices fetch failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ElevenLabs voices fetch failed: %d", res.StatusCode)
	}

	var resp struct {
		Voices []struct {
			ID     string            `json:"voice_id"`
			Name   string            `json:"name"`
			Labels map[string]string `json:"labels"`
		} `json:"voices"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	voices := make([]Voice, len(resp.Voices))
	for i, v := range resp.Voices {
		lang := "en"
		if v.Labels != nil {
			if l, ok := v.Labels["language"]; ok {
				lang = l
			}
		}
		voices[i] = Voice{
			ID:   v.ID,
			Name: v.Name,
			Lang: lang,
		}
	}

	p.voicesCache[creds.APIKey] = cachedVoices{voices: voices, time: now}
	return voices, nil
}
