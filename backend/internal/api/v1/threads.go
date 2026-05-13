package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/google/uuid"
)

// ref: OpenAI Threads API

type Thread struct {
	ID           string            `json:"id"`
	Object       string            `json:"object"`
	CreatedAt    int64             `json:"created_at"`
	ToolResources *ToolResources   `json:"tool_resources,omitempty"`
	Metadata     map[string]string `json:"metadata"`
}

type ThreadMessage struct {
	Role        string                 `json:"role"`
	Content     string                 `json:"content"`
	FileIDs     []string               `json:"file_ids,omitempty"`
	Attachments []MessageAttachment    `json:"attachments,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

type MessageAttachment struct {
	FileID string `json:"file_id"`
	Tools  []Tool `json:"tools,omitempty"`
}

type ThreadStore interface {
	CreateThread(ctx context.Context, thread *Thread) error
	GetThread(ctx context.Context, threadID string) (*Thread, error)
	UpdateThread(ctx context.Context, threadID string, updates map[string]interface{}) (*Thread, error)
	DeleteThread(ctx context.Context, threadID string) error
}

var threadStore ThreadStore

func SetThreadStore(store ThreadStore) {
	threadStore = store
}

type CreateThreadRequest struct {
	Messages     []ThreadMessage   `json:"messages,omitempty"`
	ToolResources *ToolResources   `json:"tool_resources,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

func HandleCreateThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if threadStore == nil {
		errs.WriteJSONError(w, "threads not configured", http.StatusServiceUnavailable)
		return
	}

	var req CreateThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	thread := &Thread{
		ID:            "thread_" + uuid.New().String()[:24],
		Object:        "thread",
		CreatedAt:     time.Now().Unix(),
		ToolResources: req.ToolResources,
		Metadata:      req.Metadata,
	}

	if thread.Metadata == nil {
		thread.Metadata = map[string]string{}
	}

	if err := threadStore.CreateThread(ctx, thread); err != nil {
		errs.WriteJSONError(w, "failed to create thread: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(thread)
}

func HandleGetThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threadID := r.PathValue("thread_id")

	if threadID == "" {
		errs.WriteJSONError(w, "thread_id is required", http.StatusBadRequest)
		return
	}

	if threadStore == nil {
		errs.WriteJSONError(w, "threads not configured", http.StatusServiceUnavailable)
		return
	}

	thread, err := threadStore.GetThread(ctx, threadID)
	if err != nil {
		errs.WriteJSONError(w, "thread not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thread)
}

func HandleUpdateThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threadID := r.PathValue("thread_id")

	if threadID == "" {
		errs.WriteJSONError(w, "thread_id is required", http.StatusBadRequest)
		return
	}

	if threadStore == nil {
		errs.WriteJSONError(w, "threads not configured", http.StatusServiceUnavailable)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	thread, err := threadStore.UpdateThread(ctx, threadID, updates)
	if err != nil {
		errs.WriteJSONError(w, "failed to update thread: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thread)
}

func HandleDeleteThread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threadID := r.PathValue("thread_id")

	if threadID == "" {
		errs.WriteJSONError(w, "thread_id is required", http.StatusBadRequest)
		return
	}

	if threadStore == nil {
		errs.WriteJSONError(w, "threads not configured", http.StatusServiceUnavailable)
		return
	}

	err := threadStore.DeleteThread(ctx, threadID)
	if err != nil {
		errs.WriteJSONError(w, "failed to delete thread: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":      threadID,
		"object":  "thread.deleted",
		"deleted": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
