package generator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/ai"
	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/prompts"
	"github.com/gagipress/gagipress-cli/internal/repository"
)

// IdeaGenerator generates content ideas using AI
type IdeaGenerator struct {
	openaiClient  *ai.OpenAIClient
	geminiClient  *ai.GeminiClient
	contentRepo   *repository.ContentRepository
	useGemini     bool
	geminiHeadless bool
}

// NewIdeaGenerator creates a new idea generator
func NewIdeaGenerator(cfg *config.Config, useGemini bool) *IdeaGenerator {
	return &IdeaGenerator{
		openaiClient:   ai.NewOpenAIClient(&cfg.OpenAI),
		geminiClient:   ai.NewGeminiClient(true), // headless by default
		contentRepo:    repository.NewContentRepository(&cfg.Supabase),
		useGemini:      useGemini,
		geminiHeadless: true,
	}
}

// GeneratedIdea represents a generated content idea from AI
type GeneratedIdea struct {
	Type           string `json:"type"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Hook           string `json:"hook"`
	CTA            string `json:"cta"`
	RelevanceScore int    `json:"relevance_score"`
}

// GenerateIdeas generates content ideas for a book
func (g *IdeaGenerator) GenerateIdeas(bookTitle, genre, targetAudience string, niche prompts.BookNiche, count int) ([]GeneratedIdea, error) {
	// Build prompt
	prompt := prompts.IdeaPromptTemplate(bookTitle, genre, targetAudience, niche, count)

	var responseText string
	var err error

	// Try OpenAI first unless explicitly using Gemini
	if !g.useGemini {
		fmt.Println("🤖 Using OpenAI for generation...")
		responseText, err = g.openaiClient.GenerateText(prompt, 0.8)
		if err != nil {
			fmt.Printf("⚠️  OpenAI failed: %v\n", err)
			fmt.Println("🔄 Falling back to Gemini...")
			g.useGemini = true
		}
	}

	// Fallback to Gemini if OpenAI failed or explicitly requested
	if g.useGemini {
		fmt.Println("🤖 Using Gemini for generation...")
		responseText, err = g.geminiClient.GenerateText(prompt)
		if err != nil {
			return nil, fmt.Errorf("both OpenAI and Gemini failed: %w", err)
		}
	}

	// Parse JSON response
	ideas, err := g.parseIdeasFromResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return ideas, nil
}

// parseIdeasFromResponse parses the AI response into structured ideas
func (g *IdeaGenerator) parseIdeasFromResponse(response string) ([]GeneratedIdea, error) {
	// Extract JSON array from response (AI might add text around it)
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")

	if start == -1 || end == -1 || start > end {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	jsonStr := response[start : end+1]

	var ideas []GeneratedIdea
	if err := json.Unmarshal([]byte(jsonStr), &ideas); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return ideas, nil
}

// SaveIdeas saves generated ideas to the database
func (g *IdeaGenerator) SaveIdeas(ideas []GeneratedIdea, bookID *string) ([]models.ContentIdea, error) {
	var savedIdeas []models.ContentIdea

	for _, idea := range ideas {
		input := &models.ContentIdeaInput{
			Type:             idea.Type,
			BriefDescription: idea.Title + ": " + idea.Description,
			RelevanceScore:   &idea.RelevanceScore,
			BookID:           bookID,
			Metadata: map[string]string{
				"hook": idea.Hook,
				"cta":  idea.CTA,
			},
		}

		if err := input.Validate(); err != nil {
			fmt.Printf("⚠️  Skipping invalid idea: %v\n", err)
			continue
		}

		savedIdea, err := g.contentRepo.CreateIdea(input)
		if err != nil {
			fmt.Printf("⚠️  Failed to save idea: %v\n", err)
			continue
		}

		savedIdeas = append(savedIdeas, *savedIdea)
	}

	return savedIdeas, nil
}
