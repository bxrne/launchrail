package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bxrne/launchrail/pkg/config"
	"github.com/bxrne/launchrail/pkg/logger"
	"github.com/charmbracelet/bubbles/textinput"
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
	containerStyle  = lipgloss.NewStyle().Margin(1, 2)
	contentStyle    = lipgloss.NewStyle().Padding(1, 2).Margin(1, 2)
)

type Phase int

const (
	PhaseNumericInput Phase = iota
	PhaseTextInput
	PhaseDateInput
	PhaseSummary
)

type model struct {
	width  int
	height int
	logger *charm_log.Logger
	cfg    *config.Config
	phase  Phase
	// Inputs
	numericInput textinput.Model
	textInput    textinput.Model
	dateInput    textinput.Model
	// Collected values
	numericValue float64
	textValue    string
	dateValue    time.Time
}

func initialModel(cfg *config.Config, logger *charm_log.Logger) model {
	numericInput := textinput.New()
	numericInput.Placeholder = "Enter a number"
	numericInput.Focus()
	numericInput.CharLimit = 20
	numericInput.Width = 20

	return model{
		logger:       logger,
		cfg:          cfg,
		phase:        PhaseNumericInput,
		numericInput: numericInput,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			switch m.phase {
			case PhaseNumericInput:
				// Validate numeric input
				var value float64
				_, err := fmt.Sscanf(m.numericInput.Value(), "%f", &value)
				if err != nil {
					m.numericInput.SetValue("")
					m.numericInput.Placeholder = "Invalid number, please enter again"
					return m, nil
				}
				m.numericValue = value
				m.phase = PhaseTextInput

				m.textInput = textinput.New()
				m.textInput.Placeholder = "Enter text"
				m.textInput.Focus()
				m.textInput.CharLimit = 100
				m.textInput.Width = 40

			case PhaseTextInput:
				m.textValue = m.textInput.Value()
				if m.textValue == "" {
					m.textInput.Placeholder = "Please enter some text"
					return m, nil
				}
				m.phase = PhaseDateInput

				m.dateInput = textinput.New()
				m.dateInput.Placeholder = "Enter date (YYYY-MM-DD)"
				m.dateInput.Focus()
				m.dateInput.CharLimit = 10
				m.dateInput.Width = 20

			case PhaseDateInput:
				dateStr := m.dateInput.Value()
				date, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					m.dateInput.SetValue("")
					m.dateInput.Placeholder = "Invalid date, please enter YYYY-MM-DD"
					return m, nil
				}
				m.dateValue = date
				m.phase = PhaseSummary
			case PhaseSummary:
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.phase {
	case PhaseNumericInput:
		m.numericInput, cmd = m.numericInput.Update(msg)
	case PhaseTextInput:
		m.textInput, cmd = m.textInput.Update(msg)
	case PhaseDateInput:
		m.dateInput, cmd = m.dateInput.Update(msg)
	}

	return m, cmd
}
func (m model) View() string {
	header := m.headerView()
	footer := m.footerView()
	var content string

	switch m.phase {
	case PhaseNumericInput:
		content = "Please enter a number:\n\n" + m.numericInput.View()
	case PhaseTextInput:
		content = "Please enter text:\n\n" + m.textInput.View()
	case PhaseDateInput:
		content = "Please enter a date (YYYY-MM-DD):\n\n" + m.dateInput.View()
	case PhaseSummary:
		content = fmt.Sprintf(
			"Summary:\n\nNumber: %f\nText: %s\nDate: %s",
			m.numericValue, m.textValue, m.dateValue.Format("2006-01-02"),
		)
	}

	// Apply contentStyle to the content
	content = contentStyle.Render(content)

	// Wrap the entire view with containerStyle
	return containerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, header, content, footer),
	)
}
func (m *model) headerView() string {
	title := titleStyle.Render("🚀 Launchrail")
	desc := headerStyle.Render("Risk-neutral trajectory simulation for sounding rockets via the Black-Scholes model.\n'Esc' or 'Ctrl+c' to quit.")
	return fmt.Sprintf("%s\n%s\n", title, desc)
}

func (m *model) footerView() string {
	githubText := footerLinkStyle.Render(m.cfg.App.Repo)
	licenseText := footerStyle.Render(m.cfg.App.License)
	versionText := footerStyle.Render(m.cfg.App.Version)
	return fmt.Sprintf("%s | %s | %s\n", versionText, licenseText, githubText)
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
