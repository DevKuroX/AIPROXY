package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
)

func (db *DB) InsertUsageLogEntry(ctx context.Context, log *models.UsageLogEntry) error {
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	_, err := db.pool.Exec(ctx,
		`INSERT INTO usage_log (timestamp, model, provider_id, account_id, prompt_tokens, completion_tokens, total_tokens, cost_usd, rtk_bytes_saved, caveman_active, api_key_id, duration_ms, status, error_message)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		log.Timestamp, log.Model, log.ProviderID, log.AccountID, log.PromptTokens, log.CompletionTokens, log.TotalTokens, log.CostUSD, log.RTKBytesSaved, log.CavemanActive, log.APIKeyID, log.DurationMs, log.Status, log.ErrorMessage,
	)
	return err
}

func (db *DB) GetUsageLogEntries(ctx context.Context, filter models.UsageLogFilter) ([]models.UsageLogEntry, error) {
	query := `SELECT id, timestamp, model, provider_id, account_id, prompt_tokens, completion_tokens, total_tokens, cost_usd, rtk_bytes_saved, caveman_active, api_key_id, duration_ms, status, error_message FROM usage_log WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, *filter.StartTime)
		argNum++
	}
	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, *filter.EndTime)
		argNum++
	}
	if filter.Model != "" {
		query += fmt.Sprintf(" AND model = $%d", argNum)
		args = append(args, filter.Model)
		argNum++
	}
	if filter.ProviderID != "" {
		query += fmt.Sprintf(" AND provider_id = $%d", argNum)
		args = append(args, filter.ProviderID)
		argNum++
	}
	if filter.APIKeyID != "" {
		query += fmt.Sprintf(" AND api_key_id = $%d", argNum)
		args = append(args, filter.APIKeyID)
		argNum++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filter.Status)
		argNum++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.UsageLogEntry
	for rows.Next() {
		var l models.UsageLogEntry
		var id int
		var apiKeyID, accountID, errorMsg *string
		err := rows.Scan(&id, &l.Timestamp, &l.Model, &l.ProviderID, &accountID, &l.PromptTokens, &l.CompletionTokens, &l.TotalTokens, &l.CostUSD, &l.RTKBytesSaved, &l.CavemanActive, &apiKeyID, &l.DurationMs, &l.Status, &errorMsg)
		if err != nil {
			return nil, err
		}
		l.ID = fmt.Sprintf("%d", id)
		if apiKeyID != nil {
			l.APIKeyID = apiKeyID
		}
		if accountID != nil {
			l.AccountID = accountID
		}
		if errorMsg != nil {
			l.ErrorMessage = errorMsg
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (db *DB) GetUsageLogStats(ctx context.Context, startTime, endTime time.Time) (*models.UsageLogStats, error) {
	var stats models.UsageLogStats
	err := db.pool.QueryRow(ctx,
		`SELECT 
			COUNT(*) as total_requests,
			COUNT(*) FILTER (WHERE status = 'success') as successful_requests,
			COUNT(*) FILTER (WHERE status = 'error') as failed_requests,
			COALESCE(SUM(prompt_tokens), 0) as total_prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as total_completion_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost_usd), 0) as total_cost_usd,
			COALESCE(SUM(rtk_bytes_saved), 0) as total_rtk_bytes_saved,
			COALESCE(AVG(duration_ms), 0) as avg_duration_ms
		 FROM usage_log 
		 WHERE timestamp >= $1 AND timestamp <= $2`,
		startTime, endTime,
	).Scan(&stats.TotalRequests, &stats.SuccessfulRequests, &stats.FailedRequests, &stats.TotalPromptTokens, &stats.TotalCompletionTok, &stats.TotalTokens, &stats.TotalCostUSD, &stats.TotalRTKBytesSaved, &stats.AvgDurationMs)

	if err != nil {
		return nil, err
	}
	return &stats, nil
}
