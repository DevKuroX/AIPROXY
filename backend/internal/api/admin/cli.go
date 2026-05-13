package admin

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/errs"
)

type CLIHandler struct{}

func NewCLIHandler() *CLIHandler {
	return &CLIHandler{}
}

type ClaudeSettings struct {
	APIProvider string `json:"apiProvider"`
	BaseURL     string `json:"baseURL"`
	APIKey      string `json:"apiKey"`
}

type CodeAssistantConfigRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

type CodexConfigRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model,omitempty"`
}

type CursorConfigRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

func (h *CLIHandler) CodeAssistantConfig(w http.ResponseWriter, r *http.Request) {
	var req CodeAssistantConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.BaseURL == "" || req.APIKey == "" {
		errs.WriteJSONError(w, "base_url and api_key are required", http.StatusBadRequest)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		errs.WriteJSONError(w, "failed to get home directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	if err := backupConfig(settingsPath); err != nil {
		errs.WriteJSONError(w, "failed to backup existing config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	settings := ClaudeSettings{
		APIProvider: "openai-compatible",
		BaseURL:     req.BaseURL,
		APIKey:      req.APIKey,
	}

	content, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		errs.WriteJSONError(w, "failed to marshal settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		errs.WriteJSONError(w, "failed to create config directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(settingsPath, content, 0600); err != nil {
		errs.WriteJSONError(w, "failed to write settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "configuration written successfully",
		"config_path":   settingsPath,
		"config_content": string(content),
	})
}

func (h *CLIHandler) CodexConfig(w http.ResponseWriter, r *http.Request) {
	var req CodexConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.BaseURL == "" || req.APIKey == "" {
		errs.WriteJSONError(w, "base_url and api_key are required", http.StatusBadRequest)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		errs.WriteJSONError(w, "failed to get home directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	configPath := filepath.Join(homeDir, ".codex", "config.json")

	if err := backupConfig(configPath); err != nil {
		errs.WriteJSONError(w, "failed to backup existing config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	config := map[string]interface{}{
		"base_url": req.BaseURL,
		"api_key":  req.APIKey,
	}
	if req.Model != "" {
		config["model"] = req.Model
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		errs.WriteJSONError(w, "failed to marshal config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		errs.WriteJSONError(w, "failed to create config directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(configPath, content, 0600); err != nil {
		errs.WriteJSONError(w, "failed to write config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "configuration written successfully",
		"config_path":    configPath,
		"config_content": string(content),
	})
}

func (h *CLIHandler) CursorConfig(w http.ResponseWriter, r *http.Request) {
	var req CursorConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.BaseURL == "" || req.APIKey == "" {
		errs.WriteJSONError(w, "base_url and api_key are required", http.StatusBadRequest)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		errs.WriteJSONError(w, "failed to get home directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	configPath := filepath.Join(homeDir, ".cursor", "mcp.json")

	if err := backupConfig(configPath); err != nil {
		errs.WriteJSONError(w, "failed to backup existing config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	config := map[string]interface{}{
		"openai": map[string]interface{}{
			"base_url": req.BaseURL,
			"api_key":  req.APIKey,
		},
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		errs.WriteJSONError(w, "failed to marshal config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		errs.WriteJSONError(w, "failed to create config directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(configPath, content, 0600); err != nil {
		errs.WriteJSONError(w, "failed to write config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "configuration written successfully",
		"config_path":    configPath,
		"config_content": string(content),
	})
}

func backupConfig(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	backupPath := fmt.Sprintf("%s.backup.%s", path, time.Now().Format("20060102-150405"))
	return os.WriteFile(backupPath, content, 0600)
}

var _ = fs.FileMode(0755)
