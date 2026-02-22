# Campaign Strategy Guide

## Goal: Increase Amazon KDP Book Sales

All decisions should be evaluated against one question: **does this likely increase book sales?**

---

## Interpreting `stats correlate` Output

The command returns a **Pearson correlation coefficient** between social post frequency/engagement and daily book sales.

| Coefficient | Meaning | Action |
|---|---|---|
| > 0.6 | Strong positive correlation | Double down: more posts, same content style |
| 0.3–0.6 | Moderate correlation | Optimize: test different angles, better CTAs |
| 0–0.3 | Weak/no correlation | Investigate: wrong audience, weak CTA, no ASIN in scripts |
| < 0 | Negative correlation | Stop current approach, pivot content strategy |

**Always check**:
1. Do generated scripts contain the Amazon link? (`amazon.it/dp/<ASIN>?tag=gagipress-21&utm_source=...`)
2. Is the ASIN set on the book? (`bin/gagipress books list`)
3. Have sales been imported recently? (`bin/gagipress stats import`)

---

## Platform Strategy

### TikTok
- **Best for**: Discovery, viral reach, new audiences
- **Script length**: 3min max, hook in first 3 seconds
- **CTA placement**: Mid-video + end of video
- **Posting frequency**: 1–2x daily for growth phase
- **Optimal times**: 7–9 AM, 12–2 PM, 7–9 PM (local audience timezone)

### Instagram Reels
- **Best for**: Engaged audience, higher purchase intent
- **Script length**: 60 seconds max
- **CTA placement**: End of video + caption
- **Posting frequency**: 1x daily
- **Optimal times**: 9–11 AM, 8–10 PM

---

## Content Angle Optimization

When correlation is weak, test different content angles. Prioritize ideas that:

1. **Show the product result** ("My daughter learned to color in 3 days with this book")
2. **Address the purchase decision** ("Why this coloring book is different from the rest")
3. **Social proof** ("200+ reviews on Amazon — here's what parents say")
4. **Behind the scenes** ("How I created the illustrations for this book")
5. **Educational** ("5 benefits of coloring books for toddlers")

Avoid angles that are generic (no product tie-in) or purely entertainment (no purchase intent signal).

---

## Weekly Rhythm

| Day | Action |
|---|---|
| Monday | Review `calendar status`, retry failures, check `stats show` |
| Tuesday | `generate ideas` if pipeline is thin (<5 approved) |
| Wednesday | Approve ideas, `generate batch` scripts |
| Thursday | `calendar plan`, review and approve entries |
| Friday | `stats import` latest KDP CSV, run `stats correlate` |
| Weekend | Automated cron handles publishing; monitor only |

---

## Diagnosing Poor Performance

**No sales despite posts?**
1. Check scripts have Amazon links: read a recent script from DB
2. Check ASIN is set: `bin/gagipress books list`
3. Check posts actually published: `bin/gagipress calendar show --status published`
4. Import fresh KDP data: `bin/gagipress stats import`

**Calendar all failed?**
1. `bin/gagipress calendar status`
2. `bin/gagipress calendar retry`
3. Check Blotato API key: `supabase secrets list`

**Low engagement metrics?**
1. Review recent scripts for hook strength (first 3 seconds)
2. Switch platforms: if TikTok weak, try Instagram batch
3. Generate new ideas with different angles
4. Try `--with-media` flag for visual posts

---

## Strategy Recommendations Template

When reporting strategy suggestions, structure them as:

```
📊 Current State: [X posts published, Y days of data, correlation: Z]
🎯 What's Working: [specific content angles or platforms]
⚠️ What's Not: [weak areas with evidence]
🚀 Next 3 Actions:
  1. [Specific action with exact command]
  2. [Specific action with exact command]
  3. [Specific action with exact command]
```
