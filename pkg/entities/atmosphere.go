package entities

type Atmosphere interface {
	GetDensity(altitude float64) float64
	GetPressure(altitude float64) float64
	GetTemperature(altitude float64) float64
	GetSpeedOfSound(altitude float64) float64
}

type StandardAtmosphere struct {
	// TODO: Implementation
}
