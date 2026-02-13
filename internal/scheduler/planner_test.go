package scheduler

import (
	"testing"

	"github.com/gagipress/gagipress-cli/internal/models"
)

func TestPlanner_BalanceContentMix(t *testing.T) {
	planner := &Planner{
		optimizer: NewOptimizer(),
	}

	scripts := []models.ContentScript{
		{ID: "script-3"},
		{ID: "script-1"},
		{ID: "script-2"},
	}

	balanced := planner.BalanceContentMix(scripts)

	// Should return same number of scripts
	if len(balanced) != len(scripts) {
		t.Errorf("Expected %d scripts, got %d", len(scripts), len(balanced))
	}

	// Should be sorted by ID (deterministic ordering)
	if balanced[0].ID != "script-1" {
		t.Errorf("Expected first script ID 'script-1', got '%s'", balanced[0].ID)
	}
	if balanced[1].ID != "script-2" {
		t.Errorf("Expected second script ID 'script-2', got '%s'", balanced[1].ID)
	}
	if balanced[2].ID != "script-3" {
		t.Errorf("Expected third script ID 'script-3', got '%s'", balanced[2].ID)
	}
}

func TestTimeSlot_Structure(t *testing.T) {
	// Test that TimeSlot has expected fields
	slot := TimeSlot{
		Platform: "tiktok",
		Type:     "scheduled",
	}

	if slot.Platform != "tiktok" {
		t.Errorf("Expected platform 'tiktok', got '%s'", slot.Platform)
	}

	if slot.Type != "scheduled" {
		t.Errorf("Expected type 'scheduled', got '%s'", slot.Type)
	}
}
