package ui

import (
	"strings"
	"testing"
	"time"
)

func TestApp_RenderMetricsView_EmptyData(t *testing.T) {
	app := &App{
		metricsTimeSeries: NewMetricsTimeSeries(12),
	}

	result := app.renderMetricsView()

	// Should show "collecting metrics" message
	if !strings.Contains(result, "Collecting metrics data") {
		t.Error("Should show collecting message when no data")
	}

	if strings.Contains(result, "Indexing Rate") {
		t.Error("Should not show graphs when no data")
	}
}

func TestApp_RenderMetricsView_WithData(t *testing.T) {
	app := &App{
		metricsTimeSeries: NewMetricsTimeSeries(12),
		lastMetricsUpdate: time.Now(),
	}

	// Add some test data
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	result := app.renderMetricsView()

	// Should show metric headers
	if !strings.Contains(result, "Indexing Rate") {
		t.Error("Should show Indexing Rate header")
	}

	if !strings.Contains(result, "Search Rate") {
		t.Error("Should show Search Rate header")
	}

	// Should show statistics
	if !strings.Contains(result, "Current:") {
		t.Error("Should show Current statistic")
	}

	if !strings.Contains(result, "Average:") {
		t.Error("Should show Average statistic")
	}

	if !strings.Contains(result, "Peak:") {
		t.Error("Should show Peak statistic")
	}

	// Should show footer info
	if !strings.Contains(result, "Last updated:") {
		t.Error("Should show last updated time")
	}

	if !strings.Contains(result, "Data points:") {
		t.Error("Should show data point count")
	}
}

func TestRenderMetricStats(t *testing.T) {
	summary := MetricsSummary{
		Current: 1234.5,
		Average: 1000.0,
		Peak:    2000.0,
		Min:     500.0,
	}

	result := renderMetricStats(summary)

	// Should contain all values
	if !strings.Contains(result, "1,234") || !strings.Contains(result, "1,000") || !strings.Contains(result, "2,000") {
		t.Errorf("Should contain formatted numbers, got: %s", result)
	}

	if !strings.Contains(result, "Current:") {
		t.Error("Should have Current label")
	}

	if !strings.Contains(result, "Average:") {
		t.Error("Should have Average label")
	}

	if !strings.Contains(result, "Peak:") {
		t.Error("Should have Peak label")
	}
}

func TestRenderMetricsGraph_EmptyData(t *testing.T) {
	dataPoints := []MetricsDataPoint{}

	result := renderMetricsGraph(dataPoints, "insert", 68, 8)

	if !strings.Contains(result, "No data") {
		t.Error("Should show 'No data' for empty input")
	}
}

func TestRenderMetricsGraph_WithData(t *testing.T) {
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 100, SearchRate: 50},
		{Timestamp: time.Unix(5, 0), InsertRate: 200, SearchRate: 100},
		{Timestamp: time.Unix(10, 0), InsertRate: 150, SearchRate: 75},
	}

	result := renderMetricsGraph(dataPoints, "insert", 68, 8)

	// Should return some graph content (not empty)
	if len(result) == 0 {
		t.Error("Graph should not be empty with valid data")
	}

	// Should not show error message
	if strings.Contains(result, "No data") {
		t.Error("Should not show 'No data' with valid input")
	}
}

func TestRenderMetricsGraph_SinglePoint(t *testing.T) {
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 100, SearchRate: 50},
	}

	result := renderMetricsGraph(dataPoints, "insert", 68, 8)

	// Should handle single point without panic
	if len(result) == 0 {
		t.Error("Should render graph even with single point")
	}
}

func TestRenderMetricsGraph_ZeroValues(t *testing.T) {
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 0, SearchRate: 0},
		{Timestamp: time.Unix(5, 0), InsertRate: 0, SearchRate: 0},
	}

	result := renderMetricsGraph(dataPoints, "insert", 68, 8)

	// Should handle all zeros without panic
	if len(result) == 0 {
		t.Error("Should render graph even with zero values")
	}
}

