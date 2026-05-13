package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
)

// ref: open-sse/handlers/fetch/index.js:1-237

const (
	defaultTimeoutMs   = 15000
	defaultFormat      = "markdown"
	defaultMaxChars    = 50000
	maxContentSize     = 10 * 1024 * 1024 // 10MB max content size
	fetchTimeoutMs     = 30000            // 30s total fetch timeout
	rateLimitPerMin    = 60               // requests per minute per API key
)

// FetchRequest represents the request body for /v1/fetch
// ref: open-sse/handlers/fetch/index.js:88-94
type FetchRequest struct {
	URL           string `json:"url"`
	Format        string `json:"format,omitempty"`
	MaxCharacters int    `json:"max_characters,omitempty"`
	Provider      string `json:"provider,omitempty"`
}

// FetchResponse represents the response for /v1/fetch
// ref: open-sse/handlers/fetch/index.js:56-66
type FetchResponse struct {
	Provider string        `json:"provider"`
	URL      string        `json:"url"`
	Title    string        `json:"title,omitempty"`
	Content  ContentData   `json:"content"`
	Metadata FetchMetadata `json:"metadata"`
	Usage    FetchUsage    `json:"usage"`
	Metrics  FetchMetrics  `json:"metrics"`
}

// ContentData holds the fetched content
type ContentData struct {
	Format string `json:"format"`
	Text   string `json:"text"`
	Length int    `json:"length"`
}

// FetchMetadata holds optional metadata
type FetchMetadata struct {
	Author      string `json:"author"`
	PublishedAt string `json:"published_at"`
	Language    string `json:"language"`
}

// FetchUsage holds usage/cost info
type FetchUsage struct {
	FetchCostUSD float64 `json:"fetch_cost_usd"`
}

// FetchMetrics holds timing metrics
type FetchMetrics struct {
	ResponseTimeMs    int64 `json:"response_time_ms"`
	UpstreamLatencyMs int64 `json:"upstream_latency_ms"`
}

// ErrorResponse for fetch failures
// ref: open-sse/handlers/fetch/index.js:89-90
type FetchErrorResponse struct {
	Success bool   `json:"success"`
	Status  int    `json:"status,omitempty"`
	Error   string `json:"error"`
}

// DomainAllowlist - domains allowed for fetching (security)
var domainAllowlist = map[string]bool{
	"r.jina.ai":          true, // jina reader
	"api.firecrawl.dev":  true,
	"api.tavily.com":     true,
	"api.exa.ai":         true,
	// Add more allowed domains as needed
}

// isURLAllowed checks if the target URL domain is allowed
func isURLAllowed(targetURL string) bool {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	host := parsed.Host
	// Strip port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return domainAllowlist[host]
}

// sanitizeHeaders removes non-ASCII characters from header values
// ref: open-sse/handlers/fetch/index.js:21-29
func sanitizeHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	out := make(map[string]string)
	for k, v := range headers {
		// Remove non-ASCII chars
		cleaned := strings.Map(func(r rune) rune {
			if r < 128 {
				return r
			}
			return -1
		}, v)
		out[k] = strings.TrimSpace(cleaned)
	}
	return out
}

// truncate truncates text to max characters
// ref: open-sse/handlers/fetch/index.js:45-49
func truncate(text string, max int) string {
	if text == "" || max <= 0 {
		return text
	}
	if len(text) > max {
		return text[:max]
	}
	return text
}

// parseJinaTitle extracts title from Jina markdown output
// ref: open-sse/handlers/fetch/index.js:51-54
func parseJinaTitle(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

// tryFetch performs HTTP request with timeout
// ref: open-sse/handlers/fetch/index.js:31-43
func tryFetch(ctx context.Context, targetURL string, headers map[string]string, timeoutMs int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	cleanHeaders := sanitizeHeaders(headers)
	for k, v := range cleanHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	return client.Do(req)
}

// tryFetchPost performs POST request with timeout
func tryFetchPost(ctx context.Context, targetURL string, headers map[string]string, body io.Reader, timeoutMs int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, body)
	if err != nil {
		return nil, err
	}

	cleanHeaders := sanitizeHeaders(headers)
	for k, v := range cleanHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	return client.Do(req)
}

// runJina fetches content via Jina Reader
// ref: open-sse/handlers/fetch/index.js:153-177
func runJina(ctx context.Context, targetURL string, apiKey string, maxChars int, startedAt time.Time) (*FetchResponse, *FetchErrorResponse) {
	encodedURL := url.QueryEscape(targetURL)
	jinaURL := fmt.Sprintf("https://r.jina.ai/%s", encodedURL)
	
	upstreamStart := time.Now()
	
	headers := map[string]string{}
	if apiKey != "" {
		headers["Authorization"] = "Bearer " + apiKey
	}
	
	resp, err := tryFetch(ctx, jinaURL, headers, defaultTimeoutMs)
	if err != nil {
		return nil, &FetchErrorResponse{
			Success: false,
			Status:  http.StatusGatewayTimeout,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()
	
	upstreamMs := time.Since(upstreamStart).Milliseconds()
	
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxContentSize))
	if err != nil {
		return nil, &FetchErrorResponse{
			Success: false,
			Status:  http.StatusBadGateway,
			Error:   "failed to read response body",
		}
	}
	
	if resp.StatusCode != http.StatusOK {
		errMsg := string(body)
		if len(errMsg) > 500 {
			errMsg = errMsg[:500]
		}
		return nil, &FetchErrorResponse{
			Success: false,
			Status:  resp.StatusCode,
			Error:   errMsg,
		}
	}
	
	text := truncate(string(body), maxChars)
	
	return &FetchResponse{
		Provider: "jina-reader",
		URL:      targetURL,
		Title:    parseJinaTitle(string(body)),
		Content: ContentData{
			Format: "markdown",
			Text:   text,
			Length: len(text),
		},
		Metadata: FetchMetadata{},
		Usage:    FetchUsage{},
		Metrics: FetchMetrics{
			ResponseTimeMs:    time.Since(startedAt).Milliseconds(),
			UpstreamLatencyMs: upstreamMs,
		},
	}, nil
}

