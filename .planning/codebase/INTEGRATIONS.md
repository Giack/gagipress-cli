# External Integrations

**Analysis Date:** 2026-02-25

## APIs & External Services

**AI/Content Generation:**
- **OpenAI** - Primary AI provider for content generation
  - SDK: Direct HTTP calls to `https://api.openai.com/v1`
  - Client: `internal/ai/openai.go` - `OpenAIClient` struct
  - Auth: API key in `Authorization: Bearer` header
  - Model: `gpt-4o-mini` (default, configurable)
  - Retry: Exponential backoff (3 attempts) via `internal/errors/retry.go`
  - Used for: Idea generation, script creation, prompt completion

- **Google Gemini** - Fallback AI provider (free, via browser automation)
  - Method: Browser automation (chromedp) to `https://gemini.google.com`
  - Client: `internal/ai/gemini.go` - `GeminiClient` struct
  - Auth: Accessed without API auth (free tier)
  - Selectors: DOM-based (`textarea[placeholder*="Enter a prompt"]`) — fragile
  - Used when: OpenAI unavailable or `--gemini` flag set
  - Risk: DOM changes will break automation

- **Google Imagen (via Blotato)** - Image generation for post covers
  - Method: Blotato proxy API
  - Details: See Blotato integration below

**Social Media Publishing:**
- **Blotato API** - Proxy service for Instagram and TikTok publishing
  - Base URL: `https://backend.blotato.com/v2`
  - Client: `internal/social/blotato.go` - `BlotatoClient` struct
  - Auth: `blotato-api-key` header
  - Endpoints:
    - `GET /users/me/accounts?platform={platform}` - Fetch connected accounts
    - `POST /videos/from-templates` - Create video from template with Imagen
    - `GET /videos/creations/{id}` - Poll media generation status
    - `POST /create` - Publish post to platform
  - Edge Function: `supabase/functions/publish-scheduled/index.ts` calls Blotato every 15 minutes

- **Instagram Graph API** - Direct Instagram integration (stub implementation)
  - Client: `internal/social/instagram.go` - `InstagramClient` struct
  - Status: Not fully implemented (awaiting OAuth setup)
  - Required: Access token + account ID (stored in config)
  - OAuth: Deferred to Phase 2

- **TikTok Creator API** - Direct TikTok integration (stub implementation)
  - Client: `internal/social/tiktok.go` - `TikTokClient` struct
  - Status: Not fully implemented (awaiting OAuth setup)
  - Required: Access token + account ID (stored in config)
  - OAuth: Deferred to Phase 2

## Data Storage

**Databases:**
- **Supabase (PostgreSQL 13+)**
  - Connection: Via REST API (`PostgREST`)
  - Client: Direct HTTP calls to `{SUPABASE_URL}/rest/v1/`
  - Auth: `apikey` header + `Authorization: Bearer` (ServiceKey preferred)
  - Tables:
    - `books` - Book catalog with KDP metadata
    - `content_ideas` - Generated ideas with approval workflow
    - `content_scripts` - Platform-specific scripts
    - `content_calendar` - Scheduled posts
    - `post_metrics` - Social engagement data
    - `sales_data` - KDP sales import
  - Cron: pg_cron schedules `publish-scheduled` Edge Function every 15 minutes

**File Storage:**
- **Supabase Storage** - Public bucket for campaign media
  - Bucket: `campaign-media` (public)
  - Path pattern: Generated images stored with `media_url` reference in `content_calendar`
  - RLS Policies: Enabled for read/write/delete/update operations
  - Media generation: Triggered by Edge Function on demand

**Caching:**
- None detected

## Authentication & Identity

**Auth Provider:**
- Custom (no external OAuth provider integrated for CLI)
- Config-based authentication via `~/.gagipress/config.yaml`

**Implementation:**
- Viper loads configuration from YAML + environment variables
- No built-in session management or user authentication
- Assumes single-user CLI environment (config stored in user home directory)

## Monitoring & Observability

**Error Tracking:**
- Not configured (no external error tracking service)
- Custom error handling in `internal/errors/errors.go`
- Error types: AppError with context and retry logic

