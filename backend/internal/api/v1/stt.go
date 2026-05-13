package v1

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/router"
)

// ref: open-sse/handlers/sttCore.js:170-194
// STT (Speech-to-Text) endpoint handler

// STTConfig holds configuration for STT providers
type STTConfig struct {
	Name       string
	BaseURL    string
	APIKey     string
	AuthHeader string // "bearer", "token", "x-api-key", "key", or "none"
}

// STTProviderStore retrieves STT provider configuration
type STTProviderStore interface {
	GetSTTProvider(model string) (*STTConfig, error)
}

var sttProviderStore STTProviderStore

func SetSTTProviderStore(store STTProviderStore) {
	sttProviderStore = store
}

// TranscriptionResponse is the standard STT response format
// ref: open-sse/handlers/sttCore.js:156-164
type TranscriptionResponse struct {
	Text string `json:"text"`
}

// HandleAudioTranscriptions handles POST /v1/audio/transcriptions
// ref: open-sse/handlers/sttCore.js:170-194
func HandleAudioTranscriptions(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 32MB)
	// ref: open-sse/handlers/sttCore.js:171-172
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		errs.WriteJSONError(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		errs.WriteJSONError(w, "missing required field: file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get model from form, default to whisper-1
	// ref: open-sse/handlers/sttCore.js:144
	model := r.FormValue("model")
	if model == "" {
		model = "whisper-1"
	}

	// Get provider configuration
	var provider *STTConfig
	if sttProviderStore != nil {
		provider, err = sttProviderStore.GetSTTProvider(model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get STT provider: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Fallback to router's provider store
	if provider == nil {
		fallbackProvider, err := router.GetProviderStore().GetProvider(model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get provider for model: "+err.Error(), http.StatusInternalServerError)
			return
		}
		provider = &STTConfig{
			Name:       fallbackProvider.Name,
			BaseURL:    fallbackProvider.BaseURL,
			APIKey:     fallbackProvider.APIKey,
			AuthHeader: "bearer",
		}
	}

	if provider == nil {
		errs.WriteJSONError(w, "no provider configured for STT", http.StatusInternalServerError)
		return
	}

	// Forward to 0penAI-compatible STT endpoint
	// ref: open-sse/handlers/sttCore.js:141-154
	resp, err := forwardSTTRequest(r, provider, file, header.Filename, model)
	if err != nil {
		errs.WriteJSONError(w, "STT request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// forwardSTTRequest forwards the STT request to an 0penAI-compatible provider
// ref: open-sse/handlers/sttCore.js:141-154
func forwardSTTRequest(r *http.Request, provider *STTConfig, file io.Reader, filename, model string) (*http.Response, error) {
	// Build multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file
	// ref: open-sse/handlers/sttCore.js:143
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	// Add model
	// ref: open-sse/handlers/sttCore.js:144
	err = writer.WriteField("model", model)
	if err != nil {
		return nil, err
	}

	// Forward optional fields
	// ref: open-sse/handlers/sttCore.js:145-148
	for _, k := range []string{"language", "prompt", "response_format", "temperature"} {
		v := r.FormValue(k)
		if v != "" {
			err = writer.WriteField(k, v)
			if err != nil {
				return nil, err
			}
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// Build URL
	upstreamURL := provider.BaseURL
	if !strings.HasSuffix(upstreamURL, "/") {
		upstreamURL += "/"
	}
	upstreamURL += "v1/audio/transcriptions"

	// Create request
	req, err := http.NewRequestWithContext(r.Context(), "POST", upstreamURL, &body)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set auth header
	// ref: open-sse/handlers/sttCore.js:7-16
	switch provider.AuthHeader {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	case "token":
		req.Header.Set("Authorization", "Token "+provider.APIKey)
	case "x-api-key":
		req.Header.Set("x-api-key", provider.APIKey)
	case "key":
		req.Header.Set("Authorization", "Key "+provider.APIKey)
	default:
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	client := router.NewProxyClient()
	return client.Do(req)
}

// HandleAudioTranscriptionsJSON handles JSON responses for STT
// ref: open-sse/handlers/sttCore.js:156-164
func HandleAudioTranscriptionsJSON(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TranscriptionResponse{Text: text})
}
