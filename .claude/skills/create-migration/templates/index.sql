-- {{DESCRIPTION}}
-- Created: {{DATE}}

-- Create index for {{TABLE_NAME}}
CREATE INDEX IF NOT EXISTS idx_{{TABLE_NAME}}_{{COLUMN}}
ON {{TABLE_NAME}}({{COLUMN}});

-- For composite indexes:
-- CREATE INDEX IF NOT EXISTS idx_{{TABLE_NAME}}_{{COLUMN1}}_{{COLUMN2}}
-- ON {{TABLE_NAME}}({{COLUMN1}}, {{COLUMN2}});

-- For partial/conditional indexes:
-- CREATE INDEX IF NOT EXISTS idx_{{TABLE_NAME}}_{{COLUMN}}_active
-- ON {{TABLE_NAME}}({{COLUMN}}) WHERE status = 'active';

-- For descending order (useful for ORDER BY ... DESC queries):
-- CREATE INDEX IF NOT EXISTS idx_{{TABLE_NAME}}_{{COLUMN}}_desc
-- ON {{TABLE_NAME}}({{COLUMN}} DESC);
