package stats

import (
	"fmt"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
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

	fmt.Println("📊 Performance Analytics")
	fmt.Println("════════════════════════")

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

	fmt.Printf("Period: %s", period)
	if platform != "" {
		fmt.Printf(" | Platform: %s", platform)
	}
	fmt.Println()

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

	// Display aggregate metrics
	fmt.Println("📈 Overview")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Total Posts:          %d\n", agg.TotalPosts)
	fmt.Printf("Total Views:          %s\n", formatNumber(agg.TotalViews))
	fmt.Printf("Total Likes:          %s\n", formatNumber(agg.TotalLikes))
	fmt.Printf("Total Comments:       %s\n", formatNumber(agg.TotalComments))
	fmt.Printf("Total Shares:         %s\n", formatNumber(agg.TotalShares))
	fmt.Printf("Avg Engagement Rate:  %.2f%%\n", agg.AvgEngagement)
	fmt.Println()

	// Top performer
	if agg.TopPost != "" {
		fmt.Println("🏆 Top Performer")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("Post ID:        %s\n", agg.TopPost[:8])
		fmt.Printf("Engagement:     %.2f%%\n", agg.TopEngagement)
		fmt.Println()
	}

	// Platform breakdown
	if platform == "" {
		fmt.Println("📱 Platform Breakdown")
		fmt.Println(strings.Repeat("─", 60))

		// Get TikTok metrics
		tiktokAgg, _ := metricsRepo.GetAggregateMetrics("tiktok", from, time.Now())
		fmt.Printf("TikTok:    %d posts | %.2f%% avg engagement\n",
			tiktokAgg.TotalPosts, tiktokAgg.AvgEngagement)

		// Get Instagram metrics
		igAgg, _ := metricsRepo.GetAggregateMetrics("instagram", from, time.Now())
		fmt.Printf("Instagram: %d posts | %.2f%% avg engagement\n",
			igAgg.TotalPosts, igAgg.AvgEngagement)
		fmt.Println()
	}

	// Insights
	fmt.Println("💡 Insights")
	fmt.Println(strings.Repeat("─", 60))

	avgViewsPerPost := 0
	if agg.TotalPosts > 0 {
		avgViewsPerPost = agg.TotalViews / agg.TotalPosts
	}
	fmt.Printf("• Average views per post: %s\n", formatNumber(avgViewsPerPost))

	if agg.AvgEngagement > 5.0 {
		fmt.Println("• ✅ Excellent engagement rate (>5%)")
	} else if agg.AvgEngagement > 3.0 {
		fmt.Println("• ✅ Good engagement rate (3-5%)")
	} else if agg.AvgEngagement > 1.0 {
		fmt.Println("• ⚠️  Moderate engagement rate (1-3%)")
	} else {
		fmt.Println("• ⚠️  Low engagement rate (<1%)")
	}

	return nil
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
