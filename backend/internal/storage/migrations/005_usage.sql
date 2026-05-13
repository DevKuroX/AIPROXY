-- Migration 005: Usage Log and Pricing
-- Creates tables for usage tracking and model pricing

-- Usage log table (append-only for auditing and analytics)
CREATE TABLE IF NOT EXISTS usage_log (
    id VARCHAR(36) PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    model VARCHAR(255) NOT NULL,
    provider_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(36),
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    cost_usd DECIMAL(18, 8) NOT NULL DEFAULT 0,
    rtk_bytes_saved INTEGER NOT NULL DEFAULT 0,
    caveman_active BOOLEAN NOT NULL DEFAULT FALSE,
    api_key_id VARCHAR(36),
    duration_ms INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message TEXT
);

-- Pricing table for model cost calculations
CREATE TABLE IF NOT EXISTS pricing (
    model_pattern VARCHAR(255) PRIMARY KEY,
    prompt_price_per_1m DECIMAL(18, 8) NOT NULL DEFAULT 0,
    completion_price_per_1m DECIMAL(18, 8) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for usage_log lookups
CREATE INDEX IF NOT EXISTS idx_usage_log_timestamp ON usage_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_usage_log_model ON usage_log(model);
CREATE INDEX IF NOT EXISTS idx_usage_log_provider_id ON usage_log(provider_id);
CREATE INDEX IF NOT EXISTS idx_usage_log_status ON usage_log(status);
CREATE INDEX IF NOT EXISTS idx_usage_log_api_key_id ON usage_log(api_key_id);

-- Index for pricing pattern lookups
CREATE INDEX IF NOT EXISTS idx_pricing_model_pattern ON pricing(model_pattern);
