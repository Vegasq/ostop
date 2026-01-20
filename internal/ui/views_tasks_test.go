package ui

import (
	"strings"
	"testing"
)

func TestRenderTasksView_Empty(t *testing.T) {
	app := &App{
		tasks: []TaskInfo{},
	}

	result := app.renderTasksView()

	if !strings.Contains(result, "Running Tasks (0)") {
		t.Errorf("Expected empty task count, got: %s", result)
	}
	if !strings.Contains(result, "No running tasks") {
		t.Errorf("Expected 'No running tasks' message, got: %s", result)
	}
}

func TestRenderTasksView_SingleTask(t *testing.T) {
	app := &App{
		tasks: []TaskInfo{
			{
				Action:      "indices:data/write/bulk[s]",
				TaskID:      "node1:12345",
				RunningTime: "5s",
				Node:        "i-0abc123",
				IP:          "10.0.1.5",
				Description: "update-by-query [my_index]",
			},
		},
	}

	result := app.renderTasksView()

	if !strings.Contains(result, "Running Tasks (1)") {
		t.Errorf("Expected task count of 1, got: %s", result)
	}
	if !strings.Contains(result, "indices:data/write/bulk[s]") {
		t.Errorf("Expected full action string, got: %s", result)
	}
	if !strings.Contains(result, "node1:12345") {
		t.Errorf("Expected task ID, got: %s", result)
	}
	if !strings.Contains(result, "10.0.1.5") {
		t.Errorf("Expected IP address, got: %s", result)
	}
	if !strings.Contains(result, "i-0abc123") {
		t.Errorf("Expected node name, got: %s", result)
	}
	if !strings.Contains(result, "update-by-query [my_index]") {
		t.Errorf("Expected description, got: %s", result)
	}
}

func TestRenderTasksView_Sorting(t *testing.T) {
	app := &App{
		tasks: []TaskInfo{
			{
				Action:      "task1",
				TaskID:      "task1",
				RunningTime: "5s",
				Node:        "node1",
			},
			{
				Action:      "task2",
				TaskID:      "task2",
				RunningTime: "65s",
				Node:        "node2",
			},
			{
				Action:      "task3",
				TaskID:      "task3",
				RunningTime: "35s",
				Node:        "node3",
			},
		},
	}

	result := app.renderTasksView()

	// Find positions of tasks in output
	pos1 := strings.Index(result, "task1")
	pos2 := strings.Index(result, "task2")
	pos3 := strings.Index(result, "task3")

	// task2 (65s) should come first, then task3 (35s), then task1 (5s)
	if pos2 >= pos3 || pos3 >= pos1 {
		t.Errorf("Tasks not sorted by running time (longest first). Order: task2=%d, task3=%d, task1=%d", pos2, pos3, pos1)
	}
}

func TestRenderTasksView_CriticalWarning(t *testing.T) {
	tests := []struct {
		name           string
		runningTime    string
		expectedMsg    string
		expectedInMsg  string
	}{
		{
			name:           "Critical task >= 60s",
			runningTime:    "65s",
			expectedMsg:    "⚠",
			expectedInMsg:  "running ≥60s",
		},
		{
			name:           "Warning task >= 30s",
			runningTime:    "35s",
			expectedMsg:    "⚠",
			expectedInMsg:  "running ≥30s",
		},
		{
			name:        "Normal task < 30s",
			runningTime: "10s",
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				tasks: []TaskInfo{
					{
						Action:      "test:action",
						TaskID:      "test:1",
						RunningTime: tt.runningTime,
						Node:        "testnode",
					},
				},
			}

			result := app.renderTasksView()

			if tt.expectedMsg != "" {
				if !strings.Contains(result, tt.expectedMsg) {
					t.Errorf("Expected warning message, got: %s", result)
				}
				if tt.expectedInMsg != "" && !strings.Contains(result, tt.expectedInMsg) {
					t.Errorf("Expected '%s' in message, got: %s", tt.expectedInMsg, result)
				}
			}
		})
	}
}

func TestRenderTasksView_TaskIDTruncation(t *testing.T) {
	longTaskID := "node1234567890123456789012345678901234567890extra"
	app := &App{
		tasks: []TaskInfo{
			{
				Action:      "test:action",
				TaskID:      longTaskID,
				RunningTime: "5s",
				Node:        "node1",
			},
		},
	}

	result := app.renderTasksView()

	if strings.Contains(result, longTaskID) {
		t.Errorf("Task ID should be truncated, but found full ID: %s", result)
	}
	if !strings.Contains(result, "...") {
		t.Errorf("Expected truncation indicator '...', got: %s", result)
	}
}

func TestRenderTasksView_DescriptionTruncation(t *testing.T) {
	longDesc := strings.Repeat("a", 130) // Longer than 120 chars
	app := &App{
		tasks: []TaskInfo{
			{
				Action:      "test:action",
				TaskID:      "test:1",
				RunningTime: "5s",
				Node:        "node1",
				Description: longDesc,
			},
		},
	}

	result := app.renderTasksView()

	if strings.Contains(result, longDesc) {
		t.Errorf("Description should be truncated, but found full description")
	}
	if !strings.Contains(result, "...") {
		t.Errorf("Expected truncation indicator '...', got: %s", result)
	}
}

