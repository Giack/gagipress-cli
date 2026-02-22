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

// fetchCoverImage downloads the book cover image.
// Tries cover_image_url first, then builds an Amazon URL from KDPASIN.
// Returns nil bytes (no error) if no URL is available or download fails.
func fetchCoverImage(book *models.Book) ([]byte, string, error) {
	if book == nil {
		return nil, "", nil
	}
	url := book.CoverImageURL
	if url == "" && book.KDPASIN != "" {
		url = fmt.Sprintf(
			"https://images-na.ssl-images-amazon.com/images/P/%s.01._SCLZZZZZZZ_.jpg",
			book.KDPASIN,
		)
	}
	if url == "" {
		return nil, "", nil
	}
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, "", nil // non-fatal: fall back to text-only generation
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", nil
	}
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	return data, mimeType, nil
}

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
			bookTitle := "(no book)"
			if e.Script != nil {
				hook = e.Script.Hook
				if e.Script.Idea != nil && e.Script.Idea.Book != nil {
					bookTitle = e.Script.Idea.Book.Title
				}
			}
			fmt.Printf("  - %s (%s) book: %q hook: %q\n", e.ID, e.Platform, bookTitle, hook)
		}
		return nil
	}

	if apiKey == "" {
		return fmt.Errorf("Gemini API key not configured. Set gemini.api_key in config or GEMINI_API_KEY env var")
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
		// Extract book from the nested join chain
		var book *models.Book
		if entry.Script != nil && entry.Script.Idea != nil {
			book = entry.Script.Idea.Book
		}

		bookTitle := "(no book)"
		if book != nil {
			bookTitle = book.Title
		}
		fmt.Printf("Generating image for entry %s (%s) book: %q... ", entry.ID[:8], entry.Platform, bookTitle)

		prompt := buildImagePrompt(entry.Platform, entry.Script, book)

		var imageBytes []byte

		// Path A: multimodal generation with book cover as visual reference
		coverBytes, coverMIME, _ := fetchCoverImage(book)
		if coverBytes != nil {
			contents := []*genai.Content{
				genai.NewContentFromParts([]*genai.Part{
					genai.NewPartFromText(prompt),
					genai.NewPartFromBytes(coverBytes, coverMIME),
				}, genai.RoleUser),
			}
			genResp, err := genaiClient.Models.GenerateContent(
				ctx,
				"gemini-2.5-flash-preview-image-generation",
				contents,
				&genai.GenerateContentConfig{
					ResponseModalities: []string{"IMAGE", "TEXT"},
				},
			)
			if err != nil {
				fmt.Printf("FAILED (gemini multimodal): %v\n", err)
				failed++
				continue
			}
			// Extract image bytes from the first candidate
			for _, cand := range genResp.Candidates {
				if cand.Content == nil {
					continue
				}
				for _, part := range cand.Content.Parts {
					if part.InlineData != nil {
						imageBytes = part.InlineData.Data
						break
					}
				}
				if imageBytes != nil {
					break
				}
			}
			if imageBytes == nil {
				fmt.Println("FAILED (no image in gemini response)")
				failed++
				continue
			}
		} else {
			// Path B: text-only Imagen generation
			response, err := genaiClient.Models.GenerateImages(
				ctx,
				"imagen-4.0-generate-001",
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
			imageBytes = response.GeneratedImages[0].Image.ImageBytes
		}

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

func buildImagePrompt(platform string, script *models.ContentScriptWithIdea, book *models.Book) string {
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

	bookDesc := "a children's book"
	if book != nil {
		parts := []string{}
		if book.Title != "" {
			parts = append(parts, fmt.Sprintf("'%s'", book.Title))
		}
		if book.Genre != "" {
			parts = append(parts, "a "+book.Genre+" book")
		}
		if book.TargetAudience != "" {
			parts = append(parts, "for "+book.TargetAudience)
		}
		if len(parts) > 0 {
			bookDesc = ""
			for i, p := range parts {
				if i > 0 {
					bookDesc += " "
				}
				bookDesc += p
			}
		}
	}

	if book != nil && (book.CoverImageURL != "" || book.KDPASIN != "") {
		return fmt.Sprintf(
			"Create a vertical 9:16 promotional image INSPIRED BY the attached book cover for %s — %s post.%s "+
				"Match the visual style, color palette, and mood of the cover. "+
				"Style: engaging, colorful, warm. No text overlay. Bright background, playful composition.",
			bookDesc, platformDesc, hook,
		)
	}

	return fmt.Sprintf(
		"Create a vertical 9:16 promotional image for %s — %s post.%s "+
			"Style: engaging, colorful, warm. No text overlay. Bright background, playful composition.",
		bookDesc, platformDesc, hook,
	)
}
