package v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/handlers/search"
)

// ref: open-sse/handlers/search/index.js:1-201

const (
	globalSearchTimeoutMs = 15000
)

// SearchRequest represents the request body for /v1/search
// ref: open-sse/handlers/search/index.js:20-25
type SearchRequest struct {
	Query           string                 `json:"query"`
	Provider        string                 `json:"provider,omitempty"`
	SearchType      string                 `json:"search_type,omitempty"`
	MaxResults      int                    `json:"max_results,omitempty"`
	Country         string                 `json:"country,omitempty"`
	Language        string                 `json:"language,omitempty"`
	TimeRange       string                 `json:"time_range,omitempty"`
	Offset          int                    `json:"offset,omitempty"`
	DomainFilter    []string               `json:"domain_filter,omitempty"`
	ContentOptions  *search.ContentOptions `json:"content_options,omitempty"`
	ProviderOptions map[string]interface{} `json:"provider_options,omitempty"`
}

// SearchResult represents a single search result item
// ref: open-sse/handlers/search/normalizers.js:9-32
type SearchResult struct {
	Title       string                 `json:"title"`
	URL         string                 `json:"url"`
	DisplayURL  string                 `json:"display_url,omitempty"`
	Snippet     string                 `json:"snippet"`
	Position    int                    `json:"position"`
	Score       *float64               `json:"score,omitempty"`
	PublishedAt *string                `json:"published_at,omitempty"`
	FaviconURL  *string                `json:"favicon_url,omitempty"`
	Content     *search.ContentData    `json:"content,omitempty"`
	Metadata    *search.ResultMetadata `json:"metadata,omitempty"`
	Citation    *search.Citation       `json:"citation,omitempty"`
	ProviderRaw interface{}            `json:"provider_raw,omitempty"`
}

// SearchResponse represents the response for /v1/search
// ref: open-sse/handlers/search/index.js:115-126
type SearchResponse struct {
	Provider string          `json:"provider"`
	Query    string          `json:"query"`
	Results  []SearchResult  `json:"results"`
	Answer   *search.Answer  `json:"answer,omitempty"`
	Usage    *search.Usage   `json:"usage,omitempty"`
	Metrics  *search.Metrics `json:"metrics,omitempty"`
	Errors   []string        `json:"errors,omitempty"`
}

// SearchProviderStore provides search provider configurations
type SearchProviderStore interface {
	GetSearchProvider(providerID string) (*search.ProviderConfig, error)
}

var searchProviderStore SearchProviderStore

// SetSearchProviderStore sets the search provider store
func SetSearchProviderStore(store SearchProviderStore) {
	searchProviderStore = store
}

var searchEnabled bool

// SetSearchEnabled enables or disables search
func SetSearchEnabled(enabled bool) {
	searchEnabled = enabled
}

// IsSearchEnabled returns whether search is enabled
func IsSearchEnabled() bool {
	return searchEnabled
}

// ref: open-sse/handlers/search/index.js:19-25
// sanitizeQuery normalizes and validates the query string
func sanitizeQuery(query string) (string, error) {
	// Check for control characters
	for _, r := range query {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return "", errs.ErrInvalidToken
		}
	}

	// Normalize and trim
	clean := strings.TrimSpace(query)
	clean = strings.Join(strings.Fields(clean), " ")

	if clean == "" {
		return "", errs.ErrInvalidToken
	}

	return clean, nil
}

// stripNonASCII removes non-ASCII characters from header values
// ref: open-sse/handlers/search/index.js:28-35
func stripNonASCII(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r <= 0xFF {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// isControlChar checks if a rune is a control character
func isControlChar(r rune) bool {
	return (r >= 0x00 && r <= 0x08) || r == 0x0B || r == 0x0C || (r >= 0x0E && r <= 0x1F) || r == 0x7F
}

// containsControlChars checks if a string contains control characters
// ref: open-sse/handlers/search/index.js:17
func containsControlChars(s string) bool {
	for _, r := range s {
		if isControlChar(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

// ref: open-sse/handlers/search/index.js:147-201
// HandleSearch handles POST /v1/search
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	if !searchEnabled {
		errs.WriteJSONError(w, "search is not enabled", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), globalSearchTimeoutMs*time.Millisecond)
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req SearchRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// ref: open-sse/handlers/search/index.js:150-153
	// Sanitize query
	if containsControlChars(req.Query) {
		errs.WriteJSONError(w, "query contains invalid control characters", http.StatusBadRequest)
		return
	}

	cleanQuery, err := sanitizeQuery(req.Query)
	if err != nil {
		errs.WriteJSONError(w, "query is empty after normalization", http.StatusBadRequest)
		return
	}
	req.Query = cleanQuery

	// Get provider config
	if searchProviderStore == nil {
		errs.WriteJSONError(w, "no search provider store configured", http.StatusInternalServerError)
		return
	}

	providerID := req.Provider
	if providerID == "" {
		providerID = "brave-search" // default provider
	}

	providerConfig, err := searchProviderStore.GetSearchProvider(providerID)
	if err != nil {
		errs.WriteJSONError(w, "failed to get search provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if providerConfig == nil {
		errs.WriteJSONError(w, "provider "+providerID+" does not support web search", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.SearchType == "" {
		req.SearchType = "web"
	}
	if req.MaxResults <= 0 {
		req.MaxResults = 5
	}

	// Execute search
	startTime := time.Now()
	result, err := search.ExecuteSearch(ctx, providerConfig, &search.SearchParams{
		Query:          req.Query,
		SearchType:     req.SearchType,
		MaxResults:     req.MaxResults,
		Country:        req.Country,
		Language:       req.Language,
		TimeRange:      req.TimeRange,
		Offset:         req.Offset,
		DomainFilter:   req.DomainFilter,
		ContentOptions: req.ContentOptions,
	})
	duration := time.Since(startTime)

	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Convert results
	results := make([]SearchResult, len(result.Results))
	for i, r := range result.Results {
		results[i] = SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			DisplayURL:  r.DisplayURL,
			Snippet:     r.Snippet,
			Position:    r.Position,
			Score:       r.Score,
			PublishedAt: r.PublishedAt,
			FaviconURL:  r.FaviconURL,
			Content:     r.Content,
			Metadata:    r.Metadata,
			Citation:    r.Citation,
		}
	}

	resp := SearchResponse{
		Provider: providerID,
		Query:    req.Query,
		Results:  results,
		Metrics: &search.Metrics{
			ResponseTimeMs:    duration.Milliseconds(),
			UpstreamLatencyMs: result.UpstreamLatencyMs,
			TotalResults:      result.TotalResults,
		},
		Usage: &search.Usage{
			QueriesUsed:  1,
			SearchCostUSD: providerConfig.CostPerQuery,
		},
		Errors: []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
