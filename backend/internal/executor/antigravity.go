// Package executor provides provider-specific request/response handling.
// ref: _ref/9router/open-sse/executors/antigravity.js
package executor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Antigravity constants.
// ref: open-sse/executors/antigravity.js:18-19
const (
	MaxRetryAfterMs           = 10000
	MaxAntigravityOutputTokens = 16384
)

// Antigravity headers constants.
// ref: open-sse/config/appConstants.js
const (
	AntigravityUserAgent      = "antigravity"
	InternalRequestHeaderName = "X-Internal-Request"
	InternalRequestHeaderValue = "true"
	AGToolSuffix              = "_ide"
)

// OAuth endpoints for Google.
// ref: open-sse/config/appConstants.js:161-163
var GoogleOAuthTokenEndpoint = "https://oauth2.googleapis.com/token"

// HTTP status codes.
// ref: open-sse/config/runtimeConfig.js
const (
	HTTPStatusRateLimited     = 429
	HTTPStatusServiceUnavailable = 503
)

// AGDefaultTools is the set of default Antigravity tool names.
// ref: open-sse/config/appConstants.js
var AGDefaultTools = map[string]bool{
	"codebase_search": true,
	"google_search":    true,
	"read_file":        true,
}

// AntigravityExecutor implements the Executor interface for Antigravity API.
// ref: open-sse/executors/antigravity.js:21-24
type AntigravityExecutor struct {
	BaseExecutor
	config      *AntigravityConfig
	baseURLs    []string
	clientID    string
	clientSecret string
}

