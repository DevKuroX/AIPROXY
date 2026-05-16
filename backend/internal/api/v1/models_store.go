package v1

import (
	"time"

	"github.com/DevKuroX/AIPROXY/internal/providers"
)

type defaultModelsStore struct{}

func NewModelsStore() ModelsStore {
	return &defaultModelsStore{}
}

func (s *defaultModelsStore) ListModels() ([]Model, error) {
	now := time.Now().Unix()
	var models []Model

	for id, cfg := range providers.PROVIDERS {
		// Only include LLM-capable formats
		switch cfg.Format {
		case providers.FormatOpenAI,
			providers.FormatClaude,
			providers.FormatGeminiCLI,
			providers.FormatKiro,
			providers.FormatGemini,
			providers.FormatAntigravity,
			providers.FormatCursor,
			providers.FormatOllama,
			providers.FormatGeminiWeb:
			models = append(models, Model{
				ID:      id,
				Object:  "model",
				Created: now,
				OwnedBy: cfg.Type,
			})
		}
	}
	return models, nil
}

func (s *defaultModelsStore) ListProviders() ([]ProviderInfo, error) {
	var infos []ProviderInfo
	for id, cfg := range providers.PROVIDERS {
		infos = append(infos, ProviderInfo{
			Name:    id,
			Type:    cfg.Type,
			BaseURL: cfg.BaseURL,
			Enabled: true,
		})
	}
	return infos, nil
}

func (s *defaultModelsStore) ListCombos() ([]ComboInfo, error) {
	return []ComboInfo{}, nil
}

func (s *defaultModelsStore) ListModelAliases() ([]AliasInfo, error) {
	return []AliasInfo{}, nil
}

func (s *defaultModelsStore) GetDisabledModels() (map[string]bool, error) {
	return map[string]bool{}, nil
}
