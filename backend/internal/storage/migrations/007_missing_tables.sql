-- Migration 007: Add missing tables (proxy_pools, kv, usage_daily, request_details)

-- Proxy Pools table
CREATE TABLE IF NOT EXISTS proxy_pools (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    test_status TEXT,
    proxies JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_proxy_pools_active ON proxy_pools(is_active);
CREATE INDEX IF NOT EXISTS idx_proxy_pools_status ON proxy_pools(test_status);

-- Key-Value store table
CREATE TABLE IF NOT EXISTS kv (
    scope TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (scope, key)
);

CREATE INDEX IF NOT EXISTS idx_kv_scope ON kv(scope);

-- Usage Daily aggregation table
CREATE TABLE IF NOT EXISTS usage_daily (
    date_key TEXT PRIMARY KEY,
    data JSONB NOT NULL
);

-- Request Details table (full request/response logs)
CREATE TABLE IF NOT EXISTS request_details (
    id TEXT PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    provider TEXT,
    model TEXT,
    connection_id TEXT,
    status TEXT,
    data JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_request_details_ts ON request_details(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_request_details_provider ON request_details(provider);
CREATE INDEX IF NOT EXISTS idx_request_details_model ON request_details(model);
CREATE INDEX IF NOT EXISTS idx_request_details_conn ON request_details(connection_id);