// AntigravityConfig holds Antigravity-specific configuration.
// ref: open-sse/config/providers.js
type AntigravityConfig struct {
	BaseURLs     []string `json:"baseUrls"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	Headers      map[string]string `json:"headers"`
}

// AntigravityCredentials holds Antigravity authentication data.
// ref: open-sse/executors/antigravity.js:110-140
type AntigravityCredentials struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	ProjectID    string `json:"projectId"`
	Email        string `json:"email"`
	ConnectionID string `json:"connectionId"`
}

// OAuthTokenResponse represents the OAuth token response from Google.
// ref: open-sse/executors/antigravity.js:127-134
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// AntigravityRequest represents a transformed Antigravity request.
// ref: open-sse/executors/antigravity.js:89-108
type AntigravityRequest struct {
	Project     string                   `json:"project"`
	Model       string                   `json:"model"`
	UserAgent   string                   `json:"userAgent"`
	RequestType string                   `json:"requestType"`
	RequestID   string                   `json:"requestId"`
	Request     *AntigravityInnerRequest `json:"request"`
}

// AntigravityInnerRequest represents the inner request body.
// ref: open-sse/executors/antigravity.js:89-108
type AntigravityInnerRequest struct {
	Contents         []AntigravityContent    `json:"contents,omitempty"`
	Tools            []AntigravityToolGroup  `json:"tools,omitempty"`
	GenerationConfig map[string]interface{}  `json:"generationConfig,omitempty"`
	SessionID        string                  `json:"sessionId,omitempty"`
	SafetySettings   interface{}             `json:"safetySettings,omitempty"`
	ToolConfig       *AntigravityToolConfig  `json:"toolConfig,omitempty"`
}

// AntigravityContent represents a message content.
type AntigravityContent struct {
	Role  string             `json:"role"`
	Parts []AntigravityPart  `json:"parts"`
}

// AntigravityPart represents a part of content.
type AntigravityPart struct {
	Text            string                   `json:"text,omitempty"`
	FunctionCall    *AntigravityFunctionCall `json:"functionCall,omitempty"`
	FunctionResponse *AntigravityFunctionResponse `json:"functionResponse,omitempty"`
	Thought         bool                     `json:"thought,omitempty"`
	ThoughtSignature string                  `json:"thoughtSignature,omitempty"`
}

// AntigravityFunctionCall represents a function call.
type AntigravityFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// AntigravityFunctionResponse represents a function response.
type AntigravityFunctionResponse struct {
	Name    string      `json:"name"`
	Response interface{} `json:"response"`
}

// AntigravityToolGroup represents a group of function declarations.
type AntigravityToolGroup struct {
	FunctionDeclarations []AntigravityFunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// AntigravityFunctionDeclaration represents a function declaration.
type AntigravityFunctionDeclaration struct {
	Name       string                 `json:"name"`
	Description string                `json:"description,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// AntigravityToolConfig represents the tool configuration.
type AntigravityToolConfig struct {
	FunctionCallingConfig *AntigravityFunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

// AntigravityFunctionCallingConfig represents function calling config.
type AntigravityFunctionCallingConfig struct {
	Mode string `json:"mode"`
}

// NewAntigravityExecutor creates a new Antigravity executor.
// ref: open-sse/executors/antigravity.js:22-24
func NewAntigravityExecutor(config *AntigravityConfig) *AntigravityExecutor {
	if config == nil {
		config = &AntigravityConfig{
			BaseURLs: []string{
				"https://cloudcode-pa.googleapis.com",
			},
			ClientID:     "GOOGLE_OAUTH_CLIENT_ID",
			ClientSecret: "GOOGLE_OAUTH_CLIENT_SECRET",
		}
	}
	return &AntigravityExecutor{
		BaseExecutor:  NewBaseExecutor("antigravity"),
		config:        config,
		baseURLs:      config.BaseURLs,
		clientID:      config.ClientID,
		clientSecret:  config.ClientSecret,
	}
}

// getBaseURLs returns the base URLs for the executor.
// ref: open-sse/executors/antigravity.js:27-28
func (e *AntigravityExecutor) getBaseURLs() []string {
	if len(e.baseURLs) == 0 {
		return []string{"https://cloudcode-pa.googleapis.com"}
	}
	return e.baseURLs
}

// BuildURL builds the request URL.
// ref: open-sse/executors/antigravity.js:26-31
func (e *AntigravityExecutor) BuildURL(model string, stream bool, urlIndex int) string {
	baseURLs := e.getBaseURLs()
	baseURL := baseURLs[0]
	if urlIndex < len(baseURLs) {
		baseURL = baseURLs[urlIndex]
	}

	action := "generateContent"
	if stream {
		action = "streamGenerateContent?alt=sse"
	}

	return fmt.Sprintf("%s/v1internal:%s", baseURL, action)
}

// BuildHeaders builds the request headers.
// ref: open-sse/executors/antigravity.js:33-42
func (e *AntigravityExecutor) BuildHeaders(credentials *AntigravityCredentials, stream bool, sessionID string) http.Header {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", credentials.AccessToken))

	userAgent := AntigravityUserAgent
	if e.config != nil && e.config.Headers != nil {
		if ua, ok := e.config.Headers["User-Agent"]; ok {
			userAgent = ua
		}
	}
	headers.Set("User-Agent", userAgent)
	headers.Set(InternalRequestHeaderName, InternalRequestHeaderValue)

	if sessionID != "" {
		headers.Set("X-Machine-Session-Id", sessionID)
	}

	if stream {
		headers.Set("Accept", "text/event-stream")
	} else {
		headers.Set("Accept", "application/json")
	}

	return headers
}

// generateProjectId generates a random project ID.
// ref: open-sse/executors/antigravity.js:142-146
func (e *AntigravityExecutor) generateProjectId() string {
	adjectives := []string{"useful", "bright", "swift", "calm", "bold"}
	nouns := []string{"fuze", "wave", "spark", "flow", "core"}

	adjIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	nounIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))

	uuid := generateUUID()
	shortUUID := uuid[:5]

	return fmt.Sprintf("%s-%s-%s", adjectives[adjIdx.Int64()], nouns[nounIdx.Int64()], shortUUID)
}

// generateSessionId generates a session ID.
// ref: open-sse/executors/antigravity.js:148-150
func generateSessionId() string {
	return generateUUID() + fmt.Sprintf("%d", time.Now().UnixMilli())
}

