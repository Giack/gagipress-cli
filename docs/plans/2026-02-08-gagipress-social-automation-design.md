# Gagipress Social Automation System - Design Document

**Date**: 2026-02-08
**Author**: Design Session with Claude
**Status**: Approved for Implementation

## Executive Summary

Sistema di automazione social media per Gagipress (casa editrice self-publishing su Amazon KDP) che genera, schedula e pubblica 7-10+ contenuti settimanali su TikTok e Instagram Reels, con analytics avanzate e correlazione vendite.

**Tech Stack**: Go CLI + Supabase (PostgreSQL + Edge Functions) + AI APIs (OpenAI + Gemini)

---

## 1. Context & Requirements

### Business Context
- **Azienda**: Gagipress - casa editrice self-publishing
- **Catalogo**:
  - Libri per bambini
  - Enigmistica per adulti (anche in dialetto milanese)
  - Libri per il risparmio
- **Piattaforme**: TikTok/Instagram Reels (primario), Instagram feed (secondario)
- **Profili**: Già attivi con follower esistenti

### Obiettivi
1. **Crescita**: Aumentare follower e visibilità brand
2. **Conversione**: Trasformare follower in acquirenti Amazon KDP
3. **Community**: Costruire base lettori fedeli e coinvolti

### Scope Automazione
- ✅ **A**: Generazione idee contenuti e script per Reels/TikTok
- ✅ **C**: Schedulazione e pubblicazione automatica
- ✅ **E**: Analisi performance e ottimizzazione
- 🔮 **B**: Creazione/editing video (Fase 2)

### Requisiti Tecnici
- Self-hosted/open source preferred
- Deployable su Vercel + Supabase
- Cron jobs per scheduling automatico
- 7-10+ post/settimana con alta automazione
- Correlazione vendite Amazon KDP

---

## 2. Architecture Overview

### Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Layer 1: CLI Locale                   │
│                  (Go Application)                        │
│  • Generazione contenuti                                 │
│  • Approvazione calendar                                 │
│  • Visualizzazione analytics                             │
│  • Browser automation (Gemini)                           │
└─────────────────┬───────────────────────────────────────┘
                  │ REST API
                  ▼
┌─────────────────────────────────────────────────────────┐
│              Layer 2: Database Cloud                     │
│                   (Supabase)                             │
│  • PostgreSQL: contenuti, metriche, calendar, libri     │
│  • Storage: media files                                  │
│  • Row Level Security                                    │
└─────────────────┬───────────────────────────────────────┘
                  │ Triggers
                  ▼