**Logs:**
- Console-based logging (via Cobra/standard library)
- Verbose flag: `--verbose` enables debug output
- No external log aggregation

## CI/CD & Deployment

**Hosting:**
- Supabase (PostgreSQL, Edge Functions, Storage) - Cloud-hosted
- CLI: Local binary deployment (users run `gagipress` command)
- Edge Function: Deployed to Supabase (`supabase/functions/publish-scheduled/`)
- Publishing Pipeline: Cron → Edge Function → Blotato → Instagram/TikTok

**CI Pipeline:**
- Not configured (no GitHub Actions, GitLab CI, etc.)
- Manual testing via `make test`, `make test-integration`
- Makefile targets for build/test/coverage

**Supabase CLI:**
- Used for: Database migrations, local development, Edge Function deployment
- Commands:
  - `supabase db push` - Apply migrations to remote
  - `supabase functions deploy publish-scheduled` - Deploy Edge Function
  - `supabase secrets set BLOTATO_API_KEY=...` - Configure secrets for cron

## Environment Configuration

**Required env vars:**
- `SUPABASE_URL` - Supabase project URL
- `SUPABASE_ANON_KEY` - Supabase anonymous key (fallback if no service key)

**Optional env vars:**
- `SUPABASE_SERVICE_KEY` - Preferred over anon key for admin operations
- `OPENAI_API_KEY` - OpenAI API key
- `GEMINI_API_KEY` - Google Gemini API key
- `BLOTATO_API_KEY` - Blotato API key (also stored as Supabase secret for cron)
- `INSTAGRAM_ACCESS_TOKEN` - Instagram Graph API token
- `INSTAGRAM_ACCOUNT_ID` - Instagram account ID
- `TIKTOK_ACCESS_TOKEN` - TikTok Creator API token
- `TIKTOK_ACCOUNT_ID` - TikTok account ID

**Secrets location:**
- `~/.gagipress/config.yaml` - User's local config (git-ignored)
- `supabase/.env.local` - Local development (git-ignored)
- Supabase Dashboard → Settings → Secrets - Production secrets for Edge Functions
  - Example: `BLOTATO_API_KEY` accessed via `Deno.env.get()` in Edge Function

## Webhooks & Callbacks

**Incoming:**
- Edge Function `publish-scheduled`: No external webhooks, triggered by pg_cron every 15 minutes

**Outgoing:**
- Blotato API calls: Publish posts to Instagram/TikTok
- No outbound webhooks configured

## Data Integration Points

**KDP Sales Import:**
- **Source**: Amazon KDP CSV export
- **Parser**: `internal/parser/kdp.go`
- **Format**: CSV with headers `Title,ASIN,Royalty,Units Sold,Date`
- **Destination**: `sales_data` table
- **Correlation**: `stats correlate` command correlates sales with social metrics via Pearson coefficient

**Metrics Collection:**
- **Source**: Instagram Graph API, TikTok Creator API
- **Destination**: `post_metrics` table
- **Sync**: Manual via `stats import` command (automatic collection deferred)

## API Key Management

**Preference Order (Supabase):**
1. `ServiceKey` (admin, preferred for CLI)
2. `AnonKey` (fallback, read-only by default)

**Reference Implementation** (`internal/repository/books.go`):
```go
apiKey := r.config.ServiceKey
if apiKey == "" {
    apiKey = r.config.AnonKey
}
req.Header.Set("apikey", apiKey)
req.Header.Set("Authorization", "Bearer "+apiKey)
```

## Third-Party Service Status

| Service | Status | Role | Phase |
|---------|--------|------|-------|
| Supabase | Active | Database + Edge Functions | Core |
| OpenAI | Active | Primary AI provider | Phase 2 |
| Blotato | Active | Publishing proxy | Phase 5 |
| Google Gemini | Active | Fallback AI (browser automation) | Phase 2 |
| Instagram Graph API | Stub | Direct publishing (OAuth deferred) | Phase 2 |
| TikTok Creator API | Stub | Direct publishing (OAuth deferred) | Phase 2 |
| Google Imagen | Via Blotato | Cover image generation | Phase 5 |

---

*Integration audit: 2026-02-25*
