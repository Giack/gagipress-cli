# Requirements: Gagipress Dashboard

**Defined:** 2026-02-25
**Core Value:** Vedere a colpo d'occhio dove si trova ogni idea nel pipeline content — da `pending` a `published` — senza aprire il terminale.

## v1 Requirements

### Infrastructure

- [ ] **INFRA-01**: Dashboard deployata come sito statico su Vercel dalla cartella `dashboard/` del repo
- [ ] **INFRA-02**: Connessione a Supabase via supabase-js v2 (CDN ESM), configurata in `dashboard/config.js`
- [ ] **INFRA-03**: `dashboard/config.js` aggiunto a `.gitignore` (contiene anon key, repo privato)
- [ ] **INFRA-04**: RLS abilitata su tutte le 5 tabelle con policy SELECT-only per ruolo `anon`
- [ ] **INFRA-05**: Tailwind CSS v4 CDN caricato come script tag in `index.html`

### Navigation

- [ ] **NAV-01**: Navbar con 3 sezioni: Books / Ideas / Calendar
- [ ] **NAV-02**: Navigazione via hash routing (`#/books`, `#/ideas`, `#/calendar`)
- [ ] **NAV-03**: Voce attiva evidenziata nella navbar

### Books View

- [ ] **BOOKS-01**: Tabella libri con colonne: titolo, ASIN, genere, target audience
- [ ] **BOOKS-02**: Click su ASIN apre la pagina prodotto Amazon in una nuova tab

### Ideas Kanban

- [ ] **IDEAS-01**: Kanban con 4 colonne: pending / approved / rejected / scripted
- [ ] **IDEAS-02**: Card mostra titolo idea e piattaforma (TikTok / Instagram)
- [ ] **IDEAS-03**: Click su card espande preview dello script generato (se presente)

### Calendar Kanban

- [ ] **CAL-01**: Kanban con 5 colonne: scheduled / approved / publishing / published / failed
- [ ] **CAL-02**: Card mostra data programmata e piattaforma
- [ ] **CAL-03**: Card mostra titolo dell'idea collegata

### Global UX

- [ ] **UX-01**: Spinner di loading durante il fetch da Supabase
- [ ] **UX-02**: Messaggio di errore visibile se il fetch fallisce
- [ ] **UX-03**: Empty state se una colonna o tabella è vuota
- [ ] **UX-04**: Design minimal dark (CSS custom properties o Tailwind dark palette)

## v2 Requirements

### Filtering

- **FILT-01**: Filtro per status nel kanban idee
- **FILT-02**: Filtro per piattaforma (TikTok / Instagram) nel kanban idee e calendario
- **FILT-03**: Filtro per libro nel kanban idee

### Analytics

- **ANLT-01**: Vista sales & metrics con dati KDP importati
- **ANLT-02**: Correlazione post pubblicati vs vendite

### UX Enhancements

- **UX-V2-01**: Card `failed` nel calendario evidenziate in rosso (allarme visivo)
- **UX-V2-02**: Contatori per colonna nel kanban (N card per status)
- **UX-V2-03**: Refresh manuale dei dati senza reload pagina
- **UX-V2-04**: Timestamp ultimo aggiornamento dati

## Out of Scope

| Feature | Reason |
|---------|--------|
| Autenticazione / login | Repo privato, link diretto sufficiente per uso personale |
| Operazioni di scrittura | CLI è source of truth, dashboard è sola lettura |
| Build process (npm, webpack, vite) | Zero tooling per semplicità massima |
| Mobile-first / responsive | Desktop personal tool, non priorità v1 |
| Real-time subscriptions Supabase | Overkill per uso personale, refresh manuale in v2 |
| Sales & metrics view | Dati KDP raramente aggiornati, bassa priorità v1 |
| Multi-utente | Tool personale, utente singolo |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| INFRA-01 | Phase 1 | Pending |
| INFRA-02 | Phase 1 | Pending |
| INFRA-03 | Phase 1 | Pending |
| INFRA-04 | Phase 1 | Pending |
| INFRA-05 | Phase 1 | Pending |
| NAV-01 | Phase 2 | Pending |
| NAV-02 | Phase 2 | Pending |
| NAV-03 | Phase 2 | Pending |
| UX-01 | Phase 2 | Pending |
| UX-02 | Phase 2 | Pending |
| UX-03 | Phase 2 | Pending |
| UX-04 | Phase 2 | Pending |
| BOOKS-01 | Phase 3 | Pending |
| BOOKS-02 | Phase 3 | Pending |
| IDEAS-01 | Phase 4 | Pending |
| IDEAS-02 | Phase 4 | Pending |
| IDEAS-03 | Phase 4 | Pending |
| CAL-01 | Phase 5 | Pending |
| CAL-02 | Phase 5 | Pending |
| CAL-03 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 20 total
- Mapped to phases: 20
- Unmapped: 0 ✓

---
*Requirements defined: 2026-02-25*
*Last updated: 2026-02-25 after initial definition*
