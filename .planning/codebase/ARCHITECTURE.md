# Architecture

**Analysis Date:** 2026-02-25

## Pattern Overview

**Overall:** Multi-layered command-driven architecture with clean separation between CLI commands and business logic

**Key Characteristics:**
- CLI command layer decoupled from business logic
- Repository pattern for data access via Supabase REST API
- Dual AI provider strategy (OpenAI primary, Gemini fallback)
- Pub-to-publishing pipeline: Books → Ideas → Scripts → Calendar → Publish
- Error handling with retry logic and context awareness
- External integration via Blotato for social media publishing

## Layers

**Command Layer (cmd/):**
- Purpose: Thin Cobra command definitions, flag parsing, user interaction
- Location: `cmd/*/` (books, calendar, generate, ideas, publish, stats, auth, db, test)
- Contains: Cobra command setup, flag definitions, output formatting
- Depends on: internal packages (config, repository, generator, scheduler, social)
- Used by: CLI entrypoint (`main.go` → `cmd/root.go` → `Execute()`)

**Business Logic Layer (internal/):**
- Purpose: Core algorithms and feature implementations
- Location: `internal/{generator,scheduler,parser,prompts,ai}`
- Contains: Idea generation, script generation, calendar planning, CSV parsing, AI clients
- Depends on: Models, repository layer, external APIs
- Used by: Command layer, repository layer (for coordination)

**Repository Layer (internal/repository/):**
- Purpose: Data access abstraction via Supabase REST API
- Location: `internal/repository/*.go` (books, content, calendar, sales, metrics)
- Contains: HTTP client wrappers for Supabase endpoints, CRUD operations
- Depends on: Models, Supabase config, HTTP client
- Used by: Command layer, business logic layer, scheduler

**Model Layer (internal/models/):**
- Purpose: Domain entities and validation
- Location: `internal/models/{book,content,date,sales,metrics}.go`
- Contains: Book, ContentIdea, ContentScript, ContentCalendar, SalesData, Metrics structs
- Depends on: Standard library only
- Used by: All layers

**Integration Layer (internal/{ai,social,config}):**
- Purpose: External API clients and configuration
- Location: `internal/ai/{openai,gemini}.go`, `internal/social/{blotato,instagram,tiktok}.go`, `internal/config/`
- Contains: OpenAI client, Gemini browser automation, Blotato publisher, config management
- Depends on: HTTP client, chromedp (for Gemini), standard library
- Used by: Business logic layer (generator)

## Data Flow

**Content Generation Pipeline:**

1. **Books Management** (`cmd/books/add` → `internal/repository/books.go`)
   - User adds book to catalog via `books add --title ... --genre ...`
   - BooksRepository creates HTTP POST to Supabase `/rest/v1/books`
   - Returns Book with ID, stored for later reference

2. **Idea Generation** (`cmd/generate/ideas` → `internal/generator/ideas.go`)
   - User runs `generate ideas --book <id>` (or all books if not specified)
   - IdeaGenerator builds prompt from `internal/prompts/templates.go`
   - Tries OpenAI first with exponential backoff (max 3 attempts)
   - Falls back to Gemini if OpenAI fails
   - JSON response parsed into GeneratedIdea structs
   - ContentRepository.SaveIdeas() stores each as ContentIdea with status `pending`

3. **Idea Approval** (`cmd/ideas/` → `internal/repository/content.go`)
   - User lists pending ideas with `ideas list`
   - User approves/rejects with `ideas approve <id>` or `ideas reject <id>`
   - ContentRepository updates idea status to `approved` or `rejected`

4. **Script Generation** (`cmd/generate/script` → `internal/generator/scripts.go`)
   - User runs `generate script --idea <id>` or `generate batch` (all approved ideas)
   - ScriptGenerator builds platform-specific prompt (Instagram vs TikTok)
   - AI generates Hook, FullScript, CTA, Hashtags
   - ContentRepository stores ContentScript and updates ContentIdea status to `scripted`

5. **Calendar Planning** (`cmd/calendar/plan` → `internal/scheduler/planner.go`)
   - User runs `calendar plan --days 7 --posts 2`
   - Planner fetches all scripts from ContentRepository
   - Optimizer calculates optimal posting times (industry best practices)
   - Scripts balanced across days and platforms
   - ContentCalendarInput created for each slot with status `pending_approval`
   - CalendarRepository saves ContentCalendar entries

6. **Calendar Approval** (`cmd/calendar/approve` → `internal/repository/calendar.go`)
   - User reviews scheduled posts with `calendar show`
   - User approves with `calendar approve <id>` or rejects
   - CalendarRepository updates status to `approved` or sets back to `pending_approval`

7. **Publishing** (`cmd/publish/publish` → `internal/social/blotato.go`)
   - Manual: User runs `publish <calendar-id>` (--with-media to generate images)
   - Automated: Supabase Edge Function runs every 15min (cron job)
   - PublishCommand fetches calendar entry and associated script
   - BlotatoClient.PublishPost() sends to Blotato API
   - BlotatoClient.GetAccountID() determines target account
   - Status updated to `publishing` → `published` (or `failed` on error)

8. **Stats Collection** (`cmd/stats/import` → `internal/repository/sales.go`)
   - User imports KDP CSV with `stats import <file.csv>`
   - KDPParser parses CSV (flexible headers, handles commas in currency)
   - SalesRepository stores KDPReportRow data
   - User runs `stats correlate --days 30` to calculate Pearson correlation

