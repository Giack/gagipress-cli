-- Gagipress Initial Database Schema
-- Created: 2026-02-08
-- Description: Core tables for books, content, calendar, metrics, and sales

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Books Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS books (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title TEXT NOT NULL,
  genre TEXT NOT NULL,
  target_audience TEXT,
  kdp_asin TEXT UNIQUE,
  cover_image_url TEXT,
  publication_date DATE,
  current_rank INTEGER,
  total_sales INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_books_genre ON books(genre);
CREATE INDEX idx_books_asin ON books(kdp_asin);

-- ============================================================================
-- Content Ideas Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS content_ideas (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  type TEXT NOT NULL CHECK (type IN ('educational', 'entertainment', 'bts', 'ugc', 'trend')),
  brief_description TEXT NOT NULL,
  relevance_score INTEGER CHECK (relevance_score >= 0 AND relevance_score <= 100),
  book_id UUID REFERENCES books(id) ON DELETE CASCADE,
  status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'scripted')),
  generated_at TIMESTAMPTZ DEFAULT NOW(),
  metadata JSONB
);

CREATE INDEX idx_ideas_status ON content_ideas(status);
CREATE INDEX idx_ideas_score ON content_ideas(relevance_score DESC);
CREATE INDEX idx_ideas_book ON content_ideas(book_id);
CREATE INDEX idx_ideas_type ON content_ideas(type);

-- ============================================================================
-- Content Scripts Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS content_scripts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  idea_id UUID REFERENCES content_ideas(id) ON DELETE CASCADE,
  hook TEXT NOT NULL,
  full_script TEXT NOT NULL,
  cta TEXT NOT NULL,
  hashtags TEXT[] NOT NULL,
  visual_notes TEXT,
  audio_suggestion TEXT,
  estimated_duration INTEGER, -- seconds
  status TEXT DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'used')),
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_scripts_idea ON content_scripts(idea_id);
CREATE INDEX idx_scripts_status ON content_scripts(status);

