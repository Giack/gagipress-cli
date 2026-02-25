# Roadmap: Gagipress Dashboard

## Overview

Five phases deliver a working read-only pipeline dashboard: secure infrastructure first, then shared data-access and UI components, then three progressively complex views (books table, ideas kanban, calendar kanban). Every phase ends with something verifiable in a browser.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation** - Secure project scaffold: RLS policies, config setup, index.html shell with hash router
- [ ] **Phase 2: Data Layer + Shared Components** - Fetch wrapper, kanban and table renderers, loading/error/empty states
- [ ] **Phase 3: Books View** - Live books table wired to Supabase, full fetch-to-render pipeline validated
- [ ] **Phase 4: Ideas Kanban** - Four-column kanban with live ideas data, column counts, script preview
- [ ] **Phase 5: Calendar Kanban** - Five-column calendar kanban with platform coloring and dark design polish

## Phase Details

### Phase 1: Foundation
**Goal**: Dashboard project structure is in place and Supabase data is secured so any live query is safe to make
**Depends on**: Nothing (first phase)
**Requirements**: INFRA-01, INFRA-02, INFRA-03, INFRA-04, INFRA-05
**Success Criteria** (what must be TRUE):
  1. Opening the Vercel URL (or local `index.html`) shows a styled shell with navbar — no blank white page
  2. All five Supabase tables have RLS enabled and a SELECT-only policy for the `anon` role — confirmed via Supabase dashboard
  3. `dashboard/config.js` exists locally with valid Supabase URL and anon key, is listed in `.gitignore`, and `config.example.js` is committed as a setup template
  4. Navigating to `#/books`, `#/ideas`, `#/calendar` renders the correct stub view without a page reload
  5. Tailwind CSS loads from CDN and utility classes render correctly on static elements
**Plans**: TBD

### Phase 2: Data Layer + Shared Components
**Goal**: Reusable fetch wrapper and pure component renderers are verified against mock data before any live view is built
**Depends on**: Phase 1
**Requirements**: NAV-01, NAV-02, NAV-03, UX-01, UX-02, UX-03, UX-04
**Success Criteria** (what must be TRUE):
  1. A fetch to a Supabase table returns live data and populates the page; a simulated network error shows a visible error banner (not an empty screen)
  2. Navigating between Books / Ideas / Calendar tabs highlights the active tab and renders the correct stub
  3. A spinner appears immediately when any view begins loading, and disappears when data arrives
  4. Empty columns and empty tables display a clear empty-state message rather than blank space
  5. Dark color palette is applied globally — background, text, and card colors are consistent
**Plans**: TBD

### Phase 3: Books View
**Goal**: Users can see all books in the catalog with live Supabase data and click ASINs to open Amazon pages
**Depends on**: Phase 2
**Requirements**: BOOKS-01, BOOKS-02
**Success Criteria** (what must be TRUE):
  1. The Books tab shows a table with columns: title, ASIN, genre, target audience — populated from live Supabase data
  2. Clicking an ASIN opens the Amazon product page in a new browser tab
  3. If the books table is empty, the empty-state message is displayed instead of a blank table
**Plans**: TBD

### Phase 4: Ideas Kanban
**Goal**: Users can see all content ideas organized by status in a four-column kanban, with script preview on demand
**Depends on**: Phase 3
**Requirements**: IDEAS-01, IDEAS-02, IDEAS-03
**Success Criteria** (what must be TRUE):
  1. The Ideas tab shows four columns (pending / approved / rejected / scripted) populated with live Supabase data
  2. Each card displays the idea title and platform (TikTok or Instagram)
  3. Clicking a card that has a generated script expands a preview of the script text inline
  4. Columns with no ideas show the empty-state message rather than blank space
**Plans**: TBD

### Phase 5: Calendar Kanban
**Goal**: Users can see the full publishing pipeline in a five-column kanban with platform color coding and a polished dark design
**Depends on**: Phase 4
**Requirements**: CAL-01, CAL-02, CAL-03
**Success Criteria** (what must be TRUE):
  1. The Calendar tab shows five columns (scheduled / approved / publishing / published / failed) with live Supabase data
  2. Each card displays the scheduled date, platform, and the title of the linked idea
  3. Cards are visually distinguishable by platform (different color accent for TikTok vs Instagram)
  4. The overall dashboard design is consistently minimal dark — all three views feel cohesive
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 0/TBD | Not started | - |
| 2. Data Layer + Shared Components | 0/TBD | Not started | - |
| 3. Books View | 0/TBD | Not started | - |
| 4. Ideas Kanban | 0/TBD | Not started | - |
| 5. Calendar Kanban | 0/TBD | Not started | - |
