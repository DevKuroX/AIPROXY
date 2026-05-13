package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/models"
)

// ref: 9router Provider Account Management

type AccountStore interface {
	ListProviderAccounts(ctx context.Context, providerID string) ([]models.ProviderAccount, error)
	GetProviderAccount(ctx context.Context, accountID string) (*models.ProviderAccount, error)
	CreateProviderAccount(ctx context.Context, account *models.ProviderAccount) error
	UpdateProviderAccount(ctx context.Context, account *models.ProviderAccount) error
	DeleteProviderAccount(ctx context.Context, accountID string) error
}

type AccountHandler struct {
	store AccountStore
}

func NewAccountHandler(store AccountStore) *AccountHandler {
	return &AccountHandler{store: store}
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	providerID := r.URL.Query().Get("provider_id")

	accounts, err := h.store.ListProviderAccounts(r.Context(), providerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	safeAccounts := make([]map[string]interface{}, len(accounts))
	for i, a := range accounts {
		safeAccounts[i] = map[string]interface{}{
			"id":          a.ID,
			"provider_id": a.ProviderID,
			"name":        a.Name,
			"is_active":   a.IsActive,
			"created_at":  a.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": safeAccounts,
	})
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProviderID string `json:"provider_id"`
		Name       string `json:"name"`
		APIKey     string `json:"api_key"`
		IsActive   bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProviderID == "" {
		http.Error(w, "provider_id is required", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.APIKey == "" {
		http.Error(w, "api_key is required", http.StatusBadRequest)
		return
	}

	account := &models.ProviderAccount{
		ProviderID: req.ProviderID,
		Name:       req.Name,
		APIKey:     req.APIKey,
		IsActive:   req.IsActive,
	}

	if err := h.store.CreateProviderAccount(r.Context(), account); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          account.ID,
		"provider_id": account.ProviderID,
		"name":        account.Name,
		"is_active":   account.IsActive,
		"created_at":  account.CreatedAt,
	})
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	existing, err := h.store.GetProviderAccount(r.Context(), accountID)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	var updates struct {
		Name     *string `json:"name"`
		APIKey   *string `json:"api_key"`
		IsActive *bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if updates.Name != nil {
		existing.Name = *updates.Name
	}
	if updates.APIKey != nil {
		existing.APIKey = *updates.APIKey
	}
	if updates.IsActive != nil {
		existing.IsActive = *updates.IsActive
	}

	if err := h.store.UpdateProviderAccount(r.Context(), existing); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          existing.ID,
		"provider_id": existing.ProviderID,
		"name":        existing.Name,
		"is_active":   existing.IsActive,
	})
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteProviderAccount(r.Context(), accountID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      accountID,
		"deleted": true,
	})
}

func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	account, err := h.store.GetProviderAccount(r.Context(), accountID)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          account.ID,
		"provider_id": account.ProviderID,
		"name":        account.Name,
		"is_active":   account.IsActive,
		"created_at":  account.CreatedAt,
	})
}
