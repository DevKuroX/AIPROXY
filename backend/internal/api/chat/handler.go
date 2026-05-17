package chat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/api/middleware"
	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/providers"
	"github.com/DevKuroX/AIPROXY/internal/router"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type Handler struct {
	db        *storage.DB
	apiSecret string
}

func NewHandler(db *storage.DB, apiSecret string) *Handler {
	return &Handler{db: db, apiSecret: apiSecret}
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		errs.WriteJSONError(w, "invalid user", http.StatusUnauthorized)
		return
	}

	conversations, err := h.db.ListConversations(r.Context(), userID)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conversations == nil {
		conversations = []storage.Conversation{}
	}

	type sessionItem struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		UpdatedAt string `json:"updated_at"`
	}

	sessions := make([]sessionItem, len(conversations))
	for i, c := range conversations {
		sessions[i] = sessionItem{
			ID:        c.ID,
			Title:     c.Title,
			UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
	})
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		errs.WriteJSONError(w, "invalid user", http.StatusUnauthorized)
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Title == "" {
		req.Title = "New Chat"
	}

	conv, err := h.db.CreateConversation(r.Context(), userID, req.Title)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         conv.ID,
		"title":      conv.Title,
		"created_at": conv.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	})
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		errs.WriteJSONError(w, "invalid user", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		errs.WriteJSONError(w, "session id required", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteConversation(r.Context(), id, userID); err != nil {
		if err == storage.ErrConversationNotFound {
			errs.WriteJSONError(w, "session not found", http.StatusNotFound)
			return
		}
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"deleted": true,
	})
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		errs.WriteJSONError(w, "session id required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	before := r.URL.Query().Get("before")

	messages, err := h.db.ListMessages(r.Context(), sessionID, limit, before)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if messages == nil {
		messages = []storage.Message{}
	}

	type msgItem struct {
		ID         string  `json:"id"`
		Role       string  `json:"role"`
		Content    string  `json:"content"`
		ArtifactID *string `json:"artifact_id,omitempty"`
		CreatedAt  string  `json:"created_at"`
	}

	msgs := make([]msgItem, len(messages))
	for i, m := range messages {
		msgs[i] = msgItem{
			ID:         m.ID,
			Role:       m.Role,
			Content:    m.Content,
			ArtifactID: m.ArtifactID,
			CreatedAt:  m.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": msgs,
	})
}

type SaveMessageRequest struct {
	Role       string  `json:"role"`
	Content    string  `json:"content"`
	ArtifactID *string `json:"artifact_id,omitempty"`
}

func (h *Handler) SaveMessage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		errs.WriteJSONError(w, "session id required", http.StatusBadRequest)
		return
	}

	var req SaveMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		errs.WriteJSONError(w, "role is required", http.StatusBadRequest)
		return
	}
	if req.Content == "" && req.ArtifactID == nil {
		errs.WriteJSONError(w, "content or artifact_id required", http.StatusBadRequest)
		return
	}

	msg, err := h.db.CreateMessage(r.Context(), sessionID, req.Role, req.Content, req.ArtifactID)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          msg.ID,
		"role":        msg.Role,
		"content":     msg.Content,
		"artifact_id": msg.ArtifactID,
		"created_at":  msg.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	})
}

func (h *Handler) GetArtifact(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		errs.WriteJSONError(w, "artifact id required", http.StatusBadRequest)
		return
	}

	artifact, err := h.db.GetArtifact(r.Context(), id)
	if err != nil {
		if err == storage.ErrArtifactNotFound {
			errs.WriteJSONError(w, "artifact not found", http.StatusNotFound)
			return
		}
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":              artifact.ID,
		"conversation_id": artifact.ConversationID,
		"type":            artifact.Type,
		"payload":         json.RawMessage(artifact.Payload),
		"created_at":      artifact.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	})
}

type CreateArtifactRequest struct {
	ConversationID string `json:"conversation_id"`
	Type           string `json:"type"`
	Payload        json.RawMessage `json:"payload"`
}

func (h *Handler) CreateArtifact(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateArtifactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ConversationID == "" || req.Type == "" || len(req.Payload) == 0 {
		errs.WriteJSONError(w, "conversation_id, type, and payload required", http.StatusBadRequest)
		return
	}

	artifact, err := h.db.CreateArtifact(r.Context(), req.ConversationID, req.Type, string(req.Payload))
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":              artifact.ID,
		"conversation_id": artifact.ConversationID,
		"type":            artifact.Type,
		"created_at":      artifact.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
	})
}

func (h *Handler) GenerateTitle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		errs.WriteJSONError(w, "session id required", http.StatusBadRequest)
		return
	}

	msgs, err := h.db.ListMessages(r.Context(), sessionID, 1, "")
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(msgs) == 0 {
		errs.WriteJSONError(w, "no messages", http.StatusBadRequest)
		return
	}

	content := msgs[0].Content
	if len(content) > 50 {
		content = content[:50] + "..."
	}

	title := content
	if len(title) > 60 {
		title = title[:60]
	}
	if title == "" {
		title = "New Chat"
	}

	if err := h.db.UpdateConversationTitle(r.Context(), sessionID, title); err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"title": title})
}

func (h *Handler) ListProviderModels(w http.ResponseWriter, r *http.Request) {
	providerID := r.URL.Query().Get("provider")
	if providerID == "" {
		errs.WriteJSONError(w, "provider query param required", http.StatusBadRequest)
		return
	}

	cfg, ok := providers.GetProviderConfig(providerID)
	if !ok {
		errs.WriteJSONError(w, "unknown provider", http.StatusNotFound)
		return
	}

	var modelURL string
	switch providerID {
	case "oc", "opencode":
		modelURL = "https://opencode.ai/zen/v1/models"
	default:
		errs.WriteJSONError(w, "model listing not supported for this provider", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), "GET", modelURL, nil)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *Handler) StreamCompletion(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	router.HandleChatCompletions(w, r)
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		errs.WriteJSONError(w, "file too large", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		errs.WriteJSONError(w, "no file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sessionID := r.FormValue("session_id")
	dir := fmt.Sprintf("/tmp/aiproxy-uploads/%s", sessionID)
	os.MkdirAll(dir, 0755)

	dst, err := os.Create(filepath.Join(dir, header.Filename))
	if err != nil {
		errs.WriteJSONError(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	io.Copy(dst, file)
	dst.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   sessionID + "/" + header.Filename,
		"name": header.Filename,
		"size": header.Size,
		"url":  "/uploads/" + sessionID + "/" + header.Filename,
	})
}