┌─────────────────────────────────────────────────────────┐
│          Layer 3: Automazioni Serverless                 │
│            (Supabase Edge Functions)                     │
│  • Cron: scarica metriche (daily)                       │
│  • Cron: pubblica contenuti (hourly)                    │
│  • Cron: genera report (weekly)                         │
└─────────────────────────────────────────────────────────┘
```

### Data Flow
1. CLI genera contenuti → Salva Supabase
2. Utente approva piano settimanale → Marca "approved"
3. Cron job pubblica contenuti schedulati → API Instagram/TikTok
4. Cron job raccoglie metriche → Salva time-series
5. Analytics correla performance → Insights
6. Ciclo ricomincia con contenuti ottimizzati

---

## 3. CLI Structure & Commands

### Tool Name
`gagipress` (alias: `ggp`)

### Command Tree

```bash
gagipress
├── init                    # Setup Supabase, API keys
├── auth                    # Autentica social accounts
├── generate
│   ├── ideas              # Genera 20-30 idee contenuti
│   ├── script <id>        # Crea script da idea
│   └── batch              # Genera script settimana completa
├── calendar
│   ├── plan               # Crea piano settimanale intelligente
│   ├── show               # Visualizza calendario
│   ├── approve            # Approva/modifica contenuti
│   └── publish            # Forza pubblicazione immediata
├── stats
│   ├── sync               # Scarica dati social
│   ├── show               # Dashboard metriche
│   └── correlate          # Analizza social→vendite
├── books
│   ├── add                # Aggiungi libro a tracking
│   └── sales              # Importa dati vendite Amazon
├── deploy
│   └── functions          # Deploy Edge Functions
└── logs                   # Visualizza logs sistema
```

### Weekly Workflow

**Lunedì mattina**:
```bash
gagipress generate batch
```
Genera idee e script per 7-10 post settimanali

**Lunedì pomeriggio**:
```bash
gagipress calendar plan
```
Sistema crea piano ottimizzato (best times, content mix)

**Review & Approve**:
```bash
gagipress calendar approve
```
Interfaccia interattiva per approvare/modificare/rigettare

**Automatico Durante Settimana**:
Cron jobs pubblicano contenuti approvati agli orari schedulati

**Domenica**:
```bash
gagipress stats correlate
```
Analisi performance e correlazione vendite

---

## 4. Content Generation Pipeline

### Phase 1: Idea Generation (Bulk)

**Trigger**: `gagipress generate ideas`

**Process**:
1. Legge catalogo libri KDP dal database
2. Analizza top performing posts storici
3. Genera 20-30 idee usando AI:
   - OpenAI API (primary, qualità alta)
   - Gemini browser automation (fallback, gratuito)
4. Categorizza idee:
   - Educational (spiega valore libro)
   - Entertainment (challenge, quiz)
   - Behind-the-scenes (processo creazione)
   - User-generated (testimonianze, unboxing)
   - Trend-jacking (cavalca trend TikTok)
5. Assegna relevance score (0-100)
6. Salva in `content_ideas` table

### Phase 2: Script Generation

**Trigger**: `gagipress generate script <idea-id>`

**Output per ogni script**:
- **Hook** (primi 3 secondi - critico per retention!)
- **Full script** (15-60 secondi)
- **Call-to-action** (link Amazon, follow, commenta)
- **Hashtags** ottimizzati (#BookTok, #enigmisticamilanese)
- **Visual notes** (inquadrature, testi sovraimpressi)
- **Audio suggestion** (trending sounds, voiceover)

### Phase 3: Niche Personalization

**Template specifici per categoria**:

**Libri Bambini**:
- Target: genitori 25-40
- Focus: educational value, preview illustrazioni
- Tone: caldo, rassicurante, pedagogico

**Enigmistica Dialetto Milanese**:
- Target: milanesi, nostalgici, dialect lovers
- Focus: humor locale, sfide interattive
- Tone: scherzoso, proud, comunitario

**Libri Risparmio**:
- Target: famiglie, budget-conscious
- Focus: tips pratici, problem→solution
- Tone: empatico, pratico, relatable

### Contextual Intelligence

- **KDP Data**: Prioritizza contenuti per libri best-seller
- **Trend Monitoring**: API TikTok per trending hashtags/sounds
- **Tone Variation**: Evita ripetitività con diversi angles
- **Seasonal**: Adatta a festività, back-to-school, etc.

---

## 5. Scheduling & Publishing

### Intelligent Weekly Planning

**Trigger**: `gagipress calendar plan`

**Algorithm considera**:
1. **Best Posting Times**: Analytics storiche → quando follower attivi
2. **Content Mix**: Bilancia tipi (non 5 enigmi di fila)
3. **Book Rotation**: Spotlight equo tra libri catalogo
4. **Trend Timing**: Contenuti trend vanno schedulati ASAP (finestra stretta)
5. **Platform Optimization**: Orari diversi per TikTok vs Instagram

**Output**: Piano 7-10 post con `scheduled_for` timestamp ottimale

### Approval & Modification

**Trigger**: `gagipress calendar approve`

**Interactive Interface**:
```
📅 Lun 10 Feb - 18:30 [PEAK TIME]
   📹 "3 Enigmi in Milanese che NESSUNO risolve"
   📚 Libro: Enigmistica Milanese Vol.2
   ⏱️  45 secondi | 🎵 Trending Sound #1234

   Hook: "Te see bon de risolv sti tre indovinei?"
   Script: [mostra preview]
   Hashtags: #BookTok #Milano #Dialetto #Enigmistica

   [A]pprova  [E]dita  [R]ifiuta  [S]posta  [P]review

