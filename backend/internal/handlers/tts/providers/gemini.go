// ref: _ref/9router/open-sse/handlers/ttsProviders/gemini.js
package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func init() {
	Register("gemini", &GeminiProvider{})
}

// GeminiProvider implements TTS for Google Gemini.
// ref: _ref/9router/open-sse/handlers/ttsProviders/gemini.js
type GeminiProvider struct{}

const defaultGeminiModel = "gemini-2.5-flash-preview-tts"
const defaultGeminiVoice = "Kore"

var knownGeminiModels = []string{"gemini-2.5-flash-preview-tts", "gemini-2.5-pro-preview-tts"}

const sampleRate = 24000
const channels = 1
const bitsPerSample = 16

func parseGeminiModelVoice(input string) (modelId, voiceId string) {
	if input == "" {
		return defaultGeminiModel, defaultGeminiVoice
	}
	for _, id := range knownGeminiModels {
		if input == id {
			return id, defaultGeminiVoice
		}
		if strings.HasPrefix(input, id+"/") {
			return id, input[len(id)+1:]
		}
	}
	return defaultGeminiModel, input
}

func buildPrompt(text, language string) string {
	if strings.Contains(text, ": ") {
		return text
	}
	if language != "" {
		return fmt.Sprintf("Say in %s: %s", language, text)
	}
	return fmt.Sprintf("Say: %s", text)
}

func pcmToWav(pcm []byte) []byte {
	dataSize := len(pcm)
	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8

	header := make([]byte, 44)
	copy(header[0:], []byte("RIFF"))
	binary.LittleEndian.PutUint32(header[4:], uint32(36+dataSize))
	copy(header[8:], []byte("WAVE"))
	copy(header[12:], []byte("fmt "))
	binary.LittleEndian.PutUint32(header[16:], 16)
	binary.LittleEndian.PutUint16(header[20:], 1)
	binary.LittleEndian.PutUint16(header[22:], uint16(channels))
	binary.LittleEndian.PutUint32(header[24:], sampleRate)
	binary.LittleEndian.PutUint32(header[28:], uint32(byteRate))
	binary.LittleEndian.PutUint16(header[32:], uint16(blockAlign))
	binary.LittleEndian.PutUint16(header[34:], bitsPerSample)
	copy(header[36:], []byte("data"))
	binary.LittleEndian.PutUint32(header[40:], uint32(dataSize))

	return append(header, pcm...)
}

func (p *GeminiProvider) Synthesize(ctx context.Context, text string, model string, creds *Credentials) (*TTSResult, error) {
	if creds.APIKey == "" {
		return nil, fmt.Errorf("no Gemini API key configured")
	}

	modelId, voiceId := parseGeminiModelVoice(model)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelId, creds.APIKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": buildPrompt(text, "")},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseModalities": []string{"AUDIO"},
			"speechConfig": map[string]interface{}{
				"voiceConfig": map[string]interface{}{
					"prebuiltVoiceConfig": map[string]string{
						"voiceName": voiceId,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("Gemini TTS failed: %d", res.StatusCode)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	candidates, ok := resp["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return nil, fmt.Errorf("Gemini TTS: invalid or empty candidates")
	}

	candidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Gemini TTS: invalid candidate type")
	}

	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Gemini TTS: invalid content type")
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return nil, fmt.Errorf("Gemini TTS: invalid or empty parts")
	}

	part, ok := parts[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Gemini TTS: invalid part type")
	}

	inlineData, ok := part["inlineData"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Gemini TTS: invalid inlineData")
	}

	mimeType, ok := inlineData["mimeType"].(string)
	if !ok {
		return nil, fmt.Errorf("Gemini TTS: invalid mimeType")
	}

	data, ok := inlineData["data"].(string)
	if !ok {
		return nil, fmt.Errorf("Gemini TTS: invalid data")
	}

	if data == "" {
		return nil, fmt.Errorf("Gemini TTS returned no audio data")
	}

	pcm, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio: %w", err)
	}

	if strings.Contains(mimeType, "wav") {
		return &TTSResult{Base64: base64.StdEncoding.EncodeToString(pcm), Format: "wav"}, nil
	}

	wav := pcmToWav(pcm)
	return &TTSResult{Base64: base64.StdEncoding.EncodeToString(wav), Format: "wav"}, nil
}
