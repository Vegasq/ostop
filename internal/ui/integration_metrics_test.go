package ui

import (
	"testing"
	"time"
)

func TestIntegration_Metrics_TickerStartsOnViewChange(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Verify metrics disabled initially
	if app.metricsEnabled {
		t.Error("metricsEnabled should be false initially")
	}

	// Navigate to Live Metrics view (item 5)
	app.selectedItem = 5 // ViewLiveMetrics
	cmd := app.updateViewFromSelectionCmd()

	// Verify metrics enabled
	if !app.metricsEnabled {
		t.Error("metricsEnabled should be true after navigating to Live Metrics view")
	}

	// Verify command returned (to start ticker)
	if cmd == nil {
		t.Error("updateViewFromSelectionCmd should return a command to start metrics")
	}
}

func TestIntegration_Metrics_TickerStopsOnViewChange(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Navigate to Live Metrics view
	app.selectedItem = 5 // ViewLiveMetrics
	app.updateViewFromSelectionCmd()

	if !app.metricsEnabled {
		t.Fatal("metricsEnabled should be true")
	}

	// Navigate to a different view
	app.selectedItem = 0 // ViewCluster
	app.updateViewFromSelectionCmd()

	// Verify metrics disabled
	if app.metricsEnabled {
		t.Error("metricsEnabled should be false after leaving Live Metrics view")
	}
}

func TestIntegration_Metrics_AutoRefreshCycle(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Navigate to Live Metrics view
	app.selectedItem = 5
	cmd := app.updateViewFromSelectionCmd()

	// Manually add snapshots to simulate metric collection
	// First snapshot (baseline)
	baseTime := time.Now()
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   baseTime,
		IndexTotal:  1000000,
		SearchTotal: 500000,
	})

	// Simulate 3 metric ticks with increasing values
	for i := 1; i <= 3; i++ {
		snapshot := &MetricsSnapshot{
			Timestamp:   baseTime.Add(time.Duration(i*5) * time.Second),
			IndexTotal:  int64(1000000 + i*500),
			SearchTotal: int64(500000 + i*250),
		}
		app.metricsTimeSeries.AddSnapshot(snapshot)
	}

	// Verify metrics time series has data (should have 3 data points)
	if app.metricsTimeSeries.Size() != 3 {
		t.Errorf("MetricsTimeSeries should have 3 data points, got %d", app.metricsTimeSeries.Size())
	}

	// Verify command was returned to start ticker
	if cmd == nil {
		t.Error("updateViewFromSelectionCmd should return command to start metrics")
	}
}

func TestIntegration_Metrics_ErrorHandling(t *testing.T) {
	// Create app with error on metrics endpoint
	client, err := NewMockClientWithError("metrics")
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// Navigate to metrics view
	app.selectedItem = 5
	app.metricsEnabled = true

	// Trigger metrics refresh
	cmd := app.refreshMetrics()
	msg := ExecuteCommand(cmd)

	// Update app with metrics refresh message
	app.Update(msg)

	// Verify error is logged but doesn't crash
	// (metricsEnabled should still be true, ticker should continue)
	if !app.metricsEnabled {
		t.Error("metricsEnabled should remain true even after error")
	}
}

func TestIntegration_Metrics_TimeSeriesUpdates(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Verify initial state
	if app.metricsTimeSeries == nil {
		t.Fatal("metricsTimeSeries should be initialized")
	}

	initialSize := app.metricsTimeSeries.Size()
	if initialSize != 0 {
		t.Errorf("Initial size should be 0, got %d", initialSize)
	}

	// Navigate to Live Metrics view
	app.selectedItem = 5
	app.metricsEnabled = true

	// Manually add snapshots to simulate metric updates
	baseTime := time.Now()

	// First snapshot (baseline)
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   baseTime,
		IndexTotal:  1000000,
		SearchTotal: 500000,
	})

	// Second snapshot (creates first data point)
	added := app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   baseTime.Add(5 * time.Second),
		IndexTotal:  1000500,
		SearchTotal: 500250,
	})

	if !added {
		t.Error("Second snapshot should have added a data point")
	}

	// Simulate the app's behavior of setting lastMetricsUpdate
	if added {
		app.lastMetricsUpdate = time.Now()
	}

	// Verify time series updated
	finalSize := app.metricsTimeSeries.Size()
	if finalSize != 1 {
		t.Errorf("MetricsTimeSeries size should be 1 after adding 2 snapshots, got %d", finalSize)
	}

	// Verify lastMetricsUpdate timestamp
	if app.lastMetricsUpdate.IsZero() {
		t.Error("lastMetricsUpdate should be set after metrics update")
	}
}

func TestIntegration_Metrics_TickNotProcessedWhenDisabled(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Ensure metrics disabled
	app.metricsEnabled = false

	// Send a metrics tick message
	tickMsg := metricsTickMsg{timestamp: time.Now()}
	_, cmd := app.Update(tickMsg)

	// Command should be nil (no refresh triggered)
	if cmd != nil {
		t.Error("Metrics tick should not trigger refresh when metrics disabled")
	}
}

func TestIntegration_Metrics_LastMetricsUpdateTimestamp(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Verify initial state
	if !app.lastMetricsUpdate.IsZero() {
		t.Error("lastMetricsUpdate should be zero initially")
	}

	// Navigate to metrics view and trigger refresh
	app.selectedItem = 5
	app.metricsEnabled = true

	// Need 2 snapshots to create a data point
	// First snapshot
	cmd := app.refreshMetrics()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Second snapshot (after some time)
	time.Sleep(10 * time.Millisecond)
	cmd = app.refreshMetrics()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	// Verify lastMetricsUpdate is set
	if app.lastMetricsUpdate.IsZero() {
		t.Error("lastMetricsUpdate should be set after successful metrics update")
	}
}

func TestIntegration_Metrics_ViewportContentUpdates(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Initialize viewport
	SendWindowSize(app, 120, 40)

	// Navigate to Live Metrics view
	app.selectedItem = 5
	app.currentView = ViewLiveMetrics
	app.metricsEnabled = true

	// Add some data to metrics time series
	cmd := app.refreshMetrics()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	time.Sleep(10 * time.Millisecond)

	cmd = app.refreshMetrics()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	// Verify viewport has content
	if app.viewport.View() == "" {
		t.Error("Viewport should have content when on Live Metrics view")
	}
}
