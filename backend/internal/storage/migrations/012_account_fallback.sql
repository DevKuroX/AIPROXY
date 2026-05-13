-- Migration 012: Account Fallback State
-- Creates tables for tracking account fallback state and cooldown periods

-- Account fallback state table
-- Tracks per-account cooldown and backoff state for intelligent fallback
CREATE TABLE IF NOT EXISTS account_fallback_state (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    provider_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    backoff_level INTEGER NOT NULL DEFAULT 0,
    unavailable_until TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    last_error_status INTEGER,
    last_failure_at TIMESTAMP WITH TIME ZONE,
    total_failures INTEGER NOT NULL DEFAULT 0,
    total_successes INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(provider_id, account_id)
);

-- Indexes for fallback lookups
CREATE INDEX IF NOT EXISTS idx_account_fallback_provider ON account_fallback_state(provider_id);
CREATE INDEX IF NOT EXISTS idx_account_fallback_account ON account_fallback_state(account_id);
CREATE INDEX IF NOT EXISTS idx_account_fallback_unavailable ON account_fallback_state(unavailable_until) 
    WHERE unavailable_until IS NOT NULL;

-- Account health metrics table
-- Tracks rolling success/failure rates for smarter fallback decisions
CREATE TABLE IF NOT EXISTS account_health_metrics (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    provider_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    window_end TIMESTAMP WITH TIME ZONE NOT NULL,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    avg_latency_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(provider_id, account_id, window_start)
);

CREATE INDEX IF NOT EXISTS idx_health_metrics_provider ON account_health_metrics(provider_id);
CREATE INDEX IF NOT EXISTS idx_health_metrics_window ON account_health_metrics(window_start, window_end);

-- Combo rotation state table
-- Persists round-robin rotation state across restarts
CREATE TABLE IF NOT EXISTS combo_rotation_state (
    combo_name TEXT PRIMARY KEY,
    current_index INTEGER NOT NULL DEFAULT 0,
    consecutive_use_count INTEGER NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
