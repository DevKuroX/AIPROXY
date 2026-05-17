package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/auth"
	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type KeyStore interface {
	ListAPIKeys(ctx context.Context) ([]models.APIKey, error)
	GetAPIKeyByID(ctx context.Context, id string) (*models.APIKey, error)
	CreateAPIKey(ctx context.Context, key *models.APIKey) error
	DeleteAPIKey(ctx context.Context, id string) error
}

type KeyHandler struct {
	store  KeyStore
	secret string
}

func NewKeyHandler(store KeyStore, secret string) *KeyHandler {
	return &KeyHandler{store: store, secret: secret}
}

func (h *KeyHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.store.ListAPIKeys(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	safeKeys := make([]map[string]interface{}, len(keys))
	for i, k := range keys {
		safeKeys[i] = map[string]interface{}{
			"id":           k.ID,
			"name":         k.Name,
			"is_active":    k.IsActive,
			"created_at":   k.CreatedAt,
			"last_used_at": k.LastUsedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"keys": safeKeys,
	})
}

func (h *KeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	key, keyHash, err := auth.GenerateAPIKey(h.secret)
	if err != nil {
		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	apiKey := &models.APIKey{
		Key:      key,
		KeyHash:  keyHash,
		Name:     req.Name,
		IsActive: true,
	}

	if err := h.store.CreateAPIKey(r.Context(), apiKey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         apiKey.ID,
		"key":        apiKey.Key,
		"name":       apiKey.Name,
		"is_active":  apiKey.IsActive,
		"created_at": apiKey.CreatedAt,
	})
}

func (h *KeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "API key ID is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteAPIKey(r.Context(), id); err != nil {
		if err == storage.ErrAPIKeyNotFound {
			http.Error(w, "API key not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
