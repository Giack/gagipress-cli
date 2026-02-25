# Codebase Structure

**Analysis Date:** 2026-02-25

## Directory Layout

```
sales-mkgt-automation/
├── main.go                          # CLI entrypoint
├── cmd/                             # Cobra command definitions (thin layer)
│   ├── root.go                      # Main CLI setup, subcommand registration
│   ├── init.go                      # CLI initialization hooks
│   ├── version.go                   # Version command
│   ├── fix-config.go                # Config migration utility
│   ├── auth/                        # OAuth setup for Instagram/TikTok
│   ├── books/                       # Book management (add, list, edit, delete, sales import)
│   ├── calendar/                    # Content calendar (plan, approve, show, status, retry, generate-media)
│   ├── db/                          # Database operations (migrate, status)
│   ├── generate/                    # AI content generation (ideas, scripts, batch)
│   ├── ideas/                       # Idea workflow (approve, reject, list)
│   ├── publish/                     # Publishing via Blotato (single, batch)
│   ├── stats/                       # Analytics (import KDP CSV, correlate, show)
│   └── test/                        # Debug utilities (gemini, openai)
├── internal/                        # Business logic (testable, reusable)
│   ├── ai/                          # AI provider clients
│   │   ├── openai.go                # OpenAI GPT-4o-mini client
│   │   └── gemini.go                # Google Gemini browser automation (via chromedp)
│   ├── config/                      # Configuration management
│   │   └── config.go                # Config struct, Load/Save, validation
│   ├── errors/                      # Error handling utilities
│   │   ├── errors.go                # AppError type with categorization
│   │   ├── retry.go                 # Exponential backoff retry logic
│   │   ├── errors_test.go           # Error handling tests
│   │   └── retry_test.go            # Retry logic tests
│   ├── generator/                   # Content generation logic
│   │   ├── ideas.go                 # Idea generation using AI
│   │   └── scripts.go               # Script generation (platform-specific)
│   ├── models/                      # Domain entities
│   │   ├── book.go                  # Book struct and BookInput
│   │   ├── content.go               # ContentIdea, ContentScript, ContentCalendar
│   │   ├── date.go                  # Custom Date type
│   │   ├── sales.go                 # KDPReportRow, SalesData
│   │   ├── metrics.go               # PostMetrics, EngagementMetrics
│   │   ├── book_test.go             # Book validation tests
│   │   ├── content_test.go          # Content model tests
│   │   ├── date_test.go             # Date parsing tests
│   │   └── metrics_test.go          # Metrics calculation tests
│   ├── parser/                      # CSV and data parsing
│   │   ├── kdp.go                   # KDP sales report CSV parser (97% coverage)
│   │   └── kdp_test.go              # Parser tests with edge cases
│   ├── prompts/                     # AI prompt templates
│   │   └── templates.go             # IdeaPromptTemplate, ScriptPromptTemplate (niche-aware)
│   ├── repository/                  # Data access layer (Supabase REST API)
│   │   ├── books.go                 # CRUD for books table
│   │   ├── content.go               # CRUD for content_ideas, content_scripts
│   │   ├── calendar.go              # CRUD for content_calendar
│   │   ├── sales.go                 # CRUD for sales_data
│   │   ├── metrics.go               # CRUD for post_metrics
│   │   ├── books_test.go            # Repository tests
│   │   ├── content_test.go          # Content repo tests
│   │   ├── calendar_test.go         # Calendar repo tests
│   │   └── metrics_test.go          # Metrics repo tests
│   ├── scheduler/                   # Content calendar planning algorithms
│   │   ├── planner.go               # Weekly plan generation
│   │   ├── optimizer.go             # Optimal posting time calculation
│   │   ├── planner_test.go          # Scheduler tests
│   │   └── optimizer_test.go        # Optimizer tests
│   ├── social/                      # Social media API clients
│   │   ├── blotato.go               # Blotato publishing API (primary)
│   │   ├── instagram.go             # Instagram Graph API (stub)
│   │   └── tiktok.go                # TikTok Creator API (stub)
│   ├── testutil/                    # Testing utilities
│   │   └── helpers.go               # Common test assertions and fixtures
│   └── ui/                          # CLI UI components
│       ├── spinner.go               # Loading spinner
│       └── [other UI helpers]
├── migrations/                      # Database schema migrations
│   ├── 001_init_schema.sql          # Initial tables (books, content_ideas, etc.)
│   ├── 002_*.sql                    # Migration files (follow naming: NNN_description.sql)
│   ├── 003_*.sql
│   └── 004_*.sql
├── supabase/                        # Supabase CLI configuration
│   ├── config.toml                  # Supabase project config
│   ├── functions/                   # Edge Functions
│   │   └── publish-scheduled/
│   │       └── index.ts             # Cron-triggered publishing (every 15min)
│   └── migrations/                  # Auto-synced migrations (don't edit)
├── test/                            # Integration tests
│   └── integration/
│       └── *_test.go                # Integration test files (require Supabase credentials)
├── docs/                            # Documentation
│   └── plans/                       # Implementation plans
└── scripts/                         # Utility scripts (makefiles, build scripts)
```

