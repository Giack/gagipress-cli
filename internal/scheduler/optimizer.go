package scheduler

import (
	"time"
)

// Optimizer handles posting time optimization
type Optimizer struct {
	historicalData map[string][]MetricPoint
}

// MetricPoint represents historical performance data
type MetricPoint struct {
	Hour           int
	DayOfWeek      time.Weekday
	EngagementRate float64
}

// NewOptimizer creates a new posting time optimizer
func NewOptimizer() *Optimizer {
	return &Optimizer{
		historicalData: make(map[string][]MetricPoint),
	}
}

// GetOptimalTimes returns optimal posting times for a period
func (o *Optimizer) GetOptimalTimes(days int, postsPerDay int) []TimeSlot {
	var slots []TimeSlot

	// Start from next day at midnight
	startDate := time.Now().AddDate(0, 0, 1)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	// Default peak times for TikTok/Instagram Reels
	// Based on industry research:
	// - TikTok: 6-10am, 7-11pm
	// - Instagram: 11am-2pm, 7-9pm
	peakHours := []int{7, 12, 19, 21}

	for day := 0; day < days; day++ {
		currentDate := startDate.AddDate(0, 0, day)

		// Distribute posts throughout the day using peak hours
		for i := 0; i < postsPerDay; i++ {
			hourIndex := i % len(peakHours)
			hour := peakHours[hourIndex]

			// Add some variation to avoid exact same time every day
			minuteVariation := (day * 7) % 60 // 0-60 minutes variation

			postTime := time.Date(
				currentDate.Year(),
				currentDate.Month(),
				currentDate.Day(),
				hour,
				minuteVariation,
				0,
				0,
				currentDate.Location(),
			)

			// Determine platform based on time
			platform := "tiktok"
			if hour >= 11 && hour <= 14 {
				platform = "instagram" // Lunch time better for Instagram
			}

			slots = append(slots, TimeSlot{
				Time:     postTime,
				Platform: platform,
				Type:     "scheduled",
			})
		}
	}

	return slots
}

// AnalyzeHistoricalData analyzes past performance to optimize times
func (o *Optimizer) AnalyzeHistoricalData(platform string, metrics []MetricPoint) {
	o.historicalData[platform] = metrics

	// In a real implementation, this would:
	// 1. Group metrics by hour and day of week
	// 2. Calculate average engagement for each time slot
	// 3. Identify top-performing times
	// 4. Return personalized peak times
}

// GetPeakTimes returns the best times based on historical data
func (o *Optimizer) GetPeakTimes(platform string, count int) []time.Time {
	// If we have historical data, use it
	if data, ok := o.historicalData[platform]; ok && len(data) > 0 {
		// Sort by engagement rate
		// Return top N times
		_ = data // Use historical data in real implementation
	}

	// Default peak times
	var times []time.Time
	peakHours := []int{7, 12, 19, 21}

	now := time.Now()
	for i := 0; i < count; i++ {
		hour := peakHours[i%len(peakHours)]
		t := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
		times = append(times, t)
	}

	return times
}

// ContentMixStrategy determines how to balance different content types
type ContentMixStrategy struct {
	Educational   float64 // 0.0 - 1.0
	Entertainment float64
	BTS           float64
	UGC           float64
	Trend         float64
}

// DefaultMixStrategy returns a balanced content mix
func DefaultMixStrategy() ContentMixStrategy {
	return ContentMixStrategy{
		Educational:   0.25,
		Entertainment: 0.25,
		BTS:           0.15,
		UGC:           0.20,
		Trend:         0.15,
	}
}
