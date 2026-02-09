package stats

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
)

var (
	bookID string
	days   int
)

var correlateCmd = &cobra.Command{
	Use:   "correlate",
	Short: "Analyze social media → sales correlation",
	Long: `Analyze the correlation between social media performance and book sales.

Shows:
  - Daily social metrics vs sales
  - Correlation coefficient (Pearson's r)
  - Impact analysis
  - Recommendations

Helps identify if social media activity drives book sales.`,
	RunE: runCorrelate,
}

func init() {
	correlateCmd.Flags().StringVar(&bookID, "book", "", "Book ID (required)")
	correlateCmd.Flags().IntVar(&days, "days", 30, "Days to analyze")
	correlateCmd.MarkFlagRequired("book")
}

func runCorrelate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("🔗 Social → Sales Correlation Analysis")
	fmt.Println("═══════════════════════════════════════\n")

	// Get book info
	booksRepo := repository.NewBooksRepository(&cfg.Supabase)
	book, err := booksRepo.GetByID(bookID)
	if err != nil {
		return fmt.Errorf("failed to get book: %w", err)
	}

	fmt.Printf("Book: %s\n", book.Title)
	fmt.Printf("Period: Last %d days\n\n", days)

	// Get date range
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	// Get sales data
	salesRepo := repository.NewSalesRepository(&cfg.Supabase)
	sales, err := salesRepo.GetSalesByBook(bookID, from, to)
	if err != nil {
		return fmt.Errorf("failed to get sales: %w", err)
	}

	if len(sales) == 0 {
		fmt.Println("⚠️  No sales data available for this period.")
		fmt.Println("\nImport sales data with: gagipress books sales import <csv>")
		return nil
	}

	// Get metrics data
	metricsRepo := repository.NewMetricsRepository(&cfg.Supabase)
	metrics, err := metricsRepo.GetMetrics("", from, to)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	if len(metrics) == 0 {
		fmt.Println("⚠️  No social metrics available for this period.")
		fmt.Println("\nNote: Correlation requires both sales and social metrics data.")
		return nil
	}

	// Aggregate by day
	dailyData := aggregateByDay(sales, metrics)

	if len(dailyData) < 3 {
		fmt.Println("⚠️  Insufficient data points for correlation analysis.")
		fmt.Printf("   Found: %d days | Need: at least 3 days\n", len(dailyData))
		return nil
	}

	// Calculate correlation
	correlation := calculateCorrelation(dailyData)

	// Display results
	fmt.Println("📊 Correlation Results")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Data points:          %d days\n", len(dailyData))
	fmt.Printf("Total views:          %s\n", formatNumber(sumViews(dailyData)))
	fmt.Printf("Total sales:          %d units\n", sumSales(dailyData))
	fmt.Printf("Correlation (r):      %.3f\n", correlation)
	fmt.Println()

	// Interpret correlation
	fmt.Println("📈 Interpretation")
	fmt.Println(strings.Repeat("─", 60))

	absCorr := math.Abs(correlation)
	var strength string
	var interpretation string

	if absCorr >= 0.7 {
		strength = "Strong"
		interpretation = "Social media activity has a strong relationship with sales"
	} else if absCorr >= 0.4 {
		strength = "Moderate"
		interpretation = "Social media activity has a moderate relationship with sales"
	} else if absCorr >= 0.2 {
		strength = "Weak"
		interpretation = "Social media activity has a weak relationship with sales"
	} else {
		strength = "Very Weak / None"
		interpretation = "Little to no relationship detected between social media and sales"
	}

	if correlation > 0 {
		fmt.Printf("Strength:   %s positive correlation\n", strength)
		fmt.Printf("Direction:  📈 Positive (more views → more sales)\n")
	} else {
		fmt.Printf("Strength:   %s negative correlation\n", strength)
		fmt.Printf("Direction:  📉 Negative (more views → fewer sales)\n")
	}
	fmt.Printf("\n%s\n", interpretation)
	fmt.Println()

	// Recommendations
	fmt.Println("💡 Recommendations")
	fmt.Println(strings.Repeat("─", 60))

	if correlation > 0.4 {
		fmt.Println("✅ Social media is driving sales! Keep posting consistently.")
		fmt.Println("   • Focus on content types with highest engagement")
		fmt.Println("   • Increase posting frequency during peak days")
		fmt.Println("   • Use top-performing content as templates")
	} else if correlation > 0.1 {
		fmt.Println("⚠️  Moderate impact. Consider:")
		fmt.Println("   • Stronger CTAs in your posts")
		fmt.Println("   • More direct product mentions")
		fmt.Println("   • Test different content types")
		fmt.Println("   • Add Amazon link in bio")
	} else {
		fmt.Println("ℹ️  Low correlation detected. Possible reasons:")
		fmt.Println("   • Need more data (try 60-90 days)")
		fmt.Println("   • Delayed conversion effect")
		fmt.Println("   • Content not targeted enough")
		fmt.Println("   • Audience mismatch")
	}

	return nil
}

// aggregateByDay groups data by day
func aggregateByDay(sales []models.BookSale, metrics []models.PostMetric) []models.CorrelationPoint {
	dataMap := make(map[string]*models.CorrelationPoint)

	// Add sales
	for _, sale := range sales {
		dateKey := sale.SaleDate.Format("2006-01-02")
		if _, ok := dataMap[dateKey]; !ok {
			dataMap[dateKey] = &models.CorrelationPoint{
				Date: sale.SaleDate,
			}
		}
		dataMap[dateKey].UnitsSold += sale.UnitsSold
		dataMap[dateKey].Royalty += sale.Royalty
	}

	// Add metrics
	for _, metric := range metrics {
		dateKey := metric.CollectedAt.Format("2006-01-02")
		if _, ok := dataMap[dateKey]; !ok {
			dataMap[dateKey] = &models.CorrelationPoint{
				Date: metric.CollectedAt,
			}
		}
		dataMap[dateKey].Views += metric.Views
		dataMap[dateKey].Engagement += metric.EngagementRate
	}

	// Convert to slice
	var points []models.CorrelationPoint
	for _, point := range dataMap {
		// Only include days with both metrics and sales
		if point.Views > 0 && point.UnitsSold > 0 {
			points = append(points, *point)
		}
	}

	return points
}

// calculateCorrelation calculates Pearson correlation coefficient
func calculateCorrelation(data []models.CorrelationPoint) float64 {
	if len(data) < 2 {
		return 0.0
	}

	n := float64(len(data))

	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for _, point := range data {
		x := float64(point.Views)
		y := float64(point.UnitsSold)

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	numerator := (n * sumXY) - (sumX * sumY)
	denominator := math.Sqrt(((n * sumX2) - (sumX * sumX)) * ((n * sumY2) - (sumY * sumY)))

	if denominator == 0 {
		return 0.0
	}

	return numerator / denominator
}

func sumViews(data []models.CorrelationPoint) int {
	total := 0
	for _, p := range data {
		total += p.Views
	}
	return total
}

func sumSales(data []models.CorrelationPoint) int {
	total := 0
	for _, p := range data {
		total += p.UnitsSold
	}
	return total
}
