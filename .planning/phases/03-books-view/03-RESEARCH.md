# Phase 3: Books View - Research

**Researched:** 2026-02-25
**Domain:** Vanilla JS ES modules — live Supabase table fetch, HTML table rendering, sticky header, skeleton loading, ASIN links
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Column order: Title → ASIN → Genre → Target Audience
- Long titles: truncate with ellipsis, reveal full title on hover tooltip
- Sticky header — stays visible when the list scrolls
- Loading: skeleton rows that match the table column structure (not a plain spinner)
- Empty state: replace the entire table with the `renderEmpty()` component — no orphaned headers
- Empty copy: "No books in your catalog yet"
- Error: red error banner using `renderError()` from Phase 2 — no retry button needed in Phase 3
- ASIN cell is a styled `<a>` tag pointing to `https://www.amazon.com/dp/{ASIN}`, `target="_blank"`
- ASIN link: indigo color + underline on hover, matches Phase 2 palette
- Row hover: subtle background highlight (signals interactivity)
- No other row-level actions — read-only table in Phase 3

### Claude's Discretion
- Column sorting behavior (user deferred)
- Exact skeleton row count and animation style
- Table border vs borderless styling
- Row hover color value (within the dark palette)

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| BOOKS-01 | Books table with columns: title, ASIN, genre, target audience — populated from live Supabase data | `fetchTable('books')` from `api.js` returns rows; map `kdp_asin` field (not `asin`) to ASIN column; render with `<table>` HTML string |
| BOOKS-02 | Click on ASIN opens Amazon product page in a new browser tab | `<a href="https://www.amazon.com/dp/{row.kdp_asin}" target="_blank" rel="noopener">` inside ASIN cell |
</phase_requirements>

---

## Summary

Phase 3 replaces the `renderBooks()` stub in `dashboard/js/views/books.js` with a full fetch-to-render pipeline. The entire implementation lives in one file — no new dependencies, no new infrastructure. All building blocks are already in place from Phases 1–2: `fetchTable()` from `api.js`, `renderLoading()` / `renderError()` / `renderEmpty()` from `components.js`, and the `#app` render target.

The most important implementation detail is the **column name**: the Supabase `books` table uses `kdp_asin` (not `asin`). This is defined in migration `001_initial_schema.sql` and confirmed in `internal/models/book.go`. The ASIN link URL must use `row.kdp_asin`, not `row.asin`.

The loading state uses skeleton rows (not the `renderLoading()` spinner) per the locked decision. This means `renderBooks()` must produce its own skeleton HTML — 3–5 animated placeholder rows matching the 4-column table structure — before the fetch resolves.

**Primary recommendation:** Replace `dashboard/js/views/books.js` completely. Follow the async view pattern: set `#app` to skeleton immediately → `await fetchTable('books')` → branch on error/empty/data → set `#app` to final HTML.

---

## Standard Stack

### Core (all established in Phase 1–2, no new installs)
| File | Purpose | Already Exists |
|------|---------|----------------|
| `dashboard/js/api.js` | `fetchTable('books')` — returns `{ data, error }` | Yes |
| `dashboard/js/components.js` | `renderError(msg)`, `renderEmpty(msg)` | Yes |
| `dashboard/js/views/books.js` | Books view — currently a stub, Phase 3 replaces it | Yes (stub) |

### Supabase books table columns
| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Not displayed |
| `title` | TEXT | Column 1 — truncate long values |
| `kdp_asin` | TEXT | Column 2 — ASIN link target; field is `kdp_asin` NOT `asin` |
| `genre` | TEXT | Column 3 |
| `target_audience` | TEXT | Column 4 |

**No npm install needed** — zero-build project, all tooling already loaded via CDN.

---

## Architecture Patterns

### Recommended File Structure (no changes to filesystem)
```
dashboard/js/
├── api.js           # unchanged — fetchTable() already handles Supabase
├── components.js    # unchanged — renderError(), renderEmpty() ready to use
├── router.js        # unchanged
├── app.js           # unchanged
└── views/
    ├── books.js     # REPLACE stub — full async view implementation
    ├── ideas.js     # unchanged (stub)
    └── calendar.js  # unchanged (stub)
```

### Pattern: Async View with Skeleton Loading

This is the complete pattern for `books.js`:

```javascript
// dashboard/js/views/books.js
import { fetchTable } from '../api.js';
import { renderError, renderEmpty } from '../components.js';

export async function renderBooks() {
  const app = document.getElementById('app');

  // 1. Show skeleton immediately — layout stable before data arrives
  app.innerHTML = renderBooksSkeleton();

  // 2. Fetch
  const { data, error } = await fetchTable('books', { order: 'title' });

  // 3. Branch on result
  if (error) {
    app.innerHTML = renderError(error);
    return;
  }
  if (data.length === 0) {
    app.innerHTML = renderEmpty('No books in your catalog yet');
    return;
  }

  // 4. Render table
  app.innerHTML = renderBooksTable(data);
}
```

**Key constraint from router.js:** `dispatch()` calls `view()` synchronously — `renderBooks` is called without `await`. The view must set `#app` synchronously (skeleton) before any async work, otherwise the DOM is empty during the fetch. Setting the skeleton in the first synchronous line satisfies this.

### Pattern: HTML Table with Sticky Header

```javascript
function renderBooksTable(books) {
  const rows = books.map(book => `
    <tr class="border-b border-gray-700 hover:bg-gray-800 transition-colors">
      <td class="py-3 px-4 max-w-xs truncate" title="${escapeHtml(book.title)}">${escapeHtml(book.title)}</td>
      <td class="py-3 px-4">
        <a href="https://www.amazon.com/dp/${escapeHtml(book.kdp_asin)}"
           target="_blank" rel="noopener"
           class="text-indigo-400 hover:underline">
          ${escapeHtml(book.kdp_asin)}
        </a>
      </td>
      <td class="py-3 px-4 text-gray-300">${escapeHtml(book.genre ?? '—')}</td>
      <td class="py-3 px-4 text-gray-300">${escapeHtml(book.target_audience ?? '—')}</td>
    </tr>`).join('');

  return `
    <div class="max-w-4xl overflow-x-auto">
      <h1 class="text-2xl font-bold text-white mb-6">Books</h1>
      <table class="w-full text-sm text-left">
        <thead class="sticky top-0 bg-gray-800 text-gray-400 uppercase text-xs">
          <tr>
            <th class="py-3 px-4">Title</th>
            <th class="py-3 px-4">ASIN</th>
            <th class="py-3 px-4">Genre</th>
            <th class="py-3 px-4">Target Audience</th>
          </tr>
        </thead>
        <tbody class="text-gray-100">${rows}</tbody>
      </table>
    </div>`;
}
```

### Pattern: Skeleton Rows (4-column match)

Skeleton rows must match the 4-column table structure exactly — prevents layout jump when real data arrives.

```javascript
function renderBooksSkeleton() {
  const skeletonRow = `
    <tr class="border-b border-gray-700">
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-40"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-24"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-20"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-32"></div></td>
    </tr>`;

  return `
    <div class="max-w-4xl overflow-x-auto">
      <h1 class="text-2xl font-bold text-white mb-6">Books</h1>
      <table class="w-full text-sm text-left">
        <thead class="sticky top-0 bg-gray-800 text-gray-400 uppercase text-xs">
          <tr>
            <th class="py-3 px-4">Title</th>
            <th class="py-3 px-4">ASIN</th>
            <th class="py-3 px-4">Genre</th>
            <th class="py-3 px-4">Target Audience</th>
          </tr>
        </thead>
        <tbody>${skeletonRow.repeat(4)}</tbody>
      </table>
    </div>`;
}
```

### Anti-Patterns to Avoid

- **Using `row.asin` instead of `row.kdp_asin`:** The Supabase column is `kdp_asin`. Using `asin` returns `undefined` — every ASIN cell renders blank and links break silently.
- **Awaiting view in router:** `dispatch()` calls views synchronously. If `renderBooks` has no synchronous skeleton render before the `await`, `#app` is blank during fetch. Always set skeleton before `await`.
- **Rendering empty `<tbody>` on empty data:** Show `renderEmpty()` in place of the table entirely — no orphaned header per locked decision.
- **Skipping `rel="noopener"` on external links:** `target="_blank"` without `rel="noopener"` allows the new tab to access `window.opener`. Always pair them.
- **Skipping `escapeHtml` on book data:** Book titles may contain `<`, `>`, `&` characters. Always escape before injecting into HTML strings.
- **Omitting `null` guard on optional columns:** `target_audience` is nullable in the schema. Use `book.target_audience ?? '—'` to avoid rendering "null".

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Fetch + error handling | Custom fetch | `fetchTable()` from `api.js` | Already handles Supabase error shape, null data, and structured return |
| Error state HTML | Custom error div | `renderError(msg)` from `components.js` | Consistent red banner styling, escapeHtml already applied |
| Empty state HTML | Custom empty div | `renderEmpty('No books in your catalog yet')` from `components.js` | Consistent gray italic styling |