-- ============================================================================
-- Content Calendar Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS content_calendar (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  script_id UUID REFERENCES content_scripts(id) ON DELETE CASCADE,
  scheduled_for TIMESTAMPTZ NOT NULL,
  platform TEXT NOT NULL CHECK (platform IN ('instagram', 'tiktok')),
  post_type TEXT NOT NULL CHECK (post_type IN ('reel', 'story', 'feed')),
  status TEXT DEFAULT 'pending_approval' CHECK (status IN ('pending_approval', 'approved', 'published', 'failed')),
  approved_at TIMESTAMPTZ,
  published_at TIMESTAMPTZ,
  post_url TEXT,
  publish_errors JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_calendar_scheduled ON content_calendar(scheduled_for);
CREATE INDEX idx_calendar_status ON content_calendar(status);
CREATE INDEX idx_calendar_platform ON content_calendar(platform);
CREATE INDEX idx_calendar_script ON content_calendar(script_id);

-- ============================================================================
-- Post Metrics Table (Time-Series)
-- ============================================================================
CREATE TABLE IF NOT EXISTS post_metrics (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  calendar_id UUID REFERENCES content_calendar(id) ON DELETE CASCADE,
  platform TEXT NOT NULL,
  post_url TEXT,
  views INTEGER DEFAULT 0,
  likes INTEGER DEFAULT 0,
  comments INTEGER DEFAULT 0,
  shares INTEGER DEFAULT 0,
  saves INTEGER DEFAULT 0,
  engagement_rate DECIMAL(5,2),
  watch_time_percentage DECIMAL(5,2),
  follower_growth INTEGER DEFAULT 0,
  is_top_performer BOOLEAN DEFAULT FALSE,
  scraped_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_metrics_calendar ON post_metrics(calendar_id);
CREATE INDEX idx_metrics_scraped ON post_metrics(scraped_at DESC);
CREATE INDEX idx_metrics_top ON post_metrics(is_top_performer) WHERE is_top_performer = TRUE;

-- ============================================================================
-- Sales Data Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS sales_data (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  book_id UUID REFERENCES books(id) ON DELETE CASCADE,
  date DATE NOT NULL,
  units_sold INTEGER DEFAULT 0,
  revenue DECIMAL(10,2),
  royalty DECIMAL(10,2),
  source TEXT DEFAULT 'amazon_reports',
  imported_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(book_id, date)
);

CREATE INDEX idx_sales_book ON sales_data(book_id);
CREATE INDEX idx_sales_date ON sales_data(date DESC);

-- ============================================================================
-- Updated At Trigger Function
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to books table
CREATE TRIGGER update_books_updated_at
    BEFORE UPDATE ON books
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Row Level Security (RLS) Setup
-- ============================================================================
-- Enable RLS on all tables
ALTER TABLE books ENABLE ROW LEVEL SECURITY;
ALTER TABLE content_ideas ENABLE ROW LEVEL SECURITY;
ALTER TABLE content_scripts ENABLE ROW LEVEL SECURITY;
ALTER TABLE content_calendar ENABLE ROW LEVEL SECURITY;
ALTER TABLE post_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE sales_data ENABLE ROW LEVEL SECURITY;

-- Create policies (allow all for authenticated users for now)
-- In production, you'd want more granular policies

CREATE POLICY "Enable all access for authenticated users" ON books
    FOR ALL USING (true);

CREATE POLICY "Enable all access for authenticated users" ON content_ideas
    FOR ALL USING (true);

CREATE POLICY "Enable all access for authenticated users" ON content_scripts
    FOR ALL USING (true);

CREATE POLICY "Enable all access for authenticated users" ON content_calendar
    FOR ALL USING (true);

CREATE POLICY "Enable all access for authenticated users" ON post_metrics
    FOR ALL USING (true);

CREATE POLICY "Enable all access for authenticated users" ON sales_data
    FOR ALL USING (true);

-- ============================================================================
-- Utility Views
-- ============================================================================

-- View: Complete content pipeline
CREATE OR REPLACE VIEW content_pipeline AS
SELECT
    ci.id as idea_id,
    ci.type as content_type,
    ci.brief_description,
    ci.relevance_score,
    ci.status as idea_status,
    b.title as book_title,
    b.genre,
    cs.id as script_id,
    cs.hook,
    cs.status as script_status,
    cc.id as calendar_id,
    cc.scheduled_for,
    cc.platform,
    cc.status as publish_status,
    cc.post_url,
    pm.views,
    pm.engagement_rate
FROM content_ideas ci
LEFT JOIN books b ON ci.book_id = b.id
LEFT JOIN content_scripts cs ON cs.idea_id = ci.id
LEFT JOIN content_calendar cc ON cc.script_id = cs.id
LEFT JOIN post_metrics pm ON pm.calendar_id = cc.id;

-- View: Performance summary by content type
CREATE OR REPLACE VIEW performance_by_type AS
SELECT
    ci.type,
    COUNT(DISTINCT cc.id) as total_posts,
    AVG(pm.views) as avg_views,
    AVG(pm.engagement_rate) as avg_engagement,
    SUM(CASE WHEN pm.is_top_performer THEN 1 ELSE 0 END) as top_performers
FROM content_ideas ci
JOIN content_scripts cs ON cs.idea_id = ci.id
JOIN content_calendar cc ON cc.script_id = cs.id
JOIN post_metrics pm ON pm.calendar_id = cc.id
WHERE cc.status = 'published'
GROUP BY ci.type;

-- View: Book performance with social correlation
CREATE OR REPLACE VIEW book_social_performance AS
SELECT
    b.id,
    b.title,
    b.genre,
    COUNT(DISTINCT cc.id) as posts_count,
    AVG(pm.views) as avg_views,
    AVG(pm.engagement_rate) as avg_engagement,
    SUM(sd.units_sold) as total_sales_last_30d
FROM books b
LEFT JOIN content_ideas ci ON ci.book_id = b.id
LEFT JOIN content_scripts cs ON cs.idea_id = ci.id
LEFT JOIN content_calendar cc ON cc.script_id = cs.id
LEFT JOIN post_metrics pm ON pm.calendar_id = cc.id
LEFT JOIN sales_data sd ON sd.book_id = b.id
    AND sd.date >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY b.id, b.title, b.genre;

-- ============================================================================
-- Helper Functions
-- ============================================================================

-- Function: Calculate engagement rate
CREATE OR REPLACE FUNCTION calculate_engagement_rate(
    p_likes INTEGER,
    p_comments INTEGER,
    p_shares INTEGER,
    p_views INTEGER
)
RETURNS DECIMAL(5,2) AS $$
BEGIN
    IF p_views = 0 THEN
        RETURN 0;
    END IF;
    RETURN ROUND(((p_likes + p_comments + p_shares)::DECIMAL / p_views * 100), 2);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function: Get top performing content for a book
CREATE OR REPLACE FUNCTION get_top_content_for_book(p_book_id UUID, p_limit INTEGER DEFAULT 5)
RETURNS TABLE (
    script_id UUID,
    hook TEXT,
    platform TEXT,
    views INTEGER,
    engagement_rate DECIMAL(5,2),
    post_url TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cs.id,
        cs.hook,
        cc.platform,
        pm.views,
        pm.engagement_rate,
        cc.post_url
    FROM content_ideas ci
    JOIN content_scripts cs ON cs.idea_id = ci.id
    JOIN content_calendar cc ON cc.script_id = cs.id
    JOIN post_metrics pm ON pm.calendar_id = cc.id
    WHERE ci.book_id = p_book_id
        AND cc.status = 'published'
        AND pm.is_top_performer = true
    ORDER BY pm.engagement_rate DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Initial Data
-- ============================================================================

-- Insert a sample configuration row (optional, for testing)
-- You can remove this in production

COMMENT ON TABLE books IS 'Catalog of Amazon KDP books being promoted';
COMMENT ON TABLE content_ideas IS 'AI-generated content ideas for social media posts';
COMMENT ON TABLE content_scripts IS 'Complete scripts ready for production';
COMMENT ON TABLE content_calendar IS 'Scheduled posts with approval workflow';
COMMENT ON TABLE post_metrics IS 'Time-series performance metrics from social platforms';
COMMENT ON TABLE sales_data IS 'Amazon KDP sales data for correlation analysis';

-- ============================================================================
-- Schema Version
-- ============================================================================
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW(),
    description TEXT
);

INSERT INTO schema_version (version, description)
VALUES (1, 'Initial schema with books, content pipeline, and analytics tables');
