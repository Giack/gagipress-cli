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

// ScriptGenerator generates content scripts from ideas
type ScriptGenerator struct {
	openaiClient *ai.OpenAIClient
	geminiClient *ai.GeminiClient
	contentRepo  *repository.ContentRepository
	useGemini    bool
}

// NewScriptGenerator creates a new script generator
func NewScriptGenerator(cfg *config.Config, useGemini bool) *ScriptGenerator {
	return &ScriptGenerator{
		openaiClient: ai.NewOpenAIClient(&cfg.OpenAI),
		geminiClient: ai.NewGeminiClient(true),
		contentRepo:  repository.NewContentRepository(&cfg.Supabase),
		useGemini:    useGemini,
	}
}

// GeneratedScript represents a generated script from AI
type GeneratedScript struct {
	Hook            string   `json:"hook"`
	MainContent     string   `json:"main_content"`
	CTA             string   `json:"cta"`
	Hashtags        []string `json:"hashtags"`
	MusicSuggestion string   `json:"music_suggestion"`
	VideoNotes      string   `json:"video_notes"`
	EstimatedLength int      `json:"estimated_length"`
}

// GenerateScript generates a complete script from an idea
func (g *ScriptGenerator) GenerateScript(idea *models.ContentIdea, bookTitle, platform string) (*GeneratedScript, error) {
	// Build prompt
	ideaDescription := idea.BriefDescription
	prompt := prompts.ScriptPromptTemplate(ideaDescription, bookTitle, platform)

	var responseText string
	var err error

	// Try OpenAI first unless explicitly using Gemini
	if !g.useGemini {
		fmt.Println("🤖 Using OpenAI for script generation...")
		responseText, err = g.openaiClient.GenerateText(prompt, 0.7)
		if err != nil {
			fmt.Printf("⚠️  OpenAI failed: %v\n", err)
			fmt.Println("🔄 Falling back to Gemini...")
			g.useGemini = true
		}
	}

	// Fallback to Gemini
	if g.useGemini {
		fmt.Println("🤖 Using Gemini for script generation...")
		responseText, err = g.geminiClient.GenerateText(prompt)
		if err != nil {
			return nil, fmt.Errorf("both OpenAI and Gemini failed: %w", err)
		}
	}

	// Parse JSON response
	script, err := g.parseScriptFromResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return script, nil
}

// parseScriptFromResponse parses the AI response into a structured script
func (g *ScriptGenerator) parseScriptFromResponse(response string) (*GeneratedScript, error) {
	// Extract JSON object from response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 || start > end {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	jsonStr := response[start : end+1]

	var script GeneratedScript
	if err := json.Unmarshal([]byte(jsonStr), &script); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate required fields
	if script.Hook == "" {
		return nil, fmt.Errorf("missing required field: hook")
	}
	if script.MainContent == "" {
		return nil, fmt.Errorf("missing required field: main_content")
	}
	if script.CTA == "" {
		return nil, fmt.Errorf("missing required field: cta")
	}

	// Set defaults if missing
	if script.EstimatedLength == 0 {
		script.EstimatedLength = 45 // default 45 seconds
	}
	if len(script.Hashtags) == 0 {
		script.Hashtags = []string{"#booktok", "#bookstagram"} // minimal defaults
	}

	return &script, nil
}

// SaveScript saves generated script to the database
func (g *ScriptGenerator) SaveScript(script *GeneratedScript, ideaID string, format string) (*models.ContentScript, error) {
	input := &models.ContentScriptInput{
		IdeaID:          ideaID,
		Hook:            script.Hook,
		MainContent:     script.MainContent,
		CTA:             script.CTA,
		Hashtags:        script.Hashtags,
		EstimatedLength: script.EstimatedLength,
		Format:          format,
	}

	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	savedScript, err := g.contentRepo.CreateScript(input)
	if err != nil {
		return nil, fmt.Errorf("failed to save script: %w", err)
	}

	// Update idea status to "scripted"
	if err := g.contentRepo.UpdateIdeaStatus(ideaID, "scripted"); err != nil {
		fmt.Printf("⚠️  Warning: failed to update idea status: %v\n", err)
	}

	return savedScript, nil
}
