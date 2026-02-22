-- Migration 005: Enable pg_cron and schedule publish-scheduled edge function
-- Date: 2026-02-22
-- NOTE: pg_cron and pg_net must be enabled in Supabase Dashboard first:
--   Dashboard → Database → Extensions → enable pg_cron and pg_net

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pg_cron;
CREATE EXTENSION IF NOT EXISTS pg_net;

-- This migration is idempotent (safe to re-run)

-- Schedule the publish-scheduled edge function every 15 minutes.
-- The function is deployed with --no-verify-jwt so no auth header needed.
-- It uses SUPABASE_SERVICE_ROLE_KEY (auto-injected by runtime) for DB access.
SELECT cron.schedule(
    'gagipress-publish-scheduled',
    '*/15 * * * *',
    $$
    SELECT net.http_post(
        url     := 'https://nhsthucdmjgodfrduzfb.supabase.co/functions/v1/publish-scheduled',
        headers := '{"Content-Type": "application/json"}'::jsonb,
        body    := '{}'::jsonb
    );
    $$
);

-- Schema version bump
INSERT INTO schema_version (version, description)
VALUES (5, 'Schedule publish-scheduled edge function via pg_cron every 15 minutes');
