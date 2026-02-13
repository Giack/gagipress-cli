package models

import (
	"testing"
)

func TestPostMetricInput_CalculateEngagementRate(t *testing.T) {
	tests := []struct {
		name     string
		input    PostMetricInput
		expected float64
	}{
		{
			name: "normal engagement",
			input: PostMetricInput{
				Views:    1000,
				Likes:    100,
				Comments: 20,
				Shares:   10,
				Saves:    5,
			},
			expected: 13.5, // (100+20+10+5) / 1000 * 100
		},
		{
			name: "zero views",
			input: PostMetricInput{
				Views:    0,
				Likes:    10,
				Comments: 5,
			},
			expected: 0.0,
		},
		{
			name: "high engagement",
			input: PostMetricInput{
				Views:    100,
				Likes:    50,
				Comments: 30,
				Shares:   10,
			},
			expected: 90.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.CalculateEngagementRate()
			if result != tt.expected {
				t.Errorf("CalculateEngagementRate() = %v, want %v", result, tt.expected)
			}
		})
	}
}
