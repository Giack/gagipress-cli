# Gagipress CLI - Comprehensive Test Plan

**Version**: 1.0
**Date**: 2026-02-13
**Status**: Ready for Execution
**Coverage**: End-to-end workflows + Unit tests

---

## 📋 Overview

This test plan covers all functionality of the Gagipress CLI system, from basic setup to complete content generation and analytics workflows.

**Scope:**
- ✅ Unit tests (automated - already implemented)
- ✅ Integration tests (automated - framework ready)
- 📋 Manual E2E tests (documented below)
- 📋 Smoke tests (critical paths)

**Test Environment:**
- Local development machine
- Real Supabase instance (test project recommended)
- OpenAI API key (test account)
- Test data (sample books, ideas)

---

## 🧪 Test Execution Summary

### Automated Tests (Already Implemented)

**Run all tests:**
```bash
mise exec -- go test ./...
```

**Coverage report:**
```bash
./scripts/test-coverage.sh
```

**Current Status:**
- ✅ 50+ unit test cases
- ✅ All passing
- ✅ ~40% coverage

---

## 📦 Test Suites

### Suite 1: Setup & Configuration

#### Test 1.1: Initial Setup
**Objective**: Verify system can be initialized from scratch

**Prerequisites**: None

**Steps:**
1. Remove existing config: `rm -rf ~/.gagipress/`
2. Run `./bin/gagipress init`
3. Enter Supabase URL
4. Enter Supabase anon key
5. Enter OpenAI API key

**Expected Results:**
- ✅ Config file created at `~/.gagipress/config.yaml`
- ✅ No errors displayed
- ✅ Success message shown

**Pass Criteria**: Config file exists and contains credentials

---

#### Test 1.2: Database Migration
**Objective**: Verify database schema creation

**Prerequisites**: Test 1.1 passed

**Steps:**
1. Run `./bin/gagipress db migrate`
2. Check output for success messages
3. Run `./bin/gagipress db status`

**Expected Results:**
- ✅ Migration executes without errors
- ✅ All tables created (books, content_ideas, content_scripts, content_calendar, post_metrics, sales_data)
- ✅ Status shows "Connected" with schema version

**Pass Criteria**: db status shows connection and version 2

**SQL Verification** (optional):
```sql
SELECT version, description FROM schema_version ORDER BY version;
-- Should show version 1 and 2
```

---

### Suite 2: Book Management

#### Test 2.1: Add Book
**Objective**: Create a new book in the catalog

**Prerequisites**: Database migrated

**Steps:**
1. Run `./bin/gagipress books add`
2. Enter title: "Test Book for Children"
3. Enter genre: "children"
4. Enter target audience: "3-5 years"
5. Enter ASIN: "B0TEST123" (or leave empty)
6. Skip cover image URL (press Enter)

**Expected Results:**
- ✅ Book created successfully
- ✅ Book ID displayed
- ✅ Confirmation message shown

**Pass Criteria**: Book appears in `books list`

---

#### Test 2.2: List Books
**Objective**: View all books in catalog

**Prerequisites**: At least 1 book created

**Steps:**
1. Run `./bin/gagipress books list`

**Expected Results:**
- ✅ Table displayed with columns: ID, Title, Genre, Audience, ASIN
- ✅ All created books shown
- ✅ Formatted output (readable)

**Pass Criteria**: Previously created book appears in list

---

#### Test 2.3: Edit Book
**Objective**: Modify existing book details

**Prerequisites**: Test 2.1 passed

**Steps:**
1. Note book ID from `books list`
2. Run `./bin/gagipress books edit <book-id>`
3. Update title: "Updated Test Book"
4. Keep other fields (press Enter)

**Expected Results:**
- ✅ Book updated successfully
- ✅ Changes reflected in `books list`

**Pass Criteria**: Title changed in database

---

#### Test 2.4: Delete Book
**Objective**: Remove book from catalog

**Prerequisites**: Multiple books exist

**Steps:**
1. Create a temporary book
2. Run `./bin/gagipress books delete <book-id>`
3. Confirm deletion
4. Check `books list`

**Expected Results:**
- ✅ Book removed from list
- ✅ Cascade delete removes related content (if any)

