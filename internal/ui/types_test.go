package ui

import (
	"encoding/json"
	"os"
	"testing"
)

func TestClusterHealth_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/cluster_health.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var health ClusterHealth
	if err := json.Unmarshal(data, &health); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Validate critical fields
	if health.ClusterName == "" {
		t.Error("expected cluster_name to be populated")
	}

	validStatuses := map[string]bool{"green": true, "yellow": true, "red": true}
	if !validStatuses[health.Status] {
		t.Errorf("invalid status: %s", health.Status)
	}

	if health.NumberOfNodes != 3 {
		t.Errorf("expected 3 nodes, got %d", health.NumberOfNodes)
	}

	if health.ActiveShards != 20 {
		t.Errorf("expected 20 active shards, got %d", health.ActiveShards)
	}
}

func TestClusterStats_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/cluster_stats.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var stats ClusterStats
	if err := json.Unmarshal(data, &stats); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if stats.ClusterName == "" {
		t.Error("expected cluster_name to be populated")
	}

	if stats.Indices.Count != 15 {
		t.Errorf("expected 15 indices, got %d", stats.Indices.Count)
	}

	if stats.Indices.Docs.Count != 1000000 {
		t.Errorf("expected 1000000 docs, got %d", stats.Indices.Docs.Count)
	}

	if stats.Nodes.Count.Total != 3 {
		t.Errorf("expected 3 total nodes, got %d", stats.Nodes.Count.Total)
	}
}

func TestNodeInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/nodes.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var nodes []NodeInfo
	if err := json.Unmarshal(data, &nodes); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	node1 := nodes[0]
	if node1.IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", node1.IP)
	}

	if node1.Name != "node-1" {
		t.Errorf("expected name node-1, got %s", node1.Name)
	}

	if node1.Master != "*" {
		t.Errorf("expected master *, got %s", node1.Master)
	}
}

func TestIndexInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/indices.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var indices []IndexInfo
	if err := json.Unmarshal(data, &indices); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(indices) != 2 {
		t.Fatalf("expected 2 indices, got %d", len(indices))
	}

	idx := indices[0]
	if idx.Health != "green" {
		t.Errorf("expected health green, got %s", idx.Health)
	}

	if idx.Index != "test-index-1" {
		t.Errorf("expected index test-index-1, got %s", idx.Index)
	}
}

func TestShardInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/shards.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var shards []ShardInfo
	if err := json.Unmarshal(data, &shards); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(shards) != 3 {
		t.Fatalf("expected 3 shards, got %d", len(shards))
	}

	shard := shards[0]
	if shard.Prirep != "p" {
		t.Errorf("expected prirep p, got %s", shard.Prirep)
	}

	if shard.State != "STARTED" {
		t.Errorf("expected state STARTED, got %s", shard.State)
	}
}

func TestAllocationInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/allocation.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var allocations []AllocationInfo
	if err := json.Unmarshal(data, &allocations); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(allocations) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(allocations))
	}

	alloc := allocations[0]
	if alloc.Node != "node-1" {
		t.Errorf("expected node node-1, got %s", alloc.Node)
	}

	if alloc.DiskPercent != "40" {
		t.Errorf("expected disk percent 40, got %s", alloc.DiskPercent)
	}
}

func TestThreadPoolInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/threadpool.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var pools []ThreadPoolInfo
	if err := json.Unmarshal(data, &pools); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(pools) != 3 {
		t.Fatalf("expected 3 thread pools, got %d", len(pools))
	}

	pool := pools[0]
	if pool.Name != "search" {
		t.Errorf("expected name search, got %s", pool.Name)
	}

	if pool.NodeName != "node-1" {
		t.Errorf("expected node_name node-1, got %s", pool.NodeName)
	}
}

func TestTaskInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/tasks.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var tasks []TaskInfo
	if err := json.Unmarshal(data, &tasks); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Action != "indices:data/write/bulk" {
		t.Errorf("expected action indices:data/write/bulk, got %s", task.Action)
	}

	if task.RunningTime != "1.5s" {
		t.Errorf("expected running time 1.5s, got %s", task.RunningTime)
	}
}

func TestPendingTaskInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/pending_tasks.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var tasks []PendingTaskInfo
	if err := json.Unmarshal(data, &tasks); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 pending tasks, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Priority != "URGENT" {
		t.Errorf("expected priority URGENT, got %s", task.Priority)
	}
}

func TestRecoveryInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/recovery.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var recoveries []RecoveryInfo
	if err := json.Unmarshal(data, &recoveries); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(recoveries) != 2 {
		t.Fatalf("expected 2 recoveries, got %d", len(recoveries))
	}

	recovery := recoveries[0]
	if recovery.Stage != "done" {
		t.Errorf("expected stage done, got %s", recovery.Stage)
	}

	if recovery.Type != "peer" {
		t.Errorf("expected type peer, got %s", recovery.Type)
	}
}

func TestSegmentInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/segments.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var segments []SegmentInfo
	if err := json.Unmarshal(data, &segments); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	segment := segments[0]
	if segment.Segment != "_0" {
		t.Errorf("expected segment _0, got %s", segment.Segment)
	}

	if segment.Committed != "true" {
		t.Errorf("expected committed true, got %s", segment.Committed)
	}
}

func TestFielddataInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/fielddata.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var fielddata []FielddataInfo
	if err := json.Unmarshal(data, &fielddata); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(fielddata) != 2 {
		t.Fatalf("expected 2 fielddata entries, got %d", len(fielddata))
	}

	fd := fielddata[0]
	if fd.Field != "user.keyword" {
		t.Errorf("expected field user.keyword, got %s", fd.Field)
	}
}

func TestPluginInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/plugins.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var plugins []PluginInfo
	if err := json.Unmarshal(data, &plugins); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	plugin := plugins[0]
	if plugin.Name != "analysis-icu" {
		t.Errorf("expected name analysis-icu, got %s", plugin.Name)
	}
}

func TestTemplateInfo_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/templates.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var templates []TemplateInfo
	if err := json.Unmarshal(data, &templates); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	template := templates[0]
	if template.Name != "logs-template" {
		t.Errorf("expected name logs-template, got %s", template.Name)
	}
}