## Directory Purposes

**cmd/ - Command Definitions:**
- Purpose: Thin Cobra command wrappers, flag parsing, output formatting
- Contains: One package per command group (books, calendar, generate, etc.)
- Key files: Each subcommand has its own file (e.g., `books/add.go`, `books/list.go`)
- Entry point: `root.go` registers all subcommands via `AddCommand()`

**internal/ai/ - AI Providers:**
- Purpose: Abstraction over external AI APIs
- Contains: OpenAI client (primary, paid), Gemini client (fallback, free browser automation)
- Key files: `openai.go` with retry logic, `gemini.go` with chromedp automation
- Usage: Created by generator.IdeaGenerator and generator.ScriptGenerator

**internal/config/ - Configuration:**
- Purpose: Centralized config management
- Contains: Config struct with all providers, Load/Save methods, Viper integration
- Key files: `config.go` with SupabaseConfig, OpenAIConfig, BlotatoConfig, etc.
- Usage: Loaded once at CLI startup, passed to repositories and generators

**internal/errors/ - Error Handling:**
- Purpose: Custom error types and retry logic
- Contains: AppError with ErrorType enum, Retry function with exponential backoff
- Key files: `errors.go` (types), `retry.go` (backoff logic)
- Usage: In all layers for consistent error categorization and handling

**internal/generator/ - AI Content Generation:**
- Purpose: Core algorithms that use AI to create content
- Contains: IdeaGenerator (20+ ideas), ScriptGenerator (platform-specific scripts)
- Key files: `ideas.go`, `scripts.go`
- Usage: Called by cmd/generate/ideas.go and cmd/generate/script.go

**internal/models/ - Domain Entities:**
- Purpose: Data structures representing business concepts
- Contains: Book, ContentIdea, ContentScript, ContentCalendar, SalesData, Metrics
- Key files: Each entity in its own file (book.go, content.go, sales.go, metrics.go)
- Pattern: Each *Input type has Validate() method, returned from API as primary type

**internal/parser/ - Data Parsing:**
- Purpose: CSV parsing for KDP sales reports
- Contains: KDPParser with flexible column name matching, multiple date format support
- Key files: `kdp.go` (parser implementation), `kdp_test.go` (97% coverage)
- Known edge cases: Commas in currency ($1,234.56), flexible date formats (YYYY-MM-DD or MM/DD/YYYY)

**internal/prompts/ - AI Prompt Templates:**
- Purpose: DRY prompt generation with niche-specific customization
- Contains: IdeaPromptTemplate, ScriptPromptTemplate functions
- Key files: `templates.go` with BookNiche enum (children, puzzles, dialect_puzzles, savings)
- Usage: Called by generators to build prompts before sending to AI

**internal/repository/ - Data Access:**
- Purpose: Abstraction over Supabase REST API
- Contains: One repository per table (BooksRepository, ContentRepository, CalendarRepository, SalesRepository, MetricsRepository)
- Key files: Each repo in own file (books.go, content.go, calendar.go, etc.)
- Pattern: All repos use HTTP client, prefer ServiceKey over AnonKey, set required headers

