package ui

import (
	"os"

	"golang.org/x/term"
)

// GetTerminalWidth returns terminal width or default
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < 40 {
		return 100 // Fallback to current hardcoded width
	}
	return width
}

// IsColorTerminal checks if terminal supports color
func IsColorTerminal() bool {
	// Respect NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if output is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false // Piped output
	}

	// Check TERM variable
	termType := os.Getenv("TERM")
	if termType == "dumb" || termType == "" {
		return false
	}

	return true
}

// IsPipedOutput checks if stdout is piped
func IsPipedOutput() bool {
	return !term.IsTerminal(int(os.Stdout.Fd()))
}
