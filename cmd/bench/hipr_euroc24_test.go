package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary directory and files for benchmark data
// Updated to create benchmark-specific subdirectory and use processed filenames.
func setupTempBenchdata(t *testing.T, benchmarkName, flightInfoContent, eventInfoContent, flightStatesContent string) string {
	t.Helper()
	tempDir := t.TempDir() // Base temporary directory

	// Create benchmark-specific subdirectory
	benchmarkSubDir := filepath.Join(tempDir, benchmarkName)
	err := os.MkdirAll(benchmarkSubDir, 0755)
	require.NoError(t, err, "Failed to create benchmark subdirectory")

	// --- Use filenames expected by LoadData ---
	flightInfoFilename := "fl001 - flight_info_processed.csv"
	eventInfoFilename := "fl001 - event_info_processed.csv"
	flightStatesFilename := "fl001 - flight_states_processed.csv"


	// Write flight_info.csv
	flightInfoPath := filepath.Join(benchmarkSubDir, flightInfoFilename) // Use correct filename
	err = os.WriteFile(flightInfoPath, []byte(flightInfoContent), 0644)
	require.NoError(t, err, "Failed to write temp "+flightInfoFilename)

	// Write event_info.csv
	eventInfoPath := filepath.Join(benchmarkSubDir, eventInfoFilename) // Use correct filename
	err = os.WriteFile(eventInfoPath, []byte(eventInfoContent), 0644)
	require.NoError(t, err, "Failed to write temp "+eventInfoFilename)

	// Write flight_states.csv
	flightStatesPath := filepath.Join(benchmarkSubDir, flightStatesFilename) // Use correct filename
	err = os.WriteFile(flightStatesPath, []byte(flightStatesContent), 0644)
	require.NoError(t, err, "Failed to write temp "+flightStatesFilename)

	return tempDir // Return the base temp directory
}

func TestNewHiprEuroc24Benchmark(t *testing.T) {
	cfg := BenchmarkConfig{BenchdataPath: "/fake/path"}
	b := NewHiprEuroc24Benchmark(cfg)
	require.NotNil(t, b, "Constructor should return a non-nil object")
	assert.IsType(t, &HiprEuroc24Benchmark{}, b, "Constructor should return a HiprEuroc24Benchmark pointer")
}

func TestHiprEuroc24Benchmark_Name(t *testing.T) {
	cfg := BenchmarkConfig{} // Config content doesn't matter for Name()
	b := NewHiprEuroc24Benchmark(cfg)
	assert.Equal(t, "hipr-euroc24", b.Name(), "Benchmark name mismatch")
}

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

