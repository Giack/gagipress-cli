package calendar

import "testing"

// TestCalendarCmd_SingleApproveSubcommand verifies that the 'approve' subcommand
// is registered exactly once. A duplicate registration causes it to appear twice
// in --help output and is confusing for users.
func TestCalendarCmd_SingleApproveSubcommand(t *testing.T) {
	count := 0
	for _, sub := range CalendarCmd.Commands() {
		if sub.Use == "approve" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 'approve' subcommand, got %d", count)
	}
}
