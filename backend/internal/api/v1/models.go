package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/router"
)

// ref: 9router/src/app/api/v1/models/route.js
// GET /v1/models - List available models

// Model represents a model in the OpenAI API response
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelsResponse is the response for GET /v1/models
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// ModelsStore provides access to model listings
type ModelsStore interface {
	ListModels() ([]Model, error)
	ListProviders() ([]ProviderInfo, error)
	ListCombos() ([]ComboInfo, error)
	ListModelAliases() ([]AliasInfo, error)
	GetDisabledModels() (map[string]bool, error)
}

// ProviderInfo contains provider model information
type ProviderInfo struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	BaseURL  string   `json:"base_url"`
	Models   []string `json:"models"`
	Enabled  bool     `json:"enabled"`
}

// ComboInfo contains combo configuration
type ComboInfo struct {
	Name    string   `json:"name"`
	Models  []string `json:"models"`
	Enabled bool     `json:"enabled"`
}

// AliasInfo contains model alias information
type AliasInfo struct {
	Alias  string `json:"alias"`
	Target string `json:"target"`
}

var modelsStore ModelsStore

// SetModelsStore sets the models store for the package
func SetModelsStore(store ModelsStore) {
	modelsStore = store
}

// HandleModels handles GET /v1/models
// ref: 9router/src/app/api/v1/models/route.js
func HandleModels(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")

	models := make([]Model, 0)

	// Add static models from provider configurations
	if modelsStore != nil {
		providerModels, err := modelsStore.ListModels()
		if err != nil {
			errs.WriteJSONError(w, "failed to list models: "+err.Error(), http.StatusInternalServerError)
			return
		}
		models = append(models, providerModels...)

		// Add combo models
		combos, err := modelsStore.ListCombos()
		if err == nil {
			for _, combo := range combos {
				if combo.Enabled {
					models = append(models, Model{
						ID:      combo.Name,
						Object:  "model",
						Created: time.Now().Unix(),
						OwnedBy: "combo",
					})
				}
			}
		}

		// Add alias models
		aliases, err := modelsStore.ListModelAliases()
		if err == nil {
			for _, alias := range aliases {
				models = append(models, Model{
					ID:      alias.Alias,
					Object:  "model",
					Created: time.Now().Unix(),
					OwnedBy: "alias",
				})
			}
		}

		// Filter by kind if specified
		if kind != "" {
			models = filterModelsByKind(models, kind)
		}

		disabled, err := modelsStore.GetDisabledModels()
		if err == nil {
			models = filterDisabledModels(models, disabled)
		}
	}

	// If no store, return default models from router aliases
	if modelsStore == nil {
		// Return at least one model to indicate the service is running
		models = append(models, Model{
			ID:      "gpt-4o",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "openai",
		})
	}

	response := ModelsResponse{
		Object: "list",
		Data:   models,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleModelsByKind handles GET /v1/models/{kind}
// ref: 9router/src/app/api/v1/models/[kind]/route.js
func HandleModelsByKind(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	if kind == "" {
		errs.WriteJSONError(w, "kind parameter is required", http.StatusBadRequest)
		return
	}

	// Set the kind query param and delegate to HandleModels
	q := r.URL.Query()
	q.Set("kind", kind)
	r.URL.RawQuery = q.Encode()

	HandleModels(w, r)
}

// filterModelsByKind filters models by their kind/service type
// ref: 9router/src/app/api/v1/models/route.js:MODEL_TYPE_TO_KIND
func filterModelsByKind(models []Model, kind string) []Model {
	// Kind mappings based on model ID patterns
	// llm, embedding, tts, stt, image, imageToText
	filtered := make([]Model, 0)

	for _, m := range models {
		modelKind := inferModelKind(m.ID, m.OwnedBy)
		if modelKind == kind {
			filtered = append(filtered, m)
		}
	}

	return filtered
}

// inferModelKind determines the model kind from its ID and owner
// ref: 9router/src/app/api/v1/models/route.js:inferKindFromUnknownModelId
func inferModelKind(modelID, owner string) string {
	// Check owner/kind hints first
	switch owner {
	case "embedding":
		return "embedding"
	case "tts", "speech", "audio":
		return "tts"
	case "image":
		return "image"
	}

	// Fallback to pattern matching on model ID
	modelIDLower := strings.ToLower(modelID)

	if strings.Contains(modelIDLower, "embed") {
		return "embedding"
	}
	if strings.Contains(modelIDLower, "tts") ||
		strings.Contains(modelIDLower, "speech") ||
		strings.Contains(modelIDLower, "audio") ||
		strings.Contains(modelIDLower, "voice") {
		return "tts"
	}
	if strings.Contains(modelIDLower, "whisper") ||
		strings.Contains(modelIDLower, "stt") {
		return "stt"
	}
	if strings.Contains(modelIDLower, "image") ||
		strings.Contains(modelIDLower, "imagen") ||
		strings.Contains(modelIDLower, "dall-e") ||
		strings.Contains(modelIDLower, "dalle") ||
		strings.Contains(modelIDLower, "flux") ||
		strings.Contains(modelIDLower, "sdxl") ||
		strings.HasPrefix(modelIDLower, "sd-") ||
		strings.Contains(modelIDLower, "stable-diffusion") {
		return "image"
	}

	// Default to LLM
	return "llm"
}

// filterDisabledModels removes disabled models from the list
func filterDisabledModels(models []Model, disabled map[string]bool) []Model {
	filtered := make([]Model, 0, len(models))
	for _, m := range models {
		if !disabled[m.ID] {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// HandleModelInfo handles GET /v1/models/info
// ref: 9router/src/app/api/v1/models/info/route.js
func HandleModelInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	modelID := r.URL.Query().Get("model")

	if modelID == "" {
		errs.WriteJSONError(w, "model parameter is required", http.StatusBadRequest)
		return
	}

	// Resolve alias if needed
	resolvedModel, err := router.ResolveAlias(ctx, modelID)
	if err != nil {
		resolvedModel = modelID
	}

	info := map[string]interface{}{
		"id":        modelID,
		"resolved":  resolvedModel,
		"available": true,
	}

	// Check if model is disabled
	if modelsStore != nil {
		disabled, err := modelsStore.GetDisabledModels()
		if err == nil && disabled[modelID] {
			info["available"] = false
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
