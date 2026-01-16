package ui

import (
	"strings"
	"testing"
)

// ==================== Allocation View Tests ====================

func TestApp_RenderAllocationView_Empty(t *testing.T) {
	app := &App{
		allocation: []AllocationInfo{},
	}

	result := app.renderAllocationView()

	if !strings.Contains(result, "No allocation data available") {
		t.Error("renderAllocationView() should show 'No allocation data available' for empty list")
	}
}

func TestApp_RenderAllocationView_HealthyNodes(t *testing.T) {
	app := &App{
		allocation: []AllocationInfo{
			{Node: "node-1", Shards: "10", DiskUsed: "50gb", DiskAvail: "150gb", DiskTotal: "200gb", DiskPercent: "25"},
			{Node: "node-2", Shards: "12", DiskUsed: "60gb", DiskAvail: "140gb", DiskTotal: "200gb", DiskPercent: "30"},
		},
	}

	result := app.renderAllocationView()

	if !strings.Contains(result, "All nodes have adequate disk space") {
		t.Error("renderAllocationView() should show healthy status for nodes under 75% usage")
	}

	if !strings.Contains(result, "node-1") {
		t.Error("renderAllocationView() should contain node-1")
	}

	if !strings.Contains(result, "node-2") {
		t.Error("renderAllocationView() should contain node-2")
	}
}

func TestApp_RenderAllocationView_WarningLevel(t *testing.T) {
	app := &App{
		allocation: []AllocationInfo{
			{Node: "node-1", Shards: "10", DiskUsed: "160gb", DiskAvail: "40gb", DiskTotal: "200gb", DiskPercent: "80"},
			{Node: "node-2", Shards: "12", DiskUsed: "150gb", DiskAvail: "50gb", DiskTotal: "200gb", DiskPercent: "75"},
		},
	}

	result := app.renderAllocationView()

	if !strings.Contains(result, "WARNING") {
		t.Error("renderAllocationView() should show WARNING for nodes at ≥75% usage")
	}

	if !strings.Contains(result, "2 node(s) at ≥75%") {
		t.Error("renderAllocationView() should show count of nodes at warning level")
	}
}

func TestApp_RenderAllocationView_CriticalLevel(t *testing.T) {
	app := &App{
		allocation: []AllocationInfo{
			{Node: "node-1", Shards: "10", DiskUsed: "180gb", DiskAvail: "20gb", DiskTotal: "200gb", DiskPercent: "90"},
			{Node: "node-2", Shards: "12", DiskUsed: "195gb", DiskAvail: "5gb", DiskTotal: "200gb", DiskPercent: "97.5"},
		},
	}

	result := app.renderAllocationView()

	if !strings.Contains(result, "CRITICAL") {
		t.Error("renderAllocationView() should show CRITICAL for nodes at ≥90% usage")
	}

	if !strings.Contains(result, "2 node(s) at ≥90%") {
		t.Error("renderAllocationView() should show count of nodes at critical level")
	}

	if !strings.Contains(result, "Immediate action required") {
		t.Error("renderAllocationView() should show action guidance for critical state")
	}
}

func TestApp_RenderAllocationView_Sorting(t *testing.T) {
	app := &App{
		allocation: []AllocationInfo{
			{Node: "node-1", DiskPercent: "50"},
			{Node: "node-2", DiskPercent: "90"},
			{Node: "node-3", DiskPercent: "30"},
		},
	}

	result := app.renderAllocationView()

	// Find positions of nodes in output
	pos1 := strings.Index(result, "node-1")
	pos2 := strings.Index(result, "node-2")
	pos3 := strings.Index(result, "node-3")

	// node-2 (90%) should appear before node-1 (50%)
	if pos2 > pos1 {
		t.Error("renderAllocationView() should sort nodes by disk usage descending (highest first)")
	}

	// node-1 (50%) should appear before node-3 (30%)
	if pos1 > pos3 {
		t.Error("renderAllocationView() should sort nodes by disk usage descending")
	}
}

// ==================== Thread Pool View Tests ====================

func TestApp_RenderThreadPoolView_Empty(t *testing.T) {
	app := &App{
		threadPool: []ThreadPoolInfo{},
	}

	result := app.renderThreadPoolView()

	if !strings.Contains(result, "No thread pool data available") {
		t.Error("renderThreadPoolView() should show 'No thread pool data available' for empty list")
	}
}

