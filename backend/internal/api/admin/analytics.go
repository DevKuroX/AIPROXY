package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/DevKuroX/AIPROXY/internal/storage"
	"github.com/google/uuid"
)

type AnalyticsStore interface {
	ListUsageLogs(ctx context.Context, start, end time.Time, provider, model string, page, limit int) ([]models.UsageLog, int, error)
	GetUsageStats(ctx context.Context, start, end time.Time) (*storage.UsageStats, error)
	ListPricingRules(ctx context.Context) ([]models.PricingRule, error)
	CreatePricingRule(ctx context.Context, rule *models.PricingRule) error
	UpdatePricingRule(ctx context.Context, rule *models.PricingRule) error
	DeletePricingRule(ctx context.Context, id string) error
}

type usageListResponse struct {
	Data  []models.UsageLog `json:"data"`
	Total int               `json:"total"`
	Page  int               `json:"page"`
}

type pricingCreateRequest struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	InputPrice  float64 `json:"input_price"`
	OutputPrice float64 `json:"output_price"`
	IsActive    bool    `json:"is_active"`
}

type AnalyticsHandler struct {
	store AnalyticsStore
}

func NewAnalyticsHandler(store AnalyticsStore) *AnalyticsHandler {
	return &AnalyticsHandler{store: store}
}

func (h *AnalyticsHandler) ListUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	start, end, provider, model, page, limit, err := parseUsageQueryParams(r)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	logs, total, err := h.store.ListUsageLogs(ctx, start, end, provider, model, page, limit)
	if err != nil {
		errs.WriteJSONError(w, "failed to list usage logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usageListResponse{
		Data:  logs,
		Total: total,
		Page:  page,
	})
}

func (h *AnalyticsHandler) GetUsageStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	start, end, _, _, _, _, err := parseUsageQueryParams(r)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	stats, err := h.store.GetUsageStats(ctx, start, end)
	if err != nil {
		errs.WriteJSONError(w, "failed to get usage stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AnalyticsHandler) ListPricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rules, err := h.store.ListPricingRules(ctx)
	if err != nil {
		errs.WriteJSONError(w, "failed to list pricing rules: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": rules})
}

func (h *AnalyticsHandler) CreatePricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req pricingCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider == "" || req.Model == "" {
		errs.WriteJSONError(w, "provider and model are required", http.StatusBadRequest)
		return
	}

	rule := &models.PricingRule{
		ID:          uuid.New().String(),
		Provider:    req.Provider,
		Model:       req.Model,
		InputPrice:  req.InputPrice,
		OutputPrice: req.OutputPrice,
		IsActive:    req.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.store.CreatePricingRule(ctx, rule); err != nil {
		errs.WriteJSONError(w, "failed to create pricing rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (h *AnalyticsHandler) UpdatePricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	if id == "" {
		errs.WriteJSONError(w, "pricing rule id is required", http.StatusBadRequest)
		return
	}

	var req pricingCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rule := &models.PricingRule{
		ID:          id,
		Provider:    req.Provider,
		Model:       req.Model,
		InputPrice:  req.InputPrice,
		OutputPrice: req.OutputPrice,
		IsActive:    req.IsActive,
		UpdatedAt:   time.Now(),
	}

	if err := h.store.UpdatePricingRule(ctx, rule); err != nil {
		errs.WriteJSONError(w, "failed to update pricing rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

func (h *AnalyticsHandler) DeletePricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	if id == "" {
		errs.WriteJSONError(w, "pricing rule id is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeletePricingRule(ctx, id); err != nil {
		errs.WriteJSONError(w, "failed to delete pricing rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseUsageQueryParams(r *http.Request) (start, end time.Time, provider, model string, page, limit int, err error) {
	now := time.Now()
	start = now.AddDate(0, 0, -7)
	end = now

	if s := r.URL.Query().Get("start"); s != "" {
		t, parseErr := time.Parse(time.RFC3339, s)
		if parseErr != nil {
			err = parseErr
			return
		}
		start = t
	}

	if e := r.URL.Query().Get("end"); e != "" {
		t, parseErr := time.Parse(time.RFC3339, e)
		if parseErr != nil {
			err = parseErr
			return
		}
		end = t
	}

	provider = r.URL.Query().Get("provider")
	model = r.URL.Query().Get("model")

	page = 1
	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
		if page < 1 {
			page = 1
		}
	}

	limit = 50
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
		if limit < 1 || limit > 100 {
			limit = 50
		}
	}

	return
}
