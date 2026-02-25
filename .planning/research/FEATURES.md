# Feature Landscape

**Domain:** Read-only content pipeline dashboard (kanban + catalog views)
**Researched:** 2026-02-25
**Confidence:** MEDIUM — grounded in UX research and kanban literature; no direct competitors for this exact personal-tool niche

---

## Table Stakes

Features the dashboard is useless without. Missing any of these = dashboard gets opened once and abandoned.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Ideas kanban with all status columns | Core value prop — see pipeline at a glance (pending / approved / rejected / scripted) | Low | Columns map 1:1 to existing DB `status` field |
| Calendar kanban with all status columns | Second core view — see what's scheduled/published (scheduled / approved / publishing / published / failed) | Low | Same pattern as ideas kanban |
| Card shows enough context to be meaningful | Title + relevant metadata visible on card face without clicking — otherwise cards are just colored rectangles | Low | Ideas: title + book name. Calendar: title + platform + scheduled date |
| Books catalog table | Reference data — need to see what books exist and their metadata (title, ASIN, genre, audience) | Low | Simple `<table>` is fine; no sorting needed for <20 books |
| Column item counts | At a glance: "7 pending, 2 approved, 0 scripted" — tells you immediately where the bottleneck is | Low | Count per column in column header |
| Data loads without stale cache | Dashboard showing yesterday's data is worse than no dashboard — creates false confidence | Low | No caching, direct Supabase fetch on page load |
| Visual distinction between columns | Color-coded columns or card borders — `failed` must look different from `published` | Low | Tailwind color utilities, no library needed |
| Empty state per column | "No items" is informative; a blank column looks broken | Low | Simple text placeholder |

## Differentiators

Features that make the dashboard worth returning to daily rather than once a week.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Last-refreshed timestamp | Tells you the data age without guessing — "fetched 30s ago" builds trust in the data | Low | `new Date()` after fetch, display in header |
| Manual refresh button | Quick re-fetch without full page reload — important when you've just run a CLI command and want to see the result | Low | One `fetch()` call wired to a button |
| Per-platform color coding on calendar cards | Instagram vs TikTok should be visually distinct — you scan for platform, not read it | Low | Two colors, Tailwind classes |
| `failed` cards visually alarming | Red border / warning icon on failed calendar items — a failed post is the one thing you must not miss | Low | CSS class on status === 'failed' |
| Book title on idea cards | You often forget which book an idea belongs to — showing the book name prevents context-switching to the books tab | Low | One extra field from JOIN or second fetch |
| Scheduled date on calendar cards | The date is the most time-sensitive data point on a calendar item — show it prominently | Low | ISO date formatted as "Mon Feb 25" |
| Navigation tabs (Ideas / Calendar / Books) | Three views is too many for one screen; tabs keep each view focused | Low | Pure CSS tab switching, no routing library |
| Page title / identity | "Gagipress Dashboard" in the header — trivial but makes it feel like a real tool | Low | Static HTML |

## Anti-Features

Features to deliberately NOT build for a personal tool.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Click-through detail modals | For a read-only tool, you rarely need more detail than the card shows. Modals add complexity with minimal payoff. | Show all important fields on the card face. |
| Search / filter | With <50 items per column, visual scan is faster than typing a query. Filter UI adds code and cognitive overhead. | Keep columns narrow, let the eye do the work. |
| Sorting controls | Same as filtering — unnecessary for small datasets. The DB default order (by created_at desc) is usually the right order. | Accept DB natural order. |
| Pagination | For a personal pipeline with <200 total items, paginating breaks the "see everything at once" mental model. | Fetch all and render all. |
| Real-time auto-refresh (polling/websocket) | Adds complexity, burns Supabase quota, and creates flickering UX. The dashboard is a snapshot tool, not a live monitor. | Manual refresh button is enough. |
| Charts / analytics | Sales and engagement data is out of scope for v1. Charts on top of a kanban create visual noise. | Dedicate a future tab to analytics if needed. |
| Write operations (approve, reject, publish) | This is the most dangerous anti-feature. A "quick approve" button sounds useful but erodes the CLI-as-source-of-truth contract and requires auth. | All mutations stay in the CLI. Dashboard is read-only by design. |
| Authentication / login | Single-user personal tool on a private URL. Auth is pure overhead with zero security benefit at this scale. | Use a non-guessable Vercel URL. |
| Mobile / responsive layout | The dashboard is a desk tool used while the CLI is running in a terminal. Mobile optimization is wasted effort for v1. | Desktop-first layout, accept horizontal scroll on mobile. |
| Dark/light theme toggle | Doubles CSS complexity. Pick one (dark) and ship it. | Dark only, Tailwind `dark:` variant not needed. |

## Feature Dependencies

```
Navigation tabs → Ideas kanban view
Navigation tabs → Calendar kanban view
Navigation tabs → Books catalog table

Ideas kanban view → Column item counts (same render pass)
Ideas kanban view → Book title on idea cards (requires book data)

Calendar kanban view → Per-platform color coding
Calendar kanban view → Failed cards visual alarm
Calendar kanban view → Scheduled date on cards
Calendar kanban view → Column item counts

Manual refresh button → Last-refreshed timestamp (update timestamp on refresh)

Book title on idea cards → Books catalog data loaded (or second fetch)
```

## MVP Recommendation

Build in this order — each step is independently shippable:

1. **Books catalog table** — simplest view, proves Supabase connection works
2. **Ideas kanban** — core pipeline view; implement column structure, cards with title + status
3. **Calendar kanban** — same pattern as ideas, add platform color coding and `failed` alarm
4. **Column counts + last-refreshed + manual refresh** — cheap polish that makes the tool trustworthy
5. **Book title on idea cards** — one extra data join; do last to keep data fetching simple

Defer indefinitely: search, filters, charts, write operations, auth, mobile layout.

## What Makes a Dashboard Actually Used Daily

Research-backed observations translated to this specific tool:

- **Bottleneck visibility is the primary value.** "7 ideas pending, 0 scripted" tells you immediately what the CLI needs to do next. Optimize for this signal.
- **Five-second rule.** The most important information must be scannable in under 5 seconds. Column headers with counts are more valuable than card details.
- **Trust requires freshness signal.** A timestamp or refresh button is non-optional — a dashboard you can't trust in terms of data age gets ignored.
- **Failed items must demand attention.** A red/alarming visual on `failed` calendar items is the one case where the dashboard can alert you to something you'd otherwise miss in the CLI.
- **Don't fight the workflow.** The CLI is the write tool; the dashboard is the read tool. Any blurring of this contract will make both worse.

## Sources

- [Dashboard Design Principles — UXPin](https://www.uxpin.com/studio/blog/dashboard-design-principles/) — MEDIUM confidence
- [From Good To Great In Dashboard Design — Smashing Magazine](https://www.smashingmagazine.com/2021/11/dashboard-design-research-decluttering-data-viz/) — MEDIUM confidence
- [What is a Kanban Board — Atlassian](https://www.atlassian.com/agile/kanban/boards) — HIGH confidence
- [Dashboard Design Best Practices — Toptal](https://www.toptal.com/designers/data-visualization/dashboard-design-best-practices) — MEDIUM confidence
- [Dashboard UX Design — Lazarev Agency](https://lazarev.agency/articles/dashboard-ux-design) — LOW confidence (single source)
