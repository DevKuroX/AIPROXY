package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/models"
)

type AliasStore interface {
	CreateModelAlias(ctx context.Context, alias *models.ModelAlias) error
	GetModelAliasByAlias(ctx context.Context, aliasName string) (*models.ModelAlias, error)
	ListModelAliases(ctx context.Context) ([]models.ModelAlias, error)
	DeleteModelAlias(ctx context.Context, id string) error
}

type AliasHandler struct {
	store AliasStore
}

func NewAliasHandler(store AliasStore) *AliasHandler {
	return &AliasHandler{store: store}
}

type createAliasRequest struct {
	NodeID      string `json:"node_id"`
	Alias       string `json:"alias"`
	TargetModel string `json:"target_model"`
}

func (h *AliasHandler) ListAliases(w http.ResponseWriter, r *http.Request) {
	aliases, err := h.store.ListModelAliases(r.Context())
	if err != nil {
		errs.WriteJSONError(w, "failed to list aliases: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aliases)
}

func (h *AliasHandler) CreateAlias(w http.ResponseWriter, r *http.Request) {
	var req createAliasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Alias == "" || req.TargetModel == "" {
		errs.WriteJSONError(w, "alias and target_model are required", http.StatusBadRequest)
		return
	}

	alias := &models.ModelAlias{
		NodeID:      req.NodeID,
		Alias:       req.Alias,
		TargetModel: req.TargetModel,
	}

	if err := h.store.CreateModelAlias(r.Context(), alias); err != nil {
		errs.WriteJSONError(w, "failed to create alias: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alias)
}

func (h *AliasHandler) DeleteAlias(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errs.WriteJSONError(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteModelAlias(r.Context(), id); err != nil {
		errs.WriteJSONError(w, "failed to delete alias: "+err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
