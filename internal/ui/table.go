package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// TableConfig defines table rendering options
type TableConfig struct {
	Headers  []string
	Rows     [][]string
	MaxWidth int
}

// RenderTable creates a formatted table with auto-width calculation
func RenderTable(cfg TableConfig) string {
	// Use lipgloss.Table for rendering
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Headers(cfg.Headers...).
		Rows(cfg.Rows...)

	// Set width if specified
	if cfg.MaxWidth > 0 {
		t = t.Width(cfg.MaxWidth)
	}

	// Disable styling for non-color terminals
	if !IsColorTerminal() {
		t = t.BorderStyle(lipgloss.NewStyle())
	}

	return t.String()
}
