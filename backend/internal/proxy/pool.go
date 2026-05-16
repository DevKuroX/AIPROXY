package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Manager struct {
	mu       sync.RWMutex
	proxies  map[string][]*Proxy
	settings ProxySettings
	store    Store
}

type Store interface {
	SaveProxy(ctx context.Context, p *Proxy) error
	GetProxies(ctx context.Context, filter map[string]interface{}) ([]*Proxy, error)
	DeleteProxy(ctx context.Context, id string) error
	SavePool(ctx context.Context, p *ProxyPool) error
	GetPools(ctx context.Context) ([]*ProxyPool, error)
	DeletePool(ctx context.Context, id string) error
	SaveSettings(ctx context.Context, s *ProxySettings) error
	GetSettings(ctx context.Context) (*ProxySettings, error)
}

func NewManager(store Store) *Manager {
	m := &Manager{
		proxies:  make(map[string][]*Proxy),
		settings: DefaultProxySettings,
		store:    store,
	}

	// Load settings from store if available
	if s, err := store.GetSettings(context.Background()); err == nil && s != nil {
		m.settings = *s
	}

	return m
}

func (m *Manager) ShouldProxy(provider string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.settings.Enabled {
		return false
	}

	switch provider {
	case "kiro":
		return m.settings.ProxyForKiro
	case "codex":
		return m.settings.ProxyForCodex
	case "openai":
		return m.settings.ProxyForOpenAI
	case "github":
		return m.settings.ProxyForGitHub
	case "claude":
		return m.settings.ProxyForClaude
	default:
		return true
	}
}

func (m *Manager) SelectProxy(provider, preferredRegion string) *Proxy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allProxies := m.proxies["all"]
	if len(allProxies) == 0 {
		return nil
	}

	var best *Proxy
	for _, p := range allProxies {
		if p.Status != StatusOK {
			continue
		}
		if p.LatencyMs > m.settings.MaxLatencyMs {
			continue
		}
		if preferredRegion != "" && p.Region != preferredRegion {
			continue
		}
		if best == nil || p.LatencyMs < best.LatencyMs {
			best = p
		}
	}

	return best
}

func (m *Manager) ImportFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var proxies []*Proxy
	if err := json.Unmarshal(data, &proxies); err != nil {
		return fmt.Errorf("invalid proxy file format: %w", err)
	}

	for _, p := range proxies {
		p.Source = "import"
		m.store.SaveProxy(context.Background(), p)
		m.addToCache(p)
	}

	return nil
}

func (m *Manager) addToCache(p *Proxy) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := "all"
	if p.Region != "" {
		key = p.Region
	}
	m.proxies[key] = append(m.proxies[key], p)
}

func (m *Manager) GetSettings() ProxySettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings
}

func (m *Manager) UpdateSettings(s ProxySettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = s
	return m.store.SaveSettings(context.Background(), &s)
}

func (m *Manager) ImportProxyList(proxies []*Proxy) int {
	count := 0
	for _, p := range proxies {
		if err := m.store.SaveProxy(context.Background(), p); err == nil {
			m.addToCache(p)
			count++
		}
	}
	return count
}

func (m *Manager) ListProxies(filter map[string]interface{}) ([]*Proxy, error) {
	proxies, err := m.store.GetProxies(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	return proxies, nil
}
func (m *Manager) AddProxy(p *Proxy) error {
	if err := m.store.SaveProxy(context.Background(), p); err != nil {
		return err
	}
	m.addToCache(p)
	return nil
}
func (m *Manager) DeleteProxy(id string) error {
	return m.store.DeleteProxy(context.Background(), id)
}
func (m *Manager) ListPools() ([]*ProxyPool, error) {
	return m.store.GetPools(context.Background(), )
}
func (m *Manager) CreatePool(p *ProxyPool) error {
	return m.store.SavePool(context.Background(), p)
}
func (m *Manager) DeletePool(id string) error {
	return m.store.DeletePool(context.Background(), id)
}
func (m *Manager) TestPool(id string) (map[string]interface{}, error) {
	pools, err := m.store.GetPools(context.Background(), )
	if err != nil {
		return nil, err
	}
	for _, pool := range pools {
		if pool.ID == id {
			testResult := TestProxy(pool.ProxyURL)
			return map[string]interface{}{
				"alive":      testResult.Alive,
				"latency_ms": testResult.LatencyMs,
				"region":     testResult.Region,
			}, nil
		}
	}
	return nil, fmt.Errorf("pool not found")
}
func (m *Manager) RunScraper() {
	results := ScrapeAllSources(m.settings.WebshareAPIKey)
	for _, r := range results {
		for _, proxyURL := range r.Proxies {
			alive, latencyMs := TestProxyQuick(proxyURL)
			p := &Proxy{
				URL:       proxyURL,
				Status:    StatusOK,
				LatencyMs: latencyMs,
				Source:    r.Source,
			}
			if !alive {
				p.Status = StatusDead
			}
			m.store.SaveProxy(context.Background(), p)
		}
	}
}
func (m *Manager) GetScraperProgress() map[string]interface{} {
	return map[string]interface{}{
		"running": false,
		"sources": len(DefaultScraperSources),
	}
}
func (m *Manager) LoadProxies() error {
	all, err := m.store.GetProxies(context.Background(), nil)
	if err != nil {
		return err
	}
	for _, p := range all {
		m.addToCache(p)
	}
	return nil
}
