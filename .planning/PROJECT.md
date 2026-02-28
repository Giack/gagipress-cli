# Gagipress Dashboard

## What This Is

Una dashboard web read-only per visualizzare i dati prodotti dalla CLI Gagipress. Backoffice personale che mostra lo stato della pipeline content (idee, script, calendario) e il catalogo libri tramite una UI kanban dark, connessa direttamente a Supabase e deployata su Vercel come sito statico.

**v1.0 shipped:** Books table, Ideas kanban (4 colonne), Calendar kanban (5 colonne) con skeleton loading, error/empty states, e platform color-coding.

## Core Value

Vedere a colpo d'occhio dove si trova ogni idea nel pipeline content — da `pending` a `published` — senza aprire il terminale.

## Requirements

### Validated

- ✓ CLI genera e gestisce idee, script, calendario, libri in Supabase — existing
- ✓ Supabase espone tutti i dati via REST API (PostgREST) — existing
- ✓ Schema: `books`, `content_ideas`, `content_calendar`, `post_metrics`, `sales_data` — existing
- ✓ Dashboard accessibile via URL Vercel (static site, no server) — v1.0
- ✓ RLS SELECT-only su tutte le tabelle (migration 009) — v1.0
- ✓ Kanban ideas: colonne pending / approved / rejected / scripted — v1.0
- ✓ Kanban calendario: colonne scheduled / approved / publishing / published / failed — v1.0
- ✓ Pagina libri: tabella con titolo, kdp_asin (link Amazon), genere, target audience — v1.0
- ✓ Design minimal dark con Tailwind CSS CDN v4 — v1.0
- ✓ Skeleton loading + error banner + empty states — v1.0
- ✓ Platform color-coding: TikTok pink, Instagram purple — v1.0

### Active

- [ ] Filtro per status / piattaforma nel kanban idee e calendario (v2)
- [ ] Sales & metrics view con correlazione KDP (v2)
- [ ] Contatori per colonna nel kanban (v2)
- [ ] Refresh manuale senza reload pagina (v2)

### Out of Scope

- Login / autenticazione — repo privato, link diretto sufficiente
- Operazioni di scrittura — CLI è source of truth
- Build process (npm, webpack, vite) — zero tooling per semplicità
- Mobile-first — desktop personal tool
- Real-time subscriptions Supabase — refresh manuale in v2
- Multi-utente — tool personale

## Context

**Current state (v1.0):** Dashboard live su Vercel. 10 file, ~500 LOC Vanilla JS + HTML. Stack: HTML + Tailwind CDN v4 + ES Modules. Zero npm.

**Tech discoveries from v1.0:**
- `book.kdp_asin` non `book.asin` — nome colonna da migration 001
- skeleton-first render pattern: `renderSkeleton()` sincrono prima di ogni `await`
- `[data-nav]` attribute selector per nav targeting robusto
- `Promise.all` + `Map` lookup per multi-table join client-side O(1)

**Next milestone focus:** Filtering, column counters, analytics (v2 requirements above).

## Constraints

- **Stack**: Vanilla HTML/CSS/JS — nessun build tool, compatibile con Vercel static hosting
- **Autenticazione**: Nessuna — accesso tramite URL non pubblico
- **Database**: Read-only via Supabase anon key (nessuna mutation)
- **Dipendenze**: Solo CDN (Tailwind CSS v4) — zero npm install
- **Deployment**: Vercel static hosting da GitHub repo

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Vanilla JS + Tailwind CDN | Zero dipendenze, zero build, funziona su Vercel static | ✓ Good — shipped v1.0 in 3 giorni |
| Kanban come layout principale | Rende immediatamente visibile lo stato della pipeline | ✓ Good — core value realizzato |
| Supabase anon key lato client | Progetto personale, dati non sensibili, RLS già attiva | ✓ Good — sicuro per uso personale |
| Nessuna auth | Overhead non giustificato per uso personale | ✓ Good — frictionless |
| skeleton-first render | Evita blank screen durante fetch async | ✓ Good — pattern riutilizzabile in ogni vista |
| Promise.all + Map per join | Client-side join O(1) senza query complesse | ✓ Good — adottato in fasi 4 e 5 |

---
*Last updated: 2026-02-28 after v1.0 milestone*
