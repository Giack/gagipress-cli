# Piano di Implementazione: Gagipress Autonomo per Aumentare Vendite

> **Data**: 2026-02-21
> **Obiettivo**: Trasformare gagipress da "generatore di script" a "macchina autonoma di vendita"
> **Stato attuale**: Il pipeline si blocca al calendario (fake save). Zero publishing. Zero tracking vendite.

---

## Phase 0: Fonti e API Scoperte

### Codebase Audit — Risultati Chiave

| Componente | File | Stato | Problema |
|---|---|---|---|
| `calendar plan` save | `cmd/calendar/plan.go:87-96` | FAKE | Loop conta senza salvare |
| `calendar show` | `cmd/calendar/show.go:43-59` | PLACEHOLDER | Slice vuota hardcoded |
| `calendar approve` | `cmd/calendar/approve.go` | FUNZIONA | Ma non trova dati (nessun save) |
| Planner PostType | `internal/scheduler/planner.go:71-75` | BUG | `PostType` mai impostato → validation fail |
| CTA prompt | `internal/prompts/templates.go:155` | GENERICO | "link in bio" senza URL Amazon |
| Book ASIN | `internal/models/book.go:13` | DISPONIBILE | `KDPASIN` esiste ma non passato al prompt |
| Script gen → prompt | `cmd/generate/script.go:93` | PARZIALE | Solo `book.Title` estratto, non ASIN |
| Instagram client | `internal/social/instagram.go` | STUB | Tutti i metodi ritornano "not implemented" |
| TikTok client | `internal/social/tiktok.go` | STUB | Tutti i metodi ritornano "not implemented" |
| Batch script gen | N/A | MANCANTE | Non esiste `cmd/generate/batch.go` |
| Bluesky | N/A | MANCANTE | Zero riferimenti nel codebase |
| Media generation | N/A | MANCANTE | Zero codice per immagini/video |

### API e Metodi Disponibili (Verified)

**CalendarRepository** (`internal/repository/calendar.go`):
```go
CreateEntry(input *models.ContentCalendarInput) (*models.ContentCalendar, error)
GetEntries(status string, limit int) ([]models.ContentCalendar, error)
UpdateEntryStatus(id string, status string) error
DeleteEntry(id string) error
```

**ContentRepository** (`internal/repository/content.go`):
```go
GetIdeas(status string, limit int) ([]ContentIdea, error)
GetIdeaByIDPrefix(prefix string) (*ContentIdea, error)
UpdateIdeaStatus(id, status string) error
GetScripts(limit int) ([]ContentScript, error)
CreateScript(input *ContentScriptInput) (*ContentScript, error)
```

**ScriptGenerator** (`internal/generator/scripts.go`):
```go
GenerateScript(idea *models.ContentIdea, bookTitle, platform string) (*GeneratedScript, error)
SaveScript(script *GeneratedScript, ideaID string) (*models.ContentScript, error)
```

**Prompt Template** (`internal/prompts/templates.go:105`):
```go
func ScriptPromptTemplate(idea, bookTitle, platform string) string
// Parametri attuali: idea description, book title, platform
// Mancante: ASIN, Amazon URL
```

### Anti-Pattern da Evitare
- NON inventare metodi repository che non esistono
- NON modificare la signature di `ScriptPromptTemplate` senza aggiornare tutti i caller
- NON usare `GetIdeas("", 0)` per batch — filtrare per status lato server
- NON saltare `PostType` nella validazione calendar — è REQUIRED ("reel", "story", "feed")
- NON hardcodare platform nel planner — rispettare la platform scelta nello script

---

## Phase 1: Fix Calendar Pipeline (P0 — Quick Win)

**Impatto**: Sblocca l'intero workflow ideas→scripts→calendario→review
**Effort**: ~30min
**File da modificare**: 3

### Task 1.1: Fix `PlanWeek` — Aggiungere PostType

**File**: `internal/scheduler/planner.go:71-75`

**Problema**: `ContentCalendarInput.PostType` mai impostato. La validazione richiede `"reel"`, `"story"`, o `"feed"`.

