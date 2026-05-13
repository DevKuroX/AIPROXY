package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type ProviderStore interface {
	ListProviders(ctx context.Context) ([]models.Provider, error)
	GetProviderByID(ctx context.Context, id string) (*models.Provider, error)
	CreateProvider(ctx context.Context, provider *models.Provider) error
	UpdateProvider(ctx context.Context, provider *models.Provider) error
	DeleteProvider(ctx context.Context, id string) error
}

type ProviderHandler struct {
	store ProviderStore
}

func NewProviderHandler(store ProviderStore) *ProviderHandler {
	return &ProviderHandler{store: store}
}

func (h *ProviderHandler) List(w http.ResponseWriter, r *http.Request) {
	providers, err := h.store.ListProviders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	safeProviders := make([]map[string]interface{}, len(providers))
	for i, p := range providers {
		safeProviders[i] = map[string]interface{}{
			"id":         p.ID,
			"name":       p.Name,
			"type":       p.Type,
			"base_url":   p.BaseURL,
			"enabled":    p.Enabled,
			"created_at": p.CreatedAt,
			"updated_at": p.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeProviders)
}

func (h *ProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
		Enabled bool   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" {
		http.Error(w, "Name and type are required", http.StatusBadRequest)
		return
	}

	provider := &models.Provider{
		Name:    req.Name,
		Type:    req.Type,
		BaseURL: req.BaseURL,
		APIKey:  req.APIKey,
		Enabled: req.Enabled,
	}

	if err := h.store.CreateProvider(r.Context(), provider); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         provider.ID,
		"name":       provider.Name,
		"type":       provider.Type,
		"base_url":   provider.BaseURL,
		"enabled":    provider.Enabled,
		"created_at": provider.CreatedAt,
		"updated_at": provider.UpdatedAt,
	})
}

func (h *ProviderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Provider ID is required", http.StatusBadRequest)
		return
	}

	provider, err := h.store.GetProviderByID(r.Context(), id)
	if err != nil {
		if err == storage.ErrProviderNotFound {
			http.Error(w, "Provider not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req struct {
		Name    *string `json:"name"`
		Type    *string `json:"type"`
		BaseURL *string `json:"base_url"`
		APIKey  *string `json:"api_key"`
		Enabled *bool   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.Type != nil {
		provider.Type = *req.Type
	}
	if req.BaseURL != nil {
		provider.BaseURL = *req.BaseURL
	}
	if req.APIKey != nil {
		provider.APIKey = *req.APIKey
	}
	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
	}

	if err := h.store.UpdateProvider(r.Context(), provider); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         provider.ID,
		"name":       provider.Name,
		"type":       provider.Type,
		"base_url":   provider.BaseURL,
		"enabled":    provider.Enabled,
		"created_at": provider.CreatedAt,
		"updated_at": provider.UpdatedAt,
	})
}

func (h *ProviderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Provider ID is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteProvider(r.Context(), id); err != nil {
		if err == storage.ErrProviderNotFound {
			http.Error(w, "Provider not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProviderHandler) Test(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Provider ID is required", http.StatusBadRequest)
		return
	}

	provider, err := h.store.GetProviderByID(r.Context(), id)
	if err != nil {
		if err == storage.ErrProviderNotFound {
			http.Error(w, "Provider not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Provider test endpoint - implementation pending",
		"provider": map[string]interface{}{
			"id":   provider.ID,
			"name": provider.Name,
			"type": provider.Type,
		},
	})
}
