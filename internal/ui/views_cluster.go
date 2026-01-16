package ui

import (
	"fmt"
	"strings"
)

// renderClusterView renders the cluster overview
func (a *App) renderClusterView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Cluster Health"))
	b.WriteString("\n")

	if a.health != nil {
		// Status with color
		statusText := strings.ToUpper(a.health.Status)
		var styledStatus string
		switch a.health.Status {
		case "green":
			styledStatus = statusGreen.Render("● " + statusText)
		case "yellow":
			styledStatus = statusYellow.Render("● " + statusText)
		case "red":
			styledStatus = statusRed.Render("● " + statusText)
		default:
			styledStatus = statusText
		}

		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status:"), styledStatus))
		b.WriteString(fmt.Sprintf("%s %s\n\n", labelStyle.Render("Name:"), valueStyle.Render(a.health.ClusterName)))
	}

	// Nodes
	b.WriteString(headerStyle.Render("Nodes"))
	b.WriteString("\n")
	if a.health != nil {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Total:"), valueStyle.Render(fmt.Sprintf("%d", a.health.NumberOfNodes))))
		b.WriteString(fmt.Sprintf("%s %s\n\n", labelStyle.Render("Data:"), valueStyle.Render(fmt.Sprintf("%d", a.health.NumberOfDataNodes))))
	}

	// Shards
	b.WriteString(headerStyle.Render("Shards"))
	b.WriteString("\n")
	if a.health != nil {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Active:"), valueStyle.Render(fmt.Sprintf("%d", a.health.ActiveShards))))
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Primary:"), valueStyle.Render(fmt.Sprintf("%d", a.health.ActivePrimaryShards))))
		if a.health.RelocatingShards > 0 {
			b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Relocating:"), statusYellow.Render(fmt.Sprintf("%d", a.health.RelocatingShards))))
		}
		if a.health.InitializingShards > 0 {
			b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Initializing:"), statusYellow.Render(fmt.Sprintf("%d", a.health.InitializingShards))))
		}
		if a.health.UnassignedShards > 0 {
			b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Unassigned:"), statusRed.Render(fmt.Sprintf("%d", a.health.UnassignedShards))))
		}
		b.WriteString("\n")
	}

	// Indices
	b.WriteString(headerStyle.Render("Indices"))
	b.WriteString("\n")
	if a.stats != nil {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Count:"), valueStyle.Render(fmt.Sprintf("%d", a.stats.Indices.Count))))
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Documents:"), valueStyle.Render(formatNumber(a.stats.Indices.Docs.Count))))
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Total Size:"), valueStyle.Render(formatBytes(a.stats.Indices.Store.SizeInBytes))))
	}

	// Last refresh
	if !a.lastRefresh.IsZero() {
		b.WriteString("\n")
		b.WriteString(labelStyle.Render(fmt.Sprintf("Last refresh: %s", a.lastRefresh.Format("15:04:05"))))
	}

	return b.String()
}

