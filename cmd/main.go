package main

import (
	"fmt"
	"os"

	"github.com/bxrne/launchrail/pkg/config"
	"github.com/bxrne/launchrail/pkg/logger"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charm_log "github.com/charmbracelet/log"
)

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

type model struct {
	spinner spinner.Model
	width   int
	height  int
	logger  *charm_log.Logger
	cfg     *config.Config
}

func initialModel(cfg *config.Config, logger *charm_log.Logger) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner: sp,
		logger:  logger,
		cfg:     cfg,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.logger.Debug("Ctrl+C or 'q' pressed, quitting")
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.logger.Debug("Window size message received")
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m *model) headerView() string {
	title := titleStyle.Render("🚀 Launchrail")
	desc := headerStyle.Render("Risk-neutral trajectory simulation for sounding rockets via the Black-Scholes model.\n'ctrl+c' or 'q' to quit.")
	return fmt.Sprintf("%s\n%s\n", title, desc)
}

func (m *model) footerView() string {
	githubText := footerLinkStyle.Render(m.cfg.App.Repo)
	licenseText := footerStyle.Render(m.cfg.App.License)
	versionText := footerStyle.Render(m.cfg.App.Version)
	return fmt.Sprintf("%s | %s | %s\n", versionText, licenseText, githubText)
}

func (m model) View() string {
	header := containerStyle.Render(m.headerView())
	footer := containerStyle.Render(m.footerView())
	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - 2 // Adjust for padding

	var content string
	content = contentStyle.Height(contentHeight).Render(m.spinner.View())

	return fmt.Sprintf("%s\n%s\n%s", header, content, footer)

}

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error loading configuration: %v", err)
		os.Exit(1) // WARNING: Process exit
	}

	log, err := logger.GetLogger(cfg.Logs.File)
	if err != nil {
		fmt.Printf("Error getting logger: %v", err)
		os.Exit(1) // WARNING: Process exit
	}

	log.Info("Starting Launchrail application")

	p := tea.NewProgram(initialModel(cfg, log))
	if _, err := p.Run(); err != nil {
		log.Errorf("Error running program: %v", err)
		os.Exit(1)
	}

	log.Info("Exiting Launchrail application")
}
