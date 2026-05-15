package proxy

import "time"

type Protocol string

const (
	ProtocolHTTP  Protocol = "http"
	ProtocolHTTPS Protocol = "https"
	ProtocolSOCKS4 Protocol = "socks4"
	ProtocolSOCKS5 Protocol = "socks5"
)

type ProxyStatus string

const (
	StatusUntested  ProxyStatus = "untested"
	StatusOK        ProxyStatus = "ok"
	StatusDead      ProxyStatus = "dead"
	StatusSlow      ProxyStatus = "slow"
)

type Proxy struct {
	ID          string      `json:"id"`
	URL         string      `json:"url"`
	Protocol    Protocol    `json:"protocol"`
	Host        string      `json:"host"`
	Port        string      `json:"port"`
	Region      string      `json:"region"`
	LatencyMs   int         `json:"latency_ms"`
	Status      ProxyStatus `json:"status"`
	Source      string      `json:"source"`
	SuccessRate float64     `json:"success_rate"`
	FailCount   int         `json:"fail_count"`
	LastChecked time.Time   `json:"last_checked"`
	CreatedAt   time.Time   `json:"created_at"`
}

type ProxyPool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	ProxyURL    string `json:"proxy_url"`
	NoProxy     string `json:"no_proxy"`
	StrictProxy bool   `json:"strict_proxy"`
	IsActive    bool   `json:"is_active"`
	TestStatus  string `json:"test_status"`
	LastError   string `json:"last_error,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ProxySettings struct {
	Enabled      bool   `json:"enabled"`
	ProxyForKiro bool   `json:"proxy_for_kiro"`
	ProxyForCodex bool  `json:"proxy_for_codex"`
	ProxyForOpenAI bool `json:"proxy_for_openai"`
	ProxyForGitHub bool `json:"proxy_for_github"`
	ProxyForClaude bool `json:"proxy_for_claude"`
	MaxLatencyMs int    `json:"max_latency_ms"`
	ScraperIntervalMin int `json:"scraper_interval_min"`
	WebshareAPIKey string `json:"webshare_api_key"`
}

type ScraperSource struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Format string `json:"format"`  // geonode, proxyscrape, raw (ip:port), webshare
	Type   string `json:"type"`   // http, https, socks4, socks5, all
}

type ScrapeProgress struct {
	Source    string `json:"source"`
	Found     int    `json:"found"`
	Tested    int    `json:"tested"`
	Alive     int    `json:"alive"`
	Errors    int    `json:"errors"`
	Running   bool   `json:"running"`
	Total     int    `json:"total"`
}

var DefaultScraperSources = []ScraperSource{
	{Name: "geonode", URL: "https://proxylist.geonode.com/api/proxy-list?filterLastChecked=10&limit=500&sort_by=lastChecked&sort_type=desc", Format: "geonode", Type: "all"},
	{Name: "proxyscrape", URL: "https://api.proxyscrape.com/v4/free-proxy-list/get?protocol=all&timeout=10000&country=all&ssl=all&anonymity=all&limit=200000&request=displayproxies", Format: "proxyscrape", Type: "all"},
	{Name: "proxifly", URL: "https://raw.githubusercontent.com/proxifly/free-proxy-list/refs/heads/main/proxies/all/data.txt", Format: "raw", Type: "all"},
	{Name: "proxripper", URL: "https://raw.githubusercontent.com/Mohammedcha/ProxRipper/refs/heads/main/full_proxies/https.txt", Format: "raw", Type: "https"},
	{Name: "iplocate", URL: "https://raw.githubusercontent.com/iplocate/free-proxy-list/refs/heads/main/all-proxies.txt", Format: "raw", Type: "all"},
	{Name: "ercin-socks4", URL: "https://raw.githubusercontent.com/ErcinDedeoglu/proxies/refs/heads/main/proxies/socks4.txt", Format: "raw", Type: "socks4"},
	{Name: "ercin-socks5", URL: "https://raw.githubusercontent.com/ErcinDedeoglu/proxies/refs/heads/main/proxies/socks5.txt", Format: "raw", Type: "socks5"},
	{Name: "webshare", URL: "", Format: "webshare", Type: "all"},  // API key needed
}

var DefaultProxySettings = ProxySettings{
	Enabled:          false,
	ProxyForKiro:     false,
	ProxyForCodex:    false,
	ProxyForOpenAI:   false,
	ProxyForGitHub:   false,
	ProxyForClaude:   false,
	MaxLatencyMs:     2000,
	ScraperIntervalMin: 30,
}
