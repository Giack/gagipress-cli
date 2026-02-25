# Codebase Concerns

**Analysis Date:** 2026-02-25

## Tech Debt

### Stub Implementations in Social API Clients

**Files:** `internal/social/instagram.go`, `internal/social/tiktok.go`

**Issue:** Instagram and TikTok clients are placeholder stubs returning "not implemented" errors. All methods (`PublishPost`, `GetPostMetrics`, `GetRecentPosts`, `TestConnection`) are unimplemented with TODO markers.

**Impact:** Direct native publishing to Instagram/TikTok is not functional. All publishing currently routes through Blotato proxy API. If Blotato becomes unavailable or deprecated, there is no fallback for direct platform publishing.

**Fix approach:**
- Implement OAuth flows for Instagram Graph API and TikTok Creator API
- Add proper token refresh mechanisms
- Create integration tests for direct platform authentication
- Keep Blotato as optional fallback, not primary path

### Brittle Gemini Browser Automation

**File:** `internal/ai/gemini.go`

**Issue:** Uses chromedp to automate Gemini's web interface with hardcoded DOM selectors:
- `textarea[placeholder*="Enter a prompt"]` (line 51, 55, 59)
- `div[data-test-id="conversation-turn-2"]` (line 63, 66)

**Impact:** Any DOM changes to Gemini's interface will break this integration. This is used as fallback when OpenAI fails, so failures cascade. No monitoring for selector breakage.

**Workaround:** `--gemini` flag allows testing, `--headless=false` for debugging.

**Fix approach:**
- Consider Gemini API instead of browser automation (if available)
- Add visual regression testing for selector stability
- Implement fallback selectors or XPath alternatives
- Add logging when selectors fail to help quickly identify changes

## Known Bugs

### Unsafe Slice Indexing in generate_media.go

**File:** `cmd/calendar/generate_media.go:166`

**Symptoms:** Panic if entry.ID is shorter than 8 characters: `entry.ID[:8]` without bounds checking.

**Code:**
```go
fmt.Printf("Generating image for entry %s (%s) book: %q... ", entry.ID[:8], entry.Platform, bookTitle)
```

**Trigger:** Generate media for an entry with ID shorter than 8 bytes (though UUIDs are 36 chars, this is defensive).

**Workaround:** None - always use full ID or verify length first.

**Fix:** Change to safe substring: `entry.ID` or use `min(len(entry.ID), 8)`.

### Error Suppression in Photo Upload

**File:** `cmd/calendar/generate_media.go:269`

**Symptoms:** Upload response body read with `_, _ := io.ReadAll(uploadResp.Body)` ignores errors silently.

**Issue:** If body read fails, error is lost; storage error detection continues with empty body string.

**Code:**
```go
uploadBody, _ := io.ReadAll(uploadResp.Body)
```

**Fix approach:** Explicitly handle read errors or use ReadAll inline with error checking.

### Silent Fallback Errors in fetchCoverImage

**File:** `cmd/calendar/generate_media.go:37, 41, 45`

**Symptoms:** HTTP errors downloading book covers return `nil, "", nil` (no error) instead of propagating them.

**Issue:**
- Line 37: Network error → silent nil return
- Line 41: Non-200 status → silent nil return
- Line 45: Body read error → silent nil return

**Impact:** Users won't know if cover image download failed; they'll just get text-only generation.

**Code:**
```go
resp, err := http.Get(url) //nolint:gosec
if err != nil {
    return nil, "", nil // non-fatal: fall back to text-only generation
}
```

**Fix approach:** Log warnings for failures; return error to caller for visibility, let them decide fallback.

### Unverified Error in HTTP Request Construction

**File:** `cmd/calendar/generate_media.go:243`

**Symptoms:** DELETE request error ignored: `delReq, _ := http.NewRequestWithContext(...)`

**Code:**
```go
delReq, _ := http.NewRequestWithContext(ctx, "DELETE", uploadURL, nil)
```

**Impact:** Malformed URL would cause silent nil dereference when calling `Do(delReq)`.

## Security Considerations

### Hardcoded Amazon Image URL Pattern

**File:** `cmd/calendar/generate_media.go:27-30`

**Risk:** URL constructed from user-provided ASIN without validation. Malicious ASIN could trigger SSRF if URL scheme isn't verified.

**Current mitigation:** Only HTTP/HTTPS allowed by http.Get() standard library; Amazon domain is hardcoded.

**Recommendations:**
- Validate ASIN format (format: B + 9 alphanumerics)
- Add URL allowlist validation before fetch
- Log all external URL fetches for audit trail

### Environment Variable in Command Execution

**File:** `cmd/calendar/generate_media.go:141-143`

**Risk:** Gemini API key set via `os.Setenv("GOOGLE_API_KEY", apiKey)` leaks to child processes and can appear in ps output during execution.

**Code:**
```go
if err := os.Setenv("GOOGLE_API_KEY", apiKey); err != nil {
    return fmt.Errorf("failed to set GOOGLE_API_KEY: %w", err)
}
```

