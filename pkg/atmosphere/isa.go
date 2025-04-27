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

// AtmosphereData contains atmospheric properties at a given altitude
type AtmosphereData struct {
	Density     float64
	Temperature float64
	Pressure    float64
	SoundSpeed  float64
}

// NewISAModel creates a new ISAModel with the given configuration
func NewISAModel(cfg *config.ISAConfiguration) *ISAModel {
	return &ISAModel{
		cache: make(map[float64]AtmosphereData),
		cfg:   cfg,
	}
}

// GetTemperature calculates the temperature at a given altitude
func (isa *ISAModel) GetTemperature(altitude float64) float64 {
	return isa.cfg.SeaLevelTemperature + isa.cfg.TemperatureLapseRate*altitude
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
	pressure := 0.0
	density := 0.0
	soundSpeed := 0.0

	if temp > 0 { // Check temperature is valid for calculations
		pressure = isa.cfg.SeaLevelPressure * math.Pow(temp/isa.cfg.SeaLevelTemperature, -isa.cfg.GravitationalAccel/(isa.cfg.TemperatureLapseRate*isa.cfg.SpecificGasConstant))
		density = pressure / (isa.cfg.SpecificGasConstant * temp)
		soundSpeed = math.Sqrt(isa.cfg.RatioSpecificHeats * isa.cfg.SpecificGasConstant * temp)
	} // Else, pressure, density, soundSpeed remain 0

	data := AtmosphereData{
		Density:     density,
		Temperature: temp,
		Pressure:    pressure,
		SoundSpeed:  soundSpeed, // Store calculated sound speed
	}

	// Cache the result
	isa.mu.Lock()
	isa.cache[roundedAlt] = data
	isa.mu.Unlock()

	return data
}

// GetSpeedOfSound calculates speed of sound at given altitude
func (isa *ISAModel) GetSpeedOfSound(altitude float64) float64 {
	// Retrieve cached/calculated data including sound speed
	return isa.GetAtmosphere(altitude).SoundSpeed
}
