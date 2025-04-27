package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHiprEuroc24Benchmark_FindApogee(t *testing.T) {
	tests := []struct {
		name          string
		flightInfo    []FlightInfo
		expectedHeight float64
		expectedTime   float64
	}{
		{
			name: "Normal Case",
			flightInfo: []FlightInfo{
				{Timestamp: 0.0, Height: 10.0},
				{Timestamp: 1.0, Height: 100.0},
				{Timestamp: 2.0, Height: 150.0},
				{Timestamp: 3.0, Height: 120.0},
			},
			expectedHeight: 150.0,
			expectedTime:   2.0,
		},
		{
			name: "Empty Data",
			flightInfo: []FlightInfo{},
			expectedHeight: 0.0,
			expectedTime:   0.0,
		},
		{
			name: "Single Point",
			flightInfo: []FlightInfo{
				{Timestamp: 5.0, Height: 50.0},
			},
			expectedHeight: 50.0,
			expectedTime:   5.0,
		},
		{
			name: "Apogee at End",
			flightInfo: []FlightInfo{
				{Timestamp: 0.0, Height: 10.0},
				{Timestamp: 1.0, Height: 100.0},
				{Timestamp: 2.0, Height: 150.0},
			},
			expectedHeight: 150.0,
			expectedTime:   2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := HiprEuroc24Benchmark{flightInfo: tt.flightInfo}
			h, ts := b.findApogee()
			assert.Equal(t, tt.expectedHeight, h, "Incorrect apogee height")
			assert.Equal(t, tt.expectedTime, ts, "Incorrect apogee timestamp")
		})
	}
}

func TestHiprEuroc24Benchmark_FindMaxVelocity(t *testing.T) {
	// Similar structure to TestHiprEuroc24Benchmark_FindApogee
	tests := []struct {
		name            string
		flightInfo      []FlightInfo
		expectedVelocity float64
		expectedTime     float64
	}{
		{
			name: "Normal Case",
			flightInfo: []FlightInfo{
				{Timestamp: 0.0, Velocity: 10.0},
				{Timestamp: 1.0, Velocity: 100.0},
				{Timestamp: 2.0, Velocity: 150.0},
				{Timestamp: 3.0, Velocity: 120.0},
			},
			expectedVelocity: 150.0,
			expectedTime:     2.0,
		},
		{
			name:           "Empty Data",
			flightInfo:     []FlightInfo{},
			expectedVelocity: 0.0,
			expectedTime:     0.0,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := HiprEuroc24Benchmark{flightInfo: tt.flightInfo}
			v, ts := b.findMaxVelocity()
			assert.Equal(t, tt.expectedVelocity, v, "Incorrect max velocity")
			assert.Equal(t, tt.expectedTime, ts, "Incorrect max velocity timestamp")
		})
	}
}

func TestCompareFloat(t *testing.T) {
	tests := []struct {
		name            string
		metricName      string
		description     string
		expected        float64
		actual          float64
		tolerancePercent float64
		expectedPass    bool
	}{
		{"Pass Within Tolerance", "Altitude", "Test altitude", 100.0, 102.0, 0.05, true},
		{"Pass Exact Match", "Velocity", "Test velocity", 50.0, 50.0, 0.10, true},
		{"Pass Edge of Tolerance (Upper)", "Time", "Test time", 10.0, 10.5, 0.05, true},
		{"Pass Edge of Tolerance (Lower)", "Pressure", "Test pressure", 1000.0, 970.0, 0.03, true},
		{"Fail Outside Tolerance (Upper)", "Altitude", "Test altitude", 100.0, 106.0, 0.05, false},
		{"Fail Outside Tolerance (Lower)", "Velocity", "Test velocity", 50.0, 44.0, 0.10, false},
		{"Zero Expected, Non-Zero Actual, Pass", "ErrorCount", "Test error count", 0.0, 0.01, 0.1, true}, // Tolerance calc needs care
		{"Zero Expected, Non-Zero Actual, Fail", "ErrorCount", "Test error count", 0.0, 1.0, 0.1, false}, // Needs absolute tolerance or special handling
		{"Negative Values, Pass", "Temperature", "Test temperature", -10.0, -10.2, 0.05, true},
		{"Negative Values, Fail", "Temperature", "Test temperature", -10.0, -11.0, 0.05, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareFloat(tt.metricName, tt.description, tt.expected, tt.actual, tt.tolerancePercent)
			assert.Equal(t, tt.expectedPass, result.Passed, "Pass/Fail status mismatch")
			assert.Equal(t, tt.metricName, result.Metric)
			assert.Equal(t, tt.expected, result.Expected)
			assert.Equal(t, tt.actual, result.Actual)
		})
	}
}

// TODO: Add tests for findEventTime and findStateTime if their logic becomes complex