// Test for the main Run method - Requires setup with temporary data
func TestHiprEuroc24Benchmark_Run(t *testing.T) {
	// --- Test Data ---
	// Simple data covering basic cases
	flightInfoCSV := `Timestamp (s),Altitude (m),Velocity (m/s),Acceleration (m/s^2)
0.0,0.0,0.0,30.0
1.0,15.0,30.0,20.0
2.0,55.0,40.0,10.0
3.0,105.0,50.0,0.0
4.0,155.0,50.0,-10.0
5.0,195.0,40.0,-10.0
6.0,225.0,30.0,-10.0
7.0,245.0,20.0,-10.0
8.0,255.0,10.0,-10.0
9.0,255.0,0.0,-10.0
10.0,245.0,-10.0,-10.0`

	// Using Event Names expected by Run's hardcoded values/findEventTime calls
	// Updated to 4 columns (#, ts, event, out_idx) as expected by LoadEventInfo
	eventInfoCSV := `#,"Timestamp (s)","Event Name",out_idx
1,0.0,EV_LIFTOFF,0
2,5.5,EV_MECO,0
3,9.0,EV_APOGEE,0
4,9.5,EV_DROGUE_DEPLOY,0
5,15.0,EV_MAIN_DEPLOY,0`

	// Using State Names expected by Run's hardcoded values/findStateTime calls
	// Updated to 3 columns (#, ts, state) as expected by LoadFlightStates
	flightStatesCSV := `#,"Timestamp (s)","State Name"
1,3.1,EnginePressure
2,3.2,EngineTemp
3,9.1,ApogeeFlag
4,9.6,DroguePressure
5,122.12,TOUCHDOWN
6,15.1,MainPressure`

	benchmarkName := "hipr-euroc24" // Match the benchmark's Name()

	// --- Setup Temp Dir ---
	tempDataPathBase := setupTempBenchdata(t, benchmarkName, flightInfoCSV, eventInfoCSV, flightStatesCSV)


	// --- Configure and Run Benchmark ---
	cfg := BenchmarkConfig{
		BenchdataPath: tempDataPathBase, // Config points to the base temp dir
		// Removed incorrect tolerance fields
	}
	b := NewHiprEuroc24Benchmark(cfg)

	// Run Setup first (as done in main.go logic flow)
	err := b.Setup()
	require.NoError(t, err, "Setup failed")

	// LoadData requires the *base* path from config, it calculates the subpath internally
	err = b.LoadData(cfg.BenchdataPath) // Pass the base path
	require.NoError(t, err, "LoadData failed")

	// Run expects simData interface{}, pass nil as it's not used by this implementation
	results, err := b.Run(nil) // Pass nil for simData
	require.NoError(t, err, "Run failed")
	require.NotEmpty(t, results, "Run should produce results")

	// --- Assert Results ---
	resultsMap := make(map[string]BenchmarkResult)
	for _, res := range results {
		resultsMap[res.Metric] = res
	}

	// Check expected metrics based on Run implementation's hardcoded ground truths
	// const expectedApogeeGroundTruth = 7448.0
	// const expectedMaxVGroundTruth = 1055.31
	// const expectedLiftoffTimeGroundTruth = 0.0
	// const expectedApogeeTimeGroundTruth = 20.60
	// const expectedTouchdownTimeGroundTruth = 122.12

	// Apogee Height (Compares actual from data: 255.0 vs hardcoded GT: 7448.0)
	apogeeRes, ok := resultsMap["Apogee"] // Metric name is "Apogee" in compareFloat call
	require.True(t, ok, "Apogee metric missing")
	assert.Equal(t, 255.0, apogeeRes.Actual, "Incorrect Actual Apogee Height from data")
	assert.Equal(t, 7448.0, apogeeRes.Expected, "Incorrect Expected Apogee Height (Ground Truth)")
	assert.False(t, apogeeRes.Passed, "Apogee Height should fail (test data vs ground truth)")

	// Max Velocity (Compares actual from data: 50.0 vs hardcoded GT: 1055.31)
	maxVelRes, ok := resultsMap["Max Velocity"]
	require.True(t, ok, "Max Velocity metric missing")
	assert.Equal(t, 50.0, maxVelRes.Actual, "Incorrect Actual Max Velocity from data")
	assert.Equal(t, 1055.31, maxVelRes.Expected, "Incorrect Expected Max Velocity (Ground Truth)")
	assert.False(t, maxVelRes.Passed, "Max Velocity should fail (test data vs ground truth)")

	// Liftoff Time (Compares actual from data: 0.0 vs hardcoded GT: 0.0)
	liftoffTimeRes, ok := resultsMap["Liftoff Time"]
	require.True(t, ok, "Liftoff Time metric missing")
	assert.Equal(t, 0.0, liftoffTimeRes.Actual, "Incorrect Actual Liftoff Time from data")
	assert.Equal(t, 0.0, liftoffTimeRes.Expected, "Incorrect Expected Liftoff Time (Ground Truth)")
	assert.True(t, liftoffTimeRes.Passed, "Liftoff Time should pass")

	// Apogee Event Time (Compares actual from data: 9.0 vs hardcoded GT: 20.60)
	apogeeEventTimeRes, ok := resultsMap["Apogee Event Time"]
	require.True(t, ok, "Apogee Event Time metric missing")
	assert.Equal(t, 9.0, apogeeEventTimeRes.Actual, "Incorrect Actual Apogee Event Time from data")
	assert.Equal(t, 20.60, apogeeEventTimeRes.Expected, "Incorrect Expected Apogee Event Time (Ground Truth)")
	assert.False(t, apogeeEventTimeRes.Passed, "Apogee Event Time should fail (test data vs ground truth)")

	// Touchdown Time (Compares actual from data: 122.12 vs hardcoded GT: 122.12)
	touchdownTimeRes, ok := resultsMap["Touchdown Time"]
	require.True(t, ok, "Touchdown Time metric missing")
	assert.Equal(t, 122.12, touchdownTimeRes.Actual, "Incorrect Actual Touchdown Time from data")
	assert.Equal(t, 122.12, touchdownTimeRes.Expected, "Incorrect Expected Touchdown Time (Ground Truth)")
	assert.True(t, touchdownTimeRes.Passed, "Touchdown Time should pass")

}

// TODO: Add specific tests for findEventTime and findStateTime for edge cases if needed
