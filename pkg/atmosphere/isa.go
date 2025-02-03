package atmosphere

import (
	"math"
	"sync"

	"github.com/bxrne/launchrail/internal/config"
)

// ISAModel implements the International Standard Atmosphere
type ISAModel struct {
	cache map[float64]AtmosphereData
	cfg   *config.ISAConfiguration
	mu    sync.RWMutex
}

type AtmosphereData struct {
	Density     float64
	Temperature float64
	Pressure    float64
}

func NewISAModel(cfg *config.ISAConfiguration) *ISAModel {
	return &ISAModel{
		cache: make(map[float64]AtmosphereData),
		cfg:   cfg,
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
	temp := isa.cfg.SeaLevelTemperature + isa.cfg.TemperatureLapseRate*altitude // T_0 (sea level temperature) - Lapse rate * altitude
	pressure := isa.cfg.SeaLevelPressure * math.Pow(temp/isa.cfg.SeaLevelTemperature, -isa.cfg.GravitationalAccel/(isa.cfg.TemperatureLapseRate*isa.cfg.SpecificGasConstant))
	density := pressure / (isa.cfg.SpecificGasConstant * temp)

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
	return math.Sqrt(isa.cfg.RatioSpecificHeats * isa.cfg.SpecificGasConstant * atm.Temperature)
}
