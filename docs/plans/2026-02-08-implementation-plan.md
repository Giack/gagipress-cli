# Gagipress MVP - Implementation Plan

**Created**: 2026-02-08
**Status**: In Progress
**Design Doc**: [2026-02-08-gagipress-social-automation-design.md](2026-02-08-gagipress-social-automation-design.md)

## Overview

Piano dettagliato step-by-step per implementare MVP del sistema Gagipress Social Automation.

**Timeline**: 5-6 settimane
**Approach**: Iterativo, ogni settimana produce componente funzionante

---

## Week 1: Foundation (Current)

### 1.1 Go Project Setup ✅ COMPLETED

**Steps**:
- [x] Create worktree isolato
- [x] Initialize Go module
- [x] Setup Cobra CLI framework
- [x] Create basic command structure
- [x] Add configuration management (Viper)

**Deliverables**:
```bash
gagipress --help    # Shows command tree ✅
gagipress version   # Shows version info ✅
```

**Files**:
- `go.mod`, `go.sum` ✅
- `cmd/root.go` ✅
- `cmd/version.go` ✅
- `internal/config/config.go` ✅

**Completed**: 2026-02-08
**Commit**: e428bac

---

### 1.2 Supabase Project Setup ✅ COMPLETED

**Steps**:
- [ ] Create Supabase account/project (manual, via web - user task)
- [x] Implement `gagipress init` command ✅
  - Interactive wizard per credentials ✅
  - Test connection ✅
  - Save config to `~/.gagipress/config.yaml` ✅
- [x] Create SQL migration file con schema completo ✅
- [x] Implement `gagipress db migrate` command ✅
  - Esegue migration via HTTP API ✅
  - Crea tabelle, indici, RLS policies ✅

**Deliverables**:
```bash
gagipress init              # Setup completo ✅
gagipress db migrate        # Crea schema ✅
gagipress db status         # Verifica connessione ✅
```

**Files**:
- `cmd/init.go` ✅
- `cmd/db/migrate.go` ✅
- `cmd/db/status.go` ✅
- `cmd/db/db.go` ✅
- `migrations/001_initial_schema.sql` ✅
- `internal/supabase/client.go` ✅
- `internal/supabase/migrate.go` ✅

**Completed**: 2026-02-08
**Commit**: 42bf9a8

---

### 1.3 API Integrations Skeleton ✅ COMPLETED

**Steps**:
- [x] OpenAI client wrapper
  - API key configuration ✅
  - Chat completion helper ✅
  - Error handling + retry logic ✅
- [x] Instagram Graph API client (skeleton)
  - OAuth placeholder ✅
  - Basic types (Post, Metrics) ✅
- [x] TikTok API client (skeleton)
  - OAuth placeholder ✅
  - Basic types ✅
- [x] Gemini browser automation (chromedp)
  - Launch browser ✅
  - Navigate to Gemini ✅
  - Submit prompt ✅
  - Extract response ✅

**Deliverables**:
```bash
gagipress auth openai       # Test OpenAI connection ✅
gagipress auth instagram    # OAuth flow (future) ✅
gagipress auth tiktok       # OAuth flow (future) ✅
gagipress test gemini "Ciao" # Test Gemini automation ✅
```

**Files**:
- `internal/ai/openai.go` ✅
- `internal/ai/gemini.go` ✅
- `internal/social/instagram.go` ✅
- `internal/social/tiktok.go` ✅
- `cmd/auth/auth.go` ✅
- `cmd/auth/openai.go` ✅
- `cmd/auth/instagram.go` ✅
- `cmd/auth/tiktok.go` ✅
- `cmd/test/test.go` ✅
- `cmd/test/gemini.go` ✅
- `cmd/root.go` (updated) ✅
- `go.mod` (chromedp dependency) ✅

**Completed**: 2026-02-08
**Commit**: 7248b65

---

### Week 1 Acceptance Criteria

- [ ] CLI funziona e mostra help
- [ ] Supabase connesso, schema creato
- [ ] OpenAI API testata con successo
- [ ] Gemini browser automation funzionante
- [ ] Config salvato correttamente
- [ ] Tests passano (unit tests base)
- [ ] Documentation aggiornata in README.md

---

## Week 2: Content Generation

### 2.1 Database Models ✅ COMPLETED

**Steps**:
- [x] Create Go structs per tabelle:
  - `Book` ✅
  - `ContentIdea` ✅
  - `ContentScript` ✅
  - `ContentCalendar` ✅
- [x] CRUD operations con Supabase client ✅
- [x] Repository pattern per clean architecture ✅

