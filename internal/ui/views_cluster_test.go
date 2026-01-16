package ui

import (
	"strings"
	"testing"
)

func TestApp_GetNodeTypeLabel(t *testing.T) {
	app := &App{}

	tests := []struct {
		name     string
		role     string
		expected string
	}{
		{"master_only", "m", "Master-eligible"},
		{"data_only", "d", "Data"},
		{"ingest_only", "i", "Ingest"},
		{"coordinating_only", "c", "Coordinating"},
		{"master_data", "md", "Master-eligible, Data"},
		{"master_data_ingest", "mdi", "Master-eligible, Data, Ingest"},
		{"all_roles", "mdic", "Master-eligible, Data, Ingest, Coordinating"},
		{"data_ingest", "di", "Data, Ingest"},
		{"empty", "", "Unknown"},
		{"unknown_role", "xyz", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.getNodeTypeLabel(tt.role)
			if got != tt.expected {
				t.Errorf("getNodeTypeLabel(%q) = %q, want %q", tt.role, got, tt.expected)
			}
		})
	}
}

func TestApp_FormatNodeRoleBadge(t *testing.T) {
	app := &App{}

	tests := []struct {
		name     string
		role     string
		contains []string // Check for presence of badges
	}{
		{"master", "m", []string{"[M]"}},
		{"data", "d", []string{"[D]"}},
		{"ingest", "i", []string{"[I]"}},
		{"coordinating", "c", []string{"[C]"}},
		{"master_data", "md", []string{"[M]", "[D]"}},
		{"all_roles", "mdic", []string{"[M]", "[D]", "[I]", "[C]"}},
		{"empty", "", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.formatNodeRoleBadge(tt.role)
			// Strip ANSI codes for easier testing
			for _, badge := range tt.contains {
				if !strings.Contains(got, badge) {
					t.Errorf("formatNodeRoleBadge(%q) = %q, should contain %q", tt.role, got, badge)
				}
			}
		})
	}
}
