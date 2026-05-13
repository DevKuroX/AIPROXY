// Package filters provides content transformation filters for RTK.
// ref: open-sse/rtk/filters/
package filters

import (
	"sync"
)

// Filter transforms content and returns the modified version.
type Filter interface {
	// Apply transforms the input content and returns the result.
	Apply(content string) (string, error)
	// Name returns the filter's registered name.
	Name() string
}

// registry holds all registered filters.
var (
	registry   = make(map[string]Filter)
	registryMu sync.RWMutex
)

// Register adds a filter to the global registry.
func Register(f Filter) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[f.Name()] = f
}

// Get retrieves a filter by name. Returns nil if not found.
func Get(name string) Filter {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}

// All returns all registered filters.
func All() map[string]Filter {
	registryMu.RLock()
	defer registryMu.RUnlock()
	result := make(map[string]Filter, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}
