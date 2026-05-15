package router

import (
	"encoding/json"
	"fmt"
)

type ProviderConfig struct {
	Name    string
	BaseURL string
	APIKey  string
}

type ProviderStore interface {
	GetProvider(model string) (*ProviderConfig, error)
	GetProviders(model string) ([]*ProviderConfig, error)
}

var providerStore ProviderStore

func SetProviderStore(store ProviderStore) {
	providerStore = store
}

func GetProviderStore() ProviderStore {
	return providerStore
}

func extractErrorText(body []byte) string {
	var errResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		return errResp.Error.Message
	}
	return string(body)
}

type UpstreamError struct {
	StatusCode int
	Message    string
}

func NewUpstreamError(statusCode int, message string) *UpstreamError {
	return &UpstreamError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("upstream error %d: %s", e.StatusCode, e.Message)
}
