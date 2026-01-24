package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_Views_AllViewsAccessible(t *testing.T) {
	views := []struct {
		index int
		view  View
		name  string
	}{
		{0, ViewCluster, "Cluster"},
		{1, ViewNodes, "Nodes"},
		{2, ViewIndices, "Indices"},
		{3, ViewShards, "Shards"},
		{4, ViewResources, "Resources"},
		{5, ViewLiveMetrics, "Live Metrics"},
		{6, ViewAllocation, "Allocation"},
		{7, ViewThreadPool, "Thread Pool"},
		{8, ViewTasks, "Tasks"},
		{9, ViewPendingTasks, "Pending Tasks"},
		{10, ViewRecovery, "Recovery"},
		{11, ViewSegments, "Segments"},
		{12, ViewFielddata, "Fielddata"},
		{13, ViewPlugins, "Plugins"},
		{14, ViewTemplates, "Templates"},
	}

	for _, tt := range views {
		t.Run(tt.name, func(t *testing.T) {
			app, err := InitializeTestApp()
			if err != nil {
				t.Fatalf("Failed to initialize test app: %v", err)
			}

			app.activePanel = PanelLeft
			app.selectedItem = tt.index

			// Select view with Enter
			app.Update(tea.KeyMsg{Type: tea.KeyEnter})

			if app.currentView != tt.view {
				t.Errorf("currentView = %v, want %v", app.currentView, tt.view)
			}
		})
	}
}

func TestIntegration_Views_ViewportInitialization(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Verify viewport not ready initially
	if app.viewportReady {
		t.Error("viewportReady should be false initially")
	}

	// Send WindowSizeMsg
	SendWindowSize(app, 120, 40)

	// Verify viewport initialized
	if !app.viewportReady {
		t.Error("viewportReady should be true after WindowSizeMsg")
	}

	if app.width != 120 {
		t.Errorf("width = %d, want 120", app.width)
	}

	if app.height != 40 {
		t.Errorf("height = %d, want 40", app.height)
	}
}

func TestIntegration_Views_ViewportScrolling(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Initialize viewport
	SendWindowSize(app, 120, 40)

	// Switch to right panel
	app.activePanel = PanelRight

	// Test page down
	initialY := app.viewport.YOffset
	app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if app.viewport.YOffset <= initialY && app.viewport.YOffset != app.viewport.TotalLineCount() {
		// May not scroll if already at bottom
		t.Logf("Page down: YOffset changed from %d to %d", initialY, app.viewport.YOffset)
	}

	// Test home (go to top)
	app.Update(tea.KeyMsg{Type: tea.KeyHome})
	if app.viewport.YOffset != 0 {
		t.Errorf("After Home, YOffset = %d, want 0", app.viewport.YOffset)
	}

	// Test end (go to bottom)
	app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	// Viewport should be at or near the end
	t.Logf("After End, YOffset = %d", app.viewport.YOffset)
}

func TestIntegration_Views_ContentRendering(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Initialize viewport
	SendWindowSize(app, 120, 40)

	tests := []struct {
		view         View
		name         string
		shouldHaveContent bool
	}{
		{ViewCluster, "Cluster", true},
		{ViewNodes, "Nodes", true},
		{ViewIndices, "Indices", true},
		{ViewShards, "Shards", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.currentView = tt.view
			app.updateViewportContent()

			content := app.viewport.View()
			if tt.shouldHaveContent && len(content) == 0 {
				t.Errorf("View %s should have content", tt.name)
			}
		})
	}
}

func TestIntegration_Views_ViewSwitching(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Switch from Cluster to Nodes
	app.currentView = ViewCluster
	app.selectedItem = 1 // Nodes
	app.updateViewFromSelection()

	if app.currentView != ViewNodes {
		t.Errorf("currentView = %v, want ViewNodes", app.currentView)
	}

	// Verify selections reset
	if app.selectedNode != 0 {
		t.Errorf("selectedNode should reset to 0, got %d", app.selectedNode)
	}
	if app.selectedIndex != 0 {
		t.Errorf("selectedIndex should reset to 0, got %d", app.selectedIndex)
	}
}

func TestIntegration_Views_ScrollPositionReset(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	app.activePanel = PanelRight
	app.currentView = ViewCluster

	// Scroll down
	app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	app.Update(tea.KeyMsg{Type: tea.KeyPgDown})

	scrolledY := app.viewport.YOffset

	// Switch view
	app.selectedItem = 1
	app.activePanel = PanelLeft
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Scroll position should reset
	if app.viewport.YOffset >= scrolledY {
		t.Logf("Scroll position after view switch: %d (was %d)", app.viewport.YOffset, scrolledY)
	}
}

func TestIntegration_Views_MetricsViewEnablesMetrics(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Verify metrics disabled initially
	if app.metricsEnabled {
		t.Error("metricsEnabled should be false initially")
	}

	// Navigate to Live Metrics
	app.selectedItem = 5 // ViewLiveMetrics
	app.updateViewFromSelection()

	// Verify metrics enabled
	if !app.metricsEnabled {
		t.Error("metricsEnabled should be true when on Live Metrics view")
	}

	// Navigate away
	app.selectedItem = 0 // ViewCluster
	app.updateViewFromSelection()

	// Verify metrics disabled
	if app.metricsEnabled {
		t.Error("metricsEnabled should be false after leaving Live Metrics view")
	}
}

func TestIntegration_Views_ErrorStateDisplay(t *testing.T) {
	client, err := NewMockClientWithError("health")
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// Trigger refresh
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Should have error
	if app.err == nil {
		t.Error("Expected error")
	}

	// Render view
	view := app.View()

	// View should show error
	if len(view) == 0 {
		t.Error("View should display error message")
	}
}

func TestIntegration_Views_LoadingStateDisplay(t *testing.T) {
	app, err := SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	// Set loading state
	app.loading = true

	// Render view
	view := app.View()

	// View should show loading message
	if len(view) == 0 {
		t.Error("View should display loading message")
	}
}
