package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/google/uuid"
)

// ref: OpenAI Messages API
// ref: 9router/src/app/api/v1/messages/route.js

type Message struct {
	ID          string             `json:"id"`
	Object      string             `json:"object"`
	CreatedAt   int64              `json:"created_at"`
	ThreadID    string             `json:"thread_id"`
	Role        string             `json:"role"`
	Content     []MessageContent   `json:"content"`
	AssistantID *string            `json:"assistant_id,omitempty"`
	RunID       *string            `json:"run_id,omitempty"`
	Attachments []MessageAttachment `json:"attachments,omitempty"`
	Metadata    map[string]string  `json:"metadata"`
}

type MessageContent struct {
	Type      string             `json:"type"`
	Text      *string            `json:"text,omitempty"`
	ImageFile *MessageImageFile  `json:"image_file,omitempty"`
	ImageURL  *MessageImageURL   `json:"image_url,omitempty"`
}

type MessageImageFile struct {
	FileID string `json:"file_id"`
	Detail string `json:"detail,omitempty"`
}

type MessageImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type MessageStore interface {
	CreateMessage(ctx context.Context, message *Message) error
	GetMessage(ctx context.Context, threadID, messageID string) (*Message, error)
	ListMessages(ctx context.Context, threadID string, limit int, after string, before string) ([]Message, error)
	UpdateMessage(ctx context.Context, threadID, messageID string, updates map[string]interface{}) (*Message, error)
	DeleteMessage(ctx context.Context, threadID, messageID string) error
}

var messageStore MessageStore

func SetMessageStore(store MessageStore) {
	messageStore = store
}

type CreateMessageRequest struct {
	Role        string              `json:"role"`
	Content     interface{}         `json:"content"`
	Attachments []MessageAttachment `json:"attachments,omitempty"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
}

func HandleCreateMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threadID := r.PathValue("thread_id")

	if threadID == "" {
		errs.WriteJSONError(w, "thread_id is required", http.StatusBadRequest)
		return
	}

	if messageStore == nil {
		errs.WriteJSONError(w, "messages not configured", http.StatusServiceUnavailable)
		return
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		errs.WriteJSONError(w, "role is required", http.StatusBadRequest)
		return
	}

	if req.Role != "user" && req.Role != "assistant" {
		errs.WriteJSONError(w, "role must be 'user' or 'assistant'", http.StatusBadRequest)
		return
	}

	content := parseMessageContent(req.Content)

	message := &Message{
		ID:          "msg_" + uuid.New().String()[:24],
		Object:      "thread.message",
		CreatedAt:   time.Now().Unix(),
		ThreadID:    threadID,
		Role:        req.Role,
		Content:     content,
		Attachments: req.Attachments,
		Metadata:    req.Metadata,
	}

	if message.Metadata == nil {
		message.Metadata = map[string]string{}
	}

	if err := messageStore.CreateMessage(ctx, message); err != nil {
		errs.WriteJSONError(w, "failed to create message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message)
}

func HandleListMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threadID := r.PathValue("thread_id")

	if threadID == "" {
		errs.WriteJSONError(w, "thread_id is required", http.StatusBadRequest)
		return
	}

	if messageStore == nil {
		errs.WriteJSONError(w, "messages not configured", http.StatusServiceUnavailable)
		return
	}

	limit := 20
	after := r.URL.Query().Get("after")
	before := r.URL.Query().Get("before")

	messages, err := messageStore.ListMessages(ctx, threadID, limit, after, before)
	if err != nil {
		errs.WriteJSONError(w, "failed to list messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"object":   "list",
		"data":     messages,
		"first_id": func() string { if len(messages) > 0 { return messages[0].ID }; return "" }(),
		"last_id":  func() string { if len(messages) > 0 { return messages[len(messages)-1].ID }; return "" }(),
		"has_more": false,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleGetMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threadID := r.PathValue("thread_id")
	messageID := r.PathValue("message_id")

	if threadID == "" || messageID == "" {
		errs.WriteJSONError(w, "thread_id and message_id are required", http.StatusBadRequest)
		return
	}

	if messageStore == nil {
		errs.WriteJSONError(w, "messages not configured", http.StatusServiceUnavailable)
		return
	}

	message, err := messageStore.GetMessage(ctx, threadID, messageID)
	if err != nil {
		errs.WriteJSONError(w, "message not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func HandleUpdateMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threadID := r.PathValue("thread_id")
	messageID := r.PathValue("message_id")

	if threadID == "" || messageID == "" {
		errs.WriteJSONError(w, "thread_id and message_id are required", http.StatusBadRequest)
		return
	}

	if messageStore == nil {
		errs.WriteJSONError(w, "messages not configured", http.StatusServiceUnavailable)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	message, err := messageStore.UpdateMessage(ctx, threadID, messageID, updates)
	if err != nil {
		errs.WriteJSONError(w, "failed to update message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func HandleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	threadID := r.PathValue("thread_id")
	messageID := r.PathValue("message_id")

	if threadID == "" || messageID == "" {
		errs.WriteJSONError(w, "thread_id and message_id are required", http.StatusBadRequest)
		return
	}

	if messageStore == nil {
		errs.WriteJSONError(w, "messages not configured", http.StatusServiceUnavailable)
		return
	}

	err := messageStore.DeleteMessage(ctx, threadID, messageID)
	if err != nil {
		errs.WriteJSONError(w, "failed to delete message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":      messageID,
		"object":  "thread.message.deleted",
		"deleted": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type CountTokensRequest struct {
	Messages []CreateMessageRequest `json:"messages"`
	Model    string                 `json:"model"`
}

type CountTokensResponse struct {
	TotalTokens int `json:"total_tokens"`
}

// ref: 9router/src/app/api/v1/messages/count_tokens/route.js
func HandleCountTokens(w http.ResponseWriter, r *http.Request) {
	var req CountTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	totalTokens := 0
	for _, msg := range req.Messages {
		switch v := msg.Content.(type) {
		case string:
			totalTokens += len(v) / 4
		case []interface{}:
			for _, c := range v {
				if cm, ok := c.(map[string]interface{}); ok {
					if text, ok := cm["text"].(string); ok {
						totalTokens += len(text) / 4
					}
				}
			}
		}
	}

	response := CountTokensResponse{
		TotalTokens: totalTokens,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func parseMessageContent(content interface{}) []MessageContent {
	switch v := content.(type) {
	case string:
		return []MessageContent{{
			Type: "text",
			Text: &v,
		}}
	case []interface{}:
		var result []MessageContent
		for _, c := range v {
			if cm, ok := c.(map[string]interface{}); ok {
				mc := MessageContent{}
				if t, ok := cm["type"].(string); ok {
					mc.Type = t
				}
				if text, ok := cm["text"].(string); ok {
					mc.Text = &text
				}
				if imgFile, ok := cm["image_file"].(map[string]interface{}); ok {
					mc.ImageFile = &MessageImageFile{}
					if fileID, ok := imgFile["file_id"].(string); ok {
						mc.ImageFile.FileID = fileID
					}
				}
				if imgURL, ok := cm["image_url"].(map[string]interface{}); ok {
					mc.ImageURL = &MessageImageURL{}
					if url, ok := imgURL["url"].(string); ok {
						mc.ImageURL.URL = url
					}
				}
				result = append(result, mc)
			}
		}
		return result
	default:
		return []MessageContent{}
	}
}
