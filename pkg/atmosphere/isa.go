package atmosphere

import (
	"math"
	"sync"
)

// ISAModel implements the International Standard Atmosphere
type ISAModel struct {
	cache map[float64]AtmosphereData
	mu    sync.RWMutex
}

type AtmosphereData struct {
	Density     float64
	Temperature float64
	Pressure    float64
}

const (
	R_AIR      = 287.05287 // Specific gas constant for air [J/(kg·K)]
	G_0        = 9.80665   // Gravitational acceleration [m/s^2]
	RHO_0      = 1.225     // Sea level density [kg/m^3]
	T_0        = 288.15    // Sea level temperature [K]
	P_0        = 101325.0  // Sea level pressure [Pa]
	GAMMA      = 1.4       // Ratio of specific heats for air
	LAPSE_RATE = -0.0065   // Temperature lapse rate [K/m]
)

func NewISAModel() *ISAModel {
	return &ISAModel{
		cache: make(map[float64]AtmosphereData),
	}
}

// GetAtmosphere returns atmospheric data for a given altitude using memoization
func (isa *ISAModel) GetAtmosphere(altitude float64) AtmosphereData {
	// Round altitude to nearest meter for caching
	roundedAlt := math.Round(altitude)

	isa.mu.RLock()
	if data, exists := isa.cache[roundedAlt]; exists {
		isa.mu.RUnlock()
		return data
	}
	isa.mu.RUnlock()

	// Calculate new values
	temp := T_0 + LAPSE_RATE*altitude
	pressure := P_0 * math.Pow(temp/T_0, -G_0/(LAPSE_RATE*R_AIR))
	density := pressure / (R_AIR * temp)

	data := AtmosphereData{
		Density:     density,
		Temperature: temp,
		Pressure:    pressure,
	}

	// Cache the result
	isa.mu.Lock()
	isa.cache[roundedAlt] = data
	isa.mu.Unlock()

	return data
}

// GetSpeedOfSound calculates speed of sound at given altitude
func (isa *ISAModel) GetSpeedOfSound(altitude float64) float64 {
	atm := isa.GetAtmosphere(altitude)
	return math.Sqrt(GAMMA * R_AIR * atm.Temperature)
}
