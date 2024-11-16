package entities

import "github.com/bxrne/launchrail/pkg/physics"

type Environment struct {
	Latitude         float64
	Longitude        float64
	Altitude         float64
	Gravity          float64
	Pressure         float64
	SeaLevelAltitude float64
	Atmosphere       physics.Atmosphere
}
