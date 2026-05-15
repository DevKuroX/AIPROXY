package router

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/pool"
	"github.com/DevKuroX/AIPROXY/internal/providers"
	"github.com/DevKuroX/AIPROXY/internal/rtk"
	reqtrans "github.com/DevKuroX/AIPROXY/internal/translator/request"
	resptrans "github.com/DevKuroX/AIPROXY/internal/translator/response"
)

var globalPool *pool.Pool
var globalProxyAPI interface{}
var globalSettingsGetter func(key string) string

func SetGlobalPool(p *pool.Pool) {
	globalPool = p
}

func GetGlobalPool() *pool.Pool {
	return globalPool
}

func SetProxyAPI(api interface{}) {
	globalProxyAPI = api
}

func GetProxyAPI() interface{} {
	return globalProxyAPI
}

func SetSettingsGetter(fn func(key string) string) {
	globalSettingsGetter = fn
}

func getSetting(key string) string {
	if globalSettingsGetter != nil {
		return globalSettingsGetter(key)
	}
	return ""
}

func isRTKEnabled() bool {
	return getSetting("rtkEnabled") != "false"
}

func isCavemanEnabled() bool {
	return getSetting("cavemanEnabled") == "true"
}

func getCavemanLevel() string {
	level := getSetting("cavemanLevel")
	if level == "" {
		return "full"
	}
	return level
}

