package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/handlers/embeddings"
	"github.com/DevKuroX/AIPROXY/internal/router"
)

type EmbeddingProviderConfig struct {
	Name    string
	BaseURL string
	APIKey  string
}

type EmbeddingProviderStore interface {
	GetEmbeddingProvider(model string) (*EmbeddingProviderConfig, error)
}

var embeddingProviderStore EmbeddingProviderStore

func SetEmbeddingProviderStore(store EmbeddingProviderStore) {
	embeddingProviderStore = store
}

func HandleEmbeddings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req embeddings.EmbeddingRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Input == nil {
		errs.WriteJSONError(w, "missing required field: input", http.StatusBadRequest)
		return
	}

	switch v := req.Input.(type) {
	case string:
		if v == "" {
			errs.WriteJSONError(w, "input must be a non-empty string or array of strings", http.StatusBadRequest)
			return
		}
	case []interface{}:
		if len(v) == 0 {
			errs.WriteJSONError(w, "input must be a non-empty string or array of strings", http.StatusBadRequest)
			return
		}
	case []string:
		if len(v) == 0 {
			errs.WriteJSONError(w, "input must be a non-empty string or array of strings", http.StatusBadRequest)
			return
		}
	default:
		errs.WriteJSONError(w, "input must be a string or array of strings", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		errs.WriteJSONError(w, "model is required", http.StatusBadRequest)
		return
	}

	resolvedModel, err := router.ResolveAlias(ctx, req.Model)
	if err != nil {
		errs.WriteJSONError(w, "failed to resolve model alias: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Model = resolvedModel

	var provider *EmbeddingProviderConfig
	if embeddingProviderStore != nil {
		provider, err = embeddingProviderStore.GetEmbeddingProvider(req.Model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get embedding provider: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if provider == nil {
		fallbackProvider, err := router.GetProviderStore().GetProvider(req.Model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get provider for model: "+err.Error(), http.StatusInternalServerError)
			return
		}
		provider = &EmbeddingProviderConfig{
			Name:    fallbackProvider.Name,
			BaseURL: fallbackProvider.BaseURL,
			APIKey:  fallbackProvider.APIKey,
		}
	}

	if provider == nil {
		errs.WriteJSONError(w, "no provider configured for embeddings", http.StatusInternalServerError)
		return
	}

	creds := &embeddings.Credentials{
		APIKey:  provider.APIKey,
		BaseURL: provider.BaseURL,
	}

	providerAdapter := embeddings.GetProvider(provider.Name)
	if providerAdapter == nil {
		errs.WriteJSONError(w, "provider '"+provider.Name+"' does not support embeddings", http.StatusBadRequest)
		return
	}

	upstreamURL := providerAdapter.BuildURL(req.Model, creds, req.Input)
	headers := providerAdapter.BuildHeaders(creds)
	requestBody := providerAdapter.BuildBody(req.Model, &req)

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		errs.WriteJSONError(w, "failed to marshal request body", http.StatusInternalServerError)
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		errs.WriteJSONError(w, "failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		errs.WriteJSONError(w, "failed to connect to provider: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read provider response", http.StatusBadGateway)
		return
	}

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		if json.Unmarshal(respBody, &errResp) == nil {
			if errMsg, ok := errResp["error"].(map[string]interface{}); ok {
				if msg, ok := errMsg["message"].(string); ok {
					errs.WriteJSONError(w, "provider error: "+msg, resp.StatusCode)
					return
				}
			}
		}
		errs.WriteJSONError(w, "provider returned error status: "+resp.Status, resp.StatusCode)
		return
	}

	var respMap map[string]interface{}
	if err := json.Unmarshal(respBody, &respMap); err != nil {
		errs.WriteJSONError(w, "invalid JSON response from provider", http.StatusBadGateway)
		return
	}

	normalized := providerAdapter.Normalize(respMap, req.Model)

	normalizedBytes, err := embeddings.MarshalEmbeddingResponse(normalized)
	if err != nil {
		errs.WriteJSONError(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(normalizedBytes)
}

func isValidEmbeddingInput(input interface{}) bool {
	switch v := input.(type) {
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	case []string:
		return len(v) > 0
	default:
		return false
	}
}

func isBlankString(s string) bool {
	return strings.TrimSpace(s) == ""
}