**Fix**: Aggiungere `PostType: "reel"` (default per TikTok/Instagram Reels content)

```go
// PRIMA (linea 71-75):
entry := &models.ContentCalendarInput{
    ScriptID:     &script.ID,
    ScheduledFor: slot.Time,
    Platform:     platform,
}

// DOPO:
entry := &models.ContentCalendarInput{
    ScriptID:     &script.ID,
    ScheduledFor: slot.Time,
    Platform:     platform,
    PostType:     "reel", // Default per contenuti video brevi
}
```

### Task 1.2: Fix `plan.go` — Salvare realmente nel DB

**File**: `cmd/calendar/plan.go:87-96`

**Fix**: Sostituire il loop finto con chiamate a `calendarRepo.CreateEntry()`.

```go
// PRIMA (fake save):
savedCount := 0
for range calendarEntries {
    savedCount++
}

// DOPO (real save):
calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
savedCount := 0
for _, entry := range calendarEntries {
    if err := entry.Validate(); err != nil {
        fmt.Printf("⚠️  Skipping invalid entry: %v\n", err)
        continue
    }
    _, err := calendarRepo.CreateEntry(entry)
    if err != nil {
        fmt.Printf("⚠️  Failed to save entry: %v\n", err)
        continue
    }
    savedCount++
}
```

**Nota**: Bisogna aggiungere l'import di `repository` al file e creare il `calendarRepo` con `cfg.Supabase`.

### Task 1.3: Fix `show.go` — Leggere dal DB

**File**: `cmd/calendar/show.go:36-59`

**Fix**: Rimuovere il tipo `CalendarEntry` locale e la slice hardcoded. Usare `calendarRepo.GetEntries()`.

```go
// PRIMA (placeholder):
type CalendarEntry struct { ... }
entries := []CalendarEntry{}

// DOPO:
calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
entries, err := calendarRepo.GetEntries(statusFilter, 0)
if err != nil {
    return fmt.Errorf("failed to get entries: %w", err)
}
```

**Nota**: Le referenze successive a `entry.ScheduledFor`, `entry.Platform`, `entry.Status`, `entry.ScriptID` rimangono valide — i campi sono gli stessi nel modello `ContentCalendar`.

### Verifica Phase 1

```bash
# 1. Build
make build

# 2. Creare un piano
bin/gagipress calendar plan --days 7 --posts 1

# 3. Verificare che show mostra i dati
bin/gagipress calendar show

# 4. Verificare che approve trova i pending
bin/gagipress calendar approve

# 5. Verificare nel DB
# (via Supabase MCP) SELECT * FROM content_calendar ORDER BY scheduled_for;
```

---

## Phase 2: CTA con Link Amazon (P1 — Impatto Vendite Diretto)

**Impatto**: Ogni script generato includerà un link Amazon cliccabile = conversione diretta
**Effort**: ~30min
**File da modificare**: 3

### Task 2.1: Estendere `ScriptPromptTemplate` con ASIN

**File**: `internal/prompts/templates.go:105`

**Fix**: Aggiungere parametro `amazonURL` alla funzione e iniettarlo nel prompt CTA.

```go
// PRIMA:
func ScriptPromptTemplate(idea, bookTitle, platform string) string {

// DOPO:
func ScriptPromptTemplate(idea, bookTitle, platform, amazonURL string) string {
```

Nella sezione CTA del prompt (linea 152-155):
```go
// PRIMA:
**CTA (5-10 secondi)**
- Invito all'azione chiaro
- Perché dovrebbero comprare il libro
- Dove trovarlo (link in bio)

// DOPO:
**CTA (5-10 secondi)**
- Invito all'azione chiaro
- Perché dovrebbero comprare il libro
- Link diretto: %s
- Menziona "Link in bio" E il link diretto Amazon
```

### Task 2.2: Passare ASIN nel flow di generazione

**File**: `cmd/generate/script.go:90-104`

