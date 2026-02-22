# Gagipress CLI Command Reference

## Pipeline State Commands (run first to assess)

```bash
bin/gagipress books list
bin/gagipress ideas list --status pending
bin/gagipress ideas list --status approved
bin/gagipress calendar status          # shows counts: scheduled/approved/published/failed
bin/gagipress calendar show            # full list
bin/gagipress calendar show --status failed
bin/gagipress stats show               # engagement dashboard
bin/gagipress stats show --period 7d
bin/gagipress stats show --platform tiktok
```

## Content Generation

```bash
# Generate 20+ ideas for a book
bin/gagipress generate ideas --book <book-id>

# Approve / reject individual ideas
bin/gagipress ideas approve <idea-id>
bin/gagipress ideas reject <idea-id>

# Generate script for one approved idea
bin/gagipress generate script <idea-id> --platform tiktok
bin/gagipress generate script <idea-id> --platform instagram

# Batch generate scripts for all approved ideas (recommended)
bin/gagipress generate batch --platform tiktok --limit 10
bin/gagipress generate batch --platform instagram --limit 10
# Add --gemini flag to use free Gemini provider instead of OpenAI
```

## Calendar & Publishing

```bash
# Create weekly publishing schedule
bin/gagipress calendar plan

# Review and approve the scheduled entries
bin/gagipress calendar show --status scheduled
bin/gagipress calendar approve

# Manual publish (test first)
bin/gagipress publish <calendar-entry-id>
bin/gagipress publish <calendar-entry-id> --with-media   # generates Blotato visual

# Batch publish
bin/gagipress publish batch --limit 5
bin/gagipress publish batch --limit 5 --with-media

# Monitor automated cron publishing (every 15min via Edge Function)
bin/gagipress calendar status
bin/gagipress calendar retry            # reset failed → approved
```

## Analytics & Correlation

```bash
# Import Amazon KDP sales CSV (download from KDP Reports)
bin/gagipress stats import <path/to/sales.csv>

# Performance dashboard
bin/gagipress stats show
bin/gagipress stats show --period 30d --platform tiktok

# Correlation: social posts ↔ book sales (Pearson coefficient)
bin/gagipress stats correlate --book <book-id> --days 30
bin/gagipress stats correlate --book <book-id> --days 7
```

## Book Management

```bash
bin/gagipress books list
bin/gagipress books add --title "..." --genre "..." --asin B0XXXXXXXX
bin/gagipress books edit <book-id> --asin B0XXXXXXXX
```

## Cron Publishing Activation

The Edge Function `supabase/functions/publish-scheduled` runs every 15min automatically.
To activate it:
```bash
supabase secrets set BLOTATO_API_KEY=<key>
```
Once active, any calendar entry with `status=approved` and `scheduled_for <= NOW()` is published automatically.
