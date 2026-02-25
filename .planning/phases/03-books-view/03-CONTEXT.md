# Phase 3: Books View - Context

**Gathered:** 2026-02-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Display all books from the Supabase `books` table in a live-data table with four columns: title, ASIN, genre, target audience. ASIN cells link to Amazon product pages. Loading, error, and empty states are handled using the Phase 2 components. No editing, filtering, or sorting beyond what's noted below.

</domain>

<decisions>
## Implementation Decisions

### Table layout & columns
- Column order: Title → ASIN → Genre → Target Audience
- Long titles: truncate with ellipsis, reveal full title on hover tooltip
- Sticky header — stays visible when the list scrolls
- Column sorting: Claude's discretion (user deferred this)

### Empty & loading states
- Loading: skeleton rows that match the table column structure (not a plain spinner)
- Empty state: replace the entire table with the `renderEmpty()` component — no orphaned headers
- Empty copy: "No books in your catalog yet"
- Error: red error banner using `renderError()` from Phase 2 — no retry button needed in Phase 3

### Row interactions
- ASIN cell is a styled `<a>` tag pointing to `https://www.amazon.com/dp/{ASIN}`, `target="_blank"`
- ASIN link: indigo color + underline on hover, matches Phase 2 palette
- Row hover: subtle background highlight (signals interactivity)
- No other row-level actions — read-only table in Phase 3

### Claude's Discretion
- Column sorting behavior (user deferred)
- Exact skeleton row count and animation style
- Table border vs borderless styling
- Row hover color value (within the dark palette)

</decisions>

<specifics>
## Specific Ideas

- Skeleton rows should visually match the table column structure so the layout doesn't jump when data arrives
- ASIN links must open in a new tab (`target="_blank" rel="noopener"`)
- Stay consistent with the indigo-500 palette established in Phase 2 for interactive elements

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-books-view*
*Context gathered: 2026-02-25*
