# Technology Stack

**Project:** Gagipress Dashboard — Static read-only web dashboard
**Researched:** 2026-02-25
**Confidence:** MEDIUM-HIGH (Supabase client/key verified via docs; Tailwind CDN production caveat verified; Vercel static deploy verified)

---

## Recommended Stack

### Core Layer

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| HTML5 | — | Structure | Zero dependency, zero tooling, works on Vercel static |
| Vanilla JS ES Modules | ES2022+ | Logic + DOM | `<script type="module">` works in all modern browsers, no bundler needed |
| Tailwind CSS (Play CDN) | v4 (`@tailwindcss/browser@4`) | Utility styling | Zero-build, single `<script>` tag. Acceptable for personal internal tools where performance is not the primary concern. See caveat below. |
| `@supabase/supabase-js` | v2 (latest: 2.97.0) | Supabase client | Handles auth headers, PostgREST query builder, error parsing — less code than raw fetch |

### CDN URLs (Pinned Imports)

```html
<!-- Tailwind CSS v4 CDN -->
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>

<!-- Supabase JS v2 ESM (use in <script type="module">) -->
<!-- Always pin to a specific version in production to avoid breaking changes -->
import { createClient } from 'https://cdn.jsdelivr.net/npm/@supabase/supabase-js@2.97.0/+esm'
```

### Hosting

| Technology | Purpose | Why |
|------------|---------|-----|
| Vercel (Static) | Deploy and serve | Free tier, GitHub push-to-deploy, no server, no config needed for pure HTML |

---

## Supabase Client: JS Client vs Raw Fetch

**Recommendation: Use `@supabase/supabase-js` via CDN ESM.**

**Rationale:**

Raw fetch against PostgREST is viable but requires manually constructing query strings (`?select=*&status=eq.pending`), setting three headers on every call (`apikey`, `Authorization`, `Content-Type`), and parsing error responses yourself. For a read-only dashboard with 5 tables and filtering logic, this is 80-100 lines of boilerplate that the supabase-js client eliminates.

The supabase-js client at v2 ships as a proper ESM module importable from jsDelivr CDN with no build step. Bundle size is ~100KB (includes realtime, auth, storage — none of which this project uses, but the overhead is acceptable for an internal tool loaded once and cached).

**When raw fetch wins:** If you have exactly one query and no filtering, raw fetch saves the CDN load. Not the case here.

**Do NOT use:** The `supabase` npm package (different, older wrapper). Only `@supabase/supabase-js`.

---

## Tailwind CDN: Caveat and Acceptance Criteria

**The Tailwind Play CDN (`@tailwindcss/browser@4`) is officially "not for production."**

What that means practically: JIT compilation runs in the browser on every page load, adding ~50-100ms first-render cost and downloading a ~100KB JS runtime. For a **personal internal tool** (1 user, desktop-only, fast connection), this cost is irrelevant.

**Acceptance criteria for using Tailwind CDN here:**
- [x] Personal use, not public-facing with SEO or Core Web Vitals requirements
- [x] No npm/build tooling constraint is hard
- [x] Desktop-only, fast connection
- [x] Dark mode theming via `@theme` block in `<style type="text/tailwindcss">`

**If build step is ever added later:** Switch to Tailwind CLI standalone binary (single binary, no npm install required) and pipe output CSS. This is a 30-minute migration.

**Alternative rejected: Pico CSS / MVP.css / Water.css** — semantic CSS frameworks that style bare HTML. Rejected because they do not support kanban column layouts without custom CSS. Tailwind's utility approach is more appropriate for a kanban-style layout.

**Alternative rejected: Bootstrap CDN** — large footprint, opinionated component styles that conflict with dark minimal aesthetic.

---

## Supabase API Key: Which Key to Use

**Recommendation: Publishable key (`sb_publishable_...`) if your project was created after November 2025; otherwise use the legacy anon key.**

**Context (HIGH confidence, verified against Supabase docs):**

- Supabase launched new API key format in June 2025: `sb_publishable_...` replaces the anon key for client-side use
- New projects created after November 1, 2025 no longer have legacy `anon` and `service_role` keys
- Legacy anon keys still work on existing projects and are safe to use until the deprecation deadline
- The `sb_publishable_` key is safe to expose in client-side code — it has the same permissions as the anon key, enforced by RLS

**Security model for this project (HIGH confidence):**

The anon/publishable key is safe to expose in browser code **when RLS is enabled on all tables**. Supabase's security model is:
- Anon key = door handle (anyone can grab it)
- RLS policies = the actual lock

For this dashboard (read-only, personal, no auth):
1. RLS must be enabled on: `books`, `content_ideas`, `content_calendar`, `post_metrics`, `sales_data`
2. Each table needs a permissive SELECT policy for the `anon` role: `USING (true)`
3. No INSERT/UPDATE/DELETE policies for anon — the dashboard only reads

**Never expose the service role key** — it bypasses RLS entirely.

