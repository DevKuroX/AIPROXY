package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/providers"
)

type ModelEntry struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	OwnedBy  string `json:"owned_by"`
	Created  int64  `json:"created"`
}

type ModelsResponse struct {
	Object string       `json:"object"`
	Data   []ModelEntry `json:"data"`
}

func HandleListModels(w http.ResponseWriter, r *http.Request) {
	entries := []ModelEntry{}
	now := time.Now().Unix()

	for id, cfg := range providers.PROVIDERS {
		if cfg.Format != providers.FormatOpenAI &&
		   cfg.Format != providers.FormatClaude &&
		   cfg.Format != providers.FormatGeminiCLI &&
		   cfg.Format != providers.FormatKiro &&
		   cfg.Format != providers.FormatGemini &&
		   cfg.Format != providers.FormatAntigravity &&
		   cfg.Format != providers.FormatCursor &&
	       cfg.Format != providers.FormatOllama &&
	       cfg.Format != providers.FormatGeminiWeb {
			continue
		}

		entries = append(entries, ModelEntry{
			ID:      id,
			Object:  "model",
			OwnedBy: cfg.Type,
			Created: now,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModelsResponse{
		Object: "list",
		Data:   entries,
	})
}

func HandleListImageModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModelsResponse{Object: "list", Data: []ModelEntry{}})
}

func HandleListTTSModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModelsResponse{Object: "list", Data: []ModelEntry{}})
}

func HandleListEmbeddingModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModelsResponse{Object: "list", Data: []ModelEntry{}})
}

func HandleListWebModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModelsResponse{Object: "list", Data: []ModelEntry{}})
}

func HandleListSTTModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ModelsResponse{Object: "list", Data: []ModelEntry{}})
}
