package ui

import (
	"strings"
	"testing"
)

// ==================== stylePanel Tests ====================

func TestApp_StylePanel_LeftPanelActive(t *testing.T) {
	app := &App{
		activePanel: PanelLeft,
	}

	result := app.stylePanel("test content", PanelLeft)

	// When panel is active, should use activePanelStyle
	// Just verify it doesn't crash and returns something
	if result == "" {
		t.Error("stylePanel() should return styled content")
	}
}

func TestApp_StylePanel_LeftPanelInactive(t *testing.T) {
	app := &App{
		activePanel: PanelRight,
	}

	result := app.stylePanel("test content", PanelLeft)

	// When panel is inactive, should use inactivePanelStyle
	// Just verify it doesn't crash and returns something
	if result == "" {
		t.Error("stylePanel() should return styled content")
	}
}

func TestApp_StylePanel_RightPanelActive(t *testing.T) {
	app := &App{
		activePanel: PanelRight,
	}

	result := app.stylePanel("test content", PanelRight)

	// When panel is active, should use activePanelStyle
	if result == "" {
		t.Error("stylePanel() should return styled content")
	}
}

func TestApp_StylePanel_RightPanelInactive(t *testing.T) {
	app := &App{
		activePanel: PanelLeft,
	}

	result := app.stylePanel("test content", PanelRight)

	// When panel is inactive, should use inactivePanelStyle
	if result == "" {
		t.Error("stylePanel() should return styled content")
	}
}

// ==================== renderLeftPanel Tests ====================

func TestApp_RenderLeftPanel_AllMenuItems(t *testing.T) {
	app := &App{
		selectedItem: 0,
	}

	result := app.renderLeftPanel()

	expectedItems := []string{
		"Cluster Overview",
		"Nodes",
		"Indices",
		"Shards",
		"Resources",
		"Allocation",
		"Thread Pools",
		"Tasks",
		"Pending Tasks",
		"Recovery",
		"Segments",
		"Fielddata",
		"Plugins",
		"Templates",
	}

	for _, item := range expectedItems {
		if !strings.Contains(result, item) {
			t.Errorf("renderLeftPanel() should contain menu item %q", item)
		}
	}
}

func TestApp_RenderLeftPanel_FirstItemSelected(t *testing.T) {
	app := &App{
		selectedItem: 0,
	}

	result := app.renderLeftPanel()

	// First item should have selection indicator
	if !strings.Contains(result, "▶") {
		t.Error("renderLeftPanel() should show selection indicator for selected item")
	}

	// Cluster Overview is the first item, should be selected
	if !strings.Contains(result, "Cluster Overview") {
		t.Error("renderLeftPanel() should contain 'Cluster Overview'")
	}
}

func TestApp_RenderLeftPanel_DifferentSelection(t *testing.T) {
	tests := []struct {
		name         string
		selectedItem int
		expectedItem string
	}{
		{"first_item", 0, "Cluster Overview"},
		{"second_item", 1, "Nodes"},
		{"third_item", 2, "Indices"},
		{"last_item", 13, "Templates"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				selectedItem: tt.selectedItem,
			}

			result := app.renderLeftPanel()

			// Should contain the expected item
			if !strings.Contains(result, tt.expectedItem) {
				t.Errorf("renderLeftPanel() should contain %q", tt.expectedItem)
			}

			// Should have exactly one selection indicator
			count := strings.Count(result, "▶")
			if count != 1 {
				t.Errorf("renderLeftPanel() should have exactly 1 selection indicator, got %d", count)
			}
		})
	}
}

// ==================== renderRightPanel Tests ====================

func TestApp_RenderRightPanel_ClusterView(t *testing.T) {
	app := &App{
		currentView: ViewCluster,
		health:      &ClusterHealth{ClusterName: "test-cluster", Status: "green"},
		stats:       &ClusterStats{ClusterName: "test-cluster"},
	}

	result := app.renderRightPanel()

	// Should call renderClusterView()
	if !strings.Contains(result, "test-cluster") {
		t.Error("renderRightPanel() should call renderClusterView() for ViewCluster")
	}
}

func TestApp_RenderRightPanel_NodesView(t *testing.T) {
	app := &App{
		currentView: ViewNodes,
		nodes:       []NodeInfo{{Name: "test-node"}},
	}

	result := app.renderRightPanel()

	// Should call renderNodesView()
	if !strings.Contains(result, "Nodes") {
		t.Error("renderRightPanel() should call renderNodesView() for ViewNodes")
	}
}

