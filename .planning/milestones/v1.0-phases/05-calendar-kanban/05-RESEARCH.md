# Phase 5: Calendar Kanban - Research

**Researched:** 2026-02-28
**Domain:** Vanilla JS dashboard — five-column kanban for content_calendar with multi-table join and platform color coding
**Confidence:** HIGH

## Summary

Phase 5 implements the Calendar tab as a five-column kanban mirroring the Ideas kanban pattern established in Phase 4. The primary complexity is data resolution: calendar cards must display the idea title, but `content_calendar` only stores a `script_id` FK. Resolving the title requires fetching `content_scripts` and `content_ideas` in parallel and building two client-side Maps for O(1) lookup — the same pattern used for script previews in Phase 4.

The second complexity is the status mapping: CAL-01 names columns as `scheduled / approved / publishing / published / failed`, but the actual DB column values are `pending_approval / approved / publishing / published / failed`. The column labeled "scheduled" must filter on `status = 'pending_approval'`. This is a naming mismatch that must be handled explicitly in the column config.

The phase is pure frontend work — one file (`dashboard/js/views/calendar.js`) replaces its stub. No new API endpoints, no migrations, no backend changes.

**Primary recommendation:** Follow the ideas.js pattern exactly — sync skeleton first, Promise.all fetch, client-side Maps for join resolution, event delegation. Use complete Tailwind class strings (no dynamic assembly) for JIT safety.

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Tailwind CSS | v4 CDN | Styling | Already loaded in index.html via script tag |
| supabase-js | v2 CDN | Data fetch | Configured in config.js, wrapped by api.js |
| Vanilla JS ES modules | Browser native | Component logic | Project decision: zero build tooling |

### No Additional Dependencies

All required infrastructure is already in place from Phases 1-4. No new installs.

## Architecture Patterns

### Recommended Project Structure

```
dashboard/js/views/
├── books.js        # Table view (canonical pattern)
├── ideas.js        # Kanban view (direct template for Phase 5)
└── calendar.js     # STUB — replace in Phase 5
```

### Pattern 1: Sync Skeleton + Promise.all Fetch

**What:** Render skeleton HTML synchronously as line 1, then fetch all required data in parallel, then replace with real content.
**When to use:** Every view — router.js does not await view functions.
**Example:**
```javascript
// Source: dashboard/js/views/ideas.js (verified)
export async function renderCalendar() {
  const app = document.getElementById('app');
  app.innerHTML = renderCalendarSkeleton();  // MUST be line 1 — sync

  const [calResult, scriptsResult, ideasResult] = await Promise.all([
    fetchTable('content_calendar', { order: 'scheduled_for' }),
    fetchTable('content_scripts'),
    fetchTable('content_ideas'),
  ]);
  // ...
}
```

### Pattern 2: Multi-Table Client-Side Join via Map

**What:** Fetch all three tables at page load, build Maps for O(1) lookup, resolve title on card render.
**When to use:** Any time a card needs data from a related table without PostgREST joins.
**Example:**
```javascript
// Build lookup chain: script_id -> idea_id -> brief_description
const scriptsMap = new Map();  // id -> idea_id
for (const s of scriptsResult.data ?? []) {
  scriptsMap.set(s.id, s.idea_id);
}
const ideasMap = new Map();  // id -> brief_description
for (const i of ideasResult.data ?? []) {
  ideasMap.set(i.id, i.brief_description);
}

// On card render:
function getIdeaTitle(scriptId) {
  const ideaId = scriptsMap.get(scriptId);
  return ideasMap.get(ideaId) ?? '(untitled)';
}
```

### Pattern 3: Status-to-Column Mapping with DB Name Mismatch

**What:** The UI column labeled "scheduled" corresponds to `status = 'pending_approval'` in the DB. Column config must map display name to DB value.
**When to use:** Required — do not filter on `status === 'scheduled'` (no rows would match).
**Example:**
```javascript
// Complete string literals required for Tailwind JIT
const COLUMNS = [
  { label: 'scheduled',   dbStatus: 'pending_approval', labelClass: 'text-blue-400'   },
  { label: 'approved',    dbStatus: 'approved',          labelClass: 'text-green-400'  },
  { label: 'publishing',  dbStatus: 'publishing',        labelClass: 'text-yellow-400' },
  { label: 'published',   dbStatus: 'published',         labelClass: 'text-indigo-400' },
  { label: 'failed',      dbStatus: 'failed',            labelClass: 'text-red-400'    },
];
```

### Pattern 4: Platform Color Accent on Cards