func TestApp_RenderThreadPoolView_NoRejections(t *testing.T) {
	app := &App{
		threadPool: []ThreadPoolInfo{
			{NodeName: "node-1", Name: "search", Active: "5", Queue: "10", Rejected: "0", Size: "25"},
			{NodeName: "node-1", Name: "write", Active: "3", Queue: "5", Rejected: "0", Size: "20"},
		},
	}

	result := app.renderThreadPoolView()

	if !strings.Contains(result, "No thread pool rejections") {
		t.Error("renderThreadPoolView() should show healthy status when no rejections")
	}

	if !strings.Contains(result, "node-1") {
		t.Error("renderThreadPoolView() should contain node name")
	}

	if !strings.Contains(result, "search") {
		t.Error("renderThreadPoolView() should contain thread pool name")
	}
}

func TestApp_RenderThreadPoolView_WithRejections(t *testing.T) {
	app := &App{
		threadPool: []ThreadPoolInfo{
			{NodeName: "node-1", Name: "search", Active: "25", Queue: "1000", Rejected: "150", Size: "25"},
			{NodeName: "node-1", Name: "write", Active: "20", Queue: "500", Rejected: "75", Size: "20"},
		},
	}

	result := app.renderThreadPoolView()

	if !strings.Contains(result, "CRITICAL") {
		t.Error("renderThreadPoolView() should show CRITICAL when rejections are detected")
	}

	if !strings.Contains(result, "225") || !strings.Contains(result, "rejections") {
		t.Error("renderThreadPoolView() should show total rejection count")
	}

	if !strings.Contains(result, "150") {
		t.Error("renderThreadPoolView() should show individual pool rejection counts")
	}
}

func TestApp_RenderThreadPoolView_QueueDepthColorCoding(t *testing.T) {
	app := &App{
		threadPool: []ThreadPoolInfo{
			{NodeName: "node-1", Name: "search", Active: "5", Queue: "50", Rejected: "0", Size: "25"},
			{NodeName: "node-1", Name: "write", Active: "5", Queue: "500", Rejected: "0", Size: "25"},
			{NodeName: "node-1", Name: "bulk", Active: "5", Queue: "2000", Rejected: "0", Size: "25"},
		},
	}

	result := app.renderThreadPoolView()

	// Should contain queue depths
	if !strings.Contains(result, "50") {
		t.Error("renderThreadPoolView() should show queue depth for healthy pool")
	}

	if !strings.Contains(result, "500") {
		t.Error("renderThreadPoolView() should show queue depth for warning level pool")
	}

	if !strings.Contains(result, "2000") {
		t.Error("renderThreadPoolView() should show queue depth for critical level pool")
	}
}

// ==================== Tasks View Tests ====================

func TestApp_RenderTasksView_Empty(t *testing.T) {
	app := &App{
		tasks: []TaskInfo{},
	}

	result := app.renderTasksView()

	if !strings.Contains(result, "No running tasks") {
		t.Error("renderTasksView() should show 'No running tasks' for empty list")
	}
}

func TestApp_RenderTasksView_WithTasks(t *testing.T) {
	app := &App{
		tasks: []TaskInfo{
			{Action: "indices:data/read/search", Type: "transport", Node: "node-1", RunningTime: "1.5s"},
			{Action: "indices:data/write/bulk", Type: "transport", Node: "node-2", RunningTime: "2.3s"},
		},
	}

	result := app.renderTasksView()

	if !strings.Contains(result, "Running Tasks (2)") {
		t.Error("renderTasksView() should show task count")
	}

	// Check for simplified action names
	if !strings.Contains(result, "Search") || !strings.Contains(result, "Bulk") {
		t.Error("renderTasksView() should show simplified action names")
	}

	if !strings.Contains(result, "node-1") {
		t.Error("renderTasksView() should show node names")
	}
}

// ==================== Pending Tasks View Tests ====================

func TestApp_RenderPendingTasksView_Empty(t *testing.T) {
	app := &App{
		pendingTasks: []PendingTaskInfo{},
	}

	result := app.renderPendingTasksView()

	if !strings.Contains(result, "No pending tasks") {
		t.Error("renderPendingTasksView() should show 'No pending tasks' for empty list")
	}
}

