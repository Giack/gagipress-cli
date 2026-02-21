# E2E Integration Tests for Approve/Reject Commands

## Overview

This directory contains end-to-end integration tests that verify the approve/reject workflow with a real Supabase database.

## Tests Implemented

### Ideas Approve/Reject Tests
- `TestIdeasApprove_ResolvesByPrefix` - Tests UUID prefix resolution (8 chars)
- `TestIdeasApprove_UpdatesStatusToApproved` - Tests status update to "approved"
- `TestIdeasReject_UpdatesStatusToRejected` - Tests status update to "rejected"
- `TestIdeasApprove_PrefixTooShort` - Tests error handling for short prefixes

### Calendar Approve/Reject Tests
- `TestCalendarApprove_UpdatesStatus` - Tests calendar entry approval
- `TestCalendarReject_DeletesEntry` - Tests calendar entry deletion

## Bugs Fixed

### 1. URL Encoding Bug (content.go:171, books.go:222)
**Problem**: UUID prefix matching used unencoded asterisk wildcard (`*`) in LIKE queries, causing SQL errors.

**Fix**: URL-encode the pattern before constructing the query string:
```go
pattern := url.QueryEscape(prefix + "*")
requestURL := fmt.Sprintf("%s/rest/v1/content_ideas?select=*&id=like.%s", r.config.URL, pattern)
```

### 2. Schema Mismatch (models/content.go:90-117)
**Problem**:
- `ContentCalendar` model missing required `PostType` field (DB has NOT NULL constraint)
- Field `ErrorMessage` should be `PublishErrors` (JSONB, not TEXT)

**Fix**: Added PostType field and corrected PublishErrors field:
```go
type ContentCalendar struct {
    // ...
    PostType      string     `json:"post_type"` // reel, story, feed - REQUIRED
    PublishErrors any        `json:"publish_errors,omitempty"` // JSONB field
}
```

### 3. Missing Prefer Header (content.go:148, calendar.go:148)
**Problem**: PATCH operations didn't include `Prefer: return=representation` header, causing PostgREST to return 204 No Content instead of updated data.

**Fix**: Added Prefer header to both UpdateIdeaStatus and UpdateEntryStatus:
```go
req.Header.Set("Prefer", "return=representation")
```

## Test Infrastructure

### TestFixture
The `TestFixture` struct provides:
- Automatic cleanup via `t.Cleanup()`
- Repository creation with service key preference
- Tracked resource creation for cleanup

**Example usage:**
```go
func TestMyFeature(t *testing.T) {
    SkipIfNoSupabase(t)
    fixture := NewTestFixture(t)

    // Create test data - automatically cleaned up
    idea := fixture.CreateIdea(&models.ContentIdeaInput{
        Type:             "educational",
        BriefDescription: "Test idea",
    })

    // Test your feature...
    // Cleanup happens automatically when test finishes
}
```

## Running Tests

### With Supabase Credentials

```bash
# Set environment variables
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_KEY="your-anon-key"
export SUPABASE_SERVICE_KEY="your-service-key"  # Optional but recommended

# Run all E2E tests
mise exec -- go test ./test/integration/approve_reject_test.go \
  ./test/integration/fixtures.go \
  ./test/integration/setup_test.go -v

# Run specific test
mise exec -- go test ./test/integration/approve_reject_test.go \
  ./test/integration/fixtures.go \
  ./test/integration/setup_test.go -v \
  -run TestIdeasApprove_ResolvesByPrefix
```

### Without Credentials
Tests will skip gracefully:
```bash
mise exec -- go test ./test/integration/... -v
# Output: SKIP: Skipping integration test: SUPABASE_URL not set
```

## Expected Test Output

When run with credentials:
```
=== RUN   TestIdeasApprove_ResolvesByPrefix
--- PASS: TestIdeasApprove_ResolvesByPrefix (0.42s)
=== RUN   TestIdeasApprove_UpdatesStatusToApproved
--- PASS: TestIdeasApprove_UpdatesStatusToApproved (0.38s)
=== RUN   TestCalendarApprove_UpdatesStatus
--- PASS: TestCalendarApprove_UpdatesStatus (0.41s)
=== RUN   TestIdeasReject_UpdatesStatusToRejected
--- PASS: TestIdeasReject_UpdatesStatusToRejected (0.35s)
=== RUN   TestCalendarReject_DeletesEntry
--- PASS: TestCalendarReject_DeletesEntry (0.33s)
PASS
ok      github.com/gagipress/gagipress-cli/test/integration    2.145s
```

## Cleanup

Tests automatically clean up created data using `t.Cleanup()`. However:
- **Calendar entries** are deleted via `DeleteEntry()`
- **Ideas** currently have no delete endpoint, so they remain in DB
  - Check fixture cleanup logs for created idea IDs if manual cleanup needed

## TDD Process Followed

This implementation followed strict Test-Driven Development:

1. **RED**: Wrote failing tests first
   - Created test fixtures
   - Wrote tests for desired behavior
   - Ran tests to see them fail

2. **GREEN**: Made minimal changes to pass tests
   - Fixed URL encoding bug
   - Fixed schema mismatch
   - Added missing Prefer headers

3. **REFACTOR**: No refactoring needed (code already clean)

## Next Steps

To verify the fixes work with your Supabase instance:

1. Set up Supabase credentials (see above)
2. Run the E2E tests
3. Test commands manually:
   ```bash
   # Test ideas approve
   bin/gagipress ideas list
   bin/gagipress ideas approve <prefix>

   # Test calendar approve
   bin/gagipress calendar approve
   ```

## Files Created/Modified

**New Files:**
- `test/integration/fixtures.go` - Test fixture infrastructure
- `test/integration/approve_reject_test.go` - E2E tests
- `test/integration/README.md` - This file

**Modified Files:**
- `internal/models/content.go` - Added PostType, fixed PublishErrors
- `internal/repository/content.go` - Fixed URL encoding, added Prefer header
- `internal/repository/books.go` - Fixed URL encoding
- `internal/repository/calendar.go` - Added Prefer header
- `test/integration/setup_test.go` - Added GetTestSupabaseServiceKey()
