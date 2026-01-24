package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderAllocationView renders the disk allocation view
func (a *App) renderAllocationView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Disk Allocation (%d nodes)", len(a.allocation))))
	b.WriteString("\n\n")

	if len(a.allocation) == 0 {
		b.WriteString(labelStyle.Render("No allocation data available"))
		return b.String()
	}

	// Sort by disk usage percentage (descending) - show fullest nodes first
	type nodeWithPercent struct {
		info    AllocationInfo
		percent float64
	}

	var nodesWithPercent []nodeWithPercent
	for _, node := range a.allocation {
		var percent float64
		fmt.Sscanf(node.DiskPercent, "%f", &percent)
		nodesWithPercent = append(nodesWithPercent, nodeWithPercent{info: node, percent: percent})
	}

	// Simple bubble sort by percent (descending)
	for i := 0; i < len(nodesWithPercent); i++ {
		for j := i + 1; j < len(nodesWithPercent); j++ {
			if nodesWithPercent[j].percent > nodesWithPercent[i].percent {
				nodesWithPercent[i], nodesWithPercent[j] = nodesWithPercent[j], nodesWithPercent[i]
			}
		}
	}

	// Check for critical nodes
	criticalCount := 0
	warningCount := 0
	for _, nwp := range nodesWithPercent {
		if nwp.percent >= 90 {
			criticalCount++
		} else if nwp.percent >= 75 {
			warningCount++
		}
	}

	// Show warnings if any nodes in danger zone
	if criticalCount > 0 {
		b.WriteString(statusRed.Render(fmt.Sprintf("⚠ CRITICAL: %d node(s) at ≥90%% disk usage", criticalCount)))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Immediate action required: Add storage or delete data"))
		b.WriteString("\n\n")
	} else if warningCount > 0 {
		b.WriteString(statusYellow.Render(fmt.Sprintf("⚠ WARNING: %d node(s) at ≥75%% disk usage", warningCount)))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Plan for capacity expansion"))
		b.WriteString("\n\n")
	} else {
		b.WriteString(statusGreen.Render("✓ All nodes have adequate disk space"))
		b.WriteString("\n\n")
	}

	// Display nodes
	for _, nwp := range nodesWithPercent {
		node := nwp.info
		percent := nwp.percent

		var nodeNameStyle lipgloss.Style
		if percent >= 90 {
			nodeNameStyle = statusRed
		} else if percent >= 75 {
			nodeNameStyle = statusYellow
		} else {
			nodeNameStyle = statusGreen
		}

		b.WriteString(nodeNameStyle.Render(fmt.Sprintf("▶ %s", node.Node)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Shards:"), node.Shards))
		b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Used:"), node.DiskUsed))
		b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Available:"), node.DiskAvail))
		b.WriteString(fmt.Sprintf("    %s %s / %s\n", labelStyle.Render("Total:"), node.DiskUsed, node.DiskTotal))
		b.WriteString(fmt.Sprintf("    %s %s %s\n",
			labelStyle.Render("Usage:"),
			renderBar(node.DiskPercent, 20),
			valueStyle.Render(node.DiskPercent+"%")))
		b.WriteString("\n")
	}

	return b.String()
}

// renderThreadPoolView renders the thread pool statistics view
func (a *App) renderThreadPoolView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Thread Pool Statistics"))
	b.WriteString("\n\n")

	if len(a.threadPool) == 0 {
		b.WriteString(labelStyle.Render("No thread pool data available"))
		return b.String()
	}

	// Check for any rejections
	totalRejections := 0
	for _, tp := range a.threadPool {
		var rejected int
		fmt.Sscanf(tp.Rejected, "%d", &rejected)
		totalRejections += rejected
	}

	if totalRejections > 0 {
		b.WriteString(statusRed.Render(fmt.Sprintf("⚠ CRITICAL: %s thread pool rejections detected", formatNumber(int64(totalRejections)))))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Thread pool rejections indicate resource exhaustion and failed requests."))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Actions: Scale cluster, reduce load, or increase thread pool sizes."))
		b.WriteString("\n\n")
	} else {
		b.WriteString(statusGreen.Render("✓ No thread pool rejections"))
		b.WriteString("\n\n")
	}

	// Group by node
	byNode := make(map[string][]ThreadPoolInfo)
	for _, tp := range a.threadPool {
		byNode[tp.NodeName] = append(byNode[tp.NodeName], tp)
	}

	// Focus on key thread pools
	keyPools := map[string]bool{
		"search":     true,
		"write":      true,
		"get":        true,
		"bulk":       true,
		"management": true,
	}

	for nodeName, pools := range byNode {
		b.WriteString(valueStyle.Render(nodeName))
		b.WriteString("\n\n")

		for _, pool := range pools {
			// Skip non-key pools unless they have rejections
			var rejected int
			fmt.Sscanf(pool.Rejected, "%d", &rejected)

			if !keyPools[pool.Name] && rejected == 0 {
				continue
			}

			var poolNameStyle lipgloss.Style
			if rejected > 0 {
				poolNameStyle = statusRed
			} else {
				poolNameStyle = labelStyle
			}

			b.WriteString(fmt.Sprintf("  %s\n", poolNameStyle.Render(pool.Name)))
			b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Active:"), pool.Active))

			// Queue depth with color coding
			var queue int
			fmt.Sscanf(pool.Queue, "%d", &queue)
			var queueStyle lipgloss.Style
			if queue > 1000 {
				queueStyle = statusRed
			} else if queue > 100 {
				queueStyle = statusYellow
			} else {
				queueStyle = statusGreen
			}
			b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Queue:"), queueStyle.Render(pool.Queue)))

			// Rejections with emphasis if > 0
			if rejected > 0 {
				b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Rejected:"), statusRed.Render(pool.Rejected)))
			} else {
				b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Rejected:"), pool.Rejected))
			}

			b.WriteString(fmt.Sprintf("    %s %s / %s\n", labelStyle.Render("Threads:"), pool.Active, pool.Size))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderRecoveryView renders the shard recovery view
func (a *App) renderRecoveryView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Active Shard Recoveries (%d)", len(a.recovery))))
	b.WriteString("\n\n")

	if len(a.recovery) == 0 {
		b.WriteString(statusGreen.Render("✓ No active recoveries"))
		return b.String()
	}

	// Group by index
	byIndex := make(map[string][]RecoveryInfo)
	for _, rec := range a.recovery {
		byIndex[rec.Index] = append(byIndex[rec.Index], rec)
	}

	for index, recoveries := range byIndex {
		b.WriteString(valueStyle.Render(fmt.Sprintf("Index: %s", index)))
		b.WriteString("\n\n")

		for _, rec := range recoveries {
			b.WriteString(fmt.Sprintf("  %s %s → %s\n",
				labelStyle.Render("Shard:"), rec.Shard,
				labelStyle.Render(fmt.Sprintf("%s to %s", rec.SourceNode, rec.TargetNode))))

			b.WriteString(fmt.Sprintf("    %s %s  %s %s\n",
				labelStyle.Render("Type:"), rec.Type,
				labelStyle.Render("Stage:"), rec.Stage))

			// Parse and render progress bars
			var filesPercent, bytesPercent float64
			fmt.Sscanf(rec.FilesPercent, "%f", &filesPercent)
			fmt.Sscanf(rec.BytesPercent, "%f", &bytesPercent)

			b.WriteString(fmt.Sprintf("    %s %s %.1f%%\n",
				labelStyle.Render("Files:"),
				renderBar(rec.FilesPercent, 15),
				filesPercent))

			b.WriteString(fmt.Sprintf("    %s %s %.1f%%\n",
				labelStyle.Render("Bytes:"),
				renderBar(rec.BytesPercent, 15),
				bytesPercent))

			b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Time:"), rec.Time))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderSegmentsView renders the Lucene segments view
func (a *App) renderSegmentsView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Lucene Segments (%d)", len(a.segments))))
	b.WriteString("\n\n")

	if len(a.segments) == 0 {
		b.WriteString(labelStyle.Render("No segment data available"))
		return b.String()
	}

	// Group by index and shard, count segments
	type shardKey struct {
		index  string
		shard  string
		prirep string
	}

	segmentCounts := make(map[shardKey]int)
	for _, seg := range a.segments {
		key := shardKey{index: seg.Index, shard: seg.Shard, prirep: seg.Prirep}
		segmentCounts[key]++
	}

	// Check for shards with too many segments
	criticalShards := []shardKey{}
	warningShards := []shardKey{}
	for key, count := range segmentCounts {
		if count > 50 {
			criticalShards = append(criticalShards, key)
		} else if count > 20 {
			warningShards = append(warningShards, key)
		}
	}

	if len(criticalShards) > 0 {
		b.WriteString(statusRed.Render(fmt.Sprintf("⚠ CRITICAL: %d shard(s) with >50 segments", len(criticalShards))))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("High segment counts impact search performance. Consider force merge."))
		b.WriteString("\n\n")
	} else if len(warningShards) > 0 {
		b.WriteString(statusYellow.Render(fmt.Sprintf("⚠ WARNING: %d shard(s) with >20 segments", len(warningShards))))
		b.WriteString("\n\n")
	} else {
		b.WriteString(statusGreen.Render("✓ Segment counts are healthy"))
		b.WriteString("\n\n")
	}

	// Display segment counts by index/shard
	currentIndex := ""
	for key, count := range segmentCounts {
		if key.index != currentIndex {
			if currentIndex != "" {
				b.WriteString("\n")
			}
			b.WriteString(valueStyle.Render(fmt.Sprintf("Index: %s", key.index)))
			b.WriteString("\n")
			currentIndex = key.index
		}

		var countStyle lipgloss.Style
		if count > 50 {
			countStyle = statusRed
		} else if count > 20 {
			countStyle = statusYellow
		} else {
			countStyle = statusGreen
		}

		typeLabel := "Primary"
		if key.prirep == "r" {
			typeLabel = "Replica"
		}

		b.WriteString(fmt.Sprintf("  %s %s (%s): %s segments\n",
			labelStyle.Render("Shard"), key.shard, typeLabel,
			countStyle.Render(fmt.Sprintf("%d", count))))
	}

	return b.String()
}

// renderFielddataView renders the fielddata cache usage view
func (a *App) renderFielddataView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Fielddata Cache Usage (%d entries)", len(a.fielddata))))
	b.WriteString("\n\n")

	if len(a.fielddata) == 0 {
		b.WriteString(statusGreen.Render("✓ No fielddata cache usage"))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Fielddata cache is empty or not in use"))
		return b.String()
	}

	// Group by node, show top fields
	byNode := make(map[string][]FielddataInfo)
	for _, fd := range a.fielddata {
		byNode[fd.Node] = append(byNode[fd.Node], fd)
	}

	for nodeName, fields := range byNode {
		b.WriteString(valueStyle.Render(fmt.Sprintf("Node: %s", nodeName)))
		b.WriteString("\n\n")

		// Show top 20 fields by size
		displayCount := len(fields)
		if displayCount > 20 {
			displayCount = 20
		}

		for i := 0; i < displayCount; i++ {
			field := fields[i]
			b.WriteString(fmt.Sprintf("  %s %s\n",
				labelStyle.Render("Field:"), field.Field))
			b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Size:"), field.Size))
			b.WriteString("\n")
		}

		if len(fields) > displayCount {
			b.WriteString(labelStyle.Render(fmt.Sprintf("  ... and %d more fields", len(fields)-displayCount)))
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

// renderPluginsView renders the installed plugins view
func (a *App) renderPluginsView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Installed Plugins (%d)", len(a.plugins))))
	b.WriteString("\n\n")

	if len(a.plugins) == 0 {
		b.WriteString(labelStyle.Render("No plugins installed"))
		return b.String()
	}

	// Check for version mismatches
	pluginVersions := make(map[string]map[string]int)
	for _, plugin := range a.plugins {
		if pluginVersions[plugin.Name] == nil {
			pluginVersions[plugin.Name] = make(map[string]int)
		}
		pluginVersions[plugin.Name][plugin.Version]++
	}

	mismatchCount := 0
	for pluginName, versions := range pluginVersions {
		if len(versions) > 1 {
			mismatchCount++
			b.WriteString(statusYellow.Render(fmt.Sprintf("⚠ Version mismatch for plugin: %s", pluginName)))
			b.WriteString("\n")
		}
	}

	if mismatchCount > 0 {
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Inconsistent plugin versions across nodes can cause issues"))
		b.WriteString("\n\n")
	}

	// Group by node
	byNode := make(map[string][]PluginInfo)
	for _, plugin := range a.plugins {
		byNode[plugin.ID] = append(byNode[plugin.ID], plugin)
	}

	for nodeID, plugins := range byNode {
		b.WriteString(valueStyle.Render(fmt.Sprintf("Node: %s", nodeID)))
		b.WriteString("\n\n")

		for _, plugin := range plugins {
			b.WriteString(fmt.Sprintf("  %s\n", valueStyle.Render(plugin.Name)))
			b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Component:"), plugin.Component))
			b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Version:"), plugin.Version))
			if plugin.Description != "" && plugin.Description != "null" {
				b.WriteString(fmt.Sprintf("    %s %s\n", labelStyle.Render("Description:"), plugin.Description))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderTemplatesView renders the index templates view
func (a *App) renderTemplatesView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Index Templates (%d)", len(a.templates))))
	b.WriteString("\n\n")

	if len(a.templates) == 0 {
		b.WriteString(labelStyle.Render("No index templates defined"))
		return b.String()
	}

	// Sort by order (higher order = higher precedence)
	type templateWithOrder struct {
		template TemplateInfo
		order    int
	}

	var templatesWithOrder []templateWithOrder
	for _, template := range a.templates {
		var order int
		fmt.Sscanf(template.Order, "%d", &order)
		templatesWithOrder = append(templatesWithOrder, templateWithOrder{template: template, order: order})
	}

	// Simple bubble sort by order (descending)
	for i := 0; i < len(templatesWithOrder); i++ {
		for j := i + 1; j < len(templatesWithOrder); j++ {
			if templatesWithOrder[j].order > templatesWithOrder[i].order {
				templatesWithOrder[i], templatesWithOrder[j] = templatesWithOrder[j], templatesWithOrder[i]
			}
		}
	}

	b.WriteString(labelStyle.Render("Templates are sorted by order (highest precedence first)"))
	b.WriteString("\n\n")

	// Display templates
	for _, two := range templatesWithOrder {
		template := two.template

		b.WriteString(valueStyle.Render(template.Name))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Patterns:"), template.IndexPatterns))
		b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Order:"), template.Order))
		if template.Version != "" && template.Version != "null" {
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Version:"), template.Version))
		}
		b.WriteString("\n")
	}

	return b.String()
}
