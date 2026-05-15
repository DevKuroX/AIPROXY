-- Migration 013: Add Kiro token fields to provider_accounts
-- Adds columns needed for OAuth token refresh and account pool management

ALTER TABLE provider_accounts ADD COLUMN IF NOT EXISTS refresh_token TEXT;
ALTER TABLE provider_accounts ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE provider_accounts ADD COLUMN IF NOT EXISTS profile_arn TEXT;
ALTER TABLE provider_accounts ADD COLUMN IF NOT EXISTS backoff_level INTEGER DEFAULT 0;
ALTER TABLE provider_accounts ADD COLUMN IF NOT EXISTS unavailable_until TIMESTAMP WITH TIME ZONE;
