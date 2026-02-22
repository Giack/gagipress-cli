# Design: publish-scheduled Edge Function

> **Data**: 2026-02-22
> **Obiettivo**: Pubblicazione automatica dei post approvati ogni 15 minuti tramite Supabase Edge Function + pg_cron
> **Approccio scelto**: Single function con optimistic locking (anti double-posting)

## Database Changes

```sql
ALTER TABLE content_calendar
  ADD COLUMN IF NOT EXISTS published_at   TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS publish_errors TEXT,
  ADD COLUMN IF NOT EXISTS generate_media BOOLEAN NOT NULL DEFAULT FALSE;
```

**Status flow**:
```
pending_approval → approved → publishing → published
                                        ↘ failed
```
`publishing` è un lock transitorio. Entry bloccate > 10 min vengono rollbackate ad `approved`.

## Edge Function Logic (`supabase/functions/publish-scheduled/index.ts`)

1. **Cleanup stale locks**: `UPDATE ... SET status='approved' WHERE status='publishing' AND updated_at < NOW() - INTERVAL '10 minutes'`
2. **Lock atomico**: `UPDATE ... SET status='publishing' WHERE status='approved' AND scheduled_for <= NOW() AND published_at IS NULL RETURNING *`
3. Per ogni entry:
   - Fetch script da `content_scripts`
   - Build testo: `hook + "\n\n" + main_content + "\n\n" + cta + "\n" + hashtags`
   - Se `generate_media=true`: chiama Blotato template API, polling max 2 min, fallback text-only su timeout
   - `GET /users/me/accounts?platform={platform}` → accountId
   - `POST /posts` → pubblica
   - `UPDATE status='published', published_at=NOW()`
   - Su errore: `UPDATE status='failed', publish_errors=msg`
4. Return `{ processed, published, failed }`

**Secrets**: `BLOTATO_API_KEY` (manuale), `SUPABASE_URL` + `SUPABASE_SERVICE_ROLE_KEY` (automatici)

## pg_cron Schedule

```sql
CREATE EXTENSION IF NOT EXISTS pg_cron;
CREATE EXTENSION IF NOT EXISTS pg_net;

SELECT cron.schedule(
  'gagipress-publish-scheduled',
  '*/15 * * * *',
  $$ SELECT net.http_post(...) $$
);
```

## Verifica

1. Deploy Edge Function via Supabase MCP
2. Aggiungere secret `BLOTATO_API_KEY` nel Dashboard (manuale)
3. Approvare un post nel calendario
4. Invocare la funzione manualmente via curl
5. Verificare `status='published'` nel DB
