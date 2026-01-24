package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderTasksView renders the currently running tasks view
func (a *App) renderTasksView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Running Tasks (%d)", len(a.tasks))))
	b.WriteString("\n\n")

	if len(a.tasks) == 0 {
		b.WriteString(statusGreen.Render("✓ No running tasks"))
		return b.String()
	}

	// Sort by running time (longest first)
	type taskWithTime struct {
		task    TaskInfo
		seconds float64
	}

	var tasksWithTime []taskWithTime
	for _, task := range a.tasks {
		seconds := parseRunningTime(task.RunningTime)
		tasksWithTime = append(tasksWithTime, taskWithTime{task: task, seconds: seconds})
	}

	// Simple bubble sort by seconds (descending)
	for i := 0; i < len(tasksWithTime); i++ {
		for j := i + 1; j < len(tasksWithTime); j++ {
			if tasksWithTime[j].seconds > tasksWithTime[i].seconds {
				tasksWithTime[i], tasksWithTime[j] = tasksWithTime[j], tasksWithTime[i]
			}
		}
	}

	// Limit to 50 tasks for performance
	displayCount := len(tasksWithTime)
	if displayCount > 50 {
		displayCount = 50
	}

	// Check for long-running tasks
	criticalCount := 0
	warningCount := 0
	for i := 0; i < displayCount; i++ {
		if tasksWithTime[i].seconds >= 60 {
			criticalCount++
		} else if tasksWithTime[i].seconds >= 30 {
			warningCount++
		}
	}

	if criticalCount > 0 {
		b.WriteString(statusRed.Render(fmt.Sprintf("⚠ %d task(s) running ≥60s", criticalCount)))
		b.WriteString("\n\n")
	} else if warningCount > 0 {
		b.WriteString(statusYellow.Render(fmt.Sprintf("⚠ %d task(s) running ≥30s", warningCount)))
		b.WriteString("\n\n")
	}

	// Display tasks
	for i := 0; i < displayCount; i++ {
		twt := tasksWithTime[i]
		task := twt.task
		seconds := twt.seconds

		var timeStyle lipgloss.Style
		if seconds >= 60 {
			timeStyle = statusRed
		} else if seconds >= 30 {
			timeStyle = statusYellow
		} else {
			timeStyle = statusGreen
		}

		// Show full action instead of simplified version
		b.WriteString(fmt.Sprintf("%s %s\n",
			timeStyle.Render(fmt.Sprintf("[%s]", task.RunningTime)),
			valueStyle.Render(task.Action)))

		// Show Task ID and IP/Node information
		taskIDDisplay := task.TaskID
		if len(taskIDDisplay) > 40 {
			taskIDDisplay = taskIDDisplay[:37] + "..."
		}

		nodeInfo := fmt.Sprintf("Node: %s", task.Node)
		if task.IP != "" && task.IP != "null" {
			nodeInfo = fmt.Sprintf("IP: %s | Node: %s", task.IP, task.Node)
		}

		b.WriteString(fmt.Sprintf("  %s %s | %s\n",
			labelStyle.Render("Task ID:"),
			labelStyle.Render(taskIDDisplay),
			labelStyle.Render(nodeInfo)))

		// Show description with more space
		if task.Description != "" && task.Description != "null" {
			desc := task.Description
			if len(desc) > 120 {
				desc = desc[:117] + "..."
			}
			b.WriteString(fmt.Sprintf("  %s %s\n",
				labelStyle.Render("Details:"),
				labelStyle.Render(desc)))
		}
		b.WriteString("\n")
	}

	if len(tasksWithTime) > displayCount {
		b.WriteString(labelStyle.Render(fmt.Sprintf("... and %d more tasks", len(tasksWithTime)-displayCount)))
		b.WriteString("\n")
	}

	return b.String()
}

// renderPendingTasksView renders the pending cluster tasks view
func (a *App) renderPendingTasksView() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("Pending Cluster Tasks (%d)", len(a.pendingTasks))))
	b.WriteString("\n\n")

	if len(a.pendingTasks) == 0 {
		b.WriteString(statusGreen.Render("✓ No pending tasks"))
		return b.String()
	}

	// Check for tasks queued too long
	criticalCount := 0
	warningCount := 0
	for _, task := range a.pendingTasks {
		ms := parseTimeInQueue(task.TimeInQueue)
		if ms >= 5000 {
			criticalCount++
		} else if ms >= 1000 {
			warningCount++
		}
	}

	if criticalCount > 0 {
		b.WriteString(statusRed.Render(fmt.Sprintf("⚠ CRITICAL: %d task(s) queued ≥5s", criticalCount)))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Long queue times indicate cluster state update delays"))
		b.WriteString("\n\n")
	} else if warningCount > 0 {
		b.WriteString(statusYellow.Render(fmt.Sprintf("⚠ %d task(s) queued ≥1s", warningCount)))
		b.WriteString("\n\n")
	}

	// Display tasks
	for _, task := range a.pendingTasks {
		ms := parseTimeInQueue(task.TimeInQueue)

		var timeStyle lipgloss.Style
		if ms >= 5000 {
			timeStyle = statusRed
		} else if ms >= 1000 {
			timeStyle = statusYellow
		} else {
			timeStyle = statusGreen
		}

		b.WriteString(fmt.Sprintf("%s %s\n",
			timeStyle.Render(fmt.Sprintf("[%s]", task.TimeInQueue)),
			valueStyle.Render(task.Source)))
		b.WriteString(fmt.Sprintf("  %s %s  %s %s\n",
			labelStyle.Render("Priority:"), task.Priority,
			labelStyle.Render("Insert Order:"), task.InsertOrder))
		b.WriteString("\n")
	}

	return b.String()
}
