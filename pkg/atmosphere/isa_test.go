package atmosphere_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/atmosphere"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestConfig returns an ISA configuration for testing
func getTestConfig() *config.ISAConfiguration {
	return &config.ISAConfiguration{
		SpecificGasConstant:  287.05287, // J/(kg·K)
		GravitationalAccel:   9.80665,   // m/s²
		SeaLevelDensity:      1.225,     // kg/m³
		SeaLevelTemperature:  288.15,    // K
		SeaLevelPressure:     101325,    // Pa
		RatioSpecificHeats:   1.4,       // dimensionless
		TemperatureLapseRate: -0.0065,   // K/m
	}
}

// TEST: GIVEN a new ISAModel WHEN NewISAModel is called THEN the model is initialized correctly
func TestNewISAModel(t *testing.T) {
	cfg := getTestConfig()
	isa := atmosphere.NewISAModel(cfg)
	require.NotNil(t, isa)
}

// TEST: GIVEN an ISAModel WHEN GetTemperature is called THEN correct temperatures are returned
func TestISAModel_GetTemperature(t *testing.T) {
	isa := atmosphere.NewISAModel(getTestConfig())

	tests := []struct {
		name     string
		altitude float64
		want     float64
	}{
		{"Sea Level", 0, 288.15},
		{"1000m", 1000, 281.65},
		{"11000m (Tropopause)", 11000, 216.65},
		{"Negative Altitude", -100, 288.80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isa.GetTemperature(tt.altitude)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

// TEST: GIVEN an ISAModel WHEN GetAtmosphere is called THEN correct atmospheric data is returned
func TestISAModel_GetAtmosphere(t *testing.T) {
	isa := atmosphere.NewISAModel(getTestConfig())

	tests := []struct {
		name         string
		altitude     float64
		wantDensity  float64
		wantTemp     float64
		wantPressure float64
	}{
		{"Sea Level", 0, 1.225, 288.15, 101325},
		{"1000m", 1000, 1.112, 281.65, 89876},
		{"2000m", 2000, 1.007, 275.15, 79501},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isa.GetAtmosphere(tt.altitude)
			assert.InDelta(t, tt.wantDensity, got.Density, 0.01)
			assert.InDelta(t, tt.wantTemp, got.Temperature, 0.01)
			assert.InDelta(t, tt.wantPressure, got.Pressure, 100)
		})
	}
}

// TEST: GIVEN an ISAModel WHEN GetSpeedOfSound is called THEN correct speed of sound is returned
func TestISAModel_GetSpeedOfSound(t *testing.T) {
	isa := atmosphere.NewISAModel(getTestConfig())

	tests := []struct {
		name     string
		altitude float64
		want     float64
	}{
		{"Sea Level", 0, 340.29},
		{"1000m", 1000, 336.43},
		{"11000m", 11000, 295.07},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isa.GetSpeedOfSound(tt.altitude)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

// TEST: GIVEN an ISAModel WHEN GetAtmosphere is called multiple times THEN results are cached
func TestISAModel_CachingBehavior(t *testing.T) {
	isa := atmosphere.NewISAModel(getTestConfig())

	// Get data for same altitude multiple times
	altitude := 1000.0
	first := isa.GetAtmosphere(altitude)
	second := isa.GetAtmosphere(altitude)

	// Results should be identical since second call should use cached data
	assert.Equal(t, first, second)

	// Slightly different altitude should calculate new values
	different := isa.GetAtmosphere(altitude + 1.5)
	assert.NotEqual(t, first, different)
}

// TEST: GIVEN an ISAModel WHEN concurrent access occurs THEN thread safety is maintained
func TestISAModel_ConcurrentAccess(t *testing.T) {
	isa := atmosphere.NewISAModel(getTestConfig())

	// Launch multiple goroutines to access the model concurrently
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(alt float64) {
			_ = isa.GetAtmosphere(alt)
			done <- true
		}(float64(i * 100))
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}
}