func TestApp_RenderRightPanel_IndicesView(t *testing.T) {
	app := &App{
		currentView: ViewIndices,
		indices:     []IndexInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderIndicesView()
	if !strings.Contains(result, "Indices") {
		t.Error("renderRightPanel() should call renderIndicesView() for ViewIndices")
	}
}

func TestApp_RenderRightPanel_ShardsView(t *testing.T) {
	app := &App{
		currentView: ViewShards,
		shards:      []ShardInfo{},
		nodes:       []NodeInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderShardsView()
	if !strings.Contains(result, "Shard Distribution") {
		t.Error("renderRightPanel() should call renderShardsView() for ViewShards")
	}
}

func TestApp_RenderRightPanel_ResourcesView(t *testing.T) {
	app := &App{
		currentView: ViewResources,
		nodes:       []NodeInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderResourcesView()
	if !strings.Contains(result, "Resource Utilization Dashboard") {
		t.Error("renderRightPanel() should call renderResourcesView() for ViewResources")
	}
}

func TestApp_RenderRightPanel_AllocationView(t *testing.T) {
	app := &App{
		currentView: ViewAllocation,
		allocation:  []AllocationInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderAllocationView()
	if !strings.Contains(result, "Disk Allocation") {
		t.Error("renderRightPanel() should call renderAllocationView() for ViewAllocation")
	}
}

func TestApp_RenderRightPanel_ThreadPoolView(t *testing.T) {
	app := &App{
		currentView: ViewThreadPool,
		threadPool:  []ThreadPoolInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderThreadPoolView()
	if !strings.Contains(result, "Thread Pool Statistics") {
		t.Error("renderRightPanel() should call renderThreadPoolView() for ViewThreadPool")
	}
}

func TestApp_RenderRightPanel_TasksView(t *testing.T) {
	app := &App{
		currentView: ViewTasks,
		tasks:       []TaskInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderTasksView()
	if !strings.Contains(result, "Running Tasks") {
		t.Error("renderRightPanel() should call renderTasksView() for ViewTasks")
	}
}

func TestApp_RenderRightPanel_PendingTasksView(t *testing.T) {
	app := &App{
		currentView: ViewPendingTasks,
		pendingTasks: []PendingTaskInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderPendingTasksView()
	if !strings.Contains(result, "Pending Cluster Tasks") {
		t.Error("renderRightPanel() should call renderPendingTasksView() for ViewPendingTasks")
	}
}

func TestApp_RenderRightPanel_RecoveryView(t *testing.T) {
	app := &App{
		currentView: ViewRecovery,
		recovery:    []RecoveryInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderRecoveryView()
	if !strings.Contains(result, "Active Shard Recoveries") {
		t.Error("renderRightPanel() should call renderRecoveryView() for ViewRecovery")
	}
}

func TestApp_RenderRightPanel_SegmentsView(t *testing.T) {
	app := &App{
		currentView: ViewSegments,
		segments:    []SegmentInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderSegmentsView()
	if !strings.Contains(result, "Lucene Segments") {
		t.Error("renderRightPanel() should call renderSegmentsView() for ViewSegments")
	}
}

func TestApp_RenderRightPanel_FielddataView(t *testing.T) {
	app := &App{
		currentView: ViewFielddata,
		fielddata:   []FielddataInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderFielddataView()
	if !strings.Contains(result, "Fielddata Cache Usage") {
		t.Error("renderRightPanel() should call renderFielddataView() for ViewFielddata")
	}
}

func TestApp_RenderRightPanel_PluginsView(t *testing.T) {
	app := &App{
		currentView: ViewPlugins,
		plugins:     []PluginInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderPluginsView()
	if !strings.Contains(result, "Installed Plugins") {
		t.Error("renderRightPanel() should call renderPluginsView() for ViewPlugins")
	}
}

func TestApp_RenderRightPanel_TemplatesView(t *testing.T) {
	app := &App{
		currentView: ViewTemplates,
		templates:   []TemplateInfo{},
	}

	result := app.renderRightPanel()

	// Should call renderTemplatesView()
	if !strings.Contains(result, "Index Templates") {
		t.Error("renderRightPanel() should call renderTemplatesView() for ViewTemplates")
	}
}

func TestApp_RenderRightPanel_IndexSchemaView(t *testing.T) {
	app := &App{
		currentView: ViewIndexSchema,
		indices:     []IndexInfo{{Index: "test-index"}},
	}

	result := app.renderRightPanel()

	// Should call renderIndexSchemaView()
	// This view shows field mappings
	if !strings.Contains(result, "Schema") || !strings.Contains(result, "Loading") {
		t.Error("renderRightPanel() should call renderIndexSchemaView() for ViewIndexSchema")
	}
}

func TestApp_RenderRightPanel_UnknownView(t *testing.T) {
	app := &App{
		currentView: View(999), // Invalid view
	}

	result := app.renderRightPanel()

	// Should return "Unknown view" for invalid view
	if !strings.Contains(result, "Unknown view") {
		t.Error("renderRightPanel() should return 'Unknown view' for invalid view types")
	}
}