// generateUUID generates a UUID v4 string.
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// SanitizeFunctionName sanitizes a function name for Gemini compatibility.
// Gemini requires [a-zA-Z_][a-zA-Z0-9_.:\\-]{0,63}
// ref: open-sse/executors/antigravity.js:11-16
func SanitizeFunctionName(name string) string {
	if name == "" {
		return "_unknown"
	}

	// Replace invalid characters with underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9_.:\-]`)
	s := re.ReplaceAllString(name, "_")

	// Ensure starts with letter or underscore
	if len(s) > 0 && !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(s) {
		s = "_" + s
	}

	// Truncate to 64 characters
	if len(s) > 64 {
		s = s[:64]
	}

	return s
}

// DeriveSessionId derives a session ID from email or connection ID.
// ref: open-sse/utils/sessionManager.js
func DeriveSessionId(email, connectionID string) string {
	if email != "" {
		return email
	}
	if connectionID != "" {
		return connectionID
	}
	return generateSessionId()
}

// TransformRequest transforms a request body for Antigravity.
// ref: open-sse/executors/antigravity.js:44-108
func (e *AntigravityExecutor) TransformRequest(model string, body map[string]interface{}, stream bool, credentials *AntigravityCredentials) map[string]interface{} {
	projectID := ""
	if credentials != nil {
		projectID = credentials.ProjectID
	}
	if projectID == "" {
		projectID = e.generateProjectId()
	}

	request, _ := body["request"].(map[string]interface{})

	contents := e.transformContents(request)
	tools := e.transformTools(request)
	generationConfig := e.transformGenerationConfig(request)

	sessionID := ""
	if sid, ok := request["sessionId"].(string); ok && sid != "" {
		sessionID = sid
	} else if credentials != nil {
		sessionID = DeriveSessionId(credentials.Email, credentials.ConnectionID)
	}

	transformedRequest := make(map[string]interface{})
	for k, v := range request {
		if k != "tools" && k != "toolConfig" && k != "safetySettings" {
			transformedRequest[k] = v
		}
	}
	transformedRequest["generationConfig"] = generationConfig
	if contents != nil {
		transformedRequest["contents"] = contents
	}
	if tools != nil {
		transformedRequest["tools"] = tools
	}
	transformedRequest["sessionId"] = sessionID
	if tools != nil && len(tools) > 0 {
		transformedRequest["toolConfig"] = map[string]interface{}{
			"functionCallingConfig": map[string]interface{}{
				"mode": "VALIDATED",
			},
		}
	}

	return map[string]interface{}{
		"project":     projectID,
		"model":       model,
		"userAgent":   "antigravity",
		"requestType": "agent",
		"requestId":   fmt.Sprintf("agent-%s", generateUUID()),
		"request":     transformedRequest,
	}
}

func (e *AntigravityExecutor) transformContents(request map[string]interface{}) []map[string]interface{} {
	contents, ok := request["contents"].([]interface{})
	if !ok {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(contents))
	for _, c := range contents {
		content, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := content["role"].(string)
		parts, _ := content["parts"].([]interface{})

		hasFunctionResponse := false
		for _, p := range parts {
			if part, ok := p.(map[string]interface{}); ok {
				if _, has := part["functionResponse"]; has {
					hasFunctionResponse = true
					break
				}
			}
		}
		if hasFunctionResponse {
			role = "user"
		}

		filteredParts := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			part, ok := p.(map[string]interface{})
			if !ok {
				filteredParts = append(filteredParts, p)
				continue
			}

			thought, _ := part["thought"].(bool)
			_, hasFunctionCall := part["functionCall"]
			thoughtSignature, _ := part["thoughtSignature"].(string)
			_, hasText := part["text"]

			if thought && !hasFunctionCall {
				continue
			}
			if thoughtSignature != "" && !hasFunctionCall && !hasText {
				continue
			}
			filteredParts = append(filteredParts, p)
		}

		if role != content["role"] || len(filteredParts) != len(parts) {
			result = append(result, map[string]interface{}{
				"role":  role,
				"parts": filteredParts,
			})
		} else {
			result = append(result, content)
		}
	}
	return result
}

