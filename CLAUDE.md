# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gagipress CLI is a social media automation tool for Amazon KDP publishers, built as a Go CLI application using the Cobra framework. It automates content generation, scheduling, and analytics for TikTok and Instagram Reels, with correlation to KDP sales data.

**Tech Stack:**
- **Language**: Go 1.24+
- **CLI Framework**: Cobra + Viper (config)
- **Database**: Supabase (PostgreSQL via REST API + Supabase CLI for migrations)
- **AI Providers**: OpenAI (GPT-4o-mini) + Gemini (browser automation via chromedp)
- **Social APIs**: Instagram Graph API, TikTok Creator API (OAuth setup in progress)

**Current Status**: MVP Phase (Week 5 complete) - Core features working, OAuth publishing pending

## Development Commands

**Always prefer Makefile targets** over raw `mise exec -- go ...` commands. Run `make help` to see all available targets. Use raw Go commands only for ad-hoc operations not covered by the Makefile (e.g., running a specific test with flags).

### Building & Running

```bash
# Build the CLI (output in bin/)
make build

# Run without building (no Makefile target — use raw command)
mise exec -- go run main.go [command]

# Build and install globally
make install
```

### Testing

```bash
# Run all unit tests
make test

# Run with coverage
make test-coverage

# Run only integration tests (requires Supabase credentials)
make test-integration

# Run tests for a specific package (no Makefile target — use raw command)
mise exec -- go test ./internal/models/... -v
mise exec -- go test ./internal/parser/... -v
```

**Note**: `mise exec` is used for Go version management. If not using mise, just use `go test` directly.

### Linting & Code Quality

```bash
# Check for issues (should be run before commits)
make vet

# Format code
make fmt
```

## Architecture & Patterns

### Project Structure

```
cmd/                    # Cobra command definitions (thin layer)
├── root.go            # Main CLI setup
├── books/             # Book management commands
├── generate/          # Content generation commands
├── ideas/             # Idea approval workflow
├── calendar/          # Scheduling commands
└── stats/             # Analytics commands

internal/              # Business logic (testable, reusable)
├── ai/                # OpenAI & Gemini clients
├── config/            # Configuration management
├── errors/            # Custom error types + retry logic
├── generator/         # Content generation logic
├── models/            # Domain models (Book, ContentIdea, etc.)
├── parser/            # KDP CSV parser
├── prompts/           # AI prompt templates
├── repository/        # Database operations (Supabase REST calls)
├── scheduler/         # Calendar planning algorithms
├── social/            # Instagram & TikTok API clients
└── ui/                # CLI UI components (spinners)

test/integration/      # Integration tests (require credentials)
migrations/            # Database migration files (SQL)
supabase/              # Supabase CLI configuration
├── config.toml        # Supabase project config
└── migrations/        # Migrations synced for CLI (auto-generated)
```

### Key Architectural Decisions

1. **Repository Pattern for Database Access**
   - All database operations go through `internal/repository/*`
   - Uses direct HTTP calls to Supabase REST API (not the Go SDK)
   - Each repository handles a single table/domain
   - API key preference: `ServiceKey` (if available) → `AnonKey`

2. **Command-Repository Separation**
   - `cmd/` packages are thin wrappers around Cobra
   - Business logic lives in `internal/`
   - Commands instantiate repositories and call internal packages
   - This makes business logic testable without CLI dependencies

3. **Dual AI Provider Strategy**
   - Primary: OpenAI API (fast, reliable, costs money)
   - Fallback: Gemini browser automation (free, slower, less reliable)
   - Flag `--gemini` forces Gemini usage
   - OpenAI client has exponential backoff retry (3 attempts)

4. **Error Handling with Retry Logic**
   - Custom `AppError` type in `internal/errors/errors.go`
   - Exponential backoff retry in `internal/errors/retry.go`
   - Retries only on server errors (5xx), not client errors (4xx)
   - Context-aware cancellation support

5. **Configuration Management**
   - Config stored in `~/.gagipress/config.yaml`
   - Viper manages config + env var overrides
   - `config.Config` struct in `internal/config/config.go`
   - Required: Supabase URL + anon_key; Optional: API tokens

### Content Generation Workflow

