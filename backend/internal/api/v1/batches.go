package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/google/uuid"
)

// ref: OpenAI Batch API

type Batch struct {
	ID                 string          `json:"id"`
	Object             string          `json:"object"`
	Endpoint           string          `json:"endpoint"`
	Errors             *BatchErrors    `json:"errors,omitempty"`
	InputFileID        string          `json:"input_file_id"`
	CompletionWindow  string          `json:"completion_window"`
	Status             string          `json:"status"`
	OutputFileID       *string         `json:"output_file_id,omitempty"`
	ErrorFileID        *string         `json:"error_file_id,omitempty"`
	CreatedAt          int64           `json:"created_at"`
	InProgressAt       *int64          `json:"in_progress_at,omitempty"`
	ExpiresAt          *int64          `json:"expires_at,omitempty"`
	FinalizingAt       *int64          `json:"finalizing_at,omitempty"`
	CompletedAt        *int64          `json:"completed_at,omitempty"`
	FailedAt           *int64          `json:"failed_at,omitempty"`
	ExpiredAt          *int64          `json:"expired_at,omitempty"`
	CancellingAt       *int64          `json:"cancelling_at,omitempty"`
	CancelledAt        *int64          `json:"cancelled_at,omitempty"`
	RequestCounts      *BatchCounts    `json:"request_counts"`
	Metadata           map[string]any  `json:"metadata,omitempty"`
}

type BatchErrors struct {
	Object string        `json:"object"`
	Data   []BatchError `json:"data"`
}

type BatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
	Line    int    `json:"line,omitempty"`
}

type BatchCounts struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

type BatchStore interface {
	CreateBatch(ctx context.Context, batch *Batch) error
	GetBatch(ctx context.Context, batchID string) (*Batch, error)
	ListBatches(ctx context.Context, limit int, after string) ([]Batch, error)
	CancelBatch(ctx context.Context, batchID string) (*Batch, error)
}

var batchStore BatchStore

func SetBatchStore(store BatchStore) {
	batchStore = store
}

type CreateBatchRequest struct {
	InputFileID       string         `json:"input_file_id"`
	Endpoint          string         `json:"endpoint"`
	CompletionWindow  string         `json:"completion_window"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

func HandleCreateBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if batchStore == nil {
		errs.WriteJSONError(w, "batch operations not configured", http.StatusServiceUnavailable)
		return
	}

	var req CreateBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.InputFileID == "" {
		errs.WriteJSONError(w, "input_file_id is required", http.StatusBadRequest)
		return
	}

	if req.Endpoint == "" {
		errs.WriteJSONError(w, "endpoint is required", http.StatusBadRequest)
		return
	}

	validEndpoints := map[string]bool{
		"/v1/chat/completions":    true,
		"/v1/embeddings":          true,
		"/v1/completions":         true,
	}
	if !validEndpoints[req.Endpoint] {
		errs.WriteJSONError(w, "invalid endpoint. Must be one of: /v1/chat/completions, /v1/embeddings, /v1/completions", http.StatusBadRequest)
		return
	}

	window := req.CompletionWindow
	if window == "" {
		window = "24h"
	}

	now := time.Now().Unix()
	batch := &Batch{
		ID:                "batch-" + uuid.New().String()[:24],
		Object:            "batch",
		Endpoint:          req.Endpoint,
		InputFileID:       req.InputFileID,
		CompletionWindow:  window,
		Status:            "validating",
		CreatedAt:         now,
		RequestCounts:     &BatchCounts{Total: 0, Completed: 0, Failed: 0},
		Metadata:          req.Metadata,
	}

	if err := batchStore.CreateBatch(ctx, batch); err != nil {
		errs.WriteJSONError(w, "failed to create batch: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(batch)
}

func HandleListBatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if batchStore == nil {
		errs.WriteJSONError(w, "batch operations not configured", http.StatusServiceUnavailable)
		return
	}

	limit := 20
	after := r.URL.Query().Get("after")

	jobs, err := batchStore.ListBatches(ctx, limit, after)
	if err != nil {
		errs.WriteJSONError(w, "failed to list batches: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   jobs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleGetBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	batchID := r.PathValue("batch_id")

	if batchID == "" {
		errs.WriteJSONError(w, "batch_id is required", http.StatusBadRequest)
		return
	}

	if batchStore == nil {
		errs.WriteJSONError(w, "batch operations not configured", http.StatusServiceUnavailable)
		return
	}

	batch, err := batchStore.GetBatch(ctx, batchID)
	if err != nil {
		errs.WriteJSONError(w, "batch not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}

func HandleCancelBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	batchID := r.PathValue("batch_id")

	if batchID == "" {
		errs.WriteJSONError(w, "batch_id is required", http.StatusBadRequest)
		return
	}

	if batchStore == nil {
		errs.WriteJSONError(w, "batch operations not configured", http.StatusServiceUnavailable)
		return
	}

	batch, err := batchStore.CancelBatch(ctx, batchID)
	if err != nil {
		errs.WriteJSONError(w, "failed to cancel batch: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}
