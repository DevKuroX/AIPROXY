package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ref: open-sse/handlers/search/callers.js:1-371
// ref: open-sse/handlers/search/normalizers.js:1-223
// ref: open-sse/handlers/search/chatSearch.js:1-409

// ProviderConfig holds search provider configuration
// ref: open-sse/handlers/search/index.js:64-70
type ProviderConfig struct {
	ID                 string
	BaseURL            string
	Method             string
	AuthType           string
	SearchTypes        []string
	DefaultMaxResults  int
	MaxMaxResults      int
	TimeoutMs          int
	CostPerQuery       float64
	DefaultSearchType  string
	ProviderSpecific   map[string]interface{}
}

// SearchParams holds parameters for a search request
// ref: open-sse/handlers/search/callers.js:18-31
type SearchParams struct {
	Query          string
	SearchType     string
	MaxResults     int
	Token          string
	Country        string
	Language       string
	TimeRange      string
	Offset         int
	DomainFilter   []string
	ContentOptions *ContentOptions
	ProviderOpts   map[string]interface{}
	ProviderData   map[string]interface{}
}

// ContentOptions controls what content to fetch
// ref: open-sse/handlers/search/callers.js:13-16
type ContentOptions struct {
	Snippet      bool   `json:"snippet,omitempty"`
	FullPage     bool   `json:"full_page,omitempty"`
	Format       string `json:"format,omitempty"`
	MaxCharacters int   `json:"max_characters,omitempty"`
}

// ContentData holds fetched content
type ContentData struct {
	Format string `json:"format"`
	Text   string `json:"text"`
	Length int    `json:"length"`
}

// ResultMetadata holds optional result metadata
// ref: open-sse/handlers/search/normalizers.js:23-28
type ResultMetadata struct {
	Author     string `json:"author,omitempty"`
	Language   string `json:"language,omitempty"`
	SourceType string `json:"source_type,omitempty"`
	ImageURL   string `json:"image_url,omitempty"`
}

// Citation holds citation info
// ref: open-sse/handlers/search/normalizers.js:29
type Citation struct {
	Provider    string `json:"provider"`
	RetrievedAt string `json:"retrieved_at"`
	Rank        int    `json:"rank"`
}

// Answer holds chat-based search answer
// ref: open-sse/handlers/search/chatSearch.js:397
type Answer struct {
	Source string `json:"source"`
	Text   string `json:"text"`
	Model  string `json:"model"`
}

// Usage holds search usage info
// ref: open-sse/handlers/search/index.js:122
type Usage struct {
	QueriesUsed  int     `json:"queries_used"`
	SearchCostUSD float64 `json:"search_cost_usd"`
	LLMTokens    int     `json:"llm_tokens,omitempty"`
}

// Metrics holds search timing metrics
// ref: open-sse/handlers/search/index.js:123
type Metrics struct {
	ResponseTimeMs    int64 `json:"response_time_ms"`
	UpstreamLatencyMs int64 `json:"upstream_latency_ms"`
	TotalResults      int   `json:"total_results_available,omitempty"`
}

// Result holds a single search result
// ref: open-sse/handlers/search/normalizers.js:9-32
type Result struct {
	Title       string
	URL         string
	DisplayURL  string
	Snippet     string
	Position    int
	Score       *float64
	PublishedAt *string
	FaviconURL  *string
	Content     *ContentData
	Metadata    *ResultMetadata
	Citation    *Citation
}

// ExecuteResult holds the result of executing a search
type ExecuteResult struct {
	Results           []Result
	TotalResults      int
	UpstreamLatencyMs int64
}

// parseDomainFilter splits domain filter into includes and excludes
// ref: open-sse/handlers/search/callers.js:40-45
func parseDomainFilter(domainFilter []string) (includes, excludes []string) {
	for _, d := range domainFilter {
		if strings.HasPrefix(d, "-") {
			excludes = append(excludes, strings.TrimPrefix(d, "-"))
		} else {
			includes = append(includes, d)
		}
	}
	return
}