**Current mitigation:** Google SDK reads this at client creation time; env var could be removed after.

**Recommendations:**
- Remove env var after client creation: `defer os.Unsetenv("GOOGLE_API_KEY")`
- Pass API key directly to SDK if API supports it
- Document that process environment may briefly contain API key

### Direct HTTP Client Without Timeouts

**File:** `cmd/calendar/generate_media.go:151`

**Risk:** `&http.Client{}` created without timeout for Supabase Storage uploads. Long-hung requests block indefinitely.

**Code:**
```go
httpClient := &http.Client{}
```

**Current mitigation:** Global context timeout from `context.Background()` is used in requests.

**Recommendations:**
- Set explicit timeout: `&http.Client{Timeout: 30 * time.Second}`
- Document timeout expectations for large media files
- Add per-request context timeout for sensitive operations

### Suppressed Gosec Warning

**File:** `cmd/calendar/generate_media.go:35`

**Risk:** `http.Get(url)` with `//nolint:gosec` suppresses URL validation check.

**Code:**
```go
resp, err := http.Get(url) //nolint:gosec
```

**Current mitigation:** URL comes from database (cover_image_url) or constructed from validated ASIN.

**Recommendations:**
- Remove nolint and validate URL scheme is http/https
- Add URL sanitization function to prevent SSRF
- Document why URL is safe in comment

## Performance Bottlenecks

### Unoptimized Image Generation Loop

**File:** `cmd/calendar/generate_media.go:155-288`

**Problem:** Sequential Imagen API calls for each entry; no batching or parallelization.

**Issue:**
- Single threaded; waits for each image to complete before next
- Each image generation takes 5+ seconds
- 10 images = 50+ seconds minimum

**Cause:** Gemini Python SDK doesn't support concurrent requests from single client.

**Improvement path:**
- Create multiple genai clients for concurrent generation
- Implement worker pool pattern with configurable concurrency
- Add --concurrency flag to control parallelism
- Monitor API rate limits and backoff

### Polling with Fixed 5-Second Interval

**File:** `internal/social/blotato.go:230`

**Problem:** `WaitForVisualCreation` polls every 5 seconds with 60 attempts (5 minute timeout).

**Issue:** Most creations finish in <30 seconds; still waits 5s between checks.

**Cause:** Conservative timeout for worst-case slow generations.

**Improvement path:**
- Start with 1-2s interval, increase exponentially
- Set adaptive timeout based on status (e.g., 30s for "generating-media")
- Cache last-checked status to reduce redundant polls

### Missing Index on Calendar Queries

**File:** `internal/repository/calendar.go` (not read, but inferred from schema)

**Problem:** `GetEntriesNeedingMedia()` query filters by `generate_media=true` and `media_url IS NULL`; likely missing composite index.

**Issue:** Table scans on large calendars (1000+ entries).

**Improvement path:**
- Add index: `CREATE INDEX idx_calendar_media_gen ON content_calendar(generate_media, media_url)`
- Monitor query performance after adding entries
- Profile other frequently-filtered queries

## Fragile Areas

### Nested Optional Chaining in Media Generation

**File:** `cmd/calendar/generate_media.go:156-164`

**Files:** `internal/models/calendar.go` (structure not fully read)

**Why fragile:** Multiple levels of optional dereferencing:
```go
if entry.Script != nil && entry.Script.Idea != nil {
    book = entry.Script.Idea.Book
}
if book != nil {
    bookTitle = book.Title
}
```

**Safe modification:**
- Create helper function `getBookFromEntry(entry) *Book`
- Add nil checks at each level
- Log warnings if data is missing in unexpected places
- Add test fixtures covering null scenarios

**Test coverage:** Likely no tests for missing Script or Idea chains.

### Error Handling in Batch Operations

**File:** `cmd/calendar/generate_media.go:155-288`

**Why fragile:** Loop continues on individual failures, counting them at end. If 9/10 fail, user may not notice.

**Safe modification:**
- Add --fail-fast flag to stop on first error
- Report failures immediately, not silently in counter
- Collect failure reasons for diagnosis
- Add --max-failures threshold to bail out early

## Scaling Limits

### Sequential Publishing via Blotato Edge Function

**Files:** `supabase/functions/publish-scheduled/index.ts` (not read)

**Current capacity:** pg_cron runs every 15 minutes, processes one entry per run.

**Limit:** Max 4 posts per hour if all scheduled for different times. With 10+ scheduled posts, backlog builds.

**Scaling path:**
- Batch multiple posts per cron execution
- Use job queue (Redis/Postgres-based)
- Implement async webhooks for immediate publishing
- Add configurable batch size with rate limiting

### Supabase Storage Bandwidth

**File:** `cmd/calendar/generate_media.go:239-278`

**Current capacity:** Each generated image uploads once, fetched by browser on click.

