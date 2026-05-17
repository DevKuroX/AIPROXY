// Package executor provides provider-specific request/response handling.
// ref: _ref/9router/open-sse/executors/vertex.js
package executor

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// VertexExecutor handles Google Vertex AI request transformations.
// ref: open-sse/executors/vertex.js:39
type VertexExecutor struct {
	BaseExecutor
}

// NewVertexExecutor creates a new Vertex executor.
// ref: open-sse/executors/vertex.js:40-42
func NewVertexExecutor(provider string) *VertexExecutor {
	if provider == "" {
		provider = "vertex"
	}
	return &VertexExecutor{
		BaseExecutor: NewBaseExecutor(provider),
	}
}

// VertexCredentials holds Vertex-specific credential data.
// ref: open-sse/executors/vertex.js:44-72
type VertexCredentials struct {
	APIKey               string
	AccessToken          string
	ProjectID            string
	Location             string
	ProviderSpecificData map[string]interface{}
}

// ServiceAccountJSON represents a GCP Service Account JSON key.
// ref: open-sse/services/tokenRefresh.js:702-713
type ServiceAccountJSON struct {
	Type        string `json:"type"`
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
	TokenURI    string `json:"token_uri"`
}

// VertexTokenCache caches Vertex tokens keyed by service account email.
// ref: open-sse/services/tokenRefresh.js:716
type VertexTokenCache struct {
	mu      sync.RWMutex
	tokens  map[string]*cachedToken
	expired time.Duration // buffer before expiry
}

type cachedToken struct {
	token     string
	expiresAt time.Time
}

// vertexTokenCache is the global token cache.
// ref: open-sse/services/tokenRefresh.js:716
var vertexTokenCache = &VertexTokenCache{
	tokens:  make(map[string]*cachedToken),
	expired: 5 * time.Minute,
}

// projectIdCache caches project IDs resolved from raw API keys.
// ref: open-sse/executors/vertex.js:7
var projectIdCache = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

// parseVertexSaJson parses Vertex AI Service Account JSON from apiKey string.
// ref: open-sse/services/tokenRefresh.js:702-713
func parseVertexSaJson(apiKey string) *ServiceAccountJSON {
	if apiKey == "" {
		return nil
	}
	var sa ServiceAccountJSON
	if err := json.Unmarshal([]byte(apiKey), &sa); err != nil {
		return nil
	}
	if sa.Type != "service_account" || sa.ClientEmail == "" || sa.PrivateKey == "" || sa.ProjectID == "" {
		return nil
	}
	return &sa
}