// getProviderSetting reads a string setting from provider options
// ref: open-sse/handlers/search/callers.js:53-63
func getProviderSetting(params *SearchParams, key string) string {
	if params.ProviderOpts != nil {
		if v, ok := params.ProviderOpts[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	if params.ProviderData != nil {
		if v, ok := params.ProviderData[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

// resolveBaseURL resolves base URL with optional override
// ref: open-sse/handlers/search/callers.js:71-74
func resolveBaseURL(config *ProviderConfig, params *SearchParams) string {
	override := getProviderSetting(params, "baseUrl")
	base := override
	if base == "" {
		base = config.BaseURL
	}
	return strings.TrimRight(base, "/")
}

// toPageNumber converts offset + maxResults to 1-indexed page number
// ref: open-sse/handlers/search/callers.js:82-85
func toPageNumber(offset, maxResults int) int {
	if offset <= 0 || maxResults <= 0 {
		return 0
	}
	return (offset / maxResults) + 1
}

// HTTPRequest represents an HTTP request to be made
type HTTPRequest struct {
	URL    string
	Method string
	Header http.Header
	Body   []byte
}

// buildSerperRequest builds a request for Serper API
// ref: open-sse/handlers/search/callers.js:89-102
func buildSerperRequest(config *ProviderConfig, params *SearchParams) *HTTPRequest {
	endpoint := "/search"
	if params.SearchType == "news" {
		endpoint = "/news"
	}

	body := map[string]interface{}{
		"q":   params.Query,
		"num": params.MaxResults,
	}
	if params.Country != "" {
		body["gl"] = strings.ToLower(params.Country)
	}
	if params.Language != "" {
		body["hl"] = params.Language
	}

	jsonBody, _ := json.Marshal(body)

	return &HTTPRequest{
		URL:    resolveBaseURL(config, params) + endpoint,
		Method: "POST",
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-API-Key":    []string{params.Token},
		},
		Body: jsonBody,
	}
}

// buildBraveRequest builds a request for Brave Search API
// ref: open-sse/handlers/search/callers.js:104-116
func buildBraveRequest(config *ProviderConfig, params *SearchParams) *HTTPRequest {
	endpoint := "/web/search"
	if params.SearchType == "news" {
		endpoint = "/news/search"
	}

	qp := url.Values{}
	qp.Set("q", params.Query)
	qp.Set("count", fmt.Sprintf("%d", params.MaxResults))
	if params.Country != "" {
		qp.Set("country", params.Country)
	}
	if params.Language != "" {
		qp.Set("search_lang", params.Language)
	}

	return &HTTPRequest{
		URL:    resolveBaseURL(config, params) + endpoint + "?" + qp.Encode(),
		Method: "GET",
		Header: http.Header{
			"Accept":                []string{"application/json"},
			"X-Subscription-Token":  []string{params.Token},
		},
	}
}

// buildExaRequest builds a request for Exa API
// ref: open-sse/handlers/search/callers.js:133-153
func buildExaRequest(config *ProviderConfig, params *SearchParams) *HTTPRequest {
	includes, excludes := parseDomainFilter(params.DomainFilter)

	body := map[string]interface{}{
		"query":       params.Query,
		"numResults":  params.MaxResults,
		"type":        "auto",
		"text":        true,
		"highlights":  true,
	}
	if len(includes) > 0 {
		body["includeDomains"] = includes
	}
	if len(excludes) > 0 {
		body["excludeDomains"] = excludes
	}
	if params.SearchType == "news" {
		body["category"] = "news"
	}

	jsonBody, _ := json.Marshal(body)

	return &HTTPRequest{
		URL:    resolveBaseURL(config, params),
		Method: "POST",
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"x-api-key":    []string{params.Token},
		},
		Body: jsonBody,
	}
}

// buildTavilyRequest builds a request for Tavily API
// ref: open-sse/handlers/search/callers.js:155-173
func buildTavilyRequest(config *ProviderConfig, params *SearchParams) *HTTPRequest {
	includes, excludes := parseDomainFilter(params.DomainFilter)

	body := map[string]interface{}{
		"query":       params.Query,
		"max_results": params.MaxResults,
		"topic":       "general",
	}
	if params.SearchType == "news" {
		body["topic"] = "news"
	}
	if len(includes) > 0 {
		body["include_domains"] = includes
	}
	if len(excludes) > 0 {
		body["exclude_domains"] = excludes
	}
	if params.Country != "" {
		body["country"] = params.Country
	}

	jsonBody, _ := json.Marshal(body)

	return &HTTPRequest{
		URL:    resolveBaseURL(config, params),
		Method: "POST",
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer " + params.Token},
		},
		Body: jsonBody,
	}
}

// buildSearxngRequest builds a request for SearXNG API
// ref: open-sse/handlers/search/callers.js:307-328
func buildSearxngRequest(config *ProviderConfig, params *SearchParams) *HTTPRequest {
	baseURL := resolveBaseURL(config, params)
	endpoint := baseURL
	if !strings.HasSuffix(baseURL, "/search") {
		endpoint = baseURL + "/search"
	}

	qp := url.Values{}
	qp.Set("q", params.Query)
	qp.Set("format", "json")
	if params.SearchType == "news" {
		qp.Set("categories", "news")
	} else {
		qp.Set("categories", "general")
	}
	if params.Language != "" {
		qp.Set("language", params.Language)
	}
	if params.TimeRange != "" && params.TimeRange != "any" {
		qp.Set("time_range", params.TimeRange)
	}

	page := toPageNumber(params.Offset, params.MaxResults)
	if page > 0 {
		qp.Set("pageno", fmt.Sprintf("%d", page))
	}

	return &HTTPRequest{
		URL:    endpoint + "?" + qp.Encode(),
		Method: "GET",
		Header: http.Header{
			"Accept": []string{"application/json"},
		},
	}
}

// buildSearchRequest dispatches to the correct provider builder
// ref: open-sse/handlers/search/callers.js:352-371
func buildSearchRequest(config *ProviderConfig, params *SearchParams) *HTTPRequest {
	switch config.ID {
	case "serper":
		return buildSerperRequest(config, params)
	case "brave-search":
		return buildBraveRequest(config, params)
	case "exa":
		return buildExaRequest(config, params)
	case "tavily":
		return buildTavilyRequest(config, params)
	case "searxng":
		return buildSearxngRequest(config, params)
	default:
		// Generic POST with bearer auth
		body := map[string]interface{}{
			"query":        params.Query,
			"max_results":  params.MaxResults,
			"search_type":  params.SearchType,
		}
		jsonBody, _ := json.Marshal(body)
		headers := http.Header{
			"Content-Type": []string{"application/json"},
		}
		if params.Token != "" {
			headers.Set("Authorization", "Bearer "+params.Token)
		}
		return &HTTPRequest{
			URL:    resolveBaseURL(config, params),
			Method: "POST",
			Header: headers,
			Body:   jsonBody,
		}
	}
}

// makeResult creates a unified SearchResult object
// ref: open-sse/handlers/search/normalizers.js:9-32
func makeResult(providerID string, item map[string]interface{}, idx int, now string) Result {
	urlStr, _ := item["url"].(string)
	if urlStr == "" {
		urlStr, _ = item["link"].(string)
	}

	displayURL := ""
	if urlStr != "" {
		// Strip protocol and www prefix
		u, err := url.Parse(urlStr)
		if err == nil {
			displayURL = u.Host + u.Path
			displayURL = strings.TrimPrefix(displayURL, "www.")
		}
	}

	title, _ := item["title"].(string)
	snippet, _ := item["snippet"].(string)
	if snippet == "" {
		snippet, _ = item["description"].(string)
	}

	var score *float64
	if s, ok := item["score"].(float64); ok {
		clamped := s
		if clamped < 0 {
			clamped = 0
		}
		if clamped > 1 {
			clamped = 1
		}
		score = &clamped
	}

	var publishedAt *string
	if p, ok := item["published_at"].(string); ok && p != "" {
		publishedAt = &p
	} else if p, ok := item["publishedDate"].(string); ok && p != "" {
		publishedAt = &p
	} else if p, ok := item["date"].(string); ok && p != "" {
		publishedAt = &p
	} else if p, ok := item["page_age"].(string); ok && p != "" {
		publishedAt = &p
	}

	var faviconURL *string
	if f, ok := item["favicon_url"].(string); ok && f != "" {
		faviconURL = &f
	} else if f, ok := item["favicon"].(string); ok && f != "" {
		faviconURL = &f
	}

	var content *ContentData
	if fullText, ok := item["full_text"].(string); ok && fullText != "" {
		format := "text"
		if f, ok := item["text_format"].(string); ok {
			format = f
		}
		content = &ContentData{
			Format: format,
			Text:   fullText,
			Length: len(fullText),
		}
	}

	var metadata *ResultMetadata
	author, _ := item["author"].(string)
	sourceType, _ := item["source_type"].(string)
	imageURL, _ := item["image_url"].(string)
	if imageURL == "" {
		imageURL, _ = item["image"].(string)
	}
	if author != "" || sourceType != "" || imageURL != "" {
		metadata = &ResultMetadata{
			Author:     author,
			SourceType: sourceType,
			ImageURL:   imageURL,
		}
	}

	return Result{
		Title:       title,
		URL:         urlStr,
		DisplayURL:  displayURL,
		Snippet:     snippet,
		Position:    idx + 1,
		Score:       score,
		PublishedAt: publishedAt,
		FaviconURL:  faviconURL,
		Content:     content,
		Metadata:    metadata,
		Citation: &Citation{
			Provider:    providerID,
			RetrievedAt: now,
			Rank:        idx + 1,
		},
	}
}

// normalizeSerper normalizes Serper API response
// ref: open-sse/handlers/search/normalizers.js:34-43
func normalizeSerper(data map[string]interface{}, searchType string) []Result {
	now := time.Now().UTC().Format(time.RFC3339)

	var items []interface{}
	if searchType == "news" {
		items, _ = data["news"].([]interface{})
	} else {
		items, _ = data["organic"].([]interface{})
	}

	if len(items) == 0 {
		return nil
	}

	results := make([]Result, len(items))
	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		normalized := map[string]interface{}{
			"title":    m["title"],
			"url":      m["link"],
			"snippet":  m["snippet"],
			"date":     m["date"],
		}
		if s, ok := m["description"].(string); ok && normalized["snippet"] == nil {
			normalized["snippet"] = s
		}
		results[i] = makeResult("serper", normalized, i, now)
	}

	return results
}

// normalizeBrave normalizes Brave Search API response
// ref: open-sse/handlers/search/normalizers.js:45-60
func normalizeBrave(data map[string]interface{}, searchType string) []Result {
	now := time.Now().UTC().Format(time.RFC3339)

	var items []interface{}
	if searchType == "news" {
		container, _ := data["news"].(map[string]interface{})
		if container == nil {
			container = data
		}
		items, _ = container["results"].([]interface{})
	} else {
		container, _ := data["web"].(map[string]interface{})
		items, _ = container["results"].([]interface{})
	}

	if len(items) == 0 {
		return nil
	}

	results := make([]Result, len(items))
	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		normalized := map[string]interface{}{
			"title":    m["title"],
			"url":      m["url"],
			"snippet":  m["description"],
			"page_age": m["page_age"],
		}
		if age, ok := m["age"].(string); ok {
			normalized["page_age"] = age
		}
		if meta, ok := m["meta_url"].(map[string]interface{}); ok {
			if f, ok := meta["favicon"].(string); ok {
				normalized["favicon"] = f
			}
		}
		results[i] = makeResult("brave-search", normalized, i, now)
	}

	return results
}