**Files**:
- `internal/models/book.go` ✅
- `internal/models/content.go` ✅
- `internal/repository/books.go` ✅
- `internal/repository/content.go` ✅

**Completed**: 2026-02-09
**Commit**: 0b2bd57

---

### 2.2 Book Management ✅ COMPLETED

**Steps**:
- [x] Implement `gagipress books add` command ✅
  - Interactive input: titolo, genere, audience, ASIN ✅
  - Upload cover image (optional) ✅
  - Save to Supabase ✅
- [x] Implement `gagipress books list` command ✅
  - Tabella con tutti libri ✅
- [x] Implement `gagipress books edit <id>` command ✅
- [x] Implement `gagipress books delete <id>` command ✅

**Deliverables**:
```bash
gagipress books add         ✅
gagipress books list        ✅
gagipress books edit <id>   ✅
gagipress books delete <id> ✅
```

**Files**:
- `cmd/books/books.go` ✅
- `cmd/books/add.go` ✅
- `cmd/books/list.go` ✅
- `cmd/books/edit.go` ✅
- `cmd/books/delete.go` ✅

**Completed**: 2026-02-09
**Commit**: 0a32a92

---

### 2.3 Idea Generation ✅ COMPLETED

**Steps**:
- [x] Implement `gagipress generate ideas` command ✅
  - Legge libri dal database ✅
  - Costruisce prompt per AI (OpenAI primary) ✅
  - Fallback a Gemini se OpenAI fail ✅
  - Genera 20-30 idee ✅
  - Categorizza (educational, entertainment, etc.) ✅
  - Salva in `content_ideas` table ✅
- [x] Create prompt templates per niche: ✅
  - Libri bambini ✅
  - Enigmistica dialetto ✅
  - Libri risparmio ✅
- [x] Relevance scoring algorithm ✅

**Deliverables**:
```bash
gagipress generate ideas --count 30  ✅
gagipress ideas list                 ✅
gagipress ideas approve <id>         ✅
gagipress ideas reject <id>          ✅
```

**Files**:
- `cmd/generate/generate.go` ✅
- `cmd/generate/ideas.go` ✅
- `cmd/ideas/ideas.go` ✅
- `cmd/ideas/list.go` ✅
- `cmd/ideas/approve.go` ✅
- `cmd/ideas/reject.go` ✅
- `internal/generator/ideas.go` ✅
- `internal/prompts/templates.go` ✅

**Completed**: 2026-02-09
**Commit**: fbdbd8f

---

### 2.4 Script Generation ✅ COMPLETED

**Steps**:
- [x] Implement `gagipress generate script <idea-id>` command ✅
  - Legge idea dal database ✅
  - Costruisce prompt dettagliato ✅
  - Genera: hook, full script, CTA, hashtags, visual notes ✅
  - Salva in `content_scripts` table ✅
- [ ] Implement `gagipress generate batch` command ⏭️ (future enhancement)
  - Approva automaticamente top N idee
  - Genera script per tutte
- [x] Create script templates per content type ✅

**Deliverables**:
```bash
gagipress generate script <idea-id> --platform tiktok/instagram  ✅
```

**Files**:
- `cmd/generate/script.go` ✅
- `internal/generator/scripts.go` ✅
- `internal/prompts/templates.go` (updated) ✅

