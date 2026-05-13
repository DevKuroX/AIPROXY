package services

import (
	"testing"
	"time"
)

func TestNewUsageService(t *testing.T) {
	logger := &mockLogger{}
	svc := NewUsageService(nil, logger)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}

	if svc.logger == nil {
		t.Error("expected logger to be set")
	}

	if svc.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestUsageFilters(t *testing.T) {
	now := time.Now()
	filters := UsageFilters{
		StartDate:  &now,
		EndDate:    &now,
		ProviderID: "openai",
		Model:      "gpt-4",
		APIKeyID:   "key-123",
	}

	if filters.ProviderID != "openai" {
		t.Errorf("ProviderID = %q, want openai", filters.ProviderID)
	}

	if filters.Model != "gpt-4" {
		t.Errorf("Model = %q, want gpt-4", filters.Model)
	}

	if filters.APIKeyID != "key-123" {
		t.Errorf("APIKeyID = %q, want key-123", filters.APIKeyID)
	}
}

func TestProviderQuota(t *testing.T) {
	quota := ProviderQuota{
		Plan:      "pro",
		ResetDate: "2024-12-31",
		Quotas: map[string]Quota{
			"chat": {
				Used:      100,
				Total:     500,
				Unlimited: false,
			},
		},
		Message: "Quota reset on reset date",
	}

	if quota.Plan != "pro" {
		t.Errorf("Plan = %q, want pro", quota.Plan)
	}

	if quota.ResetDate != "2024-12-31" {
		t.Errorf("ResetDate = %q, want 2024-12-31", quota.ResetDate)
	}

	if chatQuota, ok := quota.Quotas["chat"]; ok {
		if chatQuota.Used != 100 {
			t.Errorf("chat.Used = %d, want 100", chatQuota.Used)
		}
		if chatQuota.Total != 500 {
			t.Errorf("chat.Total = %d, want 500", chatQuota.Total)
		}
	} else {
		t.Error("missing chat quota")
	}
}

func TestQuota(t *testing.T) {
	quota := Quota{
		Used:               75,
		Total:              100,
		Remaining:          25,
		RemainingPercentage: 25.0,
		ResetAt:            "2024-01-01T00:00:00Z",
		Unlimited:          false,
		DisplayName:        "Chat Messages",
	}

	if quota.Used != 75 {
		t.Errorf("Used = %d, want 75", quota.Used)
	}

	if quota.Total != 100 {
		t.Errorf("Total = %d, want 100", quota.Total)
	}

	if quota.Remaining != 25 {
		t.Errorf("Remaining = %d, want 25", quota.Remaining)
	}

	if quota.RemainingPercentage != 25.0 {
		t.Errorf("RemainingPercentage = %f, want 25.0", quota.RemainingPercentage)
	}

	if quota.DisplayName != "Chat Messages" {
		t.Errorf("DisplayName = %q, want Chat Messages", quota.DisplayName)
	}
}

func TestUsageSummary(t *testing.T) {
	summary := UsageSummary{
		TotalRequests:      100,
		SuccessfulRequests: 95,
		FailedRequests:     5,
		TotalPromptTokens:  10000,
		TotalCompletionTok: 5000,
		TotalTokens:        15000,
		TotalCostUSD:       1.50,
		AvgDurationMs:      250.5,
	}

	if summary.TotalRequests != 100 {
		t.Errorf("TotalRequests = %d, want 100", summary.TotalRequests)
	}

	if summary.SuccessfulRequests != 95 {
		t.Errorf("SuccessfulRequests = %d, want 95", summary.SuccessfulRequests)
	}

	if summary.FailedRequests != 5 {
		t.Errorf("FailedRequests = %d, want 5", summary.FailedRequests)
	}

	if summary.TotalPromptTokens != 10000 {
		t.Errorf("TotalPromptTokens = %d, want 10000", summary.TotalPromptTokens)
	}

	if summary.TotalCompletionTok != 5000 {
		t.Errorf("TotalCompletionTok = %d, want 5000", summary.TotalCompletionTok)
	}

	if summary.TotalCostUSD != 1.50 {
		t.Errorf("TotalCostUSD = %f, want 1.50", summary.TotalCostUSD)
	}
}

func TestUsageByProvider(t *testing.T) {
	usage := UsageByProvider{
		ProviderID:         "anthr0pic",
		TotalRequests:      50,
		TotalPromptTokens:  5000,
		TotalCompletionTok: 2500,
		TotalTokens:        7500,
		TotalCostUSD:       0.75,
	}

	if usage.ProviderID != "anthr0pic" {
		t.Errorf("ProviderID = %q, want anthr0pic", usage.ProviderID)
	}

	if usage.TotalRequests != 50 {
		t.Errorf("TotalRequests = %d, want 50", usage.TotalRequests)
	}
}

func TestUsageByModel(t *testing.T) {
	usage := UsageByModel{
		Model:              "gpt-4o",
		TotalRequests:      30,
		TotalPromptTokens:  3000,
		TotalCompletionTok: 1500,
		TotalTokens:        4500,
		TotalCostUSD:       0.45,
	}

	if usage.Model != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", usage.Model)
	}

	if usage.TotalRequests != 30 {
		t.Errorf("TotalRequests = %d, want 30", usage.TotalRequests)
	}
}

func TestUsageByAPIKey(t *testing.T) {
	usage := UsageByAPIKey{
		APIKeyID:           "key-abc123",
		TotalRequests:      200,
		TotalPromptTokens:  20000,
		TotalCompletionTok: 10000,
		TotalTokens:        30000,
		TotalCostUSD:       3.00,
	}

	if usage.APIKeyID != "key-abc123" {
		t.Errorf("APIKeyID = %q, want key-abc123", usage.APIKeyID)
	}

	if usage.TotalRequests != 200 {
		t.Errorf("TotalRequests = %d, want 200", usage.TotalRequests)
	}
}