func (e *AntigravityExecutor) transformTools(request map[string]interface{}) []map[string]interface{} {
	tools, ok := request["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		return nil
	}

	allDeclarations := make([]map[string]interface{}, 0)
	for _, t := range tools {
		tool, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		declarations, _ := tool["functionDeclarations"].([]interface{})
		for _, d := range declarations {
			fn, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := fn["name"].(string)
			params := fn["parameters"]
			if params == nil {
				params = map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"reason": map[string]interface{}{
							"type":        "string",
							"description": "Brief explanation",
						},
					},
					"required": []string{"reason"},
				}
			} else {
				params = CleanJSONSchemaForAntigravity(params)
			}
			allDeclarations = append(allDeclarations, map[string]interface{}{
				"name":       SanitizeFunctionName(name),
				"description": fn["description"],
				"parameters":  params,
			})
		}
	}

	if len(allDeclarations) == 0 {
		return nil
	}
	return []map[string]interface{}{
		{"functionDeclarations": allDeclarations},
	}
}

func (e *AntigravityExecutor) transformGenerationConfig(request map[string]interface{}) map[string]interface{} {
	gc, _ := request["generationConfig"].(map[string]interface{})
	if gc == nil {
		gc = make(map[string]interface{})
	} else {
		gc = copyMap(gc)
	}
	if maxTokens, ok := gc["maxOutputTokens"].(float64); ok && maxTokens > MaxAntigravityOutputTokens {
		gc["maxOutputTokens"] = MaxAntigravityOutputTokens
	}
	return gc
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// CleanJSONSchemaForAntigravity cleans JSON schema for Antigravity compatibility.
// ref: open-sse/translator/helpers/geminiHelper.js
func CleanJSONSchemaForAntigravity(schema interface{}) interface{} {
	s, ok := schema.(map[string]interface{})
	if !ok {
		return schema
	}

	result := make(map[string]interface{})
	for k, v := range s {
		if k == "$schema" || k == "additionalProperties" {
			continue
		}
		if k == "type" {
			if vt, ok := v.(string); ok {
				result["type"] = strings.ToUpper(vt)
			} else {
				result[k] = v
			}
			continue
		}
		if k == "properties" {
			props, ok := v.(map[string]interface{})
			if ok {
				cleanedProps := make(map[string]interface{})
				for pk, pv := range props {
					cleanedProps[pk] = CleanJSONSchemaForAntigravity(pv)
				}
				result["properties"] = cleanedProps
			} else {
				result[k] = v
			}
			continue
		}
		if k == "items" {
			result["items"] = CleanJSONSchemaForAntigravity(v)
			continue
		}
		if k == "anyOf" || k == "oneOf" || k == "allOf" {
			arr, ok := v.([]interface{})
			if ok {
				cleanedArr := make([]interface{}, len(arr))
				for i, item := range arr {
					cleanedArr[i] = CleanJSONSchemaForAntigravity(item)
				}
				result[k] = cleanedArr
			} else {
				result[k] = v
			}
			continue
		}
		result[k] = v
	}
	return result
}

// RefreshCredentials refreshes OAuth credentials.
// ref: open-sse/executors/antigravity.js:110-140
func (e *AntigravityExecutor) RefreshCredentials(ctx context.Context, credentials *AntigravityCredentials, log func(string, string)) *AntigravityCredentials {
	if credentials.RefreshToken == "" {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", credentials.RefreshToken)
	data.Set("client_id", e.clientID)
	data.Set("client_secret", e.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", GoogleOAuthTokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		if log != nil {
			log("ERROR", fmt.Sprintf("Antigravity refresh error: %v", err))
		}
		return nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if log != nil {
			log("ERROR", fmt.Sprintf("Antigravity refresh error: %v", err))
		}
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var tokens OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil
	}

	if log != nil {
		log("TOKEN", "Antigravity refreshed")
	}

	refreshToken := tokens.RefreshToken
	if refreshToken == "" {
		refreshToken = credentials.RefreshToken
	}

	return &AntigravityCredentials{
		AccessToken:  tokens.AccessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		ProjectID:    credentials.ProjectID,
	}
}

// ParseRetryHeaders parses retry headers from response.
// ref: open-sse/executors/antigravity.js:152-181
func ParseRetryHeaders(headers http.Header) *time.Duration {
	retryAfter := headers.Get("retry-after")
	if retryAfter != "" {
		seconds, err := strconv.Atoi(retryAfter)
		if err == nil && seconds > 0 {
			d := time.Duration(seconds) * time.Second
			return &d
		}

		t, err := http.ParseTime(retryAfter)
		if err == nil {
			d := time.Until(t)
			if d > 0 {
				return &d
			}
		}
	}

	resetAfter := headers.Get("x-ratelimit-reset-after")
	if resetAfter != "" {
		seconds, err := strconv.Atoi(resetAfter)
		if err == nil && seconds > 0 {
			d := time.Duration(seconds) * time.Second
			return &d
		}
	}

	resetTimestamp := headers.Get("x-ratelimit-reset")
	if resetTimestamp != "" {
		ts, err := strconv.ParseInt(resetTimestamp, 10, 64)
		if err == nil {
			t := time.Unix(ts, 0)
			d := time.Until(t)
			if d > 0 {
				return &d
			}
		}
	}

	return nil
}

// ParseRetryFromErrorMessage parses retry time from error message.
// ref: open-sse/executors/antigravity.js:183-197
func ParseRetryFromErrorMessage(errorMessage string) *time.Duration {
	if errorMessage == "" {
		return nil
	}

	re := regexp.MustCompile(`reset after (\d+h)?(\d+m)?(\d+s)?`)
	matches := re.FindStringSubmatch(strings.ToLower(errorMessage))
	if matches == nil {
		return nil
	}

	var totalMs int64
	if matches[1] != "" {
		hours, _ := strconv.Atoi(strings.TrimSuffix(matches[1], "h"))
		totalMs += int64(hours) * 3600 * 1000
	}
	if matches[2] != "" {
		minutes, _ := strconv.Atoi(strings.TrimSuffix(matches[2], "m"))
		totalMs += int64(minutes) * 60 * 1000
	}
	if matches[3] != "" {
		seconds, _ := strconv.Atoi(strings.TrimSuffix(matches[3], "s"))
		totalMs += int64(seconds) * 1000
	}

	if totalMs > 0 {
		d := time.Duration(totalMs) * time.Millisecond
		return &d
	}
	return nil
}

// PrepareRequest implements Executor interface.
// ref: open-sse/executors/antigravity.js
func (e *AntigravityExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	return nil
}

// TransformResponse implements Executor interface.
// ref: open-sse/executors/antigravity.js
func (e *AntigravityExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// HandleError implements Executor interface.
// ref: open-sse/executors/antigravity.js
func (e *AntigravityExecutor) HandleError(ctx context.Context, err error) error {
	return err
}

// ExecuteResult represents the result of an Execute call.
// ref: open-sse/executors/antigravity.js:279
type ExecuteResult struct {
	Response       *http.Response
	URL            string
	Headers        http.Header
	TransformedBody map[string]interface{}
}

// Execute executes the Antigravity request with retry logic.
// ref: open-sse/executors/antigravity.js:199-291
func (e *AntigravityExecutor) Execute(ctx context.Context, model string, body map[string]interface{}, stream bool, credentials *AntigravityCredentials, log func(string, string)) (*ExecuteResult, error) {
	fallbackCount := len(e.getBaseURLs())
	if fallbackCount == 0 {
		fallbackCount = 1
	}

	var lastError error
	var lastStatus int
	const maxAutoRetries = 3
	const maxRetryAfterRetries = 3

	retryAttemptsByURL := make(map[int]int)
	retryAfterAttemptsByURL := make(map[int]int)

	for urlIndex := 0; urlIndex < fallbackCount; urlIndex++ {
		urlStr := e.BuildURL(model, stream, urlIndex)
		transformedBody := e.TransformRequest(model, body, stream, credentials)
		sessionID, _ := transformedBody["request"].(map[string]interface{})["sessionId"].(string)
		headers := e.BuildHeaders(credentials, stream, sessionID)

		if _, ok := retryAttemptsByURL[urlIndex]; !ok {
			retryAttemptsByURL[urlIndex] = 0
		}
		if _, ok := retryAfterAttemptsByURL[urlIndex]; !ok {
			retryAfterAttemptsByURL[urlIndex] = 0
		}

		bodyBytes, err := json.Marshal(transformedBody)
		if err != nil {
			lastError = err
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(string(bodyBytes)))
		if err != nil {
			lastError = err
			continue
		}
		req.Header = headers

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastError = err
			if urlIndex+1 < fallbackCount {
				if log != nil {
					log("RETRY", fmt.Sprintf("Error on %s, trying fallback %d", urlStr, urlIndex+1))
				}
				continue
			}
			return nil, err
		}

		if resp.StatusCode == HTTPStatusRateLimited || resp.StatusCode == HTTPStatusServiceUnavailable {
			var retryMs *time.Duration
			retryMs = ParseRetryHeaders(resp.Header)

			if retryMs == nil {
				bodyBytes, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				var errorJson map[string]interface{}
				if json.Unmarshal(bodyBytes, &errorJson) == nil {
					errorMsg := ""
					if em, ok := errorJson["error"].(map[string]interface{}); ok {
						errorMsg, _ = em["message"].(string)
					} else if em, ok := errorJson["message"].(string); ok {
						errorMsg = em
					}
					retryMs = ParseRetryFromErrorMessage(errorMsg)
				}
			} else {
				resp.Body.Close()
			}

			if retryMs != nil && *retryMs <= MaxRetryAfterMs*time.Millisecond && retryAfterAttemptsByURL[urlIndex] < maxRetryAfterRetries {
				retryAfterAttemptsByURL[urlIndex]++
				if log != nil {
					log("RETRY", fmt.Sprintf("%d with Retry-After: %ds, waiting... (%d/%d)", resp.StatusCode, int(retryMs.Seconds()), retryAfterAttemptsByURL[urlIndex], maxRetryAfterRetries))
				}
				time.Sleep(*retryMs)
				urlIndex--
				continue
			}

			if resp.StatusCode == HTTPStatusRateLimited && (retryMs == nil || *retryMs == 0) && retryAttemptsByURL[urlIndex] < maxAutoRetries {
				retryAttemptsByURL[urlIndex]++
				backoffMs := time.Duration(math.Min(float64(1000)*math.Pow(2, float64(retryAttemptsByURL[urlIndex])), float64(MaxRetryAfterMs))) * time.Millisecond
				if log != nil {
					log("RETRY", fmt.Sprintf("429 auto retry %d/%d after %ds", retryAttemptsByURL[urlIndex], maxAutoRetries, int(backoffMs.Seconds())))
				}
				time.Sleep(backoffMs)
				urlIndex--
				continue
			}

			if log != nil {
				retryMsg := "missing"
				if retryMs != nil {
					retryMsg = fmt.Sprintf("too long (%ds)", int(retryMs.Seconds()))
				}
				log("RETRY", fmt.Sprintf("%d, Retry-After %s, trying fallback", resp.StatusCode, retryMsg))
			}
			lastStatus = resp.StatusCode
			resp.Body.Close()

			if urlIndex+1 < fallbackCount {
				continue
			}
		}

		if e.shouldRetry(resp.StatusCode, urlIndex, fallbackCount) {
			if log != nil {
				log("RETRY", fmt.Sprintf("%d on %s, trying fallback %d", resp.StatusCode, urlStr, urlIndex+1))
			}
			lastStatus = resp.StatusCode
			resp.Body.Close()
			continue
		}

		return &ExecuteResult{
			Response:        resp,
			URL:             urlStr,
			Headers:         headers,
			TransformedBody: transformedBody,
		}, nil
	}

	if lastError != nil {
		return nil, lastError
	}
	return nil, fmt.Errorf("all %d URLs failed with status %d", fallbackCount, lastStatus)
}

func (e *AntigravityExecutor) shouldRetry(status int, urlIndex int, fallbackCount int) bool {
	retryableStatuses := map[int]bool{
		500: true,
		502: true,
		503: true,
		504: true,
	}
	return retryableStatuses[status] && urlIndex+1 < fallbackCount
}

func (e *AntigravityExecutor) getFallbackCount() int {
	return len(e.getBaseURLs())
}

// CloakTools renames client tools with _ide suffix and injects AG decoy tools.
// ref: open-sse/executors/antigravity.js:299-392
func CloakTools(body map[string]interface{}, clientTool string) (map[string]interface{}, map[string]string) {
	request, _ := body["request"].(map[string]interface{})
	tools, _ := request["tools"].([]interface{})
	if tools == nil || len(tools) == 0 {
		return body, nil
	}

	isCopilot := clientTool == "github-copilot"
	toolNameMap := make(map[string]string)
	clientDeclarations := make([]map[string]interface{}, 0)
	decoyNames := make(map[string]bool)
	for _, t := range AGDecoyTools {
		decoyNames[t["name"].(string)] = true
	}

	for _, t := range tools {
		tool, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		declarations, _ := tool["functionDeclarations"].([]interface{})
		for _, d := range declarations {
			fn, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := fn["name"].(string)

			if isCopilot && AGDefaultTools[name] {
				continue
			}
			if isCopilot && decoyNames[name] {
				continue
			}
			if AGDefaultTools[name] {
				clientDeclarations = append(clientDeclarations, fn)
				continue
			}

			suffixed := name + AGToolSuffix
			toolNameMap[suffixed] = name
			copiedFn := copyMap(fn)
			copiedFn["name"] = suffixed
			clientDeclarations = append(clientDeclarations, copiedFn)
		}
	}

	allDeclarations := make([]map[string]interface{}, 0)
	seenNames := make(map[string]bool)
	for _, decl := range append(clientDeclarations, AGDecoyTools...) {
		name, _ := decl["name"].(string)
		if name == "" || seenNames[name] {
			continue
		}
		seenNames[name] = true
		allDeclarations = append(allDeclarations, decl)
	}

	cloakedContents := eCloakContents(request, toolNameMap)

	cloakedRequest := copyMap(request)
	cloakedRequest["tools"] = []interface{}{
		map[string]interface{}{"functionDeclarations": allDeclarations},
	}
	if cloakedContents != nil {
		cloakedRequest["contents"] = cloakedContents
	}

	cloakedBody := copyMap(body)
	cloakedBody["request"] = cloakedRequest

	return cloakedBody, toolNameMap
}

func eCloakContents(request map[string]interface{}, toolNameMap map[string]string) []interface{} {
	contents, _ := request["contents"].([]interface{})
	if contents == nil {
		return nil
	}

	cloaked := make([]interface{}, 0, len(contents))
	for _, c := range contents {
		content, ok := c.(map[string]interface{})
		if !ok {
			cloaked = append(cloaked, c)
			continue
		}

		parts, _ := content["parts"].([]interface{})
		if parts == nil {
			cloaked = append(cloaked, content)
			continue
		}

		cloakedParts := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			part, ok := p.(map[string]interface{})
			if !ok {
				cloakedParts = append(cloakedParts, p)
				continue
			}

			if fc, ok := part["functionCall"].(map[string]interface{}); ok {
				fcName, _ := fc["name"].(string)
				if !AGDefaultTools[fcName] {
					copiedPart := copyMap(part)
					copiedFC := copyMap(fc)
					copiedFC["name"] = fcName + AGToolSuffix
					copiedPart["functionCall"] = copiedFC
					cloakedParts = append(cloakedParts, copiedPart)
					continue
				}
			}

			if fr, ok := part["functionResponse"].(map[string]interface{}); ok {
				frName, _ := fr["name"].(string)
				if !AGDefaultTools[frName] {
					copiedPart := copyMap(part)
					copiedFR := copyMap(fr)
					copiedFR["name"] = frName + AGToolSuffix
					copiedPart["functionResponse"] = copiedFR
					cloakedParts = append(cloakedParts, copiedPart)
					continue
				}
			}

			cloakedParts = append(cloakedParts, p)
		}

		cloaked = append(cloaked, map[string]interface{}{
			"role":  content["role"],
			"parts": cloakedParts,
		})
	}
	return cloaked
}

// AGDecoyTools is the list of AG decoy tools.
// ref: open-sse/executors/antigravity.js:396-502
var AGDecoyTools = []map[string]interface{}{
}
