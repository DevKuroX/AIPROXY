package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

type AnthropicRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []AnthropicMsg  `json:"messages"`
	Stream    bool            `json:"stream,omitempty"`
}

type AnthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Role       string            `json:"role"`
	Model      string            `json:"model"`
	Content    []AnthropicContent `json:"content"`
	StopReason string            `json:"stop_reason,omitempty"`
	Usage      AnthropicUsage    `json:"usage"`
}

type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func HandleAnthropicMessages(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, 400)
		return
	}
	r.Body.Close()

	var anReq AnthropicRequest
	if err := json.Unmarshal(body, &anReq); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, 400)
		return
	}

	// Convert Anthropic → OpenAI format
	openAIReq := map[string]interface{}{
		"model":      anReq.Model,
		"max_tokens": anReq.MaxTokens,
		"stream":     anReq.Stream,
	}

	msgs := make([]interface{}, len(anReq.Messages))
	for i, m := range anReq.Messages {
		msgs[i] = map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
	}
	openAIReq["messages"] = msgs

	convertedBody, _ := json.Marshal(openAIReq)

	// Reuse existing chat logic by calling internal handler
	internalReq := &ChatRequest{
		Model:    anReq.Model,
		Stream:   anReq.Stream,
		MaxTokens: anReq.MaxTokens,
	}

	if anReq.Stream {
		handleStreamingAnthropic(w, r, internalReq, convertedBody)
	} else {
		handleNonStreamingAnthropic(w, r, internalReq, convertedBody)
	}
}

func handleNonStreamingAnthropic(w http.ResponseWriter, r *http.Request, req *ChatRequest, body []byte) {
	providerCfg, account, err := prepareChat(r.Context(), req.Model)
	if err != nil {
		writeAnthropicError(w, err.Error())
		return
	}
	_, modelName := ParseModel(req.Model)
	refreshIfNeeded(&providerCfg, account)

	providerResp, err := callProviderAPI(r.Context(), &providerCfg, modelName, account, body)
	if err != nil {
		writeAnthropicError(w, err.Error())
		return
	}

	// Extract content from OpenAI response
	content := ""
	if choices, ok := providerResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if first, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := first["message"].(map[string]interface{}); ok {
				if c, ok := msg["content"].(string); ok {
					content = c
				}
			}
		}
	}

	resp := AnthropicResponse{
		ID:         fmt.Sprintf("msg_%d", time.Now().UnixMilli()),
		Type:       "message",
		Role:       "assistant",
		Model:      req.Model,
		Content:    []AnthropicContent{{Type: "text", Text: content}},
		StopReason: "end_turn",
		Usage:      AnthropicUsage{InputTokens: 0, OutputTokens: 0},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleStreamingAnthropic(w http.ResponseWriter, r *http.Request, req *ChatRequest, body []byte) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAnthropicError(w, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	providerCfg, account, err := prepareChat(r.Context(), req.Model)
	if err != nil {
		writeAnthropicSSEWarn(w, flusher, err.Error())
		return
	}
	_, modelName := ParseModel(req.Model)
	refreshIfNeeded(&providerCfg, account)

	// Build translated body for Kiro
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
				"currentMessage":  map[string]interface{}{
					"userInputMessage": map[string]interface{}{
						"content": userContent,
						"modelId": modelName,
					},
				},
			},
			"profileArn":     "arn:aws:codewhisperer:us-east-1:699475941385:profile/EHGA3GRVQMUK",
			"inferenceConfig": map[string]interface{}{"maxTokens": 32000},
			"model":           modelName,
		}
		translatedBody, _ = json.Marshal(kiroReq)
	default:
		translatedBody = body
	}

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
		writeAnthropicSSEWarn(w, flusher, err.Error())
		return
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	contentStr := string(raw)

	// Event: message_start
	startEvent := fmt.Sprintf(`{"type":"message_start","message":{"id":"msg_%d","type":"message","role":"assistant","model":"%s","content":[]}}`, time.Now().UnixMilli(), req.Model)
	fmt.Fprintf(w, "event: message_start\ndata: %s\n\n", startEvent)
	flusher.Flush()

	// Parse chunks from EventStream
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
				deltaEvent := fmt.Sprintf(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"%s"}}`, escapeJSON(content))
				fmt.Fprintf(w, "event: content_block_delta\ndata: %s\n\n", deltaEvent)
				flusher.Flush()
			}
		}
		nextPos := int(math.Min(float64(end+1), float64(len(contentStr)-1)))
		if nextPos >= len(contentStr) {
			nextPos = len(contentStr) - 1
		}
		contentStr = contentStr[nextPos:]
	}

	// Event: message_delta (done)
	deltaDone := `{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":0}}`
	fmt.Fprintf(w, "event: message_delta\ndata: %s\n\n", deltaDone)
	fmt.Fprintf(w, "event: message_stop\ndata: {}\n\n")
	flusher.Flush()
}

func writeAnthropicError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{"message": msg, "type": "invalid_request_error"},
	})
}

func writeAnthropicSSEWarn(w http.ResponseWriter, flusher http.Flusher, msg string) {
	fmt.Fprintf(w, "event: error\ndata: {\"error\":{\"message\":%q}}\n\n", msg)
	flusher.Flush()
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}