func TestApp_RenderPendingTasksView_WithTasks(t *testing.T) {
	app := &App{
		pendingTasks: []PendingTaskInfo{
			{Priority: "URGENT", Source: "create-index", TimeInQueue: "500ms"},
			{Priority: "NORMAL", Source: "update-mapping", TimeInQueue: "100ms"},
		},
	}

	result := app.renderPendingTasksView()

	if !strings.Contains(result, "Pending Cluster Tasks (2)") {
		t.Error("renderPendingTasksView() should show task count")
	}

	if !strings.Contains(result, "URGENT") {
		t.Error("renderPendingTasksView() should show task priorities")
	}

	if !strings.Contains(result, "create-index") {
		t.Error("renderPendingTasksView() should show task sources")
	}
}

// ==================== Recovery View Tests ====================

func TestApp_RenderRecoveryView_Empty(t *testing.T) {
	app := &App{
		recovery: []RecoveryInfo{},
	}

	result := app.renderRecoveryView()

	if !strings.Contains(result, "No active recoveries") {
		t.Error("renderRecoveryView() should show 'No active recoveries' for empty list")
	}
}

func TestApp_RenderRecoveryView_WithRecovery(t *testing.T) {
	app := &App{
		recovery: []RecoveryInfo{
			{Index: "test-index", Shard: "0", Stage: "index", Type: "peer", SourceNode: "node-1", TargetNode: "node-2", FilesPercent: "50", BytesPercent: "45", Time: "5s"},
			{Index: "test-index", Shard: "1", Stage: "done", Type: "peer", SourceNode: "node-1", TargetNode: "node-3", FilesPercent: "100", BytesPercent: "100", Time: "10s"},
		},
	}

	result := app.renderRecoveryView()

	if !strings.Contains(result, "Active Shard Recoveries (2)") {
		t.Error("renderRecoveryView() should show recovery count")
	}

	if !strings.Contains(result, "test-index") {
		t.Error("renderRecoveryView() should show index names")
	}

	if !strings.Contains(result, "node-1") || !strings.Contains(result, "node-2") {
		t.Error("renderRecoveryView() should show source and target nodes")
	}

	if !strings.Contains(result, "index") || !strings.Contains(result, "done") {
		t.Error("renderRecoveryView() should show recovery stages")
	}
}

// ==================== Segments View Tests ====================

func TestApp_RenderSegmentsView_Empty(t *testing.T) {
	app := &App{
		segments: []SegmentInfo{},
	}

	result := app.renderSegmentsView()

	if !strings.Contains(result, "No segment data available") {
		t.Error("renderSegmentsView() should show 'No segment data available' for empty list")
	}
}

func TestApp_RenderSegmentsView_WithSegments(t *testing.T) {
	app := &App{
		segments: []SegmentInfo{
			{Index: "index-1", Shard: "0", Prirep: "p", Segment: "_0", Committed: "true", DocsCount: "1000", Size: "1mb"},
			{Index: "index-1", Shard: "1", Prirep: "p", Segment: "_1", Committed: "true", DocsCount: "2000", Size: "2mb"},
		},
	}

	result := app.renderSegmentsView()

	if !strings.Contains(result, "Lucene Segments") {
		t.Error("renderSegmentsView() should show 'Lucene Segments' header")
	}

	if !strings.Contains(result, "index-1") {
		t.Error("renderSegmentsView() should show index names")
	}

	// Should show segment counts, not doc counts
	if !strings.Contains(result, "Shard") {
		t.Error("renderSegmentsView() should show shard information")
	}

	if !strings.Contains(result, "1 segments") {
		t.Error("renderSegmentsView() should show segment counts per shard")
	}
}

// ==================== Fielddata View Tests ====================

func TestApp_RenderFielddataView_Empty(t *testing.T) {
	app := &App{
		fielddata: []FielddataInfo{},
	}

	result := app.renderFielddataView()

	if !strings.Contains(result, "No fielddata cache usage") {
		t.Error("renderFielddataView() should show 'No fielddata cache usage' for empty list")
	}
}

