package storage

import (
	"context"
	"errors"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrAPIKeyNotFound = errors.New("api key not found")

func (db *DB) CreateAPIKey(ctx context.Context, key *models.APIKey) error {
	if key.ID == "" {
		key.ID = uuid.New().String()
	}
	key.CreatedAt = time.Now()

	_, err := db.pool.Exec(ctx,
		"INSERT INTO api_keys (id, key, key_hash, name, is_active, created_at, last_used_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		key.ID, key.Key, key.KeyHash, key.Name, key.IsActive, key.CreatedAt, key.LastUsedAt,
	)
	return err
}

func (db *DB) GetAPIKeyByKey(ctx context.Context, key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := db.pool.QueryRow(ctx,
		"SELECT id, key, key_hash, name, is_active, created_at, last_used_at FROM api_keys WHERE key = $1",
		key,
	).Scan(&apiKey.ID, &apiKey.Key, &apiKey.KeyHash, &apiKey.Name, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.LastUsedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (db *DB) GetAPIKeyByID(ctx context.Context, id string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := db.pool.QueryRow(ctx,
		"SELECT id, key, key_hash, name, is_active, created_at, last_used_at FROM api_keys WHERE id = $1",
		id,
	).Scan(&apiKey.ID, &apiKey.Key, &apiKey.KeyHash, &apiKey.Name, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.LastUsedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (db *DB) ListAPIKeys(ctx context.Context) ([]models.APIKey, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT id, key, key_hash, name, is_active, created_at, last_used_at FROM api_keys ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.Key, &k.KeyHash, &k.Name, &k.IsActive, &k.CreatedAt, &k.LastUsedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (db *DB) UpdateAPIKey(ctx context.Context, key *models.APIKey) error {
	result, err := db.pool.Exec(ctx,
		"UPDATE api_keys SET name = $1, is_active = $2 WHERE id = $3",
		key.Name, key.IsActive, key.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (db *DB) DeleteAPIKey(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, "DELETE FROM api_keys WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (db *DB) UpdateAPIKeyLastUsed(ctx context.Context, id string) error {
	now := time.Now()
	result, err := db.pool.Exec(ctx,
		"UPDATE api_keys SET last_used_at = $1 WHERE id = $2",
		now, id,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}
