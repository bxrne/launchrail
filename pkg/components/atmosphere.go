package components

type Atmosphere int

const (
	StandardAtmosphere Atmosphere = iota
	ForecastAtmosphere
)

func (a Atmosphere) String() string {
	switch a {
	case StandardAtmosphere:
		return "Standard Atmosphere"
	case ForecastAtmosphere:
		return "Forecast Atmosphere"
	default:
		return "Unknown Atmosphere Model"
	}
}

// INFO: Implementing the list.Item interface
func (a Atmosphere) Title() string { return a.String() }
func (a Atmosphere) Description() string {
	switch a {
	case StandardAtmosphere:
		return "Constant temperature, pressure and density"
	case ForecastAtmosphere:
		return "Variable temperature, pressure and density"
	default:
		return ""
	}
}
func (a Atmosphere) FilterValue() string { return a.String() }