**State Management:**
- Configuration: Viper + YAML file (`~/.gagipress/config.yaml`)
- Database state: Supabase PostgreSQL via REST API
- Content statuses: pending → approved/rejected → scripted → scheduled → publishing → published
- Error state: Stored in `publish_errors` JSONB field on ContentCalendar entries

## Key Abstractions

**Repository Pattern:**
- Purpose: Isolate HTTP/database access from business logic
- Examples: `internal/repository/{books,content,calendar,sales,metrics}.go`
- Pattern: Each repo handles one table, uses Supabase REST API with structured requests
- API key fallback: ServiceKey (if configured) → AnonKey

**AI Client Abstraction:**
- Purpose: Provide consistent interface for AI text generation with dual providers
- Examples: `internal/ai/{openai,gemini}.go` both implement GenerateText(prompt) → (text, error)
- Pattern: OpenAI is primary (fast, paid), Gemini is fallback (free, slower, browser-based)
- Retry logic: Only in OpenAI, not Gemini (Gemini has no built-in retry)

**Error Handling Abstraction:**
- Purpose: Categorize errors and support context-aware retry logic
- Examples: `internal/errors/{errors,retry}.go`
- ErrorType enum: validation, api, database, not_found, network
- Retry: Only retries on API/network errors, not validation errors

**Configuration Abstraction:**
- Purpose: Centralized config management with both mapstructure and yaml tags
- Examples: `internal/config/config.go`
- Uses Viper for loading/saving, supports env var overrides
- Required: Supabase URL + AnonKey; Optional: OpenAI, Instagram, TikTok, Amazon, Blotato, Gemini

**Prompt Templates:**
- Purpose: DRY prompt generation for AI with niche-specific guidelines
- Examples: `internal/prompts/templates.go`
- Supports BookNiche: children, puzzles, dialect_puzzles, savings
- Each niche has custom guidelines, all share idea categories (educational, entertainment, bts, ugc, trend)

## Entry Points

**CLI Root Command:**
- Location: `cmd/root.go`
- Triggers: `gagipress` binary with any subcommand
- Responsibilities: Load config, setup Viper, register all subcommands, global flags (--config, --verbose)

**Subcommand Groups:**
- `books`: Add/list/edit/delete books, import sales from CSV
- `generate`: Create ideas and scripts (ai content generation)
- `ideas`: Approve/reject pending ideas
- `calendar`: Plan, approve, show, retry, status, generate-media for scheduled posts
- `publish`: Single or batch publish via Blotato
- `stats`: Import KDP sales, correlate with engagement
- `auth`: Setup Instagram/TikTok OAuth (WIP)
- `db`: Migrate schema, check connection status
- `test`: Debug Gemini or other integrations

**Database Entry Point:**
- Location: `migrations/` and `supabase/migrations/`
- Command: `gagipress db migrate` calls `supabase db push`
- Uses: Native PostgreSQL functions (gen_random_uuid()), no extensions

**External Integration Entry Points:**
- OpenAI: `internal/ai/openai.go` NewOpenAIClient(config)
- Gemini: `internal/ai/gemini.go` NewGeminiClient(headless bool)
- Blotato: `internal/social/blotato.go` NewBlotatoClient(apiKey)
- Supabase REST: Direct HTTP via `internal/repository/*`

## Error Handling

**Strategy:** Custom AppError type with categorization, no panics in production code

**Patterns:**
- Validation errors: Return immediately, no retry
- API errors (5xx): Retry with exponential backoff (max 3 attempts, 1-30 sec)
- Network errors: Wrap with context, propagate up
- Database errors: Wrap with operation context (create, update, query failed)
- Gemini failures: Fall back to... (nowhere, so returns error)

**Error Wrapping:**
```go
// Wrap with context
return nil, fmt.Errorf("failed to get book: %w", err)

// Custom AppError type
return nil, errors.Wrap(err, errors.ErrorTypeAPI, "OpenAI API call failed")
```

**Context Support:**
- Retry logic respects context.Done() for cancellation
- All long-running operations accept context parameter (if needed)

## Cross-Cutting Concerns

**Logging:**
- Approach: fmt.Println() for user-facing messages, no structured logging framework
- Spinners: `internal/ui/spinner.go` provides visual feedback during long operations
- Output: Colored headers and formatted tables via lipgloss

**Validation:**
- Approach: Model-level Validate() methods on *Input types
- Validates presence, enum values, format constraints
- Called before repository Create/Update operations

**Authentication:**
- Approach: Config-based API keys (Supabase, OpenAI, Blotato)
- ServiceKey vs AnonKey: Prefer ServiceKey for admin operations, fallback to AnonKey
- HTTP headers: Set both `apikey` and `Authorization: Bearer` for Supabase

**Rate Limiting:**
- Approach: No explicit rate limiting, relies on API providers' quotas
- OpenAI: Retry logic with exponential backoff
- Blotato: 30-second timeout per request
- Supabase: REST API rate limits apply

**Data Validation:**
- CSV parsing: Flexible column name matching, graceful date format fallback
- JSON marshaling: Use json struct tags, support JSONB fields (metadata, publish_errors)
- UUID handling: Stored as text in Supabase, used as ID references

---

*Architecture analysis: 2026-02-25*
