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
// atmosphericLayer represents a layer of the atmosphere with its properties
type atmosphericLayer struct {
    baseAltitude float64
    lapseRate    float64
    temperature  float64
    pressure     float64
}

// calculateLayerConditions calculates temperature and pressure at a given altitude within a layer
func (isa *ISAModel) calculateLayerConditions(alt float64, layer atmosphericLayer, prevLayer *atmosphericLayer) (temp, pressure float64) {
    g0 := isa.cfg.GravitationalAccel
    R := isa.cfg.SpecificGasConstant

    if layer.lapseRate != 0 {
        // Layer with temperature gradient
        temp = layer.temperature + layer.lapseRate*(alt-layer.baseAltitude)
        if temp <= 0 {
            temp = 1.0
        }
        pressure = layer.pressure * math.Pow(temp/layer.temperature, -g0/(layer.lapseRate*R))
    } else if prevLayer != nil && alt < layer.baseAltitude {
        // Calculate conditions using previous layer's properties
        return isa.calculateLayerConditions(alt, *prevLayer, nil)
    } else {
        // Isothermal layer
        temp = layer.temperature
        pressure = layer.pressure * math.Exp(-g0*(alt-layer.baseAltitude)/(R*temp))
    }
    return temp, pressure
}

// calculateDensity calculates air density from temperature and pressure
func (isa *ISAModel) calculateDensity(temp, pressure float64) float64 {
    R := isa.cfg.SpecificGasConstant
    density := pressure / (R * temp)
    if density <= 1e-9 {
        density = 1e-9
    }
    return density
}

// calculateSoundSpeed calculates speed of sound from temperature
func (isa *ISAModel) calculateSoundSpeed(temp float64) float64 {
    if temp <= 0 || isa.cfg.SpecificGasConstant <= 0 {
        return 1.0
    }
    return math.Sqrt(isa.cfg.RatioSpecificHeats * isa.cfg.SpecificGasConstant * temp)
}

// initializeAtmosphericLayers sets up the atmospheric layers with their properties
func (isa *ISAModel) initializeAtmosphericLayers(T0, P0 float64) []atmosphericLayer {
    g0 := isa.cfg.GravitationalAccel
    R := isa.cfg.SpecificGasConstant
    lapseRateTrop := isa.cfg.TemperatureLapseRate

    // Calculate temperatures at layer boundaries
    T1 := T0 + lapseRateTrop*11000  // Temperature at 11km
    T2 := T1                        // Temperature at 20km (isothermal layer)
    T3 := T2 + 0.001*(32000-20000)  // Temperature at 32km
    T4 := T3 + 0.0028*(47000-32000) // Temperature at 47km

    layers := []atmosphericLayer{
        {baseAltitude: 0, lapseRate: lapseRateTrop, temperature: T0, pressure: P0},  // Troposphere
        {baseAltitude: 11000, lapseRate: 0, temperature: T1},                         // Lower Stratosphere
        {baseAltitude: 20000, lapseRate: 0.001, temperature: T2},                    // Upper Stratosphere
        {baseAltitude: 32000, lapseRate: 0.0028, temperature: T3},                   // Stratopause
        {baseAltitude: 47000, lapseRate: 0, temperature: T4},                        // Mesosphere
    }

    // Calculate pressures for each layer
    for i := 1; i < len(layers); i++ {
        prevLayer := layers[i-1]
        currentLayer := &layers[i]
        
        if prevLayer.lapseRate != 0 {
            currentLayer.pressure = prevLayer.pressure * math.Pow(
                currentLayer.temperature/prevLayer.temperature,
                -g0/(prevLayer.lapseRate*R))
        } else {
            currentLayer.pressure = prevLayer.pressure * math.Exp(
                -g0*(currentLayer.baseAltitude-prevLayer.baseAltitude)/(R*prevLayer.temperature))
        }
    }

    return layers
}

func (isa *ISAModel) GetAtmosphere(altitude float64) AtmosphereData {
    // Check cache first
    roundedAlt := math.Round(altitude)
    isa.mu.RLock()
    if data, exists := isa.cache[roundedAlt]; exists {
        isa.mu.RUnlock()
        return data
    }
    isa.mu.RUnlock()

    // Initialize atmospheric layers
    layers := isa.initializeAtmosphericLayers(
        isa.cfg.SeaLevelTemperature,
        isa.cfg.SeaLevelPressure)

    // Find appropriate layer and calculate conditions
    var temp, pressure float64
    for i, layer := range layers {
        if altitude <= layer.baseAltitude || i == len(layers)-1 {
            if i > 0 {
                temp, pressure = isa.calculateLayerConditions(altitude, layer, &layers[i-1])
            } else {
                temp, pressure = isa.calculateLayerConditions(altitude, layer, nil)
            }
            break
        }
    }

    // Ensure valid values
    if temp <= 0 || math.IsNaN(temp) || math.IsInf(temp, 0) {
        temp = 1.0
    }
    if pressure <= 0 || math.IsNaN(pressure) || math.IsInf(pressure, 0) {
        pressure = 1e-5
    }

    // Calculate derived properties
    density := isa.calculateDensity(temp, pressure)
    soundSpeed := isa.calculateSoundSpeed(temp)

    // Create and cache result
    data := AtmosphereData{
        Density:     density,
        Temperature: temp,
        Pressure:    pressure,
        SoundSpeed:  soundSpeed,
    }

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
