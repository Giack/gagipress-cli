package integration

import (
	"testing"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
)

// TestFixture provides test data creation and automatic cleanup
type TestFixture struct {
	t               *testing.T
	contentRepo     *repository.ContentRepository
	calendarRepo    *repository.CalendarRepository
	booksRepo       *repository.BooksRepository
	createdIdeaIDs  []string
	createdEntryIDs []string
	createdBookIDs  []string
}

// NewTestFixture creates a fixture with automatic cleanup on test completion
func NewTestFixture(t *testing.T) *TestFixture {
	cfg := &config.Config{
		Supabase: config.SupabaseConfig{
			URL:        GetTestSupabaseURL(),
			AnonKey:    GetTestSupabaseKey(),
			ServiceKey: GetTestSupabaseServiceKey(), // Prefer service key for tests
		},
	}

	fixture := &TestFixture{
		t:            t,
		contentRepo:  repository.NewContentRepository(&cfg.Supabase),
		calendarRepo: repository.NewCalendarRepository(&cfg.Supabase),
		booksRepo:    repository.NewBooksRepository(&cfg.Supabase),
	}

	// Register cleanup function
	t.Cleanup(func() {
		fixture.Cleanup()
	})

	return fixture
}

// CreateIdea creates a test idea and tracks it for cleanup
func (f *TestFixture) CreateIdea(input *models.ContentIdeaInput) *models.ContentIdea {
	idea, err := f.contentRepo.CreateIdea(input)
	if err != nil {
		f.t.Fatalf("failed to create test idea: %v", err)
	}
	f.createdIdeaIDs = append(f.createdIdeaIDs, idea.ID)
	return idea
}

// CreateCalendarEntry creates a test calendar entry and tracks it for cleanup
func (f *TestFixture) CreateCalendarEntry(input *models.ContentCalendarInput) *models.ContentCalendar {
	entry, err := f.calendarRepo.CreateEntry(input)
	if err != nil {
		f.t.Fatalf("failed to create test calendar entry: %v", err)
	}
	f.createdEntryIDs = append(f.createdEntryIDs, entry.ID)
	return entry
}

// Cleanup deletes all test data created by this fixture
func (f *TestFixture) Cleanup() {
	// Clean up calendar entries
	for _, id := range f.createdEntryIDs {
		_ = f.calendarRepo.DeleteEntry(id) // Ignore errors during cleanup
	}

	// Clean up ideas (no delete method exists, so leave for manual cleanup)
	// Note: Could add DeleteIdea method to repository if needed
	if len(f.createdIdeaIDs) > 0 {
		f.t.Logf("Created %d test ideas that may require manual cleanup: %v",
			len(f.createdIdeaIDs), f.createdIdeaIDs)
	}
}