```go
// PRIMA (linea 93-96):
bookTitle := "Your Book"
if idea.BookID != nil {
    book, err := booksRepo.GetByID(*idea.BookID)
    if err == nil {
        bookTitle = book.Title
    }
}

// DOPO:
bookTitle := "Your Book"
amazonURL := ""
if idea.BookID != nil {
    book, err := booksRepo.GetByID(*idea.BookID)
    if err == nil {
        bookTitle = book.Title
        if book.KDPASIN != "" {
            amazonURL = fmt.Sprintf("https://www.amazon.it/dp/%s", book.KDPASIN)
        }
    }
}
```

### Task 2.3: Aggiornare chiamata in `generator/scripts.go`

**File**: `internal/generator/scripts.go:47-50`

```go
// PRIMA:
func (g *ScriptGenerator) GenerateScript(idea *models.ContentIdea, bookTitle, platform string) (*GeneratedScript, error) {
    prompt := prompts.ScriptPromptTemplate(ideaDescription, bookTitle, platform)

// DOPO:
func (g *ScriptGenerator) GenerateScript(idea *models.ContentIdea, bookTitle, platform, amazonURL string) (*GeneratedScript, error) {
    prompt := prompts.ScriptPromptTemplate(ideaDescription, bookTitle, platform, amazonURL)
```

### Verifica Phase 2

```bash
# 1. Verificare che il libro abbia un ASIN
bin/gagipress books list -v

# 2. Se manca ASIN, aggiungerlo
bin/gagipress books edit <book-id> --asin B0XXXXXXX

# 3. Generare un nuovo script e verificare che il CTA contenga il link Amazon
bin/gagipress generate script <approved-idea-id> --platform tiktok

# 4. Grep per confermare che nessun vecchio caller è rimasto
grep -rn "ScriptPromptTemplate" internal/ cmd/
grep -rn "GenerateScript(" internal/ cmd/
```

---

## Phase 3: Batch Script Generation (P2)

**Impatto**: Elimina la necessità di generare script uno-a-uno
**Effort**: ~45min
**File da creare**: 1 (`cmd/generate/batch.go`)

### Task 3.1: Creare `cmd/generate/batch.go`

**Pattern da copiare**: `cmd/generate/ideas.go` (batch idea generation, lines 102-140)

Struttura del comando:
```go
// gagipress generate batch [--platform tiktok|instagram] [--limit N]
var batchCmd = &cobra.Command{
    Use:   "batch",
    Short: "Generate scripts for all approved ideas",
    RunE:  runBatch,
}
```

**Logica**:
1. Query `contentRepo.GetIdeas("approved", limit)` — tutte le idee approved non-scripted
2. Per ogni idea:
   - Fetch book data (title + ASIN)
   - Genera script con `ScriptGenerator.GenerateScript()`
   - Salva con `ScriptGenerator.SaveScript()`
   - Continua su errore (warning, non fail)
3. Mostra statistiche finali

### Task 3.2: Registrare il comando

**File**: `cmd/generate/generate.go` — aggiungere `GenerateCmd.AddCommand(batchCmd)` nell'`init()`

### Verifica Phase 3

```bash
# 1. Verificare idee approved non-scripted
bin/gagipress ideas list | grep approved

# 2. Eseguire batch
bin/gagipress generate batch --platform tiktok

# 3. Verificare che tutte siano state scriptate
bin/gagipress ideas list | grep scripted
```

---

## Phase 4: Bluesky Publishing (P3 — Primo Canale Autonomo)

**Impatto**: Pubblicazione automatica su Bluesky (NO OAuth complesso — usa App Password)
**Effort**: ~2-3h
**Perché Bluesky**: Instagram/TikTok richiedono OAuth + Facebook Developer App. Bluesky usa app password = si implementa in ore, non settimane.

### Task 4.1: Aggiungere dipendenza ATProto

```bash
go get github.com/bluesky-social/indigo
```

### Task 4.2: Creare Bluesky client

**File**: `internal/social/bluesky.go` (nuovo)

**API ATProto necessarie**:
- `com.atproto.server.createSession` — login con app password
- `com.atproto.repo.createRecord` — pubblicare un post
- `com.atproto.repo.uploadBlob` — upload immagine
- `app.bsky.feed.post` — record type per i post

