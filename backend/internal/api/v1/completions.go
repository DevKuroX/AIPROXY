package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/router"
)

// ref: 9router/src/app/api/v1/chat/completions/route.js
// POST /v1/completions - Legacy completions endpoint (non-chat)

// CompletionRequest represents a legacy completion request
type CompletionRequest struct {
	Model       string         `json:"model"`
	Prompt      interface{}    `json:"prompt"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	N           int            `json:"n,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
	Logprobs    int            `json:"logprobs,omitempty"`
	Echo        bool           `json:"echo,omitempty"`
	Stop        interface{}    `json:"stop,omitempty"`
	Suffix      string         `json:"suffix,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CompletionChoice represents a completion choice
type CompletionChoice struct {
	Text         string    `json:"text"`
	Index        int       `json:"index"`
	Logprobs     *Logprobs `json:"logprobs,omitempty"`
	FinishReason string    `json:"finish_reason"`
}

// Logprobs represents log probabilities
type Logprobs struct {
	Tokens        []string             `json:"tokens,omitempty"`
	TokenLogprobs []float64            `json:"token_logprobs,omitempty"`
	TopLogprobs   []map[string]float64 `json:"top_logprobs,omitempty"`
	TextOffset    []int                `json:"text_offset,omitempty"`
}

// CompletionResponse represents the completion response
type CompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   *UsageInfo         `json:"usage,omitempty"`
}

// UsageInfo contains token usage information
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// HandleCompletions handles POST /v1/completions
// ref: 9router/src/app/api/v1/chat/completions/route.js
func HandleCompletions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req CompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		errs.WriteJSONError(w, "model is required", http.StatusBadRequest)
		return
	}

	if req.Prompt == nil {
		errs.WriteJSONError(w, "prompt is required", http.StatusBadRequest)
		return
	}

	resolvedModel, err := router.ResolveAlias(ctx, req.Model)
	if err != nil {
		errs.WriteJSONError(w, "failed to resolve model alias: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Model = resolvedModel

	if router.GetProviderStore() == nil {
		errs.WriteJSONError(w, "provider store not configured", http.StatusInternalServerError)
		return
	}

	provider, err := router.GetProviderStore().GetProvider(req.Model)
	if err != nil {
		errs.WriteJSONError(w, "failed to get provider for model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	upstreamURL := provider.BaseURL
	if !strings.HasSuffix(upstreamURL, "/") {
		upstreamURL += "/"
	}
	upstreamURL += "v1/completions"

	upstreamReq, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, bytes.NewReader(body))
	if err != nil {
		errs.WriteJSONError(w, "failed to create upstream request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Authorization", "Bearer "+provider.APIKey)

	for k, v := range r.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-") ||
			k == "Accept" ||
			k == "Accept-Encoding" {
			upstreamReq.Header[k] = v
		}
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(upstreamReq)
	if err != nil {
		errs.WriteJSONError(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

// HandleCompletionsStream handles streaming completions
func HandleCompletionsStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req CompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Stream = true
	streamBody, _ := json.Marshal(req)

	resolvedModel, err := router.ResolveAlias(ctx, req.Model)
	if err != nil {
		errs.WriteJSONError(w, "failed to resolve model alias: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Model = resolvedModel

	provider, err := router.GetProviderStore().GetProvider(req.Model)
	if err != nil {
		errs.WriteJSONError(w, "failed to get provider for model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	upstreamURL := provider.BaseURL
	if !strings.HasSuffix(upstreamURL, "/") {
		upstreamURL += "/"
	}
	upstreamURL += "v1/completions"

	upstreamReq, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, bytes.NewReader(streamBody))
	if err != nil {
		errs.WriteJSONError(w, "failed to create upstream request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Authorization", "Bearer "+provider.APIKey)
	upstreamReq.Header.Set("Accept", "text/event-stream")

	for k, v := range r.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-") {
			upstreamReq.Header[k] = v
		}
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(upstreamReq)
	if err != nil {
		errs.WriteJSONError(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(resp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		io.Copy(w, resp.Body)
		return
	}

	reader := io.Reader(resp.Body)
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
}
