package ui

import (
	"strings"
	"testing"
)

func TestApp_GetNodeTypeLabel(t *testing.T) {
	app := &App{}

	tests := []struct {
		name     string
		role     string
		expected string
	}{
		{"master_only", "m", "Master-eligible"},
		{"data_only", "d", "Data"},
		{"ingest_only", "i", "Ingest"},
		{"coordinating_only", "c", "Coordinating"},
		{"master_data", "md", "Master-eligible, Data"},
		{"master_data_ingest", "mdi", "Master-eligible, Data, Ingest"},
		{"all_roles", "mdic", "Master-eligible, Data, Ingest, Coordinating"},
		{"data_ingest", "di", "Data, Ingest"},
		{"empty", "", "Unknown"},
		{"unknown_role", "xyz", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.getNodeTypeLabel(tt.role)
			if got != tt.expected {
				t.Errorf("getNodeTypeLabel(%q) = %q, want %q", tt.role, got, tt.expected)
			}
		})
	}
}

func TestApp_FormatNodeRoleBadge(t *testing.T) {
	app := &App{}

	tests := []struct {
		name     string
		role     string
		contains []string // Check for presence of badges
	}{
		{"master", "m", []string{"[M]"}},
		{"data", "d", []string{"[D]"}},
		{"ingest", "i", []string{"[I]"}},
		{"coordinating", "c", []string{"[C]"}},
		{"master_data", "md", []string{"[M]", "[D]"}},
		{"all_roles", "mdic", []string{"[M]", "[D]", "[I]", "[C]"}},
		{"empty", "", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.formatNodeRoleBadge(tt.role)
			// Strip ANSI codes for easier testing
			for _, badge := range tt.contains {
				if !strings.Contains(got, badge) {
					t.Errorf("formatNodeRoleBadge(%q) = %q, should contain %q", tt.role, got, badge)
				}
			}
		})
	}
}

func TestApp_RenderClusterView_NoData(t *testing.T) {
	app := &App{
		health: nil,
		stats:  nil,
	}

	result := app.renderClusterView()

	if !strings.Contains(result, "Cluster Health") {
		t.Error("renderClusterView() should contain 'Cluster Health' header")
	}
}

func TestApp_RenderClusterView_GreenStatus(t *testing.T) {
	stats := &ClusterStats{}
	stats.Indices.Count = 5
	stats.Indices.Docs.Count = 1000
	stats.Indices.Store.SizeInBytes = 1024 * 1024 * 100

	app := &App{
		health: &ClusterHealth{
			Status:              "green",
			ClusterName:         "test-cluster",
			NumberOfNodes:       3,
			NumberOfDataNodes:   2,
			ActiveShards:        10,
			ActivePrimaryShards: 5,
			RelocatingShards:    0,
			InitializingShards:  0,
			UnassignedShards:    0,
		},
		stats: stats,
	}

	result := app.renderClusterView()

	expectedStrings := []string{
		"Cluster Health",
		"GREEN",
		"test-cluster",
		"Nodes",
		"Shards",
		"Indices",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("renderClusterView() should contain %q", expected)
		}
	}
}

func TestApp_RenderClusterView_YellowStatus(t *testing.T) {
	app := &App{
		health: &ClusterHealth{
			Status:              "yellow",
			ClusterName:         "test-cluster",
			NumberOfNodes:       1,
			NumberOfDataNodes:   1,
			ActiveShards:        5,
			ActivePrimaryShards: 5,
			RelocatingShards:    1,
			InitializingShards:  0,
			UnassignedShards:    5,
		},
	}

	result := app.renderClusterView()

	if !strings.Contains(result, "YELLOW") {
		t.Error("renderClusterView() should show YELLOW status")
	}

	if !strings.Contains(result, "Relocating") {
		t.Error("renderClusterView() should show relocating shards when > 0")
	}

	if !strings.Contains(result, "Unassigned") {
		t.Error("renderClusterView() should show unassigned shards when > 0")
	}
}

func TestApp_RenderClusterView_RedStatus(t *testing.T) {
	app := &App{
		health: &ClusterHealth{
			Status:              "red",
			ClusterName:         "test-cluster",
			NumberOfNodes:       2,
			NumberOfDataNodes:   2,
			ActiveShards:        3,
			ActivePrimaryShards: 3,
			RelocatingShards:    0,
			InitializingShards:  2,
			UnassignedShards:    10,
		},
	}

	result := app.renderClusterView()

	if !strings.Contains(result, "RED") {
		t.Error("renderClusterView() should show RED status")
	}

	if !strings.Contains(result, "Initializing") {
		t.Error("renderClusterView() should show initializing shards when > 0")
	}
}

