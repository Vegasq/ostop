package ui

import (
	"strings"
	"testing"
)

func TestApp_RenderIndicesView_Empty(t *testing.T) {
	app := &App{
		indices: []IndexInfo{},
	}

	result := app.renderIndicesView()

	if !strings.Contains(result, "No indices data available") {
		t.Error("renderIndicesView() should show 'No indices data available' for empty list")
	}

	if !strings.Contains(result, "Indices (0)") {
		t.Error("renderIndicesView() should show count of 0")
	}
}

func TestApp_RenderIndicesView_SingleIndex(t *testing.T) {
	app := &App{
		indices: []IndexInfo{
			{
				Health:    "green",
				Index:     "test-index",
				DocsCount: "1000",
				StoreSize: "1.5mb",
				Pri:       "1",
				Rep:       "1",
			},
		},
		selectedIndex: 0,
	}

	result := app.renderIndicesView()

	expectedStrings := []string{
		"Indices (1)",
		"test-index",
		"Docs:",
		"1000",
		"Size:",
		"1.5mb",
		"Shards:",
		"1/1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("renderIndicesView() should contain %q, got:\n%s", expected, result)
		}
	}
}

func TestApp_RenderIndicesView_MultipleIndices(t *testing.T) {
	app := &App{
		indices: []IndexInfo{
			{Health: "green", Index: "index-1", DocsCount: "100", StoreSize: "1mb", Pri: "1", Rep: "0"},
			{Health: "yellow", Index: "index-2", DocsCount: "200", StoreSize: "2mb", Pri: "2", Rep: "1"},
			{Health: "red", Index: "index-3", DocsCount: "300", StoreSize: "3mb", Pri: "3", Rep: "2"},
		},
		selectedIndex: 1,
	}

	result := app.renderIndicesView()

	if !strings.Contains(result, "Indices (3)") {
		t.Error("renderIndicesView() should show count of 3")
	}

	// Check all indices are present
	expectedIndices := []string{"index-1", "index-2", "index-3"}
	for _, idx := range expectedIndices {
		if !strings.Contains(result, idx) {
			t.Errorf("renderIndicesView() should contain index %q", idx)
		}
	}

	// Check doc counts
	expectedCounts := []string{"100", "200", "300"}
	for _, count := range expectedCounts {
		if !strings.Contains(result, count) {
			t.Errorf("renderIndicesView() should contain doc count %q", count)
		}
	}
}

func TestApp_RenderIndicesView_HealthIndicators(t *testing.T) {
	tests := []struct {
		name   string
		health string
	}{
		{"green_health", "green"},
		{"yellow_health", "yellow"},
		{"red_health", "red"},
		{"unknown_health", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				indices: []IndexInfo{
					{Health: tt.health, Index: "test-index", DocsCount: "0", StoreSize: "0b", Pri: "1", Rep: "0"},
				},
			}

			result := app.renderIndicesView()

			// Just verify it produces output without errors
			if result == "" {
				t.Error("renderIndicesView() should produce output")
			}

			// Verify the index name is present
			if !strings.Contains(result, "test-index") {
				t.Error("renderIndicesView() should contain index name")
			}
		})
	}
}

func TestApp_RenderShardsView_Empty(t *testing.T) {
	app := &App{
		shards: []ShardInfo{},
		nodes:  []NodeInfo{},
	}

	result := app.renderShardsView()

	if !strings.Contains(result, "No shard data available") {
		t.Error("renderShardsView() should show 'No shard data available' for empty list")
	}

	if !strings.Contains(result, "Shard Distribution (0 shards)") {
		t.Error("renderShardsView() should show count of 0")
	}
}

func TestApp_RenderShardsView_AssignedShards(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{Name: "node-1"},
			{Name: "node-2"},
		},
		shards: []ShardInfo{
			{Index: "test-index", Shard: "0", Prirep: "p", State: "STARTED", Node: "node-1"},
			{Index: "test-index", Shard: "0", Prirep: "r", State: "STARTED", Node: "node-2"},
			{Index: "test-index", Shard: "1", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "test-index", Shard: "1", Prirep: "r", State: "STARTED", Node: "node-1"},
		},
	}

	result := app.renderShardsView()

	if !strings.Contains(result, "Shard Distribution (4 shards)") {
		t.Error("renderShardsView() should show correct shard count")
	}

	// Check node names appear
	if !strings.Contains(result, "node-1") {
		t.Error("renderShardsView() should contain node-1")
	}
	if !strings.Contains(result, "node-2") {
		t.Error("renderShardsView() should contain node-2")
	}

	// Check for primary/replica indicators
	if !strings.Contains(result, "P:") {
		t.Error("renderShardsView() should show primary count indicator")
	}
	if !strings.Contains(result, "R:") {
		t.Error("renderShardsView() should show replica count indicator")
	}

	// Check for balance section
	if !strings.Contains(result, "Shard Balance") {
		t.Error("renderShardsView() should show shard balance section")
	}
}