**Pass Criteria**: Book no longer appears in list

---

### Suite 3: Content Generation - Ideas

#### Test 3.1: Generate Ideas (OpenAI)
**Objective**: Generate content ideas using OpenAI

**Prerequisites**:
- At least 1 book in catalog
- OpenAI API key configured

**Steps:**
1. Run `./bin/gagipress generate ideas --count 10`
2. Observe spinner animation
3. Wait for generation to complete

**Expected Results:**
- ✅ Spinner shows "Generating 10 ideas..."
- ✅ Success message: "Generated X ideas"
- ✅ Ideas saved to database
- ✅ Uses OpenAI (check console output)

**Pass Criteria**:
- Ideas created (check with `gagipress ideas list`)
- No errors during generation

**Validation:**
```bash
./bin/gagipress ideas list --status pending
# Should show generated ideas
```

---

#### Test 3.2: Generate Ideas (Gemini Fallback)
**Objective**: Verify Gemini fallback works

**Prerequisites**:
- Book in catalog
- Invalid OpenAI key (temporarily modify config)

**Steps:**
1. Edit `~/.gagipress/config.yaml` - set invalid OpenAI key
2. Run `./bin/gagipress generate ideas --count 5`
3. Observe fallback message

**Expected Results:**
- ✅ OpenAI fails
- ✅ Message: "Falling back to Gemini..."
- ✅ Gemini generates ideas
- ✅ Ideas saved successfully

**Pass Criteria**: Ideas created despite OpenAI failure

**Cleanup**: Restore valid OpenAI key

---

#### Test 3.3: Generate Ideas (Force Gemini)
**Objective**: Directly use Gemini with flag

**Prerequisites**: Book in catalog

**Steps:**
1. Run `./bin/gagipress generate ideas --count 3 --gemini`

**Expected Results:**
- ✅ Message: "Using Gemini for generation..."
- ✅ Browser automation runs (visible if headless=false)
- ✅ Ideas generated and saved

**Pass Criteria**: Ideas created using Gemini

---

#### Test 3.4: List Ideas (All Statuses)
**Objective**: View generated ideas with different filters

**Prerequisites**: Multiple ideas generated

**Steps:**
1. Run `./bin/gagipress ideas list`
2. Run `./bin/gagipress ideas list --status pending`
3. Run `./bin/gagipress ideas list --status approved`
4. Run `./bin/gagipress ideas list --limit 5`

**Expected Results:**
- ✅ All commands work without errors
- ✅ Filtering works correctly
- ✅ Limit restricts output
- ✅ Table formatted properly

**Pass Criteria**: Different filters show expected results

---

#### Test 3.5: Approve Idea
**Objective**: Approve an idea for script generation

**Prerequisites**: Pending ideas exist

**Steps:**
1. Get idea ID from `ideas list --status pending`
2. Run `./bin/gagipress ideas approve <idea-id>`
3. Verify with `ideas list --status approved`

**Expected Results:**
- ✅ Success message shown
- ✅ Idea status changed to "approved"
- ✅ Appears in approved list

**Pass Criteria**: Idea moves from pending to approved

---

#### Test 3.6: Reject Idea
**Objective**: Reject a poor quality idea

**Prerequisites**: Pending ideas exist

**Steps:**
1. Get idea ID from `ideas list --status pending`
2. Run `./bin/gagipress ideas reject <idea-id>`
3. Verify with `ideas list --status rejected`

**Expected Results:**
- ✅ Success message shown
- ✅ Idea status changed to "rejected"
- ✅ Won't be used for scripts

**Pass Criteria**: Idea moves to rejected status

---

### Suite 4: Content Generation - Scripts

#### Test 4.1: Generate Script from Idea
**Objective**: Create complete video script from approved idea

**Prerequisites**: At least 1 approved idea

**Steps:**
1. Get approved idea ID: `ideas list --status approved`
2. Run `./bin/gagipress generate script <idea-id>`
3. Observe spinner and output

**Expected Results:**
- ✅ Spinner shows "Generating script with AI..."
- ✅ Script generated with all sections:
  - Hook
  - Main Content
  - CTA
  - Hashtags
  - Music suggestion
  - Video notes
