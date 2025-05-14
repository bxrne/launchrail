package reporting

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
)

func TestLoadSimulationConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "launchrail-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock current config
	currentConfig := &config.Config{
		Setup: config.Setup{
			App: config.App{
				Version: "1.0.0-test",
			},
		},
	}

	// Test 1: No config file exists, should return current config
	logObj := logf.New(logf.Opts{Level: logf.DebugLevel})
	logger := &logObj
	resultConfig := LoadSimulationConfig(tmpDir, currentConfig, logger)
	assert.Equal(t, currentConfig, resultConfig, "Should return current config when no config file exists")

	// Test 2: Create a valid config file
	validConfig := `{"setup":{"app":{"version":"2.0.0-stored"}}}`
	configPath := filepath.Join(tmpDir, "engine_config.json")
	err = os.WriteFile(configPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	resultConfig = LoadSimulationConfig(tmpDir, currentConfig, logger)
	assert.Equal(t, "2.0.0-stored", resultConfig.Setup.App.Version, "Should load the stored config when it exists")

	// Test 3: Invalid config file (not valid JSON)
	err = os.WriteFile(configPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	resultConfig = LoadSimulationConfig(tmpDir, currentConfig, logger)
	assert.Equal(t, currentConfig, resultConfig, "Should return current config when config file is invalid")
}

func TestLoadCSVData(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "launchrail-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logObj := logf.New(logf.Opts{Level: logf.DebugLevel})
	logger := &logObj

	// Create a mock record
	record := &storage.Record{
		Hash: "test-hash",
		Path: tmpDir,
	}

	// Test 1: No CSV files available
	simData, err := LoadCSVData(record, logger)
	assert.NoError(t, err, "Should not return error when no files are available")
	assert.NotNil(t, simData, "Should return non-nil SimulationData")
	assert.Nil(t, simData.MotionData, "MotionData should be nil when no files are available")
	assert.Nil(t, simData.EventsData, "EventsData should be nil when no files are available")

	// For a more comprehensive test, we would need to mock the Storage interface
	// which requires more setup. I'll add stub methods for now.
}

func TestParseToFloat64(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"123.45", 123.45, false},
		{"0", 0.0, false},
		{"-42.5", -42.5, false},
		{"abc", 0.0, true},
		{"", 0.0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseToFloat64(tc.input)

			if tc.hasError {
				assert.Error(t, err, "Should return error for input: %s", tc.input)
			} else {
				assert.NoError(t, err, "Should not return error for input: %s", tc.input)
				assert.Equal(t, tc.expected, result, "Should parse %s to %f", tc.input, tc.expected)
			}
		})
	}
}

func TestLoadMotorData(t *testing.T) {
	logObj := logf.New(logf.Opts{Level: logf.DebugLevel})
	logger := &logObj

	// Test 1: Empty motor designation
	records, headers := LoadMotorData("", logger)
	assert.Nil(t, records, "Should return nil records for empty designation")
	assert.Nil(t, headers, "Should return nil headers for empty designation")

	// Test 2: Valid motor designation
	records2, headers2 := LoadMotorData("A8-3", logger)
	assert.NotNil(t, records2, "Should return non-nil records for valid designation")
	assert.Equal(t, 10, len(records2), "Should return 10 sample records")
	assert.Equal(t, 2, len(headers2), "Should return 2 headers")
	assert.Equal(t, ColumnTimeSeconds, headers2[0], "First header should be time")
	assert.Equal(t, ColumnThrustNewtons, headers2[1], "Second header should be thrust")

	// Check shape of thrust curve (first should be low, middle should be peak, last should be low)
	assert.Less(t, records2[0].GetFloat(ColumnThrustNewtons), records2[4].GetFloat(ColumnThrustNewtons), 
		"Thrust should increase from start to middle")
	assert.Greater(t, records2[4].GetFloat(ColumnThrustNewtons), records2[9].GetFloat(ColumnThrustNewtons), 
		"Thrust should decrease from middle to end")
}

// Helper method for PlotSimRecord to get float value with error checking
func (p PlotSimRecord) GetFloat(key string) float64 {
	if val, ok := p[key].(float64); ok {
		return val
	}
	return 0.0
}
