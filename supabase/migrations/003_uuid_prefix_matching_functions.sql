-- Migration: Add UUID prefix matching functions for PostgREST
-- These functions enable LIKE queries on UUID columns by casting to text

-- Function to find content ideas by UUID prefix
CREATE OR REPLACE FUNCTION find_idea_by_prefix(prefix_pattern TEXT)
RETURNS SETOF content_ideas
LANGUAGE sql
STABLE
AS $$
  SELECT *
  FROM content_ideas
  WHERE id::text LIKE prefix_pattern || '%';
$$;

-- Function to find books by UUID prefix
CREATE OR REPLACE FUNCTION find_book_by_prefix(prefix_pattern TEXT)
RETURNS SETOF books
LANGUAGE sql
STABLE
AS $$
  SELECT *
  FROM books
  WHERE id::text LIKE prefix_pattern || '%';
$$;

-- Grant execute permissions to anon and authenticated roles
GRANT EXECUTE ON FUNCTION find_idea_by_prefix(TEXT) TO anon, authenticated;
GRANT EXECUTE ON FUNCTION find_book_by_prefix(TEXT) TO anon, authenticated;

-- Add helpful comment
COMMENT ON FUNCTION find_idea_by_prefix IS 'Find content ideas by UUID prefix (case-sensitive). Example: find_idea_by_prefix(''abcd1234'')';
COMMENT ON FUNCTION find_book_by_prefix IS 'Find books by UUID prefix (case-sensitive). Example: find_book_by_prefix(''abcd1234'')';
