-- Migration 004: OAuth Tokens
-- Stores encrypted OAuth tokens for provider accounts

CREATE TABLE IF NOT EXISTS oauth_tokens (
    provider_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36) NOT NULL,
    encrypted_access_token TEXT NOT NULL,
    encrypted_refresh_token TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (provider_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_oauth_tokens_expires_at ON oauth_tokens(expires_at);
