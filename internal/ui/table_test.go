package ui

import (
	"strings"
	"testing"
)

func TestRenderTable_BasicUsage(t *testing.T) {
	cfg := TableConfig{
		Headers: []string{"ID", "Name", "Status"},
		Rows: [][]string{
			{"1", "Alice", "active"},
			{"2", "Bob", "inactive"},
		},
		MaxWidth: 100,
	}

	result := RenderTable(cfg)

	if !strings.Contains(result, "ID") {
		t.Error("Table missing header")
	}
	if !strings.Contains(result, "Alice") {
		t.Error("Table missing data")
	}
}

func TestRenderTable_AutoWidth(t *testing.T) {
	cfg := TableConfig{
		Headers: []string{"Short", "Very Long Column Name"},
		Rows: [][]string{
			{"A", "Data"},
		},
		MaxWidth: 0, // Let table auto-size
	}

	result := RenderTable(cfg)

	// Just verify it renders something
	if len(result) == 0 {
		t.Error("Table rendered empty")
	}
	if !strings.Contains(result, "Short") {
		t.Error("Table missing header")
	}
}
