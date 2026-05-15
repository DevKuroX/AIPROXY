package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type TestResult struct {
	URL        string
	ProxyURL   string
	LatencyMs  int
	Region     string
	Alive      bool
	Error      string
}

func TestProxy(proxyURL string) TestResult {
	result := TestResult{ProxyURL: proxyURL}
	start := time.Now()

	// Try to fetch a known URL through the proxy
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxyURL)
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Get("http://httpbin.org/ip")
	if err != nil {
		result.Alive = false
		result.Error = err.Error()
		result.LatencyMs = int(time.Since(start).Milliseconds())
		return result
	}
	defer resp.Body.Close()

	result.LatencyMs = int(time.Since(start).Milliseconds())
	body, _ := io.ReadAll(resp.Body)
	result.Alive = true

	// Detect region via ip-api.com
	regionResult := detectRegion(proxyURL)
	result.Region = regionResult

	_ = body // could parse for origin IP
	return result
}

func TestProxyQuick(proxyURL string) (alive bool, latencyMs int) {
	start := time.Now()
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxyURL)
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	resp, err := client.Get("http://httpbin.org/ip")
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}
	resp.Body.Close()
	return true, int(time.Since(start).Milliseconds())
}

func detectRegion(proxyURL string) string {
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxyURL)
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	type geoResponse struct {
		CountryCode string `json:"countryCode"`
		Country     string `json:"country"`
		Region      string `json:"region"`
		City        string `json:"city"`
	}

	resp, err := client.Get("http://ip-api.com/json/")
	if err != nil {
		return "??"
	}
	defer resp.Body.Close()

	var geo geoResponse
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return "??"
	}

	if geo.CountryCode != "" {
		return fmt.Sprintf("%s-%s", geo.CountryCode, geo.Region)
	}
	return geo.Country
}