**What:** TikTok cards get one color accent, Instagram cards get another. Use a complete class string map.
**When to use:** CAL-02 requires platform to be visually distinguishable.
**Example:**
```javascript
// Complete literals — Tailwind JIT cannot detect dynamic strings
const PLATFORM_BADGE_CLASSES = {
  tiktok:    'text-pink-400',
  instagram: 'text-purple-400',
};
```

### Anti-Patterns to Avoid

- **Dynamic Tailwind class assembly:** `'text-' + color + '-400'` — Tailwind JIT purges classes not found as complete strings. Always use full string literals in a config object.
- **Per-card network requests:** Fetching idea title on click rather than pre-loading — causes N+1 requests. Pre-fetch with Promise.all.
- **Filtering on `status === 'scheduled'`:** No row has that value. Filter on `'pending_approval'`.
- **Awaiting before setting skeleton:** Any await before `app.innerHTML = skeleton` causes a blank flash.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTML escaping | Custom regex | Copy `escapeHtml` from ideas.js verbatim | Already battle-tested, handles null via `??` |
| Error/empty states | Custom UI | `renderError` / `renderEmpty` from components.js | Consistent with all other views |
| Data fetching | Direct supabase calls | `fetchTable` from api.js | Handles error normalization, always resolves |
| Loading state | CSS spinner | Skeleton HTML (same pattern as ideas.js) | Consistent UX, already established pattern |

**Key insight:** Every shared UI primitive already exists. Phase 5 is purely composition of existing patterns — no new infrastructure.

## Common Pitfalls

### Pitfall 1: Wrong Status Value for "Scheduled" Column
**What goes wrong:** Column shows 0 cards even when DB has pending_approval rows.
**Why it happens:** Developer filters `idea.status === 'scheduled'` but DB stores `'pending_approval'`.
**How to avoid:** Use a column config object with separate `label` and `dbStatus` fields. Group by `dbStatus`.
**Warning signs:** Empty "scheduled" column in browser despite CLI having created calendar entries.

### Pitfall 2: Dynamic Tailwind Classes Purged
**What goes wrong:** Column headers or platform badges render without color.
**Why it happens:** Tailwind CDN v4 JIT scans for complete class strings — dynamic concatenation (`'text-' + colorVar`) is not detected.
**How to avoid:** All classes in config objects must be full literals: `'text-pink-400'`, not `'text-' + platform + '-400'`.
**Warning signs:** Classes appear correct in source but no color in browser.

### Pitfall 3: Idea Title Resolution Fails Silently
**What goes wrong:** Cards show "(untitled)" or blank for idea title.
**Why it happens:** calendar row has `script_id` → content_scripts has `idea_id` → content_ideas has `brief_description`. If either Map lookup fails, the chain breaks.
**How to avoid:** Always provide a fallback: `ideasMap.get(ideaId) ?? '(no title)'`. Log if scriptsMap misses `script_id`.
**Warning signs:** All cards show "(untitled)" in production.

### Pitfall 4: Duplicate Event Listeners on Re-navigation
**What goes wrong:** Click handler fires multiple times after returning to Calendar tab.
**Why it happens:** Each call to `renderCalendar()` adds a new listener to `#app` without removing the previous one.
**How to avoid:** Use event delegation on `#app` — this is already safe because `app.innerHTML =` replaces the DOM, and listeners on `#app` element itself accumulate. Use `{ once: false }` but guard with `if (!card) return` to avoid double execution, OR use `app.replaceChildren()` pattern. Simplest fix: add listener before innerHTML assignment so it only fires when the Calendar view DOM is active.
**Warning signs:** Single click triggers multiple previews or multiple state changes.

## Code Examples

Verified patterns from existing codebase:

### Complete renderCalendar() Structure
```javascript
// Source: dashboard/js/views/ideas.js (verified pattern)
export async function renderCalendar() {
  const app = document.getElementById('app');
  app.innerHTML = renderCalendarSkeleton();  // line 1 — sync

  const [calResult, scriptsResult, ideasResult] = await Promise.all([
    fetchTable('content_calendar', { order: 'scheduled_for' }),
    fetchTable('content_scripts'),
    fetchTable('content_ideas'),
  ]);

  if (calResult.error) {
    app.innerHTML = renderError(calResult.error);
    return;
  }

  const entries = calResult.data;

  if (entries.length === 0) {
    app.innerHTML = renderEmpty('No calendar entries yet — plan some with the CLI');
    return;
  }

  // Build join Maps
  const scriptsMap = new Map();
  for (const s of scriptsResult.data ?? []) scriptsMap.set(s.id, s.idea_id);
  const ideasMap = new Map();
  for (const i of ideasResult.data ?? []) ideasMap.set(i.id, i.brief_description);

  app.innerHTML = renderCalendarKanban(entries, scriptsMap, ideasMap);
}
```

