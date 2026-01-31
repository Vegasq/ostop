package ui

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
)

// App is the main Bubble Tea model
type App struct {
	client            *opensearch.Client
	endpoint          string
	health            *ClusterHealth
	stats             *ClusterStats
	nodes             []NodeInfo
	indices           []IndexInfo
	shards            []ShardInfo
	allocation        []AllocationInfo
	threadPool        []ThreadPoolInfo
	tasks             []TaskInfo
	pendingTasks      []PendingTaskInfo
	recovery          []RecoveryInfo
	segments          []SegmentInfo
	fielddata         []FielddataInfo
	plugins           []PluginInfo
	templates         []TemplateInfo
	loading           bool
	err               error
	lastRefresh       time.Time
	currentView       View
	activePanel       Panel
	selectedItem      int
	width             int
	height            int
	leftPanelWidth    int
	selectedNode      int
	selectedIndex     int
	selectedIndexName string
	indexMapping      *IndexMapping
	viewport          viewport.Model
	viewportReady     bool

	// Metrics state
	metricsTimeSeries *MetricsTimeSeries
	metricsEnabled    bool
	lastMetricsUpdate time.Time

	// Thread Pool Monitor state
	threadPoolTimeSeries *ThreadPoolTimeSeries
	threadPoolEnabled    bool
	lastThreadPoolUpdate time.Time
}

// NewApp creates a new application instance
func NewApp(client *opensearch.Client, endpoint string) *App {
	return &App{
		client:               client,
		endpoint:             endpoint,
		loading:              true,
		currentView:          ViewCluster,
		activePanel:          PanelLeft,
		selectedItem:         0,
		leftPanelWidth:       28,
		metricsTimeSeries:    NewMetricsTimeSeries(12),    // Last 60 seconds at 5-second intervals
		metricsEnabled:       false,                       // Enabled when user navigates to Live Metrics view
		threadPoolTimeSeries: NewThreadPoolTimeSeries(12), // Last 60 seconds at 5-second intervals
		threadPoolEnabled:    false,                       // Enabled when user navigates to Thread Pool Monitor view
	}
}

// Init initializes the app and triggers first data load
func (a *App) Init() tea.Cmd {
	return a.refresh()
}

// metricsTick creates a command that triggers after 5 seconds
func metricsTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return metricsTickMsg{timestamp: t}
	})
}

// refreshMetrics fetches cluster metrics in the background
func (a *App) refreshMetrics() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		snapshot, err := a.fetchClusterMetrics(ctx)
		return metricsRefreshMsg{
			snapshot: snapshot,
			err:      err,
		}
	}
}

// refreshThreadPoolMetrics fetches thread pool metrics in the background
func (a *App) refreshThreadPoolMetrics() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		snapshot, err := a.fetchThreadPoolMetrics(ctx)
		return threadPoolRefreshMsg{
			snapshot: snapshot,
			err:      err,
		}
	}
}

