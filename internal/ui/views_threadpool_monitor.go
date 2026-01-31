package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
)

func (a *App) renderThreadPoolMonitorView() string {
	var content string

	// Header
	header := headerStyle.Render("Thread Pool Monitor") + " " + subtleStyle.Render("(Last 60 seconds, auto-refresh: 5s)")
	content += header + "\n"
	content += dividerStyle.Render("────────────────────────────────────────────────────────────────") + "\n\n"

	// Check if we have data
	if a.threadPoolTimeSeries.Size() == 0 {
		content += subtleStyle.Render("Collecting metrics data...") + "\n"
		content += subtleStyle.Render("Graphs will appear after first data point (5 seconds)") + "\n"
		return content
	}

	dataPoints := a.threadPoolTimeSeries.GetDataPoints()
	summary := a.threadPoolTimeSeries.CalculateSummary()

	// Get list of pools (sorted for consistent ordering)
	pools := make([]string, 0, len(summary))
	for poolName := range summary {
		pools = append(pools, poolName)
	}
	sort.Strings(pools)

	// Determine health status
	healthStatus := "Healthy"
	healthStyled := statusGreen.Render(fmt.Sprintf("Status: %s", healthStatus))
	for _, poolName := range pools {
		s := summary[poolName]
		if s.CurrentRejections > 0 {
			healthStatus = "Critical - Active Rejections"
			healthStyled = statusRed.Render(fmt.Sprintf("Status: %s", healthStatus))
			break
		}
		if s.CurrentQueue > 100 {
			healthStatus = "Warning - High Queue Depth"
			healthStyled = statusYellow.Render(fmt.Sprintf("Status: %s", healthStatus))
		}
	}

	content += healthStyled + "\n\n"

	// Queue Depth Chart
	content += metricHeaderStyle.Render("Queue Depth (Tasks Queued)") + "\n"
	queueChart := a.renderThreadPoolMultiGraph(dataPoints, pools, "queue", 68, 8)
	content += queueChart + "\n"
	content += subtleStyle.Render("0s            30s             60s") + "\n\n"

	// Rejection Rate Chart
	content += metricHeaderStyle.Render("Rejection Rate (Rejections/Second)") + "\n"
	rejectionChart := a.renderThreadPoolMultiGraph(dataPoints, pools, "rejection", 68, 8)
	content += rejectionChart + "\n"
	content += subtleStyle.Render("0s            30s             60s") + "\n\n"

	// Summary Statistics Table
	content += a.renderThreadPoolStatsTable(pools, summary) + "\n\n"

	// Footer
	lastUpdate := "Never"
	if !a.lastThreadPoolUpdate.IsZero() {
		lastUpdate = a.lastThreadPoolUpdate.Format("15:04:05")
	}
	start, end := a.threadPoolTimeSeries.GetTimeRange()
	timeRange := end.Sub(start).Seconds()
	footer := fmt.Sprintf("ℹLast updated: %s  │  Data points: %d  │  Time range: %.0fs",
		lastUpdate,
		a.threadPoolTimeSeries.Size(),
		timeRange)
	content += subtleStyle.Render(footer) + "\n"

	return content
}

func (a *App) renderThreadPoolMultiGraph(dataPoints []ThreadPoolDataPoint, pools []string, metricType string, width, height int) string {
	if len(dataPoints) == 0 {
		return subtleStyle.Render("No data available")
	}

	// Prepare data for each pool and find max value
	poolData := make(map[string][]float64)
	maxValue := 0.0

	for _, dp := range dataPoints {
		for _, poolName := range pools {
			if metrics, ok := dp.Pools[poolName]; ok {
				var value float64
				if metricType == "queue" {
					value = metrics.QueueDepth
				} else {
					value = metrics.RejectionRate
				}

				poolData[poolName] = append(poolData[poolName], value)
				if value > maxValue {
					maxValue = value
				}
			}
		}
	}

	// Add padding to Y axis range
	yRange := maxValue
	if yRange == 0 {
		yRange = 1 // Prevent zero range
	}
	maxY := maxValue + yRange*0.1

	// X axis goes from 0 to len(dataPoints)-1
	minX := 0.0
	maxX := float64(len(dataPoints) - 1)
	if maxX == 0 {
		maxX = 1 // Prevent zero range
	}

	// Create linechart
	lc := linechart.New(width, height, minX, maxX, 0, maxY)

	// Pool colors (using ANSI color codes in drawing)
	poolColors := []string{"42", "39", "201", "51", "226"} // Green, Blue, Magenta, Cyan, Yellow

	// Draw lines for each pool
	for poolIdx, poolName := range pools {
		if data, ok := poolData[poolName]; ok && len(data) > 1 {
			// Draw lines connecting data points
			for i := 0; i < len(data)-1; i++ {
				p1 := canvas.Float64Point{X: float64(i), Y: data[i]}
				p2 := canvas.Float64Point{X: float64(i + 1), Y: data[i+1]}
				lc.DrawBrailleLine(p1, p2)
			}
		}
		_ = poolIdx
		_ = poolColors
	}

	// Build legend
	var legend strings.Builder
	for i, poolName := range pools {
		if i > 0 {
			legend.WriteString("  ")
		}
		legend.WriteString(poolName)
	}

	return lc.View() + "\n" + subtleStyle.Render(legend.String())
}

func (a *App) renderThreadPoolStatsTable(pools []string, summary map[string]ThreadPoolSummary) string {
	var content string

	content += metricHeaderStyle.Render("Summary Statistics") + "\n\n"

	headerRow := fmt.Sprintf("%-12s %10s %10s %10s %10s %12s %12s",
		"Pool", "Cur Queue", "Avg Queue", "Peak Queue", "Min Queue", "Cur Rej/s", "Avg Rej/s")
	content += valueStyle.Render(headerRow) + "\n"
	content += strings.Repeat("─", 80) + "\n"

	// Table rows
	for _, poolName := range pools {
		s := summary[poolName]
		row := fmt.Sprintf("%-12s %10.1f %10.1f %10.1f %10.1f %12.2f %12.2f",
			poolName,
			s.CurrentQueue,
			s.AverageQueue,
			s.PeakQueue,
			s.MinQueue,
			s.CurrentRejections,
			s.AverageRejections)

		// Color code based on health
		if s.CurrentRejections > 0 {
			content += statusRed.Render(row) + "\n"
		} else if s.CurrentQueue > 100 {
			content += statusYellow.Render(row) + "\n"
		} else {
			content += row + "\n"
		}
	}

	return content
}

func threadPoolTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return threadPoolTickMsg{timestamp: t}
	})
}
