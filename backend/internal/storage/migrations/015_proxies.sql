CREATE TABLE IF NOT EXISTS proxies (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    protocol TEXT,
    host TEXT,
    port TEXT,
    region TEXT,
    latency_ms INTEGER DEFAULT 0,
    status TEXT DEFAULT 'untested',
    source TEXT,
    success_rate REAL DEFAULT 0,
    fail_count INTEGER DEFAULT 0,
    last_checked TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Ensure UNIQUE on url for ON CONFLICT (url) to work
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='proxies_url_key') THEN
        ALTER TABLE proxies ADD CONSTRAINT proxies_url_key UNIQUE (url);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_proxies_status ON proxies(status);
CREATE INDEX IF NOT EXISTS idx_proxies_latency ON proxies(latency_ms);
