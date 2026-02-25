# Phase 2: Data Layer + Shared Components - Research

**Researched:** 2026-02-25
**Domain:** Vanilla JS ES modules — fetch wrapper, active-nav highlighting, loading/error/empty states, dark Tailwind palette
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| NAV-01 | Navbar with 3 sections: Books / Ideas / Calendar | Navbar HTML exists in `index.html`; Phase 2 must wire active-state CSS toggling via JS in router dispatch |
| NAV-02 | Hash routing (`#/books`, `#/ideas`, `#/calendar`) | Router already implemented in `js/router.js`; Phase 2 must ensure dispatch updates nav link classes on every route change |
| NAV-03 | Active nav item highlighted | Must read `location.hash` on each dispatch and toggle a CSS active class (e.g., `text-white border-b-2 border-indigo-500`) on the matching `<a>` tag |
| UX-01 | Spinner during Supabase fetch | Render spinner HTML into `#app` before fetch starts; replace with content or error after promise settles |
| UX-02 | Visible error banner when fetch fails | Catch fetch errors and render a styled error `<div>` — never leave `#app` empty or unchanged on error |
| UX-03 | Empty-state message when column/table empty | Check array length after successful fetch; if empty, render a placeholder message (not blank HTML) |
| UX-04 | Dark color palette globally | `bg-gray-900` / `bg-gray-800` / `text-gray-100` / `text-gray-400` already in `index.html`; Phase 2 must ensure all dynamic HTML uses the same palette via consistent class constants |
</phase_requirements>

---

## Summary

Phase 2 builds the two foundational layers that all subsequent phases (3–5) will consume: a **fetch wrapper** (`js/api.js`) that encapsulates Supabase query calls with loading/error/empty handling, and a **nav highlight function** that wires active-tab visual feedback to the existing hash router.

The critical insight is that both layers are already partially scaffolded. `index.html` has the static navbar HTML and dark palette. `router.js` has the hash dispatch logic. `config.example.js` shows the supabase-js ESM import pattern. Phase 2 fills the gaps: `api.js` with a `fetchTable(table)` wrapper, a `renderLoading()` / `renderError()` / `renderEmpty()` shared UI utilities module, and an `updateNav(hash)` function called from the router on every dispatch.

No new libraries are needed. The entire phase is pure vanilla JS using the supabase-js client already established in Phase 1.

**Primary recommendation:** Build `js/api.js` (fetch wrapper) and `js/components.js` (shared UI renderers) first as pure utility modules with no DOM side effects. Then wire `updateNav()` into the router. This order lets each piece be tested independently before any view uses it.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| @supabase/supabase-js | v2 (2.97.0 via CDN ESM) | Data fetching | Already established in Phase 1 `config.example.js`; provides `.from(table).select('*')` query builder and structured error object |
| Tailwind CSS Play CDN | v4 (`@tailwindcss/browser@4`) | Styling | Already in `index.html`; all dark palette classes established; no additional setup required |
| Vanilla JS ES Modules | ES2022+ | Module system | Zero-build pattern locked in; all modules use `.js` extensions |

### New Files for Phase 2
| File | Purpose |
|------|---------|
| `dashboard/js/api.js` | Fetch wrapper — exports `fetchTable(table)` returning `{ data, error }` |
| `dashboard/js/components.js` | Pure renderers — exports `renderLoading()`, `renderError(msg)`, `renderEmpty(msg)` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `fetchTable()` wrapper in `api.js` | Inline fetch in each view | Inline fetch duplicates error/loading handling in 3+ views; centralized wrapper is mandatory for DRY |
| `components.js` module for UI states | Inline HTML strings in each view | Inline HTML creates inconsistent dark palette across views; single module enforces visual consistency |

---

## Architecture Patterns

### Recommended Project Structure After Phase 2
```
dashboard/
├── index.html           # Existing — navbar + #app + CDN scripts
├── config.js            # Existing (gitignored) — Supabase URL + key
├── config.example.js    # Existing — setup template
├── js/
│   ├── app.js           # Existing — entry point
│   ├── router.js        # MODIFIED — add updateNav() call in dispatch()
│   ├── api.js           # NEW — fetchTable() wrapper
│   ├── components.js    # NEW — renderLoading(), renderError(), renderEmpty()
│   └── views/
│       ├── books.js     # Existing stub — unchanged in Phase 2
│       ├── ideas.js     # Existing stub — unchanged in Phase 2
│       └── calendar.js  # Existing stub — unchanged in Phase 2
```

