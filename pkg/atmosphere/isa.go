package atmosphere

import (
	"math"
	"sync"

	"github.com/bxrne/launchrail/internal/config"
)

// Model defines the interface for atmospheric models that can provide atmospheric data and speed of sound
// Implementers include ISAModel and plugin-wrapped models.
type Model interface {
	GetAtmosphere(altitude float64) AtmosphereData
	GetSpeedOfSound(altitude float64) float64
}

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
// Implements a simplified multi-layer ISA model (up to ~51km for now)
func (isa *ISAModel) GetAtmosphere(altitude float64) AtmosphereData {
	// Round altitude to nearest meter for caching
	roundedAlt := math.Round(altitude)

	isa.mu.RLock()
	if data, exists := isa.cache[roundedAlt]; exists {
		isa.mu.RUnlock()
		return data
	}
	isa.mu.RUnlock()

	// ISA constants from config or use defaults if cfg is nil (should not happen in normal operation)
	T0 := isa.cfg.SeaLevelTemperature    // Sea level temperature (K)
	P0 := isa.cfg.SeaLevelPressure       // Sea level pressure (Pa)
	R := isa.cfg.SpecificGasConstant    // Specific gas constant for dry air (J/kg·K)
	g0 := isa.cfg.GravitationalAccel   // Standard gravitational acceleration (m/s^2)
	gamma := isa.cfg.RatioSpecificHeats // Ratio of specific heats

	// Default lapse rate from config (used for troposphere)
	L_tropo := isa.cfg.TemperatureLapseRate // K/m, typically -0.0065

	var temp, pressure, density, soundSpeed float64

	// Layer boundaries (geopotential altitude h in meters)
	h1 := 11000.0  // Tropopause / Base of lower Stratosphere (Lapse rate = 0)
	h2 := 20000.0  // Top of lower Stratosphere / Base of upper Stratosphere (Lapse rate = +0.001)
	h3 := 32000.0  // Top of upper Stratosphere / Base of Stratopause (Lapse rate = +0.0028)
	h4 := 47000.0  // Top of Stratopause / Base of Mesosphere (Lapse rate = 0)
	// h5 := 51000.0  // Top of this simplified model's Mesosphere layer (Lapse rate = -0.0028)
	// We can extend this further, but this covers typical high-power rocket altitudes well.

	if altitude <= h1 { // Troposphere (0 to 11km)
		temp = T0 + L_tropo*altitude
		if temp <= 0 { // Temperature safety check
			temp = 1.0 // Prevent division by zero / log of zero, ensure minimal positive temp
		}
		pressure = P0 * math.Pow(temp/T0, -g0/(L_tropo*R))
		density = P0 / (R * T0) * math.Pow(temp/T0, -g0/(L_tropo*R)-1.0)
	} else if altitude <= h2 { // Lower Stratosphere (11km to 20km) - Isothermal layer
		T1 := T0 + L_tropo*h1 // Temperature at h1
		P1 := P0 * math.Pow(T1/T0, -g0/(L_tropo*R))
		temp = T1
		pressure = P1 * math.Exp(-g0*(altitude-h1)/(R*T1))
		density = (P1 / (R * T1)) * math.Exp(-g0*(altitude-h1)/(R*T1))
	} else if altitude <= h3 { // Upper Stratosphere (20km to 32km) - Positive lapse rate
		L_strato_upper := 0.001 // K/m
		T1 := T0 + L_tropo*h1
		P1 := P0 * math.Pow(T1/T0, -g0/(L_tropo*R))
		T2 := T1 // Temp at h2 is same as T1 (isothermal layer below)
		P2 := P1 * math.Exp(-g0*(h2-h1)/(R*T1))
		temp = T2 + L_strato_upper*(altitude-h2)
		if temp <= 0 { temp = 1.0 }
		pressure = P2 * math.Pow(temp/T2, -g0/(L_strato_upper*R))
		density = (P2 / (R * T2)) * math.Pow(temp/T2, -g0/(L_strato_upper*R)-1.0)
	} else if altitude <= h4 { // Stratopause (32km to 47km) - Positive lapse rate
		L_strato_upper := 0.001
		L_stratopause := 0.0028 // K/m
		T1 := T0 + L_tropo*h1
		P1 := P0 * math.Pow(T1/T0, -g0/(L_tropo*R))
		T2 := T1
		P2 := P1 * math.Exp(-g0*(h2-h1)/(R*T1))
		T3 := T2 + L_strato_upper*(h3-h2)
		P3 := P2 * math.Pow(T3/T2, -g0/(L_strato_upper*R))
		temp = T3 + L_stratopause*(altitude-h3)
		if temp <= 0 { temp = 1.0 }
		pressure = P3 * math.Pow(temp/T3, -g0/(L_stratopause*R))
		density = (P3 / (R * T3)) * math.Pow(temp/T3, -g0/(L_stratopause*R)-1.0)
	} else { // Altitudes above 47km - For now, model as isothermal at T47 conditions
		// This is a simplification; the Mesosphere has negative lapse rates.
		// For a more complete model, add Mesosphere layers.
		L_tropo_val := L_tropo // Use the tropospheric lapse rate from config
		L_strato_upper_val := 0.001
		L_stratopause_val := 0.0028

		T1_val := T0 + L_tropo_val*h1
		P1_val := P0 * math.Pow(T1_val/T0, -g0/(L_tropo_val*R))
		T2_val := T1_val
		P2_val := P1_val * math.Exp(-g0*(h2-h1)/(R*T2_val))
		T3_val := T2_val + L_strato_upper_val*(h3-h2)
		P3_val := P2_val * math.Pow(T3_val/T2_val, -g0/(L_strato_upper_val*R))
		T4_val := T3_val + L_stratopause_val*(h4-h3)
		P4_val := P3_val * math.Pow(T4_val/T3_val, -g0/(L_stratopause_val*R))

		temp = T4_val // Isothermal at T4 conditions
		pressure = P4_val * math.Exp(-g0*(altitude-h4)/(R*T4_val))
		density = (P4_val / (R * T4_val)) * math.Exp(-g0*(altitude-h4)/(R*T4_val))
		// If density becomes too low or negative, clamp to a minimum physical value
		if density <= 1e-9 { density = 1e-9 }
	}

	// Final safety check for density (should be covered by layer logic, but as a safeguard)
	if density <= 0 || math.IsNaN(density) || math.IsInf(density, 0) {
		density = 1e-9 // Fallback to a minimal positive density
	}
	if temp <= 0 || math.IsNaN(temp) || math.IsInf(temp, 0) {
		temp = 1.0 // Fallback to minimal positive temperature (1K)
	}
	if pressure <= 0 || math.IsNaN(pressure) || math.IsInf(pressure, 0) {
		pressure = 1e-5 // Fallback to minimal positive pressure
	}

	if temp > 0 && R > 0 { // Ensure temp and R are positive for sound speed calc
		soundSpeed = math.Sqrt(gamma * R * temp)
	} else {
		soundSpeed = 1.0 // Fallback sound speed if temp or R is not valid
	}

	data := AtmosphereData{
		Density:     density,
		Temperature: temp,
		Pressure:    pressure,
		SoundSpeed:  soundSpeed,
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
