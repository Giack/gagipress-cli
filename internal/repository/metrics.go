package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
)

// MetricsRepository handles metrics database operations
type MetricsRepository struct {
	config *config.SupabaseConfig
	client *http.Client
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(cfg *config.SupabaseConfig) *MetricsRepository {
	return &MetricsRepository{
		config: cfg,
		client: &http.Client{},
	}
}

// CreateMetric creates a new post metric
func (r *MetricsRepository) CreateMetric(input *models.PostMetricInput) (*models.PostMetric, error) {
	url := fmt.Sprintf("%s/rest/v1/post_metrics", r.config.URL)

	// Calculate engagement rate
	engagementRate := input.CalculateEngagementRate()

	data := map[string]interface{}{
		"calendar_id":     input.CalendarID,
		"platform":        input.Platform,
		"views":           input.Views,
		"likes":           input.Likes,
		"comments":        input.Comments,
		"shares":          input.Shares,
		"saves":           input.Saves,
		"engagement_rate": engagementRate,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create metric: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var metrics []models.PostMetric
	if err := json.Unmarshal(body, &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no metric returned from API")
	}

	return &metrics[0], nil
}

// GetMetrics retrieves metrics with optional filters
func (r *MetricsRepository) GetMetrics(platform string, from, to time.Time) ([]models.PostMetric, error) {
	url := fmt.Sprintf("%s/rest/v1/post_metrics?select=*&order=collected_at.desc", r.config.URL)

	if platform != "" {
		url += fmt.Sprintf("&platform=eq.%s", platform)
	}
	if !from.IsZero() {
		url += fmt.Sprintf("&collected_at=gte.%s", from.Format(time.RFC3339))
	}
	if !to.IsZero() {
		url += fmt.Sprintf("&collected_at=lte.%s", to.Format(time.RFC3339))
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get metrics: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var metrics []models.PostMetric
	if err := json.Unmarshal(body, &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return metrics, nil
}

// GetAggregateMetrics retrieves aggregated metrics for a period
func (r *MetricsRepository) GetAggregateMetrics(platform string, from, to time.Time) (*models.AggregateMetrics, error) {
	metrics, err := r.GetMetrics(platform, from, to)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return &models.AggregateMetrics{}, nil
	}

	agg := &models.AggregateMetrics{
		TotalPosts: len(metrics),
	}

	var totalEngagement float64
	var topEngagement float64
	var topPostID string

	for _, m := range metrics {
		agg.TotalViews += m.Views
		agg.TotalLikes += m.Likes
		agg.TotalComments += m.Comments
		agg.TotalShares += m.Shares
		totalEngagement += m.EngagementRate

		if m.EngagementRate > topEngagement {
			topEngagement = m.EngagementRate
			topPostID = m.CalendarID
		}
	}

	agg.AvgEngagement = totalEngagement / float64(len(metrics))
	agg.TopPost = topPostID
	agg.TopEngagement = topEngagement

	return agg, nil
}
