package publish

import (
	"fmt"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/social"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	withMedia bool
)

// PublishCmd represents the publish command group
var PublishCmd = &cobra.Command{
	Use:   "publish [calendar-entry-id]",
	Short: "Publish or schedule a post using Blotato",
	Long: `Publish or schedule a post on social media using Blotato's API.
This command takes a scheduled post from the content calendar and
sends it to Blotato for publishing or scheduling on the target platform.`,
	Args: cobra.ExactArgs(1),
	RunE: runPublish,
}

func init() {
	PublishCmd.Flags().BoolVar(&withMedia, "with-media", false, "Generate media using Blotato template before publishing (requires TemplateID in config)")
}

func runPublish(cmd *cobra.Command, args []string) error {
	entryID := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Blotato.APIKey == "" {
		return fmt.Errorf("blotato API key is not configured. Please run 'gagipress config set blotato.api_key YOUR_KEY'")
	}

	fmt.Println(ui.StyleHeader.Render("🚀 Publish Post via Blotato"))

	// Initialize repositories
	calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
	contentRepo := repository.NewContentRepository(&cfg.Supabase)

	// 1. Get calendar entry
	spinner := ui.NewSpinner("Fetching calendar entry...")
	spinner.Start()
	entry, err := calendarRepo.GetEntryByID(entryID)
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("failed to get calendar entry: %w", err)
	}

	if entry.Status == "published" {
		ui.Warning("This post has already been published!")
		return nil
	}
	if entry.ScriptID == nil {
		return fmt.Errorf("calendar entry does not have a script attached")
	}

	// 2. Get script
	spinner = ui.NewSpinner("Fetching associated script...")
	spinner.Start()
	script, err := contentRepo.GetScriptByID(*entry.ScriptID)
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("failed to get script: %w", err)
	}

	// 3. Build post text
	var postText strings.Builder
	postText.WriteString(script.Hook)
	postText.WriteString("\n\n")
	postText.WriteString(script.FullScript)
	postText.WriteString("\n\n")
	postText.WriteString(script.CTA)

	if len(script.Hashtags) > 0 {
		postText.WriteString("\n\n")
		postText.WriteString(strings.Join(script.Hashtags, " "))
	}

	fmt.Printf("\nTarget Platform: %s\n", entry.Platform)
	fmt.Printf("Scheduled For: %s\n", entry.ScheduledFor.Format("2006-01-02 15:04:05"))

	// 4. Initialize Blotato Client
	blotatoClient := social.NewBlotatoClient(cfg.Blotato.APIKey)

	// 5. Get Account ID for platform
	spinner = ui.NewSpinner(fmt.Sprintf("Fetching Blotato account ID for %s...", entry.Platform))
	spinner.Start()
	accountID, err := blotatoClient.GetAccountID(entry.Platform)
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("failed to get Blotato account ID: %w", err)
	}
	fmt.Printf("✅ Found Blotato Account ID: %s\n", accountID)

	// 6. Optional: Generate Media
	var mediaUrls []string
	if withMedia {
		if cfg.Blotato.TemplateID == "" {
			return fmt.Errorf("blotato TemplateID is not configured but --with-media was requested")
		}
		// Media generation will be implemented in Phase 5
		// For now, it's just a placeholder or we can implement it if needed right away
		ui.Warning("Media generation is scheduled for Phase 5 implementation.")
	}

	// 7. Publish/Schedule Post
	spinner = ui.NewSpinner("Sending to Blotato...")
	spinner.Start()
	submissionID, err := blotatoClient.PublishPost(accountID, entry.Platform, postText.String(), mediaUrls, &entry.ScheduledFor)
	spinner.Stop()

	if err != nil {
		// If it failed, we can mark it as failed in our DB
		_ = calendarRepo.UpdateEntryStatus(entry.ID, "failed")
		return fmt.Errorf("blotato publish failed: %w", err)
	}

	fmt.Printf("\n✅ Successfully submitted to Blotato!\n")
	fmt.Printf("Submission ID: %s\n", submissionID)

	// 8. Update DB status
	err = calendarRepo.UpdateEntryStatus(entry.ID, "published")
	if err != nil {
		ui.Warning(fmt.Sprintf("Post submitted to Blotato, but failed to update local status: %v", err))
	} else {
		fmt.Printf("✅ Local status updated to 'published'\n")
	}

	return nil
}
