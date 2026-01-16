package ui

import (
	"fmt"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
)

// renderMetricsView renders the Live Metrics view with graphs
func (a *App) renderMetricsView() string {
	var content string

	// Header
	header := headerStyle.Render("Live Cluster Metrics") + " " + subtleStyle.Render("(Last 60 seconds, auto-refresh: 5s)")
	content += header + "\n"
	content += dividerStyle.Render("────────────────────────────────────────────────────────────────") + "\n\n"

	// Check if we have data
	if a.metricsTimeSeries.Size() == 0 {
		content += subtleStyle.Render("Collecting metrics data...") + "\n"
		content += subtleStyle.Render("Graphs will appear after first data point (5 seconds)") + "\n"
		return content
	}

	// Get data points
	dataPoints := a.metricsTimeSeries.GetDataPoints()

	// Calculate summaries
	insertSummary := a.metricsTimeSeries.CalculateSummary("insert")
	searchSummary := a.metricsTimeSeries.CalculateSummary("search")

	// Indexing section
	content += metricHeaderStyle.Render("Indexing Rate") + "\n"
	content += renderMetricStats(insertSummary) + "\n\n"

	// Render insert rate graph
	graphWidth := 68
	graphHeight := 8
	insertGraph := renderMetricsGraph(dataPoints, "insert", graphWidth, graphHeight)
	content += insertGraph + "\n"
	content += subtleStyle.Render("0s            30s             60s") + "\n\n"

	// Search section
	content += metricHeaderStyle.Render("Search Rate") + "\n"
	content += renderMetricStats(searchSummary) + "\n\n"

	// Render search rate graph
	searchGraph := renderMetricsGraph(dataPoints, "search", graphWidth, graphHeight)
	content += searchGraph + "\n"
	content += subtleStyle.Render("0s            30s             60s") + "\n\n"

	// Footer with info
	lastUpdate := "Never"
	if !a.lastMetricsUpdate.IsZero() {
		lastUpdate = a.lastMetricsUpdate.Format("15:04:05")
	}
	timeRange := a.metricsTimeSeries.GetTimeRange()

	footer := fmt.Sprintf("ℹLast updated: %s  │  Data points: %d  │  Time range: %.0fs",
		lastUpdate,
		a.metricsTimeSeries.Size(),
		timeRange.Seconds(),
	)
	content += subtleStyle.Render(footer) + "\n"

	return content
}

// renderMetricStats renders statistics for a metric (current, average, peak)
func renderMetricStats(summary MetricsSummary) string {
	return fmt.Sprintf("Current: %s/s  │  Average: %s/s  │  Peak: %s/s",
		highlightStyle.Render(formatMetricNumber(summary.Current)),
		statsStyle.Render(formatMetricNumber(summary.Average)),
		highlightStyle.Render(formatMetricNumber(summary.Peak)),
	)
}

// renderMetricsGraph creates a line chart using ntcharts
func renderMetricsGraph(dataPoints []MetricsDataPoint, metricType string, width, height int) string {
	if len(dataPoints) == 0 {
		return subtleStyle.Render("No data")
	}

	// Extract values
	var values []float64
	for _, dp := range dataPoints {
		var val float64
		switch metricType {
		case "insert":
			val = dp.InsertRate
		case "search":
			val = dp.SearchRate
		}
		values = append(values, val)
	}

	// Find min/max for Y axis
	minY, maxY := values[0], values[0]
	for _, v := range values {
		if v < minY {
			minY = v
		}
		if v > maxY {
			maxY = v
		}
	}

	// Add padding to Y axis range
	yRange := maxY - minY
	if yRange == 0 {
		yRange = maxY * 0.1
		if yRange == 0 {
			yRange = 1 // Prevent zero range
		}
	}
	minY -= yRange * 0.1
	maxY += yRange * 0.1
	if minY < 0 {
		minY = 0
	}

	// X axis goes from 0 to len(dataPoints)-1
	minX := 0.0
	maxX := float64(len(dataPoints) - 1)
	if maxX == 0 {
		maxX = 1 // Prevent zero range
	}

	// Create linechart
	lc := linechart.New(width, height, minX, maxX, minY, maxY)

	// Draw lines connecting data points using Braille
	for i := 0; i < len(values)-1; i++ {
		p1 := canvas.Float64Point{X: float64(i), Y: values[i]}
		p2 := canvas.Float64Point{X: float64(i + 1), Y: values[i+1]}
		lc.DrawBrailleLine(p1, p2)
	}

	return lc.View()
}

// formatMetricNumber formats large numbers with commas
func formatMetricNumber(n float64) string {
	// For very small numbers, show decimal places
	if n < 1.0 && n > 0 {
		return fmt.Sprintf("%.2f", n)
	}

	// For larger numbers, format with commas
	s := fmt.Sprintf("%.0f", n)
	if len(s) > 3 {
		parts := []rune(s)
		result := ""
		for i, c := range parts {
			if i > 0 && (len(parts)-i)%3 == 0 {
				result += ","
			}
			result += string(c)
		}
		return result
	}
	return s
}
