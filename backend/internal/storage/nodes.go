package storage

import (
	"context"
	"errors"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrProviderNodeNotFound = errors.New("provider node not found")

func (db *DB) CreateProviderNode(ctx context.Context, node *models.ProviderNode) error {
	if node.ID == "" {
		node.ID = uuid.New().String()
	}
	now := time.Now()
	node.CreatedAt = now
	node.UpdatedAt = now

	encKey, err := db.encryptAPIKey(node.APIKey)
	if err != nil {
		return err
	}

	_, err = db.pool.Exec(ctx,
		"INSERT INTO provider_nodes (id, name, base_url, api_key, compatible_format, enabled, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		node.ID, node.Name, node.BaseURL, encKey, node.CompatibleFormat, node.Enabled, node.CreatedAt, node.UpdatedAt,
	)
	return err
}

func (db *DB) GetProviderNodeByID(ctx context.Context, id string) (*models.ProviderNode, error) {
	var node models.ProviderNode
	err := db.pool.QueryRow(ctx,
		"SELECT id, name, base_url, api_key, compatible_format, enabled, created_at, updated_at FROM provider_nodes WHERE id = $1",
		id,
	).Scan(&node.ID, &node.Name, &node.BaseURL, &node.APIKey, &node.CompatibleFormat, &node.Enabled, &node.CreatedAt, &node.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProviderNodeNotFound
	}
	if err != nil {
		return nil, err
	}

	decKey, err := db.decryptAPIKey(node.APIKey)
	if err != nil {
		return nil, err
	}
	node.APIKey = decKey

	return &node, nil
}

func (db *DB) ListProviderNodes(ctx context.Context) ([]models.ProviderNode, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT id, name, base_url, api_key, compatible_format, enabled, created_at, updated_at FROM provider_nodes ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []models.ProviderNode
	for rows.Next() {
		var n models.ProviderNode
		if err := rows.Scan(&n.ID, &n.Name, &n.BaseURL, &n.APIKey, &n.CompatibleFormat, &n.Enabled, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		decKey, err := db.decryptAPIKey(n.APIKey)
		if err != nil {
			return nil, err
		}
		n.APIKey = decKey
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (db *DB) UpdateProviderNode(ctx context.Context, node *models.ProviderNode) error {
	node.UpdatedAt = time.Now()

	encKey, err := db.encryptAPIKey(node.APIKey)
	if err != nil {
		return err
	}

	result, err := db.pool.Exec(ctx,
		"UPDATE provider_nodes SET name = $1, base_url = $2, api_key = $3, compatible_format = $4, enabled = $5, updated_at = $6 WHERE id = $7",
		node.Name, node.BaseURL, encKey, node.CompatibleFormat, node.Enabled, node.UpdatedAt, node.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrProviderNodeNotFound
	}
	return nil
}

func (db *DB) DeleteProviderNode(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, "DELETE FROM provider_nodes WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrProviderNodeNotFound
	}
	return nil
}

func (db *DB) GetProviderNodeByModelAlias(ctx context.Context, alias string) (*models.ProviderNode, string, error) {
	var node models.ProviderNode
	var targetModel string
	err := db.pool.QueryRow(ctx,
		`SELECT pn.id, pn.name, pn.base_url, pn.api_key, pn.compatible_format, pn.enabled, pn.created_at, pn.updated_at, ma.target_model
		 FROM provider_nodes pn
		 JOIN model_aliases ma ON pn.id = ma.node_id
		 WHERE ma.alias = $1 AND pn.enabled = TRUE`,
		alias,
	).Scan(&node.ID, &node.Name, &node.BaseURL, &node.APIKey, &node.CompatibleFormat, &node.Enabled, &node.CreatedAt, &node.UpdatedAt, &targetModel)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrProviderNodeNotFound
	}
	if err != nil {
		return nil, "", err
	}

	decKey, err := db.decryptAPIKey(node.APIKey)
	if err != nil {
		return nil, "", err
	}
	node.APIKey = decKey

	return &node, targetModel, nil
}