### Card Rendering with Platform Accent
```javascript
// Source: pattern derived from ideas.js renderCard()
const PLATFORM_BADGE_CLASSES = {
  tiktok:    'text-pink-400',
  instagram: 'text-purple-400',
};

function renderCard(entry, ideaTitle) {
  const badgeClass = PLATFORM_BADGE_CLASSES[entry.platform] ?? 'text-gray-400';
  const date = entry.scheduled_for
    ? new Date(entry.scheduled_for).toLocaleDateString()
    : '—';
  return `
    <div class="bg-gray-800 rounded-lg p-3">
      <p class="text-white text-sm font-medium truncate">${escapeHtml(ideaTitle)}</p>
      <span class="text-xs ${badgeClass} mt-1 block">${escapeHtml(entry.platform)}</span>
      <span class="text-xs text-gray-400 mt-1 block">${escapeHtml(date)}</span>
    </div>`;
}
```

### escapeHtml (copy verbatim from ideas.js)
```javascript
// Source: dashboard/js/views/ideas.js (verified)
function escapeHtml(str) {
  return String(str ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `status = 'pending_approval'` | Same — unchanged | Always | Column label "scheduled" is purely UI; must not change DB filter |
| Single fetch per view | Promise.all parallel fetch | Phase 4 | Established pattern — use for all three tables |

## Database Schema (Verified)

`content_calendar` columns relevant to Phase 5:
- `id` (UUID)
- `script_id` (UUID FK → content_scripts.id)
- `scheduled_for` (TIMESTAMPTZ) — use for ordering and display
- `platform` (TEXT: `'instagram'` or `'tiktok'`)
- `status` (TEXT: `'pending_approval'` | `'approved'` | `'publishing'` | `'published'` | `'failed'`)
- `post_type` (TEXT: `'reel'` | `'story'` | `'feed'`) — available but not required for CAL-01/02/03

Join chain to resolve idea title:
```
content_calendar.script_id
  -> content_scripts.id (gives idea_id)
  -> content_ideas.id (gives brief_description)
```

## Open Questions

1. **Cards without `script_id`**
   - What we know: `script_id` is nullable in the schema (REFERENCES ... ON DELETE CASCADE, no NOT NULL)
   - What's unclear: Can calendar entries exist with null script_id? CLI flow requires script before calendar, so unlikely in practice
   - Recommendation: Guard with `?? '(no title)'` fallback — safe either way

2. **Date display format**
   - What we know: `scheduled_for` is TIMESTAMPTZ
   - What's unclear: Whether user prefers locale date, ISO date, or relative ("tomorrow")
   - Recommendation: Use `new Date(entry.scheduled_for).toLocaleDateString()` — simple, locale-aware, no library needed

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CAL-01 | Kanban with 5 columns: scheduled / approved / publishing / published / failed | DB statuses confirmed; "scheduled" maps to `pending_approval`; column config pattern defined |
| CAL-02 | Card shows scheduled date and platform | `scheduled_for` and `platform` columns verified in schema; platform badge color pattern defined |
| CAL-03 | Card shows title of linked idea | Three-table join chain documented: calendar → scripts → ideas; client-side Map pattern defined |
</phase_requirements>

## Sources

### Primary (HIGH confidence)
- `dashboard/js/views/ideas.js` — canonical kanban pattern for Phase 5
- `dashboard/js/api.js` — fetchTable interface (verified)
- `dashboard/js/components.js` — renderError, renderEmpty (verified in-use)
- `migrations/001_initial_schema.sql` — content_calendar table definition
- `migrations/004_add_cron_publishing_support.sql` — publishing status and updated_at added
- `.planning/REQUIREMENTS.md` — CAL-01, CAL-02, CAL-03 definitions
- `.planning/STATE.md` — project decisions (Tailwind JIT literal constraint, no-bundler, etc.)

### Secondary (MEDIUM confidence)
- Phase 4 plan (`04-01-PLAN.md`) — implementation patterns cross-referenced

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — identical to Phases 1-4, no new dependencies
- Architecture: HIGH — verified against existing ideas.js implementation
- DB schema: HIGH — verified against actual migration files
- Pitfalls: HIGH — derived from documented Phase 4 decisions and known Tailwind JIT constraint

**Research date:** 2026-02-28
**Valid until:** 2026-04-28 (stable vanilla JS + Supabase stack, no fast-moving dependencies)