// Update handles messages and updates the model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		log.Println(a.width, a.height)
		a.viewport = viewport.New(a.width, a.height/4)
		// a.viewport = viewport.New(rightWidth-4, viewportHeight)
		a.viewport.YPosition = 0
		a.viewportReady = true

		// Initialize viewport on first window size message
		// Calculate viewport dimensions (subtract title, borders, help)
		// leftWidth := 50

		// var lp string = a.renderLeftPanel()
		// fmt.Println(lp)

		// rightWidth := a.width - leftWidth - 10
		// viewportHeight := a.height - 8 // Account for title, borders, help

		log.Println(a.width, a.height)

		widthPaddingOffset := 6
		heightPaddingOffset := 8

		a.viewport = viewport.New(a.width-a.leftPanelWidth-widthPaddingOffset, a.height-heightPaddingOffset)

		a.viewport.YPosition = 0
		a.viewportReady = true

		// Set initial content if we have data
		if a.health != nil {
			a.updateViewportContent()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit

		case "r":
			a.loading = true
			a.err = nil
			return a, a.refresh()

		case "tab":
			// Switch between panels
			if a.activePanel == PanelLeft {
				a.activePanel = PanelRight
			} else {
				a.activePanel = PanelLeft
			}

		case "up", "k":
			if a.activePanel == PanelLeft {
				// Navigate menu
				if a.selectedItem > 0 {
					a.selectedItem--
					return a, a.updateViewFromSelectionCmd()
				}
			} else {
				// Navigate indices or scroll viewport
				if a.currentView == ViewIndices && len(a.indices) > 0 {
					if a.selectedIndex > 0 {
						a.selectedIndex--
						a.updateViewportContent()
						// Scroll viewport to follow cursor (each index takes 5 lines)
						a.viewport.LineUp(2)
					}
				} else {
					// Scroll viewport up when in right panel
					a.viewport.LineUp(1)
				}
			}

		case "down", "j":
			if a.activePanel == PanelLeft {
				// Navigate menu
				if a.selectedItem < 15 { // Updated for ViewThreadPoolMonitor
					a.selectedItem++
					return a, a.updateViewFromSelectionCmd()
				}
			} else {
				// Navigate indices or scroll viewport
				if a.currentView == ViewIndices && len(a.indices) > 0 {
					if a.selectedIndex < len(a.indices)-1 {
						a.selectedIndex++
						a.updateViewportContent()
						// Scroll viewport to follow cursor (each index takes 5 lines)
						a.viewport.LineDown(2)
					}
				} else {
					// Scroll viewport down when in right panel
					a.viewport.LineDown(1)
				}
			}

		case "pgup", "b":
			if a.activePanel == PanelRight {
				a.viewport.ViewUp()
			}

		case "pgdown", "f", " ":
			if a.activePanel == PanelRight {
				a.viewport.ViewDown()
			}

		case "home", "g":
			if a.activePanel == PanelRight {
				a.viewport.GotoTop()
			}

		case "end", "G":
			if a.activePanel == PanelRight {
				a.viewport.GotoBottom()
			}

		case "enter":
			if a.activePanel == PanelLeft {
				cmd = a.updateViewFromSelectionCmd()
				a.activePanel = PanelRight
				a.viewport.GotoTop() // Reset scroll when switching views
				return a, cmd
			} else if a.activePanel == PanelRight {
				// When in indices view, drill down to schema
				if a.currentView == ViewIndices && len(a.indices) > 0 {
					if a.selectedIndex >= 0 && a.selectedIndex < len(a.indices) {
						a.selectedIndexName = a.indices[a.selectedIndex].Index
						a.currentView = ViewIndexSchema
						a.loading = true
						return a, a.fetchIndexMapping()
					}
				}
			}

		case "esc", "backspace":
			// Return from schema view to indices view
			if a.currentView == ViewIndexSchema {
				a.currentView = ViewIndices
				a.selectedIndexName = ""
				a.indexMapping = nil
				a.updateViewportContent()
				// Reset scroll position when returning to indices view
				if a.viewportReady {
					a.viewport.GotoTop()
				}
			}
		}

	case refreshMsg:
		a.loading = false
		a.err = msg.err
		if msg.err == nil {
			a.health = msg.health
			a.stats = msg.stats
			a.nodes = msg.nodes
			a.indices = msg.indices
			a.shards = msg.shards
			a.allocation = msg.allocation
			a.threadPool = msg.threadPool
			a.tasks = msg.tasks
			a.pendingTasks = msg.pendingTasks
			a.recovery = msg.recovery
			a.segments = msg.segments
			a.fielddata = msg.fielddata
			a.plugins = msg.plugins
			a.templates = msg.templates
			a.lastRefresh = time.Now()

			// Update viewport content when data refreshes
			a.updateViewportContent()
		}

	case mappingMsg:
		a.loading = false
		a.err = msg.err
		if msg.err == nil {
			a.indexMapping = msg.mapping
			a.updateViewportContent()
		}

	case metricsTickMsg:
		// Only process tick if metrics are enabled (user is on Live Metrics view)
		if a.metricsEnabled {
			// Fetch new metrics and schedule next tick
			return a, tea.Batch(a.refreshMetrics(), metricsTick())
		}
		// If metrics disabled, don't schedule next tick
		return a, nil

	case metricsRefreshMsg:
		if msg.err != nil {
			// Log error but don't stop ticker
			log.Printf("Metrics fetch error: %v", msg.err)
		} else if msg.snapshot != nil {
			// Add snapshot to time series
			added := a.metricsTimeSeries.AddSnapshot(msg.snapshot)
			if added {
				a.lastMetricsUpdate = time.Now()
				// Update viewport if we're on the metrics view
				if a.currentView == ViewLiveMetrics {
					a.updateViewportContent()
				}
			}
		}

	case threadPoolTickMsg:
		// Only process tick if thread pool monitoring is enabled
		if a.threadPoolEnabled {
			// Fetch new metrics and schedule next tick
			return a, tea.Batch(a.refreshThreadPoolMetrics(), threadPoolTick())
		}
		// If disabled, don't schedule next tick
		return a, nil

	case threadPoolRefreshMsg:
		if msg.err != nil {
			// Log error but don't stop ticker
			log.Printf("Thread pool metrics fetch error: %v", msg.err)
		} else if msg.snapshot != nil {
			// Add snapshot to time series
			a.threadPoolTimeSeries.AddSnapshot(msg.snapshot)
			a.lastThreadPoolUpdate = time.Now()
			// Update viewport if we're on the thread pool monitor view
			if a.currentView == ViewThreadPoolMonitor {
				a.updateViewportContent()
			}
		}
	}

	// Update viewport (for smooth scrolling) only when right panel is active
	if a.viewportReady && a.activePanel == PanelRight {
		a.viewport, cmd = a.viewport.Update(msg)
	}

	return a, cmd
}

