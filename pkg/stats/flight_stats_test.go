package stats_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/stats"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN nothing WHEN NewFlightStats is called THEN a new FlightStats is returned
func TestNewFlightStats(t *testing.T) {
	fs := stats.NewFlightStats()
	assert.NotNil(t, fs)
}

// TEST: GIVEN a FlightStats WHEN Update is called THEN the FlightStats is updated
func TestFlightStatsUpdate(t *testing.T) {
	fs := stats.NewFlightStats()
	fs.Update(1.0, 100.0, 10.0, 1.0, 0.1)
	assert.Equal(t, 100.0, fs.Apogee)
	assert.Equal(t, 10.0, fs.MaxVelocity)
	assert.Equal(t, 1.0, fs.MaxAccel)
	assert.Equal(t, 1.0, fs.TotalFlightTime)
	assert.Equal(t, 0.1, fs.MaxMach)
}

// TEST: GIVEN a FlightStats WHEN String is called THEN a string representation is returned
func TestFlightStatsString(t *testing.T) {
	fs := stats.NewFlightStats()
	fs.Update(1.0, 100.0, 10.0, 1.0, 0.1)
	assert.NotEmpty(t, fs.String())
	expected := "Apogee=100.00m, MaxVelocity=10.00m/s, MaxAccel=1.00m/s², MaxMach=0.10, GroundHitVelocity=0.00m/s"
	assert.Equal(t, expected, fs.String())
}
