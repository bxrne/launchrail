package main

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareFloat(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string // Name arg for compareFloat
		description string // Description arg for compareFloat
		expected    float64
		actual      float64
		tolerance   float64
		wantPassed  bool
	}{
		{"Within Positive Tolerance", "Metric1", "Desc1", 100.0, 101.0, 0.02, true},
		{"Outside Positive Tolerance", "Metric2", "Desc2", 100.0, 103.0, 0.02, false},
		{"Within Negative Tolerance", "Metric3", "Desc3", 100.0, 99.0, 0.02, true},
		{"Outside Negative Tolerance", "Metric4", "Desc4", 100.0, 97.0, 0.02, false},
		{"Exact Match", "Metric5", "Desc5", 50.0, 50.0, 0.01, true},
		{"Expected Zero, Actual Non-Zero within Abs Tolerance", "Metric6", "Desc6", 0.0, 0.005, 0.01, true},
		{"Expected Zero, Actual Non-Zero outside Abs Tolerance", "Metric7", "Desc7", 0.0, 0.015, 0.01, false},
		{"Actual Zero, Expected Non-Zero", "Metric8", "Desc8", 10.0, 0.0, 0.01, false},
		{"Both Zero", "Metric9", "Desc9", 0.0, 0.0, 0.01, true},
		{"Negative Values Within Tolerance", "Metric10", "Desc10", -100.0, -101.0, 0.02, true},
		{"Negative Values Outside Tolerance", "Metric11", "Desc11", -100.0, -103.0, 0.02, false},
		{"Mixed Signs", "Metric12", "Desc12", -10.0, 10.0, 0.1, false},
		{"Large Numbers Within Tolerance", "Metric16", "Desc16", 1e9, 1.01e9, 0.02, true},
		{"NaN Expected", "Metric13", "Desc13", math.NaN(), 10.0, 0.1, false},
		{"NaN Actual", "Metric14", "Desc14", 10.0, math.NaN(), 0.1, false},
		{"NaN Both", "Metric15", "Desc15", math.NaN(), math.NaN(), 0.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareFloat(tt.metricName, tt.description, tt.expected, tt.actual, tt.tolerance)
			assert.Equal(t, tt.wantPassed, result.Passed)
			if tt.expected == 0 {
				assert.Equal(t, "absolute", result.ToleranceType, "ToleranceType should be absolute when expected is 0")
			} else {
				assert.Equal(t, "relative", result.ToleranceType, "ToleranceType should be relative when expected is non-zero")
			}
		})
	}
}

func TestFindGroundTruthApogee(t *testing.T) {
	tests := []struct {
		name         string
		data         []FlightInfo
		wantAltitude float64
		wantTime     float64
	}{
		{"Empty Data", []FlightInfo{}, 0, 0},
		{"Single Point", []FlightInfo{{Timestamp: 1.0, Height: 100.0}}, 100.0, 1.0},
		{"Multiple Points Ascending", []FlightInfo{{Timestamp: 1.0, Height: 100.0}, {Timestamp: 2.0, Height: 200.0}}, 200.0, 2.0},
		{"Multiple Points Descending", []FlightInfo{{Timestamp: 1.0, Height: 200.0}, {Timestamp: 2.0, Height: 100.0}}, 200.0, 1.0},
		{"Peak in Middle", []FlightInfo{{Timestamp: 1.0, Height: 100.0}, {Timestamp: 2.0, Height: 300.0}, {Timestamp: 3.0, Height: 200.0}}, 300.0, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alt, time := findGroundTruthApogee(tt.data)
			assert.Equal(t, tt.wantAltitude, alt)
			assert.Equal(t, tt.wantTime, time)
		})
	}
}

func TestFindGroundTruthMaxVelocity(t *testing.T) {
	tests := []struct {
		name         string
		data         []FlightInfo
		wantVelocity float64
		wantTime     float64
	}{
		{"Empty Data", []FlightInfo{}, 0, 0},
		{"Single Point", []FlightInfo{{Timestamp: 1.0, Velocity: 50.0}}, 50.0, 1.0},
		{"Multiple Points Increasing", []FlightInfo{{Timestamp: 1.0, Velocity: 50.0}, {Timestamp: 2.0, Velocity: 100.0}}, 100.0, 2.0},
		{"Multiple Points Decreasing", []FlightInfo{{Timestamp: 1.0, Velocity: 100.0}, {Timestamp: 2.0, Velocity: 50.0}}, 100.0, 1.0},
		{"Peak in Middle", []FlightInfo{{Timestamp: 1.0, Velocity: 50.0}, {Timestamp: 2.0, Velocity: 150.0}, {Timestamp: 3.0, Velocity: 100.0}}, 150.0, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vel, time := findGroundTruthMaxVelocity(tt.data)
			assert.Equal(t, tt.wantVelocity, vel)
			assert.Equal(t, tt.wantTime, time)
		})
	}
}

func TestFindGroundTruthEventTime(t *testing.T) {
	tests := []struct {
		name        string
		events      []EventInfo
		targetEvent string
		wantTime    float64
	}{
		{"Empty Events", []EventInfo{}, "APOGEE", -1},
		{"Event Found", []EventInfo{{Timestamp: 5.0, Event: "LAUNCH"}, {Timestamp: 15.2, Event: "APOGEE"}}, "APOGEE", 15.2},
		{"Event Found Case Insensitive", []EventInfo{{Timestamp: 5.0, Event: "LAUNCH"}, {Timestamp: 15.2, Event: "Apogee"}}, "apogee", 15.2},
		{"Event Not Found", []EventInfo{{Timestamp: 5.0, Event: "LAUNCH"}, {Timestamp: 30.0, Event: "LANDED"}}, "APOGEE", -1},
		{"First Event Match", []EventInfo{{Timestamp: 15.2, Event: "APOGEE"}, {Timestamp: 16.0, Event: "APOGEE_DUPE"}}, "APOGEE", 15.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time := findGroundTruthEventTime(tt.events, tt.targetEvent)
			assert.Equal(t, tt.wantTime, time)
		})
	}
}
