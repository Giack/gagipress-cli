# Piano Gagipress: Stato Reale e Prossime Implementazioni

> **Data**: 2026-02-22
> **Branch**: feature/terminal-ux-improvements
> **Testato con**: `bin/gagipress [command]` — eseguire dopo `make build`

---

## Sezione 1: Stato Reale delle Fasi (Aggiornamento al 2026-02-21)

Il piano precedente (`2026-02-21-sales-automation-improvement-plan.md`) descriveva 7 fasi. Ecco il confronto tra quanto pianificato e quanto **effettivamente implementato**.

### Mappa Fasi → Stato Reale

| Fase | Obiettivo Originale | Stato Reale | Note |
|---|---|---|---|
| Phase 1 | Fix calendar pipeline (plan/show/planner) | ✅ **COMPLETO** | Tutto implementato correttamente |
| Phase 2 | CTA con link Amazon (ASIN nei prompt) | ✅ **COMPLETO + BONUS** | UTM tracking aggiunto |
| Phase 3 | Batch script generation | ✅ **COMPLETO** | `cmd/generate/batch.go` creato |
| Phase 4 | Bluesky publishing | ⚠️ **DEVIATO → BLOTATO** | Usato Blotato invece di Bluesky ATProto |
| Phase 5 | Generazione immagini con Gemini API | ⚠️ **PARZIALE via BLOTATO** | Template-based, non Gemini custom |
| Phase 6 | Supabase Edge Functions cron | ❌ **NON INIZIATO** | Prossima priorità |
| Phase 7 | Feedback loop vendite | ✅ **QUASI COMPLETO** | stats import+show+correlate implementati |

---

### Dettaglio Componenti Verificati

#### ✅ Completato: Calendar Pipeline (Phase 1)

**`cmd/calendar/plan.go`** (linee 87-102):
```go
// REALE: salva nel DB via calendarRepo.CreateEntry()
calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
for _, entry := range calendarEntries {
    if err := entry.Validate(); err != nil { ... continue }
    if _, err := calendarRepo.CreateEntry(entry); err != nil { ... continue }
    savedCount++
}
```

**`cmd/calendar/show.go`** (linee 45-50):
```go
// REALE: legge dal DB via calendarRepo.GetEntries()
calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
entries, err := calendarRepo.GetEntries(statusFilter, 0)
```

**`internal/scheduler/planner.go`** (linea 75):
```go
PostType: "reel", // ✅ AGGIUNTO - era mancante
```

#### ✅ Completato: Amazon CTA con UTM (Phase 2)

**`internal/prompts/templates.go`** (linea 107):
```go
func ScriptPromptTemplate(idea, bookTitle, platform, amazonURL string) string
```

**`cmd/generate/script.go`** (linee 90-103) — estrae ASIN e costruisce URL con UTM:
```go
amazonURL = fmt.Sprintf(
    "https://www.amazon.it/dp/%s?tag=gagipress-21&utm_source=%s&utm_medium=social&utm_campaign=%s",
    book.KDPASIN, platform, idea.ID)
```

**`internal/generator/scripts.go`** (linea 48):
```go
func (g *ScriptGenerator) GenerateScript(idea *models.ContentIdea, bookTitle, platform, amazonURL string) (*GeneratedScript, error)
```

#### ✅ Completato: Batch Script Generation (Phase 3)

**`cmd/generate/batch.go`**: Esiste e funziona.
- Flag: `--platform` (default: tiktok), `--gemini`, `--limit` (default: 10)
- Flusso: query `approved` ideas → fetch book ASIN → GenerateScript → SaveScript
- Gestione errori: continua su failure singola, mostra statistiche finali

#### ⚠️ Deviazione Strategica: Blotato invece di Bluesky (Phase 4-5)

**Decisione**: invece di implementare Bluesky (ATProto) + Gemini Image API custom, è stato scelto **Blotato** come backend di publishing all-in-one.

