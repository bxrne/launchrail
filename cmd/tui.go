package main

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/config"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charm_log "github.com/charmbracelet/log"
)

type model struct {
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

var (
	accentStyle    = lipgloss.Color("#FFA500")
	promptStyle    = lipgloss.NewStyle().Foreground(accentStyle)
	titleStyle     = lipgloss.NewStyle().Foreground(accentStyle).Bold(true)
	descStyle      = lipgloss.NewStyle().Bold(true).PaddingTop(1)
	footerStyle    = lipgloss.NewStyle().Foreground(accentStyle).Underline(true)
	textInputStyle = lipgloss.NewStyle().Foreground(accentStyle).Bold(true)
	textStyle      = lipgloss.NewStyle().Foreground(accentStyle).Bold(true)
	containerStyle = lipgloss.NewStyle().Padding(0).Margin(1, 2)
)

func initialModel(cfg *config.Config, logger *charm_log.Logger) model {
	fp := filepicker.New()

	ti := textinput.New()
	ti.Placeholder = "Enter value here..."
	ti.PromptStyle = textInputStyle
	ti.TextStyle = textStyle
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
	return tea.Batch(m.filePicker.Init())
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
		content = promptStyle.Render("Pick an OpenRocket design file (.ork):")
		content = lipgloss.JoinVertical(lipgloss.Top, content, m.filePicker.View())

	case selectMotorThrustFile:
		m.filePicker.Height = m.height - 4
		m.filePicker.FileAllowed = true
		m.filePicker.DirAllowed = false
		m.filePicker.AllowedTypes = []string{"eng"}
		content = promptStyle.Render("Pick Motor thrust curve file (.eng):")
		content = lipgloss.JoinVertical(lipgloss.Top, content, m.filePicker.View())

	case finalPhase:
		content = m.finalView()
	}

	return containerStyle.Render(lipgloss.JoinVertical(lipgloss.Top, header, content, footer))
}

func (m model) headerView() string {
	title := titleStyle.Render("🚀 Launchrail")
	desc := descStyle.Render("Risk-neutral trajectory simulation for sounding rockets.\nPress 'ctrl+c' or 'q' to quit.\n")
	return fmt.Sprintf("%s\n%s", title, desc)
}

func (m model) footerView() string {
	githubText := footerStyle.Render(m.cfg.App.Repo)
	licenseText := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(m.cfg.App.License)
	versionText := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(m.cfg.App.Version)
	return fmt.Sprintf("%s | %s | %s\n", versionText, licenseText, githubText)
}

func (m model) finalView() string {
	return fmt.Sprintf("Final Rocket Configuration:\nRocket File: %s\nMotor File: %s", m.data.rocketFile, m.data.motorFile)
}
