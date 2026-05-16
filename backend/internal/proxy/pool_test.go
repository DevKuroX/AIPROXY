package proxy

import (
	"context"
	"errors"
	"testing"
)

type mockStore struct {
	proxies  []*Proxy
	settings *ProxySettings
}

func (m *mockStore) SaveProxy(ctx context.Context, p *Proxy) error {
	for i, existing := range m.proxies {
		if existing.ID == p.ID {
			m.proxies[i] = p
			return nil
		}
	}
	m.proxies = append(m.proxies, p)
	return nil
}
func (m *mockStore) GetProxies(ctx context.Context, f map[string]interface{}) ([]*Proxy, error) { return m.proxies, nil }
func (m *mockStore) DeleteProxy(ctx context.Context, id string) error {
	for i, p := range m.proxies {
		if p.ID == id {
			m.proxies = append(m.proxies[:i], m.proxies[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}
func (m *mockStore) SavePool(ctx context.Context, p *ProxyPool) error           { return nil }
func (m *mockStore) GetPools(ctx context.Context) ([]*ProxyPool, error)       { return nil, nil }
func (m *mockStore) DeletePool(ctx context.Context, id string) error            { return nil }
func (m *mockStore) SaveSettings(ctx context.Context, s *ProxySettings) error   { m.settings = s; return nil }
func (m *mockStore) GetSettings(ctx context.Context) (*ProxySettings, error) {
	if m.settings != nil {
		return m.settings, nil
	}
	return &DefaultProxySettings, nil
}

func newTestManager() *Manager {
	return NewManager(&mockStore{})
}

func TestNewManager(t *testing.T) {
	m := newTestManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestAddAndListProxy(t *testing.T) {
	m := newTestManager()
	p := &Proxy{
		ID:     "p1",
		URL:    "http://proxy1:8080",
		Host:   "proxy1",
		Port:   "8080",
		Status: StatusUntested,
	}
	if err := m.AddProxy(p); err != nil {
		t.Fatalf("AddProxy failed: %v", err)
	}

	proxies, err := m.ListProxies(nil)
	if err != nil {
		t.Fatalf("ListProxies failed: %v", err)
	}
	if len(proxies) != 1 {
		t.Fatalf("expected 1 proxy, got %d", len(proxies))
	}
	if proxies[0].URL != "http://proxy1:8080" {
		t.Fatalf("expected http://proxy1:8080, got %s", proxies[0].URL)
	}
}

func TestDeleteProxy(t *testing.T) {
	m := newTestManager()
	m.AddProxy(&Proxy{ID: "p1", URL: "http://proxy1:8080", Host: "proxy1", Port: "8080"})

	if err := m.DeleteProxy("p1"); err != nil {
		t.Fatalf("DeleteProxy failed: %v", err)
	}

	proxies, _ := m.ListProxies(nil)
	if len(proxies) != 0 {
		t.Fatalf("expected 0 proxies after delete, got %d", len(proxies))
	}
}

func TestSelectProxyLowestLatency(t *testing.T) {
	m := newTestManager()
	m.AddProxy(&Proxy{ID: "p1", URL: "http://slow:8080", Status: StatusOK, LatencyMs: 500})
	m.AddProxy(&Proxy{ID: "p2", URL: "http://fast:8080", Status: StatusOK, LatencyMs: 100})
	m.AddProxy(&Proxy{ID: "p3", URL: "http://medium:8080", Status: StatusOK, LatencyMs: 250})

	selected := m.SelectProxy("", "")
	if selected == nil {
		t.Fatal("SelectProxy returned nil")
	}
	if selected.ID != "p2" {
		t.Fatalf("expected p2 (fastest 100ms), got %s (latency=%d)", selected.ID, selected.LatencyMs)
	}
}

func TestSelectProxySkipsDead(t *testing.T) {
	m := newTestManager()
	m.AddProxy(&Proxy{ID: "p1", URL: "http://dead:8080", Status: StatusDead})
	m.AddProxy(&Proxy{ID: "p2", URL: "http://alive:8080", Status: StatusOK, LatencyMs: 200})

	selected := m.SelectProxy("", "")
	if selected == nil {
		t.Fatal("SelectProxy returned nil")
	}
	if selected.ID != "p2" {
		t.Fatalf("expected p2 (alive), got %s", selected.ID)
	}
}

func TestSelectProxyNoAvailable(t *testing.T) {
	m := newTestManager()
	selected := m.SelectProxy("", "")
	if selected != nil {
		t.Fatal("SelectProxy should return nil when no proxies")
	}
}

func TestImportProxyList(t *testing.T) {
	m := newTestManager()
	count := m.ImportProxyList([]*Proxy{
		{ID: "p1", URL: "http://p1:8080", Status: StatusUntested},
		{ID: "p2", URL: "http://p2:8080", Status: StatusUntested},
	})
	if count != 2 {
		t.Fatalf("expected 2 imported, got %d", count)
	}

	proxies, _ := m.ListProxies(nil)
	if len(proxies) != 2 {
		t.Fatalf("expected 2 proxies, got %d", len(proxies))
	}
}

func TestLoadProxies(t *testing.T) {
	m := newTestManager()
	err := m.LoadProxies()
	t.Logf("LoadProxies returned: %v", err)
}

func TestGetSettings(t *testing.T) {
	m := newTestManager()
	s := m.GetSettings()
	if s.Enabled {
		t.Fatal("default settings should have Enabled=false")
	}
}

func TestShouldProxy(t *testing.T) {
	m := newTestManager()
	if m.ShouldProxy("kiro") {
		t.Fatal("ShouldProxy should return false by default")
	}
}

func TestProxyStatusConstants(t *testing.T) {
	tests := []struct {
		status   ProxyStatus
		expected string
	}{
		{StatusUntested, "untested"},
		{StatusOK, "ok"},
		{StatusDead, "dead"},
		{StatusSlow, "slow"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Fatalf("expected %s, got %s", tt.expected, string(tt.status))
		}
	}
}

func TestProtocolConstants(t *testing.T) {
	tests := []struct {
		protocol Protocol
		expected string
	}{
		{ProtocolHTTP, "http"},
		{ProtocolHTTPS, "https"},
		{ProtocolSOCKS4, "socks4"},
		{ProtocolSOCKS5, "socks5"},
	}
	for _, tt := range tests {
		if string(tt.protocol) != tt.expected {
			t.Fatalf("expected %s, got %s", tt.expected, string(tt.protocol))
		}
	}
}
