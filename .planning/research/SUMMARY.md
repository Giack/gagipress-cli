# Project Research Summary

**Project:** Gagipress Dashboard — Static read-only web dashboard
**Domain:** Kanban pipeline dashboard (vanilla JS + Supabase REST + Vercel static)
**Researched:** 2026-02-25
**Confidence:** MEDIUM-HIGH

## Executive Summary

This project is a read-only operations dashboard for the Gagipress CLI — a personal internal tool used to monitor the content pipeline (ideas, calendar, books) stored in Supabase. The correct approach is a zero-build-step static site: plain HTML, vanilla JS ES modules, Tailwind CSS via Play CDN, and the Supabase anon key used directly in browser code. No framework, no npm, no bundler. The stack is intentionally minimal because the audience is one developer who needs fast visibility into pipeline state, not a production SaaS.

The recommended architecture is a single `index.html` shell with hash-based routing (`#/ideas`, `#/calendar`, `#/books`). Each view fetches its own data from Supabase's PostgREST REST API using a thin raw `fetch` wrapper (not the supabase-js SDK, which adds unnecessary weight for simple SELECT queries). Components are pure renderers — data in, HTML string out — with no shared state. This pattern is well-understood, requires no build tooling, and deploys to Vercel on every `git push`.

The key risk is security misconfiguration: if RLS is not enabled on all five Supabase tables, the hardcoded anon key in the JS source exposes the entire database to public read/write. This must be resolved before the dashboard makes its first live query. A secondary risk is the Tailwind Play CDN causing visible Flash of Unstyled Content (FOUC) on dynamically rendered kanban cards — addressable with an opacity mask or by switching to the Tailwind CLI binary. Both risks have clear, well-documented mitigations.

## Key Findings

### Recommended Stack

The stack is intentionally constrained to zero local tooling and zero build step. HTML5 + vanilla JS ES modules (ES2022+) provide the structure and logic layer. Tailwind CSS is loaded via the Play CDN (`@tailwindcss/browser@4` from jsDelivr) — acceptable for a personal internal tool where JIT compilation overhead (~50-100ms) is irrelevant. Data access uses raw `fetch` against Supabase PostgREST rather than the supabase-js SDK, saving ~100KB of CDN payload for a dashboard that only needs `SELECT *` with basic filtering. Vercel static hosting autodeploys from `git push` with no configuration required for root-level `index.html`.

**Core technologies:**
- HTML5 + Vanilla JS ES Modules: structure and logic — no bundler, no framework, maximum simplicity
- Tailwind Play CDN v4 (`@tailwindcss/browser@4`): utility styling — zero-build, single script tag; FOUC mitigation required
- Raw `fetch` against PostgREST: data access — 15 lines replaces ~100KB SDK for simple SELECT queries
- Vercel Static Hosting: deployment — autodeploy on push, no `vercel.json` needed for root `index.html`

**Critical version note:** Supabase projects created after November 2025 use `sb_publishable_...` keys instead of legacy anon keys. This project predates that cutoff, so the legacy anon key applies and is safe to hardcode in browser JS when RLS is enabled.

### Expected Features

The dashboard value proposition is bottleneck visibility — "7 ideas pending, 0 scripted" tells you immediately what the CLI needs to do next. All features are low complexity. The MVP is achievable in a single focused session.

**Must have (table stakes):**
- Ideas kanban (pending / approved / rejected / scripted columns) — core pipeline visibility
- Calendar kanban (scheduled / approved / publishing / published / failed columns) — publishing state at a glance
- Books catalog table — reference data for book metadata (title, ASIN, genre, audience)
- Column item counts in headers — bottleneck signal without reading individual cards
- Visual distinction between columns, especially `failed` cards (red/alarming) — a failed post must not be missable
- Empty state per column — blank columns look broken without it
- Fresh data on load (no caching) — stale data creates false confidence

**Should have (differentiators that drive daily use):**
- Manual refresh button + last-refreshed timestamp — trust signal; essential when used alongside the CLI
- Per-platform color coding on calendar cards (Instagram vs TikTok) — scan by platform, not read it
- Book title on idea cards — prevents context-switching to the books tab
- Navigation tabs (Ideas / Calendar / Books) — three views cannot coexist on one screen
- Scheduled date prominently on calendar cards — most time-sensitive data point

**Defer indefinitely (anti-features):**
- Write operations (approve, reject, publish) — erodes CLI-as-source-of-truth contract, requires auth
- Search, filter, sort controls — unnecessary for <50 items per column
- Charts / analytics — out of scope for v1, add visual noise to kanban
- Auth / login — single-user personal tool on a private URL
- Mobile/responsive layout — desk tool used alongside a terminal

### Architecture Approach

