package tui

import "github.com/charmbracelet/lipgloss"

var (
	accentColor      = lipgloss.Color("#5a56e0")
	secondaryColor   = lipgloss.Color("#888888")
	textColor        = lipgloss.Color("#FFFFFF")
	titleStyle       = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	descriptionStyle = lipgloss.NewStyle().Foreground(secondaryColor).MarginBottom(1)
	promptStyle      = lipgloss.NewStyle().Foreground(accentColor)
	linkStyle        = lipgloss.NewStyle().Foreground(accentColor).Underline(true)
	accentStyle      = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	textStyle        = lipgloss.NewStyle().Foreground(textColor)
	secondaryStyle   = lipgloss.NewStyle().Foreground(secondaryColor)
	containerStyle   = lipgloss.NewStyle().Padding(1, 2)
)