func TestApp_RenderShardsView_UnassignedShards(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{Name: "node-1"},
		},
		shards: []ShardInfo{
			{Index: "test-index", Shard: "0", Prirep: "p", State: "STARTED", Node: "node-1"},
			{Index: "test-index", Shard: "1", Prirep: "p", State: "UNASSIGNED", Node: ""},
			{Index: "test-index", Shard: "2", Prirep: "r", State: "UNASSIGNED", Node: ""},
		},
	}

	result := app.renderShardsView()

	if !strings.Contains(result, "Unassigned Shards") {
		t.Error("renderShardsView() should show unassigned shards section")
	}

	if !strings.Contains(result, "2") {
		t.Error("renderShardsView() should show count of 2 unassigned shards")
	}

	if !strings.Contains(result, "test-index") {
		t.Error("renderShardsView() should show index name for unassigned shards")
	}
}

func TestApp_RenderShardsView_ShardBalance(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{Name: "node-1"},
			{Name: "node-2"},
			{Name: "node-3"},
		},
		shards: []ShardInfo{
			// node-1: 1 shard
			{Index: "idx1", Shard: "0", Prirep: "p", State: "STARTED", Node: "node-1"},
			// node-2: 3 shards
			{Index: "idx2", Shard: "0", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "1", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "2", Prirep: "p", State: "STARTED", Node: "node-2"},
			// node-3: 2 shards
			{Index: "idx3", Shard: "0", Prirep: "p", State: "STARTED", Node: "node-3"},
			{Index: "idx3", Shard: "1", Prirep: "p", State: "STARTED", Node: "node-3"},
		},
	}

	result := app.renderShardsView()

	// Check balance statistics
	if !strings.Contains(result, "Average per node:") {
		t.Error("renderShardsView() should show average shards per node")
	}

	if !strings.Contains(result, "Min per node:") {
		t.Error("renderShardsView() should show minimum shards per node")
	}

	if !strings.Contains(result, "Max per node:") {
		t.Error("renderShardsView() should show maximum shards per node")
	}
}

func TestApp_RenderShardsView_ImbalanceWarning(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{Name: "node-1"},
			{Name: "node-2"},
		},
		shards: []ShardInfo{
			// node-1: 1 shard
			{Index: "idx1", Shard: "0", Prirep: "p", State: "STARTED", Node: "node-1"},
			// node-2: 10 shards (heavily imbalanced)
			{Index: "idx2", Shard: "0", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "1", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "2", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "3", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "4", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "5", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "6", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "7", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "8", Prirep: "p", State: "STARTED", Node: "node-2"},
			{Index: "idx2", Shard: "9", Prirep: "p", State: "STARTED", Node: "node-2"},
		},
	}

	result := app.renderShardsView()

	// Should show imbalance warning (max-min = 9, avg = 5.5, imbalance/avg = 1.64 > 0.3)
	if !strings.Contains(result, "unbalanced") {
		t.Error("renderShardsView() should show imbalance warning when distribution is uneven")
	}
}

func TestApp_RenderShardsView_PrimaryAndReplicaCounting(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{Name: "node-1"},
		},
		shards: []ShardInfo{
			{Index: "idx1", Shard: "0", Prirep: "p", State: "STARTED", Node: "node-1"},
			{Index: "idx1", Shard: "0", Prirep: "r", State: "STARTED", Node: "node-1"},
			{Index: "idx1", Shard: "1", Prirep: "p", State: "STARTED", Node: "node-1"},
			{Index: "idx1", Shard: "1", Prirep: "r", State: "STARTED", Node: "node-1"},
			{Index: "idx1", Shard: "2", Prirep: "p", State: "STARTED", Node: "node-1"},
		},
	}

	result := app.renderShardsView()

	// Should show 5 total shards with 3 primary and 2 replica
	if !strings.Contains(result, "5 shards") {
		t.Error("renderShardsView() should show total shard count of 5")
	}

	// The result should contain P: 3 and R: 2
	if !strings.Contains(result, "3") {
		t.Error("renderShardsView() should show primary count of 3")
	}

	if !strings.Contains(result, "2") {
		t.Error("renderShardsView() should show replica count of 2")
	}
}