func TestRenderMetricsGraph_SearchMetric(t *testing.T) {
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 100, SearchRate: 50},
		{Timestamp: time.Unix(5, 0), InsertRate: 200, SearchRate: 100},
	}

	result := renderMetricsGraph(dataPoints, "search", 68, 8)

	// Should render for search metric
	if len(result) == 0 {
		t.Error("Should render search metric graph")
	}
}

func TestFormatMetricNumber_SmallNumber(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0.5, "0.50"},
		{0.12, "0.12"},
		{0.99, "0.99"},
	}

	for _, tt := range tests {
		result := formatMetricNumber(tt.input)
		if result != tt.expected {
			t.Errorf("formatMetricNumber(%.2f) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatMetricNumber_LargeNumber(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{1234.0, "1,234"},
		{1000000.0, "1,000,000"},
		{999.0, "999"},
		{1000.0, "1,000"},
	}

	for _, tt := range tests {
		result := formatMetricNumber(tt.input)
		if result != tt.expected {
			t.Errorf("formatMetricNumber(%.0f) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatMetricNumber_EdgeCases(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0.0, "0"},
		{1.0, "1"},
		{10.0, "10"},
		{100.0, "100"},
	}

	for _, tt := range tests {
		result := formatMetricNumber(tt.input)
		if result != tt.expected {
			t.Errorf("formatMetricNumber(%.0f) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestApp_RenderMetricsView_TimeRange(t *testing.T) {
	app := &App{
		metricsTimeSeries: NewMetricsTimeSeries(12),
		lastMetricsUpdate: time.Now(),
	}

	// Add data with known time range
	baseTime := time.Unix(0, 0)
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   baseTime,
		IndexTotal:  0,
		SearchTotal: 0,
	})
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   baseTime.Add(30 * time.Second),
		IndexTotal:  1500,
		SearchTotal: 750,
	})

	result := app.renderMetricsView()

	// Should show time range
	if !strings.Contains(result, "Time range:") {
		t.Error("Should show time range in footer")
	}
}

func TestRenderMetricsGraph_LargeValues(t *testing.T) {
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 10000, SearchRate: 5000},
		{Timestamp: time.Unix(5, 0), InsertRate: 20000, SearchRate: 10000},
		{Timestamp: time.Unix(10, 0), InsertRate: 15000, SearchRate: 7500},
	}

	result := renderMetricsGraph(dataPoints, "insert", 68, 8)

	// Should handle large values without panic
	if len(result) == 0 {
		t.Error("Should render graph with large values")
	}
}

func TestRenderMetricsGraph_VaryingValues(t *testing.T) {
	// Test with highly varying values to ensure scaling works
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 10, SearchRate: 5},
		{Timestamp: time.Unix(5, 0), InsertRate: 1000, SearchRate: 500},
		{Timestamp: time.Unix(10, 0), InsertRate: 50, SearchRate: 25},
	}

	result := renderMetricsGraph(dataPoints, "insert", 68, 8)

	// Should handle varying values without panic
	if len(result) == 0 {
		t.Error("Should render graph with varying values")
	}
}

// Edge case tests for views

func TestFormatMetricNumber_Negative(t *testing.T) {
	// Negative numbers shouldn't occur in practice but should be handled gracefully
	result := formatMetricNumber(-123.45)

	// Should format the negative number
	if len(result) == 0 {
		t.Error("Should format negative numbers")
	}
}

func TestFormatMetricNumber_VeryLargeNumber(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{1000000000.0, "1,000,000,000"},
		{9999999999.0, "9,999,999,999"},
	}

	for _, tt := range tests {
		result := formatMetricNumber(tt.input)
		if result != tt.expected {
			t.Errorf("formatMetricNumber(%.0f) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatMetricNumber_BoundaryValues(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"exactly 1", 1.0, "1"},
		{"just under 1", 0.999, "1.00"},
		{"tiny value", 0.001, "0.00"},
		{"999 to 1000 boundary", 999.0, "999"},
		{"1000 boundary", 1000.0, "1,000"},
	}

	for _, tt := range tests {
		result := formatMetricNumber(tt.input)
		if result != tt.expected {
			t.Errorf("%s: formatMetricNumber(%.3f) = %s, want %s", tt.name, tt.input, result, tt.expected)
		}
	}
}

func TestRenderMetricsGraph_SmallDimensions(t *testing.T) {
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 100, SearchRate: 50},
		{Timestamp: time.Unix(5, 0), InsertRate: 200, SearchRate: 100},
	}

	// Test with very small dimensions
	result := renderMetricsGraph(dataPoints, "insert", 10, 2)

	// Should handle small dimensions without panic
	if len(result) == 0 {
		t.Error("Should render graph even with small dimensions")
	}
}

