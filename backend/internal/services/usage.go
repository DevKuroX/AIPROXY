package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

// UsageService handles usage tracking, cost calculation, and provider quota fetching.
// ref: 9router/open-sse/services/usage.js
type UsageService struct {
	storage    *storage.DB
	httpClient *http.Client
	logger     Logger
}

// NewUsageService creates a new UsageService instance.
func NewUsageService(db *storage.DB, logger Logger) *UsageService {
	return &UsageService{
		storage: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// UsageFilters holds filter parameters for usage queries
// ref: 9router/open-sse/services/usage.js (filter patterns in getUsageForProvider)
type UsageFilters struct {
	StartDate  *time.Time
	EndDate    *time.Time
	ProviderID string
	Model      string
	APIKeyID   string
}

// ProviderQuota represents quota information for a provider
// ref: 9router/open-sse/services/usage.js:162-194 (GitHub quota format)
type ProviderQuota struct {
	Plan      string            `json:"plan,omitempty"`
	ResetDate string            `json:"resetDate,omitempty"`
	Quotas    map[string]Quota  `json:"quotas,omitempty"`
	Message   string            `json:"message,omitempty"`
}

// Quota represents a single quota bucket
type Quota struct {
	Used               int     `json:"used"`
	Total              int     `json:"total"`
	Remaining          int     `json:"remaining,omitempty"`
	RemainingPercentage float64 `json:"remainingPercentage,omitempty"`
	ResetAt            string  `json:"resetAt,omitempty"`
	Unlimited          bool    `json:"unlimited"`
	DisplayName        string  `json:"displayName,omitempty"`
}

// UsageSummary aggregates usage statistics
type UsageSummary struct {
	TotalRequests      int     `json:"total_requests"`
	SuccessfulRequests int     `json:"successful_requests"`
	FailedRequests     int     `json:"failed_requests"`
	TotalPromptTokens  int64   `json:"total_prompt_tokens"`
	TotalCompletionTok int64   `json:"total_completion_tokens"`
	TotalTokens        int64   `json:"total_tokens"`
	TotalCostUSD       float64 `json:"total_cost_usd"`
	AvgDurationMs      float64 `json:"avg_duration_ms"`
}

// UsageByProvider groups usage by provider
type UsageByProvider struct {
	ProviderID         string  `json:"provider_id"`
	TotalRequests      int     `json:"total_requests"`
	TotalPromptTokens  int64   `json:"total_prompt_tokens"`
	TotalCompletionTok int64   `json:"total_completion_tokens"`
	TotalTokens        int64   `json:"total_tokens"`
	TotalCostUSD       float64 `json:"total_cost_usd"`
}

// UsageByModel groups usage by model
type UsageByModel struct {
	Model              string  `json:"model"`
	TotalRequests      int     `json:"total_requests"`
	TotalPromptTokens  int64   `json:"total_prompt_tokens"`
	TotalCompletionTok int64   `json:"total_completion_tokens"`
	TotalTokens        int64   `json:"total_tokens"`
	TotalCostUSD       float64 `json:"total_cost_usd"`
}

// UsageByAPIKey groups usage by API key
type UsageByAPIKey struct {
	APIKeyID           string  `json:"api_key_id"`
	TotalRequests      int     `json:"total_requests"`
	TotalPromptTokens  int64   `json:"total_prompt_tokens"`
	TotalCompletionTok int64   `json:"total_completion_tokens"`
	TotalTokens        int64   `json:"total_tokens"`
	TotalCostUSD       float64 `json:"total_cost_usd"`
}

// RecordUsage records a usage entry with detailed breakdown.
// ref: 9router/open-sse/services/usage.js (getUsageForProvider tracks all provider usage)
func (s *UsageService) RecordUsage(ctx context.Context, usage *models.UsageLogEntry) error {
	if usage.Timestamp.IsZero() {
		usage.Timestamp = time.Now()
	}
	if usage.ID == "" {
		usage.ID = generateID()
	}

	// Calculate cost if not already set
	if usage.CostUSD == 0 && usage.TotalTokens > 0 {
		cost, err := s.CalculateCost(ctx, usage)
		if err != nil {
			s.logger.Debug("failed to calculate cost, using zero", "error", err)
		} else {
			usage.CostUSD = cost
		}
	}

	err := s.storage.InsertUsageLogEntry(ctx, usage)
	if err != nil {
		s.logger.Error("failed to record usage", "error", err)
		return err
	}

	s.logger.Debug("recorded usage", "id", usage.ID, "model", usage.Model, "tokens", usage.TotalTokens)
	return nil
}

// GetUsageSummary returns aggregated usage statistics for a time period.
// ref: 9router/open-sse/services/usage.js (aggregate usage patterns)
func (s *UsageService) GetUsageSummary(ctx context.Context, filters UsageFilters) (*UsageSummary, error) {
	startTime := time.Now().AddDate(0, -1, 0) // Default: last 30 days
	endTime := time.Now()

	if filters.StartDate != nil {
		startTime = *filters.StartDate
	}
	if filters.EndDate != nil {
		endTime = *filters.EndDate
	}

	stats, err := s.storage.GetUsageLogStats(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	return &UsageSummary{
		TotalRequests:      stats.TotalRequests,
		SuccessfulRequests: stats.SuccessfulRequests,
		FailedRequests:     stats.FailedRequests,
		TotalPromptTokens:  stats.TotalPromptTokens,
		TotalCompletionTok: stats.TotalCompletionTok,
		TotalTokens:        stats.TotalTokens,
		TotalCostUSD:       stats.TotalCostUSD,
		AvgDurationMs:      stats.AvgDurationMs,
	}, nil
}

// GetUsageByProvider returns usage aggregated by provider.
// ref: 9router/open-sse/services/usage.js:60-91 (provider-specific usage tracking)
func (s *UsageService) GetUsageByProvider(ctx context.Context, filters UsageFilters) ([]UsageByProvider, error) {
	query := `
		SELECT provider_id, 
		       COUNT(*) as total_requests,
		       COALESCE(SUM(prompt_tokens), 0) as total_prompt_tokens,
		       COALESCE(SUM(completion_tokens), 0) as total_completion_tokens,
		       COALESCE(SUM(total_tokens), 0) as total_tokens,
		       COALESCE(SUM(cost_usd), 0) as total_cost_usd
		FROM usage_log 
		WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, *filters.StartDate)
		argNum++
	}
	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, *filters.EndDate)
		argNum++
	}

	query += " GROUP BY provider_id ORDER BY total_requests DESC"

	rows, err := s.storage.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage by provider: %w", err)
	}
	defer rows.Close()

	var results []UsageByProvider
	for rows.Next() {
		var u UsageByProvider
		if err := rows.Scan(&u.ProviderID, &u.TotalRequests, &u.TotalPromptTokens, 
			&u.TotalCompletionTok, &u.TotalTokens, &u.TotalCostUSD); err != nil {
			return nil, fmt.Errorf("failed to scan usage by provider: %w", err)
		}
		results = append(results, u)
	}

	return results, rows.Err()
}

// GetUsageByModel returns usage aggregated by model.
// ref: 9router/open-sse/services/usage.js (model-specific usage patterns)
func (s *UsageService) GetUsageByModel(ctx context.Context, filters UsageFilters) ([]UsageByModel, error) {
	query := `
		SELECT model, 
		       COUNT(*) as total_requests,
		       COALESCE(SUM(prompt_tokens), 0) as total_prompt_tokens,
		       COALESCE(SUM(completion_tokens), 0) as total_completion_tokens,
		       COALESCE(SUM(total_tokens), 0) as total_tokens,
		       COALESCE(SUM(cost_usd), 0) as total_cost_usd
		FROM usage_log 
		WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, *filters.StartDate)
		argNum++
	}
	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, *filters.EndDate)
		argNum++
	}

	query += " GROUP BY model ORDER BY total_requests DESC"

	rows, err := s.storage.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage by model: %w", err)
	}
	defer rows.Close()

	var results []UsageByModel
	for rows.Next() {
		var u UsageByModel
		if err := rows.Scan(&u.Model, &u.TotalRequests, &u.TotalPromptTokens, 
			&u.TotalCompletionTok, &u.TotalTokens, &u.TotalCostUSD); err != nil {
			return nil, fmt.Errorf("failed to scan usage by model: %w", err)
		}
		results = append(results, u)
	}

	return results, rows.Err()
}