→
```

**Actions disponibili**:
- **Approve**: Marca `status = approved`
- **Edit**: Modifica script inline, salva nuova versione
- **Reject**: `status = rejected`, genera replacement
- **Reschedule**: Cambia `scheduled_for`
- **Preview**: Mostra come apparirà pubblicato
- **Batch approve**: Approva multipli in blocco

### Automatic Publishing

**Edge Function**: `auto-publish`
**Cron**: `0 * * * *` (ogni ora)

**Process**:
```sql
SELECT * FROM content_calendar
WHERE status = 'approved'
  AND scheduled_for <= NOW()
  AND published_at IS NULL
ORDER BY scheduled_for ASC
```

Per ogni contenuto:
1. Recupera script, media, metadata
2. Chiama API Instagram/TikTok:
   - Upload video (se presente)
   - Set caption, hashtags, location
   - Publish
3. Salva `post_url`, `published_at`
4. Marca `status = published`

**Error Handling**:
- 3 retry con exponential backoff (1min, 5min, 15min)
- Log errore dettagliato in `publish_errors` table
- Email alert dopo 3 failures
- Marca `status = failed` per review manuale

---

## 6. Analytics & Optimization

### Automatic Metrics Collection

**Edge Function**: `sync-metrics`
**Cron**: `0 2 * * *` (daily 2am)

**Process**:
1. Query tutti post pubblicati ultimi 7 giorni
2. Per ogni post, chiama API social:
   - Instagram Graph API: insights endpoint
   - TikTok Business API: video analytics
3. Salva in `post_metrics`:
   - `views`, `likes`, `comments`, `shares`, `saves`
   - `watch_time_percentage` (retention critica)
   - `follower_growth` (net gain quel giorno)
4. Calcola `engagement_rate`:
   ```
   (likes + comments + shares) / views * 100
   ```
5. Identifica top 20% performers → tag `is_top_performer = true`

**Time-Series Storage**: Ogni sync crea nuovo record per tracciare evoluzione metriche nel tempo

### Social → KDP Sales Correlation

**Trigger**: `gagipress stats correlate`

**The Gold Question**: Quali contenuti vendono libri?

**Analysis Process**:

1. **Temporal Correlation**:
   ```sql
   -- Spike vendite 3-7 giorni post pubblicazione?
   SELECT
     c.id, c.script, c.published_at,
     SUM(s.units_sold) as sales_spike
   FROM content_calendar c
   JOIN books b ON c.script.book_id = b.id
   JOIN sales_data s ON s.book_id = b.id
   WHERE s.date BETWEEN c.published_at AND c.published_at + INTERVAL '7 days'
   GROUP BY c.id
   ```

2. **Book Mention Attribution**:
   - Quali libri menzionati nei top performing posts?
   - Cross-reference con vendite quel libro

3. **Link Tracking**:
   - UTM parameters su link bio Amazon
   - Click tracking (se disponibile da social APIs)

4. **Comment Analysis**:
   - Scrape commenti contenenti "comprato", "dove trovo", "link?"
   - Indicatore di purchase intent

**Output: Content Performance Score**

```
📊 Content Type Performance (ultimo mese)

Enigmi Challenge
   Views: 125K | Engagement: 12.3% | Vendite stimate: 23
   ⭐⭐⭐⭐⭐ ROI: EXCELLENT

Behind-the-Scenes
   Views: 45K | Engagement: 8.1% | Vendite stimate: 5
   ⭐⭐⭐ ROI: MODERATE

Tips Risparmio
   Views: 89K | Engagement: 10.8% | Vendite stimate: 18
   ⭐⭐⭐⭐ ROI: GOOD

Dialetto Milanese 🔥
   Views: 156K | Engagement: 15.2% | Vendite stimate: 31
   ⭐⭐⭐⭐⭐ ROI: OUTSTANDING
