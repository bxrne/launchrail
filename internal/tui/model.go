package tui

import (
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	charm_log "github.com/charmbracelet/log"

	tea "github.com/charmbracelet/bubbletea"
)

type phase int

const (
	selectOpenRocketFile phase = iota
	selectMotorThrustFile
	enterMotorDryMass
	enterMotorPropellantMass
	selectEarthModel
	selectAtmosphericalModel
	enterLatLong
	enterElevation
	confirmPhase
)

type promptedData struct {
	rocketFile          string
	motorFile           string
	motorDryMass        float64
	motorPropellantMass float64
	earthModel          components.Earth
	atmosphericModel    components.Atmosphere
	latitude            float64
	longitude           float64
	elevation           float64
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

func (m model) Init() tea.Cmd {
	return tea.Batch(m.filePicker.Init())
}

func New(cfg *config.Config, logger *charm_log.Logger) tea.Program {
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

	return *tea.NewProgram(model{
		filePicker:     fp,
		logger:         logger,
		cfg:            cfg,
		phase:          selectOpenRocketFile,
		earthList:      earthList,
		atmosphereList: atmosphereList,
		textInput:      textInput,
	}, tea.WithAltScreen())

}
