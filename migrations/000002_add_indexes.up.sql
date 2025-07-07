-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at);
CREATE INDEX IF NOT EXISTS idx_urls_updated_at ON urls(updated_at);
CREATE INDEX IF NOT EXISTS idx_urls_clicks ON urls(clicks);

-- Add partial index for active URLs (with clicks > 0)
CREATE INDEX IF NOT EXISTS idx_urls_active ON urls(short_code, clicks) WHERE clicks > 0;

-- Add index for cleanup operations (finding old unused URLs)
CREATE INDEX IF NOT EXISTS idx_urls_cleanup ON urls(created_at, clicks) WHERE clicks = 0;
