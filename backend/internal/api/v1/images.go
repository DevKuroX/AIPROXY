package v1

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/config"
	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/router"
	"github.com/DevKuroX/AIPROXY/internal/translator/request"
)

type ImageProviderConfig struct {
	Name    string
	BaseURL string
	APIKey  string
}

type ImageProviderStore interface {
	GetImageProvider(model string) (*ImageProviderConfig, error)
}

var imageProviderStore ImageProviderStore

func SetImageProviderStore(store ImageProviderStore) {
	imageProviderStore = store
}

var imageEnabled bool

func SetImageEnabled(enabled bool) {
	imageEnabled = enabled
}

func IsImageEnabled() bool {
	return imageEnabled || config.IsImageGenerationEnabled()
}

type ImageGenerationHandler struct {
	providerStore ImageProviderStore
}

func NewImageGenerationHandler(store ImageProviderStore) *ImageGenerationHandler {
	return &ImageGenerationHandler{
		providerStore: store,
	}
}

func HandleImageGenerations(w http.ResponseWriter, r *http.Request) {
	if !IsImageEnabled() {
		errs.WriteJSONError(w, "image generation is not enabled", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req request.ImageGenerationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		errs.WriteJSONError(w, "prompt is required", http.StatusBadRequest)
		return
	}

	model := req.Model
	if model == "" {
		model = "dall-e-3"
	}

	var provider *ImageProviderConfig
	if imageProviderStore != nil {
		provider, err = imageProviderStore.GetImageProvider(model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get image provider: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if provider == nil {
		fallbackProvider, err := router.GetProviderStore().GetProvider(model)
		if err != nil {
			errs.WriteJSONError(w, "failed to get provider for model: "+err.Error(), http.StatusInternalServerError)
			return
		}
		provider = &ImageProviderConfig{
			Name:    fallbackProvider.Name,
			BaseURL: fallbackProvider.BaseURL,
			APIKey:  fallbackProvider.APIKey,
		}
	}

	if provider == nil {
		errs.WriteJSONError(w, "no provider configured for image generation", http.StatusInternalServerError)
		return
	}

	dalleReq := request.TranslateImageRequest(&req)

	reqBody, err := json.Marshal(dalleReq)
	if err != nil {
		errs.WriteJSONError(w, "failed to marshal request", http.StatusInternalServerError)
		return
	}

	upstreamURL := provider.BaseURL
	if !strings.HasSuffix(upstreamURL, "/") {
		upstreamURL += "/"
	}
	upstreamURL += "v1/images/generations"

	upstreamResp, err := router.ForwardRequest(ctx, upstreamURL, provider.APIKey, reqBody, false)
	if err != nil {
		errs.WriteJSONError(w, "failed to forward request: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer upstreamResp.Body.Close()

	respBody, err := io.ReadAll(upstreamResp.Body)
	if err != nil {
		errs.WriteJSONError(w, "failed to read response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(upstreamResp.StatusCode)
	w.Write(respBody)
}
