# Phase 1: Foundation - Research

**Researched:** 2026-02-25
**Domain:** Static HTML/JS dashboard — project scaffolding, Supabase RLS, config management, hash routing, Tailwind CDN
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| INFRA-01 | Dashboard deployed as static site on Vercel from `dashboard/` folder | Vercel auto-detects root `index.html`; `vercel.json` with `outputDirectory: "dashboard"` needed since dashboard is a subdirectory of the CLI repo |
| INFRA-02 | Supabase connection via supabase-js v2 (CDN ESM), configured in `dashboard/config.js` | supabase-js v2 is importable as ESM from jsDelivr CDN; `config.js` is the single source for URL + key |
| INFRA-03 | `dashboard/config.js` added to `.gitignore` (contains anon key, private repo) | Gitignore must have entry added before `config.js` is created; `config.example.js` serves as setup template |
| INFRA-04 | RLS enabled on all 5 tables with SELECT-only policy for `anon` role | Existing migration has `FOR ALL USING (true)` policies — these allow anon INSERT/UPDATE/DELETE. A new migration must replace them with SELECT-only anon policies |
| INFRA-05 | Tailwind CSS v4 CDN loaded as script tag in `index.html` | CDN tag is `<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>`; FOUC mitigation required for dynamically rendered content |
</phase_requirements>

---

## Summary

Phase 1 creates the entire skeleton for the dashboard: the `dashboard/` folder with `index.html`, the Supabase client config, a Vercel deployment config, hash routing stubs, and the RLS migration. All five INFRA requirements can be satisfied with static file creation plus one SQL migration.

The single most important finding for planning is that **RLS is already enabled on all tables but the current policies allow `FOR ALL` (not SELECT-only) and use no role restriction.** This means the anon key currently has full read/write/delete access to all tables. A new migration must drop the existing policies and create SELECT-only policies scoped to the `anon` role before the dashboard ships.

The Tailwind CDN vs CLI binary decision should be made in Wave 1 before any HTML is written. The recommended approach is CDN + FOUC mitigation (`body { opacity: 0 }` revealed post-load) since there is no npm toolchain and the Tailwind CLI binary requires a local binary download. This is a personal tool where a brief flash is acceptable, but the mitigation must be in place from the start.

**Primary recommendation:** Scaffold `dashboard/` structure first, add `.gitignore` entry second, create `config.js` third, write RLS migration fourth — in that strict order to avoid security missteps.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| HTML5 + Vanilla JS ES Modules | ES2022+ | Structure and logic | Zero dependencies, no bundler, works on Vercel static; `<script type="module">` supported in all modern browsers |
| @supabase/supabase-js | v2 (2.97.0) | Supabase client | Handles auth headers, PostgREST query builder, error parsing; ~100KB CDN overhead acceptable for internal tool |
| Tailwind CSS Play CDN | v4 (`@tailwindcss/browser@4`) | Utility styling | Zero-build single script tag; JIT compilation overhead irrelevant for personal desktop tool |
| Vercel Static Hosting | — | Deployment | Auto-deploys on `git push`; `vercel.json` with `outputDirectory` needed for subdirectory |

### CDN URLs (Pinned)
```html
<!-- Tailwind v4 CDN — add to <head> BEFORE body content -->
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>

<!-- Supabase JS v2 ESM — use in <script type="module"> -->
import { createClient } from 'https://cdn.jsdelivr.net/npm/@supabase/supabase-js@2.97.0/+esm'
```

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| supabase-js CDN ESM | Raw `fetch` against PostgREST | Raw fetch saves ~100KB but requires manual header management on every call; supabase-js adds query builder for filtering |
| Tailwind Play CDN | Tailwind CLI standalone binary | CLI binary eliminates FOUC and `@apply` limitations but adds a local binary step with no npm; CDN is acceptable for personal tool |
| Vercel + `vercel.json` | GitHub Pages | Vercel is simpler (no `gh-pages` branch needed) and already in scope |

---

## Architecture Patterns

### Recommended Project Structure
```
dashboard/
├── index.html           # Shell: navbar + #app div + CDN script tags
├── config.js            # Supabase URL + anon key — GITIGNORED
├── config.example.js    # Template for setup — committed
├── js/
│   ├── app.js           # Entry point: imports router, bootstraps on DOMContentLoaded
│   ├── router.js        # hashchange listener → dispatches to view functions
│   ├── api.js           # Supabase createClient + query functions (fetchBooks, etc.)
│   └── views/
│       ├── books.js     # Stub: renders placeholder content in #app
│       ├── ideas.js     # Stub: renders placeholder content in #app
│       └── calendar.js  # Stub: renders placeholder content in #app
vercel.json              # { "outputDirectory": "dashboard" }
```

