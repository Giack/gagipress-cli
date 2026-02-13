package scheduler

import (
	"testing"
	"time"
)

func TestOptimizer_GetOptimalTimes(t *testing.T) {
	optimizer := NewOptimizer()

	tests := []struct {
		name        string
		days        int
		postsPerDay int
		wantCount   int
	}{
		{
			name:        "single day with 2 posts",
			days:        1,
			postsPerDay: 2,
			wantCount:   2,
		},
		{
			name:        "week with 1 post per day",
			days:        7,
			postsPerDay: 1,
			wantCount:   7,
		},
		{
			name:        "week with 2 posts per day",
			days:        7,
			postsPerDay: 2,
			wantCount:   14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slots := optimizer.GetOptimalTimes(tt.days, tt.postsPerDay)

			if len(slots) != tt.wantCount {
				t.Errorf("GetOptimalTimes() returned %d slots, want %d", len(slots), tt.wantCount)
			}
		})
	}
}

func TestOptimizer_GetOptimalTimes_PeakHours(t *testing.T) {
	optimizer := NewOptimizer()
	peakHours := []int{7, 12, 19, 21}

	slots := optimizer.GetOptimalTimes(1, 4)

	if len(slots) != 4 {
		t.Fatalf("Expected 4 slots, got %d", len(slots))
	}

	// Check that we're using peak hours
	for i, slot := range slots {
		expectedHour := peakHours[i%len(peakHours)]
		if slot.Time.Hour() != expectedHour {
			t.Errorf("Slot %d: expected hour %d, got %d", i, expectedHour, slot.Time.Hour())
		}
	}
}

func TestOptimizer_GetOptimalTimes_FutureDates(t *testing.T) {
	optimizer := NewOptimizer()
	now := time.Now()

	slots := optimizer.GetOptimalTimes(3, 1)

	// All slots should be in the future
	for i, slot := range slots {
		if slot.Time.Before(now) {
			t.Errorf("Slot %d is in the past: %v", i, slot.Time)
		}
	}
}

func TestOptimizer_GetOptimalTimes_PlatformDistribution(t *testing.T) {
	optimizer := NewOptimizer()

	slots := optimizer.GetOptimalTimes(7, 2)

	// Should have both TikTok and Instagram
	hasTikTok := false
	hasInstagram := false

	for _, slot := range slots {
		if slot.Platform == "tiktok" {
			hasTikTok = true
		}
		if slot.Platform == "instagram" {
			hasInstagram = true
		}
	}

	if !hasTikTok {
		t.Error("Expected at least one TikTok slot")
	}

	if !hasInstagram {
		t.Error("Expected at least one Instagram slot")
	}
}

func TestOptimizer_GetPeakTimes(t *testing.T) {
	optimizer := NewOptimizer()

	tests := []struct {
		name     string
		platform string
		count    int
	}{
		{"get 4 peak times", "tiktok", 4},
		{"get 2 peak times", "instagram", 2},
		{"get 8 peak times", "tiktok", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			times := optimizer.GetPeakTimes(tt.platform, tt.count)

			if len(times) != tt.count {
				t.Errorf("GetPeakTimes() returned %d times, want %d", len(times), tt.count)
			}
		})
	}
}

func TestDefaultMixStrategy(t *testing.T) {
	strategy := DefaultMixStrategy()

	// All percentages should be positive
	if strategy.Educational <= 0 {
		t.Error("Educational percentage should be positive")
	}
	if strategy.Entertainment <= 0 {
		t.Error("Entertainment percentage should be positive")
	}

	// Total should be approximately 1.0 (100%)
	total := strategy.Educational + strategy.Entertainment + strategy.BTS + strategy.UGC + strategy.Trend

	if total < 0.99 || total > 1.01 {
		t.Errorf("Total strategy percentage should be ~1.0, got %v", total)
	}
}

func TestOptimizer_AnalyzeHistoricalData(t *testing.T) {
	optimizer := NewOptimizer()

	metrics := []MetricPoint{
		{Hour: 7, DayOfWeek: time.Monday, EngagementRate: 5.2},
		{Hour: 12, DayOfWeek: time.Monday, EngagementRate: 4.8},
		{Hour: 19, DayOfWeek: time.Monday, EngagementRate: 6.1},
	}

	// Should not panic
	optimizer.AnalyzeHistoricalData("tiktok", metrics)

	// Data should be stored
	if len(optimizer.historicalData["tiktok"]) != 3 {
		t.Errorf("Expected 3 data points stored, got %d", len(optimizer.historicalData["tiktok"]))
	}
}