**Limit:** Public bucket can hit bandwidth limits with viral content (lots of shares = lots of views = lots of downloads).

**Scaling path:**
- Configure CDN in front of Supabase Storage
- Implement image resizing/optimization before upload
- Add cache control headers to storage objects
- Monitor bandwidth usage and set alerts

## Dependencies at Risk

### chromedp Browser Automation

**Package:** `github.com/chromedp/chromedp` (used in `internal/ai/gemini.go`)

**Risk:** Depends on Chromium binary availability. GitHub Actions may not have it; local dev requires installation.

**Impact:** Gemini fallback becomes unusable without Chrome/Chromium on system.

**Migration plan:**
- Provide Docker container for CI/CD that includes Chromium
- Document Chromium installation requirements in README
- Consider Puppeteer/Playwright alternatives that auto-download browsers

### Google GenAI SDK

**Package:** `google.golang.org/genai`

**Risk:** This SDK is experimental/new. API surface may change between versions.

**Impact:** Breaking changes could require code refactor.

**Migration plan:**
- Pin to specific version in go.mod
- Monitor Google's SDK releases for deprecations
- Have fallback to REST API if SDK becomes unmaintained

### Blotato Dependency for Publishing

**Integration:** `internal/social/blotato.go`

**Risk:** Third-party service; no SLA documented. Downtime blocks all publishing.

**Impact:** 100% dependent on Blotato for Instagram/TikTok publishing. No fallback to native APIs.

**Migration plan:**
- Implement native Instagram Graph API client (follow Phase 2 plan)
- Keep Blotato as optional secondary path
- Add health check endpoint to monitor Blotato availability

## Test Coverage Gaps

### No Tests for Media Generation

**What's not tested:**
- `cmd/calendar/generate_media.go` entire file
- Image upload to Supabase Storage
- Gemini API fallback logic
- Cover image fetching from Amazon

**Files:** `cmd/calendar/generate_media.go`

**Risk:** Silent failures in critical path. Image generation failures won't be caught until production.

**Priority:** High (affects all media posts)

**Suggested tests:**
- Mock genai.Client for image generation
- Mock http.Client for Amazon cover image fetch
- Test upload URL construction and headers
- Test fallback from Gemini to Imagen when unavailable

### No Tests for Blotato Integration

**What's not tested:**
- Visual creation polling loop
- Timeout handling in `WaitForVisualCreation`
- Error responses from Blotato API
- Publishing post with/without media

**Files:** `internal/social/blotato.go`

**Risk:** Polling bugs or API changes go unnoticed.

**Priority:** High (critical publishing path)

**Suggested tests:**
- Mock Blotato API with different status sequences
- Test timeout boundary (60 attempts * 5s)
- Test error recovery in polling
- Test empty media URLs handling

### Incomplete Instagram/TikTok Client Tests

**What's not tested:**
- Any method in `internal/social/instagram.go` or `tiktok.go`
- These are stubs, but structure exists

**Files:** `internal/social/instagram.go`, `internal/social/tiktok.go`

**Risk:** When OAuth is implemented, no regression tests exist.

**Priority:** Medium (future work, but plan now)

**Suggested tests:**
- Mock OAuth flow
- Test token refresh
- Test post publishing
- Test metrics retrieval

### Gemini Browser Automation Not Tested

**What's not tested:**
- `internal/ai/gemini.go` browser automation
- Selector resilience
- Timeout handling
- Response parsing

**Files:** `internal/ai/gemini.go`

**Risk:** Silently fails when Gemini DOM changes.

**Priority:** High (fallback path could fail silently)

**Suggested tests:**
- Integration test with real Gemini (or mock browser)
- Test selector failures gracefully
- Test timeout behavior
- Test response extraction

## Missing Critical Features

### No Input Validation for ASIN

**Problem:** Book ASIN accepted without format validation.

**Blocks:** Cross-platform content linking; malformed ASINs cause broken Amazon links.

**Implementation:** Add validation in `internal/models/book.go`:
```go
func ValidateASIN(asin string) error {
    if !regexp.MustCompile(`^B[A-Z0-9]{9}$`).MatchString(asin) {
        return fmt.Errorf("invalid ASIN format: %s", asin)
    }
    return nil
}
```

### No Retry Logic for Failed Image Uploads

**Problem:** If Supabase Storage upload fails, media generation fails silently.

**Blocks:** Resilient media pipeline; single network blip aborts whole batch.

**Implementation:** Add exponential backoff retry in `cmd/calendar/generate_media.go`:
- 3 attempts with 1s, 2s, 4s backoff
- Different retry logic for 5xx vs 4xx errors

### No Dry-Run for Publishing

**Problem:** `calendar generate-media --dry-run` exists, but `publish` has no dry-run option.

**Blocks:** Safe testing of publishing flow; users must submit real posts to test.

**Implementation:** Add `--dry-run` flag to `cmd/publish/publish.go` to log without submitting.

---

*Concerns audit: 2026-02-25*
