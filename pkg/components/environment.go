package components

type Atmosphere struct {
	Temperature []float64
	Pressure    []float64
	Density     []float64
}

type Environment struct {
	Latitude   float64
	Longitude  float64
	Altitude   float64
	Gravity    float64
	Pressure   float64
	Atmosphere Atmosphere
}

func NewEnvironment(latitude, longitude, altitude, pressure float64) Environment {
	return Environment{
		Latitude:   latitude,
		Longitude:  longitude,
		Altitude:   altitude,
		Gravity:    9.81, // TODO: Implement gravity model
		Pressure:   pressure,
		Atmosphere: Atmosphere{}, // TODO: Implement atmosphere model
	}
}
