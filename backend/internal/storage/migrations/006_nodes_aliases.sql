-- Migration 006: Provider Nodes and Model Aliases
-- Creates tables for custom 0penAI-compatible endpoints

CREATE TABLE IF NOT EXISTS provider_nodes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    base_url VARCHAR(500) NOT NULL,
    api_key VARCHAR(500) NOT NULL,
    compatible_format VARCHAR(50) NOT NULL DEFAULT 'openai',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS model_aliases (
    id VARCHAR(36) PRIMARY KEY,
    node_id VARCHAR(36) NOT NULL REFERENCES provider_nodes(id) ON DELETE CASCADE,
    alias VARCHAR(255) NOT NULL,
    target_model VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(alias)
);

CREATE INDEX IF NOT EXISTS idx_provider_nodes_enabled ON provider_nodes(enabled);
CREATE INDEX IF NOT EXISTS idx_model_aliases_alias ON model_aliases(alias);
CREATE INDEX IF NOT EXISTS idx_model_aliases_node_id ON model_aliases(node_id);
