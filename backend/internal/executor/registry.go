// ref: _ref/9router/open-sse/executors/index.js (registry pattern)
package executor

import (
	"fmt"
	"sync"
)

var (
	registry   = make(map[string]Executor)
	registryMu sync.RWMutex
)

// Register adds an executor for a provider type to the global registry.
// Panics if an executor is already registered for the provider type.
func Register(providerType string, exec Executor) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[providerType]; exists {
		panic(fmt.Sprintf("executor already registered for provider type: %s", providerType))
	}
	registry[providerType] = exec
}

// Get retrieves an executor for a provider type from the registry.
// Returns the executor and true if found, nil and false otherwise.
func Get(providerType string) (Executor, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	exec, ok := registry[providerType]
	return exec, ok
}

// MustGet retrieves an executor for a provider type, panicking if not found.
func MustGet(providerType string) Executor {
	exec, ok := Get(providerType)
	if !ok {
		panic(fmt.Sprintf("no executor registered for provider type: %s", providerType))
	}
	return exec
}

// ListProviders returns all registered provider types.
func ListProviders() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	providers := make([]string, 0, len(registry))
	for p := range registry {
		providers = append(providers, p)
	}
	return providers
}

func init() {
	Register("codex", NewCodexExecutor(""))
	Register("github", NewGitHubExecutor())
	Register("opencode", NewOpenCodeExecutor())
	Register("opencode-go", NewOpenCodeGoExecutor())
	Register("ollama-local", NewOllamaLocalExecutor())
    Register("antigravity", NewAntigravityExecutor(nil))
    Register("azure", NewAzureExecutor())
    Register("commandcode", NewCommandCodeExecutor())
    Register("cursor", NewCursorExecutor())
    Register("gemini-cli", NewGeminiCLIExecutor(nil))
    Register("grok-web", NewGrokWebExecutor())
    Register("iflow", NewIFlowExecutor())
    Register("kiro", NewKiroExecutor())
    Register("perplexity-web", NewPerplexityWebExecutor())
    Register("qoder", NewQoderExecutor())
    Register("qwen", NewQwenExecutor())
}
