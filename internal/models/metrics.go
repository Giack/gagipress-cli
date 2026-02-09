package models

import (
	"time"
)

// PostMetric represents performance metrics for a published post
type PostMetric struct {
	ID             string    `json:"id"`
	CalendarID     string    `json:"calendar_id"`
	Platform       string    `json:"platform"`
	Views          int       `json:"views"`
	Likes          int       `json:"likes"`
	Comments       int       `json:"comments"`
	Shares         int       `json:"shares"`
	Saves          int       `json:"saves"`
	EngagementRate float64   `json:"engagement_rate"`
	CollectedAt    time.Time `json:"collected_at"`
}

// PostMetricInput represents input for creating a post metric
type PostMetricInput struct {
	CalendarID string  `json:"calendar_id"`
	Platform   string  `json:"platform"`
	Views      int     `json:"views"`
	Likes      int     `json:"likes"`
	Comments   int     `json:"comments"`
	Shares     int     `json:"shares"`
	Saves      int     `json:"saves"`
}

// Validate validates post metric input
func (p *PostMetricInput) Validate() error {
	if p.CalendarID == "" {
		return ErrInvalidInput{Field: "calendar_id", Message: "calendar ID is required"}
	}
	if p.Platform != "instagram" && p.Platform != "tiktok" {
		return ErrInvalidInput{Field: "platform", Message: "platform must be 'instagram' or 'tiktok'"}
	}
	return nil
}

// CalculateEngagementRate calculates engagement rate for the metric
func (p *PostMetricInput) CalculateEngagementRate() float64 {
	if p.Views == 0 {
		return 0.0
	}
	totalEngagement := p.Likes + p.Comments + p.Shares + p.Saves
	return float64(totalEngagement) / float64(p.Views) * 100.0
}

// AggregateMetrics represents aggregated metrics for a period
type AggregateMetrics struct {
	TotalPosts     int     `json:"total_posts"`
	TotalViews     int     `json:"total_views"`
	TotalLikes     int     `json:"total_likes"`
	TotalComments  int     `json:"total_comments"`
	TotalShares    int     `json:"total_shares"`
	AvgEngagement  float64 `json:"avg_engagement"`
	TopPost        string  `json:"top_post,omitempty"`
	TopEngagement  float64 `json:"top_engagement"`
}

// CorrelationPoint represents a data point for correlation analysis
type CorrelationPoint struct {
	Date        time.Time `json:"date"`
	Views       int       `json:"views"`
	Engagement  float64   `json:"engagement"`
	UnitsSold   int       `json:"units_sold"`
	Royalty     float64   `json:"royalty"`
}
