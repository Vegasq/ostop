package ui

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestMockTransport_NewMockTransport(t *testing.T) {
	transport := NewMockTransport()

	if transport == nil {
		t.Fatal("NewMockTransport returned nil")
	}

	if transport.fixtures == nil {
		t.Error("fixtures map should be initialized")
	}

	if transport.errors == nil {
		t.Error("errors map should be initialized")
	}

	if transport.callCount == nil {
		t.Error("callCount map should be initialized")
	}
}

func TestMockTransport_SetGetFixture(t *testing.T) {
	transport := NewMockTransport()

	testData := []byte(`{"test":"data"}`)
	transport.SetFixture("health", testData)

	// Verify fixture was set
	req := &http.Request{
		URL: &url.URL{Path: "/_cluster/health"},
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	if string(body) != string(testData) {
		t.Errorf("Body = %s, want %s", string(body), string(testData))
	}
}

func TestMockTransport_SetError(t *testing.T) {
	transport := NewMockTransport()

	expectedErr := fmt.Errorf("test error")
	transport.SetError("health", expectedErr)

	req := &http.Request{
		URL: &url.URL{Path: "/_cluster/health"},
	}

	_, err := transport.RoundTrip(req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("Error = %v, want %v", err, expectedErr)
	}
}

func TestMockTransport_CallCount(t *testing.T) {
	transport := NewMockTransport()
	transport.SetFixture("health", []byte(`{}`))

	req := &http.Request{
		URL: &url.URL{Path: "/_cluster/health"},
	}

	// Make 3 calls
	for i := 0; i < 3; i++ {
		_, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}
	}

	count := transport.GetCallCount("health")
	if count != 3 {
		t.Errorf("CallCount = %d, want 3", count)
	}
}