- ✅ Script saved to database
- ✅ Idea marked as "scripted"

**Pass Criteria**: Complete script displayed and saved

---

#### Test 4.2: Generate Script for Instagram
**Objective**: Generate platform-specific script

**Prerequisites**: Approved idea exists

**Steps:**
1. Run `./bin/gagipress generate script <idea-id> --platform instagram`

**Expected Results:**
- ✅ Script generated
- ✅ Platform-specific optimizations
- ✅ Appropriate hashtags for Instagram

**Pass Criteria**: Script mentions Instagram or has platform-specific content

---

#### Test 4.3: Generate Script with Gemini
**Objective**: Use Gemini for script generation

**Prerequisites**: Approved idea exists

**Steps:**
1. Run `./bin/gagipress generate script <idea-id> --gemini`

**Expected Results:**
- ✅ Uses Gemini for generation
- ✅ Complete script generated
- ✅ All required fields present

**Pass Criteria**: Script created using Gemini

---

### Suite 5: Calendar Planning & Approval

#### Test 5.1: Plan Weekly Calendar
**Objective**: Generate weekly posting schedule

**Prerequisites**: At least 7 scripts in database

**Steps:**
1. Run `./bin/gagipress calendar plan --days 7`
2. Review plan output

**Expected Results:**
- ✅ 7 days of posts scheduled
- ✅ Peak times used (7am, 12pm, 7pm, 9pm)
- ✅ Platform distribution (TikTok and Instagram)
- ✅ Content mix balanced
- ✅ Posts saved with status "pending_approval"

**Pass Criteria**: Calendar entries created with future dates

---

#### Test 5.2: Show Calendar
**Objective**: View planned content calendar

**Prerequisites**: Calendar planned

**Steps:**
1. Run `./bin/gagipress calendar show`

**Expected Results:**
- ✅ Table with scheduled posts
- ✅ Shows date, time, platform, status
- ✅ Properly formatted output

**Pass Criteria**: Planned posts displayed

---

#### Test 5.3: Approve Calendar Entries
**Objective**: Interactive approval workflow

**Prerequisites**: Pending calendar entries exist

**Steps:**
1. Run `./bin/gagipress calendar approve`
2. For first entry: choose "a" (approve)
3. For second entry: choose "s" (skip)
4. For third entry: choose "r" (reject)
5. Continue or quit

**Expected Results:**
- ✅ Interactive prompts for each entry
- ✅ Approved entries change status to "approved"
- ✅ Rejected entries removed or marked rejected
- ✅ Skipped entries remain pending
- ✅ Summary shown at end

**Pass Criteria**: Different actions work correctly

---

### Suite 6: Sales Data Import

#### Test 6.1: Import KDP Sales Data
**Objective**: Import Amazon KDP sales report

**Prerequisites**: Sample CSV file prepared

**Sample CSV** (`test-sales.csv`):
```csv
Title,ASIN,Date,Units Sold,Royalty
Test Book for Children,B0TEST123,2024-01-15,10,$25.50
Test Book for Children,B0TEST123,2024-01-16,5,$12.75
```

**Steps:**
1. Create sample CSV file
2. Run `./bin/gagipress books sales import test-sales.csv`
3. Check output

**Expected Results:**
- ✅ CSV parsed successfully
- ✅ Both rows imported
- ✅ ASIN matched to book (if exists)
- ✅ Success message with count

**Pass Criteria**: Sales data imported without errors

---

#### Test 6.2: Show Sales Data
**Objective**: View sales for a specific book

**Prerequisites**: Test 6.1 passed, book exists

**Steps:**
1. Get book ID from `books list`
2. Run `./bin/gagipress books sales show <book-id>`

**Expected Results:**
- ✅ Sales data displayed
- ✅ Shows date, units, royalty
- ✅ Formatted table

**Pass Criteria**: Imported sales visible

---

### Suite 7: Analytics & Correlation

#### Test 7.1: Add Post Metrics (Manual)
**Objective**: Manually add metrics for a published post

**Prerequisites**: Approved calendar entry exists

**Note**: Since OAuth publishing isn't implemented, manually insert metrics for testing

