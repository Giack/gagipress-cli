# Architecture Patterns

**Domain:** Static vanilla JS dashboard — read-only, Supabase-backed, zero build step
**Researched:** 2026-02-25

## Recommended Architecture

A single HTML shell (`index.html`) loads Tailwind via CDN and a root ES module (`app.js`). All views are rendered into a single `#app` div by plain JS functions. Hash-based routing (`#/ideas`, `#/calendar`, `#/books`) triggers view swaps with no server involvement — correct for Vercel static hosting.

```
Browser
  └── index.html (shell, Tailwind CDN)
        └── app.js (root module, router init)
              ├── router.js          — hashchange → view dispatch
              ├── api/supabase.js    — raw fetch wrapper (headers, URL builder)
              ├── views/ideas.js     — kanban: content_ideas
              ├── views/calendar.js  — kanban: content_calendar
              ├── views/books.js     — table: books
              ├── components/
              │     ├── kanban.js    — reusable kanban column renderer
              │     ├── table.js     — reusable table renderer
              │     └── loader.js    — spinner / skeleton state
              └── config.js          — Supabase URL + anon key constants
```

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| `router.js` | Listens to `hashchange`, maps hash → view function, swaps `#app` content | All view modules |
| `api/supabase.js` | Builds PostgREST URLs, attaches `apikey` + `Authorization` headers, returns parsed JSON | Called by views |
| `views/ideas.js` | Fetches `content_ideas`, groups by status, delegates render to `kanban.js` | `api/supabase.js`, `components/kanban.js` |
| `views/calendar.js` | Fetches `content_calendar`, groups by status, delegates render to `kanban.js` | `api/supabase.js`, `components/kanban.js` |
| `views/books.js` | Fetches `books`, renders via `table.js` | `api/supabase.js`, `components/table.js` |
| `components/kanban.js` | Renders columns + cards from data array, returns HTML string or DOM node | Called by views, reads no external state |
| `components/table.js` | Renders header + rows from data array, returns HTML string or DOM node | Called by views |
| `components/loader.js` | Returns loading skeleton markup | Called by views before fetch resolves |
| `config.js` | Exports `SUPABASE_URL` and `SUPABASE_ANON_KEY` constants | Imported by `api/supabase.js` |

### Data Flow

```
URL hash change
  → router.js dispatches to view function
    → view calls loader.js, injects skeleton into #app
    → view calls api/supabase.js (fetch)
      → PostgREST returns JSON array
    → view groups/transforms data
    → view calls component renderer (kanban.js or table.js)
      → component returns HTML
    → view injects rendered HTML into #app
```

Data flows in one direction: Supabase → fetch wrapper → view → component → DOM. No two-way binding, no reactive state library needed. Views own their data; components are pure renderers (data in, HTML out).

---

## Patterns to Follow

### Pattern 1: Hash Router

**What:** A single `hashchange` event listener maps `location.hash` to a view function.

**When:** Always — required for multi-view navigation on a Vercel static site with no server-side routing.

**Example:**
```javascript
// router.js
import { renderIdeas }   from './views/ideas.js';
import { renderCalendar } from './views/calendar.js';
import { renderBooks }   from './views/books.js';

const routes = {
  '#/ideas':    renderIdeas,
  '#/calendar': renderCalendar,
  '#/books':    renderBooks,
};

export function initRouter() {
  const navigate = () => {
    const view = routes[location.hash] ?? renderIdeas;
    view(document.getElementById('app'));
  };
  window.addEventListener('hashchange', navigate);
  navigate(); // initial load
}
```

### Pattern 2: Fetch Wrapper (raw PostgREST)

**What:** A thin wrapper around `fetch` that constructs PostgREST URLs and injects required headers. No Supabase SDK — avoids CDN import complexity.

**When:** All data fetching. Never call `fetch` directly from view code.