**`internal/social/blotato.go`** — client reale e completo:
```go
func (c *BlotatoClient) GetAccountID(platform string) (string, error)
func (c *BlotatoClient) GenerateVisual(templateID, prompt string) (string, error)
func (c *BlotatoClient) WaitForVisualCreation(creationID string) (string, error) // polling fino a 5min
func (c *BlotatoClient) PublishPost(accountId, platform, text string, mediaUrls []string, scheduledTime *time.Time) (string, error)
```

**`cmd/publish/publish.go`** — comando completo:
```go
// Single publish: gagipress publish <calendar-entry-id> [--with-media]
// Batch publish:  gagipress publish batch [--limit 10] [--with-media]
```

**`internal/config/config.go`** — BlotatoConfig presente:
```go
type BlotatoConfig struct {
    APIKey     string `mapstructure:"api_key" yaml:"api_key"`
    TemplateID string `mapstructure:"template_id" yaml:"template_id"`
}
```

**Nota**: Instagram/TikTok direct clients (`internal/social/instagram.go`, `internal/social/tiktok.go`) rimangono **STUB** — restituiscono "not implemented". Il publishing avviene tramite Blotato come proxy.

#### ✅ Quasi Completo: Stats Feedback Loop (Phase 7)

**`cmd/stats/`** contiene 4 file:
- `stats.go` — parent command
- `import.go` — `gagipress stats import sales.csv` (KDP CSV import)
- `show.go` — `gagipress stats show [--period 30d] [--platform tiktok]`
- `correlate.go` — `gagipress stats correlate --book <id> [--days 30]` (Pearson correlation coefficient)

---

## Sezione 2: Bug Trovati Durante Testing (2026-02-22)

### Bug 1: `stats show` crash — Timestamp encoding nel URL (CRITICO)

**Errore osservato**:
```
Error: failed to get metrics: HTTP 400: {"code":"22007","details":null,"hint":null,
"message":"invalid input syntax for type timestamp with time zone: \"2026-01-23T10:26:43 01:00\""}
```

**Root cause**: `internal/repository/metrics.go` righe 102-106 usa `time.RFC3339` per formattare i timestamp nella query URL. Quando il sistema è in timezone `+01:00`, il formato produce `2026-01-23T10:26:43+01:00`. Il carattere `+` nel query string viene decodificato come spazio dal server, risultando in `2026-01-23T10:26:43 01:00` → Postgres rifiuta.

**File da modificare**: `internal/repository/metrics.go:102-106`

**Fix**:
```go
// PRIMA (bug):
if !from.IsZero() {
    url += fmt.Sprintf("&collected_at=gte.%s", from.Format(time.RFC3339))
}
if !to.IsZero() {
    url += fmt.Sprintf("&collected_at=lte.%s", to.Format(time.RFC3339))
}

// DOPO (fix — usa UTC, produce "Z" suffix invece di "+01:00"):
if !from.IsZero() {
    url += fmt.Sprintf("&collected_at=gte.%s", from.UTC().Format(time.RFC3339))
}
if !to.IsZero() {
    url += fmt.Sprintf("&collected_at=lte.%s", to.UTC().Format(time.RFC3339))
}
```

**Stesso pattern va controllato in**: `internal/repository/sales.go` (stesso tipo di query)

### Bug 2: `calendar --help` mostra "approve" due volte (MINOR)

**Osservato**: output di `gagipress calendar --help` mostra:
```
Available Commands:
  approve     Approve pending calendar entries
  approve     Approve pending calendar entries   ← DUPLICATO
```

**Root cause**: `approveCmd` viene registrato due volte:
1. `cmd/calendar/calendar.go:21` → `CalendarCmd.AddCommand(approveCmd)`
2. `cmd/calendar/approve.go:29` → `CalendarCmd.AddCommand(approveCmd)`

`plan.go` e `show.go` NON hanno `CalendarCmd.AddCommand()` nel loro `init()` — li registra solo `calendar.go`. Ma `approve.go` ha un pattern inconsistente.

**File da modificare**: `cmd/calendar/approve.go` (rimuovere la riga dal suo `init()`)

