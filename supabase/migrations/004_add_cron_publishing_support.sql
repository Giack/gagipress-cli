-- Migration 004: Add cron publishing support
-- Adds: generate_media flag, updated_at for stale lock detection,
--       and 'publishing' as a transient lock status.
-- Date: 2026-02-22

-- 1. Add updated_at to content_calendar (needed to detect stale 'publishing' locks)
ALTER TABLE content_calendar
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Backfill updated_at from created_at for existing rows
UPDATE content_calendar SET updated_at = created_at WHERE updated_at IS NULL;

-- Create trigger to auto-update updated_at on every row change
CREATE OR REPLACE FUNCTION update_calendar_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_content_calendar_updated_at
    BEFORE UPDATE ON content_calendar
    FOR EACH ROW
    EXECUTE FUNCTION update_calendar_updated_at();

-- 2. Add generate_media flag (default false = text-only publish)
ALTER TABLE content_calendar
  ADD COLUMN IF NOT EXISTS generate_media BOOLEAN NOT NULL DEFAULT FALSE;

-- 3. Extend the status CHECK constraint to include 'publishing' (transient lock)
ALTER TABLE content_calendar DROP CONSTRAINT IF EXISTS content_calendar_status_check;
ALTER TABLE content_calendar
  ADD CONSTRAINT content_calendar_status_check
  CHECK (status IN ('pending_approval', 'approved', 'publishing', 'published', 'failed'));

-- 4. Add an index on (status, scheduled_for) to make the cron query fast
CREATE INDEX IF NOT EXISTS idx_calendar_status_scheduled
  ON content_calendar(status, scheduled_for)
  WHERE status IN ('approved', 'publishing');

-- Schema version bump
INSERT INTO schema_version (version, description)
VALUES (4, 'Add updated_at, generate_media, and publishing lock status to content_calendar');
