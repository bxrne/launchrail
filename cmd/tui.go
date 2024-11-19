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

type model struct {
	logger         *charm_log.Logger
	cfg            *config.Config
	phase          phase
	promptedData   promptedData
	earthList      list.Model
	atmosphereList list.Model
	filePicker     filepicker.Model // WARN: independent file picker
	textInput      textinput.Model
	windowWidth    int
	windowHeight   int
}

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

var (
	accentColor      = lipgloss.Color("#5a56e0")
	secondaryColor   = lipgloss.Color("#888888")
	titleStyle       = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	descriptionStyle = lipgloss.NewStyle().Foreground(secondaryColor).MarginBottom(1)
	promptStyle      = lipgloss.NewStyle().Foreground(accentColor)
	linkStyle        = lipgloss.NewStyle().Foreground(accentColor).Underline(true)
	textStyle        = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	footerStyle      = lipgloss.NewStyle().Foreground(secondaryColor)
	textInputStyle   = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	containerStyle   = lipgloss.NewStyle().Padding(1, 2)
)

func initialModel(cfg *config.Config, logger *charm_log.Logger) model {
	fp := filepicker.New()
	fp.AutoHeight = false // INFO: Controlled in update
	fp.Height = 5

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
	atmosphereList := list.New(atmosphereItems, listd, 15, 4)
	atmosphereList.Title = "Choose an Atmosphere model"

	textInput := textinput.New()
	textInput.Placeholder = "Enter a value"
	textInput.Prompt = "Enter a value:"
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
		contentHeight := m.contentHeight()

		m.filePicker.Height = contentHeight - 2 // WARN: -1 for the prompt
		m.earthList.SetSize(msg.Width, contentHeight)
		m.atmosphereList.SetSize(msg.Width, contentHeight)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.logger.Debug("Ctrl+C or 'q' pressed, quitting")
			return m, tea.Quit

		case "enter":
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

	selected, file := m.filePicker.DidSelectFile(msg)
	if selected {
		switch m.phase {
		case selectOpenRocketFile:
			m.promptedData.rocketFile = file
			m.phase = selectMotorThrustFile
		case selectMotorThrustFile:
			m.promptedData.motorFile = file
			m.phase = selectEarthModel
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	title := titleStyle.Render("🚀 Launchrail")
	desc := descriptionStyle.Render("Risk-neutral trajectory simulation for sounding rockets.")
	header := fmt.Sprintf("%s\n%s", title, desc)

	githubText := linkStyle.Render(m.cfg.App.Repo)
	licenseText := footerStyle.Render(m.cfg.App.License)
	versionText := footerStyle.Render(m.cfg.App.Version)
	footer := fmt.Sprintf("%s | %s | %s", versionText, licenseText, githubText)

	var content string
	switch m.phase {
	case selectOpenRocketFile:
		content = m.renderFilePicker("Pick an OpenRocket design file (.ork):", []string{"ork"})
	case selectMotorThrustFile:
		content = m.renderFilePicker("Pick Motor thrust curve file (.eng):", []string{"eng"})
	case selectEarthModel:
		content = m.earthList.View()
	case selectAtmosphericalModel:
		content = m.atmosphereList.View()
	case enterLatLong:
		m.textInput.Prompt = "Enter launch coordinates (HDMS):"
		m.textInput.Placeholder = "N 40 45 36.0 W 73 59 02.4"
		content = m.textInput.View()
	case enterElevation:
		m.textInput.Prompt = "Enter Elevation (m):"
		m.textInput.Placeholder = "42"
		content = m.textInput.View()
	case confirmPhase:
		content = m.confirmView()
	}

	return containerStyle.Render(lipgloss.JoinVertical(lipgloss.Top, header, content, footer))
}

func (m model) renderFilePicker(prompt string, allowedTypes []string) string {
	m.filePicker.FileAllowed = true
	m.filePicker.DirAllowed = false
	m.filePicker.AllowedTypes = allowedTypes
	q := promptStyle.Render(prompt)
	return lipgloss.JoinVertical(lipgloss.Top, q, m.filePicker.View())
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
	title := titleStyle.Render("🚀 Launchrail")
	desc := descriptionStyle.Render("Risk-neutral trajectory simulation for sounding rockets.")
	header := fmt.Sprintf("%s\n%s", title, desc)
	return strings.Count(header, "\n") + 1
}

func (m model) footerHeight() int {
	return 1 // Footer is single line
}

func (m model) contentHeight() int {
	headerH := m.headerHeight()
	footerH := m.footerHeight()
	paddingV := containerStyle.GetPaddingTop() + containerStyle.GetPaddingBottom()
	return m.windowHeight - headerH - footerH - paddingV
}