// normalizeExa normalizes Exa API response
// ref: open-sse/handlers/search/normalizers.js:72-91
func normalizeExa(data map[string]interface{}) []Result {
	now := time.Now().UTC().Format(time.RFC3339)

	items, _ := data["results"].([]interface{})
	if len(items) == 0 {
		return nil
	}

	results := make([]Result, len(items))
	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		snippet := ""
		if highlights, ok := m["highlights"].([]interface{}); ok && len(highlights) > 0 {
			if s, ok := highlights[0].(string); ok {
				snippet = s
			}
		}
		if snippet == "" {
			if text, ok := m["text"].(string); ok && len(text) > 300 {
				snippet = text[:300]
			} else if text, ok := m["text"].(string); ok {
				snippet = text
			}
		}
		normalized := map[string]interface{}{
			"title":         m["title"],
			"url":           m["url"],
			"snippet":       snippet,
			"score":         m["score"],
			"publishedDate": m["publishedDate"],
			"favicon":       m["favicon"],
			"author":        m["author"],
			"image":         m["image"],
			"full_text":     m["text"],
			"text_format":   "text",
		}
		results[i] = makeResult("exa", normalized, i, now)
	}

	return results
}

// normalizeTavily normalizes Tavily API response
// ref: open-sse/handlers/search/normalizers.js:93-109
func normalizeTavily(data map[string]interface{}) []Result {
	now := time.Now().UTC().Format(time.RFC3339)

	items, _ := data["results"].([]interface{})
	if len(items) == 0 {
		return nil
	}

	results := make([]Result, len(items))
	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		normalized := map[string]interface{}{
			"title":          m["title"],
			"url":            m["url"],
			"snippet":        m["content"],
			"score":          m["score"],
			"published_date": m["published_date"],
			"full_text":      m["raw_content"],
			"text_format":    "text",
		}
		results[i] = makeResult("tavily", normalized, i, now)
	}

	return results
}

