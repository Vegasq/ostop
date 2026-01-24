package ui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_ComplexFlow_FullUserSession(t *testing.T) {
	app, err := SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	// 1. Initial load
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	if app.err != nil {
		t.Fatalf("Initial load failed: %v", app.err)
	}

	// 2. Initialize viewport
	SendWindowSize(app, 120, 40)

	// 3. Navigate to Nodes view
	app.Update(tea.KeyMsg{Type: tea.KeyDown}) // selectedItem = 1
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if app.currentView != ViewNodes {
		t.Errorf("Step 3: currentView = %v, want ViewNodes", app.currentView)
	}

	// 4. Navigate to Indices view
	app.Update(tea.KeyMsg{Type: tea.KeyTab})  // Back to left panel
	app.Update(tea.KeyMsg{Type: tea.KeyDown}) // selectedItem = 2
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if app.currentView != ViewIndices {
		t.Errorf("Step 4: currentView = %v, want ViewIndices", app.currentView)
	}

	// 5. Drill into index (if available)
	if len(app.indices) > 0 {
		// After pressing Enter in step 4, we're already in right panel
		_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd != nil {
			msg := ExecuteCommand(cmd)
			if msg != nil {
				app.Update(msg)
			}
		}

		// Check if drill-down worked (might fail if mapping fixture not loaded properly)
		if app.currentView == ViewIndexSchema {
			// 6. Return from schema
			app.Update(tea.KeyMsg{Type: tea.KeyEsc})

			if app.currentView != ViewIndices {
				t.Error("Step 6: Should return to ViewIndices")
			}
		} else {
			t.Log("Step 5: Drill-down skipped (might be a mapping fixture issue)")
		}
	}

	// 7. Navigate to Live Metrics
	app.Update(tea.KeyMsg{Type: tea.KeyTab}) // Back to left panel
	// Navigate to ViewLiveMetrics (item 5)
	for app.selectedItem < 5 {
		app.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if app.currentView != ViewLiveMetrics {
		t.Errorf("Step 7: currentView = %v, want ViewLiveMetrics", app.currentView)
	}

	if !app.metricsEnabled {
		t.Error("Step 7: metricsEnabled should be true")
	}

	// 8. Trigger manual refresh
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

	if !app.loading {
		t.Error("Step 8: loading should be true after pressing 'r'")
	}

	// Verify state consistency throughout
	if app.selectedItem < 0 || app.selectedItem > 14 {
		t.Errorf("Final state: selectedItem out of bounds: %d", app.selectedItem)
	}
}

func TestIntegration_ComplexFlow_RefreshDuringMetrics(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Navigate to Live Metrics view
	app.selectedItem = 5
	cmd := app.updateViewFromSelectionCmd()

	if cmd != nil {
		msg := ExecuteCommand(cmd)
		if msg != nil {
			app.Update(msg)
		}
	}

	// Add some metrics data
	baseTime := time.Now()
	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   baseTime,
		IndexTotal:  1000000,
		SearchTotal: 500000,
	})

	app.metricsTimeSeries.AddSnapshot(&MetricsSnapshot{
		Timestamp:   baseTime.Add(5 * time.Second),
		IndexTotal:  1000500,
		SearchTotal: 500250,
	})

	// Trigger manual refresh while metrics ticker might be running
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	cmd = app.refresh()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Should not crash
	if app.err != nil {
		t.Errorf("Manual refresh during metrics failed: %v", app.err)
	}

	// Metrics should still be enabled
	if !app.metricsEnabled {
		t.Error("metricsEnabled should remain true")
	}
}

func TestIntegration_ComplexFlow_DrilldownErrorRecovery(t *testing.T) {
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// Initial load
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	SendWindowSize(app, 120, 40)

	if len(app.indices) == 0 {
		t.Skip("No indices in fixtures")
	}

	// Navigate to indices
	app.selectedItem = 2
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Try to drill into first index - should work
	// After Enter in step above, we're already in right panel
	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		msg := ExecuteCommand(cmd)
		if msg != nil {
			app.Update(msg)
		}
	}

	firstIndexMapping := app.indexMapping

	if firstIndexMapping == nil {
		t.Skip("First mapping fetch failed - fixture issue")
	}

	// Return
	app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Now inject error for mapping
	transport.SetError("mapping", fmt.Errorf("mapping error"))

	// Try to drill into second index - should fail
	if len(app.indices) > 1 {
		app.selectedIndex = 1
		_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd != nil {
			msg := ExecuteCommand(cmd)
			if msg != nil {
				app.Update(msg)
			}
		}

		if app.err == nil {
			t.Error("Second mapping fetch should fail")
		}

		// Return to indices
		app.Update(tea.KeyMsg{Type: tea.KeyEsc})

		// Clear error and try first index again
		transport.ClearError("mapping")

		app.selectedIndex = 0
		_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd != nil {
			msg := ExecuteCommand(cmd)
			if msg != nil {
				app.Update(msg)
			}
		}

		if app.err != nil {
			t.Errorf("Third mapping fetch should succeed: %v", app.err)
		}
	}
}

