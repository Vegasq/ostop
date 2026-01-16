package ui

import (
	"strings"
)

// stylePanel applies the appropriate style based on active panel
func (a *App) stylePanel(content string, panel Panel) string {
	if a.activePanel == panel {
		return activePanelStyle.Render(content)
	}
	return inactivePanelStyle.Render(content)
}

// renderLeftPanel renders the navigation menu
func (a *App) renderLeftPanel() string {
	var b strings.Builder

	menuItems := []string{
		"Cluster Overview",
		"Nodes",
		"Indices",
		"Shards",
		"Resources",
		"Live Metrics",
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

	for i, item := range menuItems {
		if i == a.selectedItem {
			b.WriteString(selectedMenuItemStyle.Render("▶ " + item))
		} else {
			b.WriteString(menuItemStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderRightPanel renders the detail view based on current view
func (a *App) renderRightPanel() string {
	switch a.currentView {
	case ViewCluster:
		return a.renderClusterView()
	case ViewNodes:
		return a.renderNodesView()
	case ViewIndices:
		return a.renderIndicesView()
	case ViewShards:
		return a.renderShardsView()
	case ViewResources:
		return a.renderResourcesView()
	case ViewLiveMetrics:
		return a.renderMetricsView()
	case ViewIndexSchema:
		return a.renderIndexSchemaView()
	case ViewAllocation:
		return a.renderAllocationView()
	case ViewThreadPool:
		return a.renderThreadPoolView()
	case ViewTasks:
		return a.renderTasksView()
	case ViewPendingTasks:
		return a.renderPendingTasksView()
	case ViewRecovery:
		return a.renderRecoveryView()
	case ViewSegments:
		return a.renderSegmentsView()
	case ViewFielddata:
		return a.renderFielddataView()
	case ViewPlugins:
		return a.renderPluginsView()
	case ViewTemplates:
		return a.renderTemplatesView()
	default:
		return "Unknown view"
	}
}
