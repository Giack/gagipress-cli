package ui

import (
	"fmt"
	"time"
)

// FormatUUID shows prefix with ellipsis if needed
func FormatUUID(id string, maxWidth int) string {
	if maxWidth == 0 || len(id) <= maxWidth {
		return id
	}
	return id[:maxWidth] + "…"
}

// FormatNumber formats large numbers with K, M suffixes
func FormatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// FormatDate shows relative dates
func FormatDate(t time.Time) string {
	if time.Since(t) < 24*time.Hour && t.Day() == time.Now().Day() {
		return "Today at " + t.Format("15:04")
	}
	if time.Since(t) < 48*time.Hour && t.Day() == time.Now().Add(-24*time.Hour).Day() {
		return "Yesterday at " + t.Format("15:04")
	}
	return t.Format("Jan 02, 2006")
}
