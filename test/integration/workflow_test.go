package integration

import (
	"testing"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
)

func TestBookWorkflow_CreateAndRetrieve(t *testing.T) {
	SkipIfNoSupabase(t)

	// Create config
	cfg := &config.Config{
		Supabase: config.SupabaseConfig{
			URL:     GetTestSupabaseURL(),
			AnonKey: GetTestSupabaseKey(),
		},
	}

	// Create repository
	repo := repository.NewBooksRepository(&cfg.Supabase)

	// Create test book
	bookInput := &models.BookInput{
		Title:          "Integration Test Book",
		Genre:          "test",
		TargetAudience: "testers",
		KDPASIN:        "B0INTTEST",
	}

	// Test Create
	created, err := repo.Create(bookInput)
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	if created.ID == "" {
		t.Error("Created book has no ID")
	}

	if created.Title != bookInput.Title {
		t.Errorf("Expected title %q, got %q", bookInput.Title, created.Title)
	}

	// Test GetByID
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve book: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}

	if retrieved.Title != bookInput.Title {
		t.Errorf("Expected title %q, got %q", bookInput.Title, retrieved.Title)
	}

	// Cleanup note: Manual cleanup required
	t.Logf("Created test book with ID: %s (manual cleanup may be required)", created.ID)
}

func TestContentIdeaWorkflow_CreateAndValidate(t *testing.T) {
	SkipIfNoSupabase(t)

	cfg := &config.Config{
		Supabase: config.SupabaseConfig{
			URL:     GetTestSupabaseURL(),
			AnonKey: GetTestSupabaseKey(),
		},
	}

	repo := repository.NewContentRepository(&cfg.Supabase)

	// Create test idea
	score := 85
	ideaInput := &models.ContentIdeaInput{
		Type:             "educational",
		BriefDescription: "Test integration idea",
		RelevanceScore:   &score,
	}

	// Test validation
	err := ideaInput.Validate()
	if err != nil {
		t.Fatalf("Valid idea failed validation: %v", err)
	}

	// Test Create
	created, err := repo.CreateIdea(ideaInput)
	if err != nil {
		t.Fatalf("Failed to create idea: %v", err)
	}

	if created.ID == "" {
		t.Error("Created idea has no ID")
	}

	if created.Type != ideaInput.Type {
		t.Errorf("Expected type %q, got %q", ideaInput.Type, created.Type)
	}

	t.Logf("Created test idea with ID: %s", created.ID)
}

func TestKDPParserWorkflow_Integration(t *testing.T) {
	// This test doesn't require Supabase, just tests the parser logic

	// Test data simulating Amazon KDP report
	_ = `Title,ASIN,Date,Units Sold,Royalty,KENP Read
Test Book,B0ABC123,2024-01-15,10,$25.50,1500
Another Book,B0DEF456,2024-01-16,5,$12.75,800`

	// Parse the data
	// Note: This would use the actual parser
	// For now, just verify the test can run

	t.Log("KDP parser integration test structure ready")
}

func TestErrorHandling_RetryLogic(t *testing.T) {
	// Test error handling without external dependencies

	t.Log("Error handling integration test structure ready")
}