// GetUsageByAPIKey returns usage aggregated by API key.
// ref: 9router/open-sse/services/usage.js (API key tracking)
func (s *UsageService) GetUsageByAPIKey(ctx context.Context, filters UsageFilters) ([]UsageByAPIKey, error) {
	query := `
		SELECT api_key_id, 
		       COUNT(*) as total_requests,
		       COALESCE(SUM(prompt_tokens), 0) as total_prompt_tokens,
		       COALESCE(SUM(completion_tokens), 0) as total_completion_tokens,
		       COALESCE(SUM(total_tokens), 0) as total_tokens,
		       COALESCE(SUM(cost_usd), 0) as total_cost_usd
		FROM usage_log 
		WHERE api_key_id IS NOT NULL`
	args := []interface{}{}
	argNum := 1

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, *filters.StartDate)
		argNum++
	}
	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, *filters.EndDate)
		argNum++
	}
	if filters.APIKeyID != "" {
		query += fmt.Sprintf(" AND api_key_id = $%d", argNum)
		args = append(args, filters.APIKeyID)
	}

	query += " GROUP BY api_key_id ORDER BY total_requests DESC"

	rows, err := s.storage.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage by API key: %w", err)
	}
	defer rows.Close()

	var results []UsageByAPIKey
	for rows.Next() {
		var u UsageByAPIKey
		if err := rows.Scan(&u.APIKeyID, &u.TotalRequests, &u.TotalPromptTokens, 
			&u.TotalCompletionTok, &u.TotalTokens, &u.TotalCostUSD); err != nil {
			return nil, fmt.Errorf("failed to scan usage by API key: %w", err)
		}
		results = append(results, u)
	}

	return results, rows.Err()
}

// CalculateCost calculates the cost for a usage entry based on model pricing.
// ref: 9router/open-sse/services/usage.js (cost calculation patterns)
func (s *UsageService) CalculateCost(ctx context.Context, usage *models.UsageLogEntry) (float64, error) {
	if usage.Model == "" {
		return 0, nil
	}

	// Get pricing for model
	var pricing models.ModelPricing
	err := s.storage.Pool().QueryRow(ctx,
		`SELECT model_pattern, prompt_price_per_1m, completion_price_per_1m, updated_at 
		 FROM model_pricing 
		 WHERE $1 LIKE model_pattern 
		 ORDER BY LENGTH(model_pattern) DESC 
		 LIMIT 1`,
		usage.Model,
	).Scan(&pricing.ModelPattern, &pricing.PromptPricePer1M, &pricing.CompletionPricePer1M, &pricing.UpdatedAt)

	if err != nil {
		// No pricing found, return zero cost
		s.logger.Debug("no pricing found for model", "model", usage.Model)
		return 0, nil
	}

	// Calculate cost: (prompt_tokens / 1M * prompt_price) + (completion_tokens / 1M * completion_price)
	promptCost := float64(usage.PromptTokens) / 1_000_000 * pricing.PromptPricePer1M
	completionCost := float64(usage.CompletionTokens) / 1_000_000 * pricing.CompletionPricePer1M

	return promptCost + completionCost, nil
}

