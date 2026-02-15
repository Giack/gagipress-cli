-- {{DESCRIPTION}}
-- Created: {{DATE}}

CREATE OR REPLACE FUNCTION {{FUNCTION_NAME}}(
    -- Add parameters here
    -- p_param_name TYPE
)
RETURNS {{RETURN_TYPE}} AS $$
BEGIN
    -- Add function logic here

    -- Example for trigger function:
    -- NEW.updated_at = NOW();
    -- RETURN NEW;

    -- Example for query function:
    -- RETURN QUERY
    -- SELECT * FROM table_name WHERE condition;

END;
$$ LANGUAGE plpgsql;

-- For trigger functions, create trigger:
-- CREATE TRIGGER {{TRIGGER_NAME}}
--     BEFORE UPDATE ON {{TABLE_NAME}}
--     FOR EACH ROW
--     EXECUTE FUNCTION {{FUNCTION_NAME}}();

COMMENT ON FUNCTION {{FUNCTION_NAME}} IS '{{DESCRIPTION}}';
