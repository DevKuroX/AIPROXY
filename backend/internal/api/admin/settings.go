package admin

import (
	"context"
	"encoding/json"
	"net/http"
)

type SettingsStore interface {
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error
}

type SettingsHandler struct {
	store SettingsStore
}

func NewSettingsHandler(store SettingsStore) *SettingsHandler {
	return &SettingsHandler{store: store}
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings := make(map[string]interface{})

	keys := []string{
		"requireApiKey",
		"requireLogin",
		"tunnelEnabled",
		"tunnelUrl",
		"tunnelDashboardAccess",
		"rtkEnabled",
		"cavemanEnabled",
		"cavemanLevel",
		"outboundProxyEnabled",
		"outboundProxyUrl",
		"outboundNoProxy",
	}

	for _, key := range keys {
		value, err := h.store.GetSetting(r.Context(), key)
		if err == nil {
			if value == "true" {
				settings[key] = true
			} else if value == "false" {
				settings[key] = false
			} else {
				settings[key] = value
			}
		} else {
			switch key {
			case "requireLogin":
				settings[key] = true
			case "requireApiKey", "tunnelEnabled", "tunnelDashboardAccess", "cavemanEnabled", "outboundProxyEnabled":
				settings[key] = false
			case "rtkEnabled":
				settings[key] = true
			case "cavemanLevel":
				settings[key] = "full"
			default:
				settings[key] = ""
			}
		}
	}

	settings["hasPassword"] = true

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for key, value := range updates {
		var strValue string
		switch v := value.(type) {
		case bool:
			if v {
				strValue = "true"
			} else {
				strValue = "false"
			}
		case string:
			strValue = v
		case float64:
			strValue = string(rune(int(v)))
		default:
			continue
		}

		if err := h.store.SetSetting(r.Context(), key, strValue); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.Get(w, r)
}
