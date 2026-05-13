package admin

import (
	"encoding/json"
	"net/http"
)

type OAuthHandler struct {
}

func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{}
}

func (h *OAuthHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (h *OAuthHandler) Start(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "OAuth start endpoint - implementation pending",
		"provider": provider,
	})
}

func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "OAuth callback endpoint - implementation pending",
		"provider": provider,
	})
}

func (h *OAuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