### Pattern 1: Hash Router
**What:** Listen for `hashchange` and `load` events; map `location.hash` to view function calls.
**When to use:** All navigation in this project — avoids Vercel 404 on direct URL access with no server-side routing.
**Example:**
```javascript
// dashboard/js/router.js
import { renderBooks } from './views/books.js';
import { renderIdeas } from './views/ideas.js';
import { renderCalendar } from './views/calendar.js';

const routes = {
  '#/books': renderBooks,
  '#/ideas': renderIdeas,
  '#/calendar': renderCalendar,
};

function dispatch() {
  const hash = location.hash || '#/ideas';
  const view = routes[hash] ?? renderIdeas;
  view();
}

window.addEventListener('hashchange', dispatch);
window.addEventListener('load', dispatch);
export { dispatch };
```

### Pattern 2: Config Module
**What:** Single `config.js` exports Supabase URL and anon key; all modules import from it.
**When to use:** Everywhere in this project — provides a single place to change credentials.
**Example:**
```javascript
// dashboard/config.js  — LISTED IN .gitignore
export const SUPABASE_URL = "https://YOUR_PROJECT.supabase.co";
export const SUPABASE_ANON_KEY = "eyJ...";

// dashboard/config.example.js  — committed as template
export const SUPABASE_URL = "https://YOUR_PROJECT.supabase.co";
export const SUPABASE_ANON_KEY = "your-anon-key-here";
```

### Pattern 3: FOUC Mitigation for Tailwind CDN
**What:** Hide body until Tailwind CDN finishes JIT compilation.
**When to use:** Required when using Tailwind Play CDN with dynamically rendered content.
**Example:**
```html
<!-- In <head>, immediately after Tailwind CDN script tag -->
<style>body { opacity: 0; transition: opacity 0.1s; }</style>
<script>
  document.addEventListener('DOMContentLoaded', function() {
    // Give Tailwind time to process classes
    requestAnimationFrame(function() {
      document.body.style.opacity = '1';
    });
  });
</script>
```

### Anti-Patterns to Avoid
- **Import without `.js` extension:** `import x from './router'` fails in browser — no bundler resolution. Always use `import x from './router.js'`.
- **Path-based routing (`/ideas`):** Causes Vercel 404 on refresh without server config. Use hash routing (`#/ideas`) exclusively.
- **Using `process.env` for config:** Undefined at runtime in zero-build static sites. Use gitignored `config.js` instead.
- **Mixed `<script>` and `<script type="module">`:** Classic scripts don't have access to ES module scope. Use a single entry point via `type="module"`.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Supabase auth headers + query construction | Manual `fetch` with headers on every call | `@supabase/supabase-js` CDN ESM | Query builder, consistent error format, less boilerplate |
| RLS policy management | Custom middleware or per-query key logic | Supabase native RLS with `CREATE POLICY` | Database-level enforcement; not bypassable from client |
| CSS utility system | Custom `app.css` with spacing/color scales | Tailwind CDN | Kanban column layout needs utility classes; custom CSS for 5 columns is 200+ lines |
| Build step for config injection | Template substitution scripts | Gitignored `config.js` | Zero complexity; anon key is public-by-design anyway |

---

## Common Pitfalls

### Pitfall 1: Existing RLS Policies Are Too Permissive
**What goes wrong:** Migration `001_initial_schema.sql` already created RLS policies but they use `FOR ALL USING (true)` with no role restriction. This means the `anon` role can INSERT, UPDATE, and DELETE data — not just SELECT. The dashboard hardcodes the anon key, so this is the live security posture.
**Why it happens:** The CLI was the only client so full access for `authenticated` role was fine. The policy names say "authenticated users" but the `USING (true)` clause with `FOR ALL` applies to all roles including `anon`.
**How to avoid:** Write migration `009_dashboard_rls_anon_select.sql` that drops the existing `FOR ALL` policies and creates `FOR SELECT TO anon` policies on all five tables.
**Warning signs:** `SELECT policyname, cmd, roles FROM pg_policies WHERE schemaname = 'public';` shows `ALL` as `cmd` and empty `roles` array.

### Pitfall 2: `config.js` Added to Git Before `.gitignore` Entry
**What goes wrong:** Developer creates `config.js` with credentials and commits before adding the gitignore entry. The key is now in git history even after gitignore is added. GitHub Secret Scanning may auto-revoke the key.
**Why it happens:** `.gitignore` is added in a follow-up commit after the file already exists.
**How to avoid:** Add `dashboard/config.js` to `.gitignore` as the very first commit in Phase 1, before creating the file. Order: `.gitignore` entry → `config.js` (already ignored) → git commit.
**Warning signs:** `git log --all -p -- dashboard/config.js` shows the key in history.

