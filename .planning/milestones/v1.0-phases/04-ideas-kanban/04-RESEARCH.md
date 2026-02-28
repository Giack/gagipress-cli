# Phase 4: Ideas Kanban — Research

**Researched:** 2026-02-25
**Domain:** Vanilla JS kanban view — Supabase PostgREST fetch, multi-column layout, inline expand/collapse
**Confidence:** HIGH

---

## Summary

Phase 4 replaces the stub in `dashboard/js/views/ideas.js` with a four-column kanban board. The column order is fixed: pending / approved / rejected / scripted. Each card shows an idea title and its content type. Clicking a scripted card expands a preview of the linked script inline.

The established project pattern (books view) dictates the entire implementation approach: synchronous skeleton render before any await, `fetchTable` + `renderError` + `renderEmpty` imported from existing modules, Tailwind CDN JIT with no dynamic class construction, and a single-file view module that does not touch any other file.

**Critical schema finding:** The `content_ideas` table does not have `title`, `platform`, or `generated_script` columns as suggested in the phase brief. The actual columns are `brief_description` (the idea text), `type` (educational / entertainment / bts / ugc / trend), and `status`. Scripts live in the separate `content_scripts` table with `idea_id` as the FK. Platform lives on `content_calendar`, not on ideas. This affects card design and the script-preview fetch strategy.

**Primary recommendation:** Cards display `brief_description` (truncated) and `type` badge. Script preview requires a second `fetchTable` call on `content_scripts` filtered by `idea_id`. Only ideas with `status = 'scripted'` have an associated script.

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| IDEAS-01 | Kanban with 4 columns: pending / approved / rejected / scripted | fetchTable('content_ideas') → group by status client-side; 4-column flex layout |
| IDEAS-02 | Card shows idea title and platform (TikTok / Instagram) | `brief_description` is the closest to "title"; `type` replaces "platform" — schema has no platform on content_ideas |
| IDEAS-03 | Click on card expands preview of generated script (if present) | Only `status = 'scripted'` rows have a script; second fetchTable on 'content_scripts' keyed by idea_id; toggle inline expansion |
</phase_requirements>

---

## Schema Reality vs Phase Brief

**This is the most important finding for the planner.**

The phase brief (additional_context) describes columns `title`, `platform`, and `generated_script` on `content_ideas`. These do not exist.

### Actual `content_ideas` columns (from migrations/001_initial_schema.sql)
| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| type | TEXT | `educational`, `entertainment`, `bts`, `ugc`, `trend` |
| brief_description | TEXT | The idea text — maps to "title" in UI |
| relevance_score | INTEGER | 0-100, nullable |
| book_id | UUID | FK to books, nullable |
| status | TEXT | `pending`, `approved`, `rejected`, `scripted` |
| generated_at | TIMESTAMPTZ | Creation timestamp |
| metadata | JSONB | Stores `hook` and `cta` strings |

### Actual `content_scripts` columns (from migrations/001_initial_schema.sql)
| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| idea_id | UUID | FK to content_ideas |
| hook | TEXT | Opening hook text |
| full_script | TEXT | Complete script — this is the "script preview" content |
| cta | TEXT | Call to action |
| hashtags | TEXT[] | Array of hashtags |
| estimated_duration | INTEGER | Seconds |
| status | TEXT | `draft`, `approved`, `used` |
| created_at | TIMESTAMPTZ | |

### Platform field location
`platform` (`instagram` or `tiktok`) lives on `content_calendar`, not on `content_ideas`. Ideas have no platform column. Cards in Phase 4 cannot show platform from this table without a join.

### Reconciling IDEAS-02
The requirement says "card shows title and platform". Given the schema:
- **"title"** → render `brief_description` (truncated to ~80 chars)
- **"platform"** → render `type` badge instead (is present on the row, meaningful, same visual purpose)

This is a valid interpretation given the actual data model. Alternatively, platform could be omitted from idea cards entirely and noted as not applicable until Phase 5.

---

## Standard Stack

### Core (all already installed — zero new dependencies)
| Library | Source | Purpose |
|---------|--------|---------|
| `dashboard/js/api.js` | Project | `fetchTable(table, options)` — all data fetching |
| `dashboard/js/components.js` | Project | `renderLoading()`, `renderError(msg)`, `renderEmpty(msg)` |
| Tailwind CSS v4 CDN | `index.html` script tag | All styling — JIT, complete class strings only |

**Installation:** None required. All dependencies are already present.

---

## Architecture Patterns

### Recommended File
Single file: `dashboard/js/views/ideas.js` — replace stub entirely. No other files modified.

### Established Pattern (from books.js — HIGH confidence)
```
export async function renderIdeas() {
  1. [SYNC] app.innerHTML = renderIdeasSkeleton()   ← MUST be line 1
  2. [AWAIT] fetchTable('content_ideas', { order: 'generated_at' })
  3. [BRANCH]
     - error → app.innerHTML = renderError(error)
     - data.length === 0 → app.innerHTML = renderEmpty(...)
     - else → app.innerHTML = renderIdeasKanban(data)
}
```

