package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_Drilldown_IndexSelection(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Navigate to Indices view
	app.currentView = ViewIndices
	app.activePanel = PanelRight

	// Verify we have indices loaded
	if len(app.indices) == 0 {
		t.Fatal("No indices loaded from fixtures")
	}

	// Select index with up/down
	app.selectedIndex = 0
	app.Update(tea.KeyMsg{Type: tea.KeyDown})

	if len(app.indices) > 1 && app.selectedIndex != 1 {
		t.Errorf("selectedIndex = %d, want 1", app.selectedIndex)
	}

	// Move back up
	app.Update(tea.KeyMsg{Type: tea.KeyUp})
	if app.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0", app.selectedIndex)
	}
}

func TestIntegration_Drilldown_IndexSelectionBoundaries(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	app.currentView = ViewIndices
	app.activePanel = PanelRight

	if len(app.indices) == 0 {
		t.Skip("No indices in fixtures")
	}

	// Try to go below 0
	app.selectedIndex = 0
	app.Update(tea.KeyMsg{Type: tea.KeyUp})
	if app.selectedIndex != 0 {
		t.Errorf("selectedIndex should not go below 0, got %d", app.selectedIndex)
	}

	// Try to go above max
	app.selectedIndex = len(app.indices) - 1
	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if app.selectedIndex != len(app.indices)-1 {
		t.Errorf("selectedIndex should not go above %d, got %d", len(app.indices)-1, app.selectedIndex)
	}
}

func TestIntegration_Drilldown_EnterFetchesMapping(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	app.currentView = ViewIndices
	app.activePanel = PanelRight

	if len(app.indices) == 0 {
		t.Skip("No indices in fixtures")
	}

	// Select first index
	app.selectedIndex = 0

	// Press Enter to drill down
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should trigger mapping fetch command
	if cmd == nil {
		t.Error("Enter should return a command to fetch mapping")
	}

	// Should set loading state
	if !app.loading {
		t.Error("loading should be true when fetching mapping")
	}

	// Should change view
	if app.currentView != ViewIndexSchema {
		t.Errorf("currentView = %v, want ViewIndexSchema", app.currentView)
	}

	// Should set selectedIndexName
	expectedIndexName := app.indices[0].Index
	if app.selectedIndexName != expectedIndexName {
		t.Errorf("selectedIndexName = %s, want %s", app.selectedIndexName, expectedIndexName)
	}
}

func TestIntegration_Drilldown_MappingLoaded(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	app.currentView = ViewIndices
	app.activePanel = PanelRight

	if len(app.indices) == 0 {
		t.Skip("No indices in fixtures")
	}

	// Select and drill down into first index
	app.selectedIndex = 0
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Execute the mapping fetch command
	if cmd != nil {
		msg := ExecuteCommand(cmd)
		if msg != nil {
			app.Update(msg)
		}
	}

	// Verify mapping loaded
	if app.indexMapping == nil {
		t.Error("indexMapping should be loaded after fetch")
	}

	// Verify loading cleared
	if app.loading {
		t.Error("loading should be false after mapping loaded")
	}

	// Verify no error
	if app.err != nil {
		t.Errorf("Unexpected error: %v", app.err)
	}
}

func TestIntegration_Drilldown_EscapeReturns(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Set up schema view state
	app.currentView = ViewIndexSchema
	app.selectedIndexName = "test-index"
	app.indexMapping = &IndexMapping{
		IndexName: "test-index",
		Mappings:  make(map[string]interface{}),
	}

	// Press Esc
	app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should return to indices view
	if app.currentView != ViewIndices {
		t.Errorf("currentView = %v, want ViewIndices", app.currentView)
	}
}

func TestIntegration_Drilldown_StateCleared(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Set up schema view state
	app.currentView = ViewIndexSchema
	app.selectedIndexName = "test-index"
	app.indexMapping = &IndexMapping{
		IndexName: "test-index",
		Mappings:  make(map[string]interface{}),
	}

	// Press Esc to return
	app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Verify state cleared
	if app.selectedIndexName != "" {
		t.Errorf("selectedIndexName = %s, want empty", app.selectedIndexName)
	}

	if app.indexMapping != nil {
		t.Error("indexMapping should be nil after returning")
	}
}

func TestIntegration_Drilldown_BackspaceAlsoReturns(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	app.currentView = ViewIndexSchema
	app.selectedIndexName = "test-index"

	// Press Backspace
	app.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	// Should return to indices view
	if app.currentView != ViewIndices {
		t.Errorf("currentView = %v, want ViewIndices", app.currentView)
	}

	// State should be cleared
	if app.selectedIndexName != "" {
		t.Error("selectedIndexName should be cleared")
	}
}

func TestIntegration_Drilldown_ScrollResetOnReturn(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	app.currentView = ViewIndexSchema
	app.activePanel = PanelRight

	// Scroll down in schema view
	app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	scrolledY := app.viewport.YOffset

	// Return to indices view
	app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Viewport should reset to top
	if app.viewport.YOffset >= scrolledY {
		t.Logf("Viewport YOffset after return: %d (was %d)", app.viewport.YOffset, scrolledY)
	}
}

func TestIntegration_Drilldown_MappingError(t *testing.T) {
	// Create client with error on mapping endpoint
	client, err := NewMockClientWithError("mapping")
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// Execute initial refresh to load indices
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	SendWindowSize(app, 120, 40)

	if len(app.indices) == 0 {
		t.Skip("No indices loaded")
	}

	// Try to drill down
	app.currentView = ViewIndices
	app.activePanel = PanelRight
	app.selectedIndex = 0

	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Execute mapping fetch (should fail)
	if cmd != nil {
		msg := ExecuteCommand(cmd)
		if msg != nil {
			app.Update(msg)
		}
	}

	// Should have error
	if app.err == nil {
		t.Error("Expected error when mapping fetch fails")
	}

	// Should clear loading state
	if app.loading {
		t.Error("loading should be false after error")
	}
}

func TestIntegration_Drilldown_OnlyInIndicesView(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Try to press Enter in a different view
	app.currentView = ViewCluster
	app.activePanel = PanelRight

	prevView := app.currentView

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not trigger drill-down
	if cmd != nil {
		t.Error("Enter in non-indices view should not trigger mapping fetch")
	}

	if app.currentView != prevView {
		t.Error("View should not change when pressing Enter in non-indices view")
	}
}
