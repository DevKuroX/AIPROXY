// ref: _ref/9router/open-sse/executors/index.js (registry pattern)
package executor

import (
	"fmt"
	"log"
	"sync"
)

var (
	registry   = make(map[string]Executor)
	registryMu sync.RWMutex
)

// Register adds an executor for a provider type to the global registry.
// Returns an error if an executor is already registered for the provider type.
func Register(providerType string, exec Executor) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[providerType]; exists {
		return fmt.Errorf("executor already registered for provider type: %s", providerType)
	}
	registry[providerType] = exec
	return nil
}

// Get retrieves an executor for a provider type from the registry.
// Returns the executor and true if found, nil and false otherwise.
func Get(providerType string) (Executor, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	exec, ok := registry[providerType]
	return exec, ok
}

// MustGet retrieves an executor for a provider type, returning an error if not found.
func MustGet(providerType string) (Executor, error) {
	exec, ok := Get(providerType)
	if !ok {
		return nil, fmt.Errorf("no executor registered for provider type: %s", providerType)
	}
	return exec, nil
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
	mustRegister := func(providerType string, exec Executor) {
		if err := Register(providerType, exec); err != nil {
			log.Fatalf("failed to register executor %s: %v", providerType, err)
		}
	}
	mustRegister("codex", NewCodexExecutor(""))
	mustRegister("github", NewGitHubExecutor())
	mustRegister("opencode", NewOpenCodeExecutor())
	mustRegister("opencode-go", NewOpenCodeGoExecutor())
	mustRegister("ollama-local", NewOllamaLocalExecutor())
	mustRegister("antigravity", NewAntigravityExecutor(nil))
	mustRegister("azure", NewAzureExecutor())
	mustRegister("commandcode", NewCommandCodeExecutor())
	mustRegister("cursor", NewCursorExecutor())
	mustRegister("gemini-cli", NewGeminiCLIExecutor(nil))
	mustRegister("grok-web", NewGrokWebExecutor())
	mustRegister("iflow", NewIFlowExecutor())
	mustRegister("kiro", NewKiroExecutor())
	mustRegister("perplexity-web", NewPerplexityWebExecutor())
	mustRegister("qoder", NewQoderExecutor())
	mustRegister("qwen", NewQwenExecutor())
}
