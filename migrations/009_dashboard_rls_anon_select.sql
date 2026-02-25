-- Migration 009: Dashboard RLS — SELECT-only anon policies
--
-- Purpose: Replace over-permissive FOR ALL policies with:
--   1. Full access for the authenticated role (required for CLI operations)
--   2. SELECT-only access for the anon role (dashboard read-only access)
--
-- Tables covered: books, content_ideas, content_scripts,
--                 content_calendar, post_metrics, sales_data
--
-- Apply with: supabase db push

-- ============================================================
-- Drop existing FOR ALL policies (allow anon INSERT/UPDATE/DELETE)
-- ============================================================

DROP POLICY IF EXISTS "Enable all access for authenticated users" ON books;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON content_ideas;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON content_scripts;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON content_calendar;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON post_metrics;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON sales_data;

-- ============================================================
-- Re-create full access for the authenticated role
-- ============================================================

CREATE POLICY "authenticated full access" ON books
  FOR ALL TO authenticated USING (true) WITH CHECK (true);

CREATE POLICY "authenticated full access" ON content_ideas
  FOR ALL TO authenticated USING (true) WITH CHECK (true);

CREATE POLICY "authenticated full access" ON content_scripts
  FOR ALL TO authenticated USING (true) WITH CHECK (true);

CREATE POLICY "authenticated full access" ON content_calendar
  FOR ALL TO authenticated USING (true) WITH CHECK (true);

CREATE POLICY "authenticated full access" ON post_metrics
  FOR ALL TO authenticated USING (true) WITH CHECK (true);

CREATE POLICY "authenticated full access" ON sales_data
  FOR ALL TO authenticated USING (true) WITH CHECK (true);

-- ============================================================
-- Add SELECT-only policies for the anon role (dashboard reads)
-- ============================================================

CREATE POLICY "anon read-only" ON books
  FOR SELECT TO anon USING (true);

CREATE POLICY "anon read-only" ON content_ideas
  FOR SELECT TO anon USING (true);

CREATE POLICY "anon read-only" ON content_scripts
  FOR SELECT TO anon USING (true);

CREATE POLICY "anon read-only" ON content_calendar
  FOR SELECT TO anon USING (true);

CREATE POLICY "anon read-only" ON post_metrics
  FOR SELECT TO anon USING (true);

CREATE POLICY "anon read-only" ON sales_data
  FOR SELECT TO anon USING (true);
