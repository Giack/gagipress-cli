package ui

import (
	"fmt"
	"time"
)

// Spinner provides visual feedback for long operations
type Spinner struct {
	message string
	done    chan bool
}

// NewSpinner creates a new spinner with message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    make(chan bool),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				fmt.Printf("\r%s %s", frames[i], s.message)
				i = (i + 1) % len(frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.done <- true
	fmt.Print("\r\033[K") // Clear line
}

// Success shows success message
func Success(message string) {
	fmt.Printf("✓ %s\n", message)
}

// Error shows error message
func Error(message string) {
	fmt.Printf("✗ %s\n", message)
}

// Info shows info message
func Info(message string) {
	fmt.Printf("ℹ %s\n", message)
}

// Warning shows warning message
func Warning(message string) {
	fmt.Printf("⚠ %s\n", message)
}