### Kanban Layout Pattern
```
<div class="flex gap-4 overflow-x-auto">          ← horizontal scroll for narrow viewports
  <div class="flex-shrink-0 w-64">               ← fixed-width column
    <h2>pending</h2>
    <div class="flex flex-col gap-2">
      [cards...]
    </div>
  </div>
  ... repeat for approved, rejected, scripted
</div>
```

**Column order (fixed):** pending → approved → rejected → scripted

### Card Pattern
```html
<div class="bg-gray-800 rounded-lg p-3 cursor-pointer hover:bg-gray-700 transition-colors"
     data-idea-id="${idea.id}">
  <p class="text-white text-sm font-medium truncate">${escapeHtml(brief_description)}</p>
  <span class="text-xs text-gray-400 mt-1">${escapeHtml(idea.type)}</span>
  <!-- script preview container, hidden by default -->
  <div class="hidden mt-2 text-xs text-gray-300 whitespace-pre-wrap" data-script-preview>
  </div>
</div>
```

### Script Preview Pattern (IDEAS-03)
**Trigger:** click on any card where `idea.status === 'scripted'`
**Mechanism:** event delegation on the kanban container (one listener, not one per card)
**Data fetch:** second `fetchTable('content_scripts')` call, filtered with `options.filter = 'idea_id=eq.{id}'`

Note: `fetchTable` currently only supports `order`. For filtered queries, either:
- (a) Fetch ALL content_scripts once upfront alongside ideas, then filter client-side — simpler, no API changes
- (b) Extend fetchTable to support a `filter` option — one-line addition to api.js
- **Recommendation:** Option (a) for Phase 4 — fetch all scripts in a single call at load time, build a Map keyed by idea_id, inject into cards. Eliminates per-click network request, no api.js change needed.

### Skeleton Pattern (same discipline as books.js)
```
renderIdeasSkeleton() → 4-column kanban structure with 3 skeleton cards per column
Each skeleton card: bg-gray-800 rounded-lg, two animated pulse div bars
```

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead |
|---------|-------------|-------------|
| Data fetching | Custom fetch with headers/error handling | `fetchTable` from api.js |
| Error display | Custom red banner HTML | `renderError(msg)` from components.js |
| Empty column | Custom empty markup | `renderEmpty(msg)` from components.js |
| HTML escaping | Custom regex replace | `escapeHtml()` module-private in view file (copied from books.js pattern) |

---

## Common Pitfalls

### Pitfall 1: Dynamic Tailwind class construction
**What goes wrong:** Class names assembled from variables (e.g., `bg-${color}-800`) are never seen by the JIT compiler and render unstyled.
**How to avoid:** All Tailwind classes MUST appear as complete string literals. For status-based coloring, use an explicit object map:
```javascript
const statusColor = {
  pending:  'text-yellow-400',
  approved: 'text-green-400',
  rejected: 'text-red-400',
  scripted: 'text-indigo-400',
};
```
**Warning signs:** Column header or badge renders with no color applied.

### Pitfall 2: Blank #app during fetch
**What goes wrong:** If `app.innerHTML = skeleton` is not the first synchronous statement, `#app` is blank for the entire network roundtrip because `dispatch()` in router.js calls `view()` without await.
**How to avoid:** `app.innerHTML = renderIdeasSkeleton()` must be line 1 in `renderIdeas()`, before any `await`.

### Pitfall 3: Using wrong column name
**What goes wrong:** Using `idea.title` or `idea.platform` returns `undefined` — cards render empty or broken.
**How to avoid:** Use `idea.brief_description` for the idea text, `idea.type` for the type badge.

### Pitfall 4: Per-card click listeners in a loop
**What goes wrong:** Adding individual `addEventListener` to each card element inside a `.map()` on an HTML string is impossible — you can't attach listeners to strings, and using innerHTML loses listeners.
**How to avoid:** After setting `app.innerHTML = kanbanHtml`, attach a single delegated click listener on the kanban container element. Read `dataset.ideaId` from the click target (or closest card).

### Pitfall 5: Empty columns showing nothing
**What goes wrong:** If a status bucket has zero ideas, the column renders with no content — should show `renderEmpty()` message.
**How to avoid:** In each column render, check `columnIdeas.length === 0` and inject `renderEmpty('No ideas yet')` in the card area.

### Pitfall 6: Script preview fetch on every click
**What goes wrong:** Fetching content_scripts on each card click causes visible latency and unnecessary network calls.
**How to avoid:** Fetch all scripts once at load time, build a `Map<ideaId, script>` client-side, use it synchronously in the click handler.

---

## Code Examples

