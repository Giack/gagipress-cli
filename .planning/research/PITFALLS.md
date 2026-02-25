# Domain Pitfalls

**Domain:** Static vanilla JS dashboard — Supabase REST + Tailwind CDN + Vercel static hosting
**Researched:** 2026-02-25
**Confidence:** HIGH (Supabase RLS, CORS, Tailwind CDN limits confirmed via official docs and 2025 sources)

---

## Critical Pitfalls

Mistakes that cause rewrites, data leaks, or broken deployments.

---

### Pitfall 1: RLS Not Enabled on Tables — Anon Key Becomes a Master Key

**What goes wrong:** The Supabase anon key is hardcoded in client-side JS (by design — it is a public key). But if Row Level Security is disabled on any table, anyone who inspects the page source or network tab gets full read (and potentially write/delete) access to that table's data. The anon key alone provides zero data protection without RLS.

**Why it happens:** Tables created via SQL migrations (not the Supabase dashboard) have RLS disabled by default. The CLI migrations in this project (`migrations/001_*.sql` etc.) may not have included `ENABLE ROW LEVEL SECURITY` or `CREATE POLICY` statements. 83% of exposed Supabase databases involve RLS misconfigurations (per 2025 security audits).

**Consequences:** All five tables (`books`, `content_ideas`, `content_calendar`, `post_metrics`, `sales_data`) become publicly readable — and writable/deletable — by anyone with the project URL.

**Prevention:**
1. Audit every table: `SELECT tablename, rowsecurity FROM pg_tables WHERE schemaname = 'public';`
2. Enable RLS on all tables: `ALTER TABLE books ENABLE ROW LEVEL SECURITY;` (repeat for each table)
3. Create a permissive SELECT-only policy for `anon` role on each table:
   ```sql
   CREATE POLICY "anon read-only" ON books
     FOR SELECT TO anon USING (true);
   ```
4. Block INSERT/UPDATE/DELETE for anon (default when no policy exists for those operations, but verify)
5. Run Supabase Security Advisor from dashboard after enabling

**Detection warning signs:**
- `SELECT tablename, rowsecurity FROM pg_tables WHERE schemaname = 'public';` shows `f` (false) for any table
- Supabase dashboard shows "RLS disabled" warning on table pages

**Phase:** Address in Phase 1 (before the dashboard reads any data). Create a migration `00X_enable_rls_policies.sql`.

---

### Pitfall 2: Vercel Environment Variables Are Not Available at Runtime in Pure Static Sites

**What goes wrong:** The Supabase URL and anon key need to reach the browser. Vercel environment variables set in the dashboard are injected at **build time** via `process.env` replacement — they require a build step. A zero-build-step static site has no build process, so `process.env.SUPABASE_URL` resolves to `undefined` at runtime. The page loads but all fetch calls fail silently.

**Why it happens:** Vercel's env var system is designed for frameworks (Next.js, SvelteKit) with build pipelines. Without a build command, there is no substitution step. For pure static HTML/JS, Vercel serves files as-is — no templating, no injection.

**Consequences:** Either the config is hardcoded in JS (acceptable for personal use, but requires care with git history), or a workaround is needed.

**Prevention (pick one approach — decide in Phase 1):**

Option A — **Hardcode directly in a `config.js` file** (simplest; acceptable for personal/private use):
```js
// dashboard/config.js  — .gitignore this file
export const SUPABASE_URL = "https://xxx.supabase.co";
export const SUPABASE_ANON_KEY = "eyJ...";
```
Add `dashboard/config.js` to `.gitignore`. Document a `config.example.js` for setup.

Option B — **`vercel.json` with a minimal build command** that replaces placeholder strings in a template file using `sed` or `envsubst`. Adds complexity but keeps secrets out of the repo.

Option C — **Accept hardcoding and rely on the Vercel URL obscurity** (personal use, anon key is public by design anyway).

**Detection warning signs:**
- `console.log(window.SUPABASE_URL)` prints `undefined` after deploy
- All fetch calls return network errors immediately

