package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/models"
)

type NodeStore interface {
	CreateProviderNode(ctx context.Context, node *models.ProviderNode) error
	GetProviderNodeByID(ctx context.Context, id string) (*models.ProviderNode, error)
	ListProviderNodes(ctx context.Context) ([]models.ProviderNode, error)
	UpdateProviderNode(ctx context.Context, node *models.ProviderNode) error
	DeleteProviderNode(ctx context.Context, id string) error
}

type nodeCreateRequest struct {
	Name             string `json:"name"`
	BaseURL          string `json:"base_url"`
	APIKey           string `json:"api_key"`
	CompatibleFormat string `json:"compatible_format"`
	Enabled          bool   `json:"enabled"`
}

type nodeUpdateRequest struct {
	Name             *string `json:"name"`
	BaseURL          *string `json:"base_url"`
	APIKey           *string `json:"api_key"`
	CompatibleFormat *string `json:"compatible_format"`
	Enabled          *bool   `json:"enabled"`
}

type NodeHandler struct {
	store NodeStore
}

func NewNodeHandler(store NodeStore) *NodeHandler {
	return &NodeHandler{store: store}
}

func (h *NodeHandler) ListNodes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	nodes, err := h.store.ListProviderNodes(ctx)
	if err != nil {
		errs.WriteJSONError(w, "failed to list nodes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if nodes == nil {
		nodes = []models.ProviderNode{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

func (h *NodeHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req nodeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		errs.WriteJSONError(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.BaseURL == "" {
		errs.WriteJSONError(w, "base_url is required", http.StatusBadRequest)
		return
	}
	if req.APIKey == "" {
		errs.WriteJSONError(w, "api_key is required", http.StatusBadRequest)
		return
	}

	compatibleFormat := req.CompatibleFormat
	if compatibleFormat == "" {
		compatibleFormat = "openai"
	}

	node := &models.ProviderNode{
		Name:             req.Name,
		BaseURL:          req.BaseURL,
		APIKey:           req.APIKey,
		CompatibleFormat: compatibleFormat,
		Enabled:          req.Enabled,
	}

	if err := h.store.CreateProviderNode(ctx, node); err != nil {
		errs.WriteJSONError(w, "failed to create node: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

func (h *NodeHandler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	if id == "" {
		errs.WriteJSONError(w, "id is required", http.StatusBadRequest)
		return
	}

	existing, err := h.store.GetProviderNodeByID(ctx, id)
	if err != nil {
		errs.WriteJSONError(w, "node not found", http.StatusNotFound)
		return
	}

	var req nodeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.BaseURL != nil {
		existing.BaseURL = *req.BaseURL
	}
	if req.APIKey != nil {
		existing.APIKey = *req.APIKey
	}
	if req.CompatibleFormat != nil {
		existing.CompatibleFormat = *req.CompatibleFormat
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	if err := h.store.UpdateProviderNode(ctx, existing); err != nil {
		errs.WriteJSONError(w, "failed to update node: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

func (h *NodeHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	if id == "" {
		errs.WriteJSONError(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteProviderNode(ctx, id); err != nil {
		errs.WriteJSONError(w, "failed to delete node: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NodeHandler) TestNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	if id == "" {
		errs.WriteJSONError(w, "id is required", http.StatusBadRequest)
		return
	}

	node, err := h.store.GetProviderNodeByID(ctx, id)
	if err != nil {
		errs.WriteJSONError(w, "node not found", http.StatusNotFound)
		return
	}

	testReq := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": "Say 'test'"},
		},
		"max_tokens": 5,
	}

	testBody, _ := json.Marshal(testReq)

	baseURL := strings.TrimSuffix(node.BaseURL, "/")
	testURL := baseURL + "/v1/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", testURL, strings.NewReader(string(testBody)))
	if err != nil {
		errs.WriteJSONError(w, "failed to create test request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+node.APIKey)

	client := &http.Client{Timeout: 10}
	resp, err := client.Do(httpReq)
	if err != nil {
		errs.WriteJSONError(w, "connection test failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status_code":  resp.StatusCode,
		"message":      "Connection test completed",
	})
}