**Example:**
```javascript
// api/supabase.js
import { SUPABASE_URL, SUPABASE_ANON_KEY } from '../config.js';

export async function query(table, params = {}) {
  const url = new URL(`${SUPABASE_URL}/rest/v1/${table}`);
  Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v));

  const res = await fetch(url.toString(), {
    headers: {
      'apikey': SUPABASE_ANON_KEY,
      'Authorization': `Bearer ${SUPABASE_ANON_KEY}`,
    },
  });
  if (!res.ok) throw new Error(`Supabase error: ${res.status}`);
  return res.json();
}

// Usage: query('content_ideas', { select: '*', order: 'created_at.desc' })
```

PostgREST filtering uses query string params: `?status=eq.pending`, `?select=id,title,status`.

### Pattern 3: Pure Component Renderers

**What:** Component functions receive data and return an HTML string (or DocumentFragment). They have no side effects, fetch nothing, and hold no state.

**When:** All reusable UI pieces (kanban, table).

**Example:**
```javascript
// components/kanban.js
export function renderKanban(columns) {
  // columns: [{ label: 'Pending', status: 'pending', items: [...] }]
  return `
    <div class="flex gap-4 overflow-x-auto p-4">
      ${columns.map(col => `
        <div class="flex-shrink-0 w-72 bg-gray-800 rounded-lg p-3">
          <h2 class="text-sm font-semibold text-gray-400 uppercase mb-3">
            ${col.label} <span class="text-gray-500">(${col.items.length})</span>
          </h2>
          <div class="space-y-2">
            ${col.items.map(item => `
              <div class="bg-gray-700 rounded p-3 text-sm text-white">
                <p class="font-medium">${item.title}</p>
              </div>
            `).join('')}
          </div>
        </div>
      `).join('')}
    </div>
  `;
}
```

### Pattern 4: View Owns Fetch + Group

**What:** Each view fetches its own data, groups it into the shape the component expects, shows a loader during fetch, handles errors.

**When:** Every view function.

**Example:**
```javascript
// views/ideas.js
import { query } from '../api/supabase.js';
import { renderKanban } from '../components/kanban.js';

const COLUMNS = [
  { label: 'Pending',  status: 'pending' },
  { label: 'Approved', status: 'approved' },
  { label: 'Rejected', status: 'rejected' },
  { label: 'Scripted', status: 'scripted' },
];

export async function renderIdeas(container) {
  container.innerHTML = '<p class="text-gray-400 p-4">Loading...</p>';
  try {
    const items = await query('content_ideas', {
      select: 'id,title,platform,status,created_at',
      order: 'created_at.desc',
    });
    const columns = COLUMNS.map(col => ({
      ...col,
      items: items.filter(i => i.status === col.status),
    }));
    container.innerHTML = renderKanban(columns);
  } catch (err) {
    container.innerHTML = `<p class="text-red-400 p-4">Error: ${err.message}</p>`;
  }
}
```

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Global Mutable State Object

**What:** A shared `window.state = {}` or module-level object mutated across files.

**Why bad:** Creates implicit coupling between views. Since this is a read-only dashboard that fetches fresh on every navigation, shared state adds complexity with no benefit.

**Instead:** Each view fetches its own data on mount. No caching needed for a personal tool at this scale.

### Anti-Pattern 2: Mixing Fetch Logic into Components

**What:** Kanban or table components that call `fetch` themselves.

**Why bad:** Makes components non-reusable and untestable. Kills the one-direction data flow.

**Instead:** Views fetch and transform; components only render.

### Anti-Pattern 3: History API Routing on a Static Host

**What:** Using `pushState` and `popstate` for clean URLs (`/ideas`, `/calendar`).

**Why bad:** Vercel static hosting returns 404 for direct URL access to `/calendar` — the server has no HTML file at that path. Requires a `vercel.json` rewrite rule, adding complexity.

**Instead:** Use hash routing (`#/ideas`). It is purely client-side, works on any static host, no server config needed. SEO is irrelevant for a personal internal tool.

### Anti-Pattern 4: Supabase JS SDK via CDN for Simple Reads

**What:** Loading `@supabase/supabase-js` from a CDN script tag just to do `SELECT *`.

**Why bad:** Adds ~50KB of overhead for functionality replaceable by 15 lines of `fetch`. The SDK also targets Node environments; CDN builds add extra complexity.

**Instead:** Use the raw `fetch` wrapper pattern (Pattern 2). The PostgREST HTTP API is simple: `GET /rest/v1/tablename?select=*` with two headers.