**internal/scheduler/ - Calendar Planning:**
- Purpose: Algorithms for creating optimal content schedules
- Contains: Planner (creates weekly schedule), Optimizer (calculates posting times)
- Key files: `planner.go` (orchestrates scheduling), `optimizer.go` (best practice times)
- Usage: Called by cmd/calendar/plan.go to generate ContentCalendarInput entries

**internal/social/ - Social Media APIs:**
- Purpose: Publishing and account management
- Contains: BlotatoClient (primary, active), InstagramClient (stub), TikTokClient (stub)
- Key files: `blotato.go` (implementation), instagram.go and tiktok.go (stubs)
- Usage: Called by cmd/publish/publish.go and Edge Functions

**internal/testutil/ - Test Utilities:**
- Purpose: Common testing helpers
- Contains: Assertion functions, fixture builders, mock factories
- Key files: `helpers.go`
- Usage: Imported by *_test.go files throughout codebase

**internal/ui/ - CLI UI Components:**
- Purpose: User-facing output formatting
- Contains: Spinners, colored headers, formatted tables
- Key files: `spinner.go` (loading indicators)
- Usage: In command implementations for visual feedback

**migrations/ - Database Schema:**
- Purpose: PostgreSQL DDL migrations
- Contains: SQL files with version numbering (NNN_description.sql)
- Key files: 001_init_schema.sql (tables), 002-004_*.sql (enhancements)
- Applied via: `gagipress db migrate` → `supabase db push`

**supabase/ - Supabase Configuration:**
- Purpose: Project config and Edge Functions
- Contains: config.toml for project settings, functions/ for serverless
- Key files: `publish-scheduled/index.ts` (cron job runs every 15min)
- Note: supabase/migrations/ auto-synced from migrations/, don't edit directly

**test/integration/ - Integration Tests:**
- Purpose: End-to-end tests requiring Supabase credentials
- Contains: Tests that call real APIs (repositories, generators)
- Key files: *_test.go files in test/integration/
- Usage: Run with `make test-integration` (requires SUPABASE_URL, SUPABASE_ANON_KEY)

## Key File Locations

**Entry Points:**
- `main.go`: CLI entrypoint
- `cmd/root.go`: Cobra root command, subcommand registration
- `cmd/*/[command].go`: Individual subcommand implementations (e.g., cmd/books/add.go)

**Configuration:**
- `internal/config/config.go`: Config struct, Load/Save, validation

**Core Logic:**
- `internal/generator/ideas.go`: Idea generation
- `internal/generator/scripts.go`: Script generation
- `internal/scheduler/planner.go`: Calendar planning
- `internal/ai/openai.go`: OpenAI client with retry
- `internal/ai/gemini.go`: Gemini browser automation
- `internal/social/blotato.go`: Blotato publishing API

**Data Access:**
- `internal/repository/books.go`: Book CRUD
- `internal/repository/content.go`: Content idea/script CRUD
- `internal/repository/calendar.go`: Calendar entry CRUD
- `internal/repository/sales.go`: Sales data import/query

**Models:**
- `internal/models/book.go`: Book entity and validation
- `internal/models/content.go`: ContentIdea, ContentScript, ContentCalendar
- `internal/models/sales.go`: KDPReportRow, SalesData

**Testing:**
- `internal/parser/kdp_test.go`: CSV parser tests (97% coverage)
- `internal/models/*_test.go`: Model validation tests
- `internal/errors/*_test.go`: Error and retry logic tests
- `test/integration/*_test.go`: Integration tests

## Naming Conventions

**Files:**
- `*.go`: Go source files
- `*_test.go`: Test files (same package as implementation)
- `*_integration_test.go`: Integration tests in test/integration/
- Migrations: `NNN_description.sql` (e.g., 001_init_schema.sql, 002_add_media_fields.sql)

**Directories:**
- lowercase with underscores: `internal/config/`, `cmd/auth/`
- Plural form: `migrations/`, `internal/models/`, `test/integration/`
- Group by domain: `cmd/`, `internal/repository/`, `internal/ai/`