func TestMockTransport_MatchEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/_cluster/health", "health"},
		{"/_cluster/stats", "stats"},
		{"/_cat/nodes?format=json", "nodes"},
		{"/_cat/indices", "indices"},
		{"/_cat/shards", "shards"},
		{"/_cat/allocation", "allocation"},
		{"/_cat/thread_pool", "threadpool"},
		{"/_cat/tasks", "tasks"},
		{"/_cat/pending_tasks", "pending_tasks"},
		{"/_cat/recovery", "recovery"},
		{"/_cat/segments", "segments"},
		{"/_cat/fielddata", "fielddata"},
		{"/_cat/plugins", "plugins"},
		{"/_cat/templates", "templates"},
		{"/myindex/_mapping", "mapping"},
		{"/_stats", "metrics"},
		{"/unknown/path", "unknown"},
	}

	transport := NewMockTransport()

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := transport.matchEndpoint(tt.path)
			if result != tt.expected {
				t.Errorf("matchEndpoint(%s) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestMockTransport_NotFound(t *testing.T) {
	transport := NewMockTransport()

	req := &http.Request{
		URL: &url.URL{Path: "/_cluster/health"},
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestMockTransport_Reset(t *testing.T) {
	transport := NewMockTransport()

	// Set some data
	transport.SetFixture("health", []byte(`{}`))
	transport.SetError("stats", fmt.Errorf("test error"))

	// Make a call to increment counter
	req := &http.Request{
		URL: &url.URL{Path: "/_cluster/health"},
	}
	transport.RoundTrip(req)

	// Verify data is set
	if len(transport.fixtures) == 0 {
		t.Error("Fixtures should not be empty before reset")
	}
	if transport.GetCallCount("health") == 0 {
		t.Error("Call count should not be 0 before reset")
	}

	// Reset
	transport.Reset()

	// Verify everything is cleared
	if len(transport.fixtures) != 0 {
		t.Errorf("Fixtures should be empty after reset, got %d items", len(transport.fixtures))
	}
	if len(transport.errors) != 0 {
		t.Errorf("Errors should be empty after reset, got %d items", len(transport.errors))
	}
	if transport.GetCallCount("health") != 0 {
		t.Errorf("Call count should be 0 after reset, got %d", transport.GetCallCount("health"))
	}
}

func TestMockTransport_LoadAllFixtures(t *testing.T) {
	transport := NewMockTransport()

	err := transport.LoadAllFixtures()
	if err != nil {
		t.Fatalf("LoadAllFixtures failed: %v", err)
	}

	// Verify all expected fixtures are loaded
	expectedEndpoints := []string{
		"health", "stats", "nodes", "indices", "shards",
		"allocation", "threadpool", "tasks", "pending_tasks",
		"recovery", "segments", "fielddata", "plugins",
		"templates", "mapping", "metrics",
	}

	for _, endpoint := range expectedEndpoints {
		if _, ok := transport.fixtures[endpoint]; !ok {
			t.Errorf("Fixture for %s not loaded", endpoint)
		}
	}
}

func TestLoadFixture_AllFixtures(t *testing.T) {
	fixtures := []string{
		"cluster_health.json",
		"cluster_stats.json",
		"nodes.json",
		"indices.json",
		"shards.json",
		"allocation.json",
		"threadpool.json",
		"tasks.json",
		"pending_tasks.json",
		"recovery.json",
		"segments.json",
		"fielddata.json",
		"plugins.json",
		"templates.json",
		"index_mapping.json",
		"cluster_metrics.json",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			data, err := LoadFixture(fixture)
			if err != nil {
				t.Errorf("Failed to load %s: %v", fixture, err)
				return
			}

			if len(data) == 0 {
				t.Errorf("Fixture %s is empty", fixture)
			}
		})
	}
}

func TestLoadFixture_NotFound(t *testing.T) {
	_, err := LoadFixture("nonexistent.json")
	if err == nil {
		t.Error("Expected error for nonexistent fixture, got nil")
	}
}

func TestMockResponse_NewMockResponse(t *testing.T) {
	data := []byte(`{"test":"data"}`)
	resp := NewMockResponse(data, 200)

	if resp == nil {
		t.Fatal("NewMockResponse returned nil")
	}

	if resp.StatusCode() != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode())
	}

	if resp.IsError() {
		t.Error("IsError should be false for 200 status")
	}

	body, err := io.ReadAll(resp.Body())
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	if string(body) != string(data) {
		t.Errorf("Body = %s, want %s", string(body), string(data))
	}
}

func TestMockResponse_ErrorResponse(t *testing.T) {
	resp := NewMockResponse([]byte(`{}`), 500)

	if resp.StatusCode() != 500 {
		t.Errorf("StatusCode = %d, want 500", resp.StatusCode())
	}

	if !resp.IsError() {
		t.Error("IsError should be true for 500 status")
	}
}

func TestMockResponse_NewMockErrorResponse(t *testing.T) {
	resp := NewMockErrorResponse(404)

	if resp.StatusCode() != 404 {
		t.Errorf("StatusCode = %d, want 404", resp.StatusCode())
	}

	if !resp.IsError() {
		t.Error("IsError should be true for error response")
	}
}

func TestNewMockClient(t *testing.T) {
	transport := NewMockTransport()
	client, err := NewMockClient(transport)

	if err != nil {
		t.Fatalf("NewMockClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewMockClient returned nil client")
	}
}

func TestNewMockClientWithFixtures(t *testing.T) {
	client, err := NewMockClientWithFixtures()

	if err != nil {
		t.Fatalf("NewMockClientWithFixtures failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewMockClientWithFixtures returned nil client")
	}
}

func TestNewMockClientWithError(t *testing.T) {
	client, err := NewMockClientWithError("health")

	if err != nil {
		t.Fatalf("NewMockClientWithError failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewMockClientWithError returned nil client")
	}

	// Verify the client returns an error for the specified endpoint
	res, err := client.Cluster.Health()
	if err == nil {
		t.Error("Expected error from health endpoint, got nil")
	}
	if res != nil {
		res.Body.Close()
	}
}
