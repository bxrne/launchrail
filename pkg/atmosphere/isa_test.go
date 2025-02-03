package atmosphere_test

import (
	"math"
	"sync"
	"testing"

	"github.com/bxrne/launchrail/pkg/atmosphere"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewISAModel(t *testing.T) {
	isa := atmosphere.NewISAModel()
	require.NotNil(t, isa)
}

func TestGetAtmosphere(t *testing.T) {
	isa := atmosphere.NewISAModel()

	altitude := 1000.0 // 1000 meters
	data := isa.GetAtmosphere(altitude)

	// Calculate expected values
	expectedTemp := atmosphere.T_0 + atmosphere.LAPSE_RATE*altitude
	expectedPressure := atmosphere.P_0 * math.Pow(expectedTemp/atmosphere.T_0, -atmosphere.G_0/(atmosphere.LAPSE_RATE*atmosphere.R_AIR))
	expectedDensity := expectedPressure / (atmosphere.R_AIR * expectedTemp)

	assert.InEpsilon(t, expectedTemp, data.Temperature, 1e-6, "Temperature mismatch")
	assert.InEpsilon(t, expectedPressure, data.Pressure, 1e-6, "Pressure mismatch")
	assert.InEpsilon(t, expectedDensity, data.Density, 1e-6, "Density mismatch")
}

func TestCachingMechanism(t *testing.T) {
	isa := atmosphere.NewISAModel()
	altitude := 500.0

	firstCall := isa.GetAtmosphere(altitude)
	secondCall := isa.GetAtmosphere(altitude)

	assert.Equal(t, firstCall, secondCall, "Cached result should match")
}

func TestConcurrencySafety(t *testing.T) {
	isa := atmosphere.NewISAModel()
	altitude := 2000.0
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			isa.GetAtmosphere(altitude)
		}()
	}

	wg.Wait()
}

func TestGetSpeedOfSound(t *testing.T) {
	isa := atmosphere.NewISAModel()
	altitude := 5000.0
	data := isa.GetAtmosphere(altitude)
	expectedSpeed := math.Sqrt(atmosphere.GAMMA * atmosphere.R_AIR * data.Temperature)
	actualSpeed := isa.GetSpeedOfSound(altitude)

	assert.InEpsilon(t, expectedSpeed, actualSpeed, 1e-6, "Speed of sound mismatch")
}