func TestRenderTasksView_IPDisplay(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		expectedIP  bool
		expectedFmt string
	}{
		{
			name:        "With IP",
			ip:          "10.0.1.5",
			expectedIP:  true,
			expectedFmt: "IP: 10.0.1.5",
		},
		{
			name:        "Without IP",
			ip:          "",
			expectedIP:  false,
			expectedFmt: "Node:",
		},
		{
			name:        "Null IP",
			ip:          "null",
			expectedIP:  false,
			expectedFmt: "Node:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				tasks: []TaskInfo{
					{
						Action:      "test:action",
						TaskID:      "test:1",
						RunningTime: "5s",
						Node:        "node1",
						IP:          tt.ip,
					},
				},
			}

			result := app.renderTasksView()

			if tt.expectedIP {
				if !strings.Contains(result, tt.expectedFmt) {
					t.Errorf("Expected IP format '%s', got: %s", tt.expectedFmt, result)
				}
			} else {
				if strings.Contains(result, "IP:") {
					t.Errorf("Should not contain 'IP:' when IP is empty/null, got: %s", result)
				}
			}
		})
	}
}

func TestRenderTasksView_NullDescription(t *testing.T) {
	app := &App{
		tasks: []TaskInfo{
			{
				Action:      "test:action",
				TaskID:      "test:1",
				RunningTime: "5s",
				Node:        "node1",
				Description: "null",
			},
		},
	}

	result := app.renderTasksView()

	// Should not display "Details:" line when description is "null"
	if strings.Contains(result, "Details:") {
		t.Errorf("Should not display Details when description is 'null', got: %s", result)
	}
}

func TestRenderTasksView_Limit50(t *testing.T) {
	// Create 60 tasks
	tasks := make([]TaskInfo, 60)
	for i := 0; i < 60; i++ {
		tasks[i] = TaskInfo{
			Action:      "test:action",
			TaskID:      "test:1",
			RunningTime: "5s",
			Node:        "node1",
		}
	}

	app := &App{
		tasks: tasks,
	}

	result := app.renderTasksView()

	if !strings.Contains(result, "Running Tasks (60)") {
		t.Errorf("Expected total task count of 60, got: %s", result)
	}
	if !strings.Contains(result, "... and 10 more tasks") {
		t.Errorf("Expected message about 10 more tasks, got: %s", result)
	}
}

func TestRenderPendingTasksView_Empty(t *testing.T) {
	app := &App{
		pendingTasks: []PendingTaskInfo{},
	}

	result := app.renderPendingTasksView()

	if !strings.Contains(result, "Pending Cluster Tasks (0)") {
		t.Errorf("Expected empty pending task count, got: %s", result)
	}
	if !strings.Contains(result, "No pending tasks") {
		t.Errorf("Expected 'No pending tasks' message, got: %s", result)
	}
}

func TestRenderPendingTasksView_SingleTask(t *testing.T) {
	app := &App{
		pendingTasks: []PendingTaskInfo{
			{
				InsertOrder: "1",
				TimeInQueue: "100ms",
				Priority:    "URGENT",
				Source:      "create-index [test_index]",
			},
		},
	}

	result := app.renderPendingTasksView()

	if !strings.Contains(result, "Pending Cluster Tasks (1)") {
		t.Errorf("Expected pending task count of 1, got: %s", result)
	}
	if !strings.Contains(result, "create-index [test_index]") {
		t.Errorf("Expected source, got: %s", result)
	}
	if !strings.Contains(result, "URGENT") {
		t.Errorf("Expected priority, got: %s", result)
	}
	if !strings.Contains(result, "Insert Order:") {
		t.Errorf("Expected insert order label, got: %s", result)
	}
}

func TestRenderPendingTasksView_CriticalWarning(t *testing.T) {
	tests := []struct {
		name           string
		timeInQueue    string
		expectedMsg    string
		expectedInMsg  string
	}{
		{
			name:           "Critical task >= 5s",
			timeInQueue:    "6s",
			expectedMsg:    "⚠ CRITICAL:",
			expectedInMsg:  "queued ≥5s",
		},
		{
			name:           "Warning task >= 1s",
			timeInQueue:    "1500ms",
			expectedMsg:    "⚠",
			expectedInMsg:  "queued ≥1s",
		},
		{
			name:        "Normal task < 1s",
			timeInQueue: "500ms",
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				pendingTasks: []PendingTaskInfo{
					{
						InsertOrder: "1",
						TimeInQueue: tt.timeInQueue,
						Priority:    "NORMAL",
						Source:      "test-task",
					},
				},
			}

			result := app.renderPendingTasksView()

			if tt.expectedMsg != "" {
				if !strings.Contains(result, tt.expectedMsg) {
					t.Errorf("Expected warning message '%s', got: %s", tt.expectedMsg, result)
				}
				if tt.expectedInMsg != "" && !strings.Contains(result, tt.expectedInMsg) {
					t.Errorf("Expected '%s' in message, got: %s", tt.expectedInMsg, result)
				}
			}
		})
	}
}

func TestRenderPendingTasksView_MultipleTasks(t *testing.T) {
	app := &App{
		pendingTasks: []PendingTaskInfo{
			{
				InsertOrder: "1",
				TimeInQueue: "100ms",
				Priority:    "URGENT",
				Source:      "create-index [index1]",
			},
			{
				InsertOrder: "2",
				TimeInQueue: "50ms",
				Priority:    "HIGH",
				Source:      "update-mapping [index2]",
			},
		},
	}

	result := app.renderPendingTasksView()

	if !strings.Contains(result, "Pending Cluster Tasks (2)") {
		t.Errorf("Expected pending task count of 2, got: %s", result)
	}
	if !strings.Contains(result, "create-index [index1]") {
		t.Errorf("Expected first task source, got: %s", result)
	}
	if !strings.Contains(result, "update-mapping [index2]") {
		t.Errorf("Expected second task source, got: %s", result)
	}
	if !strings.Contains(result, "URGENT") {
		t.Errorf("Expected URGENT priority, got: %s", result)
	}
	if !strings.Contains(result, "HIGH") {
		t.Errorf("Expected HIGH priority, got: %s", result)
	}
}
