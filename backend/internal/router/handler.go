package router

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/stream"
)

type ProviderConfig struct {
	Name    string
	BaseURL string
	APIKey  string
}

type ProviderStore interface {
	GetProvider(model string) (*ProviderConfig, error)
	GetProviders(model string) ([]*ProviderConfig, error)
}

var providerStore ProviderStore

func SetProviderStore(store ProviderStore) {
	providerStore = store
}

func GetProviderStore() ProviderStore {
	return providerStore
}

type ChatCompletionRequest struct {
	Model    string          `json:"model"`
	Messages json.RawMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

func HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req ChatCompletionRequest
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

		upstreamResp, err := ForwardRequest(ctx, upstreamURL, account.APIKey, body, req.Stream)
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

			_, tryErr := fc.TryNext(ctx, account, lastErr, lastStatus, lastErrorText)
			if tryErr != nil {
				upstreamResp.Body.Close()
				break
			}
			continue
		}

		if req.Stream {
			handleStreamingResponse(w, upstreamResp)
		} else {
			handleNonStreamingResponse(w, upstreamResp)
		}
		return
	}

	if lastErr != nil {
		if exhausted, ok := lastErr.(*AllAccountsExhaustedError); ok {
			errs.WriteJSONError(w, exhausted.Error(), http.StatusServiceUnavailable)
			return
		}
		errs.WriteJSONError(w, "all providers failed: "+lastErrorText, lastStatus)
		return
	}

	errs.WriteJSONError(w, "no providers available", http.StatusServiceUnavailable)
}

func extractErrorText(body []byte) string {
	var errResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Error.Message != "" {
			return errResp.Error.Message
		}
		if errResp.Message != "" {
			return errResp.Message
		}
	}
	return string(body)
}

type UpstreamError struct {
	Status    int
	ErrorText string
}

func NewUpstreamError(status int, errorText string) *UpstreamError {
	return &UpstreamError{Status: status, ErrorText: errorText}
}

func (e *UpstreamError) Error() string {
	return e.ErrorText
}

func handleUpstreamError(w http.ResponseWriter, resp *http.Response) {
	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func handleStreamingResponse(w http.ResponseWriter, resp *http.Response) {
	sseWriter := stream.NewSSEWriter(w)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				sseWriter.WriteDone()
				return
			}
			sseWriter.WriteRaw(context.Background(), []byte(data))
		} else {
			sseWriter.WriteRaw(context.Background(), []byte(line))
		}
	}

	if err := scanner.Err(); err != nil {
		sseWriter.WriteError(context.Background(), fmt.Errorf("stream error: %w", err))
		return
	}

	sseWriter.WriteDone()
}

func handleNonStreamingResponse(w http.ResponseWriter, resp *http.Response) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		errs.WriteJSONError(w, "failed to read upstream response", http.StatusInternalServerError)
		return
	}

	CopyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	w.Write(buf.Bytes())
}

type ComboHandler struct {
	providerStore ProviderStore
	comboResolver *ComboResolver
	healthTracker *AccountHealthState
}

func NewComboHandler(store ProviderStore, resolver *ComboResolver, tracker *AccountHealthState) *ComboHandler {
	return &ComboHandler{
		providerStore: store,
		comboResolver: resolver,
		healthTracker: tracker,
	}
}

func (h *ComboHandler) HandleComboChat(ctx context.Context, body []byte, models []string, stream bool) (*http.Response, error) {
	var lastErr error
	var lastStatus int
	var lastErrorText string

	for _, modelStr := range models {
		providers, err := h.providerStore.GetProviders(modelStr)
		if err != nil || len(providers) == 0 {
			continue
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

			upstreamResp, err := ForwardRequest(ctx, upstreamURL, account.APIKey, body, stream)
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

				return upstreamResp, nil
			}

			return upstreamResp, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ErrAllAccountsExhausted
}
