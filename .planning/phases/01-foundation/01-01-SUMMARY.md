---
phase: 01-foundation
plan: 01
subsystem: infra
tags: [gitignore, rls, supabase, security, dashboard, postgresql]

requires: []
provides:
  - dashboard/config.js excluded from git history before credentials are written
  - dashboard/config.example.js ESM template committed for developer setup
  - migration 009 replacing FOR ALL anon policies with SELECT-only anon policies on 6 tables
affects:
  - 01-02-static-shell
  - 01-03-books-view
  - 01-04-ideas-view
  - 01-05-calendar-view

tech-stack:
  added: []
  patterns:
    - "gitignore credentials before creating them — block credential leak at the infrastructure layer"
    - "RLS two-tier pattern: authenticated full access + anon SELECT-only per table"

key-files:
  created:
    - dashboard/config.example.js
    - migrations/009_dashboard_rls_anon_select.sql
  modified:
    - .gitignore

key-decisions:
  - "Include content_scripts in migration even though Phase 1 dashboard does not query it — complete RLS coverage is safer than partial"
  - "Re-create authenticated full access policies explicitly alongside DROP — maintains policy clarity even though service key bypasses RLS"

patterns-established:
  - "Dashboard JS uses ES module syntax (import/export const) — no CommonJS or bundler required"
  - "config.example.js committed, config.js gitignored — always create gitignore entry before the secret file"

requirements-completed: [INFRA-03, INFRA-04]

duration: 8min
completed: 2026-02-25
---

# Phase 1 Plan 01: Security Baseline (gitignore + RLS) Summary

**dashboard/config.js gitignored before first write, and migration 009 replaces FOR ALL anon RLS with SELECT-only on all 6 Supabase tables**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-02-25T00:00:00Z
- **Completed:** 2026-02-25T00:08:00Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added `dashboard/config.js` to `.gitignore` before the file is created — no credential leak risk
- Created `dashboard/config.example.js` with ESM import/export syntax as the developer setup template
- Wrote `migrations/009_dashboard_rls_anon_select.sql` covering all 6 tables with DROP + re-create authenticated + anon SELECT policies

## Task Commits

Each task was committed atomically:

1. **Task 1: Add dashboard/config.js to .gitignore and create config.example.js** - `1a13a57` (chore)
2. **Task 2: Write RLS migration 009 — SELECT-only anon policies** - `b70ce4b` (feat)

## Files Created/Modified

- `.gitignore` — Added `dashboard/config.js` under Dashboard credentials comment
- `dashboard/config.example.js` — ESM template: imports supabase-js from CDN, exports SUPABASE_URL, SUPABASE_ANON_KEY, and supabase client
- `migrations/009_dashboard_rls_anon_select.sql` — Drops FOR ALL anon policies, re-creates authenticated full access, creates anon SELECT-only on: books, content_ideas, content_scripts, content_calendar, post_metrics, sales_data

## Decisions Made

- Included `content_scripts` table in migration even though Phase 1 dashboard does not query it — complete RLS coverage is safer than leaving one table with an over-permissive policy
- Re-created authenticated full access policies explicitly after DROP to maintain clarity (service key bypasses RLS, but explicit policies are best practice)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None at this stage. Migration must be applied manually before the dashboard can query data:

```bash
supabase db push
```

## Next Phase Readiness

- `.gitignore` is ready — `dashboard/config.js` can be created safely at any time
- Migration 009 is ready to apply — run `supabase db push` with valid Supabase credentials
- Phase 01-02 (static HTML shell + Vercel config) can proceed immediately

---
*Phase: 01-foundation*
*Completed: 2026-02-25*