**Phase:** Decide strategy in Phase 1 before writing any fetch logic. The config loading pattern must be established first.

---

### Pitfall 3: Tailwind Play CDN Is Not Production-Grade — FOUC and Runtime Overhead

**What goes wrong:** The Tailwind Play CDN (`<script src="https://cdn.tailwindcss.com">`) generates styles at runtime by scanning the DOM. This causes a Flash of Unstyled Content (FOUC) on initial load — elements render briefly without styles before the script processes class names. For a data-heavy dashboard that renders content dynamically via JS, the FOUC is amplified: every time new DOM nodes are inserted (kanban cards, table rows), the CDN must re-scan.

**Why it happens:** The Play CDN is a JavaScript runtime that processes Tailwind utility classes — it is explicitly labelled "for development and prototypes only" in the Tailwind v4 documentation. It does not produce pre-compiled CSS.

**Consequences:**
- Visible style flash on page load (bad UX for a dashboard)
- ~90-300KB JavaScript payload that runs style processing on every render cycle
- `@apply` is not available in CDN mode — cannot write utility-composed custom classes in `<style>` tags
- CDN downtime or latency causes the entire UI to render unstyled

**Prevention:**
Use the Tailwind CLI standalone executable (zero npm required) to generate a compiled `tailwind.css`:
```bash
# One-time setup — download tailwindcss binary
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
chmod +x tailwindcss-macos-arm64
./tailwindcss-macos-arm64 -i input.css -o dashboard/tailwind.css --watch
```
This keeps zero npm dependencies while producing a proper compiled stylesheet. Add `tailwind.css` to the repo (or Vercel build step).

If the CDN is chosen anyway (for simplicity): hide the `<body>` with `opacity-0` and reveal with a tiny inline script after Tailwind loads, to mask FOUC.

**Detection warning signs:**
- Page renders with all text/structure but no colors or spacing for 100-300ms
- Browser DevTools shows `cdn.tailwindcss.com` as a render-blocking script >100ms
- Console shows: "cdn.tailwindcss.com should not be used in production"

**Phase:** Decide CDN vs CLI build in Phase 1 (affects project structure). If CDN: add FOUC mitigation. If CLI: add `tailwindcss` binary to Makefile or `package.json` scripts.

---

## Moderate Pitfalls

---

### Pitfall 4: ES Module Import Paths Require `.js` Extension — Bare Specifiers Fail

**What goes wrong:** In vanilla JS ES modules without a bundler, import statements must use explicit file extensions and relative paths. `import { fetchIdeas } from './api'` works in Node.js and bundlers — it fails in the browser with a 404. The browser has no module resolution algorithm; it fetches the literal URL.

**Why it happens:** Bundlers (webpack, vite) resolve bare specifiers as a build-time transformation. The browser cannot resolve them — it needs a full URL or explicit relative path with extension.

**Prevention:**
- Always use `import { x } from './module.js'` (with `.js`)
- Never use `import { x } from './module'` (missing extension)
- Never use bare specifiers like `import { x } from 'lodash'` without an import map
- Organize files so relative paths stay short (`./api.js`, `./ui.js` rather than `../../../../utils/api.js`)

**Detection warning signs:**
- Browser console: `Failed to resolve module specifier "..."` or 404 on import
- Modules load in Node but fail in browser

**Phase:** Phase 1 (project structure setup). Establish the module file naming convention before writing any JS.

---

### Pitfall 5: Supabase PostgREST CORS — Wildcard Origin Means No Origin Restriction

**What goes wrong:** Supabase's REST API (`/rest/v1/*`) allows requests from any origin — CORS is set to `*`. This is intentional for Supabase's hosted service. You cannot restrict the REST API to only your Vercel domain via dashboard settings. Any origin that knows the URL and anon key can query the API.

**Why it happens:** The SaaS version of Supabase overrides PostgREST's `server-cors-allowed-origins` setting at the proxy layer. The feature exists in self-hosted Supabase but not in the managed service as of 2025.

