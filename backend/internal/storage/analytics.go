package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrPricingRuleNotFound = errors.New("pricing rule not found")

type UsageStats struct {
	TotalTokens int64                `json:"total_tokens"`
	TotalCost   float64              `json:"total_cost"`
	ByModel     map[string]ModelStat `json:"by_model"`
	ByProvider  map[string]ModelStat `json:"by_provider"`
}

type ModelStat struct {
	Tokens int64   `json:"tokens"`
	Cost   float64 `json:"cost"`
}

func (db *DB) ListUsageLogs(ctx context.Context, start, end time.Time, provider, model string, page, limit int) ([]models.UsageLog, int, error) {
	offset := (page - 1) * limit

	var args []interface{}
	argIdx := 1
	conditions := "WHERE timestamp BETWEEN $1 AND $2"
	args = append(args, start, end)

	if provider != "" {
		argIdx++
		conditions += " AND provider_id = $" + strconv.Itoa(argIdx)
		args = append(args, provider)
	}
	if model != "" {
		argIdx++
		conditions += " AND model = $" + strconv.Itoa(argIdx)
		args = append(args, model)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM usage_log " + conditions
	err := db.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT id, COALESCE(api_key_id,''), provider_id, model, prompt_tokens, completion_tokens, total_tokens, cost_usd, COALESCE(status,''), duration_ms, timestamp FROM usage_log " + conditions + " ORDER BY timestamp DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.UsageLog
	for rows.Next() {
		var l models.UsageLog
		var id int
		var apiKeyID, status string
		if err := rows.Scan(&id, &apiKeyID, &l.Provider, &l.Model, &l.InputTokens, &l.OutputTokens, &l.TotalTokens, &l.Cost, &status, &l.LatencyMs, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		l.ID = fmt.Sprintf("%d", id)
		l.APIKeyID = apiKeyID
		if status == "success" || status == "ok" {
			l.Status = 1
		}
		logs = append(logs, l)
	}

	return logs, total, rows.Err()
}

func (db *DB) GetUsageStats(ctx context.Context, start, end time.Time) (*UsageStats, error) {
	stats := &UsageStats{
		ByModel:    make(map[string]ModelStat),
		ByProvider: make(map[string]ModelStat),
	}

	err := db.pool.QueryRow(ctx,
		"SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(cost_usd), 0) FROM usage_log WHERE timestamp BETWEEN $1 AND $2",
		start, end,
	).Scan(&stats.TotalTokens, &stats.TotalCost)
	if err != nil {
		return nil, err
	}

	rows, err := db.pool.Query(ctx,
		"SELECT model, SUM(total_tokens), SUM(cost_usd) FROM usage_log WHERE timestamp BETWEEN $1 AND $2 GROUP BY model",
		start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var model string
		var stat ModelStat
		if err := rows.Scan(&model, &stat.Tokens, &stat.Cost); err != nil {
			return nil, err
		}
		stats.ByModel[model] = stat
	}

	rows, err = db.pool.Query(ctx,
		"SELECT provider_id, SUM(total_tokens), SUM(cost_usd) FROM usage_log WHERE timestamp BETWEEN $1 AND $2 GROUP BY provider_id",
		start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var provider string
		var stat ModelStat
		if err := rows.Scan(&provider, &stat.Tokens, &stat.Cost); err != nil {
			return nil, err
		}
		stats.ByProvider[provider] = stat
	}

	return stats, rows.Err()
}

func (db *DB) ListPricingRules(ctx context.Context) ([]models.PricingRule, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT id, provider, model, input_price, output_price, is_active, created_at, updated_at FROM pricing_rules ORDER BY provider, model",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.PricingRule
	for rows.Next() {
		var r models.PricingRule
		if err := rows.Scan(&r.ID, &r.Provider, &r.Model, &r.InputPrice, &r.OutputPrice, &r.IsActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (db *DB) GetPricingRule(ctx context.Context, id string) (*models.PricingRule, error) {
	var rule models.PricingRule
	err := db.pool.QueryRow(ctx,
		"SELECT id, provider, model, input_price, output_price, is_active, created_at, updated_at FROM pricing_rules WHERE id = $1",
		id,
	).Scan(&rule.ID, &rule.Provider, &rule.Model, &rule.InputPrice, &rule.OutputPrice, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPricingRuleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (db *DB) CreatePricingRule(ctx context.Context, rule *models.PricingRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	_, err := db.pool.Exec(ctx,
		"INSERT INTO pricing_rules (id, provider, model, input_price, output_price, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		rule.ID, rule.Provider, rule.Model, rule.InputPrice, rule.OutputPrice, rule.IsActive, rule.CreatedAt, rule.UpdatedAt,
	)
	return err
}

func (db *DB) UpdatePricingRule(ctx context.Context, rule *models.PricingRule) error {
	rule.UpdatedAt = time.Now()
	result, err := db.pool.Exec(ctx,
		"UPDATE pricing_rules SET provider = $1, model = $2, input_price = $3, output_price = $4, is_active = $5, updated_at = $6 WHERE id = $7",
		rule.Provider, rule.Model, rule.InputPrice, rule.OutputPrice, rule.IsActive, rule.UpdatedAt, rule.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPricingRuleNotFound
	}
	return nil
}

func (db *DB) DeletePricingRule(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, "DELETE FROM pricing_rules WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPricingRuleNotFound
	}
	return nil
}
