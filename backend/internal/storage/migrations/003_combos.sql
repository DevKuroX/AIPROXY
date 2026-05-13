-- Combos table for model combos with fallback/round-robin strategies
-- ref: open-sse/services/combo.js:8-12

CREATE TABLE IF NOT EXISTS combos (
    name VARCHAR(255) PRIMARY KEY,
    models JSONB NOT NULL,
    strategy VARCHAR(50) NOT NULL DEFAULT 'fallback',
    sticky_limit INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add account_status columns to provider_accounts if not exists
-- ref: open-sse/services/combo.js (account health tracking)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'provider_accounts' AND column_name = 'status') THEN
        ALTER TABLE provider_accounts ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'active';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'provider_accounts' AND column_name = 'status_changed_at') THEN
        ALTER TABLE provider_accounts ADD COLUMN status_changed_at TIMESTAMP WITH TIME ZONE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'provider_accounts' AND column_name = 'retry_after') THEN
        ALTER TABLE provider_accounts ADD COLUMN retry_after TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;
