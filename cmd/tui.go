package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/integrations/openrocket"
	"github.com/bxrne/launchrail/pkg/simulation"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	charm_log "github.com/charmbracelet/log"
	"github.com/flopp/go-coordsparser"
)

var (
	accentColor      = lipgloss.Color("#5a56e0")
	secondaryColor   = lipgloss.Color("#888888")
	titleStyle       = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	descriptionStyle = lipgloss.NewStyle().Foreground(secondaryColor).MarginBottom(1)
	promptStyle      = lipgloss.NewStyle().Foreground(accentColor)
	linkStyle        = lipgloss.NewStyle().Foreground(accentColor).Underline(true)
	accentStyle      = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	textStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	secondaryStyle   = lipgloss.NewStyle().Foreground(secondaryColor)
	containerStyle   = lipgloss.NewStyle().Padding(1, 2)
)

type phase int

const (
	selectOpenRocketFile phase = iota
	selectMotorThrustFile
	selectEarthModel
	selectAtmosphericalModel
	enterLatLong
	enterElevation
	confirmPhase
)

type promptedData struct {
	rocketFile       string
	motorFile        string
	earthModel       components.Earth
	atmosphericModel components.Atmosphere
	latitude         float64
	longitude        float64
	elevation        float64
}

type model struct {
	logger         *charm_log.Logger
	cfg            *config.Config
	phase          phase
	promptedData   promptedData
	earthList      list.Model
	atmosphereList list.Model
	filePicker     filepicker.Model
	textInput      textinput.Model
	windowWidth    int
	windowHeight   int
}

