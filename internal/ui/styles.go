package ui

import "github.com/charmbracelet/lipgloss"

// Color scheme
var (
	ColorPrimary = lipgloss.Color("#5B8FF9") // Blue
	ColorSuccess = lipgloss.Color("#52C41A") // Green
	ColorWarning = lipgloss.Color("#FAAD14") // Yellow
	ColorError   = lipgloss.Color("#F5222D") // Red
	ColorMuted   = lipgloss.Color("#8C8C8C") // Gray
)

// Pre-configured styles
var (
	StyleHeader  = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleError   = lipgloss.NewStyle().Foreground(ColorError)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
)

// Status badge styles
var (
	BadgePending  = lipgloss.NewStyle().Foreground(ColorWarning).Render("pending")
	BadgeApproved = lipgloss.NewStyle().Foreground(ColorSuccess).Render("approved")
	BadgeRejected = lipgloss.NewStyle().Foreground(ColorError).Render("rejected")
)

// FormatStatus returns a colored status badge
func FormatStatus(status string) string {
	if !IsColorTerminal() {
		return status // Graceful degradation
	}

	switch status {
	case "pending":
		return BadgePending
	case "approved":
		return BadgeApproved
	case "rejected":
		return BadgeRejected
	default:
		return status
	}
}
