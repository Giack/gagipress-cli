package stats

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	period   string
	platform string
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show performance analytics dashboard",
	Long: `Display performance analytics and key metrics.

Shows:
  - Total posts published
  - Total views, likes, comments, shares
  - Average engagement rate
  - Top performing posts
  - Platform breakdown`,
	RunE: runShow,
}

func init() {
	showCmd.Flags().StringVar(&period, "period", "30d", "Time period (7d, 30d, 90d, all)")
	showCmd.Flags().StringVar(&platform, "platform", "", "Filter by platform (instagram, tiktok)")
}

func runShow(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse period
	var from time.Time
	switch period {
	case "7d":
		from = time.Now().AddDate(0, 0, -7)
	case "30d":
		from = time.Now().AddDate(0, 0, -30)
	case "90d":
		from = time.Now().AddDate(0, 0, -90)
	case "all":
		from = time.Time{} // Zero time = no filter
	default:
		return fmt.Errorf("invalid period: %s (use 7d, 30d, 90d, or all)", period)
	}

	// Header
	header := "📊 Performance Analytics"
	if platform != "" {
		header += fmt.Sprintf(" (%s)", platform)
	}
	fmt.Println(ui.StyleHeader.Render(header))
	fmt.Printf("Period: %s\n\n", period)

	// Get metrics
	metricsRepo := repository.NewMetricsRepository(&cfg.Supabase)
	agg, err := metricsRepo.GetAggregateMetrics(platform, from, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	if agg.TotalPosts == 0 {
		fmt.Println("No metrics data available.")
		fmt.Println("\nAdd metrics manually or sync from social platforms.")
		return nil
	}

	// Create section style
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorMuted).
		Padding(1, 2).
		Width(ui.GetTerminalWidth() - 4)

	// Display aggregate metrics
	overviewContent := fmt.Sprintf(
		"Total Posts:          %d\n"+
			"Total Views:          %s\n"+
			"Total Likes:          %s\n"+
			"Total Comments:       %s\n"+
			"Total Shares:         %s\n"+
			"Avg Engagement Rate:  %.2f%%",
		agg.TotalPosts,
		ui.FormatNumber(agg.TotalViews),
		ui.FormatNumber(agg.TotalLikes),
		ui.FormatNumber(agg.TotalComments),
		ui.FormatNumber(agg.TotalShares),
		agg.AvgEngagement,
	)

	fmt.Println(ui.StyleHeader.Render("📈 Overview"))
	fmt.Println(sectionStyle.Render(overviewContent))
	fmt.Println()

	// Top performer
	if agg.TopPost != "" {
		topPostID := agg.TopPost
		if len(topPostID) > 8 {
			topPostID = topPostID[:8] + "…"
		}

		topContent := fmt.Sprintf(
			"Post ID:        %s\n"+
				"Engagement:     %.2f%%",
			topPostID,
			agg.TopEngagement,
		)

		fmt.Println(ui.StyleHeader.Render("🏆 Top Performer"))
		fmt.Println(sectionStyle.Render(topContent))
		fmt.Println()
	}

	// Platform breakdown
	if platform == "" {
		// Get TikTok metrics
		tiktokAgg, _ := metricsRepo.GetAggregateMetrics("tiktok", from, time.Now())

		// Get Instagram metrics
		igAgg, _ := metricsRepo.GetAggregateMetrics("instagram", from, time.Now())

		platformContent := fmt.Sprintf(
			"TikTok:    %d posts | %.2f%% avg engagement\n"+
				"Instagram: %d posts | %.2f%% avg engagement",
			tiktokAgg.TotalPosts, tiktokAgg.AvgEngagement,
			igAgg.TotalPosts, igAgg.AvgEngagement,
		)

		fmt.Println(ui.StyleHeader.Render("📱 Platform Breakdown"))
		fmt.Println(sectionStyle.Render(platformContent))
		fmt.Println()
	}

	// Insights
	avgViewsPerPost := 0
	if agg.TotalPosts > 0 {
		avgViewsPerPost = agg.TotalViews / agg.TotalPosts
	}

	var engagementInsight string
	if agg.AvgEngagement > 5.0 {
		engagementInsight = ui.StyleSuccess.Render("✅ Excellent engagement rate (>5%)")
	} else if agg.AvgEngagement > 3.0 {
		engagementInsight = ui.StyleSuccess.Render("✅ Good engagement rate (3-5%)")
	} else if agg.AvgEngagement > 1.0 {
		engagementInsight = ui.StyleWarning.Render("⚠️  Moderate engagement rate (1-3%)")
	} else {
		engagementInsight = ui.StyleWarning.Render("⚠️  Low engagement rate (<1%)")
	}

	insightsContent := fmt.Sprintf(
		"• Average views per post: %s\n"+
			"• %s",
		ui.FormatNumber(avgViewsPerPost),
		engagementInsight,
	)

	fmt.Println(ui.StyleHeader.Render("💡 Insights"))
	fmt.Println(sectionStyle.Render(insightsContent))

	return nil
}