// normalizeSearxng normalizes SearXNG API response
// ref: open-sse/handlers/search/normalizers.js:187-201
func normalizeSearxng(data map[string]interface{}) []Result {
	now := time.Now().UTC().Format(time.RFC3339)

	items, _ := data["results"].([]interface{})
	if len(items) == 0 {
		return nil
	}

	results := make([]Result, len(items))
	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		snippet, _ := m["content"].(string)
		if snippet == "" {
			snippet, _ = m["snippet"].(string)
		}
		normalized := map[string]interface{}{
			"title":          m["title"],
			"url":            m["url"],
			"snippet":        snippet,
			"publishedDate":  m["publishedDate"],
			"source_type":    m["engine"],
		}
		if engines, ok := m["engines"].([]interface{}); ok && len(engines) > 0 {
			var engineStrs []string
			for _, e := range engines {
				if s, ok := e.(string); ok {
					engineStrs = append(engineStrs, s)
				}
			}
			if len(engineStrs) > 0 {
				normalized["source_type"] = strings.Join(engineStrs, ", ")
			}
		}
		results[i] = makeResult("searxng", normalized, i, now)
	}

	return results
}

// normalizeSearchResponse dispatches to the correct normalizer
// ref: open-sse/handlers/search/normalizers.js:220-223
func normalizeSearchResponse(providerID string, data map[string]interface{}, searchType string) []Result {
	switch providerID {
	case "serper":
		return normalizeSerper(data, searchType)
	case "brave-search":
		return normalizeBrave(data, searchType)
	case "exa":
		return normalizeExa(data)
	case "tavily":
		return normalizeTavily(data)
	case "searxng":
		return normalizeSearxng(data)
	default:
		// Try generic normalization
		items, _ := data["results"].([]interface{})
		if len(items) == 0 {
			items, _ = data["organic"].([]interface{})
		}
		if len(items) == 0 {
			return nil
		}
		now := time.Now().UTC().Format(time.RFC3339)
		results := make([]Result, len(items))
		for i, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				results[i] = makeResult(providerID, m, i, now)
			}
		}
		return results
	}
}

