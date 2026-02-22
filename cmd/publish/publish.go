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
	withMedia  bool
	batchLimit int
)

// PublishCmd represents the publish command group
var PublishCmd = &cobra.Command{
	Use:   "publish [calendar-entry-id]",
	Short: "Publish or schedule a post using Blotato",
	Long: `Publish or schedule a post on social media using Blotato's API.
This command takes a scheduled post from the content calendar and
sends it to Blotato for publishing or scheduling on the target platform.
If no arguments are provided, it can run as a parent command for subcommands like 'batch'.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPublish,
}

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Publish/schedule all approved calendar entries",
	RunE:  runBatchPublish,
}

func init() {
	PublishCmd.Flags().BoolVar(&withMedia, "with-media", false, "Generate media using Blotato template before publishing (requires TemplateID in config)")

	batchCmd.Flags().BoolVar(&withMedia, "with-media", false, "Generate media for each post before publishing")
	batchCmd.Flags().IntVar(&batchLimit, "limit", 10, "Maximum number of posts to submit in this batch")
	PublishCmd.AddCommand(batchCmd)
}

func runPublish(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
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

		// Media generation logic
		// We pass the script as a prompt for the AI template generator
		spinner = ui.NewSpinner(fmt.Sprintf("Requesting Blotato visual creation (Template %s)...", cfg.Blotato.TemplateID))
		spinner.Start()

		prompt := fmt.Sprintf("Create a promotional visual for a book post.\nHook: %s\nMain topic: %s", script.Hook, script.FullScript)
		creationID, err := blotatoClient.GenerateVisual(cfg.Blotato.TemplateID, prompt)
		spinner.Stop()

		if err != nil {
			return fmt.Errorf("failed to start visual generation: %w", err)
		}

		fmt.Printf("✅ Blotato generation started (ID: %s). Waiting for render to finish...\n", creationID)

		spinner = ui.NewSpinner("Waiting for Blotato visual render...")
		spinner.Start()
		mediaURL, err := blotatoClient.WaitForVisualCreation(creationID)
		spinner.Stop()

		if err != nil {
			return fmt.Errorf("failed during visual generation polling: %w", err)
		}

		fmt.Printf("✅ Blotato visual generated successfully: %s\n", mediaURL)
		mediaUrls = append(mediaUrls, mediaURL)
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

func runBatchPublish(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Blotato.APIKey == "" {
		return fmt.Errorf("blotato API key is not configured")
	}

	fmt.Println(ui.StyleHeader.Render("🚀 Batch Publish Posts via Blotato"))

	calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
	contentRepo := repository.NewContentRepository(&cfg.Supabase)
	blotatoClient := social.NewBlotatoClient(cfg.Blotato.APIKey)

	spinner := ui.NewSpinner("Fetching approved calendar entries...")
	spinner.Start()
	entries, err := calendarRepo.GetEntries("approved", batchLimit)
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("failed to get calendar entries: %w", err)
	}

	if len(entries) == 0 {
		ui.Success("No approved posts ready to be scheduled/published.")
		return nil
	}

	fmt.Printf("Found %d approved posts to process.\n\n", len(entries))

	successCount := 0
	failedCount := 0

	// Cache Account IDs to avoid multiple API calls
	accountIDs := make(map[string]string)

	for i, entry := range entries {
		fmt.Printf("[%d/%d] Submitting entry: %s (Platform: %s)\n", i+1, len(entries), entry.ID[:8], entry.Platform)

		if entry.ScriptID == nil {
			fmt.Println("❌ Failed: no script attached")
			failedCount++
			continue
		}

		script, err := contentRepo.GetScriptByID(*entry.ScriptID)
		if err != nil {
			fmt.Printf("❌ Failed to get script: %v\n", err)
			failedCount++
			continue
		}

		// Account caching
		if accountIDs[entry.Platform] == "" {
			accID, err := blotatoClient.GetAccountID(entry.Platform)
			if err != nil {
				fmt.Printf("❌ Failed to get Blotato account for %s: %v\n", entry.Platform, err)
				failedCount++
				continue
			}
			accountIDs[entry.Platform] = accID
		}
		accountID := accountIDs[entry.Platform]

		// Build text
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

		// Media
		var mediaUrls []string
		if withMedia {
			prompt := fmt.Sprintf("Create a promotional visual for a book post.\nHook: %s\nMain topic: %s", script.Hook, script.FullScript)
			creationID, err := blotatoClient.GenerateVisual(cfg.Blotato.TemplateID, prompt)
			if err != nil {
				fmt.Printf("❌ Failed to request visual: %v\n", err)
				failedCount++
				continue
			}
			mediaURL, err := blotatoClient.WaitForVisualCreation(creationID)
			if err != nil {
				fmt.Printf("❌ Failed to generate visual: %v\n", err)
				failedCount++
				continue
			}
			mediaUrls = append(mediaUrls, mediaURL)
			fmt.Printf("   🖼️  Media generated: %s\n", mediaURL)
		}

		// Submit
		submissionID, err := blotatoClient.PublishPost(accountID, entry.Platform, postText.String(), mediaUrls, &entry.ScheduledFor)
		if err != nil {
			fmt.Printf("❌ Failed to submit to Blotato: %v\n", err)
			_ = calendarRepo.UpdateEntryStatus(entry.ID, "failed")
			failedCount++
			continue
		}

		err = calendarRepo.UpdateEntryStatus(entry.ID, "published")
		if err != nil {
			fmt.Printf("⚠️ Submitted (ID: %s) but failed to update local status: %v\n", submissionID, err)
		} else {
			fmt.Printf("✅ Success (Submission ID: %s)\n", submissionID)
			successCount++
		}
	}

	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Printf("Batch publish complete!\n")
	fmt.Printf("Total: %d | Success: %d | Failed: %d\n", len(entries), successCount, failedCount)

	return nil
}
