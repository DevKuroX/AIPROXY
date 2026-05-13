package storage

import (
	"context"
	"errors"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/jackc/pgx/v5"
)

var ErrUserNotFound = errors.New("user not found")

func (db *DB) CreateUser(ctx context.Context, username, passwordHash string, isAdmin bool) error {
	_, err := db.pool.Exec(ctx,
		"INSERT INTO users (username, password_hash, is_admin) VALUES ($1, $2, $3)",
		username, passwordHash, isAdmin,
	)
	return err
}

func (db *DB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := db.pool.QueryRow(ctx,
		"SELECT id, username, password_hash, is_admin, created_at, updated_at FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := db.pool.QueryRow(ctx,
		"SELECT id, username, password_hash, is_admin, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	_, err := db.pool.Exec(ctx,
		"UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3",
		passwordHash, time.Now(), id,
	)
	return err
}