func initialModel(cfg *config.Config, logger *charm_log.Logger) model {
	fp := filepicker.New()
	fp.AutoHeight = false
	fp.Height = 0

	listd := list.NewDefaultDelegate()
	earthItems := []list.Item{
		components.FlatEarth,
		components.SphericalEarth,
		components.TopographicalEarth,
	}
	earthList := list.New(earthItems, listd, 0, 0)
	earthList.Title = "Choose an Earth model"

	atmosphereItems := []list.Item{
		components.StandardAtmosphere,
		components.ForecastAtmosphere,
	}
	atmosphereList := list.New(atmosphereItems, listd, 0, 0)
	atmosphereList.Title = "Choose an Atmosphere model"

	textInput := textinput.New()
	textInput.TextStyle = textStyle
	textInput.PlaceholderStyle = secondaryStyle
	textInput.PromptStyle = accentStyle
	textInput.Focus()

	return model{
		filePicker:     fp,
		logger:         logger,
		cfg:            cfg,
		phase:          selectOpenRocketFile,
		earthList:      earthList,
		atmosphereList: atmosphereList,
		textInput:      textInput,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.filePicker.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width
		m.updateComponentSizes()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.logger.Debug("Ctrl+C or 'q' pressed, quitting")
			return m, tea.Quit

		case "enter":
			cmd := m.handleEnterKey()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		switch m.phase {
		case selectEarthModel:
			var cmd tea.Cmd
			m.earthList, cmd = m.earthList.Update(msg)
			cmds = append(cmds, cmd)
		case selectAtmosphericalModel:
			var cmd tea.Cmd
			m.atmosphereList, cmd = m.atmosphereList.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	var tiCmd tea.Cmd
	m.textInput, tiCmd = m.textInput.Update(msg)
	cmds = append(cmds, tiCmd)

	var fpCmd tea.Cmd
	m.filePicker, fpCmd = m.filePicker.Update(msg)
	cmds = append(cmds, fpCmd)

	if selected, file := m.filePicker.DidSelectFile(msg); selected {
		m.handleFileSelection(file)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateComponentSizes() {
	contentHeight := m.contentHeight()
	m.filePicker.Height = contentHeight - 2 // 2 = prompt + padding
	m.earthList.SetSize(m.windowWidth, contentHeight)
	m.atmosphereList.SetSize(m.windowWidth, contentHeight)
}

func (m *model) handleFileSelection(file string) {
	switch m.phase {
	case selectOpenRocketFile:
		m.promptedData.rocketFile = file
		m.phase = selectMotorThrustFile
	case selectMotorThrustFile:
		m.promptedData.motorFile = file
		m.phase = selectEarthModel
	}
}

func (m *model) handleEnterKey() tea.Cmd {
	switch m.phase {
	case selectEarthModel:
		m.promptedData.earthModel = components.Earth(m.earthList.Index())
		m.phase = selectAtmosphericalModel

	case selectAtmosphericalModel:
		m.promptedData.atmosphericModel = components.Atmosphere(m.atmosphereList.Index())
		m.phase = enterLatLong

	case enterLatLong:
		lat, long, err := coordsparser.ParseHDMS(m.textInput.Value())
		if err != nil {
			m.logger.Fatalf("Error parsing coordinates: %v", err)
		}
		m.textInput.Reset()
		m.promptedData.latitude = lat
		m.promptedData.longitude = long
		m.phase = enterElevation

	case enterElevation:
		elevationValue := m.textInput.Value()
		elev, err := strconv.ParseFloat(elevationValue, 64)
		if err != nil {
			m.logger.Fatalf("Error parsing elevation: %v", err)
		}
		m.textInput.Reset()
		m.promptedData.elevation = elev
		m.phase = confirmPhase
	}
	return nil
}

func (m model) View() string {
	header := m.renderHeader()
	content := m.renderContent()
	footer := m.renderFooter()

	if m.phase == enterLatLong || m.phase == enterElevation {
		remainingHeight := m.windowHeight - m.headerHeight() - m.footerHeight() - containerStyle.GetPaddingTop() - containerStyle.GetPaddingBottom()
		content = m.fillRemainingSpace(content, remainingHeight-1) // -1 for the text input line
	}

	return containerStyle.Render(lipgloss.JoinVertical(lipgloss.Top, header, content, footer))
}

func (m model) renderHeader() string {
	title := titleStyle.Render("🚀 Launchrail")
	desc := descriptionStyle.Render("Risk-neutral trajectory simulation for sounding rockets.")
	return fmt.Sprintf("%s\n%s", title, desc)
}

func (m model) renderFooter() string {
	githubText := linkStyle.Render(m.cfg.App.Repo)
	licenseText := secondaryStyle.Render(m.cfg.App.License)
	versionText := secondaryStyle.Render(m.cfg.App.Version)
	return fmt.Sprintf("%s | %s | %s", versionText, licenseText, githubText)
}

func (m model) renderContent() string {
	switch m.phase {
	case selectOpenRocketFile:
		return m.renderFilePicker("Pick an OpenRocket design file (.ork):", []string{"ork"})
	case selectMotorThrustFile:
		return m.renderFilePicker("Pick Motor thrust curve file (.eng):", []string{"eng"})
	case selectEarthModel:
		return m.earthList.View()
	case selectAtmosphericalModel:
		return m.atmosphereList.View()
	case enterLatLong:
		m.textInput.Prompt = "Enter launch coordinates (HDMS):"
		m.textInput.Placeholder = " N 40 45 36.0 W 73 59 02.4"
		return m.textInput.View()
	case enterElevation:
		m.textInput.Prompt = "Enter Elevation (m):"
		m.textInput.Placeholder = " 42"
		return m.textInput.View()
	case confirmPhase:
		return m.confirmView()
	default:
		return ""
	}
}

func (m model) renderFilePicker(prompt string, allowedTypes []string) string {
	m.filePicker.FileAllowed = true
	m.filePicker.DirAllowed = false
	m.filePicker.AllowedTypes = allowedTypes
	q := promptStyle.Render(prompt)
	return lipgloss.JoinVertical(lipgloss.Top, q, m.filePicker.View())
}

func (m model) fillRemainingSpace(content string, remainingHeight int) string {
	contentHeight := strings.Count(content, "\n") + 1
	emptyLines := remainingHeight - contentHeight
	if emptyLines <= 0 {
		return content
	}
	return content + strings.Repeat("\n", emptyLines)
}

func (m model) confirmView() string {
	orkData, err := openrocket.Decompress(m.promptedData.rocketFile)
	if err != nil {
		return fmt.Sprintf("Error reading OpenRocket file: %v", err)
	}

	motorData, err := components.NewSolidMotor(m.promptedData.motorFile)
	if err != nil {
		return fmt.Sprintf("Error reading Motor file: %v", err)
	}

	rocket := components.NewRocket(orkData, motorData)
	environment := simulation.NewEnvironment(0, 0, 0, 9.81, 101325, &m.promptedData.atmosphericModel, &m.promptedData.earthModel)
	sim := simulation.NewSimulation(rocket, *environment)
	return sim.Info()
}

func (m model) headerHeight() int {
	return strings.Count(m.renderHeader(), "\n") + 1
}

func (m model) footerHeight() int {
	return strings.Count(m.renderFooter(), "\n") + 1
}

func (m model) contentHeight() int {
	headerH := m.headerHeight()
	footerH := m.footerHeight()
	paddingV := containerStyle.GetPaddingTop() + containerStyle.GetPaddingBottom()
	return m.windowHeight - headerH - footerH - paddingV
}
