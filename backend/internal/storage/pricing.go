package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
)

var ErrPricingNotFound = errors.New("pricing not found")

func (db *DB) GetPricing(ctx context.Context, model string) (*models.ModelPricing, error) {
	rows, err := db.pool.Query(ctx,
		`SELECT model_pattern, prompt_price_per_1m, completion_price_per_1m, updated_at FROM pricing ORDER BY LENGTH(model_pattern) DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pricings []models.ModelPricing
	for rows.Next() {
		var p models.ModelPricing
		if err := rows.Scan(&p.ModelPattern, &p.PromptPricePer1M, &p.CompletionPricePer1M, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pricings = append(pricings, p)
	}

	for _, p := range pricings {
		if matchPattern(p.ModelPattern, model) {
			return &p, nil
		}
	}

	return nil, ErrPricingNotFound
}

func matchPattern(pattern, model string) bool {
	if pattern == model {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(model, prefix)
	}
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(model, suffix)
	}
	return false
}

func (db *DB) SetPricing(ctx context.Context, pricing *models.ModelPricing) error {
	pricing.UpdatedAt = time.Now()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO pricing (model_pattern, prompt_price_per_1m, completion_price_per_1m, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (model_pattern) 
		 DO UPDATE SET prompt_price_per_1m = $2, completion_price_per_1m = $3, updated_at = $4`,
		pricing.ModelPattern, pricing.PromptPricePer1M, pricing.CompletionPricePer1M, pricing.UpdatedAt,
	)
	return err
}

func (db *DB) ListPricing(ctx context.Context) ([]models.ModelPricing, error) {
	rows, err := db.pool.Query(ctx,
		`SELECT model_pattern, prompt_price_per_1m, completion_price_per_1m, updated_at FROM pricing ORDER BY model_pattern`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pricings []models.ModelPricing
	for rows.Next() {
		var p models.ModelPricing
		if err := rows.Scan(&p.ModelPattern, &p.PromptPricePer1M, &p.CompletionPricePer1M, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pricings = append(pricings, p)
	}
	return pricings, rows.Err()
}
