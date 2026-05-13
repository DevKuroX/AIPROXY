// ref: _ref/9router/open-sse/handlers/ttsProviders/edgeTts.js
package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

func init() {
	Register("edge-tts", &EdgeTTSProvider{})
}

// EdgeTTSProvider implements TTS for Microsoft Edge.
// ref: _ref/9router/open-sse/handlers/ttsProviders/edgeTts.js
type EdgeTTSProvider struct {
	mu         sync.Mutex
	token      *edgeToken
	tokenTime  time.Time
	voices     []Voice
	voicesTime time.Time
}

type edgeToken struct {
	Key    string `json:"key"`
	Token  string `json:"token"`
	Cookie string `json:"cookie"`
}

const refreshMS = 5 * time.Minute
const voicesTTL = 24 * time.Hour

func (p *EdgeTTSProvider) getToken(ctx context.Context) (*edgeToken, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if p.token != nil && now.Sub(p.tokenTime) < refreshMS {
		return p.token, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.bing.com/translator", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Accept-Language", "vi,en-US;q=0.9,en;q=0.8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Bing translator fetch failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bing translator fetch failed: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	re := regexp.MustCompile(`params_AbusePreventionHelper\s*=\s*\[([^,]+),([^,]+),`)
	match := re.FindStringSubmatch(html)
	if match == nil {
		return nil, fmt.Errorf("failed to parse Bing token")
	}

	cookies := res.Cookies()
	var cookieStrs []string
	for _, c := range cookies {
		cookieStrs = append(cookieStrs, c.Name+"="+c.Value)
	}

	p.token = &edgeToken{
		Key:    match[1],
		Token:  strings.Trim(match[2], `"`),
		Cookie: strings.Join(cookieStrs, "; "),
	}
	p.tokenTime = now

	return p.token, nil
}

func (p *EdgeTTSProvider) Synthesize(ctx context.Context, text string, model string, creds *Credentials) (*TTSResult, error) {
	token, err := p.getToken(ctx)
	if err != nil {
		return nil, err
	}

	voiceId := model
	if voiceId == "" {
		voiceId = "en-US-AvaNeural"
	}

	parts := strings.Split(voiceId, "-")
	xmlLang := strings.Join(parts[:2], "-")
	gender := "Female"
	if strings.Contains(strings.ToLower(voiceId), "male") {
		gender = "Male"
	}

	ssml := fmt.Sprintf(`<speak version='1.0' xml:lang='%s'><voice xml:lang='%s' xml:gender='%s' name='%s'><prosody rate='0.00%%'>%s</prosody></voice></speak>`,
		xmlLang, xmlLang, gender, voiceId, text)

	formData := strings.NewReader(fmt.Sprintf("ssml=%s&token=%s&key=%s",
		escapeForm(ssml), escapeForm(token.Token), escapeForm(token.Key)))

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://www.bing.com/tfettts?isVertical=1&&IG=1&IID=translator.5023&SFX=1", formData)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://www.bing.com")
	req.Header.Set("Referer", "https://www.bing.com/translator")
	req.Header.Set("User-Agent", UA)
	if token.Cookie != "" {
		req.Header.Set("Cookie", token.Cookie)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Edge TTS failed: %d", res.StatusCode)
	}

	audio, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if len(audio) < 100 {
		return nil, fmt.Errorf("Edge TTS returned empty audio")
	}

	return &TTSResult{
		Base64: base64.StdEncoding.EncodeToString(audio),
		Format: "mp3",
	}, nil
}

func (p *EdgeTTSProvider) FetchVoices(ctx context.Context, creds *Credentials) ([]Voice, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if p.voices != nil && now.Sub(p.voicesTime) < voicesTTL {
		return p.voices, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://speech.platform.bing.com/consumer/speech/synthesize/readaloud/voices/list?trustedclienttoken=6A5AA1D4EAFF4E9FB37E23D68491D6F4", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UA)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Edge TTS voices fetch failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Edge TTS voices fetch failed: %d", res.StatusCode)
	}

	var voices []Voice
	if err := json.NewDecoder(res.Body).Decode(&voices); err != nil {
		return nil, err
	}

	p.voices = voices
	p.voicesTime = now
	return voices, nil
}

func escapeForm(s string) string {
	s = strings.ReplaceAll(s, " ", "+")
	s = strings.ReplaceAll(s, "'", "%27")
	s = strings.ReplaceAll(s, "<", "%3C")
	s = strings.ReplaceAll(s, ">", "%3E")
	s = strings.ReplaceAll(s, "\"", "%22")
	return s
}