**SQL Insert** (via Supabase dashboard or psql):
```sql
INSERT INTO post_metrics (calendar_id, platform, views, likes, comments, shares, saves, engagement_rate, collected_at)
VALUES (
  '<calendar-id>',
  'tiktok',
  1000,
  150,
  20,
  10,
  5,
  18.5,
  NOW()
);
```

**Steps:**
1. Insert test metrics via SQL
2. Run `./bin/gagipress stats show`

**Expected Results:**
- ✅ Metrics displayed
- ✅ Aggregate stats calculated
- ✅ Top performer identified

**Pass Criteria**: Stats command shows metrics

---

#### Test 7.2: Show Analytics Dashboard
**Objective**: View performance analytics

**Prerequisites**: Post metrics exist

**Steps:**
1. Run `./bin/gagipress stats show`
2. Run `./bin/gagipress stats show --period 7d`
3. Run `./bin/gagipress stats show --platform tiktok`

**Expected Results:**
- ✅ Total posts, views, likes displayed
- ✅ Average engagement rate calculated
- ✅ Top performing post shown
- ✅ Platform breakdown
- ✅ Insights provided
- ✅ Filtering works

**Pass Criteria**: Dashboard displays correctly

---

#### Test 7.3: Analyze Correlation
**Objective**: Correlate social metrics with sales

**Prerequisites**:
- Post metrics exist
- Sales data exists for same date range

**Steps:**
1. Get book ID from `books list`
2. Run `./bin/gagipress stats correlate --book <book-id>`
3. Review correlation output

**Expected Results:**
- ✅ Pearson correlation calculated
- ✅ Correlation strength shown (strong/moderate/weak)
- ✅ Direction shown (positive/negative)
- ✅ Recommendations provided
- ✅ Data points count shown

**Pass Criteria**: Correlation analysis completes

**Expected Output Example:**
```
Correlation: 0.72 (strong positive)
Recommendation: Social media is driving sales...
```

---

### Suite 8: Error Handling & Resilience

#### Test 8.1: Retry Logic - Temporary API Failure
**Objective**: Verify retry mechanism works

**Prerequisites**: OpenAI API configured

**Simulation**:
- Temporarily set invalid API key
- Run generation
- Restore key quickly

**Expected Results:**
- ✅ First attempt fails
- ✅ Retry attempts occur (2-3 times)
- ✅ Eventually fails or succeeds
- ✅ Fallback to Gemini occurs

**Pass Criteria**: Retry behavior visible in console output

---

#### Test 8.2: Validation Error - No Retry
**Objective**: Verify validation errors fail fast

**Steps:**
1. Attempt to create book with empty title (if possible via API)
2. Or modify code temporarily to test

**Expected Results:**
- ✅ Validation error returned immediately
- ✅ No retry attempts
- ✅ Clear error message

**Pass Criteria**: Fails fast without retries

---

#### Test 8.3: Graceful Degradation
**Objective**: System continues working with one service down

**Steps:**
1. Disable OpenAI (invalid key)
2. Run idea generation
3. Verify Gemini takes over

**Expected Results:**
- ✅ OpenAI fails
- ✅ Automatic fallback to Gemini
- ✅ Ideas still generated
- ✅ No data loss

**Pass Criteria**: System remains functional

---

### Suite 9: UX & Progress Indicators

#### Test 9.1: Spinner Animation
**Objective**: Verify progress indicators work

**Steps:**
1. Run `./bin/gagipress generate ideas`
2. Observe spinner during generation
3. Check success/error messages

**Expected Results:**
- ✅ Spinner animates during long operations
- ✅ Spinner stops when complete
- ✅ Success (✓) or Error (✗) shown
- ✅ Messages clear and informative

**Pass Criteria**: Good UX during waits

---

### Suite 10: Data Integrity

#### Test 10.1: Duplicate Prevention
**Objective**: Ensure no duplicate data created

**Steps:**
1. Import same sales CSV twice
2. Check sales data

**Expected Results:**
- ✅ Second import detects duplicates
- ✅ No duplicate rows created
- ✅ Appropriate message shown

**Pass Criteria**: Data deduplication works

---

#### Test 10.2: Cascade Delete
**Objective**: Verify cascade deletes work

