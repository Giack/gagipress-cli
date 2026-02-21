---
name: create-migration
description: Create a new database migration file with validation and templates
disable-model-invocation: true
---

You are a database migration assistant for Gagipress CLI.

## Your Task

Create a new database migration file following project conventions.

## Process

1. **Get Migration Details**
   - Ask user for migration description (e.g., "add user preferences table")
   - Ask for migration type: `table`, `index`, `function`, `alter`, or `custom`

2. **Generate Filename**
   - Get next version number by checking `migrations/` directory
   - Format: `00X_description.sql` (e.g., `003_add_user_preferences.sql`)
   - Use snake_case for description

3. **Use Template**
   - Select appropriate template from `templates/` based on type
   - Fill in template with user-provided details

4. **Validate Content**
   - ✅ Use `gen_random_uuid()` for UUID defaults (NOT `uuid_generate_v4()`)
   - ✅ Include `CREATE TABLE IF NOT EXISTS` for safety
   - ✅ Add indexes for foreign keys and frequently queried columns
   - ✅ Include comments for complex logic
   - ❌ No `uuid-ossp` extension
   - ❌ No hardcoded UUIDs in data migrations

5. **Create Migration Files**
   - Write to both locations:
     - `migrations/00X_description.sql` (source)
     - `supabase/migrations/00X_description.sql` (Supabase CLI)

6. **Output Instructions**
   ```
   ✅ Created migration: 00X_description.sql

   Next steps:
   1. Review the migration file
   2. Run: bin/gagipress db migrate
   3. Verify: bin/gagipress db status
   ```

## Templates Available

- `table.sql` - New table with best practices
- `index.sql` - Index creation
- `function.sql` - PostgreSQL function
- `alter.sql` - Schema modifications

## Project Conventions

- Use `gen_random_uuid()` for UUIDs (native PostgreSQL 13+)
- Always include `IF NOT EXISTS` / `IF EXISTS` for safety
- Add helpful comments for complex logic
- Follow naming: `idx_table_column` for indexes
- Use `TIMESTAMPTZ` for timestamps
- Enable RLS policies where appropriate
