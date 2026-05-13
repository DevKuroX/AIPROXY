// ref: _ref/9router/open-sse/utils/proxyFetch.js
package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type DNSCacheEntry struct {
	IP     string
	Expiry int64
}

var dnsCache = struct {
	sync.RWMutex
	entries map[string]DNSCacheEntry
}{
	entries: make(map[string]DNSCacheEntry),
}

var MitmBypassHosts = []string{
	"cloudcode-pa.googleapis.com",
	"daily-cloudcode-pa.googleapis.com",
	"api.individual.githubcopilot.com",
	"q.us-east-1.amazonaws.com",
	"codewhisperer.us-east-1.amazonaws.com",
	"api2.cursor.sh",
}

var GoogleDNSServers = []string{"8.8.8.8", "8.8.4.4"}

var defaultDNSTTL = 5 * time.Minute

func ShouldBypassMitmDns(targetURL string) bool {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	hostname := parsed.Hostname()
	for _, host := range MitmBypassHosts {
		if strings.Contains(hostname, host) {
			return true
		}
	}
	return false
}

func ShouldBypassByNoProxy(targetURL, noProxy string) bool {
	if noProxy == "" {
		return false
	}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	hostname := strings.ToLower(parsed.Hostname())

	patterns := strings.Split(noProxy, ",")
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(strings.ToLower(pattern))
		if pattern == "" {
			continue
		}
		if pattern == "*" {
			return true
		}
		if strings.HasPrefix(pattern, ".") {
			if strings.HasSuffix(hostname, pattern) || hostname == pattern[1:] {
				return true
			}
		}
		if hostname == pattern || strings.HasSuffix(hostname, "."+pattern) {
			return true
		}
	}

	return false
}

func GetEnvProxyURL(targetURL string) string {
	noProxy := os.Getenv("NO_PROXY")
	if noProxy == "" {
		noProxy = os.Getenv("no_proxy")
	}
	if ShouldBypassByNoProxy(targetURL, noProxy) {
		return ""
	}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}

	if parsed.Scheme == "https" {
		if proxy := os.Getenv("HTTPS_PROXY"); proxy != "" {
			return proxy
		}
		if proxy := os.Getenv("https_proxy"); proxy != "" {
			return proxy
		}
	}

	if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("http_proxy"); proxy != "" {
		return proxy
	}

	if proxy := os.Getenv("ALL_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("all_proxy"); proxy != "" {
		return proxy
	}

	return ""
}

type ProxyTransport struct {
	transport *http.Transport
}

func NewProxyTransport() *ProxyTransport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: os.Getenv("NODE_TLS_REJECT_UNAUTHORIZED") == "0",
		},
	}

	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		proxyURL := GetEnvProxyURL(req.URL.String())
		if proxyURL == "" {
			return nil, nil
		}
		return url.Parse(proxyURL)
	}

	return &ProxyTransport{transport: transport}
}

func (t *ProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.transport.RoundTrip(req)
}

func NewProxyClient() *http.Client {
	return &http.Client{
		Transport: NewProxyTransport(),
		Timeout:   10 * time.Minute,
	}
}

func NewProxyClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: NewProxyTransport(),
		Timeout:   timeout,
	}
}

func ProxyFetch(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	if client == nil {
		client = NewProxyClient()
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("proxy fetch failed: %w", err)
	}

	return resp, nil
}

func ResolveRealIP(hostname string) (string, error) {
	dnsCache.RLock()
	if entry, ok := dnsCache.entries[hostname]; ok {
		if time.Now().Unix() < entry.Expiry {
			dnsCache.RUnlock()
			return entry.IP, nil
		}
	}
	dnsCache.RUnlock()

	resolver := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addrs, err := resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return "", fmt.Errorf("DNS resolve failed for %s: %w", hostname, err)
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("no addresses found for %s", hostname)
	}

	ip := addrs[0].IP.String()

	dnsCache.Lock()
	dnsCache.entries[hostname] = DNSCacheEntry{
		IP:     ip,
		Expiry: time.Now().Add(defaultDNSTTL).Unix(),
	}
	dnsCache.Unlock()

	return ip, nil
}
