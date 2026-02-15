package ui

import (
	"testing"
	"time"
)

func TestFormatUUID_FullWidth(t *testing.T) {
	id := "abc12345-1234-5678-90ab-cdef12345678"
	result := FormatUUID(id, 0)
	if result != id {
		t.Errorf("Expected full UUID, got %s", result)
	}
}

func TestFormatUUID_Truncated(t *testing.T) {
	id := "abc12345-1234-5678-90ab-cdef12345678"
	result := FormatUUID(id, 8)
	expected := "abc12345…"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFormatNumber_Thousands(t *testing.T) {
	result := FormatNumber(1234)
	expected := "1.2K"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFormatNumber_Millions(t *testing.T) {
	result := FormatNumber(1234567)
	expected := "1.2M"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFormatDate_Today(t *testing.T) {
	now := time.Now()
	result := FormatDate(now)
	if !contains(result, "Today") {
		t.Errorf("Expected 'Today', got %s", result)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
