-- Migration 001: Initialize schema
-- Creates users and settings tables with default admin user

-- Users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Settings table for key-value configuration
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Migration tracking table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for faster username lookups
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Insert default admin user (password: admin123)
-- bcrypt hash generated with cost 10
INSERT INTO users (username, password_hash, is_admin)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMy.Mrj5HuGQX5.J4UzOJX5qK8YJ5xG9F8K', TRUE)
ON CONFLICT (username) DO NOTHING;

-- Record this migration
INSERT INTO schema_migrations (version) VALUES ('001') ON CONFLICT DO NOTHING;