**On key placement in static files:**

Vercel environment variables are only injected at build time (into the build process). For a no-build-step static site, there is no build process, so Vercel env vars do not help. The two valid approaches are:

1. **Hardcode the publishable/anon key directly in the JS file.** This is the intended pattern for Supabase client-side keys. The key is public by design. For a personal internal tool not linked publicly, this is fully acceptable.
2. **Fetch from a `config.json` excluded from git.** More complex, no real security benefit for anon keys.

**Decision for this project: Hardcode the publishable key in `supabase.js` config module. The key is designed to be public. The URL is the only thing to protect (keep Vercel deployment URL unlinked from public indexes).**

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| JS client | `@supabase/supabase-js` v2 ESM | Raw `fetch()` against PostgREST | Too much boilerplate for 5-table filtering dashboard |
| CSS | Tailwind Play CDN v4 | Bootstrap CDN | Heavier, opinionated components conflict with dark minimal design |
| CSS | Tailwind Play CDN v4 | Pico CSS / Water.css | No utility classes — kanban layout requires too much custom CSS |
| CSS | Tailwind Play CDN v4 | Tailwind CLI standalone | Adds one CLI step; CDN is acceptable for personal internal tool |
| Hosting | Vercel static | GitHub Pages | Vercel is simpler (no `gh-pages` branch, direct root deploy) |
| Hosting | Vercel static | Netlify | Equivalent, but Vercel is already in scope per PROJECT.md |
| Auth | None | Vercel password protection | Not needed for personal tool not publicly indexed |

---

## Project Structure

```
dashboard/               # or root of repo if separate repo
├── index.html           # Main dashboard (kanban + books)
├── js/
│   ├── supabase.js      # createClient export (key + URL here)
│   ├── api.js           # Query functions (fetchIdeas, fetchCalendar, fetchBooks)
│   └── ui.js            # DOM rendering (kanban columns, tables)
├── css/
│   └── app.css          # Custom dark theme tokens (if needed beyond Tailwind)
└── vercel.json          # Optional: only needed for custom routing/headers
```

**Note on ES module organization:** Keep `supabase.js` as a dedicated config module so the key is in exactly one place. All other modules import from it.

---

## Vercel Deployment

No `vercel.json` required for a root-level `index.html` static site. Vercel auto-detects it.

**If the dashboard lives in a subdirectory of the existing CLI repo:**
```json
{
  "outputDirectory": "dashboard"
}
```

**Custom headers for security (optional but recommended):**
```json
{
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        { "key": "X-Frame-Options", "value": "DENY" },
        { "key": "X-Content-Type-Options", "value": "nosniff" }
      ]
    }
  ]
}
```

---

## Installation

No `npm install`. No build step. Zero local tooling required.

**To develop locally:**
```bash
# Serve the dashboard directory with any static server
npx serve dashboard/
# or
python3 -m http.server 8080
```

**To deploy:**
```bash
git push origin main   # Vercel auto-deploys on push
```

---

## Confidence Assessment

| Decision | Confidence | Source |
|----------|------------|--------|
| supabase-js v2 ESM via jsDelivr | HIGH | npm registry (v2.97.0 confirmed), official Supabase docs |
| Tailwind v4 Play CDN tag | HIGH | Official Tailwind docs (play-cdn), confirmed not-for-production caveat |
| Anon key safe to expose with RLS | HIGH | Official Supabase docs (api-keys, securing-your-api) |
| New `sb_publishable_` key format | HIGH | Supabase GitHub discussion #29260, official changelog |
| Vercel root-level static HTML deploy | HIGH | Official Vercel docs + multiple implementation guides |
| Tailwind CDN acceptable for personal tool | MEDIUM | Engineering judgment applied to "not for production" caveat |
| Raw fetch vs supabase-js recommendation | MEDIUM | Based on feature analysis; no direct 2025 benchmark found |

---

## Sources

- [Supabase JS Reference — Installing](https://supabase.com/docs/reference/javascript/installing)
- [Supabase — Understanding API Keys](https://supabase.com/docs/guides/api/api-keys)
- [Supabase — Securing Your Data](https://supabase.com/docs/guides/database/secure-data)
- [Supabase — Row Level Security](https://supabase.com/docs/guides/database/postgres/row-level-security)
- [Supabase GitHub Discussion #29260 — Upcoming changes to API Keys](https://github.com/orgs/supabase/discussions/29260)
- [@supabase/supabase-js on npm (v2.97.0)](https://www.npmjs.com/package/@supabase/supabase-js)
- [@supabase/supabase-js on jsDelivr CDN](https://www.jsdelivr.com/package/npm/@supabase/supabase-js)
- [Tailwind CSS — Play CDN (v4)](https://tailwindcss.com/docs/installation/play-cdn)
- [Vercel — Environment Variables](https://vercel.com/docs/projects/environment-variables)
- [Vercel — Configuring a Build (static sites)](https://vercel.com/docs/builds/configure-a-build)
