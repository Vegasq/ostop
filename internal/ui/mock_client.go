package ui

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// MockTransport implements http.RoundTripper for mocking OpenSearch API calls
type MockTransport struct {
	mu        sync.Mutex
	fixtures  map[string][]byte  // endpoint pattern -> JSON data
	errors    map[string]error   // endpoint pattern -> error to return
	callCount map[string]int     // endpoint pattern -> call counter
}

// NewMockTransport creates a new mock HTTP transport
func NewMockTransport() *MockTransport {
	return &MockTransport{
		fixtures:  make(map[string][]byte),
		errors:    make(map[string]error),
		callCount: make(map[string]int),
	}
}

// RoundTrip implements the http.RoundTripper interface
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Determine which endpoint is being called based on URL path
	path := req.URL.Path
	endpoint := m.matchEndpoint(path)

	// Increment call count
	m.callCount[endpoint]++

	// Check if we should return an error
	if err, ok := m.errors[endpoint]; ok {
		return nil, err
	}

	// Get fixture data
	data, ok := m.fixtures[endpoint]
	if !ok {
		// Return 404 if no fixture found
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"not found"}`))),
			Header:     make(http.Header),
		}, nil
	}

	// Return successful response with fixture data
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     make(http.Header),
	}, nil
}

// matchEndpoint maps URL paths to endpoint names
func (m *MockTransport) matchEndpoint(path string) string {
	switch {
	case strings.Contains(path, "/_cluster/health"):
		return "health"
	case strings.Contains(path, "/_cluster/stats"):
		return "stats"
	case strings.Contains(path, "/_cat/nodes"):
		return "nodes"
	case strings.Contains(path, "/_cat/indices"):
		return "indices"
	case strings.Contains(path, "/_cat/shards"):
		return "shards"
	case strings.Contains(path, "/_cat/allocation"):
		return "allocation"
	case strings.Contains(path, "/_cat/thread_pool"):
		return "threadpool"
	case strings.Contains(path, "/_cat/tasks"):
		return "tasks"
	case strings.Contains(path, "/_cat/pending_tasks"):
		return "pending_tasks"
	case strings.Contains(path, "/_cat/recovery"):
		return "recovery"
	case strings.Contains(path, "/_cat/segments"):
		return "segments"
	case strings.Contains(path, "/_cat/fielddata"):
		return "fielddata"
	case strings.Contains(path, "/_cat/plugins"):
		return "plugins"
	case strings.Contains(path, "/_cat/templates"):
		return "templates"
	case strings.Contains(path, "/_mapping"):
		return "mapping"
	case strings.Contains(path, "/_stats"):
		return "metrics"
	default:
		return "unknown"
	}
}

// SetFixture sets fixture data for a specific endpoint
func (m *MockTransport) SetFixture(endpoint string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fixtures[endpoint] = data
}

// SetError sets an error to return for a specific endpoint
func (m *MockTransport) SetError(endpoint string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[endpoint] = err
}

// ClearError removes an error for a specific endpoint
func (m *MockTransport) ClearError(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.errors, endpoint)
}

// GetCallCount returns the number of times an endpoint was called
func (m *MockTransport) GetCallCount(endpoint string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount[endpoint]
}

// Reset clears all fixtures, errors, and call counts
func (m *MockTransport) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fixtures = make(map[string][]byte)
	m.errors = make(map[string]error)
	m.callCount = make(map[string]int)
}

// LoadAllFixtures loads all standard test fixtures into the transport
func (m *MockTransport) LoadAllFixtures() error {
	fixtureMap := map[string]string{
		"health":        "cluster_health.json",
		"stats":         "cluster_stats.json",
		"nodes":         "nodes.json",
		"indices":       "indices.json",
		"shards":        "shards.json",
		"allocation":    "allocation.json",
		"threadpool":    "threadpool.json",
		"tasks":         "tasks.json",
		"pending_tasks": "pending_tasks.json",
		"recovery":      "recovery.json",
		"segments":      "segments.json",
		"fielddata":     "fielddata.json",
		"plugins":       "plugins.json",
		"templates":     "templates.json",
		"mapping":       "index_mapping.json",
		"metrics":       "cluster_metrics.json",
	}

	for endpoint, filename := range fixtureMap {
		data, err := LoadFixture(filename)
		if err != nil {
			return fmt.Errorf("failed to load fixture %s: %w", filename, err)
		}
		m.SetFixture(endpoint, data)
	}

	return nil
}
