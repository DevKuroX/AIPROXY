package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dbURL string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}
