package scheduler

import (
	"fmt"
	"sort"
	"time"

	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
)

// Planner handles content calendar planning
type Planner struct {
	contentRepo *repository.ContentRepository
	optimizer   *Optimizer
}

// NewPlanner creates a new calendar planner
func NewPlanner(contentRepo *repository.ContentRepository) *Planner {
	return &Planner{
		contentRepo: contentRepo,
		optimizer:   NewOptimizer(),
	}
}

// PlanWeek creates a weekly content plan
func (p *Planner) PlanWeek(days int, postsPerDay int) ([]*models.ContentCalendarInput, error) {
	// Get available scripts (from scripted ideas)
	scripts, err := p.contentRepo.GetScripts(0)
	if err != nil {
		return nil, fmt.Errorf("failed to get scripts: %w", err)
	}

	if len(scripts) == 0 {
		return nil, fmt.Errorf("no scripts available for planning")
	}

	totalPosts := days * postsPerDay
	if len(scripts) < totalPosts {
		return nil, fmt.Errorf("not enough scripts: need %d, have %d", totalPosts, len(scripts))
	}

	// Group scripts by type for balanced distribution
	scriptsByType := make(map[string][]models.ContentScript)
	for _, script := range scripts {
		// We need to get the idea to know the type
		// For now, we'll just use the scripts directly
		scriptsByType["mixed"] = append(scriptsByType["mixed"], script)
	}

	// Get optimal posting times
	postingTimes := p.optimizer.GetOptimalTimes(days, postsPerDay)

	// Create calendar entries
	var calendar []*models.ContentCalendarInput
	scriptIndex := 0

	for _, slot := range postingTimes {
		if scriptIndex >= len(scripts) {
			break
		}

		script := scripts[scriptIndex]

		// Determine platform based on script characteristics
		platform := "tiktok"
		if script.EstimatedDuration > 60 {
			platform = "instagram" // Longer content for Instagram
		}

		entry := &models.ContentCalendarInput{
			ScriptID:     &script.ID,
			ScheduledFor: slot.Time,
			Platform:     platform,
			PostType:     "reel",
		}

		calendar = append(calendar, entry)
		scriptIndex++

		if len(calendar) >= totalPosts {
			break
		}
	}

	return calendar, nil
}

// TimeSlot represents a scheduled time slot
type TimeSlot struct {
	Time     time.Time
	Platform string
	Type     string
}

// BalanceContentMix ensures diverse content types in the calendar
func (p *Planner) BalanceContentMix(scripts []models.ContentScript) []models.ContentScript {
	// Sort by ID for now (deterministic)
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].ID < scripts[j].ID
	})

	// In a real implementation, this would:
	// 1. Group by content type (educational, entertainment, etc.)
	// 2. Ensure balanced distribution
	// 3. Avoid same type back-to-back
	// 4. Rotate between books

	return scripts
}
