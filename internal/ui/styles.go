package ui

import "github.com/charmbracelet/lipgloss"

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("39")).
				Padding(1, 2)

	inactivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(1, 2)

	menuItemStyle = lipgloss.NewStyle().
			Padding(0, 2)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Foreground(lipgloss.Color("39")).
				Bold(true)

	statusGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	statusYellow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	statusRed = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	valueStyle = lipgloss.NewStyle().
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			MarginBottom(1)

	// Metrics-specific styles
	metricHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	highlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("yellow"))

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))
)