func TestRenderMetricsGraph_UnknownMetricType(t *testing.T) {
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 100, SearchRate: 50},
		{Timestamp: time.Unix(5, 0), InsertRate: 200, SearchRate: 100},
	}

	// Test with unknown metric type
	result := renderMetricsGraph(dataPoints, "unknown", 68, 8)

	// Should handle unknown metric type gracefully
	if len(result) == 0 {
		t.Error("Should render something even with unknown metric type")
	}
}

func TestRenderMetricsGraph_AllZeroRates(t *testing.T) {
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 0, SearchRate: 0},
		{Timestamp: time.Unix(5, 0), InsertRate: 0, SearchRate: 0},
		{Timestamp: time.Unix(10, 0), InsertRate: 0, SearchRate: 0},
	}

	result := renderMetricsGraph(dataPoints, "insert", 68, 8)

	// Should render graph even with all zero values
	if len(result) == 0 {
		t.Error("Should render graph with all zero values")
	}
}

func TestRenderMetricStats_ZeroValues(t *testing.T) {
	summary := MetricsSummary{
		Current: 0,
		Average: 0,
		Peak:    0,
		Min:     0,
	}

	result := renderMetricStats(summary)

	// Should contain formatted zeros
	if !strings.Contains(result, "0") {
		t.Error("Should format zero values")
	}

	if !strings.Contains(result, "Current:") && !strings.Contains(result, "Average:") && !strings.Contains(result, "Peak:") {
		t.Error("Should contain stat labels even with zero values")
	}
}

func TestRenderMetricStats_VeryLargeValues(t *testing.T) {
	summary := MetricsSummary{
		Current: 1234567.0,
		Average: 9876543.0,
		Peak:    11111111.0,
		Min:     100.0,
	}

	result := renderMetricStats(summary)

	// Should contain comma-formatted large numbers
	if !strings.Contains(result, "1,234,567") {
		t.Errorf("Should format large numbers with commas, got: %s", result)
	}

	if !strings.Contains(result, "9,876,543") {
		t.Errorf("Should format large average with commas, got: %s", result)
	}
}

func TestApp_RenderMetricsView_NeverUpdated(t *testing.T) {
	app := &App{
		metricsTimeSeries: NewMetricsTimeSeries(12),
		// lastMetricsUpdate is zero value (never updated)
	}

	// Add some data
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(0, 0),
		IndexTotal:  0,
		SearchTotal: 0,
	})
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   time.Unix(5, 0),
		IndexTotal:  500,
		SearchTotal: 250,
	})

	result := app.renderMetricsView()

	// Should show "Never" for last update time
	if !strings.Contains(result, "Never") {
		t.Error("Should show 'Never' when lastMetricsUpdate is zero")
	}
}

func TestRenderMetricsGraph_ConsistentOutput(t *testing.T) {
	// Test that same input produces same output (deterministic rendering)
	dataPoints := []MetricsDataPoint{
		{Timestamp: time.Unix(0, 0), InsertRate: 100, SearchRate: 50},
		{Timestamp: time.Unix(5, 0), InsertRate: 200, SearchRate: 100},
		{Timestamp: time.Unix(10, 0), InsertRate: 150, SearchRate: 75},
	}

	result1 := renderMetricsGraph(dataPoints, "insert", 68, 8)
	result2 := renderMetricsGraph(dataPoints, "insert", 68, 8)

	if result1 != result2 {
		t.Error("renderMetricsGraph should produce consistent output for same input")
	}
}
