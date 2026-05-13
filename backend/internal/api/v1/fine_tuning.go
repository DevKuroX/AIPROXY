package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/google/uuid"
)

// ref: OpenAI Fine-tuning API

type FineTuningJob struct {
	ID              string                 `json:"id"`
	Object          string                 `json:"object"`
	Model           string                 `json:"model"`
	CreatedAt       int64                  `json:"created_at"`
	FinishedAt      *int64                 `json:"finished_at,omitempty"`
	FineTunedModel  *string                `json:"fine_tuned_model,omitempty"`
	OrganizationID  string                 `json:"organization_id"`
	Status          string                 `json:"status"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
	TrainingFile    string                 `json:"training_file"`
	ValidationFile  *string                `json:"validation_file,omitempty"`
	ResultFiles     []string               `json:"result_files"`
	TrainedTokens   *int                   `json:"trained_tokens,omitempty"`
	Error           *FineTuningError       `json:"error,omitempty"`
}

type FineTuningError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

type FineTuningEvent struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type FineTuningStore interface {
	CreateJob(ctx context.Context, job *FineTuningJob) error
	GetJob(ctx context.Context, jobID string) (*FineTuningJob, error)
	ListJobs(ctx context.Context, limit int, after string) ([]FineTuningJob, error)
	CancelJob(ctx context.Context, jobID string) (*FineTuningJob, error)
	ListJobEvents(ctx context.Context, jobID string) ([]FineTuningEvent, error)
}

var fineTuningStore FineTuningStore

func SetFineTuningStore(store FineTuningStore) {
	fineTuningStore = store
}

type CreateFineTuningJobRequest struct {
	Model           string                 `json:"model"`
	TrainingFile    string                 `json:"training_file"`
	Hyperparameters map[string]interface{} `json:"hyperparameters,omitempty"`
	Suffix          string                 `json:"suffix,omitempty"`
	ValidationFile  string                 `json:"validation_file,omitempty"`
}

func HandleCreateFineTuningJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if fineTuningStore == nil {
		errs.WriteJSONError(w, "fine-tuning not configured", http.StatusServiceUnavailable)
		return
	}

	var req CreateFineTuningJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		errs.WriteJSONError(w, "model is required", http.StatusBadRequest)
		return
	}

	if req.TrainingFile == "" {
		errs.WriteJSONError(w, "training_file is required", http.StatusBadRequest)
		return
	}

	job := &FineTuningJob{
		ID:              "ftjob-" + uuid.New().String()[:24],
		Object:          "fine_tuning.job",
		Model:           req.Model,
		CreatedAt:       time.Now().Unix(),
		OrganizationID:  "org-default",
		Status:          "validating_files",
		Hyperparameters: req.Hyperparameters,
		TrainingFile:    req.TrainingFile,
		ResultFiles:     []string{},
	}

	if req.ValidationFile != "" {
		job.ValidationFile = &req.ValidationFile
	}

	if err := fineTuningStore.CreateJob(ctx, job); err != nil {
		errs.WriteJSONError(w, "failed to create job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}

func HandleListFineTuningJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if fineTuningStore == nil {
		errs.WriteJSONError(w, "fine-tuning not configured", http.StatusServiceUnavailable)
		return
	}

	limit := 20
	after := r.URL.Query().Get("after")
	if l := r.URL.Query().Get("limit"); l != "" {
		var lim int
		if _, err := json.Marshal(l); err == nil {
			lim = limit
		}
		limit = lim
	}

	jobs, err := fineTuningStore.ListJobs(ctx, limit, after)
	if err != nil {
		errs.WriteJSONError(w, "failed to list jobs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   jobs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleGetFineTuningJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := r.PathValue("fine_tuning_job_id")

	if jobID == "" {
		errs.WriteJSONError(w, "fine_tuning_job_id is required", http.StatusBadRequest)
		return
	}

	if fineTuningStore == nil {
		errs.WriteJSONError(w, "fine-tuning not configured", http.StatusServiceUnavailable)
		return
	}

	job, err := fineTuningStore.GetJob(ctx, jobID)
	if err != nil {
		errs.WriteJSONError(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func HandleCancelFineTuningJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := r.PathValue("fine_tuning_job_id")

	if jobID == "" {
		errs.WriteJSONError(w, "fine_tuning_job_id is required", http.StatusBadRequest)
		return
	}

	if fineTuningStore == nil {
		errs.WriteJSONError(w, "fine-tuning not configured", http.StatusServiceUnavailable)
		return
	}

	job, err := fineTuningStore.CancelJob(ctx, jobID)
	if err != nil {
		errs.WriteJSONError(w, "failed to cancel job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func HandleListFineTuningEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := r.PathValue("fine_tuning_job_id")

	if jobID == "" {
		errs.WriteJSONError(w, "fine_tuning_job_id is required", http.StatusBadRequest)
		return
	}

	if fineTuningStore == nil {
		errs.WriteJSONError(w, "fine-tuning not configured", http.StatusServiceUnavailable)
		return
	}

	events, err := fineTuningStore.ListJobEvents(ctx, jobID)
	if err != nil {
		errs.WriteJSONError(w, "failed to list events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   events,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
