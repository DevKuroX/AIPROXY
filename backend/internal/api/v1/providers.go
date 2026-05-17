package v1

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/DevKuroX/AIPROXY/internal/providers"
)

type ProviderResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	AuthType string `json:"auth_type"`
	BaseURL  string `json:"base_url,omitempty"`
}

type ProvidersResponse struct {
	Providers []ProviderResponse `json:"providers"`
}

// HandleListProviders returns all known providers with their config
func HandleListProviders(w http.ResponseWriter, r *http.Request) {
	all := providers.GetAllProviderConfigs()
	resp := make([]ProviderResponse, 0, len(all))

	// Sort by name
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, id := range keys {
		cfg := all[id]
		resp = append(resp, ProviderResponse{
			ID:       id,
			Name:     cfg.Name,
			Type:     cfg.Type,
			AuthType: cfg.AuthType,
			BaseURL:  cfg.BaseURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProvidersResponse{Providers: resp})
}
