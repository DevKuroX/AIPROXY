package oauth

import (
	"sync"
)

// Registry maintains a mapping of provider types to Flow implementations.
type Registry struct {
	mu    sync.RWMutex
	flows map[string]Flow
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		flows: make(map[string]Flow),
	}
}

// RegisterFlow registers a Flow implementation for a provider type.
// ref: open-sse/services/tokenRefresh.js:35
func (r *Registry) RegisterFlow(providerType string, flow Flow) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.flows[providerType] = flow
}

// GetFlow retrieves a Flow implementation for a provider type.
// Returns the flow and true if found, nil and false otherwise.
// ref: open-sse/services/tokenRefresh.js:35-36
func (r *Registry) GetFlow(providerType string) (Flow, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	flow, ok := r.flows[providerType]
	return flow, ok
}
