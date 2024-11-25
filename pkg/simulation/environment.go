package simulation

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/entities"
)

type Environment struct {
	Latitude  float64
	Longitude float64
	Elevation float64
	Gravity   float64
	Pressure  float64

	Earth *entities.Earth
}

func NewEnvironment(lat, lon, elev, grav, press float64, earth *entities.Earth) *Environment {
	return &Environment{
		Latitude:  lat,
		Longitude: lon,
		Elevation: elev,
		Gravity:   grav,
		Pressure:  press,
		Earth:     earth,
	}
}

func (e *Environment) Info() string {
	return fmt.Sprintf("Latitude: %.2f°\nLongitude: %.2f°\nElevation: %.2f m\nGravity: %.2f m/s²\nPressure: %.2f Pa\n", e.Latitude, e.Longitude, e.Elevation, e.Gravity, e.Pressure)
}
