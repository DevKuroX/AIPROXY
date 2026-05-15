package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/pool"
)

type UsageHandler struct {
	pool *pool.Pool
}

func NewUsageHandler(p *pool.Pool) *UsageHandler {
	return &UsageHandler{pool: p}
}

func (h *UsageHandler) GetKiroUsage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	account, err := h.pool.GetAccount("kiro")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "no kiro account available",
			"message": "Unable to fetch Kiro usage right now.",
		})
		return
	}

	// Try multiple endpoints like 9router
	type attempt struct {
		name string
		url  string
		headers map[string]string
	}

	attempts := []attempt{
		{
			name: "codewhisperer-get",
			url:  "https://codewhisperer.us-east-1.amazonaws.com/getUsageLimits?isEmailRequired=true&origin=AI_EDITOR&resourceType=AGENTIC_REQUEST",
			headers: map[string]string{
				"Authorization":   "Bearer " + account.AccessToken,
				"Accept":          "application/json",
				"x-amz-user-agent": "aws-sdk-js/1.0.0 KiroIDE",
			},
		},
		{
			name: "codewhisperer-post",
			url:  "https://codewhisperer.us-east-1.amazonaws.com",
			headers: map[string]string{
				"Authorization":  "Bearer " + account.AccessToken,
				"Content-Type":   "application/x-amz-json-1.0",
				"x-amz-target":   "AmazonCodeWhispererService.GetUsageLimits",
				"Accept":         "application/json",
			},
		},
	}

	for _, a := range attempts {
		req, _ := http.NewRequest("POST", a.url, nil)
		if a.name == "codewhisperer-get" {
			req.Method = "GET"
		}
		for k, v := range a.headers {
			req.Header.Set(k, v)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			continue
		}

		json.NewEncoder(w).Encode(parseKiroQuota(data))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "usage_unavailable",
		"message": "Unable to fetch Kiro usage right now.",
		"quotas":  map[string]interface{}{},
	})
}

func parseKiroQuota(data map[string]interface{}) map[string]interface{} {
	usageList, _ := data["usageBreakdownList"].([]interface{})
	quotas := make(map[string]interface{})

	for _, item := range usageList {
		if b, ok := item.(map[string]interface{}); ok {
			resourceType, _ := b["resourceType"].(string)
			used, _ := b["currentUsageWithPrecision"].(float64)
			total, _ := b["usageLimitWithPrecision"].(float64)

			quotas[resourceType] = map[string]interface{}{
				"used":      used,
				"total":     total,
				"remaining": total - used,
			}
		}
	}

	return map[string]interface{}{
		"provider": "kiro",
		"quotas":   quotas,
		"resetAt":  data["nextDateReset"],
	}
}
