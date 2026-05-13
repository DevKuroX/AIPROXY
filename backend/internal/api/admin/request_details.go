package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type RequestDetailsHandler struct {
	storage *storage.RequestDetailStore
}

func NewRequestDetailsHandler(store *storage.RequestDetailStore) *RequestDetailsHandler {
	return &RequestDetailsHandler{
		storage: store,
	}
}

func (h *RequestDetailsHandler) HandleRequestDetails(w http.ResponseWriter, r *http.Request) {
	filters := storage.RequestDetailFilters{
		Limit:  parseInt(r.URL.Query().Get("limit")),
		Offset: parseInt(r.URL.Query().Get("offset")),
	}

	startTime := parseTime(r.URL.Query().Get("start_time"))
	if !startTime.IsZero() {
		filters.StartTime = &startTime
	}
	endTime := parseTime(r.URL.Query().Get("end_time"))
	if !endTime.IsZero() {
		filters.EndTime = &endTime
	}
	if providerID := parseInt64(r.URL.Query().Get("provider_id")); providerID > 0 {
		filters.ProviderID = &providerID
	}
	if model := r.URL.Query().Get("model"); model != "" {
		filters.Model = &model
	}
	if statusCode := parseInt(r.URL.Query().Get("status_code")); statusCode > 0 {
		filters.StatusCode = &statusCode
	}

	details, err := h.storage.GetRequestDetails(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]map[string]any, len(details))
	for i, d := range details {
		resp := map[string]any{
			"id":        d.ID,
			"timestamp": d.Timestamp,
			"method":    d.Method,
			"path":      d.Path,
		}
		if d.StatusCode != nil {
			resp["status_code"] = *d.StatusCode
		}
		if d.DurationMs != nil {
			resp["duration_ms"] = *d.DurationMs
		}
		if d.Error != nil {
			resp["error"] = *d.Error
		}
		if d.ProviderID != nil {
			resp["provider_id"] = *d.ProviderID
		}
		if d.AccountID != nil {
			resp["account_id"] = *d.AccountID
		}
		if d.Model != nil {
			resp["model"] = *d.Model
		}
		if d.TokensPrompt != nil {
			resp["tokens_prompt"] = *d.TokensPrompt
		}
		if d.TokensCompletion != nil {
			resp["tokens_completion"] = *d.TokensCompletion
		}
		response[i] = resp
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RequestDetailsHandler) HandleRequestDetail(w http.ResponseWriter, r *http.Request, id string) {
	detail, err := h.storage.GetRequestDetailByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := map[string]any{
		"id":        detail.ID,
		"timestamp": detail.Timestamp,
		"method":    detail.Method,
		"path":      detail.Path,
	}
	if detail.StatusCode != nil {
		resp["status_code"] = *detail.StatusCode
	}
	if detail.DurationMs != nil {
		resp["duration_ms"] = *detail.DurationMs
	}
	if detail.Error != nil {
		resp["error"] = *detail.Error
	}
	if detail.ProviderID != nil {
		resp["provider_id"] = *detail.ProviderID
	}
	if detail.AccountID != nil {
		resp["account_id"] = *detail.AccountID
	}
	if detail.Model != nil {
		resp["model"] = *detail.Model
	}
	if detail.TokensPrompt != nil {
		resp["tokens_prompt"] = *detail.TokensPrompt
	}
	if detail.TokensCompletion != nil {
		resp["tokens_completion"] = *detail.TokensCompletion
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *RequestDetailsHandler) HandleDeleteRequestDetails(w http.ResponseWriter, r *http.Request) {
	olderThan := parseTime(r.URL.Query().Get("older_than"))
	if olderThan.IsZero() {
		olderThan = time.Now().Add(-24 * time.Hour)
	}

	err := h.storage.DeleteOldRequestDetails(r.Context(), olderThan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseInt(s string) int {
	var v int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + int(c-'0')
		}
	}
	return v
}

func parseInt64(s string) int64 {
	var v int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + int64(c-'0')
		}
	}
	return v
}
