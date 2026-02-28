# Retrospective

## Milestone: v1.0 — Dashboard MVP

**Shipped:** 2026-02-28
**Phases:** 5 | **Plans:** 7

### What Was Built

- Secure static dashboard on Vercel with Supabase RLS SELECT-only policies (migration 009)
- Hash router + Tailwind CDN v4 shell with dark navbar and FOUC mitigation
- Shared `api.js` fetch wrapper + `components.js` (skeleton, error, empty state renderers)
- Books view: live table with kdp_asin → Amazon product page links
- Ideas kanban: four-column live view with inline script preview on click
- Calendar kanban: five-column live view, multi-table join via client-side Map, platform color-coding

### What Worked

- **Skeleton-first render pattern** — render skeleton synchronously before any `await`, zero blank-screen flicker
- **Zero build step** — Tailwind CDN + ES modules, deploy via `git push`, no config friction
- **Shared renderers** — `renderError()` / `renderEmpty()` / `renderSkeleton()` reused across all 3 views saved time
- **Promise.all + Map join** — fetching both tables in parallel then resolving O(1) was cleaner than nested fetches

### What Was Inefficient

- Phase 2 roadmap entry was incomplete (0/2 shown despite files on disk) — STATE.md / ROADMAP.md sync issue
- `[data-nav]` selector pattern discovered mid-Phase 2 instead of planned upfront
- `book.kdp_asin` vs `book.asin` mismatch required checking migration SQL directly — column names should be documented

### Patterns Established

- **skeleton-first**: Every async view function renders skeleton synchronously as line 1
- **`[data-nav="${route}"]`**: Attribute selector for nav highlighting instead of structural selectors
- **`Promise.all` + `new Map()`**: Standard pattern for multi-table client-side joins
- **Literal Tailwind classes in JS objects**: Use full class strings in PLATFORM_BADGE_CLASSES-style maps for JIT compat

### Key Lessons

- Check actual Supabase column names against migration SQL before wiring views — don't assume field names
- Router `dispatch()` is fire-and-forget (no await) — view functions must handle their own async internally
- Tailwind JIT requires literal class strings — no dynamic concatenation in JS

### Cost Observations

- Model mix: 100% Sonnet 4.6
- Sessions: ~5-6 sessions across 3 days (Feb 25-28)
- Notable: vanilla stack and zero tooling kept implementation fast; most time was verification/UAT

---

## Cross-Milestone Trends

| Milestone | Phases | Plans | Days | Stack |
|-----------|--------|-------|------|-------|
| v1.0 Dashboard MVP | 5 | 7 | 3 | Vanilla JS + Tailwind CDN |