### Pattern 1: Fetch Wrapper (`api.js`)
**What:** Single async function wraps supabase-js `.from().select()` and always returns a normalized `{ data, error }` object.
**When to use:** Called at the start of every view render function in phases 3–5.
**Example:**
```javascript
// dashboard/js/api.js
// NOTE: All imports must use .js extensions
import { supabase } from '../config.js';

/**
 * Fetch all rows from a Supabase table.
 * Always resolves (never throws) — errors returned as { data: null, error }.
 * @param {string} table — table name matching PostgREST endpoint
 * @returns {Promise<{ data: Array|null, error: string|null }>}
 */
export async function fetchTable(table) {
  const { data, error } = await supabase.from(table).select('*');
  if (error) {
    return { data: null, error: error.message ?? 'Unknown error' };
  }
  return { data: data ?? [], error: null };
}
```

**Why this signature matters:**
- Always resolves — callers never need `try/catch`
- `error` is a string message, not a raw Supabase error object — simpler for callers
- Returns `data: []` (empty array) on success, never `null` — callers distinguish "empty" from "error" without extra checks

### Pattern 2: Shared UI State Renderers (`components.js`)
**What:** Pure functions that return HTML strings for loading, error, and empty states.
**When to use:** Called by every view before/after fetch to maintain consistent visual feedback.
**Example:**
```javascript
// dashboard/js/components.js

/** Spinner shown while data is loading */
export function renderLoading() {
  return `
    <div class="flex items-center gap-3 text-gray-400 py-12">
      <svg class="animate-spin h-5 w-5 text-indigo-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
      </svg>
      <span>Loading...</span>
    </div>
  `;
}

/** Error banner shown when fetch fails */
export function renderError(message) {
  return `
    <div class="bg-red-900/40 border border-red-700 text-red-300 rounded-lg px-4 py-3">
      <strong class="font-semibold">Error:</strong> ${escapeHtml(message)}
    </div>
  `;
}

/** Empty state shown when fetch returns no rows */
export function renderEmpty(message = 'No items found.') {
  return `
    <div class="text-gray-500 italic py-8 text-center">${escapeHtml(message)}</div>
  `;
}

/** Escape user-supplied strings before injecting into innerHTML */
function escapeHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}
```

**Why `escapeHtml` matters:** Error messages from Supabase may contain characters that break HTML injection. This is not a security concern (no untrusted users), but prevents broken UI if error messages contain `<` or `&`.

### Pattern 3: Active Nav Highlight
**What:** On every route dispatch, toggle a CSS active class on the matching `<a>` in the navbar.
**When to use:** Called inside `dispatch()` in `router.js` after determining the current hash.
**Example:**
```javascript
// In dashboard/js/router.js — add updateNav() and call it in dispatch()

/** Highlight the active nav link matching the current hash */
function updateNav(hash) {
  document.querySelectorAll('nav a').forEach(link => {
    if (link.getAttribute('href') === hash) {
      link.classList.add('text-white', 'border-b-2', 'border-indigo-500');
      link.classList.remove('text-gray-400');
    } else {
      link.classList.remove('text-white', 'border-b-2', 'border-indigo-500');
      link.classList.add('text-gray-400');
    }
  });
}

export function dispatch() {
  const hash = location.hash || '#/ideas';
  const view = routes[hash] ?? renderIdeas;
  updateNav(hash);  // ADD THIS LINE
  view();
}
```

**Key detail:** `updateNav` must run EVERY dispatch including on `load` — the initial page load must also set the active state or the first tab will appear unselected.

### Pattern 4: View Render Pattern (for Views to Follow in Phases 3–5)
**What:** The standard pattern every view (`books.js`, `ideas.js`, `calendar.js`) will follow when consuming the data layer.
**When to use:** This is the contract Phase 2 establishes; phases 3–5 implement it.
**Example:**
```javascript
// Template every view follows — Phase 2 documents this pattern
import { fetchTable } from '../api.js';
import { renderLoading, renderError, renderEmpty } from '../components.js';

export async function renderBooks() {
  const app = document.getElementById('app');

  // 1. Show spinner immediately
  app.innerHTML = renderLoading();

  // 2. Fetch data
  const { data, error } = await fetchTable('books');

  // 3. Handle error
  if (error) {
    app.innerHTML = renderError(error);
    return;
  }

  // 4. Handle empty
  if (data.length === 0) {
    app.innerHTML = renderEmpty('No books found. Add a book with the CLI.');
    return;
  }

  // 5. Render content
  app.innerHTML = `...`;
}
```