---

## File / Folder Structure

```
dashboard/                  ← Vercel root (or a /dashboard subfolder in existing repo)
├── index.html              ← Shell: <div id="app">, Tailwind CDN, <script type="module" src="app.js">
├── app.js                  ← Entry: imports router, calls initRouter()
├── config.js               ← SUPABASE_URL + SUPABASE_ANON_KEY constants
├── router.js               ← Hash router: maps #/route → view function
├── api/
│   └── supabase.js         ← fetch wrapper: query(table, params)
├── views/
│   ├── ideas.js            ← content_ideas kanban view
│   ├── calendar.js         ← content_calendar kanban view
│   └── books.js            ← books table view
└── components/
    ├── kanban.js           ← pure kanban renderer (columns + cards)
    ├── table.js            ← pure table renderer (header + rows)
    └── loader.js           ← loading skeleton HTML
```

No `node_modules`. No `package.json`. No build output directory. Every file is shipped as-is.

---

## Suggested Build Order

Build in this order — each step is independently testable:

1. **`config.js` + `api/supabase.js`** — Verify Supabase connection returns data. Test in browser console first.
2. **`index.html` shell** — Static HTML with Tailwind CDN, `<div id="app">`, nav links with `href="#/ideas"` etc.
3. **`router.js`** — Hash routing wired up, each route logs to console. Navigation works before any real views exist.
4. **`components/kanban.js` + `components/table.js`** — Build with hardcoded mock data. Get the layout right before touching the API.
5. **`views/books.js`** — Simplest view (table, no grouping logic). Validates the full fetch → render pipeline end to end.
6. **`views/ideas.js`** — Kanban with 4 columns. Validates grouping logic.
7. **`views/calendar.js`** — Kanban with 5 columns (scheduled/approved/publishing/published/failed). Reuses the same pattern.
8. **Polish** — Loading states, error states, empty column states, nav active state highlighting.

---

## Deployment to Vercel

Vercel treats any directory with an `index.html` as a static site. No `vercel.json` needed with hash routing.

- Set `SUPABASE_ANON_KEY` and `SUPABASE_URL` as Vercel environment variables, or inline them directly in `config.js` (acceptable since anon key is public and RLS is the security boundary).
- For the simplest deploy: inline the values in `config.js` and commit. The anon key is designed to be public — RLS policies enforce access control server-side.

---

## Scalability Considerations

| Concern | At current scale (personal) | If it grows |
|---------|------------------------------|-------------|
| Data volume | Full-table fetch is fine (< 1000 rows) | Add PostgREST pagination (`limit`, `offset` params) |
| View complexity | String concatenation templates are sufficient | Migrate to tagged template literals or Web Components |
| Multiple dashboards | Single `index.html` is fine | Split into multiple HTML files, keep shared `api/` and `components/` |
| Auth | None needed (personal, obscure URL) | Add Supabase Auth — swap anon key flow for session token |

---

## Sources

- Project context: `/Users/gsortino/Workspace/sales-mkgt-automation/.planning/PROJECT.md`
- Existing CLI HTTP pattern: `internal/repository/*` (documented in CLAUDE.md — same headers apply in browser)
- [Hash Routing — MDN Glossary](https://developer.mozilla.org/en-US/docs/Glossary/Hash_routing)
- [Build a SPA Router in Vanilla JavaScript](https://jsdev.space/spa-vanilla-js/) — MEDIUM confidence (WebSearch)
- [Routing in Vanilla JS: Hash vs History API](https://medium.com/@RyuotheGreate/routing-in-vanilla-javascript-hash-vs-history-api-a65382121871) — MEDIUM confidence (WebSearch)
- [Supabase PostgREST JS SDK npm](https://www.npmjs.com/package/@supabase/postgrest-js) — header pattern inferred from SDK source + existing Go CLI patterns (HIGH confidence, cross-validated)
- [State Management in Vanilla JS: 2026 Trends](https://medium.com/@chirag.dave/state-management-in-vanilla-js-2026-trends-f9baed7599de) — LOW confidence (WebSearch only)