```

### Automatic Optimization

Sistema usa insights per:

1. **Content Mix Adjustment**:
   - Se "Dialetto Milanese" performa 2x meglio → genera 40% idee dialetto invece di 20%

2. **Timing Optimization**:
   - Se post Educational performano meglio 10am vs 6pm → reschedula

3. **Trend Detection**:
   - Se engagement cala 20% ultimi 7 giorni → alert
   - Suggerisce: "Prova più contenuti tipo X che performava bene"

4. **Hashtag Optimization**:
   - Traccia quali hashtag correlano con reach
   - Auto-suggerisce hashtag mix ottimale

**Weekly Report Email**:
- Top 3 performing posts
- Content type raccomandations
- Best time slots discovered
- Action items

---

## 7. Database Schema

### Core Tables

#### `books`
Catalogo libri KDP tracciati

```sql
CREATE TABLE books (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title TEXT NOT NULL,
  genre TEXT NOT NULL,
  target_audience TEXT,
  kdp_asin TEXT UNIQUE,
  cover_image_url TEXT,
  publication_date DATE,
  current_rank INTEGER,
  total_sales INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_books_genre ON books(genre);
CREATE INDEX idx_books_asin ON books(kdp_asin);
```

#### `content_ideas`
Idee generate da AI

```sql
CREATE TABLE content_ideas (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  type TEXT NOT NULL, -- educational, entertainment, bts, ugc, trend
  brief_description TEXT NOT NULL,
  relevance_score INTEGER CHECK (relevance_score >= 0 AND relevance_score <= 100),
  book_id UUID REFERENCES books(id),
  status TEXT DEFAULT 'pending', -- pending, approved, rejected, scripted
  generated_at TIMESTAMPTZ DEFAULT NOW(),
  metadata JSONB -- extra context
);

CREATE INDEX idx_ideas_status ON content_ideas(status);
CREATE INDEX idx_ideas_score ON content_ideas(relevance_score DESC);
CREATE INDEX idx_ideas_book ON content_ideas(book_id);
```

#### `content_scripts`
Script completi pronti per produzione

```sql
CREATE TABLE content_scripts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  idea_id UUID REFERENCES content_ideas(id),
  hook TEXT NOT NULL,
  full_script TEXT NOT NULL,
  cta TEXT NOT NULL,
  hashtags TEXT[] NOT NULL,
  visual_notes TEXT,
  audio_suggestion TEXT,
  estimated_duration INTEGER, -- secondi
  status TEXT DEFAULT 'draft', -- draft, approved, used
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_scripts_idea ON content_scripts(idea_id);
CREATE INDEX idx_scripts_status ON content_scripts(status);
```

#### `content_calendar`
Piano pubblicazione e tracking

```sql
CREATE TABLE content_calendar (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  script_id UUID REFERENCES content_scripts(id),
  scheduled_for TIMESTAMPTZ NOT NULL,
  platform TEXT NOT NULL, -- instagram, tiktok
  post_type TEXT NOT NULL, -- reel, story, feed
  status TEXT DEFAULT 'pending_approval',
    -- pending_approval, approved, published, failed
  approved_at TIMESTAMPTZ,
  published_at TIMESTAMPTZ,
  post_url TEXT,
  publish_errors JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_calendar_scheduled ON content_calendar(scheduled_for);
CREATE INDEX idx_calendar_status ON content_calendar(status);
CREATE INDEX idx_calendar_platform ON content_calendar(platform);
```

#### `post_metrics`
Performance metriche (time-series)

```sql
CREATE TABLE post_metrics (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  calendar_id UUID REFERENCES content_calendar(id),
  platform TEXT NOT NULL,
  post_url TEXT,
  views INTEGER DEFAULT 0,
  likes INTEGER DEFAULT 0,
  comments INTEGER DEFAULT 0,
  shares INTEGER DEFAULT 0,
  saves INTEGER DEFAULT 0,
  engagement_rate DECIMAL(5,2),
  watch_time_percentage DECIMAL(5,2),
  follower_growth INTEGER DEFAULT 0,
  is_top_performer BOOLEAN DEFAULT FALSE,
  scraped_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_metrics_calendar ON post_metrics(calendar_id);
CREATE INDEX idx_metrics_scraped ON post_metrics(scraped_at DESC);
CREATE INDEX idx_metrics_top ON post_metrics(is_top_performer) WHERE is_top_performer = TRUE;
```

#### `sales_data`
Vendite Amazon KDP

```sql
CREATE TABLE sales_data (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  book_id UUID REFERENCES books(id),
  date DATE NOT NULL,
  units_sold INTEGER DEFAULT 0,
  revenue DECIMAL(10,2),
  royalty DECIMAL(10,2),
  source TEXT DEFAULT 'amazon_reports',
  imported_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(book_id, date)
);

CREATE INDEX idx_sales_book ON sales_data(book_id);
CREATE INDEX idx_sales_date ON sales_data(date DESC);
```

### Relationships

```
books (1) ──→ (N) content_ideas
content_ideas (1) ──→ (N) content_scripts
content_scripts (1) ──→ (N) content_calendar
content_calendar (1) ──→ (N) post_metrics
books (1) ──→ (N) sales_data
```

### Key Queries

**Top performing content for a book**:
```sql
SELECT cs.*, pm.engagement_rate, pm.views
FROM content_scripts cs
JOIN content_calendar cc ON cc.script_id = cs.id
JOIN post_metrics pm ON pm.calendar_id = cc.id
JOIN content_ideas ci ON ci.id = cs.idea_id
WHERE ci.book_id = $1
  AND pm.is_top_performer = TRUE
ORDER BY pm.engagement_rate DESC;
```

**Sales spike after post**:
```sql
SELECT
  cc.id,
  cc.published_at,
  SUM(sd.units_sold) as units_sold_7days
FROM content_calendar cc
JOIN content_scripts cs ON cs.id = cc.script_id
JOIN content_ideas ci ON ci.id = cs.idea_id
JOIN sales_data sd ON sd.book_id = ci.book_id
WHERE sd.date BETWEEN cc.published_at::date
  AND (cc.published_at + INTERVAL '7 days')::date
GROUP BY cc.id, cc.published_at
ORDER BY units_sold_7days DESC;
```

---

## 8. Deployment & Operations

### Initial Setup (One-Time)

#### 1. Supabase Project Creation

```bash
# User creates project on supabase.com (free tier)
# Gets: SUPABASE_URL, SUPABASE_ANON_KEY, SUPABASE_SERVICE_KEY

gagipress init
# Interactive wizard:
# - Prompts for Supabase credentials
# - Runs migration SQL to create schema
# - Sets up RLS policies
# - Creates Storage bucket "media"
# - Saves config to ~/.gagipress/config.yaml
```

#### 2. API Keys Configuration

```bash
gagipress auth
# Guides OAuth flow for:
# - Instagram Business Account
# - TikTok Business Account
# Prompts for:
# - OpenAI API key (or skip for Gemini-only)
# - Amazon KDP credentials (optional, for sales import)
# Encrypts and stores in Supabase Vault
```

#### 3. Edge Functions Deployment

```bash
gagipress deploy functions
# Uploads 3 Edge Functions to Supabase:
# - sync-metrics/index.ts
# - auto-publish/index.ts
# - weekly-report/index.ts
# Configures cron triggers
```

### Cron Jobs Configuration

**Function**: `sync-metrics`
```yaml
Schedule: 0 2 * * *  # Daily 2am UTC
Purpose: Fetch social metrics, calculate engagement
Timeout: 5 minutes
Retry: 3 attempts
```

**Function**: `auto-publish`
```yaml
Schedule: 0 * * * *  # Every hour
Purpose: Publish approved scheduled content
Timeout: 2 minutes
Retry: 3 attempts (per post)
```

**Function**: `weekly-report`
```yaml
Schedule: 0 9 * * 0  # Sunday 9am UTC
Purpose: Generate performance report, send email
Timeout: 3 minutes
Retry: 1 attempt
```

### Secrets Management

**Storage**: Supabase Vault (encrypted at rest)

**Secrets**:
- `OPENAI_API_KEY`
- `INSTAGRAM_ACCESS_TOKEN`
- `TIKTOK_ACCESS_TOKEN`
- `AMAZON_KDP_EMAIL`
- `AMAZON_KDP_PASSWORD`
- `SMTP_CONFIG` (for email alerts)

**Access**: Edge Functions fetch via `vault.get(secret_name)`

### Hybrid Architecture

**Local (CLI)**:
- Interactive operations: generate, approve, view
- Heavy operations: browser automation with Gemini
- Connection: Supabase REST API (authenticated with anon key + RLS)

**Cloud (Supabase)**:
- Database: Always accessible, single source of truth
- Storage: Media files (future: generated videos)
- Edge Functions: Time-based automation

**Network Flow**:
```
CLI ←→ [Internet] ←→ Supabase API Gateway ←→ PostgreSQL
                              ↓
                       Edge Functions (cron)
                              ↓
                    External APIs (Instagram, TikTok)
```

### Monitoring & Logging

**Edge Function Logs**:
- Visible in Supabase Dashboard → Edge Functions → Logs
- Real-time streaming
- Retention: 7 days

**CLI Logs**:
```bash
gagipress logs --function sync-metrics --last 24h
gagipress logs --function auto-publish --errors-only
```

**Alerts**:
- Email when cron job fails 3x consecutive
- Email weekly report (performance summary)
- Slack webhook (optional)

### Cost Estimation

**Supabase Free Tier** (more than enough):
- 500 MB database space
- 1 GB file storage
- 500K Edge Function invocations/month
- 50K monthly active users
- Unlimited API requests

**Estimated Usage**:
- Database: ~50 MB (thousands of posts)
- Storage: ~200 MB (if storing media)
- Edge Functions: ~10K invocations/month
  - auto-publish: 24/day * 30 = 720
  - sync-metrics: 30/month = 30
  - weekly-report: 4/month = 4
  - Total: ~754 + overhead

**External API Costs**:
- OpenAI API: ~$5-10/month (20-30 content generations/week)
- Instagram/TikTok APIs: Free (business accounts)
- Gemini fallback: Free via browser automation

**Total Monthly Cost**: $5-10 (OpenAI only, Supabase free)

### Scaling Considerations

**When to upgrade Supabase**:
- \> 500 MB database (unlikely for years)
- \> 1 GB media storage (if storing all videos locally)
- \> 500K Edge Function calls (very unlikely)

**Horizontal Scaling**:
- Multiple brands: Add `brand_id` column, partition data
- Multiple users: Implement team collaboration, share approvals
- High volume: Supabase scales automatically with Pro plan

---

## 9. Future Roadmap

### Phase 2: Video Editing Automation

**Command**: `gagipress render <script-id>`

**Capabilities**:
- Template-based video generation
- FFmpeg compositing: intro, text overlays, outro
- Text-to-speech voiceover (ElevenLabs API or Coqui local)
- Auto-subtitles (Whisper API or local model)
- Stock footage integration for B-roll
- Preview before approval

**Template Types**:
- Enigma reveal (question → think → answer)
- Book showcase (cover → flip pages → CTA)
- Tips format (problem → tips list → CTA)
- Testimonial (quote → book cover → link)

**Storage**: Generated videos in Supabase Storage, auto-cleanup after 30 days

### Phase 3: Interaction Automation

**Auto-Reply Comments**:
- AI categorizes comments (question, praise, request, spam)
- Template replies for common questions
- Flag important comments for manual response

**DM Automation**:
- Auto-respond to common DM requests ("Where to buy?")
- Link to Amazon, provide info
- Human handoff for complex queries

**Community Management**:
- Identify most active/engaged followers
- Suggest to feature in content or send free book
- Build VIP community

### Phase 4: Multi-Brand Support

**Features**:
- `gagipress brand add <name>`
- Switch contexts: `gagipress use <brand>`
- Separate database rows per brand
- Team collaboration: multiple users approve

**Use Case**: Manage multiple imprints or client accounts

### Phase 5: Advanced Analytics

**A/B Testing**:
- Auto-generate 2 versions same content (different hooks)
- Publish both, compare performance
- Learn which angles work better

**Predictive Analytics**:
- ML model trained on historical data
- Predicts engagement before publishing
- Suggests: "This script likely underperforms, regenerate?"

**Competitor Analysis**:
- Monitor similar accounts (book publishers, enigmistica)
- Track their top content
- Suggest: "Competitor X got 200K views with [topic], try similar?"

### Phase 6: Cross-Platform Expansion

**Additional Platforms**:
- Facebook (groups, pages)
- X/Twitter (threads, engagement)
- LinkedIn (for B2B books, educational)
- YouTube Shorts (similar to Reels)

**Platform-Specific Optimization**:
- Auto-adapt content format/length
- Different hashtag strategies
- Cross-post with platform-specific tweaks

---

## 10. Implementation Plan

### MVP Scope (Phase 1)

**Core Features**:
- ✅ CLI with generate, calendar, stats commands
- ✅ AI content generation (OpenAI + Gemini fallback)
- ✅ Supabase database schema
- ✅ Manual approval workflow
- ✅ Auto-publish Edge Function
- ✅ Metrics collection Edge Function
- ✅ Basic analytics dashboard (CLI)

**Out of Scope (MVP)**:
- ❌ Video generation (Phase 2)
- ❌ Auto-replies (Phase 3)
- ❌ Web UI (future, CLI-first)
- ❌ Multi-brand (Phase 4)

### Development Phases

**Week 1: Foundation**
- Go CLI skeleton with Cobra
- Supabase project setup, schema migration
- API integrations (OpenAI, Instagram, TikTok)

**Week 2: Content Generation**
- `generate ideas` command
- `generate script` command
- Niche-specific templates

**Week 3: Scheduling & Publishing**
- `calendar plan` algorithm
- `calendar approve` interactive UI
- `auto-publish` Edge Function

**Week 4: Analytics**
- `sync-metrics` Edge Function
- `stats show` dashboard
- `stats correlate` logic

**Week 5: Polish & Testing**
- Error handling, retry logic
- Logging, monitoring
- Documentation
- End-to-end testing

**Week 6: Deployment & Handoff**
- Deploy Edge Functions
- User onboarding guide
- First week supervised use

---

## 11. Success Metrics

**Automation Success**:
- [ ] Generate 7-10 content scripts in <5 minutes
- [ ] 90%+ approval rate on generated content
- [ ] Zero manual intervention for publishing (after approval)
- [ ] Daily metrics sync with 99% reliability

**Business Success**:
- [ ] 20%+ follower growth/month
- [ ] 10%+ engagement rate average
- [ ] Measurable correlation social spike → sales spike
- [ ] 5+ hours/week saved on social media management

**Technical Success**:
- [ ] <2 second CLI response time (local operations)
- [ ] <10 minute Edge Function execution time
- [ ] Zero data loss
- [ ] Stay within Supabase free tier

---

## 12. Risk Mitigation

**Risk: API Rate Limits**
- **Instagram**: 200 calls/hour → batch operations
- **TikTok**: 100 calls/day → spread across day
- **Mitigation**: Exponential backoff, queue system

**Risk: AI Hallucinations/Poor Content**
- **Problem**: Generated content off-brand or factually wrong
- **Mitigation**: Template constraints, human approval required, A/B testing

**Risk: Account Bans**
- **Problem**: Social platforms ban for automation
- **Mitigation**: Use official APIs only, respect rate limits, gradual ramp-up

**Risk: Data Loss**
- **Problem**: Supabase downtime or accidental deletion
- **Mitigation**: Daily backups (pg_dump), RLS prevents accidents

**Risk: Budget Overrun**
- **Problem**: OpenAI costs spike unexpectedly
- **Mitigation**: Cost caps in code, fallback to Gemini, monitoring

---

## Conclusion

Gagipress Social Automation System è un'architettura hybrid Go CLI + Supabase che bilancia automazione aggressiva con controllo umano strategico.

**Punti di forza**:
- ✅ Self-hosted, zero recurring costs (oltre AI APIs)
- ✅ Scalabile da 1 a 100+ post/settimana
- ✅ Data-driven optimization loop
- ✅ Low maintenance (cron jobs autonomi)

**Prossimi Step**:
1. Setup git worktree per sviluppo isolato
2. Creare piano implementazione dettagliato
3. Iniziare sviluppo Week 1 (Foundation)

---

**Document Status**: ✅ Approved
**Ready for Implementation**: Yes
**Estimated MVP Timeline**: 5-6 weeks
