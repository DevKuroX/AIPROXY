package search

import (
	"fmt"
)

// searchProviderStore implements v1.SearchProviderStore with hardcoded provider configs
type searchProviderStore struct {
	providers map[string]*ProviderConfig
}

// NewSearchProviderStore creates a new search provider store with default configurations
func NewSearchProviderStore() *searchProviderStore {
	return &searchProviderStore{
		providers: map[string]*ProviderConfig{
			"brave-search": {
				ID:                "brave-search",
				BaseURL:           "https://api.search.brave.com/res/v1",
				Method:            "GET",
				AuthType:          "apikey",
				SearchTypes:       []string{"web", "news"},
				DefaultMaxResults: 5,
				MaxMaxResults:     20,
				TimeoutMs:         10000,
				CostPerQuery:      0.005,
			},
			"tavily": {
				ID:                "tavily",
				BaseURL:           "https://api.tavily.com/search",
				Method:            "POST",
				AuthType:          "apikey",
				SearchTypes:       []string{"web", "news"},
				DefaultMaxResults: 5,
				MaxMaxResults:     20,
				TimeoutMs:         10000,
				CostPerQuery:      0.008,
			},
			"serper": {
				ID:                "serper",
				BaseURL:           "https://google.serper.dev",
				Method:            "POST",
				AuthType:          "apikey",
				SearchTypes:       []string{"web", "news"},
				DefaultMaxResults: 5,
				MaxMaxResults:     50,
				TimeoutMs:         10000,
				CostPerQuery:      0.004,
			},
			"exa": {
				ID:                "exa",
				BaseURL:           "https://api.exa.ai/search",
				Method:            "POST",
				AuthType:          "apikey",
				SearchTypes:       []string{"web", "news"},
				DefaultMaxResults: 5,
				MaxMaxResults:     50,
				TimeoutMs:         15000,
				CostPerQuery:      0.005,
			},
			"searxng": {
				ID:                "searxng",
				BaseURL:           "http://localhost:8888/search",
				Method:            "GET",
				AuthType:          "none",
				SearchTypes:       []string{"web", "news"},
				DefaultMaxResults: 5,
				MaxMaxResults:     50,
				TimeoutMs:         10000,
				CostPerQuery:      0,
			},
		},
	}
}

// GetSearchProvider returns the provider configuration for the given provider ID
func (s *searchProviderStore) GetSearchProvider(providerID string) (*ProviderConfig, error) {
	cfg, ok := s.providers[providerID]
	if !ok {
		return nil, fmt.Errorf("search provider %q not found", providerID)
	}
	return cfg, nil
}
