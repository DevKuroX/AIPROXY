package storage

import (
	"context"
	"errors"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrModelAliasNotFound = errors.New("model alias not found")

func (db *DB) CreateModelAlias(ctx context.Context, alias *models.ModelAlias) error {
	if alias.ID == "" {
		alias.ID = uuid.New().String()
	}
	alias.CreatedAt = time.Now()

	_, err := db.pool.Exec(ctx,
		"INSERT INTO model_aliases (id, node_id, alias, target_model, created_at) VALUES ($1, $2, $3, $4, $5)",
		alias.ID, alias.NodeID, alias.Alias, alias.TargetModel, alias.CreatedAt,
	)
	return err
}

func (db *DB) GetModelAliasByAlias(ctx context.Context, aliasName string) (*models.ModelAlias, error) {
	var alias models.ModelAlias
	err := db.pool.QueryRow(ctx,
		"SELECT id, node_id, alias, target_model, created_at FROM model_aliases WHERE alias = $1",
		aliasName,
	).Scan(&alias.ID, &alias.NodeID, &alias.Alias, &alias.TargetModel, &alias.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrModelAliasNotFound
	}
	if err != nil {
		return nil, err
	}
	return &alias, nil
}

func (db *DB) ListModelAliases(ctx context.Context) ([]models.ModelAlias, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT id, node_id, alias, target_model, created_at FROM model_aliases ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []models.ModelAlias
	for rows.Next() {
		var a models.ModelAlias
		if err := rows.Scan(&a.ID, &a.NodeID, &a.Alias, &a.TargetModel, &a.CreatedAt); err != nil {
			return nil, err
		}
		aliases = append(aliases, a)
	}
	return aliases, rows.Err()
}

func (db *DB) DeleteModelAlias(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx,
		"DELETE FROM model_aliases WHERE id = $1",
		id,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrModelAliasNotFound
	}
	return nil
}