**This pattern must be established in Phase 2 documentation** even though phases 3–5 implement it. The planner should include this as a "view contract" that downstream tasks must follow.

### Anti-Patterns to Avoid
- **Render function does not show spinner first:** If fetch is slow (Supabase cold start), `#app` stays empty. Always set `app.innerHTML = renderLoading()` as the very first line.
- **Error swallowed silently:** If `if (error) return;` is used without setting `app.innerHTML`, the spinner remains forever on error. Always call `renderError()` before returning on error path.
- **`router.js` imports `api.js`:** Router should NOT know about data fetching. Only views import `api.js`. Router only calls view render functions and `updateNav()`.
- **`updateNav()` not called on initial load:** The `load` event fires before the user clicks any tab. Without calling `updateNav()` in that handler, the initial active tab is not highlighted.
- **Tailwind classes defined only in JS strings never JIT-compiled:** Tailwind Play CDN scans the DOM — classes that only exist inside JS template literals at parse time may not be compiled. Test all dynamic classes by navigating to each state (loading, error, empty) at least once in the browser.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Supabase auth headers | Manual `fetch` with `Authorization: Bearer` on every call | `supabase.from(table).select('*')` | supabase-js manages auth headers, API URL construction, and error parsing — already in `config.js` |
| Loading state management | State variable + event system | `app.innerHTML = renderLoading()` before fetch | Zero state management needed; innerHTML replacement is the state machine for a single-div SPA |
| CSS animation for spinner | Custom keyframe CSS | Tailwind `animate-spin` utility | `animate-spin` is in Tailwind core; one class, no custom CSS |
| XSS escaping | Custom sanitizer | `escapeHtml()` in `components.js` (4 replaces) | Error messages from Supabase are trusted but may contain HTML-breaking chars; simple 4-line function is sufficient |

---

## Common Pitfalls

### Pitfall 1: Tailwind Dynamic Classes Not Compiled
**What goes wrong:** A Tailwind class like `border-indigo-500` used only in a JS string (never in static HTML) may not appear in the JIT-compiled stylesheet when using the Play CDN. The element renders without the expected style.
**Why it happens:** Tailwind Play CDN uses a MutationObserver to JIT-compile classes as they appear in the DOM. Classes that appear only in JS template literals BEFORE they are injected into the DOM may not be scanned in time.
**How to avoid:** After injecting HTML via `innerHTML`, verify styles apply in the browser. If a class is missing, add it to a hidden element in `index.html` to force it into the compile scan, or use inline styles as fallback for critical indicators (active nav underline).
**Warning signs:** DevTools shows the element has the class attribute but no matching CSS rule in the `<style>` tag injected by Tailwind CDN.

### Pitfall 2: Spinner Persists on Supabase Auth Failure
**What goes wrong:** If `config.js` has a wrong or missing `SUPABASE_ANON_KEY`, supabase-js returns a 401 error object. If the error path is not handled, the spinner stays forever.
**Why it happens:** `supabase.from().select()` with invalid credentials does not throw — it resolves with `{ data: null, error: { message: 'JWT expired' } }`. The caller must check `error`.
**How to avoid:** The `fetchTable()` wrapper in `api.js` always returns `{ data, error }`. Views MUST check `if (error)` before accessing `data`. The `renderError()` call must be unconditional on the error path.
**Warning signs:** After navigating to a view, spinner shows and never disappears.

### Pitfall 3: `innerHTML` Assignment Races on Slow Networks
**What goes wrong:** If a user clicks Books (starts fetch), then immediately clicks Ideas (starts another fetch), the second view may render first (fast response) then be overwritten by the first view's response (slow response).
**Why it happens:** Each `async renderX()` call captures the `app` reference at call time. The `app.innerHTML = content` assignments are not ordered.
**How to avoid:** For this personal single-user tool, this is acceptable risk — do not add complex cancellation logic. Document it as a known limitation. If it becomes a problem, use an AbortController in `fetchTable()`.
**Warning signs:** Clicking tabs rapidly shows the wrong view content.

