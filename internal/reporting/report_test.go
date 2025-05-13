package reporting_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/reporting"
	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
)

func TestCalculateMotionMetrics_Basic(t *testing.T) {
	logger := logf.New(logf.Opts{})

	motionHeaders := []string{"Time (s)", "Altitude AGL (m)", "Total Velocity (m/s)", "Total Acceleration (m/s^2)"} // Updated Acceleration header
	motionData := []*reporting.PlotSimRecord{
		// plotSimRecord is map[string]interface{}
		&reporting.PlotSimRecord{
			"Time (s)":                   0.0,
			"Altitude AGL (m)":           0.0,
			"Total Velocity (m/s)":       0.0,
			"Total Acceleration (m/s^2)": 1.0, // Standing on pad
		},
		&reporting.PlotSimRecord{
			"Time (s)":                   1.0,
			"Altitude AGL (m)":           50.0,
			"Total Velocity (m/s)":       30.0,
			"Total Acceleration (m/s^2)": 5.0,
		},
		&reporting.PlotSimRecord{
			"Time (s)":                   2.0, // Rail exit assumed here for simplicity
			"Altitude AGL (m)":           150.0,
			"Total Velocity (m/s)":       60.0,
			"Total Acceleration (m/s^2)": 10.0,
		},
		&reporting.PlotSimRecord{
			"Time (s)":                   10.0, // Apogee
			"Altitude AGL (m)":           1000.0,
			"Total Velocity (m/s)":       0.0,
			"Total Acceleration (m/s^2)": -1.0, // Freefall
		},
		&reporting.PlotSimRecord{
			"Time (s)":                   20.0, // Touchdown
			"Altitude AGL (m)":           0.0,
			"Total Velocity (m/s)":       -5.0, // Landing speed
			"Total Acceleration (m/s^2)": 2.0,  // Impact
		},
	}

	eventsData := [][]string{
		{"Event Name", "Time (s)"}, // Header row for findEventHeaderIndices
		{"Launch", "0.0"},
		{"Rail Exit", "2.0"},
		{"Motor Burnout", "5.0"},
		{"Apogee", "10.0"},
		{"Touchdown", "20.0"},
	}

	launchRailLength := 100.0 // meters

	metrics := reporting.CalculateMotionMetrics(motionData, motionHeaders, eventsData, launchRailLength, &logger) // Pass as pointer

	assert.NotNil(t, metrics, "Metrics should not be nil")
	assert.Empty(t, metrics.Error, "Metrics error should be empty for basic case")

	assert.InDelta(t, 1000.0, metrics.MaxAltitudeAGL, 0.01, "Max Altitude AGL")
	assert.InDelta(t, 60.0, metrics.MaxSpeed, 0.01, "Max Speed")
	// Max acceleration in 'g's needs to be multiplied by 9.81 if the input is in g and output in m/s^2
	// The struct comment says m/s^2. The header is now m/s^2.
	// Input data for m/s^2: 1.0, 5.0, 10.0, -1.0, 2.0. Max absolute is 10.0.
	assert.InDelta(t, 10.0, metrics.MaxAcceleration, 0.01, "Max Acceleration") // findPeakValues uses Abs
	assert.InDelta(t, 10.0, metrics.TimeAtApogee, 0.01, "Time at Apogee (sensor)")
	assert.InDelta(t, 20.0, metrics.FlightTime, 0.01, "Flight Time (event)")
	assert.InDelta(t, 60.0, metrics.RailExitVelocity, 0.01, "Rail Exit Velocity") // Based on event and finding closest point
	assert.InDelta(t, 10.0, metrics.TimeToApogee, 0.01, "Time to Apogee (event-based, launch@0, apogee_event@10)")
	assert.InDelta(t, 5.0, metrics.BurnoutTime, 0.01, "Burnout Time")
	// Burnout altitude will be interpolated or the closest point to 5.0s.

	// Re-defining motionData with a point closer to burnout for a more predictable test.
	motionDataWithBurnout := []*reporting.PlotSimRecord{
		&reporting.PlotSimRecord{"Time (s)": 0.0, "Altitude AGL (m)": 0.0, "Total Velocity (m/s)": 0.0, "Total Acceleration (m/s^2)": 1.0},
		&reporting.PlotSimRecord{"Time (s)": 1.0, "Altitude AGL (m)": 50.0, "Total Velocity (m/s)": 30.0, "Total Acceleration (m/s^2)": 5.0},
		&reporting.PlotSimRecord{"Time (s)": 2.0, "Altitude AGL (m)": 150.0, "Total Velocity (m/s)": 60.0, "Total Acceleration (m/s^2)": 10.0},
		&reporting.PlotSimRecord{"Time (s)": 5.0, "Altitude AGL (m)": 500.0, "Total Velocity (m/s)": 80.0, "Total Acceleration (m/s^2)": 2.0}, // Burnout point
		&reporting.PlotSimRecord{"Time (s)": 10.0, "Altitude AGL (m)": 1000.0, "Total Velocity (m/s)": 0.0, "Total Acceleration (m/s^2)": -1.0},
		&reporting.PlotSimRecord{"Time (s)": 20.0, "Altitude AGL (m)": 0.0, "Total Velocity (m/s)": -5.0, "Total Acceleration (m/s^2)": 2.0},
	}

	metricsWithBurnoutPoint := reporting.CalculateMotionMetrics(motionDataWithBurnout, motionHeaders, eventsData, launchRailLength, &logger) // Pass as pointer
	assert.InDelta(t, 500.0, metricsWithBurnoutPoint.BurnoutAltitude, 0.01, "Burnout Altitude")

}
