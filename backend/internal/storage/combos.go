package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/jackc/pgx/v5"
)

var ErrComboNotFound = errors.New("combo not found")

// ref: open-sse/services/combo.js:82-94
func (db *DB) CreateCombo(ctx context.Context, combo *models.Combo) error {
	now := time.Now()
	combo.CreatedAt = now
	combo.UpdatedAt = now

	modelsJSON, err := json.Marshal(combo.Models)
	if err != nil {
		return err
	}

	_, err = db.pool.Exec(ctx,
		"INSERT INTO combos (name, models, strategy, sticky_limit, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		combo.Name, modelsJSON, combo.Strategy, combo.StickyLimit, combo.CreatedAt, combo.UpdatedAt,
	)
	return err
}

// ref: open-sse/services/combo.js:82-94
func (db *DB) GetComboByName(ctx context.Context, name string) (*models.Combo, error) {
	var combo models.Combo
	var modelsJSON []byte

	err := db.pool.QueryRow(ctx,
		"SELECT name, models, strategy, sticky_limit, created_at, updated_at FROM combos WHERE name = $1",
		name,
	).Scan(&combo.Name, &modelsJSON, &combo.Strategy, &combo.StickyLimit, &combo.CreatedAt, &combo.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrComboNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(modelsJSON, &combo.Models); err != nil {
		return nil, err
	}

	return &combo, nil
}

func (db *DB) ListCombos(ctx context.Context) ([]models.Combo, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT name, models, strategy, sticky_limit, created_at, updated_at FROM combos ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var combos []models.Combo
	for rows.Next() {
		var combo models.Combo
		var modelsJSON []byte
		if err := rows.Scan(&combo.Name, &modelsJSON, &combo.Strategy, &combo.StickyLimit, &combo.CreatedAt, &combo.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(modelsJSON, &combo.Models); err != nil {
			return nil, err
		}
		combos = append(combos, combo)
	}
	return combos, rows.Err()
}

func (db *DB) UpdateCombo(ctx context.Context, combo *models.Combo) error {
	combo.UpdatedAt = time.Now()

	modelsJSON, err := json.Marshal(combo.Models)
	if err != nil {
		return err
	}

	result, err := db.pool.Exec(ctx,
		"UPDATE combos SET models = $1, strategy = $2, sticky_limit = $3, updated_at = $4 WHERE name = $5",
		modelsJSON, combo.Strategy, combo.StickyLimit, combo.UpdatedAt, combo.Name,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrComboNotFound
	}
	return nil
}

func (db *DB) DeleteCombo(ctx context.Context, name string) error {
	result, err := db.pool.Exec(ctx, "DELETE FROM combos WHERE name = $1", name)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrComboNotFound
	}
	return nil
}
