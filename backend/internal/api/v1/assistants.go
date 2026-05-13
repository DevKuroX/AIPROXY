package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/google/uuid"
)

// ref: OpenAI Assistants API

type Assistant struct {
	ID           string            `json:"id"`
	Object       string            `json:"object"`
	CreatedAt    int64             `json:"created_at"`
	Name         *string           `json:"name,omitempty"`
	Description  *string           `json:"description,omitempty"`
	Model        string            `json:"model"`
	Instructions *string           `json:"instructions,omitempty"`
	Tools        []Tool            `json:"tools"`
	ToolResources *ToolResources   `json:"tool_resources,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	TopP         *float64          `json:"top_p,omitempty"`
	Temperature  *float64          `json:"temperature,omitempty"`
}

type Tool struct {
	Type     string                 `json:"type"`
	Function *FunctionDefinition    `json:"function,omitempty"`
}

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolResources struct {
	CodeInterpreter *CodeInterpreterResources `json:"code_interpreter,omitempty"`
	FileSearch      *FileSearchResources      `json:"file_search,omitempty"`
}

type CodeInterpreterResources struct {
	FileIDs []string `json:"file_ids"`
}

type FileSearchResources struct {
	VectorStoreIDs []string `json:"vector_store_ids"`
}

type AssistantStore interface {
	CreateAssistant(ctx context.Context, assistant *Assistant) error
	GetAssistant(ctx context.Context, assistantID string) (*Assistant, error)
	ListAssistants(ctx context.Context, limit int, after string, before string) ([]Assistant, error)
	UpdateAssistant(ctx context.Context, assistantID string, updates map[string]interface{}) (*Assistant, error)
	DeleteAssistant(ctx context.Context, assistantID string) error
}

var assistantStore AssistantStore

func SetAssistantStore(store AssistantStore) {
	assistantStore = store
}

type CreateAssistantRequest struct {
	Model        string            `json:"model"`
	Name         *string           `json:"name,omitempty"`
	Description  *string           `json:"description,omitempty"`
	Instructions *string           `json:"instructions,omitempty"`
	Tools        []Tool            `json:"tools,omitempty"`
	ToolResources *ToolResources   `json:"tool_resources,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Temperature  *float64          `json:"temperature,omitempty"`
	TopP         *float64          `json:"top_p,omitempty"`
}

func HandleCreateAssistant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if assistantStore == nil {
		errs.WriteJSONError(w, "assistants not configured", http.StatusServiceUnavailable)
		return
	}

	var req CreateAssistantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		errs.WriteJSONError(w, "model is required", http.StatusBadRequest)
		return
	}

	assistant := &Assistant{
		ID:            "asst_" + uuid.New().String()[:24],
		Object:        "assistant",
		CreatedAt:     time.Now().Unix(),
		Name:          req.Name,
		Description:   req.Description,
		Model:         req.Model,
		Instructions:  req.Instructions,
		Tools:         req.Tools,
		ToolResources: req.ToolResources,
		Metadata:      req.Metadata,
		Temperature:   req.Temperature,
		TopP:          req.TopP,
	}

	if assistant.Tools == nil {
		assistant.Tools = []Tool{}
	}
	if assistant.Metadata == nil {
		assistant.Metadata = map[string]string{}
	}

	if err := assistantStore.CreateAssistant(ctx, assistant); err != nil {
		errs.WriteJSONError(w, "failed to create assistant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(assistant)
}

func HandleListAssistants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if assistantStore == nil {
		errs.WriteJSONError(w, "assistants not configured", http.StatusServiceUnavailable)
		return
	}

	limit := 20
	after := r.URL.Query().Get("after")
	before := r.URL.Query().Get("before")

	assistants, err := assistantStore.ListAssistants(ctx, limit, after, before)
	if err != nil {
		errs.WriteJSONError(w, "failed to list assistants: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   assistants,
		"first_id": func() string { if len(assistants) > 0 { return assistants[0].ID }; return "" }(),
		"last_id":  func() string { if len(assistants) > 0 { return assistants[len(assistants)-1].ID }; return "" }(),
		"has_more": false,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleGetAssistant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assistantID := r.PathValue("assistant_id")

	if assistantID == "" {
		errs.WriteJSONError(w, "assistant_id is required", http.StatusBadRequest)
		return
	}

	if assistantStore == nil {
		errs.WriteJSONError(w, "assistants not configured", http.StatusServiceUnavailable)
		return
	}

	assistant, err := assistantStore.GetAssistant(ctx, assistantID)
	if err != nil {
		errs.WriteJSONError(w, "assistant not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assistant)
}

func HandleUpdateAssistant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assistantID := r.PathValue("assistant_id")

	if assistantID == "" {
		errs.WriteJSONError(w, "assistant_id is required", http.StatusBadRequest)
		return
	}

	if assistantStore == nil {
		errs.WriteJSONError(w, "assistants not configured", http.StatusServiceUnavailable)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	assistant, err := assistantStore.UpdateAssistant(ctx, assistantID, updates)
	if err != nil {
		errs.WriteJSONError(w, "failed to update assistant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assistant)
}

func HandleDeleteAssistant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assistantID := r.PathValue("assistant_id")

	if assistantID == "" {
		errs.WriteJSONError(w, "assistant_id is required", http.StatusBadRequest)
		return
	}

	if assistantStore == nil {
		errs.WriteJSONError(w, "assistants not configured", http.StatusServiceUnavailable)
		return
	}

	err := assistantStore.DeleteAssistant(ctx, assistantID)
	if err != nil {
		errs.WriteJSONError(w, "failed to delete assistant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":      assistantID,
		"object":  "assistant.deleted",
		"deleted": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
