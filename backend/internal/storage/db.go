package storage

import (
	"context"

	"github.com/DevKuroX/AIPROXY/internal/auth/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool           *pgxpool.Pool
	encryptionKey  []byte // AES-256-GCM key (32 bytes). If nil, pass-through (no encryption).
}

func New(ctx context.Context, dbURL string, encryptionKey []byte) (*DB, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &DB{pool: pool, encryptionKey: encryptionKey}, nil
}

// encryptAPIKey encrypts apiKey with the DB encryption key.
// Returns original string unchanged if key is nil or apiKey is empty.
func (db *DB) encryptAPIKey(apiKey string) (string, error) {
	if db.encryptionKey == nil || apiKey == "" {
		return apiKey, nil
	}
	enc, err := crypto.Encrypt([]byte(apiKey), db.encryptionKey)
	if err != nil {
		return "", err
	}
	return string(enc), nil
}

// decryptAPIKey decrypts apiKey with the DB encryption key.
func (db *DB) decryptAPIKey(apiKey string) (string, error) {
	if db.encryptionKey == nil || apiKey == "" {
		return apiKey, nil
	}
	dec, err := crypto.Decrypt([]byte(apiKey), db.encryptionKey)
	if err != nil {
		return apiKey, nil
	}
	return string(dec), nil
}

func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}