**Key insight:** Phases 1–2 were specifically designed to prevent hand-rolled fetch/error/empty logic in view files. Phase 3 consumes those utilities — do not duplicate them.

---

## Common Pitfalls

### Pitfall 1: Wrong column name for ASIN
**What goes wrong:** `row.asin` is `undefined` — links render as `https://www.amazon.com/dp/undefined`, cells are blank.
**Why it happens:** The column in the `books` table is `kdp_asin`, not `asin`. Easy to assume wrong name.
**How to avoid:** Use `row.kdp_asin` everywhere. Verify against `migrations/001_initial_schema.sql`.
**Warning signs:** ASIN column blank in table, Amazon link goes to `/dp/undefined`.

### Pitfall 2: Missing synchronous skeleton render
**What goes wrong:** `#app` is blank for the duration of the fetch (~100–500ms). No loading feedback.
**Why it happens:** Developer awaits fetch before touching the DOM. `dispatch()` in router.js does not `await` the view function.
**How to avoid:** `app.innerHTML = renderBooksSkeleton()` must be the FIRST line in `renderBooks()`, before any `await`.
**Warning signs:** Brief blank screen when clicking Books tab.

### Pitfall 3: Tailwind `animate-pulse` not rendering
**What goes wrong:** Skeleton rows show gray boxes but don't animate.
**Why it happens:** Tailwind v4 CDN JIT only compiles classes it sees in the initial HTML parse. Dynamically injected classes may miss the initial scan.
**How to avoid:** This is handled — Tailwind v4 CDN browser mode scans dynamically inserted content. `animate-pulse` is a core utility and compiles reliably. Verified in Phase 2 with `animate-spin` on the spinner.
**Warning signs:** Static gray boxes, no pulse animation.

### Pitfall 4: Orphaned table header on empty data
**What goes wrong:** Empty `books` table renders a header row with no body rows — looks broken.
**Why it happens:** Rendering table structure first, then checking empty inside tbody.
**How to avoid:** Check `data.length === 0` before building any table HTML. Return `renderEmpty()` which completely replaces the table.
**Warning signs:** Visible column headers with nothing below them.

---

## Code Examples

### Complete `books.js` implementation pattern

```javascript
// dashboard/js/views/books.js
// NOTE: All imports must use .js extensions — no bundler resolution in browser
import { fetchTable }        from '../api.js';
import { renderError, renderEmpty } from '../components.js';

export async function renderBooks() {
  const app = document.getElementById('app');
  app.innerHTML = renderBooksSkeleton();  // synchronous — before any await

  const { data, error } = await fetchTable('books', { order: 'title' });

  if (error) {
    app.innerHTML = renderError(error);
    return;
  }
  if (data.length === 0) {
    app.innerHTML = renderEmpty('No books in your catalog yet');
    return;
  }
  app.innerHTML = renderBooksTable(data);
}

function renderBooksTable(books) {
  const rows = books.map(book => `
    <tr class="border-b border-gray-700 hover:bg-gray-800 transition-colors">
      <td class="py-3 px-4 max-w-xs truncate" title="${escapeHtml(book.title)}">${escapeHtml(book.title)}</td>
      <td class="py-3 px-4">
        <a href="https://www.amazon.com/dp/${escapeHtml(book.kdp_asin ?? '')}"
           target="_blank" rel="noopener"
           class="text-indigo-400 hover:underline">
          ${escapeHtml(book.kdp_asin ?? '—')}
        </a>
      </td>
      <td class="py-3 px-4 text-gray-300">${escapeHtml(book.genre ?? '—')}</td>
      <td class="py-3 px-4 text-gray-300">${escapeHtml(book.target_audience ?? '—')}</td>
    </tr>`).join('');

  return `
    <div class="max-w-4xl overflow-x-auto">
      <h1 class="text-2xl font-bold text-white mb-6">Books</h1>
      <table class="w-full text-sm text-left">
        <thead class="sticky top-0 bg-gray-800 text-gray-400 uppercase text-xs">
          <tr>
            <th class="py-3 px-4">Title</th>
            <th class="py-3 px-4">ASIN</th>
            <th class="py-3 px-4">Genre</th>
            <th class="py-3 px-4">Target Audience</th>
          </tr>
        </thead>
        <tbody class="text-gray-100">${rows}</tbody>
      </table>
    </div>`;
}

function renderBooksSkeleton() {
  const row = `
    <tr class="border-b border-gray-700">
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-40"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-24"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-20"></div></td>
      <td class="py-3 px-4"><div class="h-4 bg-gray-700 rounded animate-pulse w-32"></div></td>
    </tr>`;
  return `
    <div class="max-w-4xl overflow-x-auto">
      <h1 class="text-2xl font-bold text-white mb-6">Books</h1>
      <table class="w-full text-sm text-left">
        <thead class="sticky top-0 bg-gray-800 text-gray-400 uppercase text-xs">
          <tr>
            <th class="py-3 px-4">Title</th>
            <th class="py-3 px-4">ASIN</th>
            <th class="py-3 px-4">Genre</th>
            <th class="py-3 px-4">Target Audience</th>
          </tr>
        </thead>
        <tbody>${row.repeat(4)}</tbody>
      </table>
    </div>`;
}

function escapeHtml(str) {
  return String(str ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}
```

