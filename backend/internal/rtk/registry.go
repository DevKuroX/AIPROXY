// ref: open-sse/rtk/registry.js
package rtk

import "sync"

// Filter is a function that transforms content.
type Filter func(content string) string

// FilterWithMeta wraps a filter with its name for error reporting.
type FilterWithMeta struct {
	Name   string
	Filter Filter
}

// FilterRegistry holds registered filters by name.
type FilterRegistry struct {
	mu      sync.RWMutex
	filters map[string]Filter
	aliasToCanonical map[string]string
}

// globalRegistry is the default filter registry.
var globalRegistry = NewFilterRegistry()

// NewFilterRegistry creates a new empty filter registry.
func NewFilterRegistry() *FilterRegistry {
	return &FilterRegistry{
		filters: make(map[string]Filter),
		aliasToCanonical: make(map[string]string),
	}
}

// RegisterFilter adds a filter to the registry.
func (r *FilterRegistry) RegisterFilter(name string, filter Filter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.filters[name] = filter
}

// RegisterAlias adds an alias for an existing filter.
func (r *FilterRegistry) RegisterAlias(alias, canonical string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aliasToCanonical[alias] = canonical
}

// GetFilter retrieves a filter by name, checking aliases.
// Returns the filter and true if found, nil and false otherwise.
func (r *FilterRegistry) GetFilter(name string) (Filter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check direct match first
	if f, ok := r.filters[name]; ok {
		return f, true
	}

	// Check aliases
	if canonical, ok := r.aliasToCanonical[name]; ok {
		if f, ok := r.filters[canonical]; ok {
			return f, true
		}
	}

	return nil, false
}

// AllFilters returns a copy of all registered filters.
func (r *FilterRegistry) AllFilters() map[string]Filter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Filter, len(r.filters))
	for k, v := range r.filters {
		result[k] = v
	}
	return result
}

// Package-level functions using the global registry

// RegisterFilter adds a filter to the global registry.
func RegisterFilter(name string, filter Filter) {
	globalRegistry.RegisterFilter(name, filter)
}

// GetFilter retrieves a filter from the global registry.
func GetFilter(name string) (Filter, bool) {
	return globalRegistry.GetFilter(name)
}

// RegisterAlias adds an alias to the global registry.
// Rust resolve_filter aliases (pipe_cmd.rs): grep|rg, find|fd
func RegisterAlias(alias, canonical string) {
	globalRegistry.RegisterAlias(alias, canonical)
}

// AllFilters returns all filters from the global registry.
func AllFilters() map[string]Filter {
	return globalRegistry.AllFilters()
}
