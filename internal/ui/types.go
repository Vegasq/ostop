package ui

import "time"

// View represents different views in the application
type View int

const (
	ViewCluster View = iota
	ViewNodes
	ViewIndices
	ViewShards
	ViewResources
	ViewLiveMetrics // Real-time cluster activity metrics
	ViewAllocation
	ViewThreadPool
	ViewTasks
	ViewPendingTasks
	ViewRecovery
	ViewSegments
	ViewFielddata
	ViewPlugins
	ViewTemplates
	ViewIndexSchema // Special view accessed via drill-down from Indices
)

// Panel represents which panel is active
type Panel int

const (
	PanelLeft Panel = iota
	PanelRight
)

// ClusterHealth represents the cluster health response
type ClusterHealth struct {
	ClusterName         string `json:"cluster_name"`
	Status              string `json:"status"`
	TimedOut            bool   `json:"timed_out"`
	NumberOfNodes       int    `json:"number_of_nodes"`
	NumberOfDataNodes   int    `json:"number_of_data_nodes"`
	ActivePrimaryShards int    `json:"active_primary_shards"`
	ActiveShards        int    `json:"active_shards"`
	RelocatingShards    int    `json:"relocating_shards"`
	InitializingShards  int    `json:"initializing_shards"`
	UnassignedShards    int    `json:"unassigned_shards"`
}

// ClusterStats represents simplified cluster stats
type ClusterStats struct {
	ClusterName string `json:"cluster_name"`
	Status      string `json:"status"`
	Indices     struct {
		Count int `json:"count"`
		Docs  struct {
			Count int64 `json:"count"`
		} `json:"docs"`
		Store struct {
			SizeInBytes int64 `json:"size_in_bytes"`
		} `json:"store"`
	} `json:"indices"`
	Nodes struct {
		Count struct {
			Total int `json:"total"`
			Data  int `json:"data"`
		} `json:"count"`
	} `json:"nodes"`
}

// NodeInfo represents a node from CAT nodes API
type NodeInfo struct {
	IP              string `json:"ip"`
	HeapPercent     string `json:"heap.percent"`
	RAMPercent      string `json:"ram.percent"`
	CPU             string `json:"cpu"`
	Load1m          string `json:"load_1m"`
	Load5m          string `json:"load_5m"`
	Load15m         string `json:"load_15m"`
	NodeRole        string `json:"node.role"`
	Master          string `json:"master"`
	Name            string `json:"name"`
	DiskUsedPercent string `json:"disk.used_percent"`
	DiskUsed        string `json:"disk.used"`
	DiskAvail       string `json:"disk.avail"`
	DiskTotal       string `json:"disk.total"`
}

// IndexInfo represents an index from CAT indices API
type IndexInfo struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	UUID         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
}

// ShardInfo represents a shard from CAT shards API
type ShardInfo struct {
	Index  string `json:"index"`
	Shard  string `json:"shard"`
	Prirep string `json:"prirep"` // "p" for primary, "r" for replica
	State  string `json:"state"`  // STARTED, RELOCATING, INITIALIZING, UNASSIGNED
	Docs   string `json:"docs"`
	Store  string `json:"store"`
	IP     string `json:"ip"`
	Node   string `json:"node"`
}

// IndexMapping represents the mapping structure for an index
type IndexMapping struct {
	IndexName string
	Mappings  map[string]interface{}
}

// FieldInfo represents a field in the index mapping
type FieldInfo struct {
	Name       string
	Type       string
	Index      string // "true", "false", or ""
	Analyzer   string
	Searchable bool
	Properties map[string]FieldInfo // For nested fields
}

// AllocationInfo represents node disk allocation from CAT allocation API
type AllocationInfo struct {
	Shards      string `json:"shards"`
	DiskIndices string `json:"disk.indices"`
	DiskUsed    string `json:"disk.used"`
	DiskAvail   string `json:"disk.avail"`
	DiskTotal   string `json:"disk.total"`
	DiskPercent string `json:"disk.percent"`
	Host        string `json:"host"`
	IP          string `json:"ip"`
	Node        string `json:"node"`
}

// ThreadPoolInfo represents thread pool statistics from CAT thread_pool API
type ThreadPoolInfo struct {
	NodeName  string `json:"node_name"`
	Name      string `json:"name"`
	Active    string `json:"active"`
	Queue     string `json:"queue"`
	Rejected  string `json:"rejected"`
	Completed string `json:"completed"`
	Size      string `json:"size"`
}