**Struct**:
```go
type BlueskyClient struct {
    handle      string
    appPassword string
    pdsURL      string // default: https://bsky.social
    accessJwt   string
    did         string
}
```

**Metodi**:
```go
NewBlueskyClient(cfg *config.BlueskyConfig) *BlueskyClient
Login() error
PublishPost(text string, link string, imageBytes []byte) (string, error)
TestConnection() error
```

### Task 4.3: Aggiungere config Bluesky

**File**: `internal/config/config.go`

```go
type BlueskyConfig struct {
    Handle      string `mapstructure:"handle" yaml:"handle"`
    AppPassword string `mapstructure:"app_password" yaml:"app_password"`
}
```

### Task 4.4: Creare comando publish

**File**: `cmd/publish/publish.go` (nuovo)

```bash
# Pubblica un singolo post dal calendario
gagipress publish <calendar-entry-id> --platform bluesky

# Pubblica tutti gli entry approvati per oggi
gagipress publish today --platform bluesky
```

### Task 4.5: Aggiungere "bluesky" come piattaforma

**File**: `internal/models/content.go:113`

Estendere la validazione di `ContentCalendarInput.Platform`:
```go
// PRIMA:
if c.Platform != "instagram" && c.Platform != "tiktok" {

// DOPO:
validPlatforms := map[string]bool{"instagram": true, "tiktok": true, "bluesky": true}
if !validPlatforms[c.Platform] {
```

### Verifica Phase 4

```bash
# 1. Test connessione
bin/gagipress auth bluesky

# 2. Pubblicare un post di test
bin/gagipress publish <entry-id> --platform bluesky

# 3. Verificare su bsky.app che il post sia apparso
```

---

## Phase 5: Generazione Immagini con Gemini API (P4)

**Impatto**: Crea immagini promozionali per i post (copertine, infografiche, quote card)
**Effort**: ~2-3h
**API**: Google Gemini API (Imagen 3) via REST

### Task 5.1: Creare Gemini Image client

**File**: `internal/ai/gemini_api.go` (nuovo — separato dal browser automation)

**API Endpoint**: `https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent`

**Feature**: Gemini 2.0 supporta image generation nativo tramite `response_modalities: ["IMAGE", "TEXT"]`

```go
type GeminiAPIClient struct {
    apiKey     string
    httpClient *http.Client
    baseURL    string
}

func NewGeminiAPIClient(apiKey string) *GeminiAPIClient
func (c *GeminiAPIClient) GenerateImage(prompt string) ([]byte, string, error) // returns imageBytes, mimeType, error
func (c *GeminiAPIClient) GeneratePostImage(script *GeneratedScript, bookTitle string) ([]byte, error)
```

### Task 5.2: Aggiungere config Gemini API

**File**: `internal/config/config.go`

```go
type GeminiConfig struct {
    APIKey string `mapstructure:"api_key" yaml:"api_key"`
}
```

### Task 5.3: Integrare generazione immagine nel flow di publishing

**File**: `cmd/publish/publish.go`

Prima di pubblicare su Bluesky:
1. Caricare lo script dal calendario
2. Generare immagine promozionale con Gemini API
3. Uploadare immagine su Bluesky come blob
4. Pubblicare post con testo + immagine + link Amazon

### Task 5.4: Prompt per immagini promozionali

**File**: `internal/prompts/templates.go`

```go
func BookPromoImagePrompt(bookTitle, genre, hook string) string {
    return fmt.Sprintf(`Crea un'immagine promozionale per un post social media.
Libro: "%s"
Genere: %s
Hook: "%s"

L'immagine deve:
- Essere colorata e accattivante
- Avere il titolo del libro ben visibile
- Essere ottimizzata per social media (formato quadrato)
- Target: genitori e bambini
- Stile: allegro, moderno, invitante`, bookTitle, genre, hook)
}
```

### Verifica Phase 5