### Pitfall 4: `nav a` Selector Fails if Navbar HTML Changes
**What goes wrong:** `document.querySelectorAll('nav a')` is fragile. If Phase 5 adds a settings icon or changes the `<nav>` structure, `updateNav()` may highlight the wrong element or nothing.
**Why it happens:** CSS selector depends on DOM structure assumptions made in Phase 1.
**How to avoid:** Add `data-nav` attributes to nav links in `index.html`: `<a href="#/books" data-nav="#/books">`. Then select with `document.querySelectorAll('[data-nav]')` and match `link.dataset.nav === hash`. This is more robust than structural selectors.
**Warning signs:** Active tab highlight breaks after any HTML change to `<nav>`.

---

## Code Examples

### Complete `api.js`
```javascript
// dashboard/js/api.js
// NOTE: All imports must use .js extensions — no bundler
import { supabase } from '../config.js';

/**
 * Fetch all rows from a Supabase table.
 * Always resolves — errors returned as { data: null, error: string }.
 * @param {string} table
 * @returns {Promise<{ data: Array|null, error: string|null }>}
 */
export async function fetchTable(table) {
  const { data, error } = await supabase.from(table).select('*');
  if (error) {
    return { data: null, error: error.message ?? 'Unknown fetch error' };
  }
  return { data: data ?? [], error: null };
}
```

### Complete `components.js`
```javascript
// dashboard/js/components.js

export function renderLoading() {
  return `
    <div class="flex items-center gap-3 text-gray-400 py-12">
      <svg class="animate-spin h-5 w-5 text-indigo-400"
           xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10"
                stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
      </svg>
      <span>Loading...</span>
    </div>`;
}

export function renderError(message) {
  return `
    <div class="bg-red-900/40 border border-red-700 text-red-300 rounded-lg px-4 py-3 mt-4">
      <strong class="font-semibold">Error:</strong> ${escapeHtml(message)}
    </div>`;
}

export function renderEmpty(message = 'No items found.') {
  return `
    <div class="text-gray-500 italic py-8 text-center">${escapeHtml(message)}</div>`;
}

function escapeHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}
```

### Updated `router.js` with `updateNav()`
```javascript
// dashboard/js/router.js
import { renderBooks }   from './views/books.js';
import { renderIdeas }   from './views/ideas.js';
import { renderCalendar } from './views/calendar.js';

const routes = {
  '#/books':    renderBooks,
  '#/ideas':    renderIdeas,
  '#/calendar': renderCalendar,
};

function updateNav(hash) {
  document.querySelectorAll('[data-nav]').forEach(link => {
    const isActive = link.dataset.nav === hash;
    link.classList.toggle('text-white', isActive);
    link.classList.toggle('border-b-2', isActive);
    link.classList.toggle('border-indigo-500', isActive);
    link.classList.toggle('text-gray-400', !isActive);
  });
}

export function dispatch() {
  const hash = location.hash || '#/ideas';
  const view = routes[hash] ?? renderIdeas;
  updateNav(hash);
  view();
}

window.addEventListener('hashchange', dispatch);
window.addEventListener('load', dispatch);
```

### Required `index.html` nav update (add `data-nav` attributes)
```html
<nav class="flex gap-6 px-6 py-4 border-b border-gray-700 bg-gray-800">
  <a href="#/books"    data-nav="#/books"    class="text-gray-400 hover:text-white font-medium transition-colors">Books</a>
  <a href="#/ideas"    data-nav="#/ideas"    class="text-gray-400 hover:text-white font-medium transition-colors">Ideas</a>
  <a href="#/calendar" data-nav="#/calendar" class="text-gray-400 hover:text-white font-medium transition-colors">Calendar</a>
</nav>
```

