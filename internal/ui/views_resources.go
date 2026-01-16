package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderResourcesView renders the resource utilization dashboard
func (a *App) renderResourcesView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("Resource Utilization Dashboard"))
	b.WriteString("\n\n")

	if len(a.nodes) == 0 {
		b.WriteString(labelStyle.Render("No node data available"))
		return b.String()
	}

	// Calculate aggregate metrics
	var totalHeap, totalCPU, totalRAM, totalDisk float64
	var minHeap, minCPU, minRAM, minDisk float64 = 100, 100, 100, 100
	var maxHeap, maxCPU, maxRAM, maxDisk float64
	nodeCount := float64(len(a.nodes))

	for _, node := range a.nodes {
		var heap, cpu, ram, disk float64
		fmt.Sscanf(node.HeapPercent, "%f", &heap)
		fmt.Sscanf(node.CPU, "%f", &cpu)
		fmt.Sscanf(node.RAMPercent, "%f", &ram)
		fmt.Sscanf(node.DiskUsedPercent, "%f", &disk)

		totalHeap += heap
		totalCPU += cpu
		totalRAM += ram
		totalDisk += disk

		if heap < minHeap {
			minHeap = heap
		}
		if heap > maxHeap {
			maxHeap = heap
		}
		if cpu < minCPU {
			minCPU = cpu
		}
		if cpu > maxCPU {
			maxCPU = cpu
		}
		if ram < minRAM {
			minRAM = ram
		}
		if ram > maxRAM {
			maxRAM = ram
		}
		if disk < minDisk {
			minDisk = disk
		}
		if disk > maxDisk {
			maxDisk = disk
		}
	}

	avgHeap := totalHeap / nodeCount
	avgCPU := totalCPU / nodeCount
	avgRAM := totalRAM / nodeCount
	avgDisk := totalDisk / nodeCount

	// Display aggregate metrics
	b.WriteString(headerStyle.Render("Cluster Averages"))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("%s %s %.1f%%\n",
		labelStyle.Render("JVM Heap:"),
		renderBar(fmt.Sprintf("%.1f", avgHeap), 20),
		avgHeap))

	b.WriteString(fmt.Sprintf("%s %s %.1f%%\n",
		labelStyle.Render("CPU:     "),
		renderBar(fmt.Sprintf("%.1f", avgCPU), 20),
		avgCPU))

	b.WriteString(fmt.Sprintf("%s %s %.1f%%\n",
		labelStyle.Render("RAM:     "),
		renderBar(fmt.Sprintf("%.1f", avgRAM), 20),
		avgRAM))

	b.WriteString(fmt.Sprintf("%s %s %.1f%%\n",
		labelStyle.Render("Disk:    "),
		renderBar(fmt.Sprintf("%.1f", avgDisk), 20),
		avgDisk))

	// Resource range
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("Resource Distribution"))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("%s %.1f%% - %.1f%% (Δ %.1f%%)\n",
		labelStyle.Render("Heap Range:"),
		minHeap, maxHeap, maxHeap-minHeap))

	b.WriteString(fmt.Sprintf("%s %.1f%% - %.1f%% (Δ %.1f%%)\n",
		labelStyle.Render("CPU Range: "),
		minCPU, maxCPU, maxCPU-minCPU))

	b.WriteString(fmt.Sprintf("%s %.1f%% - %.1f%% (Δ %.1f%%)\n",
		labelStyle.Render("RAM Range: "),
		minRAM, maxRAM, maxRAM-minRAM))

	b.WriteString(fmt.Sprintf("%s %.1f%% - %.1f%% (Δ %.1f%%)\n",
		labelStyle.Render("Disk Range:"),
		minDisk, maxDisk, maxDisk-minDisk))

	// Identify hotspots
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("Resource Hotspots"))
	b.WriteString("\n")

	hotspots := []string{}
	for _, node := range a.nodes {
		var heap, cpu, ram, disk float64
		fmt.Sscanf(node.HeapPercent, "%f", &heap)
		fmt.Sscanf(node.CPU, "%f", &cpu)
		fmt.Sscanf(node.RAMPercent, "%f", &ram)
		fmt.Sscanf(node.DiskUsedPercent, "%f", &disk)

		if heap >= 90 || cpu >= 90 || ram >= 90 || disk >= 90 {
			reason := ""
			if heap >= 90 {
				reason += fmt.Sprintf("Heap: %.0f%% ", heap)
			}
			if cpu >= 90 {
				reason += fmt.Sprintf("CPU: %.0f%% ", cpu)
			}
			if ram >= 90 {
				reason += fmt.Sprintf("RAM: %.0f%% ", ram)
			}
			if disk >= 90 {
				reason += fmt.Sprintf("Disk: %.0f%% ", disk)
			}
			hotspots = append(hotspots, fmt.Sprintf("%s - %s", node.Name, statusRed.Render(reason)))
		} else if heap >= 75 || cpu >= 75 || ram >= 75 || disk >= 85 {
			reason := ""
			if heap >= 75 {
				reason += fmt.Sprintf("Heap: %.0f%% ", heap)
			}
			if cpu >= 75 {
				reason += fmt.Sprintf("CPU: %.0f%% ", cpu)
			}
			if ram >= 75 {
				reason += fmt.Sprintf("RAM: %.0f%% ", ram)
			}
			if disk >= 85 {
				reason += fmt.Sprintf("Disk: %.0f%% ", disk)
			}
			hotspots = append(hotspots, fmt.Sprintf("%s - %s", node.Name, statusYellow.Render(reason)))
		}
	}

	if len(hotspots) > 0 {
		for _, hotspot := range hotspots {
			b.WriteString(fmt.Sprintf("  %s\n", hotspot))
		}
	} else {
		b.WriteString(statusGreen.Render("  ✓ All nodes within normal thresholds"))
		b.WriteString("\n")
	}

	// Capacity warnings
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("Capacity Planning"))
	b.WriteString("\n")

	warnings := []string{}
	if avgHeap >= 75 {
		warnings = append(warnings, "⚠ Average JVM heap is high - consider adding nodes or reducing load")
	}
	if avgDisk >= 80 {
		warnings = append(warnings, "⚠ Average disk usage is high - plan for storage expansion")
	}
	if maxHeap >= 90 {
		warnings = append(warnings, "⚠ At least one node has critical heap usage - immediate action needed")
	}
	if maxDisk >= 95 {
		warnings = append(warnings, "⚠ At least one node is nearly out of disk space - urgent action needed")
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			b.WriteString(statusYellow.Render(fmt.Sprintf("  %s\n", warning)))
		}
	} else {
		b.WriteString(statusGreen.Render("  ✓ Cluster capacity is healthy"))
		b.WriteString("\n")
	}

	return b.String()
}