// GetProviderQuota fetches quota information from provider-specific APIs.
// ref: 9router/open-sse/services/usage.js:60-91 (getUsageForProvider switch statement)
func (s *UsageService) GetProviderQuota(ctx context.Context, provider, accessToken string, providerData map[string]any) (*ProviderQuota, error) {
	switch provider {
	case "github":
		return s.getGitHubQuota(ctx, accessToken)
	case "gemini-cli":
		return s.getGeminiQuota(ctx, accessToken, providerData)
	case "antigravity":
		return s.getAntigravityQuota(ctx, accessToken, providerData)
	case "claude":
		return s.getCL4udeQuota(ctx, accessToken)
	case "codex":
		return s.getCodexQuota(ctx, accessToken)
	case "kiro":
		return s.getKiroQuota(ctx, accessToken, providerData)
	case "qwen":
		return s.getQwenQuota(ctx, accessToken, providerData)
	case "glm", "glm-cn":
		return s.getGLMQuota(ctx, accessToken, provider)
	case "minimax", "minimax-cn":
		return s.getMiniMaxQuota(ctx, accessToken, provider)
	default:
		return &ProviderQuota{
			Message: fmt.Sprintf("Usage API not implemented for %s", provider),
		}, nil
	}
}

// getGitHubQuota fetches GitHub Copilot quota.
// ref: 9router/open-sse/services/usage.js:131-201
func (s *UsageService) getGitHubQuota(ctx context.Context, accessToken string) (*ProviderQuota, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("no GitHub access token available")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/copilot_internal/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "GitHubCopilotChat/0.26.7")
	req.Header.Set("Editor-Version", "vscode/1.100.0")
	req.Header.Set("Editor-Plugin-Version", "copilot-chat/0.26.7")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		CopilotPlan       string `json:"copilot_plan"`
		QuotaResetDate    string `json:"quota_reset_date"`
		QuotaSnapshots    map[string]struct {
			Entitlement int  `json:"entitlement"`
			Remaining   int  `json:"remaining"`
			Unlimited   bool `json:"unlimited"`
		} `json:"quota_snapshots"`
		MonthlyQuotas map[string]int `json:"monthly_quotas"`
		LimitedUserQuotas map[string]int `json:"limited_user_quotas"`
		LimitedUserResetDate string `json:"limited_user_reset_date"`
		AccessTypeSku string `json:"access_type_sku"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	quotas := make(map[string]Quota)

	// Paid plan format
	if data.QuotaSnapshots != nil {
		for key, snapshot := range data.QuotaSnapshots {
			used := snapshot.Entitlement - snapshot.Remaining
			quotas[key] = Quota{
				Used:      used,
				Total:     snapshot.Entitlement,
				Remaining: snapshot.Remaining,
				Unlimited: snapshot.Unlimited,
				ResetAt:   data.QuotaResetDate,
			}
		}
		return &ProviderQuota{
			Plan:      data.CopilotPlan,
			ResetDate: data.QuotaResetDate,
			Quotas:    quotas,
		}, nil
	}

	// Free/limited plan format
	if data.MonthlyQuotas != nil || data.LimitedUserQuotas != nil {
		monthlyQuotas := data.MonthlyQuotas
		if monthlyQuotas == nil {
			monthlyQuotas = make(map[string]int)
		}
		usedQuotas := data.LimitedUserQuotas
		if usedQuotas == nil {
			usedQuotas = make(map[string]int)
		}

		for _, key := range []string{"chat", "completions"} {
			used := usedQuotas[key]
			total := monthlyQuotas[key]
			quotas[key] = Quota{
				Used:      used,
				Total:     total,
				Remaining: total - used,
				Unlimited: false,
				ResetAt:   data.LimitedUserResetDate,
			}
		}

		plan := data.CopilotPlan
		if plan == "" {
			plan = data.AccessTypeSku
		}

		return &ProviderQuota{
			Plan:      plan,
			ResetDate: data.LimitedUserResetDate,
			Quotas:    quotas,
		}, nil
	}

	return &ProviderQuota{
		Plan:    data.CopilotPlan,
		Message: "GitHub Copilot connected. Unable to parse quota data.",
	}, nil
}

// getGeminiQuota fetches Gemini CLI quota via Cloud Code Assist API.
// ref: 9router/open-sse/services/usage.js:219-290
func (s *UsageService) getGeminiQuota(ctx context.Context, accessToken string, providerData map[string]any) (*ProviderQuota, error) {
	if accessToken == "" {
		return &ProviderQuota{Plan: "Free", Message: "Gemini CLI access token not available."}, nil
	}

	// Resolve project ID
	projectID, _ := providerData["projectId"].(string)
	plan := "Free"

	if projectID == "" {
		// Try to get subscription info
		subInfo, err := s.getGeminiSubscriptionInfo(ctx, accessToken)
		if err == nil && subInfo != nil {
			if pid, ok := subInfo["cloudaicompanionProject"].(string); ok {
				projectID = pid
			}
			if tier, ok := subInfo["currentTier"].(map[string]any); ok {
				if name, ok := tier["name"].(string); ok {
					plan = name
				}
			}
		}
	}

	if projectID == "" {
		return &ProviderQuota{Plan: plan, Message: "Gemini CLI project ID not available."}, nil
	}

	reqBody := map[string]any{"project": projectID}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", 
		"https://cloudcode-pa.googleapis.com/v1internal:retrieveUserQuota", 
		strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &ProviderQuota{Message: fmt.Sprintf("Gemini CLI error: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ProviderQuota{Plan: plan, Message: fmt.Sprintf("Gemini CLI quota error (%d).", resp.StatusCode)}, nil
	}

	var data struct {
		Buckets []struct {
			ModelID           string  `json:"modelId"`
			RemainingFraction float64 `json:"remainingFraction"`
			ResetTime         string  `json:"resetTime"`
		} `json:"buckets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	quotas := make(map[string]Quota)
	for _, bucket := range data.Buckets {
		if bucket.ModelID == "" || bucket.RemainingFraction == 0 {
			continue
		}

		total := 1000 // Normalized base
		remaining := int(float64(total) * bucket.RemainingFraction)
		used := total - remaining

		quotas[bucket.ModelID] = Quota{
			Used:                used,
			Total:               total,
			Remaining:           remaining,
			RemainingPercentage: bucket.RemainingFraction * 100,
			ResetAt:             bucket.ResetTime,
			Unlimited:           false,
		}
	}

	return &ProviderQuota{Plan: plan, Quotas: quotas}, nil
}

