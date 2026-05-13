package router

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/handlers/responses"
	"github.com/DevKuroX/AIPROXY/internal/stream"
)

// ResponsesRequest represents a Responses API request
// ref: open-sse/handlers/responsesHandler.js - request body structure
type ResponsesRequest struct {
	Model  string          `json:"model"`
	Input  json.RawMessage `json:"input,omitempty"`
	Stream bool            `json:"stream,omitempty"`
}

// HandleResponses handles /v1/responses endpoint
// ref: open-sse/handlers/responsesHandler.js - handleResponsesCore
func HandleResponses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req ResponsesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		errs.WriteJSONError(w, "model is required", http.StatusBadRequest)
		return
	}

	resolvedModel, err := ResolveAlias(ctx, req.Model)
	if err != nil {
		errs.WriteJSONError(w, "failed to resolve model alias: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Model = resolvedModel

	if providerStore == nil {
		errs.WriteJSONError(w, "provider store not configured", http.StatusInternalServerError)
		return
	}

	// Convert Responses API to Chat Completions format
	// ref: open-sse/handlers/responsesHandler.js:26
	convertedBody, err := responses.ConvertResponsesToChat(body)
	if err != nil {
		errs.WriteJSONError(w, "failed to convert Responses API format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if client requested streaming (default false)
	// ref: open-sse/handlers/responsesHandler.js:30-33
	clientRequestedStreaming := req.Stream

	providers, err := providerStore.GetProviders(req.Model)
	if err != nil || len(providers) == 0 {
		provider, err := providerStore.GetProvider(req.Model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get provider for model: "+err.Error(), http.StatusInternalServerError)
			return
		}
		providers = []*ProviderConfig{provider}
	}

	accounts := make([]*ProviderAccount, len(providers))
	for i, p := range providers {
		accounts[i] = &ProviderAccount{
			Name:    p.Name,
			BaseURL: p.BaseURL,
			APIKey:  p.APIKey,
		}
	}

	fc := NewFallbackChain(accounts)

	var lastErr error
	var lastStatus int
	var lastErrorText string

	for {
		account := fc.GetCurrent()
		if account == nil {
			break
		}

		upstreamURL := account.BaseURL
		if !strings.HasSuffix(upstreamURL, "/") {
			upstreamURL += "/"
		}
		upstreamURL += "v1/chat/completions"

		upstreamResp, err := ForwardRequest(ctx, upstreamURL, account.APIKey, convertedBody, clientRequestedStreaming)
		if err != nil {
			lastErr = err
			lastStatus = http.StatusBadGateway
			lastErrorText = err.Error()

			_, tryErr := fc.TryNext(ctx, account, err, lastStatus, lastErrorText)
			if tryErr != nil {
				break
			}
			continue
		}

		if upstreamResp.StatusCode >= 400 {
			errBody, _ := io.ReadAll(upstreamResp.Body)
			upstreamResp.Body.Close()

			lastStatus = upstreamResp.StatusCode
			lastErrorText = extractErrorText(errBody)
			lastErr = NewUpstreamError(lastStatus, lastErrorText)

			if ShouldRetry(lastStatus, lastErrorText) {
				_, tryErr := fc.TryNext(ctx, account, lastErr, lastStatus, lastErrorText)
				if tryErr != nil {
					break
				}
				continue
			}

			// Return error response
			CopyHeaders(w.Header(), upstreamResp.Header)
			w.WriteHeader(upstreamResp.StatusCode)
			w.Write(errBody)
			return
		}

		// Handle response based on content type and stream preference
		// ref: open-sse/handlers/responsesHandler.js:55-101
		handler := responses.NewHandler()
		config := &stream.StreamConfig{
			Provider:     account.Name,
			Model:        req.Model,
			SourceFormat: stream.FormatOpenAI,
			TargetFormat: stream.FormatOpenAIResponses,
			Stream:       clientRequestedStreaming,
		}

		if err := handler.HandleResponses(ctx, w, upstreamResp, config, clientRequestedStreaming); err != nil {
			lastErr = err
			lastStatus = http.StatusInternalServerError
			lastErrorText = err.Error()

			_, tryErr := fc.TryNext(ctx, account, err, lastStatus, lastErrorText)
			if tryErr != nil {
				break
			}
			continue
		}

		return
	}

	// All accounts exhausted
	if lastErr != nil {
		errs.WriteJSONError(w, lastErrorText, lastStatus)
		return
	}
	errs.WriteJSONError(w, "no available providers", http.StatusServiceUnavailable)
}

// HandleResponsesDirect handles /v1/responses with pre-converted body
// Used when body is already in chat completions format
func HandleResponsesDirect(w http.ResponseWriter, r *http.Request, convertedBody []byte, model string, streamRequested bool) {
	ctx := r.Context()

	if providerStore == nil {
		errs.WriteJSONError(w, "provider store not configured", http.StatusInternalServerError)
		return
	}

	providers, err := providerStore.GetProviders(model)
	if err != nil || len(providers) == 0 {
		provider, err := providerStore.GetProvider(model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get provider for model: "+err.Error(), http.StatusInternalServerError)
			return
		}
		providers = []*ProviderConfig{provider}
	}

	accounts := make([]*ProviderAccount, len(providers))
	for i, p := range providers {
		accounts[i] = &ProviderAccount{
			Name:    p.Name,
			BaseURL: p.BaseURL,
			APIKey:  p.APIKey,
		}
	}

	fc := NewFallbackChain(accounts)

	for {
		account := fc.GetCurrent()
		if account == nil {
			break
		}

		upstreamURL := account.BaseURL
		if !strings.HasSuffix(upstreamURL, "/") {
			upstreamURL += "/"
		}
		upstreamURL += "v1/chat/completions"

		upstreamResp, err := ForwardRequest(ctx, upstreamURL, account.APIKey, convertedBody, streamRequested)
		if err != nil {
			_, tryErr := fc.TryNext(ctx, account, err, http.StatusBadGateway, err.Error())
			if tryErr != nil {
				break
			}
			continue
		}

		if upstreamResp.StatusCode >= 400 {
			errBody, _ := io.ReadAll(upstreamResp.Body)
			upstreamResp.Body.Close()

			if ShouldRetry(upstreamResp.StatusCode, extractErrorText(errBody)) {
				_, tryErr := fc.TryNext(ctx, account, NewUpstreamError(upstreamResp.StatusCode, extractErrorText(errBody)), upstreamResp.StatusCode, extractErrorText(errBody))
				if tryErr != nil {
					break
				}
				continue
			}

			CopyHeaders(w.Header(), upstreamResp.Header)
			w.WriteHeader(upstreamResp.StatusCode)
			w.Write(errBody)
			return
		}

		handler := responses.NewHandler()
		config := &stream.StreamConfig{
			Provider:     account.Name,
			Model:        model,
			SourceFormat: stream.FormatOpenAI,
			TargetFormat: stream.FormatOpenAIResponses,
			Stream:       streamRequested,
		}

		if err := handler.HandleResponses(ctx, w, upstreamResp, config, streamRequested); err != nil {
			_, tryErr := fc.TryNext(ctx, account, err, http.StatusInternalServerError, err.Error())
			if tryErr != nil {
				break
			}
			continue
		}

		return
	}

	errs.WriteJSONError(w, "no available providers", http.StatusServiceUnavailable)
}

// IsResponsesAPI checks if request is for Responses API
// ref: open-sse/handlers/responsesHandler.js - detection logic
func IsResponsesAPI(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/v1/responses")
}