// renderNodesView renders the nodes list
func (a *App) renderNodesView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Nodes (%d)", len(a.nodes))))
	b.WriteString("\n\n")

	if len(a.nodes) == 0 {
		b.WriteString(labelStyle.Render("No nodes data available"))
		return b.String()
	}

	// Categorize nodes
	var masterNodes []NodeInfo
	var dataNodes []NodeInfo
	var otherNodes []NodeInfo

	for _, node := range a.nodes {
		isMaster := strings.Contains(node.NodeRole, "m")
		isData := strings.Contains(node.NodeRole, "d")

		if isMaster && !isData {
			masterNodes = append(masterNodes, node)
		} else if isData {
			dataNodes = append(dataNodes, node)
		} else {
			otherNodes = append(otherNodes, node)
		}
	}

	// Node type summary
	b.WriteString(headerStyle.Render("Node Types"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Master/Controller:"), len(masterNodes)))
	b.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Data Nodes:"), len(dataNodes)))
	if len(otherNodes) > 0 {
		b.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Other:"), len(otherNodes)))
	}
	b.WriteString("\n")

	// Master/Controller nodes section
	if len(masterNodes) > 0 {
		b.WriteString(statusGreen.Render("═══ MASTER/CONTROLLER NODES ═══"))
		b.WriteString("\n\n")

		for _, node := range masterNodes {
			a.renderNode(&b, node)
		}
	}

	// Data nodes section
	if len(dataNodes) > 0 {
		b.WriteString(statusYellow.Render("═══ DATA NODES ═══"))
		b.WriteString("\n\n")

		for _, node := range dataNodes {
			a.renderNode(&b, node)
		}
	}

	// Other nodes section
	if len(otherNodes) > 0 {
		b.WriteString(labelStyle.Render("═══ OTHER NODES ═══"))
		b.WriteString("\n\n")

		for _, node := range otherNodes {
			a.renderNode(&b, node)
		}
	}

	return b.String()
}

// renderNode renders a single node's details
func (a *App) renderNode(b *strings.Builder, node NodeInfo) {
	var nodeStr strings.Builder

	// Node name with prominent role indicator
	nodeType := a.getNodeTypeLabel(node.NodeRole)
	isMaster := node.Master == "*"

	nodeStr.WriteString("  ")
	if isMaster {
		nodeStr.WriteString(statusGreen.Render("★ "))
	} else {
		nodeStr.WriteString("  ")
	}
	nodeStr.WriteString(valueStyle.Render(node.Name))
	nodeStr.WriteString(" ")
	nodeStr.WriteString(a.formatNodeRoleBadge(node.NodeRole))

	if isMaster {
		nodeStr.WriteString(statusGreen.Render(" [ACTIVE MASTER]"))
	}
	nodeStr.WriteString("\n")

	nodeStr.WriteString(fmt.Sprintf("      %s %s\n", labelStyle.Render("Type:"), nodeType))

	// Metrics with bar charts
	nodeStr.WriteString(fmt.Sprintf("      %s %s %s\n",
		labelStyle.Render("Heap:"),
		renderBar(node.HeapPercent, 18),
		valueStyle.Render(node.HeapPercent+"%")))

	nodeStr.WriteString(fmt.Sprintf("      %s %s %s\n",
		labelStyle.Render("CPU: "),
		renderBar(node.CPU, 18),
		valueStyle.Render(node.CPU+"%")))

	nodeStr.WriteString(fmt.Sprintf("      %s %s %s\n",
		labelStyle.Render("RAM: "),
		renderBar(node.RAMPercent, 18),
		valueStyle.Render(node.RAMPercent+"%")))

	if node.DiskUsedPercent != "" {
		nodeStr.WriteString(fmt.Sprintf("      %s %s %s (%s / %s)\n",
			labelStyle.Render("Disk:"),
			renderBar(node.DiskUsedPercent, 18),
			valueStyle.Render(node.DiskUsedPercent+"%"),
			node.DiskUsed,
			node.DiskTotal))
	}

	nodeStr.WriteString("\n")
	b.WriteString(nodeStr.String())
}

// getNodeTypeLabel returns a human-readable label for node type
func (a *App) getNodeTypeLabel(role string) string {
	types := []string{}

	if strings.Contains(role, "m") {
		types = append(types, "Master-eligible")
	}
	if strings.Contains(role, "d") {
		types = append(types, "Data")
	}
	if strings.Contains(role, "i") {
		types = append(types, "Ingest")
	}
	if strings.Contains(role, "c") {
		types = append(types, "Coordinating")
	}

	if len(types) == 0 {
		return "Unknown"
	}

	return strings.Join(types, ", ")
}

// formatNodeRoleBadge formats node roles as badges
func (a *App) formatNodeRoleBadge(role string) string {
	badges := []string{}
	if strings.Contains(role, "m") {
		badges = append(badges, "[M]")
	}
	if strings.Contains(role, "d") {
		badges = append(badges, "[D]")
	}
	if strings.Contains(role, "i") {
		badges = append(badges, "[I]")
	}
	if strings.Contains(role, "c") {
		badges = append(badges, "[C]")
	}
	return labelStyle.Render(strings.Join(badges, " "))
}