// getGeminiSubscriptionInfo fetches Gemini subscription info.
// ref: 9router/open-sse/services/usage.js:295-325
func (s *UsageService) getGeminiSubscriptionInfo(ctx context.Context, accessToken string) (map[string]any, error) {
	reqBody := map[string]any{
		"metadata": map[string]any{
			"ideType":    "IDE_UNSPECIFIED",
			"platform":   "PLATFORM_UNSPECIFIED",
			"pluginType": "GEMINI",
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist",
		strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription info error: %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// getAntigravityQuota fetches Antigravity quota from Google Cloud Code API.
// ref: 9router/open-sse/services/usage.js:330-434
func (s *UsageService) getAntigravityQuota(ctx context.Context, accessToken string, providerData map[string]any) (*ProviderQuota, error) {
	// Fetch subscription info
	subInfo, err := s.getAntigravitySubscriptionInfo(ctx, accessToken)
	plan := "Unknown"
	var projectID string

	if err == nil && subInfo != nil {
		if pid, ok := subInfo["cloudaicompanionProject"].(string); ok {
			projectID = pid
		}
		if tier, ok := subInfo["currentTier"].(map[string]any); ok {
			if name, ok := tier["name"].(string); ok {
				plan = name
			}
		}
	}

	reqBody := make(map[string]any)
	if projectID != "" {
		reqBody["project"] = projectID
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels",
		strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Name", "antigravity")
	req.Header.Set("X-Client-Version", "1.107.0")
	req.Header.Set("x-request-source", "local")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("antigravity error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return &ProviderQuota{
			Message: "Antigravity quota API access forbidden. Chat may still work.",
			Quotas:  make(map[string]Quota),
		}, nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return &ProviderQuota{
			Message: "Antigravity quota API authentication expired. Chat may still work.",
			Quotas:  make(map[string]Quota),
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("antigravity API error: %d", resp.StatusCode)
	}

	var data struct {
		Models map[string]struct {
			IsInternal  bool `json:"isInternal"`
			DisplayName string `json:"displayName"`
			QuotaInfo   struct {
				RemainingFraction float64 `json:"remainingFraction"`
				ResetTime         string  `json:"resetTime"`
			} `json:"quotaInfo"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse antigravity response: %w", err)
	}

	// Filter important models
	importantModels := map[string]bool{
		"claude-opus-4-6-thinking": true,
		"claude-sonnet-4-6":        true,
		"gemini-3.1-pro-high":      true,
		"gemini-3.1-pro-low":       true,
		"gemini-3-flash":           true,
		"gpt-oss-120b-medium":      true,
	}

	quotas := make(map[string]Quota)
	for modelKey, info := range data.Models {
		if info.IsInternal || !importantModels[modelKey] {
			continue
		}

		remainingFraction := info.QuotaInfo.RemainingFraction
		total := 1000
		remaining := int(float64(total) * remainingFraction)
		used := total - remaining

		quotas[modelKey] = Quota{
			Used:                used,
			Total:               total,
			Remaining:           remaining,
			RemainingPercentage: remainingFraction * 100,
			ResetAt:             info.QuotaInfo.ResetTime,
			Unlimited:           false,
			DisplayName:         info.DisplayName,
		}
	}

	return &ProviderQuota{
		Plan:   plan,
		Quotas: quotas,
	}, nil
}

// getAntigravitySubscriptionInfo fetches Antigravity subscription info.
// ref: 9router/open-sse/services/usage.js:451-475
func (s *UsageService) getAntigravitySubscriptionInfo(ctx context.Context, accessToken string) (map[string]any, error) {
	reqBody := map[string]any{
		"metadata": map[string]any{
			"ideType":    "IDE_UNSPECIFIED",
			"platform":   "PLATFORM_UNSPECIFIED",
			"pluginType": "GEMINI",
		},
		"mode": 1,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist",
		strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-request-source", "local")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription info error: %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// getCL4udeQuota fetches CL4ude usage via OAuth endpoint.
// ref: 9router/open-sse/services/usage.js:480-594
func (s *UsageService) getCL4udeQuota(ctx context.Context, accessToken string) (*ProviderQuota, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.anthr0pic.com/api/oauth/usage", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("anthr0pic-beta", "oauth-2025-04-20")
	req.Header.Set("anthr0pic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CL4ude error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to legacy
		return s.getCL4udeQuotaLegacy(ctx, accessToken)
	}

	var data struct {
		FiveHour struct {
			Utilization int    `json:"utilization"`
			ResetsAt    string `json:"resets_at"`
		} `json:"five_hour"`
		SevenDay struct {
			Utilization int    `json:"utilization"`
			ResetsAt    string `json:"resets_at"`
		} `json:"seven_day"`
		ExtraUsage any `json:"extra_usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse CL4ude response: %w", err)
	}

	quotas := make(map[string]Quota)

	// Session (5h) window
	if data.FiveHour.Utilization > 0 || data.FiveHour.ResetsAt != "" {
		used := data.FiveHour.Utilization
		quotas["session (5h)"] = Quota{
			Used:      used,
			Total:     100,
			Remaining: 100 - used,
			ResetAt:   data.FiveHour.ResetsAt,
			Unlimited: false,
		}
	}

	// Weekly (7d) window
	if data.SevenDay.Utilization > 0 || data.SevenDay.ResetsAt != "" {
		used := data.SevenDay.Utilization
		quotas["weekly (7d)"] = Quota{
			Used:      used,
			Total:     100,
			Remaining: 100 - used,
			ResetAt:   data.SevenDay.ResetsAt,
			Unlimited: false,
		}
	}

	return &ProviderQuota{
		Plan:   "Code Assistant",
		Quotas: quotas,
	}, nil
}

// getCL4udeQuotaLegacy fetches CL4ude usage via legacy settings/org endpoint.
// ref: 9router/open-sse/services/usage.js:547-594
func (s *UsageService) getCL4udeQuotaLegacy(ctx context.Context, accessToken string) (*ProviderQuota, error) {
	// Get settings first
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.anthr0pic.com/v1/settings", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("anthr0pic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CL4ude settings error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ProviderQuota{Message: "CL4ude connected. Usage API requires admin permissions."}, nil
	}

	var settings struct {
		OrganizationID   string `json:"organization_id"`
		OrganizationName string `json:"organization_name"`
		Plan             string `json:"plan"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, fmt.Errorf("failed to parse CL4ude settings: %w", err)
	}

	if settings.OrganizationID == "" {
		return &ProviderQuota{
			Plan:    settings.Plan,
			Message: "CL4ude connected. Usage details require admin access.",
		}, nil
	}

	// Get org usage
	usageURL := fmt.Sprintf("https://api.anthr0pic.com/v1/organizations/%s/usage", settings.OrganizationID)
	req, err = http.NewRequestWithContext(ctx, "GET", usageURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("anthr0pic-version", "2023-06-01")

	usageResp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CL4ude usage error: %w", err)
	}
	defer usageResp.Body.Close()

	if usageResp.StatusCode != http.StatusOK {
		return &ProviderQuota{
			Plan:    settings.Plan,
			Message: "CL4ude connected. Usage details require admin access.",
		}, nil
	}

	var usageData map[string]any
	if err := json.NewDecoder(usageResp.Body).Decode(&usageData); err != nil {
		return nil, fmt.Errorf("failed to parse CL4ude usage: %w", err)
	}

	quotas := make(map[string]Quota)
	// Parse usage data into quotas
	for k, v := range usageData {
		if m, ok := v.(map[string]any); ok {
			quota := Quota{Unlimited: false}
			if used, ok := m["used"].(float64); ok {
				quota.Used = int(used)
			}
			if total, ok := m["total"].(float64); ok {
				quota.Total = int(total)
			}
			if reset, ok := m["reset_at"].(string); ok {
				quota.ResetAt = reset
			}
			quotas[k] = quota
		}
	}

	return &ProviderQuota{
		Plan:   settings.Plan,
		Quotas: quotas,
	}, nil
}

// getCodexQuota fetches Codex (OpenAI) usage from ChatGPT backend API.
// ref: 9router/open-sse/services/usage.js:663-694
func (s *UsageService) getCodexQuota(ctx context.Context, accessToken string) (*ProviderQuota, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://chatgpt.com/backend-api/wham/usage", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Codex error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ProviderQuota{
			Message: fmt.Sprintf("Codex connected. Usage API temporarily unavailable (%d).", resp.StatusCode),
		}, nil
	}

	var data struct {
		PlanType string `json:"plan_type"`
		Summary  struct {
			Plan string `json:"plan"`
		} `json:"summary"`
		RateLimit struct {
			LimitReached bool `json:"limit_reached"`
			PrimaryWindow struct {
				UsedPercent float64 `json:"used_percent"`
				ResetAt     string  `json:"reset_at"`
			} `json:"primary_window"`
			SecondaryWindow struct {
				UsedPercent float64 `json:"used_percent"`
				ResetAt     string  `json:"reset_at"`
			} `json:"secondary_window"`
		} `json:"rate_limit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse Codex response: %w", err)
	}

	quotas := make(map[string]Quota)

	// Session window
	if data.RateLimit.PrimaryWindow.UsedPercent > 0 || data.RateLimit.PrimaryWindow.ResetAt != "" {
		used := int(data.RateLimit.PrimaryWindow.UsedPercent)
		quotas["session"] = Quota{
			Used:      used,
			Total:     100,
			Remaining: 100 - used,
			ResetAt:   data.RateLimit.PrimaryWindow.ResetAt,
			Unlimited: false,
		}
	}

	// Weekly window
	if data.RateLimit.SecondaryWindow.UsedPercent > 0 || data.RateLimit.SecondaryWindow.ResetAt != "" {
		used := int(data.RateLimit.SecondaryWindow.UsedPercent)
		quotas["weekly"] = Quota{
			Used:      used,
			Total:     100,
			Remaining: 100 - used,
			ResetAt:   data.RateLimit.SecondaryWindow.ResetAt,
			Unlimited: false,
		}
	}

	plan := data.PlanType
	if plan == "" {
		plan = data.Summary.Plan
	}

	return &ProviderQuota{
		Plan:   plan,
		Quotas: quotas,
	}, nil
}

// getKiroQuota fetches Kiro (AWS CodeWhisperer) usage.
// ref: 9router/open-sse/services/usage.js:738-857
func (s *UsageService) getKiroQuota(ctx context.Context, accessToken string, providerData map[string]any) (*ProviderQuota, error) {
	profileArn := os.Getenv("KIRO_PROFILE_ARN")
	if pa, ok := providerData["profileArn"].(string); ok && pa != "" {
		profileArn = pa
	}

	params := url.Values{}
	params.Set("isEmailRequired", "true")
	params.Set("origin", "AI_EDITOR")
	params.Set("resourceType", "AGENTIC_REQUEST")

	// Try CodeWhisperer GET endpoint
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://codewhisperer.us-east-1.amazonaws.com/getUsageLimits?%s", params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-amz-user-agent", "aws-sdk-js/1.0.0 KiroIDE")
	req.Header.Set("user-agent", "aws-sdk-js/1.0.0 KiroIDE")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Kiro error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &ProviderQuota{
			Message: "Kiro quota API authentication expired. Chat may still work.",
			Quotas:  make(map[string]Quota),
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		// Try POST endpoint
		return s.getKiroQuotaPost(ctx, accessToken, profileArn)
	}

	var data struct {
		UsageBreakdownList []struct {
			ResourceType              string  `json:"resourceType"`
			CurrentUsageWithPrecision int     `json:"currentUsageWithPrecision"`
			UsageLimitWithPrecision   int     `json:"usageLimitWithPrecision"`
			FreeTrialInfo             *struct {
				CurrentUsageWithPrecision int    `json:"currentUsageWithPrecision"`
				UsageLimitWithPrecision   int    `json:"usageLimitWithPrecision"`
				FreeTrialExpiry           string `json:"freeTrialExpiry"`
			} `json:"freeTrialInfo"`
		} `json:"usageBreakdownList"`
		NextDateReset    string `json:"nextDateReset"`
		SubscriptionInfo struct {
			SubscriptionTitle string `json:"subscriptionTitle"`
		} `json:"subscriptionInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse Kiro response: %w", err)
	}

	quotas := make(map[string]Quota)
	resetAt := data.NextDateReset

	for _, breakdown := range data.UsageBreakdownList {
		resourceType := strings.ToLower(breakdown.ResourceType)
		if resourceType == "" {
			resourceType = "unknown"
		}

		quotas[resourceType] = Quota{
			Used:      breakdown.CurrentUsageWithPrecision,
			Total:     breakdown.UsageLimitWithPrecision,
			Remaining: breakdown.UsageLimitWithPrecision - breakdown.CurrentUsageWithPrecision,
			ResetAt:   resetAt,
			Unlimited: false,
		}

		if breakdown.FreeTrialInfo != nil {
			quotas[resourceType+"_freetrial"] = Quota{
				Used:      breakdown.FreeTrialInfo.CurrentUsageWithPrecision,
				Total:     breakdown.FreeTrialInfo.UsageLimitWithPrecision,
				Remaining: breakdown.FreeTrialInfo.UsageLimitWithPrecision - breakdown.FreeTrialInfo.CurrentUsageWithPrecision,
				ResetAt:   breakdown.FreeTrialInfo.FreeTrialExpiry,
				Unlimited: false,
			}
		}
	}

	plan := data.SubscriptionInfo.SubscriptionTitle
	if plan == "" {
		plan = "Kiro"
	}

	return &ProviderQuota{
		Plan:   plan,
		Quotas: quotas,
	}, nil
}

// getKiroQuotaPost tries the POST endpoint for Kiro quota.
// ref: 9router/open-sse/services/usage.js:768-784
func (s *UsageService) getKiroQuotaPost(ctx context.Context, accessToken, profileArn string) (*ProviderQuota, error) {
	reqBody := map[string]any{
		"origin":       "AI_EDITOR",
		"profileArn":   profileArn,
		"resourceType": "AGENTIC_REQUEST",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://codewhisperer.us-east-1.amazonaws.com", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	req.Header.Set("x-amz-target", "AmazonCodeWhispererService.GetUsageLimits")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Kiro POST error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &ProviderQuota{
			Message: "Kiro quota API authentication expired. Chat may still work.",
			Quotas:  make(map[string]Quota),
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &ProviderQuota{
			Message: fmt.Sprintf("Unable to fetch Kiro usage right now. (%d)", resp.StatusCode),
			Quotas:  make(map[string]Quota),
		}, nil
	}

	var data struct {
		UsageBreakdownList []struct {
			ResourceType              string `json:"resourceType"`
			CurrentUsageWithPrecision int    `json:"currentUsageWithPrecision"`
			UsageLimitWithPrecision   int    `json:"usageLimitWithPrecision"`
		} `json:"usageBreakdownList"`
		NextDateReset    string `json:"nextDateReset"`
		SubscriptionInfo struct {
			SubscriptionTitle string `json:"subscriptionTitle"`
		} `json:"subscriptionInfo"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse Kiro POST response: %w", err)
	}

	quotas := make(map[string]Quota)
	for _, breakdown := range data.UsageBreakdownList {
		resourceType := strings.ToLower(breakdown.ResourceType)
		quotas[resourceType] = Quota{
			Used:      breakdown.CurrentUsageWithPrecision,
			Total:     breakdown.UsageLimitWithPrecision,
			Remaining: breakdown.UsageLimitWithPrecision - breakdown.CurrentUsageWithPrecision,
			ResetAt:   data.NextDateReset,
			Unlimited: false,
		}
	}

	return &ProviderQuota{
		Plan:   data.SubscriptionInfo.SubscriptionTitle,
		Quotas: quotas,
	}, nil
}

// getQwenQuota fetches Qwen usage.
// ref: 9router/open-sse/services/usage.js:862-874
func (s *UsageService) getQwenQuota(ctx context.Context, accessToken string, providerData map[string]any) (*ProviderQuota, error) {
	resourceURL, _ := providerData["resourceUrl"].(string)
	if resourceURL == "" {
		return &ProviderQuota{Message: "Qwen connected. No resource URL available."}, nil
	}

	// Qwen may have usage endpoint at resource URL
	return &ProviderQuota{Message: "Qwen connected. Usage tracked per request."}, nil
}

// getGLMQuota fetches GLM Coding Plan usage (international + China regions).
// ref: 9router/open-sse/services/usage.js:913-966
func (s *UsageService) getGLMQuota(ctx context.Context, apiKey, provider string) (*ProviderQuota, error) {
	if apiKey == "" {
		return &ProviderQuota{Message: "GLM API key not available."}, nil
	}

	// Determine region
	quotaURL := "https://api.z.ai/api/monitor/usage/quota/limit"
	if provider == "glm-cn" {
		quotaURL = "https://open.bigmodel.cn/api/monitor/usage/quota/limit"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", quotaURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GLM error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return &ProviderQuota{Message: "GLM API key invalid or expired."}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &ProviderQuota{Message: fmt.Sprintf("GLM quota API error (%d).", resp.StatusCode)}, nil
	}

	var result struct {
		Data struct {
			Level  string `json:"level"`
			Limits []struct {
				Type          string  `json:"type"`
				Percentage    float64 `json:"percentage"`
				NextResetTime int64   `json:"nextResetTime"`
			} `json:"limits"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse GLM response: %w", err)
	}

	quotas := make(map[string]Quota)
	for _, limit := range result.Data.Limits {
		if limit.Type != "TOKENS_LIMIT" {
			continue
		}

		usedPercent := int(limit.Percentage)
		remaining := 100 - usedPercent
		var resetAt string
		if limit.NextResetTime > 0 {
			resetAt = time.UnixMilli(limit.NextResetTime).Format(time.RFC3339)
		}

		quotas["session"] = Quota{
			Used:                usedPercent,
			Total:               100,
			Remaining:           remaining,
			RemainingPercentage: float64(remaining),
			ResetAt:             resetAt,
			Unlimited:           false,
		}
		break // Only one TOKENS_LIMIT expected
	}

	// Format plan name
	level := result.Data.Level
	plan := "Unknown"
	if level != "" {
		plan = strings.ToUpper(level[:1]) + strings.ToLower(level[1:])
	}

	return &ProviderQuota{Plan: plan, Quotas: quotas}, nil
}

// getMiniMaxQuota fetches MiniMax Token Plan / Coding Plan usage.
// ref: 9router/open-sse/services/usage.js:1017-1111
func (s *UsageService) getMiniMaxQuota(ctx context.Context, apiKey, provider string) (*ProviderQuota, error) {
	if apiKey == "" {
		return &ProviderQuota{Message: "MiniMax API key not available."}, nil
	}

	// Determine usage URLs based on provider
	var usageURLs []string
	if provider == "minimax-cn" {
		usageURLs = []string{
			"https://www.minimaxi.com/v1/api/openplatform/coding_plan/remains",
			"https://api.minimaxi.com/v1/api/openplatform/coding_plan/remains",
		}
	} else {
		usageURLs = []string{
			"https://www.minimax.io/v1/token_plan/remains",
			"https://api.minimax.io/v1/api/openplatform/coding_plan/remains",
		}
	}

	var lastError string

	for _, usageURL := range usageURLs {
		req, err := http.NewRequestWithContext(ctx, "GET", usageURL, nil)
		if err != nil {
			lastError = err.Error()
			continue
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastError = err.Error()
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check for auth errors
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return &ProviderQuota{Message: "MiniMax API key invalid or inactive. Use an active Token/Coding Plan key."}, nil
		}

		if resp.StatusCode != http.StatusOK {
			lastError = fmt.Sprintf("MiniMax usage endpoint error (%d)", resp.StatusCode)
			continue
		}

		var payload struct {
			BaseResp struct {
				StatusCode int    `json:"status_code"`
				StatusMsg  string `json:"status_msg"`
			} `json:"base_resp"`
			ModelRemains []struct {
				ModelName                   string `json:"model_name"`
				CurrentIntervalTotalCount   int    `json:"current_interval_total_count"`
				CurrentIntervalUsageCount   int    `json:"current_interval_usage_count"`
				RemainsTime                 int64  `json:"remains_time"`
				CurrentWeeklyTotalCount     int    `json:"current_weekly_total_count"`
				CurrentWeeklyUsageCount     int    `json:"current_weekly_usage_count"`
				WeeklyRemainsTime           int64  `json:"weekly_remains_time"`
			} `json:"model_remains"`
		}

		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			lastError = "failed to parse MiniMax response"
			continue
		}

		if payload.BaseResp.StatusCode != 0 {
			return &ProviderQuota{Message: fmt.Sprintf("MiniMax connected. %s", payload.BaseResp.StatusMsg)}, nil
		}

		// Filter text quota models
		var textModels []struct {
			ModelName                   string `json:"model_name"`
			CurrentIntervalTotalCount   int    `json:"current_interval_total_count"`
			CurrentIntervalUsageCount   int    `json:"current_interval_usage_count"`
			RemainsTime                 int64  `json:"remains_time"`
			CurrentWeeklyTotalCount     int    `json:"current_weekly_total_count"`
			CurrentWeeklyUsageCount     int    `json:"current_weekly_usage_count"`
			WeeklyRemainsTime           int64  `json:"weekly_remains_time"`
		}

		for _, m := range payload.ModelRemains {
			modelName := strings.ToLower(m.ModelName)
			if strings.HasPrefix(modelName, "minimax-m") || strings.HasPrefix(modelName, "coding-plan") {
				textModels = append(textModels, m)
			}
		}

		if len(textModels) == 0 {
			return &ProviderQuota{Message: "MiniMax connected. No text quota data was returned."}, nil
		}

		quotas := make(map[string]Quota)
		now := time.Now().UnixMilli()
		countMeansRemaining := strings.Contains(usageURL, "/coding_plan/remains")

		// Find representative session model (highest total)
		var sessionModel *struct {
			ModelName                   string `json:"model_name"`
			CurrentIntervalTotalCount   int    `json:"current_interval_total_count"`
			CurrentIntervalUsageCount   int    `json:"current_interval_usage_count"`
			RemainsTime                 int64  `json:"remains_time"`
			CurrentWeeklyTotalCount     int    `json:"current_weekly_total_count"`
			CurrentWeeklyUsageCount     int    `json:"current_weekly_usage_count"`
			WeeklyRemainsTime           int64  `json:"weekly_remains_time"`
		}
		for i := range textModels {
			if sessionModel == nil || textModels[i].CurrentIntervalTotalCount > sessionModel.CurrentIntervalTotalCount {
				sessionModel = &textModels[i]
			}
		}

		if sessionModel != nil && sessionModel.CurrentIntervalTotalCount > 0 {
			total := sessionModel.CurrentIntervalTotalCount
			count := sessionModel.CurrentIntervalUsageCount
			var used int
			if countMeansRemaining {
				used = total - count
				if used < 0 {
					used = 0
				}
			} else {
				used = count
				if used > total {
					used = total
				}
			}
			remaining := total - used
			if remaining < 0 {
				remaining = 0
			}

			var resetAt string
			if sessionModel.RemainsTime > 0 {
				resetAt = time.UnixMilli(now + sessionModel.RemainsTime).Format(time.RFC3339)
			}

			quotas["session (5h)"] = Quota{
				Used:      used,
				Total:     total,
				Remaining: remaining,
				ResetAt:   resetAt,
				Unlimited: false,
			}
		}

		// Find representative weekly model
		var weeklyModel *struct {
			ModelName                   string `json:"model_name"`
			CurrentIntervalTotalCount   int    `json:"current_interval_total_count"`
			CurrentIntervalUsageCount   int    `json:"current_interval_usage_count"`
			RemainsTime                 int64  `json:"remains_time"`
			CurrentWeeklyTotalCount     int    `json:"current_weekly_total_count"`
			CurrentWeeklyUsageCount     int    `json:"current_weekly_usage_count"`
			WeeklyRemainsTime           int64  `json:"weekly_remains_time"`
		}
		for i := range textModels {
			if textModels[i].CurrentWeeklyTotalCount > 0 {
				if weeklyModel == nil || textModels[i].CurrentWeeklyTotalCount > weeklyModel.CurrentWeeklyTotalCount {
					weeklyModel = &textModels[i]
				}
			}
		}

		if weeklyModel != nil && weeklyModel.CurrentWeeklyTotalCount > 0 {
			total := weeklyModel.CurrentWeeklyTotalCount
			count := weeklyModel.CurrentWeeklyUsageCount
			var used int
			if countMeansRemaining {
				used = total - count
				if used < 0 {
					used = 0
				}
			} else {
				used = count
				if used > total {
					used = total
				}
			}
			remaining := total - used
			if remaining < 0 {
				remaining = 0
			}

			var resetAt string
			if weeklyModel.WeeklyRemainsTime > 0 {
				resetAt = time.UnixMilli(now + weeklyModel.WeeklyRemainsTime).Format(time.RFC3339)
			}

			quotas["weekly (7d)"] = Quota{
				Used:      used,
				Total:     total,
				Remaining: remaining,
				ResetAt:   resetAt,
				Unlimited: false,
			}
		}

		if len(quotas) == 0 {
			return &ProviderQuota{Message: "MiniMax connected. Unable to extract quota usage."}, nil
		}

		return &ProviderQuota{Quotas: quotas}, nil
	}

	return &ProviderQuota{Message: fmt.Sprintf("MiniMax connected. Unable to fetch usage: %s", lastError)}, nil
}

// parseResetTime parses reset date/time to ISO string.
// ref: 9router/open-sse/services/usage.js:97-125
func parseResetTime(resetValue any) string {
	if resetValue == nil {
		return ""
	}

	switch v := resetValue.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case string:
		// Check if numeric string
		if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			timestamp, _ := strconv.ParseInt(v, 10, 64)
			// Handle seconds vs milliseconds
			if timestamp < 1e12 {
				return time.Unix(timestamp, 0).Format(time.RFC3339)
			}
			return time.UnixMilli(timestamp).Format(time.RFC3339)
		}
		// Try parsing as ISO date string
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.Format(time.RFC3339)
		}
		return v
	case int:
		return time.Unix(int64(v), 0).Format(time.RFC3339)
	case int64:
		// Handle seconds vs milliseconds
		if v < 1e12 {
			return time.Unix(v, 0).Format(time.RFC3339)
		}
		return time.UnixMilli(v).Format(time.RFC3339)
	case float64:
		timestamp := int64(v)
		if timestamp < 1e12 {
			return time.Unix(timestamp, 0).Format(time.RFC3339)
		}
		return time.UnixMilli(timestamp).Format(time.RFC3339)
	default:
		return ""
	}
}

// generateID generates a unique ID for usage entries.
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
