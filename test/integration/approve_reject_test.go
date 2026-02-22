package integration

import (
	"testing"
	"time"

	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/testutil"
)

func TestIdeasApprove_ResolvesByPrefix(t *testing.T) {
	SkipIfNoSupabase(t)

	fixture := NewTestFixture(t)

	// Arrange: Create a pending idea
	idea := fixture.CreateIdea(&models.ContentIdeaInput{
		Type:             "educational",
		BriefDescription: "Test idea for approval",
	})

	// Act: Resolve by 8-character prefix
	prefix := idea.ID[:8]
	resolved, err := fixture.contentRepo.GetIdeaByIDPrefix(prefix)

	// Assert: Should find the idea
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, idea.ID, resolved.ID)
}

func TestIdeasApprove_UpdatesStatusToApproved(t *testing.T) {
	SkipIfNoSupabase(t)

	fixture := NewTestFixture(t)

	// Arrange
	idea := fixture.CreateIdea(&models.ContentIdeaInput{
		Type:             "entertainment",
		BriefDescription: "Test idea",
	})

	// Act: Approve the idea
	err := fixture.contentRepo.UpdateIdeaStatus(idea.ID, "approved")
	testutil.AssertNoError(t, err)

	// Assert: Verify status changed by re-fetching
	updated, err := fixture.contentRepo.GetIdeaByIDPrefix(idea.ID)
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, "approved", updated.Status)
}

func TestCalendarApprove_UpdatesStatus(t *testing.T) {
	SkipIfNoSupabase(t)

	fixture := NewTestFixture(t)

	// Arrange: Create calendar entry
	scheduledTime := time.Now().Add(24 * time.Hour)
	entry := fixture.CreateCalendarEntry(&models.ContentCalendarInput{
		ScheduledFor: scheduledTime,
		Platform:     "instagram",
		PostType:     "reel", // Required field
	})

	// Act: Approve the entry
	err := fixture.calendarRepo.UpdateEntryStatus(entry.ID, "approved")
	testutil.AssertNoError(t, err)

	// Assert: Verify status
	entries, err := fixture.calendarRepo.GetEntries("approved", 10)
	testutil.AssertNoError(t, err)

	found := false
	for _, e := range entries {
		if e.ID == entry.ID {
			found = true
			testutil.AssertEqual(t, "approved", e.Status)
		}
	}
	testutil.AssertEqual(t, true, found)
}

func TestIdeasApprove_PrefixTooShort(t *testing.T) {
	SkipIfNoSupabase(t)
	fixture := NewTestFixture(t)

	_, err := fixture.contentRepo.GetIdeaByIDPrefix("abc")
	testutil.AssertError(t, err)
	// Error message should mention minimum length
}

func TestIdeasReject_UpdatesStatusToRejected(t *testing.T) {
	SkipIfNoSupabase(t)
	fixture := NewTestFixture(t)

	idea := fixture.CreateIdea(&models.ContentIdeaInput{
		Type:             "trend",
		BriefDescription: "Test rejection",
	})

	err := fixture.contentRepo.UpdateIdeaStatus(idea.ID, "rejected")
	testutil.AssertNoError(t, err)

	updated, err := fixture.contentRepo.GetIdeaByIDPrefix(idea.ID[:8])
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, "rejected", updated.Status)
}

func TestCalendarReject_DeletesEntry(t *testing.T) {
	SkipIfNoSupabase(t)
	fixture := NewTestFixture(t)

	entry := fixture.CreateCalendarEntry(&models.ContentCalendarInput{
		ScheduledFor: time.Now().Add(48 * time.Hour),
		Platform:     "tiktok",
		PostType:     "reel",
	})

	err := fixture.calendarRepo.DeleteEntry(entry.ID)
	testutil.AssertNoError(t, err)

	// Verify deleted
	entries, _ := fixture.calendarRepo.GetEntries("pending_approval", 100)
	for _, e := range entries {
		if e.ID == entry.ID {
			t.Fatal("Entry should have been deleted")
		}
	}
}