type ChatRequest struct {
	Model    string          `json:"model"`
	Messages json.RawMessage `json:"messages"`
	Stream   bool            `json:"stream,omitempty"`
	MaxTokens int            `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
}

type refreshTokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

type providerError struct {
	statusCode int
	message    string
}

func (e *providerError) Error() string {
	return e.message
}

func HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	var req ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Model == "" {
		errs.WriteJSONError(w, "model is required", http.StatusBadRequest)
		return
	}

	// Route to streaming or non-streaming handler
	if req.Stream {
		handleStreamingChat(w, r, &req, body)
	} else {
		handleNonStreamingChat(w, r, &req, body)
	}
}

func handleNonStreamingChat(w http.ResponseWriter, r *http.Request, req *ChatRequest, body []byte) {
	w.Header().Set("Content-Type", "application/json")

	providerCfg, account, err := prepareChat(r.Context(), req.Model)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if globalPool == nil {
		errs.WriteJSONError(w, "account pool not initialized", http.StatusInternalServerError)
		return
	}

	_, modelName := ParseModel(req.Model)
	if err := refreshIfNeeded(&providerCfg, account); err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Retry loop: refresh on 401, backoff on 429
	var providerResp map[string]interface{}
	var lastErr error
	refreshed := false
	providerID, _ := ParseModel(req.Model)

	for attempt := 0; attempt < 3; attempt++ {
		providerResp, lastErr = callProviderAPI(r.Context(), &providerCfg, modelName, account, body)
		if lastErr == nil {
			break
		}

		pe, isProviderErr := lastErr.(*providerError)

		if isProviderErr && pe.statusCode == 401 && !refreshed && account.RefreshToken != "" {
			newToken, err := refreshProviderToken(&providerCfg, account.RefreshToken)
			if err == nil {
				account.AccessToken = newToken.AccessToken
				refreshed = true
				continue
			}
		}

		if isProviderErr && pe.statusCode == 429 {
			backoff := CalculateBackoff(attempt)
			globalPool.MarkRateLimited(account.ID, backoff)
			nextAcc, accErr := globalPool.GetAccount(providerID)
			if accErr != nil {
				errs.WriteJSONError(w, fmt.Sprintf("all accounts rate limited for %s", providerID), http.StatusTooManyRequests)
				return
			}
			account = nextAcc
			time.Sleep(backoff)
			continue
		}

		errs.WriteJSONError(w, lastErr.Error(), http.StatusBadGateway)
		return
	}

	if lastErr != nil {
		errs.WriteJSONError(w, "all retries exhausted", http.StatusTooManyRequests)
		return
	}

	// Step 6: Translate response based on provider format
	if providerCfg.Format == providers.FormatGeminiCLI {
		geminiResp := &resptrans.GeminiResponse{}
		data, _ := json.Marshal(providerResp)
		if json.Unmarshal(data, geminiResp) == nil {
			if openaiResp, transErr := resptrans.TranslateGeminiToOpenAIResponse(geminiResp, modelName); transErr == nil {
				json.NewEncoder(w).Encode(openaiResp)
				return
			}
		}
	}

	json.NewEncoder(w).Encode(providerResp)
}

func handleStreamingChat(w http.ResponseWriter, r *http.Request, req *ChatRequest, body []byte) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		errs.WriteJSONError(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	providerCfg, account, err := prepareChat(r.Context(), req.Model)
	if err != nil {
		writeSSE(w, flusher, map[string]interface{}{"error": err.Error()})
		writeSSE(w, flusher, "[DONE]")
		return
	}

	_, modelName := ParseModel(req.Model)
	if err := refreshIfNeeded(&providerCfg, account); err != nil {
		writeSSE(w, flusher, map[string]interface{}{"error": err.Error()})
		writeSSE(w, flusher, "[DONE]")
		return
	}

	// Build translated body
	reqMap := make(map[string]interface{})
	json.Unmarshal(body, &reqMap)
	reqMap["stream"] = true

	var translatedBody []byte
	switch providerCfg.Format {
	case "kiro":
		userContent := "Say hi"
		if msgs, ok := reqMap["messages"].([]interface{}); ok && len(msgs) > 0 {
			if last, ok := msgs[len(msgs)-1].(map[string]interface{}); ok {
				if c, ok := last["content"].(string); ok {
					userContent = c
				}
			}
		}
		kiroReq := map[string]interface{}{
			"conversationState": map[string]interface{}{
				"chatTriggerType": "MANUAL",
				"currentMessage": map[string]interface{}{
					"userInputMessage": map[string]interface{}{
						"content": userContent,
						"modelId": modelName,
					},
				},
			},
			"profileArn": "arn:aws:codewhisperer:us-east-1:699475941385:profile/EHGA3GRVQMUK",
			"inferenceConfig": map[string]interface{}{"maxTokens": 32000},
			"model": modelName,
		}
		translatedBody, _ = json.Marshal(kiroReq)
	default:
		translatedBody = body
	}

	// Build request
	url := buildProviderURL(providerCfg.BaseURL, modelName, providerCfg.Format)
	reqHTTP, _ := http.NewRequestWithContext(r.Context(), "POST", url, bytes.NewReader(translatedBody))
	reqHTTP.Header.Set("Content-Type", "application/json")
	if account.AccessToken != "" && account.AccessToken != "public" {
		reqHTTP.Header.Set("Authorization", "Bearer "+account.AccessToken)
	}
	for k, v := range providerCfg.Headers {
		reqHTTP.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		writeSSE(w, flusher, map[string]interface{}{"error": fmt.Sprintf("provider request failed: %v", err)})
		writeSSE(w, flusher, "[DONE]")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		writeSSE(w, flusher, map[string]interface{}{"error": fmt.Sprintf("provider returned %d: %s", resp.StatusCode, string(errBody))})
		writeSSE(w, flusher, "[DONE]")
		return
	}

	// Handle Kiro AWS EventStream streaming
	if providerCfg.Format == "kiro" {
		raw, _ := io.ReadAll(resp.Body)
		contentStr := string(raw)
		modelID := modelName
		chunkIndex := 0

		for {
			start := strings.Index(contentStr, `{"content":`)
			if start < 0 {
				break
			}
			contentStr = contentStr[start:]
			end := strings.Index(contentStr, `}`)
			if end < 0 {
				break
			}
			var evt map[string]interface{}
			if err := json.Unmarshal([]byte(contentStr[:end+1]), &evt); err == nil {
				if content, ok := evt["content"].(string); ok {
					chunk := map[string]interface{}{
						"id": fmt.Sprintf("chatcmpl-%d", time.Now().UnixMilli()),
						"object": "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model": modelID,
						"choices": []interface{}{
							map[string]interface{}{
								"index": chunkIndex,
								"delta": map[string]interface{}{
									"content": content,
								},
							},
						},
					}
					writeSSE(w, flusher, chunk)
					chunkIndex++
				}
			}
			nextPos := int(math.Min(float64(end+1), float64(len(contentStr)-1)))
			if nextPos >= len(contentStr) {
				nextPos = len(contentStr) - 1
			}
			contentStr = contentStr[nextPos:]
		}

		// Final chunk with finish_reason
		finalChunk := map[string]interface{}{
			"id": fmt.Sprintf("chatcmpl-%d", time.Now().UnixMilli()),
			"object": "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model": modelID,
			"choices": []interface{}{
				map[string]interface{}{
					"index": chunkIndex,
					"delta": map[string]interface{}{},
					"finish_reason": "stop",
				},
			},
		}
		writeSSE(w, flusher, finalChunk)
		writeSSE(w, flusher, "[DONE]")
		return
	}

	// For OpenAI-compatible providers, proxy the upstream SSE stream
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			w.Write([]byte(line + "\n\n"))
			flusher.Flush()
		}
	}
}

func prepareChat(ctx context.Context, modelStr string) (providers.ProviderConfig, *pool.Account, error) {
	providerID, _ := ParseModel(modelStr)
	if providerID == "" {
		return providers.ProviderConfig{}, nil, fmt.Errorf("model must be in provider/model format")
	}

	providerCfg, exists := providers.GetProviderConfig(providerID)
	if !exists {
		return providers.ProviderConfig{}, nil, fmt.Errorf("unknown provider: %s", providerID)
	}

	// For noAuth providers, create virtual account
	if providerCfg.AuthType == providers.AuthTypeNone {
		virtualAccount := &pool.Account{
			ID:          "noauth",
			Provider:    providerID,
			IsActive:    true,
			AccessToken: "public",
		}
		return providerCfg, virtualAccount, nil
	}

	if globalPool == nil {
		return providers.ProviderConfig{}, nil, fmt.Errorf("account pool not initialized")
	}

	account, err := globalPool.GetAccount(providerID)
	if err != nil {
		return providers.ProviderConfig{}, nil, fmt.Errorf("no available accounts for %s: %v", providerID, err)
	}

	return providerCfg, account, nil
}

func refreshIfNeeded(cfg *providers.ProviderConfig, account *pool.Account) error {
	if account.ExpiresAt.After(time.Time{}) {
		remaining := time.Until(account.ExpiresAt)
		if remaining < 5*time.Minute && account.RefreshToken != "" {
			newToken, err := refreshProviderToken(cfg, account.RefreshToken)
			if err == nil {
				account.AccessToken = newToken.AccessToken
				if newToken.RefreshToken != "" {
					account.RefreshToken = newToken.RefreshToken
				}
				account.ExpiresAt = time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)
			}
		}
	}
	return nil
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, data interface{}) {
	var line string
	switch v := data.(type) {
	case string:
		if v == "[DONE]" {
			line = "data: [DONE]\n\n"
		} else {
			line = fmt.Sprintf("data: %s\n\n", v)
		}
	default:
		b, _ := json.Marshal(v)
		line = fmt.Sprintf("data: %s\n\n", string(b))
	}
	w.Write([]byte(line))
	flusher.Flush()
}

func callProviderAPI(ctx context.Context, cfg *providers.ProviderConfig, model string, account *pool.Account, reqBody []byte) (map[string]interface{}, error) {
	var reqMap map[string]interface{}
	if err := json.Unmarshal(reqBody, &reqMap); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Translate request body to provider format
	var translatedBody []byte
	switch cfg.Format {
	case providers.FormatGeminiCLI:
		openaiReq := &reqtrans.OpenAIRequest{}
		if err := json.Unmarshal(reqBody, openaiReq); err == nil {
			isStream := false
			if s, ok := reqMap["stream"].(bool); ok {
				isStream = s
			}
			geminiReq := reqtrans.TranslateOpenAIToGeminiRequest(model, openaiReq, isStream)
			if geminiReq != nil {
				translatedBody, _ = json.Marshal(geminiReq)
			}
		}

	case "kiro":
		userContent := "Say hi"
		if msgs, ok := reqMap["messages"].([]interface{}); ok && len(msgs) > 0 {
			if last, ok := msgs[len(msgs)-1].(map[string]interface{}); ok {
				if c, ok := last["content"].(string); ok {
					userContent = c
				}
			}
		}
		kiroReq := map[string]interface{}{
			"conversationState": map[string]interface{}{
				"chatTriggerType": "MANUAL",
				"currentMessage": map[string]interface{}{
					"userInputMessage": map[string]interface{}{
						"content": userContent,
						"modelId": model,
					},
				},
			},
			"profileArn":     "arn:aws:codewhisperer:us-east-1:699475941385:profile/EHGA3GRVQMUK",
			"inferenceConfig": map[string]interface{}{"maxTokens": 32000},
			"model":           model,
		}
		translatedBody, _ = json.Marshal(kiroReq)

	case providers.FormatClaude:
		translatedBody = reqBody

	default:
		// Strip provider prefix from model field (e.g., "oc/model" → "model")
		reqMap["model"] = model

		// Apply RTK compression to tool messages
		if isRTKEnabled() {
			if msgs, ok := reqMap["messages"].([]interface{}); ok {
				for i, m := range msgs {
					if msg, ok := m.(map[string]interface{}); ok {
						content, hasContent := msg["content"].(string)
						if hasContent && len(content) > 500 {
							contentType := rtk.DetectContentType(content)
							filter := rtk.SelectFilter(contentType)
							if filter != nil {
								compressed := rtk.SafeApply(content, filter)
								if len(compressed) < len(content) {
									msg["content"] = compressed
									msgs[i] = msg
								}
							}
						}
					}
				}
				reqMap["messages"] = msgs
			}
		}

		// Inject Caveman terse prompt
		if isCavemanEnabled() {
			if msgs, ok := reqMap["messages"].([]interface{}); ok {
				parsed := make([]map[string]any, len(msgs))
				for i, m := range msgs {
					if msg, ok := m.(map[string]interface{}); ok {
						parsed[i] = msg
					}
				}
				parsed = rtk.InjectCavemanIntoMessages(parsed, getCavemanLevel())
				reparsed := make([]interface{}, len(parsed))
				for i, p := range parsed {
					reparsed[i] = p
				}
				reqMap["messages"] = reparsed
			}
		}

		translatedBody, _ = json.Marshal(reqMap)
	}

	if translatedBody == nil {
		translatedBody = reqBody
	}

	// Build provider-specific URL
	url := buildProviderURL(cfg.BaseURL, model, cfg.Format)

	// Build request
	req, httpErr := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(translatedBody))
	if httpErr != nil {
		return nil, httpErr
	}

	req.Header.Set("Content-Type", "application/json")
	if account.AccessToken != "" && account.AccessToken != "public" {
		req.Header.Set("Authorization", "Bearer "+account.AccessToken)
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	// Send
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("provider request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &providerError{
			statusCode: resp.StatusCode,
			message:    fmt.Sprintf("provider returned %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	// AWS EventStream binary response - extract text content
	if len(respBody) > 0 && respBody[0] == 0x00 {
		var fullContent string
		contentStr := string(respBody)
		// Find all JSON objects in binary stream by scanning for {"content":
		for {
			start := strings.Index(contentStr, `{"content":`)
			if start < 0 {
				break
			}
			contentStr = contentStr[start:]
			end := strings.Index(contentStr, `}`)
			if end < 0 {
				break
			}
			var eventData map[string]interface{}
			if err := json.Unmarshal([]byte(contentStr[:end+1]), &eventData); err == nil {
				if c, ok := eventData["content"].(string); ok {
					fullContent += c
				}
			}
			nextPos := int(math.Min(float64(end+1), float64(len(contentStr)-1)))
			if nextPos >= len(contentStr) {
				nextPos = len(contentStr) - 1
			}
			contentStr = contentStr[nextPos:]
		}

		result := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": fullContent,
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":    0,
				"completion_tokens": 0,
				"total_tokens":      0,
			},
		}
		return result, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

func buildProviderURL(baseURL, model, format string) string {
	switch format {
	case providers.FormatGeminiCLI:
		return fmt.Sprintf("%s:generateContent", baseURL)
	case providers.FormatOpenAI:
		return fmt.Sprintf("%s/chat/completions", baseURL)
	case providers.FormatClaude:
		return fmt.Sprintf("%s/messages", baseURL)
	case "kiro":
		return baseURL
	default:
		return fmt.Sprintf("%s/chat/completions", baseURL)
	}
}

func refreshProviderToken(cfg *providers.ProviderConfig, refreshToken string) (*refreshTokenResult, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	var resp *http.Response
	var err error

	// Kiro uses its own refresh endpoint
	if cfg.Type == providers.TypeKiro {
		payload := map[string]string{
			"refreshToken": refreshToken,
		}
		payloadBytes, _ := json.Marshal(payload)

		resp, err = http.Post(
			providers.KIRO_REFRESH_URL,
			"application/json",
			bytes.NewReader(payloadBytes),
		)
		if err != nil {
			return nil, fmt.Errorf("kiro refresh failed: %w", err)
		}
		defer resp.Body.Close()

		var result struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
			ExpiresIn    int    `json:"expiresIn"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode kiro refresh: %w", err)
		}
		return &refreshTokenResult{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresIn:    result.ExpiresIn,
		}, nil
	}

	// Generic OAuth refresh for other providers
	if cfg.AuthType == providers.AuthTypeOAuth && cfg.Headers != nil {
		payload := map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
		}
		payloadBytes, _ := json.Marshal(payload)

		resp, err = http.Post(
			cfg.BaseURL+"/oauth/token",
			"application/json",
			bytes.NewReader(payloadBytes),
		)
		if err != nil {
			return nil, fmt.Errorf("oauth refresh failed: %w", err)
		}
		defer resp.Body.Close()

		var result struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode oauth refresh: %w", err)
		}
		return &refreshTokenResult{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresIn:    result.ExpiresIn,
		}, nil
	}

	return nil, fmt.Errorf("refresh not implemented for %s", cfg.Type)
}
