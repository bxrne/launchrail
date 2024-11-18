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