**Flow**: `books` → `ideas` → `scripts` → `calendar` → `publish`

1. **Book Management** (`cmd/books/`)
   - Add/list/edit books in catalog
   - Import KDP sales data from CSV

2. **Idea Generation** (`cmd/generate/ideas.go`)
   - Generates 20+ content ideas using AI
   - Prompts from `internal/prompts/templates.go`
   - Saves to `content_ideas` table with status `pending`

3. **Idea Approval** (`cmd/ideas/`)
   - List pending ideas
   - Approve/reject workflow (updates status)
   - Only approved ideas get scripts generated

4. **Script Generation** (`cmd/generate/script.go`)
   - Converts approved ideas into platform-specific scripts
   - Platform: `instagram` (60s Reels) or `tiktok` (3min max)
   - Saves to `content_ideas` table with generated_script + status `scripted`

5. **Calendar Planning** (`cmd/calendar/plan.go`)
   - Uses `internal/scheduler/planner.go` + `optimizer.go`
   - Creates weekly schedule with optimal posting times
   - Saves to `content_calendar` table with status `scheduled`

6. **Publishing** (Not yet implemented)
   - Will use Supabase Edge Functions (cron jobs)
   - OAuth setup required for Instagram + TikTok

### Database Schema (Supabase)

**Tables:**
- `books` - Book catalog (title, ASIN, genre, target_audience)
- `content_ideas` - Generated ideas with approval status + scripts
- `content_calendar` - Scheduled posts with platform + publish time
- `post_metrics` - Social media engagement data
- `sales_data` - Amazon KDP sales (imported from CSV)

**Status Fields:**
- Content ideas: `pending` → `approved`/`rejected` → `scripted` → `scheduled`
- Calendar: `scheduled` → `published`
- Metrics: `collected` after automated collection

**Migration Tool:**
```bash
gagipress db migrate    # Sync migrations and apply via Supabase CLI
gagipress db status     # Check connection + schema version
```

**Important**: The `migrate` command uses Supabase CLI (`supabase db push`) to apply migrations. Migrations are stored in `migrations/` and synced to `supabase/migrations/` automatically.

### Testing Strategy

**Coverage Targets** (from Week 5 lessons learned):
- Critical code (parsers, error handling): 80%+
- Business logic (scheduler, models): 40-60%
- Commands: Manual testing (not unit tested)

**Testing Patterns:**
- Table-driven tests for validation logic
- Mock at repository layer, not HTTP layer
- Integration tests skip without credentials
- Use `internal/testutil/helpers.go` for common assertions

**Test Organization:**
- Unit tests: `*_test.go` files alongside code
- Integration tests: `test/integration/*_test.go`
- Test utilities: `internal/testutil/`

## Important Constraints

### Config File Format

**IMPORTANT**: All config structs use both `mapstructure` and `yaml` tags:

```go
type SupabaseConfig struct {
    URL        string `mapstructure:"url" yaml:"url"`
    AnonKey    string `mapstructure:"anon_key" yaml:"anon_key"`
    ServiceKey string `mapstructure:"service_key" yaml:"service_key"`
}
```

**Why both tags?**
- `mapstructure`: Used by Viper for deserialization (`Unmarshal`)
- `yaml`: Used by Viper for serialization (`WriteConfigAs`)

Without both tags, field names mismatch between save/load, causing config to appear empty.

**Historical Bug**: Versions before 2026-02-14 only had `mapstructure` tags, causing Viper to write camelCase fields (`anonkey`) but expect snake_case (`anon_key`) when reading. The `fix-config` command migrates old configs.

### Database Migration Strategy

**Supabase CLI for DDL Operations:**
- Database migrations use **Supabase CLI** (`supabase db push`)
- Migration files stored in `migrations/` directory
- Files follow naming convention: `001_description.sql`, `002_description.sql`
- Uses native PostgreSQL functions (`gen_random_uuid()`) instead of extensions

**Why Supabase CLI instead of custom migration system?**
- Official tooling with full PostgreSQL support
- Better error handling and rollback support
- Avoids PostgREST limitations (REST API not designed for DDL)
- Maintains migration history in `supabase_migrations.schema_migrations`

