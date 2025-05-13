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

func TestCalculateMotionMetrics_NilEventsData(t *testing.T) {
	logger := logf.New(logf.Opts{})

	motionHeaders := []string{"Time (s)", "Altitude AGL (m)", "Total Velocity (m/s)", "Total Acceleration (m/s^2)"}
	motionData := []*reporting.PlotSimRecord{
		{
			"Time (s)":                   0.0,
			"Altitude AGL (m)":           0.0,
			"Total Velocity (m/s)":       0.0,
			"Total Acceleration (m/s^2)": 1.0,
		},
		{
			"Time (s)":                   10.0, // Apogee from motion
			"Altitude AGL (m)":           1000.0,
			"Total Velocity (m/s)":       0.0,
			"Total Acceleration (m/s^2)": -1.0,
		},
		{
			"Time (s)":                   20.0, // Touchdown from motion (for landing speed calculation if possible)
			"Altitude AGL (m)":           0.0,
			"Total Velocity (m/s)":       -5.0,
			"Total Acceleration (m/s^2)": 2.0,
		},
	}

	var eventsData [][]string = nil // Explicitly nil
	launchRailLength := 100.0

	metrics := reporting.CalculateMotionMetrics(motionData, motionHeaders, eventsData, launchRailLength, &logger)

	assert.NotNil(t, metrics, "Metrics should not be nil")
	// Most event-based metrics should be zero or unset
	assert.Equal(t, 0.0, metrics.FlightTime, "FlightTime should be 0 when eventsData is nil")
	// TimeToApogee and TimeAtApogee might be calculated from motion data
	assert.InDelta(t, 1000.0, metrics.MaxAltitudeAGL, 0.01, "MaxAltitudeAGL from motion data")
	assert.InDelta(t, 10.0, metrics.TimeAtApogee, 0.01, "TimeAtApogee from motion data (approx launch at 0s)")
	// TimeToApogee from motion: apogeeTimeFromMotion (10.0) - launchTimeFromMotion (motionPoints[0].Time = 0.0) = 10.0
	assert.InDelta(t, 10.0, metrics.TimeToApogee, 0.01, "TimeToApogee from motion data")

	assert.Equal(t, 0.0, metrics.BurnoutTime, "BurnoutTime should be 0")
	assert.Equal(t, 0.0, metrics.BurnoutAltitude, "BurnoutAltitude should be 0")
	assert.Equal(t, 0.0, metrics.CoastToApogeeTime, "CoastToApogeeTime should be 0")
	assert.Equal(t, 0.0, metrics.DescentTime, "DescentTime should be 0 as it relies on event-based flight/apogee time")
	assert.InDelta(t, -5.0, metrics.LandingSpeed, 0.01, "LandingSpeed from motion data at last point if Alt=0")

	// RailExitVelocity might be calculated if launchRailLength > 0 and motion data allows
	// With motion point {Time:10, Alt:1000}, if railLength is 100, it should pick that up before apogee.
	// Need a point that crosses rail length. Let's refine motionData slightly for this test.
	motionDataForRailExit := []*reporting.PlotSimRecord{
		{"Time (s)": 0.0, "Altitude AGL (m)": 0.0, "Total Velocity (m/s)": 0.0},
		{"Time (s)": 1.0, "Altitude AGL (m)": 60.0, "Total Velocity (m/s)": 30.0},  // Below rail exit
		{"Time (s)": 2.0, "Altitude AGL (m)": 120.0, "Total Velocity (m/s)": 50.0}, // Above rail exit (100m)
		{"Time (s)": 10.0, "Altitude AGL (m)": 1000.0, "Total Velocity (m/s)": 0.0},
		{"Time (s)": 20.0, "Altitude AGL (m)": 0.0, "Total Velocity (m/s)": -5.0},
	}
	metricsRailExit := reporting.CalculateMotionMetrics(motionDataForRailExit, motionHeaders, eventsData, launchRailLength, &logger)
	assert.InDelta(t, 50.0, metricsRailExit.RailExitVelocity, 0.01, "RailExitVelocity from motion data and launchRailLength")
}

func TestCalculateMotionMetrics_NoMotionPoints(t *testing.T) {
	logger := logf.New(logf.Opts{})

	// Pass nil motionData and nil motionHeaders
	// This ensures ExtractMotionPoints returns an empty slice because timeIdx will be -1.
	var motionData []*reporting.PlotSimRecord = nil
	var motionHeaders []string = nil

	eventsData := [][]string{
		{"Event Name", "Time (s)"},
		{"Launch", "0.0"},
		{"Motor Burnout", "5.0"},
		{"Apogee", "10.0"},
		{"Touchdown", "20.0"},
		{"Rail Exit", "2.0"},
	}
	launchRailLength := 100.0

	metrics := reporting.CalculateMotionMetrics(motionData, motionHeaders, eventsData, launchRailLength, &logger)

	assert.NotNil(t, metrics, "Metrics should not be nil")

	// Event-based metrics should be calculated
	assert.InDelta(t, 20.0, metrics.FlightTime, 0.01, "FlightTime from events")
	assert.InDelta(t, 10.0, metrics.TimeToApogee, 0.01, "TimeToApogee from events") // Should be from event, not motion fallback
	assert.InDelta(t, 10.0, metrics.TimeAtApogee, 0.01, "TimeAtApogee from events")
	assert.InDelta(t, 5.0, metrics.BurnoutTime, 0.01, "BurnoutTime from events")
	assert.InDelta(t, 5.0, metrics.CoastToApogeeTime, 0.01, "CoastToApogeeTime from events") // 10 (apogee) - 5 (burnout)
	assert.InDelta(t, 10.0, metrics.DescentTime, 0.01, "DescentTime from events")            // 20 (flight) - 10 (apogee)

	// Motion-point dependent metrics should be zero or default
	assert.Equal(t, 0.0, metrics.MaxAltitudeAGL, "MaxAltitudeAGL should be 0 as no motion points")
	assert.Equal(t, 0.0, metrics.MaxSpeed, "MaxSpeed should be 0")
	assert.Equal(t, 0.0, metrics.MaxAcceleration, "MaxAcceleration should be 0")
	assert.Equal(t, 0.0, metrics.RailExitVelocity, "RailExitVelocity should be 0")
	assert.Equal(t, 0.0, metrics.BurnoutAltitude, "BurnoutAltitude should be 0")
	assert.Equal(t, 0.0, metrics.LandingSpeed, "LandingSpeed should be 0")

	// Check for specific log warnings if possible (optional, depends on log capture setup)
	// e.g., "No motion points extracted..."
	// e.g., "Motion points are not available. Point-based metrics... cannot be calculated."
}

// TestCalculateMotorSummary_Basic tests the basic functionality of calculateMotorSummary.
// It checks if the function correctly calculates max thrust, average thrust, total impulse, and burn time.
