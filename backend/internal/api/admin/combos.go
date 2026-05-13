package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type ComboStore interface {
	ListCombos(ctx context.Context) ([]models.Combo, error)
	GetComboByName(ctx context.Context, name string) (*models.Combo, error)
	CreateCombo(ctx context.Context, combo *models.Combo) error
	UpdateCombo(ctx context.Context, combo *models.Combo) error
	DeleteCombo(ctx context.Context, name string) error
}

type ComboHandler struct {
	store ComboStore
}

func NewComboHandler(store ComboStore) *ComboHandler {
	return &ComboHandler{store: store}
}

func (h *ComboHandler) List(w http.ResponseWriter, r *http.Request) {
	combos, err := h.store.ListCombos(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(combos)
}

func (h *ComboHandler) Create(w http.ResponseWriter, r *http.Request) {
	var combo models.Combo

	if err := json.NewDecoder(r.Body).Decode(&combo); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if combo.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if len(combo.Models) == 0 {
		http.Error(w, "At least one model is required", http.StatusBadRequest)
		return
	}

	if err := h.store.CreateCombo(r.Context(), &combo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(combo)
}

func (h *ComboHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Combo ID is required", http.StatusBadRequest)
		return
	}

	existing, err := h.store.GetComboByName(r.Context(), id)
	if err != nil {
		if err == storage.ErrComboNotFound {
			http.Error(w, "Combo not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updates struct {
		Models      []string `json:"models"`
		Strategy    *string  `json:"strategy"`
		StickyLimit *int     `json:"sticky_limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(updates.Models) > 0 {
		existing.Models = updates.Models
	}
	if updates.Strategy != nil {
		existing.Strategy = *updates.Strategy
	}
	if updates.StickyLimit != nil {
		existing.StickyLimit = *updates.StickyLimit
	}

	if err := h.store.UpdateCombo(r.Context(), existing); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

func (h *ComboHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Combo ID is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteCombo(r.Context(), id); err != nil {
		if err == storage.ErrComboNotFound {
			http.Error(w, "Combo not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
