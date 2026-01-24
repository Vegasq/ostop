package ui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_RefreshCycle_AllDataLoaded(t *testing.T) {
	app, err := SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	// Verify initial state
	if !app.loading {
		t.Error("App should start with loading=true")
	}

	// Initialize and execute refresh
	cmd := app.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil command")
	}

	msg := ExecuteCommand(cmd)
	if msg == nil {
		t.Fatal("ExecuteCommand returned nil message")
	}

	app.Update(msg)

	// Verify state after refresh
	if app.loading {
		t.Error("loading should be false after refresh")
	}

	if app.err != nil {
		t.Errorf("unexpected error: %v", app.err)
	}

	// Verify all 14 data fields are populated
	if app.health == nil {
		t.Error("health not loaded")
	}
	if app.stats == nil {
		t.Error("stats not loaded")
	}
	if app.nodes == nil {
		t.Error("nodes not loaded")
	}
	if app.indices == nil {
		t.Error("indices not loaded")
	}
	if app.shards == nil {
		t.Error("shards not loaded")
	}
	if app.allocation == nil {
		t.Error("allocation not loaded")
	}
	if app.threadPool == nil {
		t.Error("threadPool not loaded")
	}
	if app.tasks == nil {
		t.Error("tasks not loaded")
	}
	if app.pendingTasks == nil {
		t.Error("pendingTasks not loaded")
	}
	if app.recovery == nil {
		t.Error("recovery not loaded")
	}
	if app.segments == nil {
		t.Error("segments not loaded")
	}
	if app.fielddata == nil {
		t.Error("fielddata not loaded")
	}
	if app.plugins == nil {
		t.Error("plugins not loaded")
	}
	if app.templates == nil {
		t.Error("templates not loaded")
	}

	// Verify lastRefresh is set
	if app.lastRefresh.IsZero() {
		t.Error("lastRefresh not set")
	}

	// Verify health data is correct
	if app.health.ClusterName != "test-cluster" {
		t.Errorf("health.ClusterName = %s, want test-cluster", app.health.ClusterName)
	}
}

func TestIntegration_RefreshCycle_StateTransitions(t *testing.T) {
	app, err := SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	// Check initial state
	if !app.loading {
		t.Error("Initial loading should be true")
	}

	// Execute refresh
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Check state after successful refresh
	if app.loading {
		t.Error("loading should be false after refresh completes")
	}

	if app.err != nil {
		t.Errorf("err should be nil after successful refresh, got: %v", app.err)
	}

	// Trigger manual refresh with 'r' key
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

	// Loading should be set to true immediately
	if !app.loading {
		t.Error("loading should be true after pressing 'r'")
	}

	if app.err != nil {
		t.Error("err should be cleared when starting refresh")
	}
}

func TestIntegration_RefreshCycle_ErrorAtEachStage(t *testing.T) {
	// Test errors at each of the 14 fetch points
	endpoints := []struct {
		name     string
		endpoint string
	}{
		{"health", "health"},
		{"stats", "stats"},
		{"nodes", "nodes"},
		{"indices", "indices"},
		{"shards", "shards"},
		{"allocation", "allocation"},
		{"threadpool", "threadpool"},
		{"tasks", "tasks"},
		{"pending_tasks", "pending_tasks"},
		{"recovery", "recovery"},
		{"segments", "segments"},
		{"fielddata", "fielddata"},
		{"plugins", "plugins"},
		{"templates", "templates"},
	}

	for _, tt := range endpoints {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMockClientWithError(tt.endpoint)
			if err != nil {
				t.Fatalf("Failed to create mock client: %v", err)
			}

			app := NewTestApp(client, "http://localhost:9200")

			// Execute refresh
			cmd := app.Init()
			msg := ExecuteCommand(cmd)
			app.Update(msg)

			// Verify error state
			if app.loading {
				t.Error("loading should be false after error")
			}

			if app.err == nil {
				t.Errorf("Expected error for %s endpoint", tt.endpoint)
			}
		})
	}
}

func TestIntegration_RefreshCycle_MultipleRefreshes(t *testing.T) {
	app, err := SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	// First refresh
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	firstRefreshTime := app.lastRefresh

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Second refresh
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	cmd = app.refresh()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	secondRefreshTime := app.lastRefresh

	// Verify timestamp updated
	if !secondRefreshTime.After(firstRefreshTime) {
		t.Errorf("lastRefresh should update on subsequent refreshes: first=%v, second=%v",
			firstRefreshTime, secondRefreshTime)
	}

	// Verify data still populated
	if app.health == nil {
		t.Error("health should still be populated after second refresh")
	}
}

func TestIntegration_RefreshCycle_LastRefreshTimestamp(t *testing.T) {
	app, err := SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	// Verify lastRefresh starts as zero
	if !app.lastRefresh.IsZero() {
		t.Error("lastRefresh should be zero initially")
	}

	before := time.Now()

	// Execute refresh
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	after := time.Now()

	// Verify lastRefresh is set and within expected range
	if app.lastRefresh.IsZero() {
		t.Error("lastRefresh should be set after refresh")
	}

	if app.lastRefresh.Before(before) || app.lastRefresh.After(after) {
		t.Errorf("lastRefresh timestamp %v is outside expected range [%v, %v]",
			app.lastRefresh, before, after)
	}
}

func TestIntegration_RefreshCycle_ErrorRecovery(t *testing.T) {
	// First, create a working client and do a successful refresh
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}
	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// Successful refresh
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	if app.err != nil {
		t.Fatalf("First refresh should succeed: %v", app.err)
	}

	// Now inject an error
	transport.SetError("health", fmt.Errorf("simulated error"))

	// Trigger another refresh
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	cmd = app.refresh()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	// Should have error now
	if app.err == nil {
		t.Error("Expected error after injecting health error")
	}

	// Old data should be retained
	if app.health == nil {
		t.Error("health data should be retained from previous successful refresh")
	}

	// Now clear the error and refresh again
	transport.ClearError("health")

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	cmd = app.refresh()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	// Should succeed again
	if app.err != nil {
		t.Errorf("Refresh should succeed after clearing error: %v", app.err)
	}
}