### escapeHtml (module-private, same pattern as books.js)
```javascript
// Source: dashboard/js/views/books.js (established project pattern)
function escapeHtml(str) {
  return String(str ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
```

### fetchTable call (from api.js)
```javascript
// Source: dashboard/js/api.js
// Always resolves { data: Array|null, error: string|null }
const { data: ideas, error } = await fetchTable('content_ideas', { order: 'generated_at' });
const { data: scripts } = await fetchTable('content_scripts');
```

### Status-based column header colors (safe Tailwind literals)
```javascript
// Source: established Tailwind CDN JIT constraint
const COLUMN_LABEL_CLASSES = {
  pending:  'text-yellow-400',
  approved: 'text-green-400',
  rejected: 'text-red-400',
  scripted: 'text-indigo-400',
};
const COLUMNS = ['pending', 'approved', 'rejected', 'scripted'];
```

### Event delegation for script preview toggle
```javascript
// Source: standard JS pattern — attach AFTER app.innerHTML is set
document.getElementById('app').addEventListener('click', (e) => {
  const card = e.target.closest('[data-idea-id]');
  if (!card) return;
  const ideaId = card.dataset.ideaId;
  const previewEl = card.querySelector('[data-script-preview]');
  if (!previewEl) return;  // card has no preview slot (non-scripted)
  previewEl.classList.toggle('hidden');
  // If first expand, populate from scripts map
  if (!previewEl.dataset.loaded) {
    const script = scriptsMap.get(ideaId);
    previewEl.textContent = script ? script.full_script : 'No script found.';
    previewEl.dataset.loaded = 'true';
  }
});
```

### GroupBy helper (module-private)
```javascript
function groupByStatus(ideas) {
  const groups = { pending: [], approved: [], rejected: [], scripted: [] };
  for (const idea of ideas) {
    if (groups[idea.status]) groups[idea.status].push(idea);
  }
  return groups;
}
```

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | None — vanilla JS with no bundler, no test runner installed for dashboard |
| Config file | None |
| Quick run command | Open `dashboard/index.html` in browser, navigate to `#/ideas` |
| Full suite command | Manual visual verification of all 4 success criteria |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| IDEAS-01 | 4 columns visible with live data | manual smoke | Open `#/ideas`, verify column headers | N/A |
| IDEAS-02 | Card shows brief_description + type | manual smoke | Inspect card content | N/A |
| IDEAS-03 | Click scripted card → script expands inline | manual smoke | Click a scripted card, verify text appears | N/A |

### Sampling Rate
- **Per task commit:** Load `dashboard/index.html#/ideas` in browser
- **Per wave merge:** All 4 success criteria visually verified
- **Phase gate:** Human checkpoint confirms all 4 criteria TRUE before verify-work

### Wave 0 Gaps
None — no automated test infrastructure needed for this phase. Verification is manual browser inspection per established project pattern.

---

## Open Questions

1. **IDEAS-02 platform display**
   - What we know: `content_ideas` has no `platform` column; `platform` is on `content_calendar`
   - What's unclear: Whether the requirement means "show platform from calendar" (requires JOIN or second fetch) or is based on an assumed schema that doesn't match reality
   - Recommendation: Show `type` badge on cards instead of platform. Note the discrepancy in the plan. Platform display can be added in Phase 5 when calendar data is also loaded.

2. **Script preview when no scripts exist yet**
   - What we know: An idea with `status = 'scripted'` should have a matching `content_scripts` row — but is not guaranteed
   - Recommendation: For scripted cards, show preview toggle; if `scriptsMap.get(ideaId)` is undefined, show "Script not found." text.

---

## Sources

### Primary (HIGH confidence)
- `migrations/001_initial_schema.sql` — definitive column names for `content_ideas` and `content_scripts`
- `dashboard/js/api.js` — fetchTable contract (always resolves, order support)
- `dashboard/js/components.js` — renderError, renderEmpty signatures
- `dashboard/js/router.js` — dispatch() does not await view functions (sync skeleton requirement)
- `dashboard/js/views/books.js` — canonical pattern for all view modules in this project
- `internal/generator/ideas.go` — confirms `brief_description = title + ": " + description`, metadata stores hook/cta only (no platform)

### Secondary (MEDIUM confidence)
- `.planning/STATE.md` decisions log — confirms Tailwind JIT complete-string constraint, sync skeleton pattern
- `dashboard/index.html` — confirms Tailwind v4 CDN, `#app` mount point

---

## Metadata

**Confidence breakdown:**
- Schema facts: HIGH — sourced directly from migration SQL
- Architecture patterns: HIGH — sourced from existing implemented books.js
- Tailwind JIT constraint: HIGH — documented in STATE.md decisions + books.js evidence
- Script preview implementation strategy: MEDIUM — no prior art in codebase, standard JS pattern

**Research date:** 2026-02-25
**Valid until:** 2026-03-25 (stable codebase, no fast-moving dependencies)
