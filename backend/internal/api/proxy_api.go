package api

import (
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/proxy"
)

type ProxyAPI struct {
	manager *proxy.Manager
}

func NewProxyAPI(m *proxy.Manager) *ProxyAPI {
	return &ProxyAPI{manager: m}
}

// GET /api/proxies - list proxies with optional filter
func (h *ProxyAPI) ListProxies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	proxies, err := h.manager.ListProxies(nil)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"proxies": proxies})
}

// POST /api/proxies - add proxy manually
func (h *ProxyAPI) AddProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var p proxy.Proxy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}
	if err := h.manager.AddProxy(&p); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DELETE /api/proxies/:id - delete proxy
func (h *ProxyAPI) DeleteProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	if err := h.manager.DeleteProxy(id); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GET /api/proxy-pools - list proxy pools
func (h *ProxyAPI) ListPools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pools, err := h.manager.ListPools()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"proxy_pools": pools})
}

// POST /api/proxy-pools - create proxy pool
func (h *ProxyAPI) CreatePool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var p proxy.ProxyPool
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}
	if err := h.manager.CreatePool(&p); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DELETE /api/proxy-pools/:id - delete pool
func (h *ProxyAPI) DeletePool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	if err := h.manager.DeletePool(id); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// POST /api/proxy-pools/test-all - test all pools concurrently
func (h *ProxyAPI) TestAllPools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	result := h.manager.TestAllPools()
	json.NewEncoder(w).Encode(result)
}

// POST /api/proxy-pools/:id/test - test pool connection
func (h *ProxyAPI) TestPool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	result, err := h.manager.TestPool(id)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(result)
}

// GET /api/proxy/settings - get proxy settings
func (h *ProxyAPI) GetSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.manager.GetSettings())
}

// POST /api/proxy/settings - update proxy settings
func (h *ProxyAPI) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var s proxy.ProxySettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}
	if err := h.manager.UpdateSettings(s); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// POST /api/scraper/start - run proxy scraper
func (h *ProxyAPI) StartScraper(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	go h.manager.RunScraper()
	json.NewEncoder(w).Encode(map[string]string{"status": "scraper started"})
}

// POST /api/scraper/webshare - scrape webshare only
func (h *ProxyAPI) ScrapeWebshare(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	count := h.manager.ScrapeWebshare()
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "imported": count})
}

// GET /api/scraper/progress - get scraper progress
func (h *ProxyAPI) ScraperProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.manager.GetScraperProgress())
}
