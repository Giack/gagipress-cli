-- {{DESCRIPTION}}
-- Created: {{DATE}}

CREATE TABLE IF NOT EXISTS {{TABLE_NAME}} (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Add your columns here
  -- Example:
  -- name TEXT NOT NULL,
  -- email TEXT UNIQUE,
  -- status TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive')),

  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
-- CREATE INDEX idx_{{TABLE_NAME}}_{{COLUMN}} ON {{TABLE_NAME}}({{COLUMN}});

-- Enable Row Level Security (if needed)
-- ALTER TABLE {{TABLE_NAME}} ENABLE ROW LEVEL SECURITY;

-- Create RLS policies
-- CREATE POLICY "Enable all access for authenticated users" ON {{TABLE_NAME}}
--     FOR ALL USING (true);

-- Comments
COMMENT ON TABLE {{TABLE_NAME}} IS '{{DESCRIPTION}}';