func TestApp_RenderNodesView_NoNodes(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{},
	}

	result := app.renderNodesView()

	if !strings.Contains(result, "No nodes data available") {
		t.Error("renderNodesView() with no nodes should show 'No nodes data available'")
	}
}

func TestApp_RenderNodesView_MasterNodes(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{
				Name:            "master-1",
				NodeRole:        "m",
				Master:          "*",
				HeapPercent:     "50",
				CPU:             "30",
				RAMPercent:      "60",
				DiskUsedPercent: "40",
			},
		},
	}

	result := app.renderNodesView()

	expectedStrings := []string{
		"Nodes (1)",
		"Node Types",
		"Master/Controller:",
		"MASTER/CONTROLLER NODES",
		"master-1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("renderNodesView() should contain %q", expected)
		}
	}
}

func TestApp_RenderNodesView_DataNodes(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{
				Name:            "data-1",
				NodeRole:        "d",
				Master:          "-",
				HeapPercent:     "70",
				CPU:             "50",
				RAMPercent:      "80",
				DiskUsedPercent: "60",
				DiskUsed:        "100GB",
				DiskTotal:       "250GB",
			},
		},
	}

	result := app.renderNodesView()

	expectedStrings := []string{
		"Nodes (1)",
		"Data Nodes:",
		"DATA NODES",
		"data-1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("renderNodesView() should contain %q", expected)
		}
	}
}

func TestApp_RenderNodesView_MixedNodes(t *testing.T) {
	app := &App{
		nodes: []NodeInfo{
			{
				Name:            "master-1",
				NodeRole:        "m",
				Master:          "*",
				HeapPercent:     "50",
				CPU:             "30",
				RAMPercent:      "60",
				DiskUsedPercent: "40",
			},
			{
				Name:            "data-1",
				NodeRole:        "md",
				Master:          "-",
				HeapPercent:     "70",
				CPU:             "50",
				RAMPercent:      "80",
				DiskUsedPercent: "60",
			},
			{
				Name:            "ingest-1",
				NodeRole:        "i",
				Master:          "-",
				HeapPercent:     "40",
				CPU:             "20",
				RAMPercent:      "50",
				DiskUsedPercent: "30",
			},
		},
	}

	result := app.renderNodesView()

	if !strings.Contains(result, "Nodes (3)") {
		t.Error("renderNodesView() should show total node count")
	}

	if !strings.Contains(result, "Other:") {
		t.Error("renderNodesView() should show 'Other' section when there are non-master/data nodes")
	}

	if !strings.Contains(result, "OTHER NODES") {
		t.Error("renderNodesView() should have OTHER NODES section")
	}
}

func TestApp_RenderNode_WithDiskInfo(t *testing.T) {
	app := &App{}
	var b strings.Builder

	node := NodeInfo{
		Name:            "test-node",
		NodeRole:        "md",
		Master:          "-",
		HeapPercent:     "75",
		CPU:             "60",
		RAMPercent:      "85",
		DiskUsedPercent: "70",
		DiskUsed:        "700GB",
		DiskTotal:       "1TB",
	}

	app.renderNode(&b, node)
	result := b.String()

	expectedStrings := []string{
		"test-node",
		"Heap:",
		"CPU:",
		"RAM:",
		"Disk:",
		"700GB",
		"1TB",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("renderNode() should contain %q", expected)
		}
	}
}

func TestApp_RenderNode_ActiveMaster(t *testing.T) {
	app := &App{}
	var b strings.Builder

	node := NodeInfo{
		Name:            "master-node",
		NodeRole:        "m",
		Master:          "*",
		HeapPercent:     "50",
		CPU:             "30",
		RAMPercent:      "60",
		DiskUsedPercent: "40",
	}

	app.renderNode(&b, node)
	result := b.String()

	if !strings.Contains(result, "ACTIVE MASTER") {
		t.Error("renderNode() should show ACTIVE MASTER for master nodes")
	}

	if !strings.Contains(result, "master-node") {
		t.Error("renderNode() should contain node name")
	}
}

func TestApp_RenderNode_NoDiskInfo(t *testing.T) {
	app := &App{}
	var b strings.Builder

	node := NodeInfo{
		Name:            "test-node",
		NodeRole:        "m",
		Master:          "-",
		HeapPercent:     "50",
		CPU:             "30",
		RAMPercent:      "60",
		DiskUsedPercent: "",
	}

	app.renderNode(&b, node)
	result := b.String()

	// Disk line should not be present when DiskUsedPercent is empty
	if strings.Contains(result, "Disk:") {
		t.Error("renderNode() should not show Disk info when DiskUsedPercent is empty")
	}
}
