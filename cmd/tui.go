package main

import (
	"fmt"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/integrations/openrocket"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charm_log "github.com/charmbracelet/log"
)

type model struct {
	filePicker   filepicker.Model
	textInput    textinput.Model
	height       int
	logger       *charm_log.Logger
	cfg          *config.Config
	phase        phase
	promptedData promptedData
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
	latitudeValue float64
	longitudeValue float64
	altitudeValue float64
	
}

var (
	accentColor    = lipgloss.Color("#FFA500")
	promptStyle    = lipgloss.NewStyle().Foreground(accentColor)
	titleStyle     = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	descStyle      = lipgloss.NewStyle().Bold(true).PaddingTop(1)
	linkStyle      = lipgloss.NewStyle().Foreground(accentColor).Underline(true)
	textStyle      = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	containerStyle = lipgloss.NewStyle().Padding(0).MarginTop(1).MarginLeft(1)
	footerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	textInputStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
)

func initialModel(cfg *config.Config, logger *charm_log.Logger) model {
	fp := filepicker.New()
	fp.AutoHeight = false
	fp.Height = 5

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
			m.promptedData.rocketFile = file
			m.phase = selectMotorThrustFile
		case selectMotorThrustFile:
			m.promptedData.motorFile = file
			m.phase = finalPhase
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	header := containerStyle.Render(m.headerView())
	footer := containerStyle.Render(m.footerView())

	var content string
	switch m.phase {
	case selectOpenRocketFile:
		m.filePicker.Height = m.height - 4
		m.filePicker.FileAllowed = true
		m.filePicker.DirAllowed = false
		m.filePicker.AllowedTypes = []string{"ork"}

		q := containerStyle.Render(promptStyle.Render("Pick an OpenRocket design file (.ork):"))
		content = lipgloss.JoinVertical(lipgloss.Top, q, m.filePicker.View())

	case selectMotorThrustFile:
		m.filePicker.Height = m.height - 4
		m.filePicker.FileAllowed = true
		m.filePicker.DirAllowed = false
		m.filePicker.AllowedTypes = []string{"eng"}

		q := containerStyle.Render(promptStyle.Render("Pick Motor thrust curve file (.eng):"))
		content = lipgloss.JoinVertical(lipgloss.Top, q, m.filePicker.View())

	case finalPhase:
		content = m.finalView()
	}

	return containerStyle.Render(header, content, footer)
}

func (m model) headerView() string {
	title := titleStyle.Render("🚀 Launchrail")
	desc := descStyle.Render("Risk-neutral trajectory simulation for sounding rockets.")
	instructions := "Press 'ctrl+c' or 'q' to quit.\n"
	return fmt.Sprintf("%s\n%s\n%s", title, desc, instructions)
}

func (m *model) footerView() string {
	githubText := linkStyle.Render(m.cfg.App.Repo)
	licenseText := footerStyle.Render(m.cfg.App.License)
	versionText := footerStyle.Render(m.cfg.App.Version)
	return fmt.Sprintf("%s | %s | %s\n", versionText, licenseText, githubText)
}

func (m model) finalView() string {
	orkData, err := openrocket.Decompress(m.promptedData.rocketFile)
	if err != nil {
		return fmt.Sprintf("Error reading OpenRocket file: %v", err)
	}

	motorData, err := components.NewSolidMotor(m.promptedData.motorFile)
	if err != nil {
		return fmt.Sprintf("Error reading Motor file: %v", err)
	}

	rocket := components.NewRocket(orkData, motorData)

	return containerStyle.Render(rocket.Info())
}
