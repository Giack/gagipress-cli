---
name: gagipress-campaign
description: >
  Execute the complete gagipress marketing campaign loop for Amazon KDP book publishers.
  Covers the full lifecycle: content idea generation → idea approval → script writing →
  calendar planning → publishing → performance analytics → strategy improvement.
  Goal is always to increase Amazon book sales.

  Use this skill when the user asks to:
  - "Run the marketing campaign" or "esegui la campagna"
  - "Generate content for my book" or any step in the content pipeline
  - "Check campaign performance" / "come sta andando la campagna"
  - "Publish posts" / "publish content" / "pubblica i post"
  - "Improve my TikTok/Instagram strategy"
  - "Increase book sales" / "aumenta le vendite"
  - "What should I do next for marketing?"
  - Run any gagipress workflow involving ideas, scripts, calendar, publish, or stats
---

# Gagipress Campaign Skill

**Goal**: Increase Amazon KDP book sales by executing and optimizing the social media marketing loop.

**References** (load as needed):
- Full CLI syntax → `references/commands.md`
- Strategy interpretation & optimization → `references/strategy.md`

---

## Step 1: Assess Pipeline State

Always start here. Run these commands and report what you find:

```bash
bin/gagipress books list
bin/gagipress ideas list --status pending
bin/gagipress ideas list --status approved
bin/gagipress calendar status
bin/gagipress stats show --period 7d
```

Read `references/strategy.md` to interpret the stats output.

---

## Step 2: Decide What Phase to Run

Based on the state assessment, pick the right phase:

| State | Action |
|---|---|
| No books or missing ASIN | Fix first: `books edit <id> --asin B0XXXXXXXX` |
| Fewer than 5 approved ideas | → **Phase A: Generate Ideas** |
| Approved ideas but no scripts | → **Phase B: Generate Scripts** |
| Scripts exist but no calendar | → **Phase C: Plan Calendar** |
| Calendar entries not approved | → **Phase D: Approve & Publish** |
| Posts published, have sales data | → **Phase E: Analyze & Optimize** |
| Calendar has failures | → **Phase F: Retry Failures** |

Run all phases needed in sequence. Don't stop at one phase if the pipeline is stale end-to-end.

---

## Phase A: Generate & Approve Ideas

```bash
bin/gagipress generate ideas --book <book-id>
bin/gagipress ideas list --status pending
bin/gagipress ideas approve <idea-id>   # repeat for each good idea
```

**Approval criteria** (read `references/strategy.md` → Content Angle Optimization):
- Prefer ideas showing product result, purchase decision, or social proof
- Target 8–12 approved ideas per batch before moving to scripts

---

## Phase B: Generate Scripts

```bash
bin/gagipress generate batch --platform tiktok --limit 10
bin/gagipress generate batch --platform instagram --limit 5
```

Verify scripts contain Amazon URLs with UTM params. If ASIN is missing from the book, add it first.

---

## Phase C: Plan Calendar

```bash
bin/gagipress calendar plan
bin/gagipress calendar show --status scheduled
```

Review the schedule. Confirm posting frequency matches the platform strategy in `references/strategy.md`.

---

## Phase D: Approve & Publish

```bash
bin/gagipress calendar approve

# Test with a single post first
bin/gagipress publish <calendar-entry-id>

# If test succeeds, batch publish (or let cron auto-publish)
bin/gagipress publish batch --limit 5
```

The pg_cron Edge Function auto-publishes every 15 minutes for `approved` entries with `scheduled_for <= NOW()`. Manual publish is only needed for immediate testing.

---

## Phase E: Analyze & Optimize

```bash
bin/gagipress stats import <path/to/sales.csv>
bin/gagipress stats show --period 30d
bin/gagipress stats correlate --book <book-id> --days 30
```

Read `references/strategy.md` → Interpreting `stats correlate` Output to evaluate the Pearson coefficient and recommend next actions.

**Always present a structured recommendation** using the template in `references/strategy.md` → Strategy Recommendations Template.

---

## Phase F: Retry Failures

```bash
bin/gagipress calendar status
bin/gagipress calendar retry
```

If retries keep failing, check Blotato API key (`supabase secrets list`) and platform connectivity.

---

## Continuous Improvement Loop

After each full cycle, evaluate:
1. Which content angles correlated with sales spikes?
2. Which platform performed better (TikTok vs Instagram)?
3. Are scripts including clear Amazon CTAs with UTM links?
4. Should idea generation focus on different angles next batch?

Use `references/strategy.md` → Weekly Rhythm as a cadence guide between sessions.
