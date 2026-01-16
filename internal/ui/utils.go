package ui

import (
	"fmt"
	"strings"
)

// renderBar creates an ASCII progress bar with color thresholds
func renderBar(percentStr string, width int) string {
	// Parse percentage
	var percent float64
	fmt.Sscanf(percentStr, "%f", &percent)

	filled := int((percent / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	// Color based on percentage
	if percent >= 90 {
		return statusRed.Render(bar)
	} else if percent >= 75 {
		return statusYellow.Render(bar)
	}
	return statusGreen.Render(bar)
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatNumber adds comma separators to large numbers
func formatNumber(n int64) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}
	return result.String()
}

// parseRunningTime parses time strings like "1.5s", "30ms", "2m" to seconds
func parseRunningTime(runningTime string) float64 {
	if runningTime == "" {
		return 0
	}

	var value float64
	var unit string

	// Try to parse the value and unit
	n, err := fmt.Sscanf(runningTime, "%f%s", &value, &unit)
	if err != nil || n != 2 {
		return 0
	}

	switch unit {
	case "micros":
		return value / 1_000_000
	case "ms", "millis":
		return value / 1000
	case "s", "sec":
		return value
	case "m", "min":
		return value * 60
	case "h", "hour":
		return value * 3600
	default:
		return value
	}
}

// parseTimeInQueue parses time strings like "1.2s", "500ms" to milliseconds
func parseTimeInQueue(timeStr string) float64 {
	if timeStr == "" {
		return 0
	}

	var value float64
	var unit string

	n, err := fmt.Sscanf(timeStr, "%f%s", &value, &unit)
	if err != nil || n != 2 {
		return 0
	}

	switch unit {
	case "ms", "millis":
		return value
	case "s", "sec":
		return value * 1000
	case "m", "min":
		return value * 60000
	default:
		return value
	}
}

// simplifyAction simplifies task action names for grouping
func simplifyAction(action string) string {
	if strings.Contains(action, "search") {
		return "Search"
	}
	if strings.Contains(action, "bulk") {
		return "Bulk"
	}
	if strings.Contains(action, "index") {
		return "Indexing"
	}
	if strings.Contains(action, "delete") {
		return "Delete"
	}
	if strings.Contains(action, "update") {
		return "Update"
	}
	return "Other"
}
