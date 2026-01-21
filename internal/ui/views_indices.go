package ui

import (
	"fmt"
	"strings"
)

// renderIndicesView renders the indices list
func (a *App) renderIndicesView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Indices (%d)", len(a.indices))))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Enter to view schema"))
	b.WriteString("\n\n")

	if len(a.indices) == 0 {
		b.WriteString(labelStyle.Render("No indices data available"))
		return b.String()
	}

	for i, idx := range a.indices {
		var idxStr strings.Builder

		// Show selection indicator
		if i == a.selectedIndex {
			idxStr.WriteString(statusGreen.Render("▶ "))
		} else {
			idxStr.WriteString("  ")
		}

		// Health indicator
		var healthIndicator string
		switch idx.Health {
		case "green":
			healthIndicator = statusGreen.Render("●")
		case "yellow":
			healthIndicator = statusYellow.Render("●")
		case "red":
			healthIndicator = statusRed.Render("●")
		default:
			healthIndicator = "○"
		}

		// Index name - bold if selected
		indexName := idx.Index
		if i == a.selectedIndex {
			indexName = valueStyle.Render(indexName)
		}
		idxStr.WriteString(fmt.Sprintf("%s %s\n", healthIndicator, indexName))

		// Stats
		idxStr.WriteString(fmt.Sprintf("    %s %s  ", labelStyle.Render("Docs:"), idx.DocsCount))
		idxStr.WriteString(fmt.Sprintf("%s %s  ", labelStyle.Render("Size:"), idx.StoreSize))
		idxStr.WriteString(fmt.Sprintf("%s %s/%s\n", labelStyle.Render("Shards:"), idx.Pri, idx.Rep))

		idxStr.WriteString("\n")
		b.WriteString(idxStr.String())
	}

	return b.String()
}

// renderShardsView renders the shard distribution view
func (a *App) renderShardsView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Shard Distribution (%d shards)", len(a.shards))))
	b.WriteString("\n\n")

	if len(a.shards) == 0 {
		b.WriteString(labelStyle.Render("No shard data available"))
		return b.String()
	}

	// Group shards by node
	shardsByNode := make(map[string][]ShardInfo)
	unassignedShards := []ShardInfo{}

	for _, shard := range a.shards {
		if shard.Node == "" {
			unassignedShards = append(unassignedShards, shard)
		} else {
			shardsByNode[shard.Node] = append(shardsByNode[shard.Node], shard)
		}
	}

	// Show per-node shard distribution
	for _, node := range a.nodes {
		nodeShards := shardsByNode[node.Name]
		primaryCount := 0
		replicaCount := 0

		// Group shards by index for this node
		shardsByIndex := make(map[string]struct {
			primaryShards int
			replicaShards int
		})

		for _, shard := range nodeShards {
			entry := shardsByIndex[shard.Index]
			if shard.Prirep == "p" {
				primaryCount++
				entry.primaryShards++
			} else {
				replicaCount++
				entry.replicaShards++
			}
			shardsByIndex[shard.Index] = entry
		}

		// Node header with IP if available
		nodeDisplay := node.Name
		if node.IP != "" {
			nodeDisplay = fmt.Sprintf("%s (%s)", node.Name, node.IP)
		}

		b.WriteString(valueStyle.Render(nodeDisplay))
		b.WriteString(fmt.Sprintf(" - %d shards ", len(nodeShards)))
		b.WriteString(fmt.Sprintf("(%s %d / %s %d)\n",
			statusGreen.Render("P:"), primaryCount,
			statusYellow.Render("R:"), replicaCount))

		// Show indices on this node
		if len(shardsByIndex) > 0 {
			// Get sorted list of indices
			indices := make([]string, 0, len(shardsByIndex))
			for index := range shardsByIndex {
				indices = append(indices, index)
			}
			// Simple bubble sort
			for i := 0; i < len(indices); i++ {
				for j := i + 1; j < len(indices); j++ {
					if indices[j] < indices[i] {
						indices[i], indices[j] = indices[j], indices[i]
					}
				}
			}

			// Display each index with its shard counts, one per line
			for _, index := range indices {
				entry := shardsByIndex[index]
				b.WriteString(fmt.Sprintf("    %s %s (P:%d/R:%d)\n",
					labelStyle.Render("•"),
					index,
					entry.primaryShards,
					entry.replicaShards))
			}
		}
		b.WriteString("\n")
	}

	// Show unassigned shards if any
	if len(unassignedShards) > 0 {
		b.WriteString("\n")
		b.WriteString(statusRed.Render(fmt.Sprintf("⚠ Unassigned Shards: %d", len(unassignedShards))))
		b.WriteString("\n\n")

		// Group unassigned by index
		unassignedByIndex := make(map[string]int)
		for _, shard := range unassignedShards {
			unassignedByIndex[shard.Index]++
		}

		for index, count := range unassignedByIndex {
			b.WriteString(fmt.Sprintf("  %s: %d shards\n", index, count))
		}
	}

	// Summary statistics
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("Shard Balance"))
	b.WriteString("\n")

	if len(shardsByNode) > 0 {
		totalShards := 0
		minShards := 999999
		maxShards := 0

		for _, shards := range shardsByNode {
			count := len(shards)
			totalShards += count
			if count < minShards {
				minShards = count
			}
			if count > maxShards {
				maxShards = count
			}
		}

		avgShards := float64(totalShards) / float64(len(shardsByNode))

		b.WriteString(fmt.Sprintf("%s %.1f\n", labelStyle.Render("Average per node:"), avgShards))
		b.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Min per node:"), minShards))
		b.WriteString(fmt.Sprintf("%s %d\n", labelStyle.Render("Max per node:"), maxShards))

		// Balance indicator
		imbalance := float64(maxShards - minShards)
		if imbalance/avgShards > 0.3 {
			b.WriteString("\n")
			b.WriteString(statusYellow.Render("⚠ Shard distribution is unbalanced"))
		}
	}

	return b.String()
}
