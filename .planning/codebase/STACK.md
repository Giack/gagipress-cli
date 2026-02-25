# Technology Stack

**Analysis Date:** 2026-02-25

## Languages

**Primary:**
- Go 1.24.0 - Core CLI application
- TypeScript - Supabase Edge Functions

**Secondary:**
- SQL - PostgreSQL migrations and stored procedures
- YAML - Configuration files

## Runtime

**Environment:**
- Go 1.24.0 (managed via `mise`)

**Package Manager:**
- Go modules
- Lockfile: `go.sum` (present)

**Build Tool:**
- Go compiler (native binary output to `bin/gagipress`)

## Frameworks

**Core:**
- Cobra v1.10.2 - CLI command framework
- Viper v1.21.0 - Configuration management (YAML, environment variables, mapstructure unmarshaling)

**Browser Automation:**
- chromedp v0.14.2 - Headless browser automation (for Gemini fallback)
  - cdproto v0.0.0-20250724212937-08a3db8b4327 - Chrome DevTools Protocol
  - Provides browser control without external browser dependency

**Styling/UI:**
- charmbracelet/lipgloss v1.1.0 - Terminal UI styling and colors
- charmbracelet/colorprofile - Color profile detection for terminals

**Utilities:**
- golang.org/x/term v0.40.0 - Terminal control (ANSI escape sequences)

## Key Dependencies

**Critical:**
- github.com/spf13/cobra v1.10.2 - CLI framework (commands, flags, help)
- github.com/spf13/viper v1.21.0 - Configuration loading with environment variable support
- github.com/chromedp/chromedp v0.14.2 - Gemini browser automation (fallback AI provider)

**Infrastructure (indirect):**
- cloud.google.com/go v0.116.0 - Google Cloud client libraries (transitive via chromedp)
- google.golang.org/genai v1.47.0 - Google Generative AI client (transitive)

**Terminal UI (indirect):**
- charmbracelet/x/ansi, charmbracelet/x/term - Advanced terminal features
- muesli/termenv - Color and style output

## Configuration

**Environment:**
- Configuration file: `~/.gagipress/config.yaml`
- Format: YAML with structured sections for each service
- All config structs use both `mapstructure` (for Viper unmarshaling) and `yaml` (for writing) tags
  - Example from `internal/config/config.go`:
    ```go
    type SupabaseConfig struct {
        URL        string `mapstructure:"url" yaml:"url"`
        AnonKey    string `mapstructure:"anon_key" yaml:"anon_key"`
        ServiceKey string `mapstructure:"service_key" yaml:"service_key"`
    }
    ```

**Key Configuration Sections:**
- `supabase`: Database connection (URL, anon_key, service_key)
- `openai`: API key and model selection
- `instagram`: Access token and account ID
- `tiktok`: Access token and account ID
- `amazon`: Email and password (for KDP CSV export login)
- `blotato`: API key and template ID (for publishing)
- `gemini`: API key (for fallback image generation)

**Build Configuration:**
- Makefile: Targets for build, test, coverage, vet, format, install
- mise.toml: Go version pinning (1.24)

## Platform Requirements

**Development:**
- macOS/Linux/Windows with Go 1.24+
- Mise (optional, for automatic Go version management)
- Browser (headless Chrome/Chromium for Gemini automation)
- Network access to external APIs

**Production:**
- CLI deployed as single binary (`gagipress`)
- Requires Supabase project (cloud or self-hosted)
- Requires configuration in `~/.gagipress/config.yaml`
- Optional: API keys for OpenAI, Gemini, Blotato, Instagram Graph API, TikTok Creator API

## Database

**PostgreSQL:**
- Version: 13+ (uses native `gen_random_uuid()`)
- Hosted on Supabase (REST API access via PostgREST)
- Schema version tracked in `schema_version` table
- Extensions used:
  - `uuid-ossp` (implicit: `gen_random_uuid()` is native in PG 13+)
  - `pg_cron` - For scheduled cron jobs (publishing automation)
  - `pg_net` - For HTTP requests from database functions

**Migrations:**
- Location: `migrations/` (synced to `supabase/migrations/`)
- Tool: Supabase CLI (`supabase db push`)
- Naming convention: `NNN_description.sql`
- Migration history tracked in `supabase_migrations.schema_migrations`

## API Access Patterns

**HTTP Clients:**
- Go `net/http` (standard library)
- Direct HTTP calls (no SDK wrappers, except for Blotato abstraction in `internal/social/blotato.go`)
- Custom timeout handling (typically 30-60 seconds)
- Retry logic with exponential backoff in `internal/errors/retry.go`

**Authentication Patterns:**
- API keys in request headers (`apikey`, `Authorization: Bearer`, `blotato-api-key`)
- Supabase prefers ServiceKey > AnonKey
- Viper environment variable substitution (e.g., `SUPABASE_URL` maps to `supabase.url`)

---

*Stack analysis: 2026-02-25*
