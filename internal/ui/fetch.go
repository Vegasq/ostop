package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// refresh fetches cluster data in the background
func (a *App) refresh() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Fetch cluster health
		health, err := a.fetchClusterHealth(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch cluster stats
		stats, err := a.fetchClusterStats(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch nodes
		nodes, err := a.fetchNodes(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch indices
		indices, err := a.fetchIndices(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch shards
		shards, err := a.fetchShards(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch allocation
		allocation, err := a.fetchAllocation(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch thread pool
		threadPool, err := a.fetchThreadPool(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch tasks
		tasks, err := a.fetchTasks(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch pending tasks
		pendingTasks, err := a.fetchPendingTasks(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch recovery
		recovery, err := a.fetchRecovery(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch segments
		segments, err := a.fetchSegments(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch fielddata
		fielddata, err := a.fetchFielddata(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch plugins
		plugins, err := a.fetchPlugins(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		// Fetch templates
		templates, err := a.fetchTemplates(ctx)
		if err != nil {
			return refreshMsg{err: err}
		}

		return refreshMsg{
			health:       health,
			stats:        stats,
			nodes:        nodes,
			indices:      indices,
			shards:       shards,
			allocation:   allocation,
			threadPool:   threadPool,
			tasks:        tasks,
			pendingTasks: pendingTasks,
			recovery:     recovery,
			segments:     segments,
			fielddata:    fielddata,
			plugins:      plugins,
			templates:    templates,
		}
	}
}

// fetchClusterHealth calls the cluster health API
func (a *App) fetchClusterHealth(ctx context.Context) (*ClusterHealth, error) {
	res, err := a.client.Cluster.Health()
	if err != nil {
		return nil, fmt.Errorf("cluster health request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("cluster health API error: %s", res.Status())
	}

	var health ClusterHealth
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to parse cluster health: %w", err)
	}

	return &health, nil
}

// fetchClusterStats calls the cluster stats API
func (a *App) fetchClusterStats(ctx context.Context) (*ClusterStats, error) {
	res, err := a.client.Cluster.Stats()
	if err != nil {
		return nil, fmt.Errorf("cluster stats request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("cluster stats API error: %s", res.Status())
	}

	var stats ClusterStats
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to parse cluster stats: %w", err)
	}

	return &stats, nil
}

// fetchNodes calls the CAT nodes API
func (a *App) fetchNodes(ctx context.Context) ([]NodeInfo, error) {
	res, err := a.client.Cat.Nodes(
		a.client.Cat.Nodes.WithFormat("json"),
		a.client.Cat.Nodes.WithH("ip", "heap.percent", "ram.percent", "cpu", "load_1m", "load_5m", "load_15m", "node.role", "master", "name", "disk.used_percent", "disk.used", "disk.avail", "disk.total"),
	)
	if err != nil {
		return nil, fmt.Errorf("nodes request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("nodes API error: %s", res.Status())
	}

	var nodes []NodeInfo
	if err := json.NewDecoder(res.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("failed to parse nodes: %w", err)
	}

	return nodes, nil
}

// fetchIndices calls the CAT indices API
func (a *App) fetchIndices(ctx context.Context) ([]IndexInfo, error) {
	res, err := a.client.Cat.Indices(
		a.client.Cat.Indices.WithFormat("json"),
		a.client.Cat.Indices.WithH("health", "status", "index", "uuid", "pri", "rep", "docs.count", "docs.deleted", "store.size", "pri.store.size"),
	)
	if err != nil {
		return nil, fmt.Errorf("indices request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("indices API error: %s", res.Status())
	}

	var indices []IndexInfo
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, fmt.Errorf("failed to parse indices: %w", err)
	}

	return indices, nil
}

// fetchShards calls the CAT shards API
func (a *App) fetchShards(ctx context.Context) ([]ShardInfo, error) {
	res, err := a.client.Cat.Shards(
		a.client.Cat.Shards.WithFormat("json"),
		a.client.Cat.Shards.WithH("index", "shard", "prirep", "state", "docs", "store", "ip", "node"),
	)
	if err != nil {
		return nil, fmt.Errorf("shards request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("shards API error: %s", res.Status())
	}

	var shards []ShardInfo
	if err := json.NewDecoder(res.Body).Decode(&shards); err != nil {
		return nil, fmt.Errorf("failed to parse shards: %w", err)
	}

	return shards, nil
}

// fetchIndexMapping fetches the mapping for a specific index
func (a *App) fetchIndexMapping() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Call the indices get mapping API
		res, err := a.client.Indices.GetMapping(
			a.client.Indices.GetMapping.WithIndex(a.selectedIndexName),
			a.client.Indices.GetMapping.WithContext(ctx),
		)
		if err != nil {
			return mappingMsg{err: fmt.Errorf("mapping request failed: %w", err)}
		}
		defer res.Body.Close()

		if res.IsError() {
			return mappingMsg{err: fmt.Errorf("mapping API error: %s", res.Status())}
		}

		// Parse the response
		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return mappingMsg{err: fmt.Errorf("failed to parse mapping: %w", err)}
		}

		// Extract the mapping for the specific index
		// Response format: { "index_name": { "mappings": { ... } } }
		indexData, ok := response[a.selectedIndexName].(map[string]interface{})
		if !ok {
			return mappingMsg{err: fmt.Errorf("unexpected mapping response format")}
		}

		mappings, ok := indexData["mappings"].(map[string]interface{})
		if !ok {
			return mappingMsg{err: fmt.Errorf("no mappings found in response")}
		}

		mapping := &IndexMapping{
			IndexName: a.selectedIndexName,
			Mappings:  mappings,
		}

		return mappingMsg{mapping: mapping}
	}
}

// fetchAllocation calls the CAT allocation API
func (a *App) fetchAllocation(ctx context.Context) ([]AllocationInfo, error) {
	res, err := a.client.Cat.Allocation(
		a.client.Cat.Allocation.WithFormat("json"),
		a.client.Cat.Allocation.WithH("shards", "disk.indices", "disk.used", "disk.avail", "disk.total", "disk.percent", "host", "ip", "node"),
	)
	if err != nil {
		return nil, fmt.Errorf("allocation request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("allocation API error: %s", res.Status())
	}

	var allocation []AllocationInfo
	if err := json.NewDecoder(res.Body).Decode(&allocation); err != nil {
		return nil, fmt.Errorf("failed to parse allocation: %w", err)
	}

	return allocation, nil
}

// fetchThreadPool calls the CAT thread_pool API
func (a *App) fetchThreadPool(ctx context.Context) ([]ThreadPoolInfo, error) {
	res, err := a.client.Cat.ThreadPool(
		a.client.Cat.ThreadPool.WithFormat("json"),
		a.client.Cat.ThreadPool.WithH("node_name", "name", "active", "queue", "rejected", "completed", "size"),
	)
	if err != nil {
		return nil, fmt.Errorf("thread_pool request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("thread_pool API error: %s", res.Status())
	}

	var threadPool []ThreadPoolInfo
	if err := json.NewDecoder(res.Body).Decode(&threadPool); err != nil {
		return nil, fmt.Errorf("failed to parse thread_pool: %w", err)
	}

	return threadPool, nil
}

// fetchTasks calls the CAT tasks API
func (a *App) fetchTasks(ctx context.Context) ([]TaskInfo, error) {
	res, err := a.client.Cat.Tasks(
		a.client.Cat.Tasks.WithFormat("json"),
		a.client.Cat.Tasks.WithDetailed(true),
		a.client.Cat.Tasks.WithH("action", "task_id", "parent_task_id", "type", "start_time", "timestamp", "running_time", "ip", "node", "description"),
	)
	if err != nil {
		return nil, fmt.Errorf("tasks request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("tasks API error: %s", res.Status())
	}

	var tasks []TaskInfo
	if err := json.NewDecoder(res.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	return tasks, nil
}

// fetchPendingTasks calls the CAT pending_tasks API
func (a *App) fetchPendingTasks(ctx context.Context) ([]PendingTaskInfo, error) {
	res, err := a.client.Cat.PendingTasks(
		a.client.Cat.PendingTasks.WithFormat("json"),
		a.client.Cat.PendingTasks.WithH("insertOrder", "timeInQueue", "priority", "source"),
	)
	if err != nil {
		return nil, fmt.Errorf("pending_tasks request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("pending_tasks API error: %s", res.Status())
	}

	var pendingTasks []PendingTaskInfo
	if err := json.NewDecoder(res.Body).Decode(&pendingTasks); err != nil {
		return nil, fmt.Errorf("failed to parse pending_tasks: %w", err)
	}

	return pendingTasks, nil
}

// fetchRecovery calls the CAT recovery API
func (a *App) fetchRecovery(ctx context.Context) ([]RecoveryInfo, error) {
	res, err := a.client.Cat.Recovery(
		a.client.Cat.Recovery.WithFormat("json"),
		a.client.Cat.Recovery.WithActiveOnly(true),
		a.client.Cat.Recovery.WithH("index", "shard", "time", "type", "stage", "source_node", "target_node", "files", "files_recovered", "files_percent", "bytes", "bytes_recovered", "bytes_percent"),
	)
	if err != nil {
		return nil, fmt.Errorf("recovery request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("recovery API error: %s", res.Status())
	}

	var recovery []RecoveryInfo
	if err := json.NewDecoder(res.Body).Decode(&recovery); err != nil {
		return nil, fmt.Errorf("failed to parse recovery: %w", err)
	}

	return recovery, nil
}

// fetchSegments calls the CAT segments API
func (a *App) fetchSegments(ctx context.Context) ([]SegmentInfo, error) {
	res, err := a.client.Cat.Segments(
		a.client.Cat.Segments.WithFormat("json"),
		a.client.Cat.Segments.WithH("index", "shard", "prirep", "ip", "segment", "generation", "docs.count", "docs.deleted", "size", "committed"),
	)
	if err != nil {
		return nil, fmt.Errorf("segments request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("segments API error: %s", res.Status())
	}

	var segments []SegmentInfo
	if err := json.NewDecoder(res.Body).Decode(&segments); err != nil {
		return nil, fmt.Errorf("failed to parse segments: %w", err)
	}

	return segments, nil
}

// fetchFielddata calls the CAT fielddata API
func (a *App) fetchFielddata(ctx context.Context) ([]FielddataInfo, error) {
	res, err := a.client.Cat.Fielddata(
		a.client.Cat.Fielddata.WithFormat("json"),
		a.client.Cat.Fielddata.WithH("id", "host", "ip", "node", "field", "size"),
	)
	if err != nil {
		return nil, fmt.Errorf("fielddata request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("fielddata API error: %s", res.Status())
	}

	var fielddata []FielddataInfo
	if err := json.NewDecoder(res.Body).Decode(&fielddata); err != nil {
		return nil, fmt.Errorf("failed to parse fielddata: %w", err)
	}

	return fielddata, nil
}

// fetchPlugins calls the CAT plugins API
func (a *App) fetchPlugins(ctx context.Context) ([]PluginInfo, error) {
	res, err := a.client.Cat.Plugins(
		a.client.Cat.Plugins.WithFormat("json"),
		a.client.Cat.Plugins.WithH("id", "name", "component", "version", "description"),
	)
	if err != nil {
		return nil, fmt.Errorf("plugins request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("plugins API error: %s", res.Status())
	}

	var plugins []PluginInfo
	if err := json.NewDecoder(res.Body).Decode(&plugins); err != nil {
		return nil, fmt.Errorf("failed to parse plugins: %w", err)
	}

	return plugins, nil
}

// fetchTemplates calls the CAT templates API
func (a *App) fetchTemplates(ctx context.Context) ([]TemplateInfo, error) {
	res, err := a.client.Cat.Templates(
		a.client.Cat.Templates.WithFormat("json"),
		a.client.Cat.Templates.WithH("name", "index_patterns", "order", "version"),
	)
	if err != nil {
		return nil, fmt.Errorf("templates request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("templates API error: %s", res.Status())
	}

	var templates []TemplateInfo
	if err := json.NewDecoder(res.Body).Decode(&templates); err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return templates, nil
}

// fetchClusterMetrics retrieves current cumulative indexing and search metrics
func (a *App) fetchClusterMetrics(ctx context.Context) (*MetricsSnapshot, error) {
	// Use indices stats API to get cluster-wide metrics
	res, err := a.client.Indices.Stats(
		a.client.Indices.Stats.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("cluster metrics request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("cluster metrics API error: %s", res.Status())
	}

	var statsResponse struct {
		All struct {
			Primaries struct {
				Indexing struct {
					IndexTotal int64 `json:"index_total"`
				} `json:"indexing"`
				Search struct {
					QueryTotal int64 `json:"query_total"`
				} `json:"search"`
			} `json:"primaries"`
		} `json:"_all"`
	}

	if err := json.NewDecoder(res.Body).Decode(&statsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse cluster metrics: %w", err)
	}

	return &MetricsSnapshot{
		Timestamp:   time.Now(),
		IndexTotal:  statsResponse.All.Primaries.Indexing.IndexTotal,
		SearchTotal: statsResponse.All.Primaries.Search.QueryTotal,
	}, nil
}
