package main

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/config"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charm_log "github.com/charmbracelet/log"
)

type Data struct {
	rocketFile string
	motorFile  string
}

// model struct to hold the state
type model struct {
	spinner    spinner.Model
	filePicker filepicker.Model
	textInput  textinput.Model
	width      int
	height     int
	logger     *charm_log.Logger
	cfg        *config.Config
	phase      phase
	data       Data
}

type phase int

const (
	selectRocketFile phase = iota
	selectMotorFile
	finalPhase
)

// Initial model to start with spinner and file picker
func initialModel(cfg *config.Config, logger *charm_log.Logger) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	fp := filepicker.New()
	fp.AllowedTypes = []string{".ork", ".eng"} // Allowed file types
	fp.FileAllowed = true
	fp.DirAllowed = false

	ti := textinput.New()
	ti.Placeholder = "Enter value here..."
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	ti.Focus()

	return model{
		spinner:    sp,
		filePicker: fp,
		textInput:  ti,
		logger:     logger,
		cfg:        cfg,
		phase:      selectRocketFile,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.filePicker.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.logger.Debug("Ctrl+C or 'q' pressed, quitting")
			return m, tea.Quit

		case "enter":
			switch m.phase {
			case selectRocketFile:
				if m.filePicker.FileSelected != "" {
					m.data.rocketFile = m.filePicker.FileSelected
					fmt.Println("Rocket design file:", m.data.rocketFile)
					m.phase = selectMotorFile
				}
			case selectMotorFile:
				if m.filePicker.FileSelected != "" {
					m.data.motorFile = m.filePicker.FileSelected
					fmt.Println("Motor thrust curve file:", m.data.motorFile)
					m.phase = finalPhase
					return m, tea.Quit
				}
			}
		}
	}

	var cmds []tea.Cmd
	if m.phase == selectRocketFile || m.phase == selectMotorFile {
		newFilePicker, cmd := m.filePicker.Update(msg)
		m.filePicker = newFilePicker
		cmds = append(cmds, cmd)
	}

	newSpinner, spinnerCmd := m.spinner.Update(msg)
	m.spinner = newSpinner
	cmds = append(cmds, spinnerCmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	header := m.headerView()
	footer := m.footerView()

	var content string
	switch m.phase {
	case selectRocketFile:
		m.filePicker.Height = m.height - 4
		content = m.filePicker.View()
	case selectMotorFile:
		content = m.filePicker.View()
	case finalPhase:
		content = m.finalView()
	}

	return fmt.Sprintf("%s\n%s\n%s", header, content, footer)
}

func (m model) headerView() string {
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true).Padding(1, 2).MarginBottom(1).Render("🚀 Launchrail")
	desc := lipgloss.NewStyle().Bold(true).Padding(0, 2).Render("Risk-neutral trajectory simulation for sounding rockets.\nPress 'ctrl+c' or 'q' to quit.\n")
	return fmt.Sprintf("%s\n%s\n", title, desc)
}

func (m model) footerView() string {
	githubText := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Underline(true).Padding(0, 2).Render(m.cfg.App.Repo)
	licenseText := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Align(lipgloss.Center).Padding(0, 2).Render(m.cfg.App.License)
	versionText := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Align(lipgloss.Center).Padding(0, 2).Render(m.cfg.App.Version)
	return fmt.Sprintf("%s | %s | %s\n", versionText, licenseText, githubText)
}

func (m model) finalView() string {
	return fmt.Sprintf("Final Rocket Configuration:\nRocket File: %s\nMotor File: %s", m.data.rocketFile, m.data.motorFile)
}
