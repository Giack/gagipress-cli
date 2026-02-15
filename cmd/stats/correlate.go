package stats

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
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

	fmt.Println(ui.StyleHeader.Render("🔗 Social → Sales Correlation Analysis"))
	fmt.Println()

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

	// Create section style
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorMuted).
		Padding(1, 2).
		Width(ui.GetTerminalWidth() - 4)

	// Display results
	resultsContent := fmt.Sprintf(
		"Data points:          %d days\n"+
			"Total views:          %s\n"+
			"Total sales:          %d units\n"+
			"Correlation (r):      %.3f %s",
		len(dailyData),
		ui.FormatNumber(sumViews(dailyData)),
		sumSales(dailyData),
		correlation,
		renderCorrelationBar(correlation),
	)

	fmt.Println(ui.StyleHeader.Render("📊 Correlation Results"))
	fmt.Println(sectionStyle.Render(resultsContent))
	fmt.Println()

	// Interpret correlation
	absCorr := math.Abs(correlation)
	var strength string
	var interpretation string
	var strengthStyle lipgloss.Style

	if absCorr >= 0.7 {
		strength = "Strong"
		interpretation = "Social media activity has a strong relationship with sales"
		strengthStyle = ui.StyleSuccess
	} else if absCorr >= 0.4 {
		strength = "Moderate"
		interpretation = "Social media activity has a moderate relationship with sales"
		strengthStyle = ui.StyleWarning
	} else if absCorr >= 0.2 {
		strength = "Weak"
		interpretation = "Social media activity has a weak relationship with sales"
		strengthStyle = ui.StyleMuted
	} else {
		strength = "Very Weak / None"
		interpretation = "Little to no relationship detected between social media and sales"
		strengthStyle = ui.StyleMuted
	}

	var directionText string
	if correlation > 0 {
		directionText = fmt.Sprintf("Strength:   %s positive correlation\n"+
			"Direction:  📈 Positive (more views → more sales)",
			strengthStyle.Render(strength))
	} else {
		directionText = fmt.Sprintf("Strength:   %s negative correlation\n"+
			"Direction:  📉 Negative (more views → fewer sales)",
			strengthStyle.Render(strength))
	}

	interpretContent := fmt.Sprintf("%s\n\n%s", directionText, interpretation)

	fmt.Println(ui.StyleHeader.Render("📈 Interpretation"))
	fmt.Println(sectionStyle.Render(interpretContent))
	fmt.Println()

	// Recommendations
	var recsContent string
	if correlation > 0.4 {
		recsContent = ui.StyleSuccess.Render("✅ Social media is driving sales! Keep posting consistently.") + "\n" +
			"   • Focus on content types with highest engagement\n" +
			"   • Increase posting frequency during peak days\n" +
			"   • Use top-performing content as templates"
	} else if correlation > 0.1 {
		recsContent = ui.StyleWarning.Render("⚠️  Moderate impact. Consider:") + "\n" +
			"   • Stronger CTAs in your posts\n" +
			"   • More direct product mentions\n" +
			"   • Test different content types\n" +
			"   • Add Amazon link in bio"
	} else {
		recsContent = ui.StyleMuted.Render("ℹ️  Low correlation detected. Possible reasons:") + "\n" +
			"   • Need more data (try 60-90 days)\n" +
			"   • Delayed conversion effect\n" +
			"   • Content not targeted enough\n" +
			"   • Audience mismatch"
	}

	fmt.Println(ui.StyleHeader.Render("💡 Recommendations"))
	fmt.Println(sectionStyle.Render(recsContent))

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
				Date: sale.SaleDate.Time,
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

// renderCorrelationBar creates a visual bar showing correlation strength
func renderCorrelationBar(r float64) string {
	absR := math.Abs(r)
	strength := int(absR * 10)
	if strength > 10 {
		strength = 10
	}

	bar := strings.Repeat("█", strength) + strings.Repeat("░", 10-strength)

	color := ui.ColorMuted
	if absR > 0.7 {
		color = ui.ColorSuccess
	} else if absR > 0.4 {
		color = ui.ColorWarning
	}

	return lipgloss.NewStyle().Foreground(color).Render(bar)
}
