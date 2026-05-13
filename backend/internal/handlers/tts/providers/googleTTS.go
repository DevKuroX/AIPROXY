// ref: _ref/9router/open-sse/handlers/ttsProviders/googleTts.js
package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

func init() {
	Register("google-tts", &GoogleTTSProvider{})
}

// GoogleTTSProvider implements TTS for Google Translate.
// ref: _ref/9router/open-sse/handlers/ttsProviders/googleTts.js
type GoogleTTSProvider struct {
	mu        sync.Mutex
	token     *googleToken
	tokenTime time.Time
	idx       int
}

type googleToken struct {
	FSid string `json:"f.sid"`
	Bl   string `json:"bl"`
}

const refreshMSGoogle = 11 * time.Minute

func (p *GoogleTTSProvider) getToken(ctx context.Context) (*googleToken, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if p.token != nil && now.Sub(p.tokenTime) < refreshMSGoogle {
		return p.token, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://translate.google.com/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UA)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Google translate fetch failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google translate fetch failed: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)
	fSidRe := regexp.MustCompile(`"FdrFJe":"(.*?)"`)
	blRe := regexp.MustCompile(`"cfb2h":"(.*?)"`)

	fSidMatch := fSidRe.FindStringSubmatch(html)
	blMatch := blRe.FindStringSubmatch(html)

	if fSidMatch == nil || blMatch == nil {
		return nil, fmt.Errorf("failed to parse Google token")
	}

	p.token = &googleToken{
		FSid: fSidMatch[1],
		Bl:   blMatch[1],
	}
	p.tokenTime = now

	return p.token, nil
}

func (p *GoogleTTSProvider) Synthesize(ctx context.Context, text string, model string, creds *Credentials) (*TTSResult, error) {
	lang := model
	if lang == "" {
		lang = "en"
	}

	token, err := p.getToken(ctx)
	if err != nil {
		return nil, err
	}

	replacer := strings.NewReplacer(
		"@", " ", "^", " ", "*", " ", "(", " ", ")", " ",
		"\\", " ", "_", " ", "-", " ", "+", " ", "=", " ",
		">", " ", "<", " ", "\"", " ", "'", " ",
		", ", ". ",
	)
	cleanText := replacer.Replace(text)

	p.mu.Lock()
	p.idx++
	reqId := p.idx*100000 + 1000 + int(time.Now().UnixNano()%9000)
	p.mu.Unlock()

	rpcId := "jQ1olc"
	query := url.Values{}
	query.Set("rpcids", rpcId)
	query.Set("f.sid", token.FSid)
	query.Set("bl", token.Bl)
	query.Set("hl", lang)
	query.Set("soc-app", "1")
	query.Set("soc-platform", "1")
	query.Set("soc-device", "1")
	query.Set("_reqid", fmt.Sprintf("%d", reqId))
	query.Set("rt", "c")

	payload := []interface{}{cleanText, lang, nil, "undefined", []interface{}{0}}
	innerArr := []interface{}{rpcId, payload, nil, "generic"}
	outerArr := [][]interface{}{{innerArr}}
	fReq, _ := json.Marshal(outerArr)

	formData := url.Values{}
	formData.Set("f.req", string(fReq))

	targetURL := fmt.Sprintf("https://translate.google.com/_/TranslateWebserverUi/data/batchexecute?%s", query.Encode())

	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://translate.google.com/")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Google TTS failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google TTS failed: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")
	if len(lines) < 4 {
		return nil, fmt.Errorf("Google TTS returned unexpected response")
	}

	var split []interface{}
	if err := json.Unmarshal([]byte(lines[3]), &split); err != nil {
		return nil, fmt.Errorf("failed to parse Google response: %w", err)
	}

	if len(split) == 0 {
		return nil, fmt.Errorf("Google TTS returned empty audio")
	}

	inner, _ := split[0].([]interface{})
	if len(inner) < 3 {
		return nil, fmt.Errorf("Google TTS returned empty audio")
	}

	innerInner, _ := inner[2].(string)
	var parsed []interface{}
	if err := json.Unmarshal([]byte(innerInner), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse Google audio data: %w", err)
	}

	if len(parsed) == 0 {
		return nil, fmt.Errorf("Google TTS returned empty audio")
	}

	base64Audio, _ := parsed[0].(string)
	if base64Audio == "" || len(base64Audio) < 100 {
		return nil, fmt.Errorf("Google TTS returned empty audio")
	}

	return &TTSResult{Base64: base64Audio, Format: "mp3"}, nil
}
