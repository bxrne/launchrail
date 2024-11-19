package tui

import (
	"strconv"

	"github.com/bxrne/launchrail/pkg/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/flopp/go-coordsparser"
)

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
		m.phase = enterMotorMass
	}
}

func (m *model) handleEnterKey() tea.Cmd {
	switch m.phase {
	case enterMotorMass:
		massValue := m.textInput.Value()
		mass, err := strconv.ParseFloat(massValue, 64)
		if err != nil {
			m.logger.Fatalf("Error parsing mass: %v", err)
		}

		m.textInput.Reset()
		m.promptedData.motorMass = mass
		m.phase = selectEarthModel

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
