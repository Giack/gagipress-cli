---
phase: 02-data-layer-shared-components
plan: "01"
subsystem: dashboard-utilities
tags: [javascript, supabase, ui-components, fetch-wrapper]
dependency_graph:
  requires: [dashboard/config.js]
  provides: [dashboard/js/api.js, dashboard/js/components.js]
  affects: [dashboard/js/views/books.js, dashboard/js/views/ideas.js, dashboard/js/views/calendar.js]
tech_stack:
  added: []
  patterns: [fetch-wrapper-pattern, html-string-renderer-pattern]
key_files:
  created:
    - dashboard/js/api.js
    - dashboard/js/components.js
  modified: []
decisions:
  - "fetchTable always resolves with { data, error } — callers never need try/catch"
  - "escapeHtml kept internal (not exported) — it is an implementation detail not part of the public API"
  - "optional options.order parameter added to fetchTable to prevent rework in phases 3-5"
metrics:
  duration: "4 minutes"
  completed: "2026-02-25"
  tasks_completed: 2
  files_changed: 2
---

# Phase 02 Plan 01: Data Layer Shared Components Summary

Supabase fetch wrapper and dark-palette UI state renderers — pure utility modules with no DOM side effects that establish the contract for all view phases.

## What Was Built

Two foundational utility modules consumed by all view phases (3-5):

1. **`dashboard/js/api.js`** — Supabase fetch wrapper with `fetchTable(table, options)`. Always resolves with `{ data: Array|null, error: string|null }` — never throws. Accepts optional `options.order` for sorted results. Imports `supabase` client from `../config.js`.

2. **`dashboard/js/components.js`** — Three HTML string renderer functions: `renderLoading()`, `renderError(message)`, `renderEmpty(message)`. All return HTML strings only — no DOM operations. Internal `escapeHtml()` helper prevents broken HTML from Supabase error messages containing `<`, `>`, `&`, `"`.

## View Contract (Phases 3-5 Must Follow)

Every view render function must follow this exact pattern:

```javascript
import { fetchTable } from '../api.js';
import { renderLoading, renderError, renderEmpty } from '../components.js';

export async function renderBooks() {
  const app = document.getElementById('app');
  app.innerHTML = renderLoading();                                    // 1. Show spinner immediately
  const { data, error } = await fetchTable('books');
  if (error) { app.innerHTML = renderError(error); return; }         // 2. Handle error
  if (data.length === 0) { app.innerHTML = renderEmpty('No books found.'); return; }  // 3. Handle empty
  app.innerHTML = `...`;                                              // 4. Render content
}
```

This pattern ensures consistent UX across all views: loading state always shown before fetch, error state always shown on failure, empty state always shown when no rows returned.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 | 092a557 | feat(02-01): add dashboard/js/api.js — Supabase fetch wrapper |
| 2 | ecb4d7e | feat(02-01): add dashboard/js/components.js — shared UI state renderers |

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED

- dashboard/js/api.js: FOUND
- dashboard/js/components.js: FOUND
- commit 092a557: FOUND
- commit ecb4d7e: FOUND
