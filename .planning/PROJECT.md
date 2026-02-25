# Gagipress Dashboard

## What This Is

Una dashboard web read-only per visualizzare i dati prodotti dalla CLI Gagipress. Pensata come backoffice personale: mostra lo stato della pipeline content (idee, script, calendario), i libri nel catalogo, e i dati del calendario pubblicazione con una UI kanban. Si connette direttamente a Supabase e gira su Vercel come sito statico.

## Core Value

Vedere a colpo d'occhio dove si trova ogni idea nel pipeline content — da `pending` a `published` — senza aprire il terminale.

## Requirements

### Validated

- ✓ CLI genera e gestisce idee, script, calendario, libri in Supabase — existing
- ✓ Supabase espone tutti i dati via REST API (PostgREST) — existing
- ✓ Schema: `books`, `content_ideas`, `content_calendar`, `post_metrics`, `sales_data` — existing

### Active

- [ ] Dashboard accessibile via URL Vercel (static site, no server)
- [ ] Kanban ideas: colonne pending / approved / rejected / scripted
- [ ] Kanban calendario: colonne scheduled / approved / publishing / published / failed
- [ ] Pagina libri: tabella catalogo con titolo, ASIN, genere, target audience
- [ ] Design minimal dark con Tailwind CSS (CDN, zero build step)
- [ ] Nessuna operazione di scrittura — pura visualizzazione
- [ ] Connessione Supabase tramite anon key (read-only, RLS permissive per select)

### Out of Scope

- Login / autenticazione — link privato è sufficiente per uso personale
- Operazioni di scrittura (approve, reject, publish) — rimangono nella CLI
- Sales & metrics view — bassa priorità per v1
- Mobile-first — desktop first, non responsive ottimizzata
- Build process (webpack, vite, npm) — vanilla JS con CDN per zero tooling

## Context

Il progetto CLI è maturo (MVP completato, fase 5+). Il database Supabase ha già tutti i dati strutturati. La dashboard non ha bisogno di backend dedicato — Supabase PostgREST espone già tutto via HTTP con anon key. Vercel deploya siti statici gratuitamente da un repo GitHub.

Stack dashboard: HTML + Tailwind CSS (CDN) + Vanilla JS (ES modules). Nessun framework, nessun npm. Un singolo comando `git push` per deployare.

La anon key di Supabase sarà configurata come variabile d'ambiente Vercel e iniettata in build via un file di config generato, oppure inclusa direttamente (dato che è read-only e il progetto è personale).

## Constraints

- **Stack**: Vanilla HTML/CSS/JS — nessun build tool, compatibile con Vercel static hosting
- **Autenticazione**: Nessuna — accesso tramite URL non pubblico
- **Database**: Read-only via Supabase anon key (nessuna mutation)
- **Dipendenze**: Solo CDN (Tailwind CSS) — zero npm install
- **Deployment**: Vercel static hosting da GitHub repo

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Vanilla JS + Tailwind CDN | Zero dipendenze, zero build, funziona su Vercel static | — Pending |
| Kanban come layout principale | Rende immediatamente visibile lo stato della pipeline | — Pending |
| Supabase anon key lato client | Progetto personale, dati non sensibili, RLS già attiva | — Pending |
| Nessuna auth | Overhead non giustificato per uso personale | — Pending |

---
*Last updated: 2026-02-25 after initialization*