// renderIndexSchemaView renders the schema/mapping for a specific index
func (a *App) renderIndexSchemaView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Index Schema: %s", a.selectedIndexName)))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Esc or Backspace to return to indices list"))
	b.WriteString("\n\n")

	if a.indexMapping == nil {
		b.WriteString(labelStyle.Render("Loading mapping..."))
		return b.String()
	}

	// Extract properties from mappings
	properties, ok := a.indexMapping.Mappings["properties"].(map[string]interface{})
	if !ok {
		b.WriteString(errorStyle.Render("No properties found in mapping"))
		return b.String()
	}

	// Count total fields
	totalFields := a.countFields(properties)
	b.WriteString(headerStyle.Render(fmt.Sprintf("Fields (%d)", totalFields)))
	b.WriteString("\n\n")

	// Render fields recursively
	a.renderFields(&b, properties, 0)

	return b.String()
}

// countFields recursively counts the number of fields in the mapping
func (a *App) countFields(properties map[string]interface{}) int {
	count := 0
	for _, value := range properties {
		count++
		fieldMap, ok := value.(map[string]interface{})
		if ok {
			if nestedProps, ok := fieldMap["properties"].(map[string]interface{}); ok {
				count += a.countFields(nestedProps)
			}
		}
	}
	return count
}

// renderFields recursively renders field information
func (a *App) renderFields(b *strings.Builder, properties map[string]interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)

	for fieldName, value := range properties {
		fieldMap, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		// Field name
		b.WriteString(indentStr)
		b.WriteString(valueStyle.Render(fieldName))
		b.WriteString("\n")

		// Field type
		if fieldType, ok := fieldMap["type"].(string); ok {
			b.WriteString(indentStr)
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Type:"), a.formatFieldType(fieldType)))
		}

		// Index property (searchability)
		indexable := true
		if indexProp, ok := fieldMap["index"].(bool); ok {
			indexable = indexProp
		}
		if indexProp, ok := fieldMap["index"].(string); ok {
			indexable = indexProp != "false"
		}

		searchIcon := statusGreen.Render("✓")
		searchText := "Searchable"
		if !indexable {
			searchIcon = statusRed.Render("✗")
			searchText = "Not searchable"
		}
		b.WriteString(indentStr)
		b.WriteString(fmt.Sprintf("  %s %s\n", searchIcon, labelStyle.Render(searchText)))

		// Analyzer
		if analyzer, ok := fieldMap["analyzer"].(string); ok {
			b.WriteString(indentStr)
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Analyzer:"), analyzer))
		}

		// Search analyzer
		if searchAnalyzer, ok := fieldMap["search_analyzer"].(string); ok {
			b.WriteString(indentStr)
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Search Analyzer:"), searchAnalyzer))
		}

		// Normalizer (for keyword fields)
		if normalizer, ok := fieldMap["normalizer"].(string); ok {
			b.WriteString(indentStr)
			b.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Normalizer:"), normalizer))
		}

		// Doc values
		if docValues, ok := fieldMap["doc_values"].(bool); ok {
			if !docValues {
				b.WriteString(indentStr)
				b.WriteString(fmt.Sprintf("  %s\n", labelStyle.Render("Doc values: disabled")))
			}
		}

		// Store
		if store, ok := fieldMap["store"].(bool); ok {
			if store {
				b.WriteString(indentStr)
				b.WriteString(fmt.Sprintf("  %s\n", labelStyle.Render("Stored: yes")))
			}
		}

		// Nested properties
		if nestedProps, ok := fieldMap["properties"].(map[string]interface{}); ok {
			b.WriteString(indentStr)
			b.WriteString(fmt.Sprintf("  %s\n", labelStyle.Render("Nested fields:")))
			a.renderFields(b, nestedProps, indent+2)
		}

		// Fields (multi-fields)
		if fields, ok := fieldMap["fields"].(map[string]interface{}); ok {
			b.WriteString(indentStr)
			b.WriteString(fmt.Sprintf("  %s\n", labelStyle.Render("Multi-fields:")))
			a.renderFields(b, fields, indent+2)
		}

		b.WriteString("\n")
	}
}

// formatFieldType formats the field type with color coding
func (a *App) formatFieldType(fieldType string) string {
	// Color code common types
	switch fieldType {
	case "text":
		return statusGreen.Render(fieldType)
	case "keyword":
		return statusYellow.Render(fieldType)
	case "long", "integer", "short", "byte", "double", "float", "half_float", "scaled_float":
		return statusGreen.Render(fieldType)
	case "date":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Render(fieldType)
	case "boolean":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Render(fieldType)
	case "object", "nested":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render(fieldType)
	default:
		return fieldType
	}
}