// ExecuteSearch executes a search request against a provider
// ref: open-sse/handlers/search/index.js:64-134
func ExecuteSearch(ctx context.Context, config *ProviderConfig, params *SearchParams) (*ExecuteResult, error) {
	if config == nil {
		return nil, fmt.Errorf("no provider configuration")
	}

	if config.AuthType != "none" && params.Token == "" {
		return nil, fmt.Errorf("no credentials for provider: %s", config.ID)
	}

	// Set defaults
	if params.SearchType == "" {
		params.SearchType = config.DefaultSearchType
		if params.SearchType == "" {
			params.SearchType = "web"
		}
	}
	if params.MaxResults <= 0 {
		params.MaxResults = config.DefaultMaxResults
		if params.MaxResults <= 0 {
			params.MaxResults = 5
		}
	}

	// Cap max results
	if config.MaxMaxResults > 0 && params.MaxResults > config.MaxMaxResults {
		params.MaxResults = config.MaxMaxResults
	}

	// Build request
	req := buildSearchRequest(config, params)

	// Create HTTP request
	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if req.Header != nil {
		httpReq.Header = req.Header
	}

	// Execute request
	timeout := config.TimeoutMs
	if timeout <= 0 {
		timeout = 10000
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	startTime := time.Now()
	resp, err := client.Do(httpReq)
	upstreamLatency := time.Since(startTime).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("%s error: %w", config.ID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s returned %d: %s", config.ID, resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Normalize results
	results := normalizeSearchResponse(config.ID, data, params.SearchType)

	// Limit results
	if len(results) > params.MaxResults {
		results = results[:params.MaxResults]
	}

	return &ExecuteResult{
		Results:           results,
		TotalResults:      len(results),
		UpstreamLatencyMs: upstreamLatency,
	}, nil
}
