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

type model struct {
	spinner    spinner.Model
	filePicker filepicker.Model
	textInput  textinput.Model
	width      int
	height     int
	logger     *charm_log.Logger
	cfg        *config.Config
	phase      phase
	data       promptedData
}

type phase int

const (
	selectOpenRocketFile phase = iota
	selectMotorThrustFile
	finalPhase
)

type promptedData struct {
	rocketFile string
	motorFile  string
}

func initialModel(cfg *config.Config, logger *charm_log.Logger) model {
	fp := filepicker.New()

	ti := textinput.New()
	ti.Placeholder = "Enter value here..."
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	ti.Focus()

	return model{
		filePicker: fp,
		textInput:  ti,
		logger:     logger,
		cfg:        cfg,
		phase:      selectOpenRocketFile,
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
		}
	}

	var cmds []tea.Cmd

	var fpCmd tea.Cmd
	m.filePicker, fpCmd = m.filePicker.Update(msg)
	cmds = append(cmds, fpCmd)

	selected, file := m.filePicker.DidSelectFile(msg)
	if selected {
		switch m.phase {
		case selectOpenRocketFile:
			m.data.rocketFile = file
			m.phase = selectMotorThrustFile
		case selectMotorThrustFile:
			m.data.motorFile = file
			m.phase = finalPhase
		}
	}

	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	cmds = append(cmds, spinnerCmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	header := m.headerView()
	footer := m.footerView()

	var content string
	switch m.phase {
	case selectOpenRocketFile:
		m.filePicker.Height = m.height - 4
		m.filePicker.FileAllowed = true
		m.filePicker.DirAllowed = false
		m.filePicker.AllowedTypes = []string{"ork"}
		content = m.filePicker.View()
	case selectMotorThrustFile:
		m.filePicker.Height = m.height - 4
		m.filePicker.FileAllowed = true
		m.filePicker.DirAllowed = false
		m.filePicker.AllowedTypes = []string{"eng"}
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
