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

func TestCompareFloat(t *testing.T) {
	tests := []struct {
		name              string
		expected          float64
		actual            float64
		tolerancePercent  float64
		wantPassed        bool
		wantToleranceType string
		wantToleranceVal  float64 // Expected calculated tolerance
	}{
		{"within tolerance", 100.0, 102.0, 0.05, true, "relative", 5.0},
		{"exact match", 100.0, 100.0, 0.05, true, "relative", 5.0},
		{"outside tolerance", 100.0, 106.0, 0.05, false, "relative", 5.0},
		{"negative within tolerance", -100.0, -102.0, 0.05, true, "relative", 5.0},
		{"negative outside tolerance", -100.0, -106.0, 0.05, false, "relative", 5.0},
		{"zero expected, zero actual", 0.0, 0.0, 0.1, true, "absolute", 0.1},
		{"zero expected, within absolute tolerance", 0.0, 0.05, 0.1, true, "absolute", 0.1},
		{"zero expected, outside absolute tolerance", 0.0, 0.15, 0.1, false, "absolute", 0.1},
		{"zero tolerance", 100.0, 100.0, 0.0, true, "relative", 0.0},
		{"zero tolerance, diff", 100.0, 100.1, 0.0, false, "relative", 0.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := compareFloat("testMetric", "test description", tc.expected, tc.actual, tc.tolerancePercent)
			assert.Equal(t, tc.wantPassed, result.Passed)
			assert.Equal(t, tc.wantToleranceType, result.ToleranceType)
			// Use assert.InDelta for comparing the calculated tolerance value
			assert.InDelta(t, tc.wantToleranceVal, result.Tolerance, 1e-9)
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
	results, err := b.Run() // Pass nil for simData
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

// TestRunHiprEuroc24Benchmark requires setting up a mock simulation record
// or using actual simulation data, which makes it more of an integration test.
// This is a placeholder demonstrating the structure.
func TestRunHiprEuroc24Benchmark(t *testing.T) {
	// Setup: Create temp dir, mock config, mock record manager, mock record
	// Removed unused tempDir variable
	cfg := BenchmarkConfig{
		BenchdataPath: "testdata", // Assuming ground truth CSVs are in testdata/hipr-euroc24
		SimRecordHash: "mockSimHash",
		// Mock RecordManager needed here
	}

	// Mock the RecordManager and the GetRecord/LoadTelemetry interactions
	// This part is complex and requires a mocking library or manual mocks.
	// For now, we assume the benchmark can be created.
	b := NewHiprEuroc24Benchmark(cfg)

	// Load Ground Truth Data (requires testdata/hipr-euroc24 CSVs)
	err := b.LoadData(cfg.BenchdataPath)
	require.NoError(t, err, "Failed to load ground truth test data")

	// --- MOCKING LoadTelemetry --- (This is the hard part)
	// You would need to mock b.config.RecordManager.GetRecord to return a mock record,
	// and then mock the equivalent of LoadTelemetry to return specific test data.
	// Example (conceptual using testify/mock):
	// mockRM := new(MockRecordManager)
	// mockRec := new(MockRecord)
	// mockTelemetry := &storage.TelemetryData{ Time: []float64{0, 1, 2}, Altitude: []float64{0, 10, 5} ... }
	// mockRec.On("LoadTelemetry").Return(mockTelemetry, nil)
	// mockRM.On("GetRecord", "mockSimHash").Return(mockRec, nil)
	// b.config.RecordManager = mockRM

	// Execute Run (assuming mocks are set up)
	// results, err := b.Run() // This would fail without proper mocking

	// Assertions (Example - would depend on mocked telemetry)
	// assert.NoError(t, err)
	// assert.NotEmpty(t, results)
	// assert.True(t, results[0].Passed) // e.g., Check if Apogee passed based on mock data

	// Mark test as skipped until mocking is implemented
	t.Skip("Skipping TestRunHiprEuroc24Benchmark: Requires mocking RecordManager/Record/Telemetry")

	// Keep the original Run call structure for reference if needed later
	// results, err := b.Run() // Corrected call signature
	// require.NoError(t, err)
	// assert.NotEmpty(t, results)
}