### Smoke test — verify fetch + error path in browser DevTools
```javascript
// Paste in DevTools console to test api.js in isolation
// (requires config.js to be present)
import('/js/api.js').then(m => m.fetchTable('books')).then(console.log);
// Expected: { data: [...], error: null }  — or { data: null, error: 'message' }
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| jQuery `$.ajax` + DOM manipulation | Vanilla JS `async/await` + `innerHTML` | 2017–2020 (broad ES2017 support) | No jQuery dependency; `async/await` is readable and standard |
| Redux / Flux for loading state | Direct `innerHTML` replacement | N/A for zero-build static | For a single-div SPA with no client-side state, state machine via innerHTML is sufficient and has zero overhead |
| Spinner libraries (spin.js) | Tailwind `animate-spin` + SVG | Tailwind v1+ | One utility class replaces a 10KB library |

---

## Open Questions

1. **`fetchTable()` ordering — does Supabase PostgREST return rows in insertion order?**
   - What we know: PostgREST with no `order` clause returns rows in undefined order (PostgreSQL heap order)
   - What's unclear: Whether phases 3–5 will need ordered results (e.g., ideas sorted by created_at)
   - Recommendation: Add an optional `options` parameter to `fetchTable()` now: `fetchTable(table, { order: 'created_at' })` — easy to add, prevents rework in phases 3–5

2. **Which Supabase table to use for Phase 2 live-fetch smoke test**
   - What we know: `books` table has the simplest schema and is most likely to have at least 1 row (the ASIN B0GJ54MR4F is set per MEMORY.md)
   - Recommendation: Wire the books view for the Phase 2 smoke test even though books is Phase 3 work — use a read-only `.select('title')` to confirm the data layer works end-to-end

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Manual browser testing — no JS test framework (zero-build constraint, no npm) |
| Config file | None |
| Quick run command | Open `dashboard/index.html` in browser with DevTools open |
| Full suite command | Navigate to each tab; check Network tab, Console, and visual states |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| NAV-01 | Navbar renders with Books/Ideas/Calendar links | manual | Open dashboard; visually confirm all 3 links present | ✅ (index.html exists) |
| NAV-02 | Hash routing navigates without page reload | manual | Click each tab; Network tab shows no document reload | ✅ (router.js exists) |
| NAV-03 | Active nav link is highlighted | manual | Click each tab; confirm active link has white text + indigo underline | ❌ Wave 0 — updateNav() not yet implemented |
| UX-01 | Spinner appears before data arrives | manual | Throttle network in DevTools to Slow 3G; confirm spinner visible during fetch | ❌ Wave 0 — api.js + renderLoading() not yet implemented |
| UX-02 | Error banner on fetch failure | manual | Set wrong API key in config.js; confirm error banner renders (not empty/spinner) | ❌ Wave 0 — renderError() not yet implemented |
| UX-03 | Empty-state message when no rows | manual | Test with empty table or filter to 0 results; confirm message renders | ❌ Wave 0 — renderEmpty() not yet implemented |
| UX-04 | Dark palette consistent on all dynamic content | manual | Navigate all views and states; confirm no white backgrounds or default text colors | ❌ Wave 0 — components.js dark classes not yet implemented |

### Sampling Rate
- **Per task commit:** Open dashboard in browser, navigate all 3 tabs — confirm no JS errors in Console
- **Per wave merge:** Full manual checklist: all 5 success criteria from phase definition
- **Phase gate:** Full checklist green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `dashboard/js/api.js` — fetch wrapper; covers UX-01, UX-02, UX-03
- [ ] `dashboard/js/components.js` — shared UI renderers; covers UX-01, UX-02, UX-03, UX-04
- [ ] `dashboard/js/router.js` updated with `updateNav()` — covers NAV-03
- [ ] `dashboard/index.html` updated with `data-nav` attributes on nav links — required by `updateNav()` selector

---

## Sources

### Primary (HIGH confidence)
- `dashboard/js/router.js` — existing dispatch pattern; directly read
- `dashboard/js/app.js` — existing entry point; directly read
- `dashboard/index.html` — existing HTML structure; directly read
- `dashboard/config.example.js` — confirms supabase-js ESM import pattern; directly read
- `.planning/phases/01-foundation/01-RESEARCH.md` — Phase 1 architecture decisions and patterns
- `.planning/REQUIREMENTS.md` — Phase 2 requirement IDs NAV-01..03, UX-01..04
- [MDN — ES Modules](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Modules) — `.js` extension requirement
- [Tailwind CSS — animate-spin](https://tailwindcss.com/docs/animation) — spinner utility class

### Secondary (MEDIUM confidence)
- [Supabase JS Client — Error Handling](https://supabase.com/docs/reference/javascript/select) — `{ data, error }` return shape from `.select()`
- [Tailwind CSS Play CDN — Dynamic Classes](https://tailwindcss.com/docs/content-configuration#dynamic-class-names) — JIT compilation with MutationObserver behavior

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new libraries; uses Phase 1 established stack directly
- Architecture: HIGH — patterns derived directly from existing Phase 1 files, not speculative
- Pitfalls: HIGH (Tailwind dynamic classes), MEDIUM (innerHTML race) — Tailwind CDN behavior verified against docs
- Validation approach: HIGH — manual testing is correct for zero-build static dashboard; no JS test framework to install

**Research date:** 2026-02-25
**Valid until:** 2026-03-25 (stable domain; Tailwind CDN and supabase-js v2 API not changing)