### Pitfall 3: ES Module Import Missing `.js` Extension
**What goes wrong:** `import { dispatch } from './router'` throws `Failed to resolve module specifier` in the browser.
**Why it happens:** Bundlers resolve bare specifiers; browsers do not. Browser fetches the literal URL `./router` which returns 404.
**How to avoid:** Enforce `.js` extension on every import from the first line of code. Add a comment in `app.js` header noting this project convention.
**Warning signs:** Console error `Failed to resolve module specifier` or 404 on module URL.

### Pitfall 4: Vercel Deploys CLI Repo Root Instead of `dashboard/` Subdirectory
**What goes wrong:** Vercel detects the Go repo and either fails the build or serves `main.go` as a static file.
**Why it happens:** Vercel auto-detects framework. Without `vercel.json`, it may apply a Go framework preset or serve the wrong directory.
**How to avoid:** Commit `vercel.json` at repo root with `{ "outputDirectory": "dashboard" }` to tell Vercel which directory to serve.
**Warning signs:** Vercel deployment log shows "Detected Go" or deployed URL shows directory listing rather than the dashboard.

### Pitfall 5: Supabase Free Tier Auto-Pauses
**What goes wrong:** Dashboard opens to empty columns with no error banner. Supabase free tier pauses projects after 7 days of inactivity.
**Why it happens:** The existing CLI activity keeps the project alive, but if CLI hasn't been run recently, the project may be paused.
**How to avoid:** Run `gagipress books list` before opening the dashboard to wake the project. Error handling in the fetch layer (Phase 2) will surface this as a visible error when it happens.
**Warning signs:** Dashboard shows empty state on all columns; Supabase dashboard shows "Project paused" banner.

---

## Code Examples

### RLS Migration — SELECT-only Anon Policies
```sql
-- migrations/009_dashboard_rls_anon_select.sql
-- Drop existing over-permissive policies
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON books;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON content_ideas;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON content_scripts;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON content_calendar;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON post_metrics;
DROP POLICY IF EXISTS "Enable all access for authenticated users" ON sales_data;

-- Re-create full access for authenticated role (CLI operations)
CREATE POLICY "authenticated full access" ON books
  FOR ALL TO authenticated USING (true) WITH CHECK (true);
CREATE POLICY "authenticated full access" ON content_ideas
  FOR ALL TO authenticated USING (true) WITH CHECK (true);
CREATE POLICY "authenticated full access" ON content_scripts
  FOR ALL TO authenticated USING (true) WITH CHECK (true);
CREATE POLICY "authenticated full access" ON content_calendar
  FOR ALL TO authenticated USING (true) WITH CHECK (true);
CREATE POLICY "authenticated full access" ON post_metrics
  FOR ALL TO authenticated USING (true) WITH CHECK (true);
CREATE POLICY "authenticated full access" ON sales_data
  FOR ALL TO authenticated USING (true) WITH CHECK (true);

-- Add SELECT-only policy for anon role (dashboard read access)
CREATE POLICY "anon read-only" ON books
  FOR SELECT TO anon USING (true);
CREATE POLICY "anon read-only" ON content_ideas
  FOR SELECT TO anon USING (true);
CREATE POLICY "anon read-only" ON content_scripts
  FOR SELECT TO anon USING (true);
CREATE POLICY "anon read-only" ON content_calendar
  FOR SELECT TO anon USING (true);
CREATE POLICY "anon read-only" ON post_metrics
  FOR SELECT TO anon USING (true);
CREATE POLICY "anon read-only" ON sales_data
  FOR SELECT TO anon USING (true);
```

### index.html Shell
```html
<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Gagipress Dashboard</title>
  <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
  <style>body { opacity: 0; transition: opacity 0.15s; }</style>
  <script>
    document.addEventListener('DOMContentLoaded', function() {
      requestAnimationFrame(() => document.body.style.opacity = '1');
    });
  </script>
</head>
<body class="bg-gray-900 text-gray-100 min-h-screen">
  <nav id="navbar">
    <a href="#/books">Books</a>
    <a href="#/ideas">Ideas</a>
    <a href="#/calendar">Calendar</a>
  </nav>
  <main id="app">
    <!-- Views render here -->
  </main>
  <script type="module" src="js/app.js"></script>
</body>
</html>
```

### vercel.json (Subdirectory Deploy)
```json
{
  "outputDirectory": "dashboard"
}
```

