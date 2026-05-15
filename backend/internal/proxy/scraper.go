package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ScrapeResult struct {
	Source string
	Proxies []string
	Error  error
}

func ScrapeAllSources(webshareAPIKey string) []ScrapeResult {
	results := make([]ScrapeResult, 0)
	for _, src := range DefaultScraperSources {
		// Skip webshare if no API key
		if src.Format == "webshare" && webshareAPIKey == "" {
			continue
		}
		result := scrapeSource(src, webshareAPIKey)
		results = append(results, result)
	}
	return results
}

func scrapeSource(src ScraperSource, webshareAPIKey string) ScrapeResult {
	result := ScrapeResult{Source: src.Name}

	switch src.Format {
	case "geonode":
		result.Proxies, result.Error = scrapeGeonode(src.URL)
	case "proxyscrape":
		result.Proxies, result.Error = scrapeTextURL(src.URL)
	case "raw":
		result.Proxies, result.Error = scrapeTextURL(src.URL)
	case "webshare":
		result.Proxies, result.Error = scrapeWebshare(webshareAPIKey)
	}

	return result
}

func scrapeGeonode(url string) ([]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			IP    string `json:"ip"`
			Port  string `json:"port"`
			Protocols []string `json:"protocols"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var proxies []string
	for _, p := range data.Data {
		for _, proto := range p.Protocols {
			proxies = append(proxies, fmt.Sprintf("%s://%s:%s", proto, p.IP, p.Port))
		}
	}
	return proxies, nil
}

func scrapeTextURL(url string) ([]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var proxies []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		// Handle ip:port format
		if strings.Contains(line, ":") && !strings.Contains(line, "://") {
			proxies = append(proxies, "http://"+line)
		} else if strings.Contains(line, "://") {
			proxies = append(proxies, line)
		}
	}
	return proxies, scanner.Err()
}

func scrapeWebshare(apiKey string) ([]string, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("webshare API key required")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	var allProxies []string

	for page := 1; page <= 10; page++ {
		url := fmt.Sprintf("https://proxy.webshare.io/api/v2/proxy/list/?page=%d&page_size=100", page)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Token "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			break
		}

		var data struct {
			Results []struct {
				ProxyAddress string `json:"proxy_address"`
				Port         int    `json:"port"`
				Username     string `json:"username"`
				Password     string `json:"password"`
				CountryCode  string `json:"country_code"`
				Protocol     string `json:"protocol"`
			} `json:"results"`
			Next string `json:"next"`
		}
		json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()

		for _, p := range data.Results {
			proto := strings.ToLower(p.Protocol)
			if proto == "" {
				proto = "http"
			}
			proxyURL := fmt.Sprintf("%s://%s:%s@%s:%d", proto, p.Username, p.Password, p.ProxyAddress, p.Port)
			allProxies = append(allProxies, proxyURL)
		}

		if data.Next == "" || data.Next == "null" {
			break
		}
	}

	return allProxies, nil
}
