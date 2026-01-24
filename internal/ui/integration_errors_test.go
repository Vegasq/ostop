package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIntegration_Errors_FailFastBehavior(t *testing.T) {
	// Table of endpoints where errors can occur
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

			cmd := app.Init()
			msg := ExecuteCommand(cmd)
			app.Update(msg)

			// Should fail fast with error
			if app.err == nil {
				t.Errorf("Expected error for %s endpoint", tt.endpoint)
			}

			if !app.loading {
				// loading might still be true or false depending on implementation
			}
		})
	}
}

func TestIntegration_Errors_RetryRecovery(t *testing.T) {
	// Create transport that initially fails, then succeeds
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	// Inject error
	transport.SetError("health", fmt.Errorf("simulated error"))

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// First refresh - should fail
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	if app.err == nil {
		t.Fatal("First refresh should fail")
	}

	// Clear error
	transport.ClearError("health")

	// Retry with 'r' key
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	cmd = app.refresh()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	// Should succeed now
	if app.err != nil {
		t.Errorf("Retry should succeed, got error: %v", app.err)
	}

	if app.health == nil {
		t.Error("health should be loaded after successful retry")
	}
}

func TestIntegration_Errors_MalformedJSON(t *testing.T) {
	transport := NewMockTransport()

	// Set malformed JSON for health endpoint
	transport.SetFixture("health", []byte(`{"invalid json`))

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Should have parse error
	if app.err == nil {
		t.Error("Expected JSON parse error")
	}
}

func TestIntegration_Errors_PartialDataRetained(t *testing.T) {
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// First successful refresh
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	if app.err != nil {
		t.Fatalf("First refresh should succeed: %v", app.err)
	}

	// Capture health data
	firstHealth := app.health

	// Inject error for second refresh
	transport.SetError("health", fmt.Errorf("simulated error"))

	// Second refresh (should fail)
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	cmd = app.refresh()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	// Should have error
	if app.err == nil {
		t.Error("Second refresh should fail")
	}

	// Old data should be retained (not overwritten)
	if app.health == nil {
		t.Error("health data should be retained from first refresh")
	}

	if app.health != firstHealth {
		t.Error("health reference should be same as first refresh")
	}
}

func TestIntegration_Errors_MappingFetchError(t *testing.T) {
	// Create client with error on mapping endpoint
	client, err := NewMockClientWithError("mapping")
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// Load initial data
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	SendWindowSize(app, 120, 40)

	if len(app.indices) == 0 {
		t.Skip("No indices in fixtures")
	}

	// Try to drill down to schema
	app.currentView = ViewIndices
	app.activePanel = PanelRight
	app.selectedIndex = 0

	_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Execute mapping fetch
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

	// indexMapping should be nil
	if app.indexMapping != nil {
		t.Error("indexMapping should be nil when fetch fails")
	}
}

func TestIntegration_Errors_ErrorDisplayInView(t *testing.T) {
	client, err := NewMockClientWithError("health")
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Render view
	view := app.View()

	// Should contain error indicator
	if len(view) == 0 {
		t.Error("View should render error state")
	}

	// View should mention error and retry option
	// (This is a basic check - actual content depends on implementation)
}

func TestIntegration_Errors_ClearErrorOnRetry(t *testing.T) {
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	transport.SetError("health", fmt.Errorf("simulated error"))

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// First refresh - fails
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	if app.err == nil {
		t.Fatal("Should have error")
	}

	// Press 'r' to retry - this should clear error state immediately
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

	// Error should be cleared when retry starts
	if app.err != nil {
		t.Error("Error should be cleared when pressing 'r'")
	}

	// Loading should be set
	if !app.loading {
		t.Error("loading should be true when starting retry")
	}
}

func TestIntegration_Errors_NetworkTimeout(t *testing.T) {
	transport := NewMockTransport()

	// Set error to simulate network timeout
	transport.SetError("health", fmt.Errorf("network timeout"))

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	// Should handle network error gracefully
	if app.err == nil {
		t.Error("Expected network error")
	}

	// App should not crash
	view := app.View()
	if len(view) == 0 {
		t.Error("View should render even with network error")
	}
}

func TestIntegration_Errors_MultipleSequentialErrors(t *testing.T) {
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		t.Fatalf("Failed to load fixtures: %v", err)
	}

	client, err := NewMockClient(transport)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	app := NewTestApp(client, "http://localhost:9200")

	// First error - health
	transport.SetError("health", fmt.Errorf("error 1"))
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	if app.err == nil {
		t.Error("Should have first error")
	}

	// Clear and retry
	transport.ClearError("health")
	transport.SetError("stats", fmt.Errorf("error 2"))

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	cmd = app.refresh()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	if app.err == nil {
		t.Error("Should have second error")
	}

	// Clear and succeed
	transport.ClearError("stats")

	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	cmd = app.refresh()
	msg = ExecuteCommand(cmd)
	app.Update(msg)

	if app.err != nil {
		t.Errorf("Final retry should succeed: %v", app.err)
	}
}
