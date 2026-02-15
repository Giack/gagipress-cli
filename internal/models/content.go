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
	ID                string    `json:"id"`
	IdeaID            string    `json:"idea_id"`
	Hook              string    `json:"hook"`
	FullScript        string    `json:"full_script"`
	CTA               string    `json:"cta"`
	Hashtags          []string  `json:"hashtags,omitempty"`
	EstimatedDuration int       `json:"estimated_duration"` // seconds
	CreatedAt         time.Time `json:"created_at"`
}

// ContentScriptInput represents input for creating a script
type ContentScriptInput struct {
	IdeaID            string   `json:"idea_id"`
	Hook              string   `json:"hook"`
	FullScript        string   `json:"full_script"`
	CTA               string   `json:"cta"`
	Hashtags          []string `json:"hashtags,omitempty"`
	EstimatedDuration int      `json:"estimated_duration"`
}

// Validate validates content script input
func (c *ContentScriptInput) Validate() error {
	if c.IdeaID == "" {
		return ErrInvalidInput{Field: "idea_id", Message: "idea ID is required"}
	}
	if c.Hook == "" {
		return ErrInvalidInput{Field: "hook", Message: "hook is required"}
	}
	if c.FullScript == "" {
		return ErrInvalidInput{Field: "full_script", Message: "full script is required"}
	}
	if c.CTA == "" {
		return ErrInvalidInput{Field: "cta", Message: "CTA is required"}
	}
	return nil
}

// ContentCalendar represents a scheduled post
type ContentCalendar struct {
	ID            string     `json:"id"`
	ScriptID      *string    `json:"script_id,omitempty"`
	ScheduledFor  time.Time  `json:"scheduled_for"`
	Platform      string     `json:"platform"`  // instagram, tiktok
	PostType      string     `json:"post_type"` // reel, story, feed - REQUIRED
	Status        string     `json:"status"`    // pending_approval, approved, published, failed
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	PublishErrors any        `json:"publish_errors,omitempty"` // JSONB field
}

// ContentCalendarInput represents input for creating a calendar entry
type ContentCalendarInput struct {
	ScriptID     *string   `json:"script_id,omitempty"`
	ScheduledFor time.Time `json:"scheduled_for"`
	Platform     string    `json:"platform"`
	PostType     string    `json:"post_type"` // REQUIRED
}

// Validate validates content calendar input
func (c *ContentCalendarInput) Validate() error {
	if c.ScheduledFor.IsZero() {
		return ErrInvalidInput{Field: "scheduled_for", Message: "scheduled time is required"}
	}
	if c.Platform != "instagram" && c.Platform != "tiktok" {
		return ErrInvalidInput{Field: "platform", Message: "platform must be 'instagram' or 'tiktok'"}
	}
	validPostTypes := map[string]bool{"reel": true, "story": true, "feed": true}
	if !validPostTypes[c.PostType] {
		return ErrInvalidInput{Field: "post_type", Message: "post_type must be 'reel', 'story', or 'feed'"}
	}
	return nil
}