**Completed**: 2026-02-09
**Commit**: d86943f
gagipress scripts show <id>    # Mostra script completo
```

**Files**:
- `cmd/generate/script.go`
- `cmd/generate/batch.go`
- `cmd/scripts/list.go`
- `cmd/scripts/show.go`
- `internal/generator/scripts.go`
- `templates/prompts/script_*.txt`

---

### Week 2 Acceptance Criteria

- [ ] Almeno 3 libri aggiunti nel database
- [ ] `generate ideas` produce 20-30 idee sensate
- [ ] `generate script` produce script completo con tutti campi
- [ ] `generate batch` funziona end-to-end
- [ ] Prompt templates customizzati per niche
- [ ] Unit tests per generators
- [ ] Documentation aggiornata

---

## Week 3: Scheduling & Publishing

### 3.1 Calendar Planning Algorithm

**Steps**:
- [ ] Implement `gagipress calendar plan` command
  - Legge script approvati dal database
  - Algoritmo per best posting times:
    - Analizza metriche storiche (se disponibili)
    - Default a peak times standard
  - Content mix balancing
  - Book rotation
  - Genera piano con `scheduled_for` timestamps
  - Salva in `content_calendar` con status `pending_approval`

**Deliverables**:
```bash
gagipress calendar plan --days 7
gagipress calendar show         # Visualizza piano corrente
```

**Files**:
- `cmd/calendar/plan.go`
- `cmd/calendar/show.go`
- `internal/scheduler/planner.go`
- `internal/scheduler/optimizer.go`

---

### 3.2 Interactive Approval UI

**Steps**:
- [ ] Implement `gagipress calendar approve` command
  - TUI (Text User Interface) con bubbletea/lipgloss
  - Lista contenuti schedulati
  - Azioni per ogni item: Approve, Edit, Reject, Reschedule
  - Navigazione con arrow keys
  - Batch operations
  - Update database status

**Deliverables**:
```bash
gagipress calendar approve   # Interactive TUI
```

**Files**:
- `cmd/calendar/approve.go`
- `internal/tui/approval.go`
- `internal/tui/models.go`

---

### 3.3 Supabase Edge Functions

**Steps**:
- [ ] Setup Supabase Edge Functions project structure
- [ ] Implement `auto-publish` Edge Function:
  - Query contenuti approvati schedulati per ora corrente
  - Per ogni contenuto:
    - Call Instagram/TikTok API
    - Upload media (if available)
    - Publish post
    - Save post_url, published_at
    - Handle errors con retry
  - Configurare cron: `0 * * * *`
- [ ] Implement `gagipress deploy functions` command
  - Deploy Edge Functions via Supabase CLI

**Deliverables**:
```bash
gagipress deploy functions
gagipress logs --function auto-publish
```

**Files**:
- `supabase/functions/auto-publish/index.ts`
- `supabase/functions/_shared/instagram.ts`
- `supabase/functions/_shared/tiktok.ts`
- `cmd/deploy/functions.go`
- `cmd/logs.go`

---

### 3.4 Social API Integration (Real)

**Steps**:
- [ ] Instagram Graph API:
  - OAuth flow completo
  - Create media container
  - Publish container
  - Error handling
- [ ] TikTok Creator API:
  - OAuth flow
  - Upload video
  - Publish post
  - Error handling
- [ ] Implement `gagipress auth instagram` command
- [ ] Implement `gagipress auth tiktok` command
- [ ] Test pubblicazione manuale

**Deliverables**:
```bash
gagipress auth instagram    # OAuth flow
gagipress auth tiktok       # OAuth flow
gagipress calendar publish <id>  # Test manuale
```

**Files**:
- `cmd/auth/instagram.go` (complete implementation)
- `cmd/auth/tiktok.go` (complete implementation)
- `cmd/calendar/publish.go`
- `internal/social/instagram.go` (complete)
- `internal/social/tiktok.go` (complete)

---

### Week 3 Acceptance Criteria

- [ ] `calendar plan` genera piano settimanale intelligente
- [ ] `calendar approve` TUI funzionale e usabile
- [ ] Edge Function `auto-publish` deployata e testata
- [ ] OAuth flow Instagram e TikTok funzionanti
- [ ] Test pubblicazione manuale su entrambe piattaforme
- [ ] Cron job configurato e attivo
- [ ] Error handling robusto con retry logic
- [ ] Documentation aggiornata

---

## Week 4: Analytics & Correlation

### 4.1 Metrics Collection

**Steps**:
- [ ] Implement `sync-metrics` Edge Function:
  - Query post pubblicati ultimi 7 giorni
  - Per ogni post:
    - Call Instagram Insights API
    - Call TikTok Analytics API
    - Extract: views, likes, comments, shares, saves
    - Calculate engagement_rate
  - Salva in `post_metrics` table
  - Identify top performers
  - Configurare cron: `0 2 * * *`
- [ ] Implement `gagipress stats sync` command
  - Trigger manuale sync
  - Progress indicator

**Deliverables**:
```bash
gagipress stats sync
```

**Files**:
- `supabase/functions/sync-metrics/index.ts`
- `cmd/stats/sync.go`
- `internal/analytics/collector.go`

---

### 4.2 Analytics Dashboard (CLI)

**Steps**:
- [ ] Implement `gagipress stats show` command
  - Query metriche aggregate
  - Visualizzazione tabellare:
    - Total views, likes, engagement rate
    - Top performing posts
    - Growth trend (follower)
    - Content type breakdown
  - Charts in terminal (usando termui o simile)

**Deliverables**:
```bash
gagipress stats show
gagipress stats show --period 7d
gagipress stats show --period 30d
```

**Files**:
- `cmd/stats/show.go`
- `internal/analytics/dashboard.go`
- `internal/analytics/charts.go`

---

### 4.3 Sales Data Import

**Steps**:
- [ ] Implement `gagipress books sales` command
  - Supporta import da CSV (Amazon KDP report)
  - Parsing formato Amazon
  - Map ASIN → book_id
  - Insert/update `sales_data` table
  - Deduplication

**Deliverables**:
```bash
gagipress books sales import <csv-file>
gagipress books sales show <book-id>
```

**Files**:
- `cmd/books/sales.go`
- `internal/kdp/parser.go`
- `internal/repository/sales.go`

---

### 4.4 Correlation Analysis

**Steps**:
- [ ] Implement `gagipress stats correlate` command
  - Query temporal correlation:
    - Vendite spike 3-7 giorni post pubblicazione
  - Content performance score per categoria
  - Top ROI content types
  - Recommendations per ottimizzazione
  - Report visualizzazione CLI

**Deliverables**:
```bash
gagipress stats correlate
gagipress stats correlate --book <id>
```

**Files**:
- `cmd/stats/correlate.go`
- `internal/analytics/correlation.go`
- `internal/analytics/scoring.go`

---

### 4.5 Weekly Report

**Steps**:
- [ ] Implement `weekly-report` Edge Function:
  - Aggregate metriche settimana passata
  - Calcola insights e recommendations
  - Genera report markdown
  - Send via email (optional: SMTP integration)
  - Configurare cron: `0 9 * * 0` (Domenica 9am)

**Deliverables**:
- Report automatico ogni domenica

**Files**:
- `supabase/functions/weekly-report/index.ts`
- `supabase/functions/_shared/report.ts`

---

### Week 4 Acceptance Criteria

- [ ] `sync-metrics` Edge Function funzionante e schedulata
- [ ] `stats show` mostra dashboard comprensibile
- [ ] Import vendite Amazon KDP funzionante
- [ ] `stats correlate` produce insights utili
- [ ] Weekly report generato automaticamente
- [ ] Tutti cron jobs configurati e attivi
- [ ] Documentation completa
- [ ] End-to-end test completo sistema

---

## Week 5: Polish & Testing

### 5.1 Error Handling & Resilience

**Steps**:
- [ ] Audit error handling in tutti comandi
- [ ] Implement retry logic con exponential backoff
- [ ] Circuit breaker per API external
- [ ] Graceful degradation (OpenAI fail → Gemini)
- [ ] Logging strutturato (zerolog o zap)
- [ ] Error reporting mechanism

**Files**:
- `internal/errors/handler.go`
- `internal/errors/retry.go`
- `internal/logging/logger.go`

---

### 5.2 Testing Suite

**Steps**:
- [ ] Unit tests per:
  - Content generators
  - Scheduler logic
  - Analytics calculations
  - Repository operations
- [ ] Integration tests per:
  - Supabase operations
  - OpenAI API (mocked)
  - Edge Functions (local)
- [ ] E2E test:
  - Full workflow generate → schedule → publish (dry-run)
- [ ] Test coverage > 70%

**Files**:
- `*_test.go` files throughout
- `test/integration/`
- `test/e2e/`
- `test/mocks/`

---

### 5.3 Configuration & Secrets

**Steps**:
- [ ] Refactor config management
- [ ] Environment-based config (dev, prod)
- [ ] Secrets encryption at rest
- [ ] Validate config on startup
- [ ] Config migration tool

**Files**:
- `internal/config/validator.go`
- `internal/config/secrets.go`

---

### 5.4 CLI UX Improvements

**Steps**:
- [ ] Spinner/progress indicators per long operations
- [ ] Color coding per output (success=green, error=red)
- [ ] Confirmation prompts per destructive operations
- [ ] Auto-completion script (bash, zsh)
- [ ] `--dry-run` flag per comandi critici
- [ ] `--verbose` flag per debugging

---

### 5.5 Performance Optimization

**Steps**:
- [ ] Concurrent processing dove possibile
- [ ] Database query optimization (indici)
- [ ] HTTP client connection pooling
- [ ] Rate limiting rispetto API limits
- [ ] Caching (in-memory) per frequently accessed data

---

### Week 5 Acceptance Criteria

- [ ] Zero critical bugs
- [ ] Test coverage > 70%
- [ ] All error cases handled gracefully
- [ ] Performance acceptable (<2s per operazione locale)
- [ ] Logs strutturati e utili
- [ ] CLI UX polished
- [ ] Documentation completa e aggiornata

---

## Week 6: Deployment & Handoff

### 6.1 Documentation

**Steps**:
- [ ] Complete README.md:
  - Installation instructions
  - Quick start guide
  - Configuration guide
  - Command reference
- [ ] User guide (`docs/USER_GUIDE.md`)
- [ ] Architecture doc (`docs/ARCHITECTURE.md`)
- [ ] API integration guide (`docs/API_SETUP.md`)
- [ ] Troubleshooting guide (`docs/TROUBLESHOOTING.md`)
- [ ] Contributing guide (`CONTRIBUTING.md`)

---

### 6.2 Build & Release

**Steps**:
- [ ] Setup build pipeline (Makefile o task)
- [ ] Cross-compilation per OS (Darwin, Linux, Windows)
- [ ] Release automation (GitHub Actions o simile)
- [ ] Versioning strategy (semantic versioning)
- [ ] Changelog automation
- [ ] Binary distribution (GitHub Releases)

**Files**:
- `Makefile`
- `.github/workflows/release.yml`
- `scripts/build.sh`

---

### 6.3 Deployment Guide

**Steps**:
- [ ] Supabase setup guide step-by-step
- [ ] API keys acquisition guide:
  - OpenAI
  - Instagram Graph API
  - TikTok Creator API
- [ ] Initial data seeding guide
- [ ] Backup & restore procedures
- [ ] Monitoring setup (Supabase Dashboard)

**Files**:
- `docs/DEPLOYMENT.md`
- `docs/API_SETUP.md`

---

### 6.4 First Run Experience

**Steps**:
- [ ] `gagipress quickstart` command
  - Guided setup wizard
  - Tests all connections
  - Seeds sample book
  - Generates test content
  - Shows next steps
- [ ] Interactive tutorial mode

**Files**:
- `cmd/quickstart.go`

---

### 6.5 Handoff

**Steps**:
- [ ] Walktrough sistema completo
- [ ] Supervised first week usage
- [ ] Knowledge transfer
- [ ] Support plan
- [ ] Future roadmap discussion

---

### Week 6 Acceptance Criteria

- [ ] Documentation completa e chiara
- [ ] Binari compilati per Darwin/Linux/Windows
- [ ] Sistema deployed e funzionante
- [ ] First run experience fluida
- [ ] User confidence nel usare sistema
- [ ] Feedback incorporato
- [ ] Project pronto per uso produzione

---

## Success Metrics

### Technical

- [ ] Test coverage > 70%
- [ ] Zero critical bugs
- [ ] <2s response time comandi locali
- [ ] <10min Edge Function execution
- [ ] 99% uptime cron jobs

### Functional

- [ ] Genera 7-10+ content scripts in <5min
- [ ] 90%+ approval rate contenuti generati
- [ ] Pubblicazione automatica 100% success rate
- [ ] Metrics sync giornaliero funzionante
- [ ] Correlation analysis produce insights utili

### Business

- [ ] Sistema in uso quotidianamente
- [ ] Almeno 20 post pubblicati nel primo mese
- [ ] Metriche tracciate accuratamente
- [ ] Tempo risparmiato: 5+ ore/settimana

---

## Risk Mitigation

### Technical Risks

**Risk**: API rate limits hit
- **Mitigation**: Exponential backoff, queue system, monitoring

**Risk**: AI genera contenuti off-brand
- **Mitigation**: Template constraints, approval required, iterazione prompts

**Risk**: Edge Functions failure
- **Mitigation**: Retry logic, alerting, fallback manuale

### Timeline Risks

**Risk**: OAuth setup blocca sviluppo
- **Mitigation**: Mock API durante sviluppo, OAuth in parallelo

**Risk**: Supabase learning curve
- **Mitigation**: Tutorials, documentation, support channels

**Risk**: Scope creep
- **Mitigation**: Strict MVP scope, future features in backlog

---

## Daily Checklist

**During Implementation**:
- [ ] Commit atomici con messaggi descrittivi
- [ ] Tests per ogni feature
- [ ] Documentation inline (godoc)
- [ ] Update questo piano con status
- [ ] Daily progress tracking

**End of Each Week**:
- [ ] Week acceptance criteria review
- [ ] Demo features completate
- [ ] Retrospective: cosa funziona, cosa no
- [ ] Plan next week adjustments

---

## Current Status

**Week**: 1
**Day**: 1
**Current Task**: 1.3 API Integrations Skeleton
**Completed**:
  - 1.1 Go Project Setup ✅
  - 1.2 Supabase Project Setup ✅
**Blockers**: None
**Next**: Implement OpenAI client wrapper, test Gemini automation

---

## Notes

- Mantenere approccio iterativo: ogni settimana deve produrre qualcosa funzionante
- Priorità a MVP scope, future features in backlog
- Testing e documentation in parallelo allo sviluppo
- User feedback incorporato continuamente