// RefreshVertexToken mints a short-lived OAuth2 Bearer token for Google Cloud Vertex AI
// using Service Account JSON + RS256 JWT assertion flow.
// ref: open-sse/services/tokenRefresh.js:723-772
func RefreshVertexToken(ctx context.Context, sa *ServiceAccountJSON) (string, error) {
	cacheKey := sa.ClientEmail

	// Check cache first
	// ref: open-sse/services/tokenRefresh.js:725-730
	vertexTokenCache.mu.RLock()
	cached, ok := vertexTokenCache.tokens[cacheKey]
	vertexTokenCache.mu.RUnlock()

	if ok && time.Until(cached.expiresAt) > vertexTokenCache.expired {
		return cached.token, nil
	}

	// Parse private key
	// ref: open-sse/services/tokenRefresh.js:735
	privateKeyPEM := strings.ReplaceAll(sa.PrivateKey, "\\n", "\n")
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA")
	}

	// Build JWT assertion
	// ref: open-sse/services/tokenRefresh.js:738-744
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"scope": "https://www.googleapis.com/auth/cloud-platform",
		"iss":   sa.ClientEmail,
		"aud":   "https://oauth2.googleapis.com/token",
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	})

	jwtStr, err := token.SignedString(rsaKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Exchange JWT for access token
	// ref: open-sse/services/tokenRefresh.js:746-761
	tokenURL := "https://oauth2.googleapis.com/token"
	if sa.TokenURI != "" {
		tokenURL = sa.TokenURI
	}

	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	formData.Set("assertion", jwtStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	// Cache the token
	// ref: open-sse/services/tokenRefresh.js:764
	vertexTokenCache.mu.Lock()
	vertexTokenCache.tokens[cacheKey] = &cachedToken{
		token:     tokenResp.AccessToken,
		expiresAt: expiresAt,
	}
	vertexTokenCache.mu.Unlock()

	return tokenResp.AccessToken, nil
}

// resolveProjectId resolves GCP project ID from a raw Vertex API key.
// Sends a dummy 404 request and parses "projects/{id}" from the error message.
// ref: open-sse/executors/vertex.js:13-27
func resolveProjectId(ctx context.Context, apiKey string) (string, error) {
	// Check cache first
	projectIdCache.RLock()
	if id, ok := projectIdCache.m[apiKey]; ok {
		projectIdCache.RUnlock()
		return id, nil
	}
	projectIdCache.RUnlock()

	// Probe endpoint
	probeURL := fmt.Sprintf("https://aiplatform.googleapis.com/v1/publishers/google/models/__probe__:generateContent?key=%s", apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, probeURL, strings.NewReader("{}"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse error message to extract project ID
	// ref: open-sse/executors/vertex.js:21-23
	var errResp struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &errResp)

	msg := ""
	if errResp.Error != nil {
		msg = errResp.Error.Message
	}

	// Extract project ID from "projects/{id}/" pattern
	// ref: open-sse/executors/vertex.js:22-23
	idx := strings.Index(msg, "projects/")
	if idx == -1 {
		return "", fmt.Errorf("could not resolve project_id from API key")
	}
	rest := msg[idx+9:]
	endIdx := strings.Index(rest, "/")
	if endIdx == -1 {
		return "", fmt.Errorf("could not resolve project_id from API key")
	}
	projectID := rest[:endIdx]

	// Cache result
	// ref: open-sse/executors/vertex.js:25
	projectIdCache.Lock()
	projectIdCache.m[apiKey] = projectID
	projectIdCache.Unlock()

	return projectID, nil
}

// buildURL constructs the Vertex API URL based on the provider type and credentials.
// ref: open-sse/executors/vertex.js:44-73
func (e *VertexExecutor) buildURL(model string, stream bool, creds *VertexCredentials) (string, error) {
	sa := parseVertexSaJson(creds.APIKey)
	rawKey := ""
	if sa == nil {
		rawKey = creds.APIKey
	}

	projectID := ""
	if sa != nil {
		projectID = sa.ProjectID
	}
	if creds.ProjectID != "" {
		projectID = creds.ProjectID
	}

	// vertex-partner: OpenAI-compatible endpoint
	// ref: open-sse/executors/vertex.js:49-54
	if e.Provider() == "vertex-partner" {
		if projectID == "" {
			return "", fmt.Errorf("vertex partner models require a project_id")
		}
		baseURL := fmt.Sprintf("https://aiplatform.googleapis.com/v1/projects/%s/locations/global/endpoints/openapi/chat/completions", projectID)
		if rawKey != "" {
			return baseURL + "?key=" + rawKey, nil
		}
		return baseURL, nil
	}

	// Gemini on Vertex
	// ref: open-sse/executors/vertex.js:56-72
	action := "generateContent"
	if stream {
		action = "streamGenerateContent"
	}

	// SA JSON: must use project-scoped path
	// ref: open-sse/executors/vertex.js:59-64
	if sa != nil {
		location := "us-central1"
		if loc, ok := creds.ProviderSpecificData["location"].(string); ok && loc != "" {
			location = loc
		}
		urlStr := fmt.Sprintf("https://aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:%s", projectID, location, model, action)
		if stream {
			urlStr += "?alt=sse"
		}
		return urlStr, nil
	}

	// Raw API key: use global publishers endpoint with ?key= param
	// ref: open-sse/executors/vertex.js:67-72
	urlStr := fmt.Sprintf("https://aiplatform.googleapis.com/v1/publishers/google/models/%s:%s", model, action)
	if stream {
		urlStr += "?alt=sse"
		if rawKey != "" {
			urlStr += "&key=" + rawKey
		}
	} else if rawKey != "" {
		urlStr += "?key=" + rawKey
	}
	return urlStr, nil
}

// VertexCredentialsCtxKey is the context key for Vertex credentials.
type VertexCredentialsCtxKey struct{}

// extractVertexCredentials extracts Vertex credentials from context.
func extractVertexCredentials(ctx context.Context) *VertexCredentials {
	if creds, ok := ctx.Value(VertexCredentialsCtxKey{}).(*VertexCredentials); ok {
		return creds
	}
	return &VertexCredentials{}
}

// PrepareRequest transforms the request for Vertex AI API.
// ref: open-sse/executors/vertex.js:98-128
func (e *VertexExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	creds := extractVertexCredentials(ctx)
	sa := parseVertexSaJson(creds.APIKey)

	// Handle SA JSON auth: mint Bearer token
	// ref: open-sse/executors/vertex.js:101-106
	if sa != nil {
		accessToken, err := RefreshVertexToken(ctx, sa)
		if err != nil {
			return fmt.Errorf("vertex: failed to mint access token: %w", err)
		}
		creds.AccessToken = accessToken
	}

	// vertex-partner with raw key: auto-resolve project_id if not provided
	// ref: open-sse/executors/vertex.js:109-114
	if e.Provider() == "vertex-partner" && sa == nil {
		if projID, ok := creds.ProviderSpecificData["projectId"].(string); !ok || projID == "" {
			projectID, err := resolveProjectId(ctx, creds.APIKey)
			if err != nil {
				return fmt.Errorf("vertex: could not resolve project_id from API key: %w", err)
			}
			if creds.ProviderSpecificData == nil {
				creds.ProviderSpecificData = make(map[string]interface{})
			}
			creds.ProviderSpecificData["projectId"] = projectID
			creds.ProjectID = projectID
		}
	}

	// Extract model from path or use default
	model := "gemini-pro"
	pathParts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	for i, p := range pathParts {
		if p == "models" && i+1 < len(pathParts) {
			model = pathParts[i+1]
			break
		}
	}

	// Detect streaming from request body or Accept header
	stream := strings.Contains(req.Header.Get("Accept"), "text/event-stream")

	// Build Vertex URL
	// ref: open-sse/executors/vertex.js:116
	vertexURL, err := e.buildURL(model, stream, creds)
	if err != nil {
		return err
	}

	req.URL, err = url.Parse(vertexURL)
	if err != nil {
		return fmt.Errorf("failed to parse Vertex URL: %w", err)
	}

	// Set headers
	// ref: open-sse/executors/vertex.js:75-86
	req.Header.Set("Content-Type", "application/json")

	if creds.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+creds.AccessToken)
	}

	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	return nil
}

// TransformResponse reads and returns the response body unchanged.
func (e *VertexExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (e *VertexExecutor) HandleError(ctx context.Context, err error) error {
	return err
}

func init() {
	// Register both vertex and vertex-partner executors
	if err := Register("vertex", NewVertexExecutor("vertex")); err != nil {
		log.Fatalf("failed to register vertex executor: %v", err)
	}
	if err := Register("vertex-partner", NewVertexExecutor("vertex-partner")); err != nil {
		log.Fatalf("failed to register vertex-partner executor: %v", err)
	}
}