// updateViewFromSelection updates the current view based on menu selection
func (a *App) updateViewFromSelection() {
	previousView := a.currentView
	a.currentView = View(a.selectedItem)
	a.selectedNode = 0
	a.selectedIndex = 0

	// Enable/disable metrics based on view
	wasEnabled := a.metricsEnabled
	a.metricsEnabled = (a.currentView == ViewLiveMetrics)

	// Enable/disable thread pool monitoring based on view
	wasThreadPoolEnabled := a.threadPoolEnabled
	a.threadPoolEnabled = (a.currentView == ViewThreadPoolMonitor)
	_ = wasThreadPoolEnabled

	// Update viewport content
	a.updateViewportContent()

	// Reset scroll position when switching views
	if a.viewportReady {
		a.viewport.GotoTop()
	}

	// Note: Command to start ticker is handled by updateViewFromSelectionCmd()
	_ = previousView
	_ = wasEnabled
}

// updateViewFromSelectionCmd updates the view and returns a command to start metrics if needed
func (a *App) updateViewFromSelectionCmd() tea.Cmd {
	previousView := a.currentView
	wasEnabled := a.metricsEnabled
	wasThreadPoolEnabled := a.threadPoolEnabled

	// Update view state
	a.updateViewFromSelection()

	var cmds []tea.Cmd

	// Start metrics ticker if transitioning to Live Metrics view
	if !wasEnabled && a.metricsEnabled {
		// Start ticker and immediate first fetch
		cmds = append(cmds, a.refreshMetrics(), metricsTick())
	}

	// Start thread pool ticker if transitioning to Thread Pool Monitor view
	if !wasThreadPoolEnabled && a.threadPoolEnabled {
		// Start ticker and immediate first fetch
		cmds = append(cmds, a.refreshThreadPoolMetrics(), threadPoolTick())
	}

	// Stop ticker if leaving views (handled by enabled flags in tick handlers)
	_ = previousView

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// updateViewportContent updates the viewport with current view content
func (a *App) updateViewportContent() {
	if !a.viewportReady {
		return
	}
	content := a.renderRightPanel()
	a.viewport.SetContent(content)
}

// View renders the UI
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var b string

	// Title bar
	title := titleStyle.Render("ostop - OpenSearch Cluster Monitor")
	statusBar := statusBarStyle.Render(fmt.Sprintf("Endpoint: %s", a.endpoint))
	b += lipgloss.JoinHorizontal(lipgloss.Top, title, statusBar)
	b += "\n\n"

	// Loading state
	if a.loading {
		b += "Loading cluster data...\n"
		return b
	}

	// Error state
	if a.err != nil {
		b += errorStyle.Render(fmt.Sprintf("Error: %v", a.err))
		b += "\n\n"
		b += helpStyle.Render("Press 'r' to retry | 'q' to quit")
		return b
	}

	// Split panel layout
	leftPanel := a.renderLeftPanel()

	// Apply panel styles
	leftPanelStyled := a.stylePanel(leftPanel, PanelLeft)
	// a.leftPanelWidth = len(strings.Split(leftPanel, "\n")[0])

	// Render viewport content with border
	var rightPanelStyled string
	rightPanelStyled = a.stylePanel(a.viewport.View(), PanelRight)
	// if a.viewportReady {
	// 	// rightPanelStyled = a.renderRightPanel()
	// } else {
	// 	// If viewport not ready, render content directly
	// 	// rightPanel := a.renderRightPanel()
	// 	// rightPanelStyled = a.stylePanel(rightPanel, PanelRight)
	// }

	// Calculate panel widths
	// leftWidth := 30
	// rightWidth := a.width - leftWidth - 10

	leftPanelStyled = lipgloss.NewStyle().Width(a.leftPanelWidth).Render(leftPanelStyled)
	rightPanelStyled = lipgloss.NewStyle().Width(a.width - a.leftPanelWidth).Render(rightPanelStyled)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanelStyled, rightPanelStyled)
	b += panels
	b += "\n\n"

	// Help footer with scroll info
	helpText := "↑/↓: Navigate | Tab: Switch Panel | Enter: Select"
	if a.currentView == ViewIndexSchema {
		helpText += " | Esc: Back"
	}
	helpText += " | r: Refresh | q: Quit"
	if a.activePanel == PanelRight && a.viewportReady {
		scrollPercent := int(a.viewport.ScrollPercent() * 100)
		if scrollPercent < 100 {
			helpText += fmt.Sprintf(" | Scroll: %d%%", scrollPercent)
		}
	}
	help := helpStyle.Render(helpText)
	b += help

	return b
}