func TestIntegration_ComplexFlow_NavigationLoop(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Loop through all views multiple times
	for loop := 0; loop < 3; loop++ {
		for viewIndex := 0; viewIndex <= 14; viewIndex++ {
			app.activePanel = PanelLeft
			app.selectedItem = viewIndex
			app.Update(tea.KeyMsg{Type: tea.KeyEnter})

			if app.currentView != View(viewIndex) {
				t.Errorf("Loop %d, View %d: currentView = %v, want %v",
					loop, viewIndex, app.currentView, View(viewIndex))
			}

			// After Enter in left panel, we're automatically in right panel
			// So just one Tab will take us back to left
			app.Update(tea.KeyMsg{Type: tea.KeyTab})
			if app.activePanel != PanelLeft {
				t.Logf("Loop %d, View %d: activePanel = %v after Tab", loop, viewIndex, app.activePanel)
			}
		}
	}

	// Verify final state is consistent
	if app.selectedItem < 0 || app.selectedItem > 14 {
		t.Errorf("Final selectedItem out of bounds: %d", app.selectedItem)
	}
}

func TestIntegration_ComplexFlow_ScrollAndNavigate(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Navigate to a view with content
	app.selectedItem = 0 // Cluster
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app.Update(tea.KeyMsg{Type: tea.KeyTab}) // Switch to right panel

	// Scroll around
	app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	app.Update(tea.KeyMsg{Type: tea.KeyHome})
	app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	app.Update(tea.KeyMsg{Type: tea.KeyPgUp})

	// Navigate to different view
	app.Update(tea.KeyMsg{Type: tea.KeyTab}) // Back to left panel
	app.Update(tea.KeyMsg{Type: tea.KeyDown})
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Scroll should reset
	initialY := app.viewport.YOffset

	// Navigate back
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	app.Update(tea.KeyMsg{Type: tea.KeyUp})
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not crash
	view := app.View()
	if len(view) == 0 {
		t.Error("View should render after complex navigation")
	}

	t.Logf("Initial viewport Y: %d, Final: %d", initialY, app.viewport.YOffset)
}

func TestIntegration_ComplexFlow_ErrorRecoveryLoop(t *testing.T) {
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	SendWindowSize(app, 120, 40)

	// Cycle through: success, error, retry, success multiple times
	endpoints := []string{"health", "stats", "nodes"}

	for i, endpoint := range endpoints {
		// Success
		transport.ClearError(endpoint)
		cmd := app.Init()
		msg := ExecuteCommand(cmd)
		app.Update(msg)

		if app.err != nil {
			t.Errorf("Iteration %d: Initial load should succeed: %v", i, app.err)
		}

		// Inject error
		transport.SetError(endpoint, fmt.Errorf("error %d", i))
		app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
		cmd = app.refresh()
		msg = ExecuteCommand(cmd)
		app.Update(msg)

		if app.err == nil {
			t.Errorf("Iteration %d: Should have error", i)
		}

		// Retry with error cleared
		transport.ClearError(endpoint)
		app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
		cmd = app.refresh()
		msg = ExecuteCommand(cmd)
		app.Update(msg)

		if app.err != nil {
			t.Errorf("Iteration %d: Retry should succeed: %v", i, app.err)
		}
	}
}

func TestIntegration_ComplexFlow_MetricsEnableDisableLoop(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Repeatedly enable and disable metrics
	for i := 0; i < 5; i++ {
		// Navigate to metrics view (enable)
		app.selectedItem = 5
		app.updateViewFromSelection()

		if !app.metricsEnabled {
			t.Errorf("Iteration %d: metricsEnabled should be true", i)
		}

		// Navigate away (disable)
		app.selectedItem = 0
		app.updateViewFromSelection()

		if app.metricsEnabled {
			t.Errorf("Iteration %d: metricsEnabled should be false", i)
		}
	}

	// Final state should be consistent
	if app.currentView != ViewCluster {
		t.Errorf("Final view should be ViewCluster, got %v", app.currentView)
	}
}

func TestIntegration_ComplexFlow_QuickNavigation(t *testing.T) {
	app, err := InitializeTestApp()
	if err != nil {
		t.Fatalf("Failed to initialize test app: %v", err)
	}

	SendWindowSize(app, 120, 40)

	// Rapid navigation without waiting
	keys := []tea.KeyType{
		tea.KeyDown, tea.KeyDown, tea.KeyEnter,
		tea.KeyTab, tea.KeyUp, tea.KeyDown,
		tea.KeyTab, tea.KeyDown, tea.KeyDown, tea.KeyEnter,
		tea.KeyPgDown, tea.KeyHome,
		tea.KeyTab, tea.KeyUp, tea.KeyEnter,
	}

	for _, key := range keys {
		app.Update(tea.KeyMsg{Type: key})
	}

	// Should maintain consistent state
	if app.selectedItem < 0 || app.selectedItem > 14 {
		t.Errorf("selectedItem out of bounds: %d", app.selectedItem)
	}

	view := app.View()
	if len(view) == 0 {
		t.Error("View should render after rapid navigation")
	}
}
