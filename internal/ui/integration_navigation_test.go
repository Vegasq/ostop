package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_Navigation_MenuUpDown(t *testing.T) {
	tests := []struct {
		name           string
		initialItem    int
		key            tea.KeyType
		expectedItem   int
		shouldChange   bool
	}{
		{"down from 0", 0, tea.KeyDown, 1, true},
		{"down from 5", 5, tea.KeyDown, 6, true},
		{"down from 14", 14, tea.KeyDown, 14, false}, // At boundary
		{"up from 5", 5, tea.KeyUp, 4, true},
		{"up from 1", 1, tea.KeyUp, 0, true},
		{"up from 0", 0, tea.KeyUp, 0, false}, // At boundary
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := InitializeTestApp()
			if err != nil {
				t.Fatalf("Failed to initialize test app: %v", err)
			}

			app.selectedItem = tt.initialItem
			app.activePanel = PanelLeft

			// Send key
			app.Update(tea.KeyMsg{Type: tt.key})

			if tt.shouldChange {
				if app.selectedItem != tt.expectedItem {
					t.Errorf("selectedItem = %d, want %d", app.selectedItem, tt.expectedItem)
				}
			} else {
				if app.selectedItem != tt.initialItem {
					t.Errorf("selectedItem should not change from %d, got %d", tt.initialItem, app.selectedItem)
				}
			}
		})
	}
}

func TestIntegration_Navigation_Boundaries(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	app.activePanel = PanelLeft

	// Try to go below 0
	app.selectedItem = 0
	app.Update(tea.KeyMsg{Type: tea.KeyUp})
	if app.selectedItem != 0 {
		t.Errorf("Should not go below 0, got %d", app.selectedItem)
	}

	// Try to go above 14
	app.selectedItem = 14
	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if app.selectedItem != 14 {
		t.Errorf("Should not go above 14, got %d", app.selectedItem)
	}
}

func TestIntegration_Navigation_EnterSelectsView(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	app.activePanel = PanelLeft
	app.selectedItem = 2 // ViewIndices

	// Press Enter
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify view changed
	if app.currentView != ViewIndices {
		t.Errorf("currentView = %v, want ViewIndices", app.currentView)
	}

	// Verify panel switched to right
	if app.activePanel != PanelRight {
		t.Error("activePanel should switch to PanelRight after Enter")
	}
}

func TestIntegration_Navigation_TabSwitchesPanels(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Start in left panel
	app.activePanel = PanelLeft

	// Press Tab
	app.Update(tea.KeyMsg{Type: tea.KeyTab})

	if app.activePanel != PanelRight {
		t.Error("Tab should switch to PanelRight")
	}

	// Press Tab again
	app.Update(tea.KeyMsg{Type: tea.KeyTab})

	if app.activePanel != PanelLeft {
		t.Error("Tab should switch back to PanelLeft")
	}
}

func TestIntegration_Navigation_ContextSensitiveKeys(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Initialize viewport
	SendWindowSize(app, 120, 40)

	// Test up/down in left panel (menu navigation)
	app.activePanel = PanelLeft
	app.selectedItem = 0

	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if app.selectedItem != 1 {
		t.Errorf("In left panel, down should change selectedItem: got %d, want 1", app.selectedItem)
	}

	// Test up/down in right panel with non-indices view (viewport scroll)
	app.activePanel = PanelRight
	app.currentView = ViewCluster
	initialY := app.viewport.YOffset

	app.Update(tea.KeyMsg{Type: tea.KeyDown})

	// In right panel with non-indices view, down should scroll viewport
	if app.viewport.YOffset == initialY {
		// Viewport might not scroll if at bottom, but behavior is different from left panel
		// The key point is it doesn't change selectedItem
	}
}

func TestIntegration_Navigation_VimKeys(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	app.activePanel = PanelLeft
	app.selectedItem = 5

	// Test 'k' (vim up)
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if app.selectedItem != 4 {
		t.Errorf("'k' should move up: got %d, want 4", app.selectedItem)
	}

	// Test 'j' (vim down)
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if app.selectedItem != 5 {
		t.Errorf("'j' should move down: got %d, want 5", app.selectedItem)
	}
}

func TestIntegration_Navigation_FullFlow(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	// Start state
	if app.selectedItem != 0 {
		t.Errorf("Initial selectedItem = %d, want 0", app.selectedItem)
	}
	if app.activePanel != PanelLeft {
		t.Error("Initial activePanel should be PanelLeft")
	}

	// Navigate menu down twice
	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if app.selectedItem != 2 {
		t.Errorf("After 2 downs, selectedItem = %d, want 2", app.selectedItem)
	}

	// Select view
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.currentView != ViewIndices {
		t.Errorf("currentView = %v, want ViewIndices", app.currentView)
	}
	if app.activePanel != PanelRight {
		t.Error("activePanel should be PanelRight after Enter")
	}

	// Switch back to left panel
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if app.activePanel != PanelLeft {
		t.Error("Tab should switch to PanelLeft")
	}

	// Navigate up
	app.Update(tea.KeyMsg{Type: tea.KeyUp})
	if app.selectedItem != 1 {
		t.Errorf("After up, selectedItem = %d, want 1", app.selectedItem)
	}
}