The architecture is a single-page app with hash routing, unidirectional data flow, and pure component renderers. `index.html` is the shell; `app.js` bootstraps the hash router; each view (`views/ideas.js`, `views/calendar.js`, `views/books.js`) owns its fetch-and-render lifecycle; shared components (`components/kanban.js`, `components/table.js`) receive data and return HTML strings with no side effects. A single `config.js` exports the Supabase URL and anon key — all other modules import from it. No global mutable state, no reactive framework, no event bus.

**Major components:**
1. `config.js` — Supabase URL + anon key constants (single source of truth; excluded from git)
2. `api/supabase.js` — thin `fetch` wrapper: builds PostgREST URL, injects `apikey`/`Authorization` headers, checks `response.ok`
3. `router.js` — `hashchange` listener maps `#/route` to view function, dispatches on load
4. `views/*.js` (ideas, calendar, books) — fetch own data, group/transform, inject into `#app` via component renderers
5. `components/kanban.js` — pure renderer: columns array in, HTML string out
6. `components/table.js` — pure renderer: rows array in, HTML string out

### Critical Pitfalls

1. **RLS not enabled on Supabase tables** — the hardcoded anon key becomes a master key giving public read/write access to all data. Fix: enable RLS and add `SELECT TO anon` policies via a migration (`ALTER TABLE ... ENABLE ROW LEVEL SECURITY; CREATE POLICY ...`) before the dashboard makes its first live query.

2. **Vercel env vars not injected at runtime in zero-build static sites** — `process.env.SUPABASE_URL` is `undefined` at runtime without a build step. Fix: hardcode the anon key directly in `config.js` (it is a public key by design) and add `config.js` to `.gitignore` before the first commit. Provide `config.example.js` as a setup template.

3. **Tailwind Play CDN causes FOUC on dynamic content** — kanban cards inserted via JS render briefly without styles. Fix: either use the Tailwind CLI standalone binary (no npm required) to pre-compile CSS, or mask the flash with `body { opacity: 0 }` revealed after CDN loads.

4. **`fetch()` resolves on HTTP errors — silent empty dashboard** — a 401 (wrong key), 403 (RLS blocking), or Supabase pause returns a resolved promise with `response.ok === false`. Without an explicit check, the dashboard shows empty columns with no error. Fix: always check `if (!res.ok) throw new Error(...)` in the fetch wrapper and display a visible error banner in the UI.

5. **ES module imports require `.js` extension** — `import { x } from './api'` fails in the browser (no bundler resolution). Fix: always use `import { x } from './api.js'` from the first line of code. Establish this as a project convention before writing any modules.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Foundation and Security
**Rationale:** Two of the five critical pitfalls (RLS misconfiguration and config strategy) must be resolved before writing any JS that touches the database. Getting this wrong after data-fetching code is in place creates rework. The project structure decisions (hash routing, module conventions, gitignore setup) made here govern all subsequent phases.
**Delivers:** Supabase tables secured with RLS + SELECT policies; `config.js` (gitignored) with verified credentials; `index.html` shell with Tailwind CDN; `router.js` with hash routing to stub views; module file convention established (`.js` extensions enforced)
**Addresses:** Table stakes (fresh data from secured database), pitfalls 1, 2, 4, 5
**Avoids:** RLS data exposure; silent fetch failures; broken imports; key committed to git

### Phase 2: Data Layer and Components
**Rationale:** Before building any view, prove the Supabase connection works and establish reusable renderers with mock data. Building components against mock data lets layout and styling be validated without API dependency. The fetch wrapper's error handling must be correct from the start per pitfall 4.
**Delivers:** `api/supabase.js` fetch wrapper with `response.ok` check; `components/kanban.js` and `components/table.js` built against hardcoded mock data; loading and error states implemented
**Uses:** Raw fetch PostgREST pattern; Tailwind utility classes for kanban column layout
**Implements:** Fetch wrapper component, pure renderer components

### Phase 3: Books View (Simplest Pipeline View)
**Rationale:** Books is the simplest view (flat table, no grouping logic, no status columns) and validates the full fetch-to-render pipeline end-to-end before tackling kanban complexity.
**Delivers:** Working Books tab with live Supabase data; navigation tab switching working
**Implements:** `views/books.js`, `components/table.js` wired to live data

### Phase 4: Ideas Kanban
**Rationale:** Core value proposition. Four status columns (pending/approved/rejected/scripted) with item counts. Adds book title to cards (requires loading books data alongside ideas). This is the primary bottleneck-visibility view.
**Delivers:** Ideas tab with live kanban; column counts; book name on cards; empty column states
**Implements:** `views/ideas.js` with grouping logic, `components/kanban.js` with column count display

### Phase 5: Calendar Kanban and Polish
**Rationale:** Same pattern as ideas kanban but with 5 columns and two additional concerns: per-platform color coding and the `failed` alarm state. Built last because it reuses everything from Phase 4 and adds visual polish on top.
**Delivers:** Calendar tab with live kanban; platform color coding; red alarm on `failed` cards; scheduled date on cards; manual refresh button; last-refreshed timestamp
**Implements:** `views/calendar.js`, refresh/timestamp logic wired to all views