**UUID Generation:**
- Use `gen_random_uuid()` (native PostgreSQL 13+) instead of `uuid_generate_v4()` from `uuid-ossp`
- No extension installation required
- Works out-of-the-box on Supabase

**Supabase CLI Commands:**
```bash
# Apply migrations to remote database
supabase db push

# Generate a schema diff (useful when making manual changes)
supabase db diff

# Check migration status
supabase migration list
```

### Supabase Direct HTTP Usage (for CRUD only)

**Repository pattern uses REST API for data operations:**
- All CRUD operations go through PostgREST API
- Direct HTTP calls (not supabase-go SDK which is experimental)
- Each repository handles a single table/domain

**Pattern to follow:**
```go
// Always prefer ServiceKey over AnonKey
apiKey := r.config.ServiceKey
if apiKey == "" {
    apiKey = r.config.AnonKey
}

// Set required headers
req.Header.Set("apikey", apiKey)
req.Header.Set("Authorization", "Bearer "+apiKey)
req.Header.Set("Prefer", "return=representation")  // For INSERT/UPDATE
```

**Important**: Use REST API only for CRUD operations (SELECT/INSERT/UPDATE/DELETE), not for DDL (CREATE TABLE/ALTER TABLE). Migrations must use Supabase CLI.

### Gemini Browser Automation

**Important**: Gemini integration uses browser automation (chromedp), not an official API
- Selectors like `textarea[placeholder*="Enter a prompt"]` are brittle
- DOM changes will break this code
- Always test with `--headless=false` flag when debugging
- Use as fallback only, not primary AI provider

**Testing Gemini:**
```bash
gagipress test gemini "Hello" --headless=false
```

### CSV Parsing Edge Cases

The KDP CSV parser (`internal/parser/kdp.go`) has been battle-tested with 97% coverage:
- Handles commas in currency fields (`"$1,234.56"`)
- Requires headers: `Title,ASIN,Royalty,Units Sold,Date`
- Date format: `YYYY-MM-DD` or `MM/DD/YYYY`
- Royalty format: `$X.XX` or `X.XX`

**Found bugs** (from testing):
- Comma in currency was treated as field delimiter (fixed)
- Integer division in engagement calculation (fixed)

## Working with AI Prompts

Prompt templates are in `internal/prompts/templates.go`:
- `GenerateIdeasPrompt` - Content idea generation
- `GenerateScriptPrompt` - Script creation (platform-specific)

**When modifying prompts:**
- Test with both OpenAI and Gemini
- OpenAI is faster but Gemini is free
- Add platform-specific instructions (character limits, hashtags)
- Consider target_audience and genre from book metadata

## Common Tasks

### Adding a New Command

1. Create file in `cmd/<category>/`
2. Define Cobra command with flags
3. Add to parent command in `init()`
4. Implement logic in `internal/` package
5. Add tests in `internal/` (not `cmd/`)

### Adding a New Repository Method

1. Define method in `internal/repository/<domain>.go`
2. Use HTTP client pattern (see existing methods)
3. Handle Supabase-specific headers
4. Parse JSON response into domain model
5. Add error handling (wrap with context)

### Debugging Failed Tests

These are ad-hoc commands with specific flags — no Makefile targets for these:

```bash
# Run specific test with verbose output
mise exec -- go test ./internal/parser -run TestParseKDPReport -v

# Run with coverage to see what's not tested
mise exec -- go test ./internal/parser -coverprofile=coverage.out
mise exec -- go tool cover -html=coverage.out
```

## Future Development Notes

**OAuth Publishing** (deferred to Phase 2):
- Instagram requires Facebook Developer App + OAuth flow
- TikTok requires Creator API access + OAuth
- Will use Supabase Edge Functions for cron-based publishing

**Video Automation** (Phase 2):
- FFmpeg compositing for template-based videos
- Text-to-speech voiceover integration
- Auto-subtitles generation

**Performance Considerations**:
- AI generation is rate-limited by API quotas
- Database queries are simple (no complex joins needed yet)
- Parser handles large CSV files efficiently (streaming not needed yet)

## References

- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Supabase REST API Docs](https://supabase.com/docs/guides/api)
- [OpenAI Chat Completions](https://platform.openai.com/docs/guides/chat)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)