**Consequences:** For a personal read-only dashboard, this is acceptable — the anon key is public by design and RLS limits what can be read. But it means there is no network-level defense against a third party scraping your data using your anon key.

**Prevention:**
- Accept the wildcard CORS as a known constraint (RLS is your actual security layer)
- Do not treat Supabase CORS as a security boundary — it is not one for the hosted service
- Ensure RLS SELECT policies are tight enough that even if someone has the anon key, they can only read data you explicitly allow

**Detection warning signs:** This is not a runtime bug — it is a design constraint. No warning signs during development.

**Phase:** Phase 1 (architecture decisions). Document this constraint so it is not confused with a misconfiguration during debugging.

---

### Pitfall 6: Supabase anon Key Committed to Git History

**What goes wrong:** Developer hardcodes the Supabase URL and anon key in a JS file and commits it. Even if the file is later moved to `.gitignore`, the key remains in git history and is extractable via `git log -p`.

**Why it happens:** For a personal project with a public-facing anon key (which is safe to expose), the developer treats it as non-sensitive. But Supabase's GitHub Secret Scanning integration auto-revokes keys detected in public repos, causing an unexpected outage.

**Consequences:** If the repo is public, Supabase detects the key, revokes it, and the dashboard stops working with no obvious error. If the repo is private, the risk is lower but still exists if it ever goes public.

**Prevention:**
- Keep the repo private **or** use a config file excluded by `.gitignore`
- Add `dashboard/config.js` to `.gitignore` from the very first commit, before adding credentials
- Provide `dashboard/config.example.js` as a documented template
- Note: the anon key is technically safe to expose publicly (it is a publishable key by design), but automatic revocation from secret scanning is a real operational risk

**Detection warning signs:**
- Dashboard suddenly returns 401 errors after a repo visibility change
- Email from Supabase: "We detected and revoked a secret key..."

**Phase:** Phase 1 (project setup). Add `.gitignore` entry before creating the config file.

---

### Pitfall 7: No Error State Handling for Failed Fetch — Silent Empty Dashboard

**What goes wrong:** The dashboard fetches data on load. If the anon key is wrong, RLS blocks the query, or the Supabase project is paused (free tier auto-pauses after 1 week of inactivity), all fetch calls return errors. Without explicit error handling, the dashboard renders empty kanban columns with no indication of what went wrong.

**Why it happens:** `fetch()` does not throw on HTTP errors — it resolves with a `Response` object where `response.ok === false`. Developers unfamiliar with this check assume a resolved promise means success.

**Prevention:**
```js
async function fetchIdeas() {
  const res = await fetch(`${SUPABASE_URL}/rest/v1/content_ideas?select=*`, {
    headers: { apikey: SUPABASE_ANON_KEY, Authorization: `Bearer ${SUPABASE_ANON_KEY}` }
  });
  if (!res.ok) {
    throw new Error(`Supabase error ${res.status}: ${await res.text()}`);
  }
  return res.json();
}
```
Always check `response.ok`. Show a visible error banner in the UI when fetch fails.

**Detection warning signs:**
- Kanban columns appear but are empty
- No console errors (because the promise resolved successfully with a non-2xx response)
- Supabase free tier project was inactive for >7 days

**Phase:** Phase 2 (data fetching layer). Add error handling as a first-class concern, not an afterthought.

---

## Minor Pitfalls

---

### Pitfall 8: Supabase Free Tier Project Auto-Pauses After 7 Days of Inactivity

**What goes wrong:** Supabase pauses free-tier projects after 1 week of inactivity. The dashboard will return errors or timeouts. The first request after a pause incurs a cold-start delay of 10-30 seconds.

**Prevention:** Keep the project active (the existing CLI activity is sufficient). If the dashboard is the primary usage, be aware that the CLI must be run at least weekly, or upgrade to Pro tier.

**Phase:** Operations concern. Not a development pitfall — just document it.

---

### Pitfall 9: `type="module"` Scripts Are Deferred by Default — Race Conditions with DOM