**Functions/Methods:**
- Constructors: `New[Type]()` (e.g., `NewBooksRepository()`)
- CRUD methods: `Create()`, `GetByID()`, `GetAll()`, `Update()`, `Delete()`
- Getters: `Get[Field]()` (e.g., `GetOptimalTimes()`)
- Business logic: `[Verb][Noun]()` (e.g., `GenerateIdeas()`, `PublishPost()`)

**Types/Interfaces:**
- Entities: `PascalCase` (Book, ContentIdea, ContentScript)
- Repositories: `[Entity]Repository` (BooksRepository, ContentRepository)
- Clients: `[Provider]Client` (OpenAIClient, BlotatoClient)
- Errors: `Err[Reason]` or `ErrorType` enum

**Variables/Constants:**
- Package constants: `UPPERCASE` (BlotatoBaseURL)
- Package vars: `camelCase` (cfgFile in cmd/root.go)
- Local variables: `camelCase` (bookID, postsPerDay)

## Where to Add New Code

**New Feature (e.g., a new command):**
1. Create directory: `cmd/[feature]/`
2. Define command: `cmd/[feature]/[feature].go` (main command struct)
3. Define subcommands: `cmd/[feature]/[action].go` (add.go, list.go, etc.)
4. Business logic: `internal/[feature]/` package if reusable
5. Tests: `cmd/[feature]/[feature]_test.go`, `internal/[feature]/*_test.go`

**New Repository Method:**
1. Add method to: `internal/repository/[entity].go`
2. Follow pattern: HTTP request → parse JSON → return typed result or error
3. Use ServiceKey preference, set required headers (apikey, Authorization, Prefer)
4. Test: Add test case to `internal/repository/[entity]_test.go`

**New Model/Entity:**
1. Create file: `internal/models/[entity].go`
2. Define struct: `type [Entity] struct { ... }`
3. Define input type: `type [Entity]Input struct { ... }`
4. Add Validate() method to input type
5. Test: Create `internal/models/[entity]_test.go`

**New Generator/Algorithm:**
1. Create file: `internal/generator/[algorithm].go`
2. Define generator struct with dependencies (config, repos, clients)
3. Add NewGenerator() constructor
4. Implement main algorithm function
5. Command layer: Call from `cmd/generate/[algorithm].go`
6. Test: Add `internal/generator/[algorithm]_test.go`

**Utilities/Shared Helpers:**
- String utilities: `internal/ui/` or create `internal/utils/`
- Error helpers: `internal/errors/`
- Test fixtures: `internal/testutil/`
- CSV parsing: `internal/parser/`

**Configuration:**
- Add new provider config: Edit `internal/config/config.go` (add struct, tags)
- Command to set config: `cmd/auth/[provider].go`
- Save logic: `internal/config/config.go` Save() method handles all fields

## Special Directories

**migrations/:**
- Purpose: Database schema versioning
- Generated: No (manually created)
- Committed: Yes
- Apply with: `gagipress db migrate` (calls `supabase db push`)
- Format: NNN_description.sql (e.g., 001_init_schema.sql)

**supabase/migrations/:**
- Purpose: Auto-synced mirror of migrations/
- Generated: Yes (auto-synced by Supabase CLI)
- Committed: No (generated files)
- Don't edit directly, edit migrations/ instead

**supabase/functions/:**
- Purpose: Serverless Edge Functions (cron jobs, webhooks)
- Generated: No (manually created)
- Committed: Yes
- Deploy with: `supabase functions deploy publish-scheduled`
- Env: Set `BLOTATO_API_KEY` via `supabase secrets set`

**test/integration/:**
- Purpose: End-to-end tests requiring live services
- Generated: No (manually created)
- Committed: Yes
- Run with: `make test-integration` (requires credentials)
- Skipped: Tests skip gracefully if SUPABASE_URL not set

**docs/plans/:**
- Purpose: Implementation plans and design docs
- Generated: No (manually created or AI-generated)
- Committed: Yes
- Format: Markdown, dated (YYYY-MM-DD-description.md)

**bin/:**
- Purpose: Compiled CLI binary output
- Generated: Yes (`make build` creates bin/gagipress)
- Committed: No (.gitignored)
- Install: `make install` copies to system PATH

---

*Structure analysis: 2026-02-25*
