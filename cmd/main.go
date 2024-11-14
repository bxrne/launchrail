package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/pkg/logger"
	"github.com/bxrne/launchrail/pkg/utils"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// INFO: Exec constants
const (
	version     = "v0.0"
	githubLink  = "https://github.com/bxrne/launchrail"
	license     = "GNU GPL-3.0"
	logFilePath = "launchrail.log"
)

// INFO: Lipgloss styles
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

type logData struct {
	logger       *log.Logger
	logs         []string
	showLogPanel bool
}

type model struct {
	spinner spinner.Model
	width   int
	height  int
	logData logData
}

func initialModel(logger *log.Logger) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner: sp,
		logData: logData{
			logger:       logger,
			logs:         []string{},
			showLogPanel: false,
		},
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
	m.logData.logger.Debugf("Update called with message: %v", msg)
	m.logData.logs = append(m.logData.logs, fmt.Sprintf("Update called with message: %v", msg))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.logData.logger.Debugf("KeyMsg received: %s", msg.String())
		m.logData.logs = append(m.logData.logs, fmt.Sprintf("KeyMsg received: %s", msg.String()))
		switch msg.String() {
		case "ctrl+c", "q":
			m.logData.logger.Info("Exiting application")
			m.logData.logs = append(m.logData.logs, "Exiting application")
			return m, tea.Quit
		case "l":
			m.logData.showLogPanel = !m.logData.showLogPanel
		}
	case tea.WindowSizeMsg:
		m.logData.logger.Debugf("WindowSizeMsg received: width=%d, height=%d", msg.Width, msg.Height)
		m.logData.logs = append(m.logData.logs, fmt.Sprintf("WindowSizeMsg received: width=%d, height=%d", msg.Width, msg.Height))
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m model) View() string {
	m.logData.logger.Debug("View called")
	m.logData.logs = append(m.logData.logs, "View called")

	header := containerStyle.Render(headerView())
	footer := containerStyle.Render(footerView())
	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 2 // Adjust for padding

	var content string
	if m.logData.showLogPanel {
		logPanel := logStyle.Render(utils.Tail(m.logData.logs, 10))
		contentHeight = contentHeight - logPanelHeight

		content = contentStyle.Height(contentHeight).Render(m.spinner.View())
		return fmt.Sprintf("%s\n%s\n%s\n%s", header, content, logPanel, footer)
	} else {
		content = contentStyle.Height(contentHeight).Render(m.spinner.View())
		return fmt.Sprintf("%s\n%s\n%s", header, content, footer)
	}
}

func main() {
	log := logger.GetLogger(logFilePath)
	log.Info("Starting Launchrail application")

	p := tea.NewProgram(initialModel(log), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Errorf("Error running program: %v", err)
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}

	log.Info("Exiting Launchrail application")
}