func TestApp_RenderFielddataView_WithFielddata(t *testing.T) {
	app := &App{
		fielddata: []FielddataInfo{
			{Node: "node-1", Field: "user.keyword", Size: "10mb"},
			{Node: "node-1", Field: "tags.keyword", Size: "5mb"},
		},
	}

	result := app.renderFielddataView()

	if !strings.Contains(result, "Fielddata Cache Usage") {
		t.Error("renderFielddataView() should show 'Fielddata Cache Usage' header")
	}

	if !strings.Contains(result, "node-1") {
		t.Error("renderFielddataView() should show node names")
	}

	if !strings.Contains(result, "user.keyword") {
		t.Error("renderFielddataView() should show field names")
	}

	if !strings.Contains(result, "10mb") {
		t.Error("renderFielddataView() should show memory sizes")
	}
}

// ==================== Plugins View Tests ====================

func TestApp_RenderPluginsView_Empty(t *testing.T) {
	app := &App{
		plugins: []PluginInfo{},
	}

	result := app.renderPluginsView()

	if !strings.Contains(result, "No plugins installed") {
		t.Error("renderPluginsView() should show 'No plugins installed' for empty list")
	}
}

func TestApp_RenderPluginsView_ConsistentVersions(t *testing.T) {
	app := &App{
		plugins: []PluginInfo{
			{ID: "node-1", Name: "analysis-icu", Version: "1.0.0", Component: "plugin"},
			{ID: "node-2", Name: "analysis-icu", Version: "1.0.0", Component: "plugin"},
		},
	}

	result := app.renderPluginsView()

	if !strings.Contains(result, "Installed Plugins") {
		t.Error("renderPluginsView() should show 'Installed Plugins' header")
	}

	if !strings.Contains(result, "analysis-icu") {
		t.Error("renderPluginsView() should show plugin name")
	}

	if !strings.Contains(result, "1.0.0") {
		t.Error("renderPluginsView() should show plugin version")
	}

	if strings.Contains(result, "mismatch") {
		t.Error("renderPluginsView() should not show mismatch warning for consistent versions")
	}
}

func TestApp_RenderPluginsView_VersionMismatch(t *testing.T) {
	app := &App{
		plugins: []PluginInfo{
			{ID: "node-1", Name: "analysis-icu", Version: "1.0.0", Component: "plugin"},
			{ID: "node-2", Name: "analysis-icu", Version: "2.0.0", Component: "plugin"},
		},
	}

	result := app.renderPluginsView()

	if !strings.Contains(result, "mismatch") {
		t.Error("renderPluginsView() should show version mismatch warning for inconsistent versions")
	}

	if !strings.Contains(result, "1.0.0") || !strings.Contains(result, "2.0.0") {
		t.Error("renderPluginsView() should show both versions in mismatch case")
	}
}

// ==================== Templates View Tests ====================

func TestApp_RenderTemplatesView_Empty(t *testing.T) {
	app := &App{
		templates: []TemplateInfo{},
	}

	result := app.renderTemplatesView()

	if !strings.Contains(result, "No index templates") {
		t.Error("renderTemplatesView() should show 'No index templates' for empty list")
	}
}

func TestApp_RenderTemplatesView_WithTemplates(t *testing.T) {
	app := &App{
		templates: []TemplateInfo{
			{Name: "logs-template", IndexPatterns: "logs-*", Order: "100", Version: "1"},
			{Name: "metrics-template", IndexPatterns: "metrics-*", Order: "50", Version: "2"},
		},
	}

	result := app.renderTemplatesView()

	if !strings.Contains(result, "Index Templates (2)") {
		t.Error("renderTemplatesView() should show template count")
	}

	if !strings.Contains(result, "logs-template") {
		t.Error("renderTemplatesView() should show template names")
	}

	if !strings.Contains(result, "logs-*") {
		t.Error("renderTemplatesView() should show index patterns")
	}

	if !strings.Contains(result, "100") {
		t.Error("renderTemplatesView() should show template order")
	}
}

func TestApp_RenderTemplatesView_Sorting(t *testing.T) {
	app := &App{
		templates: []TemplateInfo{
			{Name: "low-priority", Order: "10"},
			{Name: "high-priority", Order: "100"},
			{Name: "medium-priority", Order: "50"},
		},
	}

	result := app.renderTemplatesView()

	// Find positions in output
	posHigh := strings.Index(result, "high-priority")
	posMedium := strings.Index(result, "medium-priority")
	posLow := strings.Index(result, "low-priority")

	// Templates should be sorted by order descending (highest first)
	if posHigh > posMedium || posMedium > posLow {
		t.Error("renderTemplatesView() should sort templates by order descending")
	}
}
