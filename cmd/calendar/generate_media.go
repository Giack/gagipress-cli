package calendar

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

var generateMediaCmd = &cobra.Command{
	Use:   "generate-media",
	Short: "Generate images for scheduled posts using Google Imagen",
	Long: `Generates AI images for calendar entries that have generate_media=true and no media_url yet.
Images are uploaded to Supabase Storage and the public URL is saved to the calendar entry.
The Edge Function will then use these pre-generated images at publish time.

Requires gemini.api_key in your config (or GEMINI_API_KEY env var).`,
	RunE: runGenerateMedia,
}

var (
	generateMediaLimit    int
	generateMediaDryRun   bool
	generateMediaPlatform string
)

func init() {
	generateMediaCmd.Flags().IntVar(&generateMediaLimit, "limit", 10, "Maximum number of images to generate")
	generateMediaCmd.Flags().BoolVar(&generateMediaDryRun, "dry-run", false, "Show what would be generated without doing it")
	generateMediaCmd.Flags().StringVar(&generateMediaPlatform, "platform", "", "Filter by platform: tiktok or instagram")
}

func runGenerateMedia(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	apiKey := cfg.Gemini.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("Gemini API key not configured. Set gemini.api_key in config or GEMINI_API_KEY env var")
	}

	supabaseServiceKey := cfg.Supabase.ServiceKey
	if supabaseServiceKey == "" {
		supabaseServiceKey = cfg.Supabase.AnonKey
	}

	calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)

	entries, err := calendarRepo.GetEntriesNeedingMedia()
	if err != nil {
		return fmt.Errorf("failed to fetch entries: %w", err)
	}

	if generateMediaPlatform != "" {
		filtered := entries[:0]
		for _, e := range entries {
			if e.Platform == generateMediaPlatform {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	if generateMediaLimit > 0 && len(entries) > generateMediaLimit {
		entries = entries[:generateMediaLimit]
	}

	if len(entries) == 0 {
		fmt.Println("No entries need media generation.")
		return nil
	}

	fmt.Printf("Found %d entries needing media generation.\n", len(entries))

	if generateMediaDryRun {
		fmt.Println("\n[dry-run] Would generate images for:")
		for _, e := range entries {
			hook := ""
			if e.Script != nil {
				hook = e.Script.Hook
			}
			fmt.Printf("  - %s (%s) hook: %q\n", e.ID, e.Platform, hook)
		}
		return nil
	}

	if err := os.Setenv("GOOGLE_API_KEY", apiKey); err != nil {
		return fmt.Errorf("failed to set GOOGLE_API_KEY: %w", err)
	}

	ctx := context.Background()
	genaiClient, err := genai.NewClient(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	httpClient := &http.Client{}
	generated := 0
	failed := 0

	for _, entry := range entries {
		prompt := buildImagePrompt(entry.Platform, entry.Script)

		fmt.Printf("Generating image for entry %s (%s)... ", entry.ID[:8], entry.Platform)

		response, err := genaiClient.Models.GenerateImages(
			ctx,
			"imagen-3.0-generate-002",
			prompt,
			&genai.GenerateImagesConfig{
				NumberOfImages: 1,
				AspectRatio:    "9:16",
				OutputMIMEType: "image/jpeg",
			},
		)
		if err != nil {
			fmt.Printf("FAILED (imagen): %v\n", err)
			failed++
			continue
		}

		if len(response.GeneratedImages) == 0 {
			fmt.Println("FAILED (no images returned)")
			failed++
			continue
		}

		imageBytes := response.GeneratedImages[0].Image.ImageBytes

		fileName := fmt.Sprintf("%s.jpg", entry.ID)
		uploadURL := fmt.Sprintf("%s/storage/v1/object/campaign-media/%s", cfg.Supabase.URL, fileName)

		uploadReq, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(imageBytes))
		if err != nil {
			fmt.Printf("FAILED (upload req): %v\n", err)
			failed++
			continue
		}
		// Lowercase content-type is required — uppercase causes Supabase Storage to store
		// the multipart envelope instead of the actual bytes (known Supabase bug).
		uploadReq.Header.Set("content-type", "image/jpeg")
		uploadReq.Header.Set("Authorization", "Bearer "+supabaseServiceKey)
		uploadReq.Header.Set("apikey", supabaseServiceKey)

		uploadResp, err := httpClient.Do(uploadReq)
		if err != nil {
			fmt.Printf("FAILED (upload): %v\n", err)
			failed++
			continue
		}
		uploadBody, _ := io.ReadAll(uploadResp.Body)
		uploadResp.Body.Close()

		if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated {
			fmt.Printf("FAILED (storage %d): %s\n", uploadResp.StatusCode, string(uploadBody))
			failed++
			continue
		}

		publicURL := fmt.Sprintf("%s/storage/v1/object/public/campaign-media/%s", cfg.Supabase.URL, fileName)

		if err := calendarRepo.UpdateMediaURL(entry.ID, publicURL); err != nil {
			fmt.Printf("FAILED (db update): %v\n", err)
			failed++
			continue
		}

		fmt.Printf("OK → %s\n", publicURL)
		generated++
	}

	fmt.Printf("\nDone. Generated: %d, Failed: %d\n", generated, failed)
	return nil
}

func buildImagePrompt(platform string, script *models.ContentScript) string {
	platformDesc := "TikTok/Instagram Reels"
	if platform == "instagram" {
		platformDesc = "Instagram Reels"
	} else if platform == "tiktok" {
		platformDesc = "TikTok"
	}

	hook := ""
	if script != nil && script.Hook != "" {
		hook = " Hook concept: " + script.Hook + "."
	}

	return fmt.Sprintf(
		"Create a vertical promotional image for a children's book %s post.%s "+
			"Style: engaging, colorful, warm, suitable for children's book promotion. "+
			"No text overlay. Bright background, playful composition.",
		platformDesc, hook,
	)
}