---

## State of the Art

| Old Approach | Current Approach | Impact |
|--------------|-----------------|--------|
| Stub view with placeholder text | Full async view with skeleton + live data | Phase 3 goal |
| `renderLoading()` spinner for all views | Table-specific skeleton rows for Books | Layout stable during load — no shift |

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Manual browser verification (no automated test framework for dashboard) |
| Config file | none |
| Quick run command | Open `dashboard/index.html` locally or Vercel preview URL |
| Full suite command | Same — manual checklist verification |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| BOOKS-01 | Books tab shows table with title/ASIN/genre/target_audience from live Supabase | manual | Open `#/books` — verify 4 columns populated | N/A |
| BOOKS-02 | Click ASIN opens Amazon product page in new tab | manual | Click ASIN link — verify new tab opens `amazon.com/dp/{ASIN}` | N/A |

**Note:** This dashboard is a static zero-build vanilla JS project. There is no test runner (no jest, vitest, or similar). All verification is manual browser testing. Wave 0 has no automated test gaps to fill — verification is done via browser checklist.

### Sampling Rate
- **Per task commit:** Open `index.html` in browser, navigate to `#/books`, verify table renders
- **Per wave merge:** Full manual checklist: skeleton visible, data loads, ASIN link opens Amazon, empty state shows if table empty, error banner shows if Supabase offline
- **Phase gate:** All 3 success criteria TRUE before `/gsd:verify-work`

### Wave 0 Gaps
None — no automated test infrastructure needed for this phase. All verification is manual browser testing.

---

## Open Questions

1. **Is `books` table populated in the live Supabase instance?**
   - What we know: The table exists (migration 001), the CLI `gagipress books` command manages it
   - What's unclear: Whether the user has any books added — if not, the empty state is what will be seen during verification
   - Recommendation: Plan verification to cover BOTH empty state and data-populated state. User should add a test book via CLI before verifying BOOKS-01.

2. **Column sorting — deferred to Claude's discretion**
   - What we know: User deferred this decision
   - What's unclear: Whether to implement client-side sort on column header click in Phase 3
   - Recommendation: Skip sorting in Phase 3. The `fetchTable('books', { order: 'title' })` call already returns rows sorted by title alphabetically. This satisfies read-only use without adding JS complexity. Sorting can be added in v2.

---

## Sources

### Primary (HIGH confidence)
- `migrations/001_initial_schema.sql` — books table DDL, confirmed `kdp_asin` column name
- `internal/models/book.go` — Go struct confirms `json:"kdp_asin"` field name
- `dashboard/js/api.js` — `fetchTable()` signature and return shape
- `dashboard/js/components.js` — `renderError()`, `renderEmpty()` signatures
- `dashboard/js/router.js` — `dispatch()` synchronous call pattern
- `dashboard/js/views/books.js` — current stub to be replaced
- `dashboard/index.html` — Tailwind v4 CDN load, `#app` target, dark palette classes

### Secondary (MEDIUM confidence)
- `.planning/phases/02-data-layer-shared-components/02-RESEARCH.md` — confirmed Phase 2 patterns and component API

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all files exist and were read directly
- Architecture: HIGH — based on reading actual source code, not documentation
- Pitfalls: HIGH — `kdp_asin` verified in migration SQL and Go model; skeleton-before-await verified from router.js

**Research date:** 2026-02-25
**Valid until:** 2026-03-25 (stable — no external dependencies added in this phase)
