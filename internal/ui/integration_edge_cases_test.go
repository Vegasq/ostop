package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_EdgeCases_EmptyCluster(t *testing.T) {
	transport := NewMockTransport()

	// Set empty arrays for all endpoints
	transport.SetFixture("health", []byte(`{"cluster_name":"empty","status":"green","number_of_nodes":0,"number_of_data_nodes":0,"active_primary_shards":0,"active_shards":0,"relocating_shards":0,"initializing_shards":0,"unassigned_shards":0}`))
	transport.SetFixture("stats", []byte(`{"cluster_name":"empty","status":"green","indices":{"count":0,"docs":{"count":0},"store":{"size_in_bytes":0}},"nodes":{"count":{"total":0,"data":0}}}`))
	transport.SetFixture("nodes", []byte(`[]`))
	transport.SetFixture("indices", []byte(`[]`))
	transport.SetFixture("shards", []byte(`[]`))
	transport.SetFixture("allocation", []byte(`[]`))
	transport.SetFixture("threadpool", []byte(`[]`))
	transport.SetFixture("tasks", []byte(`[]`))
	transport.SetFixture("pending_tasks", []byte(`[]`))
	transport.SetFixture("recovery", []byte(`[]`))
	transport.SetFixture("segments", []byte(`[]`))
	transport.SetFixture("fielddata", []byte(`[]`))
	transport.SetFixture("plugins", []byte(`[]`))
	transport.SetFixture("templates", []byte(`[]`))

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Should not crash with empty data
	if app.err != nil {
		t.Errorf("Should handle empty cluster without error: %v", app.err)
	}

	// Verify empty arrays
	if len(app.nodes) != 0 {
		t.Errorf("nodes should be empty, got %d", len(app.nodes))
	}

	if len(app.indices) != 0 {
		t.Errorf("indices should be empty, got %d", len(app.indices))
	}

	// Should render view without crashing
	SendWindowSize(app, 120, 40)
	view := app.View()

	if len(view) == 0 {
		t.Error("View should render even with empty data")
	}
}

func TestIntegration_EdgeCases_NoIndicesForDrilldown(t *testing.T) {
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	// Override indices with empty array
	transport.SetFixture("indices", []byte(`[]`))

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	SendWindowSize(app, 120, 40)

	// Navigate to indices view
	app.currentView = ViewIndices
	app.activePanel = PanelRight

	// Try to drill down (should do nothing)
	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not crash or change view
	if app.currentView != ViewIndices {
		t.Error("View should remain on Indices when no indices available")
	}
}

func TestIntegration_EdgeCases_RapidKeyPresses(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Rapidly press keys
	keys := []tea.KeyType{
		tea.KeyDown, tea.KeyDown, tea.KeyUp, tea.KeyDown,
		tea.KeyTab, tea.KeyTab, tea.KeyEnter,
		tea.KeyUp, tea.KeyDown, tea.KeyPgDown, tea.KeyPgUp,
	}

	for _, key := range keys {
		app.Update(tea.KeyMsg{Type: key})
	}

	// Should maintain consistent state
	if app.selectedItem < 0 || app.selectedItem > 14 {
		t.Errorf("selectedItem out of bounds: %d", app.selectedItem)
	}

	if app.activePanel != PanelLeft && app.activePanel != PanelRight {
		t.Errorf("Invalid activePanel: %v", app.activePanel)
	}
}

func TestIntegration_EdgeCases_WindowResize(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Send multiple window size messages
	sizes := []struct {
		width  int
		height int
	}{
		{120, 40},
		{80, 24},
		{200, 60},
		{60, 20},
	}

	for _, size := range sizes {
		SendWindowSize(app, size.width, size.height)

		if app.width != size.width {
			t.Errorf("width = %d, want %d", app.width, size.width)
		}

		if app.height != size.height {
			t.Errorf("height = %d, want %d", app.height, size.height)
		}

		if !app.viewportReady {
			t.Error("viewport should be ready after WindowSizeMsg")
		}
	}
}

func TestIntegration_EdgeCases_ZeroSizeWindow(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Send zero-size window
	SendWindowSize(app, 0, 0)

	// Should not crash
	view := app.View()

	if len(view) == 0 {
		t.Log("View might be empty with zero size window")
	}
}

func TestIntegration_EdgeCases_VerySmallWindow(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Send very small window size
	SendWindowSize(app, 10, 5)

	// Should not crash
	view := app.View()

	if len(view) == 0 {
		t.Log("View might be minimal with small window")
	}
}

func TestIntegration_EdgeCases_InvalidSelectedIndex(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	if len(app.indices) == 0 {
		t.Skip("No indices in fixtures")
	}

	// Manually set invalid selectedIndex
	app.currentView = ViewIndices
	app.activePanel = PanelRight
	app.selectedIndex = 9999

	// Try to drill down
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not crash (command should be nil or handled gracefully)
	if cmd != nil {
		msg := ExecuteCommand(cmd)
		if msg != nil {
			app.Update(msg)
		}
	}
}

func TestIntegration_EdgeCases_NilClient(t *testing.T) {
	// This tests the edge case of creating app with nil client
	// In normal operation this shouldn't happen, but good to verify behavior
	defer func() {
		if r := recover(); r != nil {
			t.Log("Creating app with nil client causes panic (expected)")
		}
	}()

	app := NewApp(nil, "http://localhost:9200")

	// If we get here without panic, try to use it
	if app != nil {
		t.Log("App created with nil client")
	}
}

// TestIntegration_EdgeCases_ConcurrentUpdates removed because Bubble Tea
// guarantees single-threaded Update() calls in real usage, so testing
// concurrent updates would be testing an impossible scenario.

func TestIntegration_EdgeCases_RefreshDuringViewChange(t *testing.T) {
	app, err := SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Start a refresh
	app.loading = true
	cmd := app.refresh()

	// While refresh is in progress, change view
	app.selectedItem = 2
	app.updateViewFromSelection()

	// Now complete the refresh
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Should not crash and should update data
	if app.err != nil {
		t.Errorf("Refresh during view change failed: %v", app.err)
	}
}

func TestIntegration_EdgeCases_MetricsWithNoData(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Navigate to metrics view without any metric data
	app.currentView = ViewLiveMetrics
	app.metricsEnabled = true

	// Render view
	app.updateViewportContent()
	view := app.viewport.View()

	// Should render without crashing (even with no metrics data)
	if len(view) == 0 {
		t.Log("Metrics view might be empty with no data")
	}
}

func TestIntegration_EdgeCases_DrilldownWithoutWindowSize(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Don't send WindowSizeMsg
	if len(app.indices) == 0 {
		t.Skip("No indices in fixtures")
	}

	app.currentView = ViewIndices
	app.activePanel = PanelRight
	app.selectedIndex = 0

	// Try to drill down without viewport initialized
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not crash
	if cmd != nil {
		msg := ExecuteCommand(cmd)
		if msg != nil {
			app.Update(msg)
		}
	}
}