**What goes wrong:** `<script type="module">` is automatically deferred — it runs after the HTML is parsed, but only once per module (modules are singletons). If two scripts import the same module, it loads once. This is correct behavior but surprises developers used to classic scripts.

**Prevention:** Load all app logic from a single `main.js` entry point that imports other modules. Do not mix `<script type="module">` with `<script>` (classic) tags for the same logic — classic scripts run in order and are not deferred.

**Phase:** Phase 1 (project structure). Establish the single entry point pattern upfront.

---

### Pitfall 10: Vercel 404 on Direct URL Navigation (SPA Routing Without a Framework)

**What goes wrong:** If the dashboard uses hash-based navigation (`/#/ideas`, `/#/calendar`), this is fine — hashes are client-side only. But if any "multi-page" pattern uses path-based routing (`/ideas`, `/calendar`) without corresponding `.html` files, Vercel returns 404 for direct navigation or page refreshes.

**Prevention:** Use hash-based routing (`location.hash`) for all navigation, or ensure each route has a corresponding `ideas.html`, `calendar.html` file. Alternatively, add a `vercel.json` with rewrites:
```json
{
  "rewrites": [{ "source": "/(.*)", "destination": "/index.html" }]
}
```
But this only works if the dashboard is a true SPA. For a simple multi-page structure, separate HTML files are simpler.

**Phase:** Phase 1 (routing decision). Choose hash routing or multi-HTML before building navigation.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Project setup | Anon key committed to git | `.gitignore` config.js before first commit |
| RLS audit | Tables without RLS = data open to public | Migration to enable RLS + SELECT policies before dashboard reads |
| Config loading | Env vars not available at runtime in static site | Decide config strategy (gitignored file vs hardcode) in Phase 1 |
| Tailwind CDN | FOUC on dynamic content insertion | Use Tailwind CLI binary or add FOUC mitigation |
| Fetch layer | Silent failures on Supabase errors | Always check `response.ok`, show error banners |
| Routing | 404 on direct URL access | Use hash routing or add `vercel.json` rewrites |
| ES modules | Missing `.js` extension on imports | Enforce in all import statements from the start |
| Supabase free tier | Project auto-pauses after 7 days | Existing CLI activity keeps project alive |

---

## Sources

- [Supabase API Keys Documentation](https://supabase.com/docs/guides/api/api-keys) — HIGH confidence (official)
- [Supabase Row Level Security Documentation](https://supabase.com/docs/guides/database/postgres/row-level-security) — HIGH confidence (official)
- [Supabase Security Retro 2025](https://supabase.com/blog/supabase-security-2025-retro) — HIGH confidence (official)
- [Supabase Security Flaw: 170+ Apps Exposed by Missing RLS](https://byteiota.com/supabase-security-flaw-170-apps-exposed-by-missing-rls/) — MEDIUM confidence (verified against official RLS docs)
- [Harden Your Supabase: Real-World Pentest Lessons](https://www.pentestly.io/blog/supabase-security-best-practices-2025-guide) — MEDIUM confidence
- [Tailwind CSS Play CDN Documentation](https://tailwindcss.com/docs/installation/play-cdn) — HIGH confidence (official)
- [Tailwind CDN Production Discussion](https://github.com/zauberzeug/nicegui/discussions/2517) — MEDIUM confidence
- [Supabase CORS for REST API — wildcard origin](https://github.com/orgs/supabase/discussions/7038) — MEDIUM confidence (GitHub discussion, corroborated by multiple sources)
- [Fix Supabase CORS Errors Guide 2025](https://corsproxy.io/blog/fix-supabase-cors-errors/) — MEDIUM confidence
- [Vercel Environment Variables Documentation](https://vercel.com/docs/environment-variables) — HIGH confidence (official)
- [JavaScript ES Modules — MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules) — HIGH confidence (official)
- [10 Common Supabase Security Misconfigurations](https://modernpentest.com/blog/supabase-security-misconfigurations) — MEDIUM confidence
