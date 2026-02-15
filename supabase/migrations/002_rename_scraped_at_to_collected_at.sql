-- Rename scraped_at to collected_at in post_metrics table
-- This better reflects the use case: metrics can be manually collected or scraped

ALTER TABLE post_metrics
RENAME COLUMN scraped_at TO collected_at;

-- Update the index
DROP INDEX IF EXISTS idx_metrics_scraped;
CREATE INDEX idx_metrics_collected ON post_metrics(collected_at DESC);

-- Update schema version
INSERT INTO schema_version (version, description)
VALUES (2, 'Rename scraped_at to collected_at in post_metrics table');
