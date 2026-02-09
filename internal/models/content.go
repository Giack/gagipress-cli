package models

import (
	"time"
)

// ContentIdea represents a content idea
type ContentIdea struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"` // educational, entertainment, bts, ugc, trend
	BriefDescription string    `json:"brief_description"`
	RelevanceScore   *int      `json:"relevance_score,omitempty"`
	BookID           *string   `json:"book_id,omitempty"`
	Status           string    `json:"status"` // pending, approved, rejected, scripted
	GeneratedAt      time.Time `json:"generated_at"`
	Metadata         any       `json:"metadata,omitempty"` // JSONB field
}

// ContentIdeaInput represents input for creating a content idea
type ContentIdeaInput struct {
	Type             string  `json:"type"`
	BriefDescription string  `json:"brief_description"`
	RelevanceScore   *int    `json:"relevance_score,omitempty"`
	BookID           *string `json:"book_id,omitempty"`
	Metadata         any     `json:"metadata,omitempty"`
}

// Validate validates content idea input
func (c *ContentIdeaInput) Validate() error {
	if c.Type == "" {
		return ErrInvalidInput{Field: "type", Message: "type is required"}
	}
	validTypes := map[string]bool{
		"educational":   true,
		"entertainment": true,
		"bts":           true,
		"ugc":           true,
		"trend":         true,
	}
	if !validTypes[c.Type] {
		return ErrInvalidInput{Field: "type", Message: "invalid type"}
	}
	if c.BriefDescription == "" {
		return ErrInvalidInput{Field: "brief_description", Message: "brief description is required"}
	}
	return nil
}

// ContentScript represents a generated script
type ContentScript struct {
	ID              string    `json:"id"`
	IdeaID          string    `json:"idea_id"`
	Hook            string    `json:"hook"`
	MainContent     string    `json:"main_content"`
	CTA             string    `json:"cta"`
	Hashtags        []string  `json:"hashtags,omitempty"`
	EstimatedLength int       `json:"estimated_length"` // seconds
	Format          string    `json:"format"`           // vertical, square
	ScriptedAt      time.Time `json:"scripted_at"`
}

// ContentScriptInput represents input for creating a script
type ContentScriptInput struct {
	IdeaID          string   `json:"idea_id"`
	Hook            string   `json:"hook"`
	MainContent     string   `json:"main_content"`
	CTA             string   `json:"cta"`
	Hashtags        []string `json:"hashtags,omitempty"`
	EstimatedLength int      `json:"estimated_length"`
	Format          string   `json:"format"`
}

// Validate validates content script input
func (c *ContentScriptInput) Validate() error {
	if c.IdeaID == "" {
		return ErrInvalidInput{Field: "idea_id", Message: "idea ID is required"}
	}
	if c.Hook == "" {
		return ErrInvalidInput{Field: "hook", Message: "hook is required"}
	}
	if c.MainContent == "" {
		return ErrInvalidInput{Field: "main_content", Message: "main content is required"}
	}
	if c.CTA == "" {
		return ErrInvalidInput{Field: "cta", Message: "CTA is required"}
	}
	return nil
}

// ContentCalendar represents a scheduled post
type ContentCalendar struct {
	ID           string     `json:"id"`
	ScriptID     *string    `json:"script_id,omitempty"`
	ScheduledFor time.Time  `json:"scheduled_for"`
	Platform     string     `json:"platform"` // instagram, tiktok
	Status       string     `json:"status"`   // pending_approval, approved, published, failed
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
}

// ContentCalendarInput represents input for creating a calendar entry
type ContentCalendarInput struct {
	ScriptID     *string   `json:"script_id,omitempty"`
	ScheduledFor time.Time `json:"scheduled_for"`
	Platform     string    `json:"platform"`
}

// Validate validates content calendar input
func (c *ContentCalendarInput) Validate() error {
	if c.ScheduledFor.IsZero() {
		return ErrInvalidInput{Field: "scheduled_for", Message: "scheduled time is required"}
	}
	if c.Platform != "instagram" && c.Platform != "tiktok" {
		return ErrInvalidInput{Field: "platform", Message: "platform must be 'instagram' or 'tiktok'"}
	}
	return nil
}