// TaskInfo represents a running task from CAT tasks API
type TaskInfo struct {
	Action       string `json:"action"`
	TaskID       string `json:"task_id"`
	ParentTaskID string `json:"parent_task_id"`
	Type         string `json:"type"`
	StartTime    string `json:"start_time"`
	Timestamp    string `json:"timestamp"`
	RunningTime  string `json:"running_time"`
	IP           string `json:"ip"`
	Node         string `json:"node"`
	Description  string `json:"description"`
}

// PendingTaskInfo represents a pending cluster task from CAT pending_tasks API
type PendingTaskInfo struct {
	InsertOrder string `json:"insertOrder"`
	TimeInQueue string `json:"timeInQueue"`
	Priority    string `json:"priority"`
	Source      string `json:"source"`
}

// RecoveryInfo represents shard recovery information from CAT recovery API
type RecoveryInfo struct {
	Index          string `json:"index"`
	Shard          string `json:"shard"`
	Time           string `json:"time"`
	Type           string `json:"type"`
	Stage          string `json:"stage"`
	SourceNode     string `json:"source_node"`
	TargetNode     string `json:"target_node"`
	Files          string `json:"files"`
	FilesRecovered string `json:"files_recovered"`
	FilesPercent   string `json:"files_percent"`
	Bytes          string `json:"bytes"`
	BytesRecovered string `json:"bytes_recovered"`
	BytesPercent   string `json:"bytes_percent"`
}

// SegmentInfo represents Lucene segment information from CAT segments API
type SegmentInfo struct {
	Index       string `json:"index"`
	Shard       string `json:"shard"`
	Prirep      string `json:"prirep"`
	IP          string `json:"ip"`
	Segment     string `json:"segment"`
	Generation  string `json:"generation"`
	DocsCount   string `json:"docs.count"`
	DocsDeleted string `json:"docs.deleted"`
	Size        string `json:"size"`
	Committed   string `json:"committed"`
}

// FielddataInfo represents fielddata cache usage from CAT fielddata API
type FielddataInfo struct {
	ID    string `json:"id"`
	Host  string `json:"host"`
	IP    string `json:"ip"`
	Node  string `json:"node"`
	Field string `json:"field"`
	Size  string `json:"size"`
}

// PluginInfo represents installed plugin information from CAT plugins API
type PluginInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Component   string `json:"component"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// TemplateInfo represents index template information from CAT templates API
type TemplateInfo struct {
	Name          string `json:"name"`
	IndexPatterns string `json:"index_patterns"`
	Order         string `json:"order"`
	Version       string `json:"version"`
}

// MetricsSnapshot represents a single point-in-time measurement from cluster stats
type MetricsSnapshot struct {
	Timestamp   time.Time
	IndexTotal  int64 // Cumulative inserts since cluster start
	SearchTotal int64 // Cumulative searches since cluster start
}

// MetricsDataPoint represents a calculated rate over an interval
type MetricsDataPoint struct {
	Timestamp  time.Time
	InsertRate float64 // Inserts per second
	SearchRate float64 // Searches per second
}

// MetricsTimeSeries manages the rolling window of metrics
type MetricsTimeSeries struct {
	DataPoints   []MetricsDataPoint
	MaxSize      int              // Maximum number of data points to store
	LastSnapshot *MetricsSnapshot // For delta calculation
}

// MetricsSummary provides aggregate statistics
type MetricsSummary struct {
	Current float64
	Average float64
	Peak    float64
	Min     float64
}

// refreshMsg is sent when data refresh completes
type refreshMsg struct {
	health       *ClusterHealth
	stats        *ClusterStats
	nodes        []NodeInfo
	indices      []IndexInfo
	shards       []ShardInfo
	allocation   []AllocationInfo
	threadPool   []ThreadPoolInfo
	tasks        []TaskInfo
	pendingTasks []PendingTaskInfo
	recovery     []RecoveryInfo
	segments     []SegmentInfo
	fielddata    []FielddataInfo
	plugins      []PluginInfo
	templates    []TemplateInfo
	err          error
}

// mappingMsg is sent when index mapping fetch completes
type mappingMsg struct {
	mapping *IndexMapping
	err     error
}

// metricsTickMsg triggers periodic metrics refresh
type metricsTickMsg struct {
	timestamp time.Time
}

// metricsRefreshMsg carries fetched metrics data
type metricsRefreshMsg struct {
	snapshot *MetricsSnapshot
	err      error
}
