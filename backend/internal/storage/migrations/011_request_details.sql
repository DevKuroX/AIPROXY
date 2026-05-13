-- Migration 011: Create request_details table
-- ref: open-sse/handlers/chatCore/requestDetail.js

CREATE TABLE IF NOT EXISTS request_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    headers JSONB,
    body JSONB,
    response JSONB,
    status_code INT,
    duration_ms INT,
    error TEXT,
    provider_id BIGINT REFERENCES providers(id),
    account_id BIGINT,
    model VARCHAR(255),
    tokens_prompt INT,
    tokens_completion INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_details_timestamp ON request_details(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_request_details_provider ON request_details(provider_id);
CREATE INDEX IF NOT EXISTS idx_request_details_model ON request_details(model);
CREATE INDEX IF NOT EXISTS idx_request_details_created_at ON request_details(created_at);

-- Add comment
COMMENT ON TABLE request_details IS 'Stores detailed request logs for debugging and analytics';
COMMENT ON COLUMN request_details.duration_ms IS 'Request duration in milliseconds';
COMMENT ON COLUMN request_details.tokens_prompt IS 'Prompt tokens used';
COMMENT ON COLUMN request_details.tokens_completion IS 'Completion tokens used';
