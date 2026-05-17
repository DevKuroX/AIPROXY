package v1

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/handlers"
)

// ref: open-sse/handlers/ttsCore.js:51
type TTSRequest struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Voice  string `json:"voice"`
	Format string `json:"response_format,omitempty"`
}

type TTSProviderStore interface {
	GetTTSProvider(model string) (*handlers.TTSProviderConfig, error)
}

var ttsProviderStore TTSProviderStore

func SetTTSProviderStore(store TTSProviderStore) {
	ttsProviderStore = store
}

var ttsEnabled bool

func SetTTSEnabled(enabled bool) {
	ttsEnabled = enabled
}

func IsTTSEnabled() bool {
	return ttsEnabled
}

// ref: open-sse/handlers/ttsCore.js:51
func HandleTTSSpeech(w http.ResponseWriter, r *http.Request) {
	if !ttsEnabled {
		errs.WriteJSONError(w, "TTS is not enabled", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req TTSRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// ref: open-sse/handlers/ttsCore.js:52-54
	if strings.TrimSpace(req.Input) == "" {
		errs.WriteJSONError(w, "missing required field: input", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = "tts-1"
	}
	if req.Voice == "" {
		req.Voice = "alloy"
	}
	if req.Format == "" {
		req.Format = "mp3"
	}

	var provider *handlers.TTSProviderConfig
	if ttsProviderStore != nil {
		provider, err = ttsProviderStore.GetTTSProvider(req.Model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get TTS provider: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if provider == nil {
		errs.WriteJSONError(w, "no TTS provider configured", http.StatusInternalServerError)
		return
	}

	// ref: open-sse/handlers/ttsCore.js:58-64
	result, err := handlers.SynthesizeTTS(r.Context(), provider, req.Input, req.Model, req.Voice, req.Format)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusBadGateway)
		return
	}

	// ref: open-sse/handlers/ttsCore.js:15-42
	w.Header().Set("Content-Type", "audio/"+result.Format)
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Audio)))
	w.WriteHeader(http.StatusOK)
	w.Write(result.Audio)
}
