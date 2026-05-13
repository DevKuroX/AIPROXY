package storage

import (
	"context"
	"errors"
	"time"
)

var ErrSettingNotFound = errors.New("setting not found")

func (db *DB) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := db.pool.QueryRow(ctx,
		"SELECT value FROM settings WHERE key = $1",
		key,
	).Scan(&value)

	if errors.Is(err, ErrSettingNotFound) {
		return "", ErrSettingNotFound
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (db *DB) SetSetting(ctx context.Context, key, value string) error {
	_, err := db.pool.Exec(ctx,
		"INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, $3) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = $3",
		key, value, time.Now(),
	)
	return err
}

func (db *DB) DeleteSetting(ctx context.Context, key string) error {
	_, err := db.pool.Exec(ctx,
		"DELETE FROM settings WHERE key = $1",
		key,
	)
	return err
}