### Audit Query — Verify RLS State
```sql
-- Run in Supabase SQL editor to confirm state before and after migration
SELECT
  tablename,
  rowsecurity AS rls_enabled
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY tablename;

SELECT
  tablename,
  policyname,
  cmd,
  roles
FROM pg_policies
WHERE schemaname = 'public'
ORDER BY tablename, policyname;
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Legacy anon key (`eyJ...`) | `sb_publishable_...` key | June 2025 (new projects after Nov 2025) | This project predates cutoff — use legacy anon key from Supabase dashboard |
| `supabase-js` npm install | CDN ESM import via jsDelivr | supabase-js v2 added ESM exports | No npm required; works in zero-build static sites |
| Tailwind CDN `<script src="https://cdn.tailwindcss.com">` | `@tailwindcss/browser@4` from jsDelivr | Tailwind v4 | New CDN endpoint; old URL still works but pinned version URL is preferred |

---

## Open Questions

1. **Service role key for CLI vs anon key for dashboard**
   - What we know: The CLI uses `ServiceKey` (bypasses RLS) for all operations; the dashboard uses `AnonKey` (subject to RLS)
   - What's unclear: Whether the Supabase project currently has a service role key configured that the CLI reads — need to verify the CLI still functions after RLS policies are replaced
   - Recommendation: Audit RLS state before migration; run `gagipress books list` after applying migration to confirm CLI still works (it uses service key which bypasses RLS)

2. **`content_scripts` table — not in REQUIREMENTS.md**
   - What we know: Migration `001` created a `content_scripts` table; REQUIREMENTS.md lists only `books`, `content_ideas`, `content_calendar`, `post_metrics`, `sales_data`
   - What's unclear: Whether this table is still used or was superseded by the `generated_script` column in `content_ideas`
   - Recommendation: Include `content_scripts` in the RLS migration for safety; the dashboard does not query it in Phase 1

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package (existing) — no JS test framework for dashboard |
| Config file | none — static site has no test config |
| Quick run command | `make vet` (Go linting; no JS tests in Phase 1) |
| Full suite command | `make test` (Go unit tests; no JS tests in Phase 1) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| INFRA-01 | `dashboard/index.html` exists and Vercel deploys it | manual | Open Vercel URL, confirm styled page renders | ❌ Wave 0 (manual verification) |
| INFRA-02 | `dashboard/config.js` exports valid Supabase URL + key | manual | Load `index.html` locally; DevTools Network shows Supabase request | ❌ Wave 0 (manual verification) |
| INFRA-03 | `dashboard/config.js` is in `.gitignore`; `config.example.js` is committed | smoke | `git check-ignore -v dashboard/config.js` returns match | ❌ Wave 0 |
| INFRA-04 | All 5 tables have RLS + SELECT-only anon policy | manual | Supabase SQL: `SELECT * FROM pg_policies WHERE schemaname='public'` shows `SELECT` + `{anon}` | ❌ Wave 0 |
| INFRA-05 | Tailwind classes render correctly on static elements | manual | Open `index.html`; confirm navbar has correct bg-color, no FOUC | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `make vet` (Go code unchanged; verify no regressions)
- **Per wave merge:** `make test` (full Go suite)
- **Phase gate:** Manual checklist (5 success criteria from phase definition) before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `dashboard/index.html` — does not exist yet
- [ ] `dashboard/config.js` + `dashboard/config.example.js` — does not exist yet
- [ ] `dashboard/vercel.json` (at repo root) — does not exist yet
- [ ] `migrations/009_dashboard_rls_anon_select.sql` — does not exist yet
- [ ] `dashboard/js/app.js`, `router.js`, `api.js`, views stubs — does not exist yet
- [ ] `.gitignore` entry for `dashboard/config.js` — not yet present

---

## Sources

### Primary (HIGH confidence)
- [Supabase — Row Level Security](https://supabase.com/docs/guides/database/postgres/row-level-security) — RLS SQL patterns
- [Supabase — API Keys](https://supabase.com/docs/guides/api/api-keys) — anon key safety model
- [Tailwind CSS — Play CDN v4](https://tailwindcss.com/docs/installation/play-cdn) — CDN script tag, FOUC behavior
- [Vercel — Configuring a Build](https://vercel.com/docs/builds/configure-a-build) — `outputDirectory` for subdirectory
- [MDN — ES Modules](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules) — `.js` extension requirement
- `.planning/research/STACK.md` — verified stack decisions
- `.planning/research/PITFALLS.md` — pitfall catalog with mitigations
- `migrations/001_initial_schema.sql` — confirmed existing RLS policies are `FOR ALL` (too permissive)

### Secondary (MEDIUM confidence)
- [Supabase GitHub Discussion #29260](https://github.com/orgs/supabase/discussions/29260) — `sb_publishable_` key format; this project uses legacy key (predates Nov 2025)
- [Supabase Security Retro 2025](https://supabase.com/blog/supabase-security-2025-retro) — RLS misconfiguration prevalence

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — CDN URLs pinned to specific versions; verified against official docs
- Architecture: HIGH — hash router + ES modules + config module are well-established vanilla JS patterns
- RLS migration: HIGH — existing migration audited directly; SQL patterns from official Supabase docs
- Pitfalls: HIGH — discovered directly from codebase inspection (existing `FOR ALL` policies) and official docs

**Research date:** 2026-02-25
**Valid until:** 2026-03-25 (stable domain; Supabase RLS API not changing)