**Steps:**
1. Create book with ideas, scripts, calendar entries
2. Delete book
3. Check related data

**Expected Results:**
- ✅ Book deleted
- ✅ Related ideas deleted
- ✅ Related scripts deleted
- ✅ Related calendar entries deleted

**Pass Criteria**: Foreign key constraints enforced

---

## 🔥 Smoke Tests (Critical Paths)

Run these before every release to verify core functionality:

### Smoke Test 1: Basic Flow
```bash
# 1. Setup
./bin/gagipress init
./bin/gagipress db migrate

# 2. Add book
./bin/gagipress books add

# 3. Generate ideas
./bin/gagipress generate ideas --count 5

# 4. Approve idea
IDEA_ID=$(./bin/gagipress ideas list --status pending --limit 1 | head -1)
./bin/gagipress ideas approve $IDEA_ID

# 5. Generate script
./bin/gagipress generate script $IDEA_ID

# 6. Check calendar
./bin/gagipress calendar show
```

**Expected**: No errors, all commands complete successfully

---

### Smoke Test 2: Analytics Flow
```bash
# 1. Import sales
./bin/gagipress books sales import test-sales.csv

# 2. Show stats (may be empty if no metrics)
./bin/gagipress stats show

# 3. Correlate (if data available)
BOOK_ID=$(./bin/gagipress books list | head -2 | tail -1 | awk '{print $1}')
./bin/gagipress stats correlate --book $BOOK_ID
```

**Expected**: Commands run without errors

---

## 📊 Test Coverage Report

### Automated Tests Summary

| Package | Coverage | Test Cases | Status |
|---------|----------|------------|--------|
| models | 37.0% | 10 | ✅ Pass |
| parser | 96.9% | 16 | ✅ Pass |
| scheduler | 51.7% | 9 | ✅ Pass |
| errors | 82.9% | 11 | ✅ Pass |
| integration | N/A | 4 | ✅ Pass (2 skip) |
| **Total** | **~40%** | **50** | ✅ Pass |

---

## 🐛 Known Issues & Limitations

### Not Implemented (Out of Scope)
1. **OAuth Publishing** (Section 3.3)
   - Manual post metrics entry required
   - Future enhancement

2. **Automated Metrics Collection** (Section 4.1)
   - Manual SQL insert for testing
   - Future enhancement

3. **Weekly Report Generation** (Section 4.5)
   - Can be added later
   - Not critical for MVP

### Manual Testing Required
1. **Commands UI** - Interactive workflows
2. **OAuth flows** - When implemented
3. **Browser automation** - Gemini with headless=false

---

## ✅ Test Execution Checklist

**Before Testing:**
- [ ] Fresh Supabase project created (or reset)
- [ ] OpenAI API key available
- [ ] Test data prepared (sample CSV)
- [ ] Config backed up (if existing)

**During Testing:**
- [ ] Document any failures
- [ ] Note performance issues
- [ ] Capture error messages
- [ ] Test edge cases

**After Testing:**
- [ ] All smoke tests passed
- [ ] Critical paths verified
- [ ] Issues logged
- [ ] Test data cleaned up

---

## 📝 Test Results Template

```markdown
## Test Execution: [Date]

**Tester**: [Name]
**Environment**: [Local/Staging]
**Build**: [Commit hash]

### Results Summary
- Total tests: XX
- Passed: XX
- Failed: XX
- Skipped: XX

### Failed Tests
1. Test X.X - [Description]
   - Error: [Error message]
   - Steps to reproduce: [Steps]
   - Severity: [Critical/High/Medium/Low]

### Notes
- [Any observations]
- [Performance issues]
- [Suggestions]
```

---

## 🎯 Success Criteria

**Test Plan Complete When:**
- ✅ All automated tests passing
- ✅ Smoke tests execute without errors
- ✅ Critical paths verified manually
- ✅ No blocking bugs found
- ✅ Documentation updated

**System Ready for Production When:**
- ✅ This test plan executed successfully
- ✅ All critical/high bugs fixed
- ✅ OAuth setup documented (when implemented)
- ✅ User guide available

---

**Test Plan Status**: ✅ **READY FOR EXECUTION**

**Last Updated**: 2026-02-13
**Version**: 1.0