### Phase Ordering Rationale

- Security (Phase 1) must precede data access (Phase 2) — RLS not retroactively fixable without risk window
- Components (Phase 2) built with mock data before views (Phases 3-5) — decouples layout work from API dependency
- Books (Phase 3) before kanbans (Phases 4-5) — validates full pipeline without grouping complexity
- Calendar (Phase 5) after Ideas (Phase 4) — reuses identical kanban pattern; platform coloring and failed alarm are incremental additions
- Polish (refresh, timestamp, empty states) in Phase 5 — deferred until all views exist so refresh logic is wired once to all three views

### Research Flags

Phases with well-documented patterns (research-phase not needed):
- **Phase 1:** RLS SQL and `.gitignore` patterns are straightforward and well-documented; use PITFALLS.md SQL snippets directly
- **Phase 2:** Raw fetch + PostgREST pattern is fully specified in ARCHITECTURE.md with working code examples
- **Phase 3:** Books table is a simple `SELECT *` with no joins or grouping
- **Phase 4:** Ideas kanban grouping is a `Array.filter()` by status — no unknowns
- **Phase 5:** Calendar kanban is identical to ideas pattern; Tailwind color utilities for platform/status are trivial

No phases need `/gsd:research-phase`. All patterns are documented with working code examples in the research files.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All technology choices verified against official docs; CDN URLs pinned to specific versions |
| Features | MEDIUM | No direct competitors for this personal-tool niche; grounded in UX research and kanban literature but not domain-specific benchmarks |
| Architecture | HIGH | Hash router, fetch wrapper, and pure component patterns are well-established vanilla JS patterns; cross-validated against existing Go CLI HTTP patterns in this codebase |
| Pitfalls | HIGH | RLS, CORS, Vercel env vars, and Tailwind CDN limits all confirmed via official docs and 2025 security audits |

**Overall confidence:** HIGH

### Gaps to Address

- **Tailwind CDN vs CLI binary decision:** STACK.md recommends CDN (acceptable for personal tool); PITFALLS.md recommends CLI binary (to avoid FOUC). Decision must be made in Phase 1. Recommendation: start with CDN + opacity-mask FOUC mitigation; migrate to CLI binary only if FOUC is visibly problematic in practice.
- **`config.js` in git vs gitignored:** STACK.md says hardcoding is acceptable; PITFALLS.md warns about GitHub secret scanning auto-revoking keys. Since this repo is private, hardcoding in a gitignored `config.js` (with `config.example.js` template) is the right call. Must add `.gitignore` entry before the first commit.
- **Supabase key type:** Verify whether this project uses legacy anon key or new `sb_publishable_` key format. Project predates November 2025 cutoff so legacy anon key is expected. Confirm in Supabase dashboard before writing config.

## Sources

### Primary (HIGH confidence)
- [Supabase JS Reference — Installing](https://supabase.com/docs/reference/javascript/installing)
- [Supabase — Understanding API Keys](https://supabase.com/docs/guides/api/api-keys)
- [Supabase — Row Level Security](https://supabase.com/docs/guides/database/postgres/row-level-security)
- [Supabase GitHub Discussion #29260 — Upcoming changes to API Keys](https://github.com/orgs/supabase/discussions/29260)
- [Tailwind CSS — Play CDN (v4)](https://tailwindcss.com/docs/installation/play-cdn)
- [Vercel — Environment Variables](https://vercel.com/docs/projects/environment-variables)
- [JavaScript ES Modules — MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules)
- [Supabase Security Retro 2025](https://supabase.com/blog/supabase-security-2025-retro)

### Secondary (MEDIUM confidence)
- [Dashboard Design Principles — UXPin](https://www.uxpin.com/studio/blog/dashboard-design-principles/)
- [From Good To Great In Dashboard Design — Smashing Magazine](https://www.smashingmagazine.com/2021/11/dashboard-design-research-decluttering-data-viz/)
- [What is a Kanban Board — Atlassian](https://www.atlassian.com/agile/kanban/boards)
- [Supabase CORS for REST API — wildcard origin](https://github.com/orgs/supabase/discussions/7038)
- [Supabase Security Flaw: 170+ Apps Exposed by Missing RLS](https://byteiota.com/supabase-security-flaw-170-apps-exposed-by-missing-rls/)

### Tertiary (LOW confidence)
- [State Management in Vanilla JS: 2026 Trends](https://medium.com/@chirag.dave/state-management-in-vanilla-js-2026-trends-f9baed7599de) — no shared state needed, finding not acted upon

---
*Research completed: 2026-02-25*
*Ready for roadmap: yes*
