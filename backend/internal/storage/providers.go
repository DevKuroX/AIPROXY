package storage

import (
	"context"
	"errors"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrProviderNotFound = errors.New("provider not found")
var ErrProviderAccountNotFound = errors.New("provider account not found")

func (db *DB) CreateProvider(ctx context.Context, provider *models.Provider) error {
	if provider.ID == "" {
		provider.ID = uuid.New().String()
	}
	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	encKey, err := db.encryptAPIKey(provider.APIKey)
	if err != nil {
		return err
	}

	_, err = db.pool.Exec(ctx,
		"INSERT INTO providers (id, name, type, base_url, api_key, enabled, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		provider.ID, provider.Name, provider.Type, provider.BaseURL, encKey, provider.Enabled, provider.CreatedAt, provider.UpdatedAt,
	)
	return err
}

func (db *DB) GetProviderByID(ctx context.Context, id string) (*models.Provider, error) {
	var provider models.Provider
	err := db.pool.QueryRow(ctx,
		"SELECT id, name, type, base_url, api_key, enabled, created_at, updated_at FROM providers WHERE id = $1",
		id,
	).Scan(&provider.ID, &provider.Name, &provider.Type, &provider.BaseURL, &provider.APIKey, &provider.Enabled, &provider.CreatedAt, &provider.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProviderNotFound
	}
	if err != nil {
		return nil, err
	}

	decKey, err := db.decryptAPIKey(provider.APIKey)
	if err != nil {
		return nil, err
	}
	provider.APIKey = decKey

	return &provider, nil
}

func (db *DB) ListProviders(ctx context.Context) ([]models.Provider, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT id, name, type, base_url, api_key, enabled, created_at, updated_at FROM providers ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []models.Provider
	for rows.Next() {
		var p models.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.APIKey, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		decKey, err := db.decryptAPIKey(p.APIKey)
		if err != nil {
			return nil, err
		}
		p.APIKey = decKey
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

func (db *DB) UpdateProvider(ctx context.Context, provider *models.Provider) error {
	provider.UpdatedAt = time.Now()

	encKey, err := db.encryptAPIKey(provider.APIKey)
	if err != nil {
		return err
	}

	result, err := db.pool.Exec(ctx,
		"UPDATE providers SET name = $1, type = $2, base_url = $3, api_key = $4, enabled = $5, updated_at = $6 WHERE id = $7",
		provider.Name, provider.Type, provider.BaseURL, encKey, provider.Enabled, provider.UpdatedAt, provider.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrProviderNotFound
	}
	return nil
}

func (db *DB) DeleteProvider(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, "DELETE FROM providers WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrProviderNotFound
	}
	return nil
}

func (db *DB) CreateProviderAccount(ctx context.Context, account *models.ProviderAccount) error {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}
	account.CreatedAt = time.Now()

	encKey, err := db.encryptAPIKey(account.APIKey)
	if err != nil {
		return err
	}

	_, err = db.pool.Exec(ctx,
		"INSERT INTO provider_accounts (id, provider_id, name, api_key, is_active, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		account.ID, account.ProviderID, account.Name, encKey, account.IsActive, account.CreatedAt,
	)
	return err
}

func (db *DB) GetProviderAccountByID(ctx context.Context, id string) (*models.ProviderAccount, error) {
	var account models.ProviderAccount
	err := db.pool.QueryRow(ctx,
		"SELECT id, provider_id, name, api_key, is_active, created_at FROM provider_accounts WHERE id = $1",
		id,
	).Scan(&account.ID, &account.ProviderID, &account.Name, &account.APIKey, &account.IsActive, &account.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProviderAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	decKey, err := db.decryptAPIKey(account.APIKey)
	if err != nil {
		return nil, err
	}
	account.APIKey = decKey

	return &account, nil
}

func (db *DB) ListProviderAccounts(ctx context.Context, providerID string) ([]models.ProviderAccount, error) {
	var rows pgx.Rows
	var err error
	if providerID == "" {
		rows, err = db.pool.Query(ctx,
			"SELECT id, provider_id, name, api_key, is_active, created_at FROM provider_accounts ORDER BY created_at DESC",
		)
	} else {
		rows, err = db.pool.Query(ctx,
			"SELECT id, provider_id, name, api_key, is_active, created_at FROM provider_accounts WHERE provider_id = $1 ORDER BY created_at DESC",
			providerID,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.ProviderAccount
	for rows.Next() {
		var a models.ProviderAccount
		if err := rows.Scan(&a.ID, &a.ProviderID, &a.Name, &a.APIKey, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, err
		}
		decKey, err := db.decryptAPIKey(a.APIKey)
		if err != nil {
			return nil, err
		}
		a.APIKey = decKey
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (db *DB) UpdateProviderAccount(ctx context.Context, account *models.ProviderAccount) error {
	encKey, err := db.encryptAPIKey(account.APIKey)
	if err != nil {
		return err
	}

	result, err := db.pool.Exec(ctx,
		"UPDATE provider_accounts SET name = $1, api_key = $2, is_active = $3 WHERE id = $4",
		account.Name, encKey, account.IsActive, account.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrProviderAccountNotFound
	}
	return nil
}

func (db *DB) DeleteProviderAccount(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, "DELETE FROM provider_accounts WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrProviderAccountNotFound
	}
	return nil
}
