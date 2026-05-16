package geminiweb

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// httpClient wraps the standard HTTP client with cookie management.
type httpClient struct {
	client  *http.Client
	cookies map[string]string
}

func newHTTPClient(proxy string) *httpClient {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{}
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &httpClient{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
			Jar:       jar,
		},
		cookies: make(map[string]string),
	}
}

func (c *httpClient) setCookies(cookies map[string]string) {
	for k, v := range cookies {
		c.cookies[k] = v
	}
}

func (c *httpClient) do(req *http.Request) (*http.Response, error) {
	// Set cookies on request
	for name, value := range c.cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value, Domain: ".google.com"})
	}
	return c.client.Do(req)
}

var snlM0eRe = regexp.MustCompile(`"SNlM0e":"([^"]+)"`)
var buildLabelRe = regexp.MustCompile(`"buildLabel":"([^"]+)"`)
var sessionIDRe = regexp.MustCompile(`"sessionId":"([^"]+)"`)

// NewSession creates a new Gemini web session from cookies.
func NewSession(secure1psid, secure1psidts, proxy string) *Session {
	s := &Session{
		Secure1PSID:   secure1psid,
		Secure1PSIDTS: secure1psidts,
		Language:      "en",
		Proxy:         proxy,
		reqCounter:    10000,
	}
	s.client = newHTTPClient(proxy)
	s.client.setCookies(map[string]string{
		"__Secure-1PSID":   secure1psid,
		"__Secure-1PSIDTS": secure1psidts,
	})
	return s
}

// Init initializes the session by fetching the access token from gemini.google.com.
func (s *Session) Init() error {
	// Step 1: Preflight to www.google.com
	req, _ := http.NewRequest("GET", endpointGoogle, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := s.client.do(req)
	if err != nil {
		return fmt.Errorf("preflight failed: %w", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Step 2: Get gemini.google.com page to extract token
	req, _ = http.NewRequest("GET", endpointInit, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://gemini.google.com")
	req.Header.Set("Referer", "https://gemini.google.com/")

	resp, err = s.client.do(req)
	if err != nil {
		return fmt.Errorf("init request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read init response: %w", err)
	}
	html := string(body)

	// Extract SNlM0e access token
	matches := snlM0eRe.FindStringSubmatch(html)
	if len(matches) < 2 {
		return fmt.Errorf("failed to extract SNlM0e token from page (cookies may be invalid or expired)")
	}
	s.AccessToken = matches[1]

	// Extract build label
	if m := buildLabelRe.FindStringSubmatch(html); len(m) >= 2 {
		s.BuildLabel = m[1]
	}

	// Extract session ID
	if m := sessionIDRe.FindStringSubmatch(html); len(m) >= 2 {
		s.SessionID = m[1]
	}

	s.lastRefresh = time.Now()
	return nil
}

// RefreshAccessToken re-fetches the access token.
func (s *Session) RefreshAccessToken() error {
	// Refresh cookies first via RotateCookies
	err := s.rotateCookies()
	if err != nil {
		return err
	}

	// Re-init to get new token
	return s.Init()
}

func (s *Session) rotateCookies() error {
	data := fmt.Sprintf(`[null,null,[[null,null,null,null,null,null,null,null,"%s"]]]`, s.Secure1PSIDTS)

	req, _ := http.NewRequest("POST", endpointRotate, strings.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://accounts.google.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	resp, err := s.client.do(req)
	if err != nil {
		return fmt.Errorf("cookie rotation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("cookie rotation returned %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	// Response contains new __Secure-1PSIDTS
	// Format: [null,null,[[null,null,[null,null,null,"<new_token>"]]]]
	parts := strings.Split(string(body), `"`)
	for i, p := range parts {
		if strings.HasPrefix(p, "sidts-") {
			s.Secure1PSIDTS = p
			s.client.setCookies(map[string]string{
				"__Secure-1PSID":   s.Secure1PSID,
				"__Secure-1PSIDTS": s.Secure1PSIDTS,
			})
			break
		}
		if i > 0 && strings.Contains(p, "sidts-") {
			s.Secure1PSIDTS = strings.TrimSuffix(p, `\u003d`)
			s.client.setCookies(map[string]string{
				"__Secure-1PSID":   s.Secure1PSID,
				"__Secure-1PSIDTS": s.Secure1PSIDTS,
			})
			break
		}
	}

	return nil
}

// IsTokenExpired checks if the access token needs refreshing.
func (s *Session) IsTokenExpired() bool {
	return time.Since(s.lastRefresh) > 50*time.Minute
}

// IsAuthenticated returns true if the session has a valid access token.
func (s *Session) IsAuthenticated() bool {
	return s.AccessToken != ""
}