// HandleFetch handles POST /v1/fetch endpoint
// ref: open-sse/handlers/fetch/index.js:88-120
func HandleFetch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startedAt := time.Now()

	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeFetchError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req FetchRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeFetchError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate URL
	if req.URL == "" {
		// ref: open-sse/handlers/fetch/index.js:89-91
		writeFetchError(w, http.StatusBadRequest, "url is required")
		return
	}

	// Validate URL format
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		writeFetchError(w, http.StatusBadRequest, "invalid url format")
		return
	}

	// Security: only allow http/https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		writeFetchError(w, http.StatusBadRequest, "only http and https urls are allowed")
		return
	}

	// Set defaults
	format := req.Format
	if format == "" {
		format = defaultFormat
	}

	maxChars := req.MaxCharacters
	if maxChars <= 0 {
		maxChars = defaultMaxChars
	}

	// Create context with timeout
	fetchCtx, cancel := context.WithTimeout(ctx, time.Duration(fetchTimeoutMs)*time.Millisecond)
	defer cancel()

	// Default to jina-reader provider (free, no API key required)
	provider := req.Provider
	if provider == "" {
		provider = "jina-reader"
	}

	var fetchResp *FetchResponse
	var fetchErr *FetchErrorResponse

	// Route to appropriate provider
	switch provider {
	case "jina-reader":
		fetchResp, fetchErr = runJina(fetchCtx, req.URL, "", maxChars, startedAt)
	default:
		// ref: open-sse/handlers/fetch/index.js:115
		writeFetchError(w, http.StatusBadRequest, fmt.Sprintf("unsupported provider: %s", provider))
		return
	}

	if fetchErr != nil {
		writeFetchError(w, fetchErr.Status, fetchErr.Error)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fetchResp)
}

// writeFetchError writes a JSON error response
func writeFetchError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(FetchErrorResponse{
		Success: false,
		Status:  status,
		Error:   message,
	})
}

// RequireFetchAuth is a middleware that checks API key for fetch endpoint
func RequireFetchAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// API key check is handled by the main API key middleware in routes.go
		// This is just a placeholder for additional fetch-specific auth if needed
		next(w, r)
	}
}

// RegisterFetchRoutes registers fetch-related routes
func RegisterFetchRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/fetch", HandleFetch)
}

// ValidateFetchURL performs security validation on fetch URLs
func ValidateFetchURL(targetURL string) error {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only HTTP/HTTPS URLs allowed")
	}

	// Block private/internal IP ranges (SSRF protection)
	host := parsed.Hostname()
	if host == "localhost" || host == "127.0.0.1" || 
	   strings.HasPrefix(host, "10.") ||
	   strings.HasPrefix(host, "192.168.") ||
	   strings.HasPrefix(host, "172.16.") ||
	   strings.HasPrefix(host, "172.17.") ||
	   strings.HasPrefix(host, "172.18.") ||
	   strings.HasPrefix(host, "172.19.") ||
	   strings.HasPrefix(host, "172.20.") ||
	   strings.HasPrefix(host, "172.21.") ||
	   strings.HasPrefix(host, "172.22.") ||
	   strings.HasPrefix(host, "172.23.") ||
	   strings.HasPrefix(host, "172.24.") ||
	   strings.HasPrefix(host, "172.25.") ||
	   strings.HasPrefix(host, "172.26.") ||
	   strings.HasPrefix(host, "172.27.") ||
	   strings.HasPrefix(host, "172.28.") ||
	   strings.HasPrefix(host, "172.29.") ||
	   strings.HasPrefix(host, "172.30.") ||
	   strings.HasPrefix(host, "172.31.") ||
	   strings.HasSuffix(host, ".local") ||
	   strings.HasSuffix(host, ".internal") {
		return fmt.Errorf("access to private/internal IPs is blocked")
	}

	return nil
}

// HandleFetchWithValidation is the main fetch handler with full validation
func HandleFetchWithValidation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startedAt := time.Now()

	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req FetchRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate URL presence
	if req.URL == "" {
		errs.WriteJSONError(w, "url is required", http.StatusBadRequest)
		return
	}

	// Security validation
	if err := ValidateFetchURL(req.URL); err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set defaults
	format := req.Format
	if format == "" {
		format = defaultFormat
	}

	maxChars := req.MaxCharacters
	if maxChars <= 0 {
		maxChars = defaultMaxChars
	}

	// Create context with timeout
	fetchCtx, cancel := context.WithTimeout(ctx, time.Duration(fetchTimeoutMs)*time.Millisecond)
	defer cancel()

	// Default to jina-reader provider
	provider := req.Provider
	if provider == "" {
		provider = "jina-reader"
	}

	var fetchResp *FetchResponse
	var fetchErr *FetchErrorResponse

	// Route to appropriate provider
	switch provider {
	case "jina-reader":
		fetchResp, fetchErr = runJina(fetchCtx, req.URL, "", maxChars, startedAt)
	default:
		errs.WriteJSONError(w, fmt.Sprintf("unsupported provider: %s", provider), http.StatusBadRequest)
		return
	}

	if fetchErr != nil {
		errs.WriteJSONError(w, fetchErr.Error, fetchErr.Status)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fetchResp)
}
