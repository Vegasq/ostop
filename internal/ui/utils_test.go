package ui

import (
	"strings"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"exact_kb", 1024, "1.0 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
		{"terabytes", 1099511627776, "1.0 TB"},
		{"fractional_kb", 1536, "1.5 KB"},
		{"fractional_mb", 5368709120, "5.0 GB"},
		{"less_than_1kb", 1000, "1000 B"},
		{"large_gb", 10737418240, "10.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.input)
			if got != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0"},
		{"single_digit", 5, "5"},
		{"two_digits", 42, "42"},
		{"three_digits", 999, "999"},
		{"thousands", 1000, "1,000"},
		{"ten_thousands", 12345, "12,345"},
		{"millions", 1000000, "1,000,000"},
		{"large_number", 1234567890, "1,234,567,890"},
		{"negative", -1000, "-1,000"},
		{"negative_millions", -1234567, "-1,234,567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatNumber(tt.input)
			if got != tt.expected {
				t.Errorf("formatNumber(%d) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseRunningTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"empty", "", 0},
		{"milliseconds", "500ms", 0.5},
		{"seconds", "1.5s", 1.5},
		{"minutes", "2m", 120},
		{"hours", "1h", 3600},
		{"fractional_ms", "123.45ms", 0.12345},
		{"fractional_seconds", "30.5s", 30.5},
		{"millis_unit", "1000millis", 1.0},
		{"sec_unit", "5sec", 5.0},
		{"min_unit", "1.5min", 90.0},
		{"hour_unit", "0.5hour", 1800.0},
		{"invalid", "invalid", 0},
		{"missing_unit", "123", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRunningTime(tt.input)
			if got != tt.expected {
				t.Errorf("parseRunningTime(%q) = %f, want %f", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseTimeInQueue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"empty", "", 0},
		{"milliseconds", "100ms", 100},
		{"seconds", "1.5s", 1500},
		{"minutes", "2m", 120000},
		{"fractional_ms", "50.5ms", 50.5},
		{"fractional_seconds", "0.5s", 500},
		{"millis_unit", "250millis", 250},
		{"sec_unit", "3sec", 3000},
		{"min_unit", "1.5min", 90000},
		{"invalid", "invalid", 0},
		{"missing_unit", "500", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTimeInQueue(tt.input)
			if got != tt.expected {
				t.Errorf("parseTimeInQueue(%q) = %f, want %f", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSimplifyAction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"search_action", "indices:data/read/search", "Search"},
		{"bulk_action", "indices:data/write/bulk", "Bulk"},
		{"index_action", "indices:data/write/index", "Indexing"},
		{"delete_action", "indices:data/write/delete", "Delete"},
		{"update_action", "indices:data/write/update", "Update"},
		{"search_lowercase", "search_query", "Search"},
		{"bulk_lowercase", "bulk_insert", "Bulk"},
		{"other_action", "cluster:monitor/health", "Other"},
		{"empty", "", "Other"},
		{"unknown", "unknown:action", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := simplifyAction(tt.input)
			if got != tt.expected {
				t.Errorf("simplifyAction(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRenderBar(t *testing.T) {
	tests := []struct {
		name         string
		percentStr   string
		width        int
		expectFilled int
		expectEmpty  int
	}{
		{"zero_percent", "0", 10, 0, 10},
		{"fifty_percent", "50", 10, 5, 5},
		{"hundred_percent", "100", 10, 10, 0},
		{"over_hundred", "150", 10, 10, 0},
		{"negative", "-10", 10, 0, 10},
		{"decimal", "33.3", 9, 2, 7}, // 33.3% of 9 = 2.997 -> 2
		{"ninety_percent", "90", 10, 9, 1},
		{"seventy_five", "75", 20, 15, 5},
		{"small_width", "50", 4, 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderBar(tt.percentStr, tt.width)

			// Count filled and empty characters (ignoring ANSI codes)
			filledCount := strings.Count(result, "█")
			emptyCount := strings.Count(result, "░")

			if filledCount != tt.expectFilled {
				t.Errorf("renderBar(%q, %d) filled count = %d, want %d",
					tt.percentStr, tt.width, filledCount, tt.expectFilled)
			}

			if emptyCount != tt.expectEmpty {
				t.Errorf("renderBar(%q, %d) empty count = %d, want %d",
					tt.percentStr, tt.width, emptyCount, tt.expectEmpty)
			}

			// Verify total width
			total := filledCount + emptyCount
			if total != tt.width {
				t.Errorf("renderBar(%q, %d) total width = %d, want %d",
					tt.percentStr, tt.width, total, tt.width)
			}
		})
	}
}

func TestRenderBar_ColorCoding(t *testing.T) {
	tests := []struct {
		name       string
		percentStr string
		// We can't easily test exact ANSI codes, but we can verify the bar contains the right characters
	}{
		{"green_zone", "50"},
		{"yellow_zone", "75"},
		{"yellow_high", "85"},
		{"red_zone", "90"},
		{"red_max", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderBar(tt.percentStr, 10)

			// Just verify it returns a non-empty string with bar characters
			if !strings.Contains(result, "█") && !strings.Contains(result, "░") {
				t.Errorf("renderBar(%q) should contain bar characters", tt.percentStr)
			}
		})
	}
}
