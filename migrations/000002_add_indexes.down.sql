-- Remove indexes created in 000002_add_indexes.up.sql
DROP INDEX IF EXISTS idx_urls_cleanup;
DROP INDEX IF EXISTS idx_urls_active;
DROP INDEX IF EXISTS idx_urls_clicks;
DROP INDEX IF EXISTS idx_urls_updated_at;
DROP INDEX IF EXISTS idx_urls_created_at;
DROP INDEX IF EXISTS idx_urls_short_code;
