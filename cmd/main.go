package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/pkg/logger"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	version    = "v0.0"
	githubLink = "https://github.com/bxrne/launchrail"
	license    = "GNU GPL-3.0"
)

var (
	accentColor     = lipgloss.Color("#FFA500")
	titleStyle      = lipgloss.NewStyle().Foreground(accentColor).Bold(true).Padding(1, 2).MarginBottom(1)
	headerStyle     = lipgloss.NewStyle().Bold(true).Padding(0, 2)
	footerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Align(lipgloss.Center).Padding(0, 2)
	footerLinkStyle = lipgloss.NewStyle().Foreground(accentColor).Underline(true).Padding(0, 2)
	containerStyle  = lipgloss.NewStyle().Margin(1, 2)
	contentStyle    = lipgloss.NewStyle().Padding(1, 2).Margin(1, 2)
)

type model struct {
	spinner spinner.Model
	width   int
	height  int
}

func initialModel() model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner: sp,
	}
}

func headerView() string {
	title := titleStyle.Render("🚀 Launchrail")
	versionText := headerStyle.Render("Risk-neutral trajectory simulation for sounding rockets via the Black-Scholes model.\n'ctrl+c' or 'q' to quit.")
	return fmt.Sprintf("%s\n%s\n", title, versionText)
}

func footerView() string {
	githubText := footerLinkStyle.Render(githubLink)
	licenseText := footerStyle.Render(license)
	versionText := footerStyle.Render(version)
	return fmt.Sprintf("%s | %s | %s\n", versionText, licenseText, githubText)
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log := logger.GetLogger()
	log.Debugf("Update called with message: %v", msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Debugf("KeyMsg received: %s", msg.String())
		switch msg.String() {
		case "ctrl+c", "q":
			log.Info("Exiting application")
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		log.Debugf("WindowSizeMsg received: width=%d, height=%d", msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m model) View() string {
	log := logger.GetLogger()
	log.Debug("View called")

	header := containerStyle.Render(headerView())
	footer := containerStyle.Render(footerView())
	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 2 // Adjust for padding
	content := contentStyle.Height(contentHeight).Render(m.spinner.View())

	return fmt.Sprintf("%s\n%s\n%s", header, content, footer)
}

func main() {
	log := logger.GetLogger()
	log.Info("Starting Launchrail application")

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Errorf("Error running program: %v", err)
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}

	log.Info("Exiting Launchrail application")
}