**Fix**:
```go
// PRIMA (approve.go:28-30):
func init() {
    CalendarCmd.AddCommand(approveCmd)  // ← RIMUOVERE questa riga
}

// DOPO: init() vuoto (o eliminare l'intera funzione init se non fa altro)
```

### Osservazione 3: Book senza ASIN (non è un bug, ma blocca le Amazon URLs)

Il libro in catalogo (`19d89b62...`) non ha un ASIN impostato. Finché non viene aggiunto, gli script generati non includeranno link Amazon nei CTA.

**Fix dati** (da fare manualmente):
```bash
bin/gagipress books edit 19d89b62-2a9a-4d31-9298-fc77d282ddae --asin B0XXXXXXXX
```
(sostituire con l'ASIN reale del libro su Amazon.it)

---

## Sezione 3: Piano Prossime Implementazioni

### Fase A: Hotfix (Priorità P0 — ~30 min)

**Obiettivo**: Correggere i due bug trovati nel testing.

#### Task A.1 — Fix timestamp UTC in metrics.go

**File**: `internal/repository/metrics.go`

1. Leggere il file (già verificato nella sezione 2)
2. Alla riga 102, cambiare `from.Format(time.RFC3339)` → `from.UTC().Format(time.RFC3339)`
3. Alla riga 105, cambiare `to.Format(time.RFC3339)` → `to.UTC().Format(time.RFC3339)`
4. Verificare se lo stesso pattern esiste in `internal/repository/sales.go` e applicare lo stesso fix
5. Build: `make build`
6. Test: `bin/gagipress stats show` deve completare senza errore 400

**Verifica**:
```bash
make build
bin/gagipress stats show  # deve mostrare dashboard (vuota OK, ma senza errore)
```

#### Task A.2 — Fix duplicate approve in calendar

**File**: `cmd/calendar/approve.go`

1. Leggere il file (già verificato nella sezione 2)
2. Rimuovere `CalendarCmd.AddCommand(approveCmd)` dalla funzione `init()` in `approve.go`
   - La registrazione rimane in `calendar.go:21`
3. Build: `make build`

**Verifica**:
```bash
make build
bin/gagipress calendar --help  # deve mostrare "approve" una sola volta
```

---

### Fase B: Configurazione Blotato e Primo Test Publish (Priorità P1 — ~1h)

**Prerequisiti esterni**:
- Account Blotato con API key
- Account TikTok/Instagram connesso a Blotato
- ASIN del libro (per avere Amazon URL nel CTA)

**Obiettivo**: Configurare l'integrazione Blotato e testare il publish di un post reale.

#### Task B.1 — Configurare Blotato

```bash
# Aggiungere al config ~/.gagipress/config.yaml:
# blotato:
#   api_key: "YOUR_BLOTATO_API_KEY"
#   template_id: "YOUR_TEMPLATE_ID"  # opzionale per media generation

bin/gagipress init  # oppure modificare direttamente il config
```

#### Task B.2 — Aggiungere ASIN al libro

```bash
bin/gagipress books edit 19d89b62-2a9a-4d31-9298-fc77d282ddae --asin B0XXXXXXXX
```

#### Task B.3 — Test del workflow completo

```bash
# 1. Verificare che il calendario abbia post approvati
bin/gagipress calendar show

# 2. Approvare i post pending
bin/gagipress calendar approve

# 3. Test publish singolo (prima di lanciare il batch)
bin/gagipress publish <calendar-entry-id>

# 4. Verificare su Blotato/piattaforma che il post sia apparso

# 5. Test batch (solo se il singolo funziona)
bin/gagipress publish batch --limit 3
```

#### Task B.4 — Test generate batch con ASIN reale

```bash
# Dopo aver aggiunto ASIN, rigenerare gli script per includere link Amazon
bin/gagipress generate batch --platform tiktok --limit 5
# Verificare che i nuovi script contengano "amazon.it/dp/B0XXXXXXXX"
```

---

### Fase C: Supabase Edge Function per Cron Publishing (Priorità P2 — ~3h)

**Obiettivo**: Pubblicazione automatica senza intervento manuale — ogni 15 minuti controlla i post approvati con `scheduled_for <= NOW()` e li pubblica via Blotato.

**Tool da usare**: Supabase MCP `deploy_edge_function`

#### Task C.1 — Verificare dipendenze database

Prima di deployare, verificare che:
1. La tabella `content_calendar` abbia i campi `published_at` e `publish_errors`
2. Se mancano, applicare una migration:

```sql
-- migrations/006_add_publish_tracking.sql
ALTER TABLE content_calendar
ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS publish_errors TEXT;
```

Verificare via Supabase MCP:
```graphql
# query per controllare schema
```

#### Task C.2 — Creare Edge Function `publish-scheduled`

**Nota sul pattern da seguire**: consultare la Supabase Edge Functions documentation via MCP prima di scrivere il codice.

**Logica da implementare**:
```typescript
// supabase/functions/publish-scheduled/index.ts
import "jsr:@supabase/functions-js/edge-runtime.d.ts";

Deno.serve(async (req) => {
    const supabase = createClient(SUPABASE_URL, SERVICE_ROLE_KEY);
    const blotatoApiKey = Deno.env.get("BLOTATO_API_KEY");

    // 1. Query approved entries scheduled for <= NOW()
    const { data: entries } = await supabase
        .from("content_calendar")
        .select("*, content_scripts(*)")
        .eq("status", "approved")
        .lte("scheduled_for", new Date().toISOString())
        .is("published_at", null);

    // 2. Per ogni entry:
    for (const entry of entries ?? []) {
        try {
            // a. Costruisci il testo del post dallo script
            const script = entry.content_scripts;
            const text = buildPostText(script);

            // b. Ottieni account ID da Blotato per la platform
            const accountId = await getBlotatoAccountId(blotatoApiKey, entry.platform);

            // c. Pubblica su Blotato
            const submissionId = await publishToBlotato(blotatoApiKey, accountId, entry.platform, text, []);

            // d. Aggiorna status nel DB
            await supabase.from("content_calendar")
                .update({ status: "published", published_at: new Date().toISOString() })
                .eq("id", entry.id);
        } catch (err) {
            await supabase.from("content_calendar")
                .update({ status: "failed", publish_errors: err.message })
                .eq("id", entry.id);
        }
    }

    return new Response(JSON.stringify({ processed: entries?.length ?? 0 }));
});
```

**Anti-pattern da evitare**:
- NON hardcodare API keys nel codice — usare `Deno.env.get()`
- NON usare `supabase-js v1` — usare v2 con `jsr:@supabase/supabase-js`
- NON dimenticare il loop di retry/backoff per Blotato

#### Task C.3 — Deploy via Supabase MCP

```typescript
// Usare: mcp__plugin_supabase_supabase__deploy_edge_function
// project_id: (leggere da ~/.gagipress/config.yaml o supabase/config.toml)
// name: "publish-scheduled"
// verify_jwt: false  // Sarà chiamato da pg_cron, non da utenti
// files: [{ name: "index.ts", content: "..." }]
```

**Secrets da configurare nel Supabase Dashboard**:
- `BLOTATO_API_KEY` — API key Blotato

#### Task C.4 — Configurare pg_cron

**Via Supabase MCP `execute_sql`** (dopo aver abilitato `pg_cron` extension):
```sql
-- Abilitare pg_cron (una sola volta)
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Schedulare ogni 15 minuti
SELECT cron.schedule(
    'publish-scheduled-posts',
    '*/15 * * * *',
    $$
    SELECT net.http_post(
        url := 'https://<PROJECT_REF>.supabase.co/functions/v1/publish-scheduled',
        headers := '{"Content-Type": "application/json", "Authorization": "Bearer <SERVICE_ROLE_KEY>"}'::jsonb,
        body := '{}'::jsonb
    );
    $$
);
```

**Verifica Phase C**:
```bash
# 1. Verificare che la funzione sia deployata
# (via Supabase MCP list_edge_functions)

# 2. Approvare un post
bin/gagipress calendar approve

# 3. Invocare manualmente la funzione per test
curl -X POST https://<PROJECT_REF>.supabase.co/functions/v1/publish-scheduled \
  -H "Authorization: Bearer <SERVICE_ROLE_KEY>"

# 4. Verificare che il post sia pubblicato
bin/gagipress calendar show --status published
```

---

### Fase D: Monitoring e Status Commands (Priorità P3 — ~1h)

**Obiettivo**: Aggiungere visibilità sullo stato del publishing automatico.

#### Task D.1 — Aggiungere `calendar status` command

**File da creare**: `cmd/calendar/status.go`

**Pattern da copiare**: `cmd/calendar/show.go` (stesso pattern di query + display)

**Comando**:
```bash
gagipress calendar status
# Output: tabella per status (pending_approval, approved, published, failed)
# con conteggi + ultime entries per categoria
```

#### Task D.2 — Aggiungere `calendar retry` command per post falliti

**File da creare**: `cmd/calendar/retry.go`

**Logica**: query entries con `status=failed`, resettare a `approved` per ritentare al prossimo cron.

```bash
gagipress calendar retry         # Resetta tutti i failed → approved
gagipress calendar retry <id>    # Resetta un singolo entry
```

---

### Fase E: UTM Correlation Verification (Priorità P3 — ~1h)

**Obiettivo**: Verificare che il ciclo completo UTM→vendite funzioni.

#### Task E.1 — Verificare formato UTM nei script generati

Dopo aver aggiunto l'ASIN al libro e rigenerato script:
```bash
# Generare uno script e leggere il CTA
bin/gagipress generate script <approved-idea-id> --platform tiktok

# Verificare nel DB che il link sia:
# https://www.amazon.it/dp/B0XXXXXXXX?tag=gagipress-21&utm_source=tiktok&utm_medium=social&utm_campaign=<idea-id>
```

#### Task E.2 — Test stats correlate

Dopo aver importato almeno 3 giorni di dati vendite:
```bash
bin/gagipress stats import sales.csv
bin/gagipress stats correlate --book 19d89b62-2a9a-4d31-9298-fc77d282ddae --days 30
```

---

## Sequenza di Esecuzione Raccomandata

```
Fase A (30min)  — Hotfix critici (stats + duplicate approve)
    ↓
Fase B (1h)     — Configurazione Blotato + primo publish manuale
    ↓
Fase C (3h)     — Edge Function cron publishing
    ↓
Fase D (1h)     — Monitoring commands (status + retry)
    ↓
Fase E (1h)     — UTM correlation verification
```

**Milestone 1** (Fase A): Zero bug nel testing base
**Milestone 2** (Fase B): Primo post pubblicato su TikTok/Instagram via Blotato
**Milestone 3** (Fase C): Sistema autonomo — pubblica senza intervento manuale
**Milestone 4** (Fasi D+E): Loop completo generate→publish→track_sales operativo

---

## Dipendenze Esterne Necessarie

| Dipendenza | Richiesta per | Come Ottenerla |
|---|---|---|
| Blotato API Key | Fase B, C | app.blotato.com → Settings → API |
| Blotato Template ID | Fase B (opz.) | app.blotato.com → Templates |
| Amazon ASIN del libro | Fase B | amazon.it → pagina prodotto → URL (`/dp/BXXXXXXXX`) |
| Supabase Service Role Key | Fase C | Supabase Dashboard → Project Settings → API |

---

## File di Riferimento per Ogni Fase

| Fase | File Chiave |
|---|---|
| A.1 | `internal/repository/metrics.go:102-106` |
| A.2 | `cmd/calendar/approve.go:28-30`, `cmd/calendar/calendar.go:18-22` |
| B | `internal/social/blotato.go`, `cmd/publish/publish.go`, `internal/config/config.go` |
| C | `supabase/functions/publish-scheduled/index.ts` (nuovo), `migrations/006_*.sql` (nuovo) |
| D | `cmd/calendar/status.go` (nuovo), `cmd/calendar/retry.go` (nuovo) |
| E | `cmd/stats/correlate.go`, `internal/repository/sales.go` |
