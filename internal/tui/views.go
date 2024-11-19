package tui

import (
	"fmt"
	"strings"

	"github.com/bxrne/launchrail/internal/openrocket"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/simulation"
	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	header := m.renderHeader()
	content := m.renderContent()
	footer := m.renderFooter()

	// TODO: This is a bit hacky, but it works for now
	if m.phase == enterLatLong || m.phase == enterElevation || m.phase == enterMotorDryMass || m.phase == enterMotorPropellantMass {
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
	case enterMotorDryMass:
		m.textInput.Prompt = "Enter Motor Dry Mass (kg):"
		m.textInput.Placeholder = " 1.5"
		return m.textInput.View()
	case enterMotorPropellantMass:
		m.textInput.Prompt = "Enter Motor Propellant Mass (kg):"
		m.textInput.Placeholder = " 0.5"
		return m.textInput.View()
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
	m.filePicker.Init()
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

	motorData, err := components.NewSolidMotor(m.promptedData.motorFile, m.promptedData.motorDryMass, m.promptedData.motorPropellantMass)
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