```bash
# 1. Test generazione immagine
bin/gagipress test gemini-api "Crea un'immagine di un libro da colorare"

# 2. Verificare che il file venga generato
ls -la /tmp/gagipress-*.png

# 3. Test publishing con immagine
bin/gagipress publish <entry-id> --platform bluesky --with-image
```

---

## Phase 6: Supabase Edge Functions per Cron Publishing (P5)

**Impatto**: Pubblicazione completamente autonoma — zero intervento umano
**Effort**: ~3-4h
**Tool**: Supabase MCP (deploy_edge_function)

### Task 6.1: Creare Edge Function `publish-scheduled`

**Deploy via**: Supabase MCP `deploy_edge_function`

```typescript
// Logica:
// 1. Query content_calendar WHERE status='approved' AND scheduled_for <= NOW()
// 2. Per ogni entry:
//    a. Fetch script associato
//    b. Generare immagine (opzionale, via Gemini API)
//    c. Pubblicare su Bluesky (via ATProto REST)
//    d. Update status → 'published', set published_at
//    e. Su errore: status → 'failed', save publish_errors
```

### Task 6.2: Configurare Cron Schedule

**Via Supabase Dashboard** o `pg_cron`:
```sql
SELECT cron.schedule(
    'publish-scheduled-posts',
    '*/15 * * * *', -- ogni 15 minuti
    $$SELECT net.http_post(
        'https://<project-ref>.supabase.co/functions/v1/publish-scheduled',
        '{}',
        'application/json',
        ARRAY[http_header('Authorization', 'Bearer <service-key>')]
    )$$
);
```

### Task 6.3: Monitoring e alerting

Aggiungere al CLI:
```bash
gagipress calendar status  # Mostra: pending, approved, published, failed
gagipress calendar retry   # Riprova i post failed
```

### Verifica Phase 6

```bash
# 1. Deploy edge function
# (via Supabase MCP deploy_edge_function)

# 2. Approvare un post nel calendario
bin/gagipress calendar approve

# 3. Verificare che venga pubblicato automaticamente entro 15min
bin/gagipress calendar status
```

---

## Phase 7: Feedback Loop Vendite (P6)

**Impatto**: Capire QUALE contenuto vende di più → generare PIÙ di quel tipo
**Effort**: ~2h

### Task 7.1: UTM tracking nei link Amazon

Aggiungere parametri UTM ai link Amazon generati:
```
https://www.amazon.it/dp/{ASIN}?tag=gagipress-21&utm_source={platform}&utm_medium=social&utm_campaign={script_id}
```

### Task 7.2: Importazione automatica vendite KDP

```bash
gagipress stats import sales.csv  # Già esistente
gagipress stats correlate         # Correlare post → vendite
```

### Task 7.3: Auto-tuning del generatore

Usare i dati di correlazione per pesare la generazione di idee future:
- Se "educational" vende più di "trend" → generare più idee educational
- Se posting ore 7:00 vende più di 12:00 → spostare scheduling

---

## Sequenza di Esecuzione Raccomandata

```
Phase 1 (30min) → Phase 2 (30min) → Phase 3 (45min)
                                           ↓
                                    Phase 4 (2-3h)
                                           ↓
                                    Phase 5 (2-3h)
                                           ↓
                                    Phase 6 (3-4h)
                                           ↓
                                    Phase 7 (2h)
```

**Milestone 1** (Phases 1-3): Pipeline funzionante end-to-end (generate → approve → schedule → show)
**Milestone 2** (Phases 4-5): Primo post automatico su Bluesky con immagine
**Milestone 3** (Phases 6-7): Pubblicazione autonoma + feedback loop vendite

---

## Dipendenze Esterne

| Dipendenza | Richiesto per | Come Ottenerla |
|---|---|---|
| Bluesky App Password | Phase 4 | bsky.app → Settings → App Passwords |
| Google Gemini API Key | Phase 5 | ai.google.dev → API Keys |
| Amazon Associate Tag | Phase 7 | programma-affiliazione.amazon.it |
| Supabase pg_cron | Phase 6 | Abilitare extension `pg_cron` nel progetto |
