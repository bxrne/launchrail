package main

import "github.com/charmbracelet/lipgloss"

var (
	accentColor     = lipgloss.Color("#FFA500")
	titleStyle      = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Padding(1, 2).MarginBottom(1)
	headerStyle     = lipgloss.NewStyle().Bold(true).Padding(0, 2)
	footerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Align(lipgloss.Center).Padding(0, 2)
	footerLinkStyle = lipgloss.NewStyle().Foreground(accentColor).Underline(true).Padding(0, 2)
	containerStyle  = lipgloss.NewStyle().Margin(1, 2)               // INFO: Layout container
	contentStyle    = lipgloss.NewStyle().Padding(1, 2).Margin(1, 2) // INFO: Layout container
	logStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(1, 2).Margin(1, 2).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#444444"))
	logPanelHeight  = 16
)
