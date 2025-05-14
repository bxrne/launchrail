package main_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	bench "github.com/bxrne/launchrail/cmd/bench"
	"github.com/bxrne/launchrail/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary CSV file for testing
func createTempCSV(t *testing.T, content string) string {
	t.Helper()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.csv")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create temp CSV file")
	return filePath
}

func TestLoadFlightInfo(t *testing.T) {
	csvContent := `ts,height,velocity,acceleration,motor_designation
-0.75,0.45,0.14,-0.02,TEST
-0.74,0.45,0.14,-0.02,TEST
21.22,7444.18,-10.90,-9.8,TEST
` // Added apogee data point
	filePath := createTempCSV(t, csvContent)

	data, err := bench.LoadFlightInfo(filePath)
	require.NoError(t, err, "LoadFlightInfo failed")
	require.Len(t, data, 3, "Incorrect number of records loaded")

	assert.Equal(t, -0.75, data[0].Timestamp)
	assert.Equal(t, 0.45, data[0].Height)
	assert.Equal(t, 0.14, data[0].Velocity)
	assert.Equal(t, -0.02, data[0].Acceleration)
	assert.Equal(t, "TEST", data[0].MotorDesignation)

	assert.Equal(t, 21.22, data[2].Timestamp)
	assert.Equal(t, 7444.18, data[2].Height)
	assert.Equal(t, -10.90, data[2].Velocity)
	assert.Equal(t, -9.8, data[2].Acceleration)
	assert.Equal(t, "TEST", data[2].MotorDesignation)
}

func TestLoadFlightInfo_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		expectedErr string
	}{
		{
			name:        "Missing Column",
			csvContent:  "ts,height,velocity\n1.0,10.0,5.0",
			expectedErr: "unexpected number of columns",
		},
		{
			name:        "Invalid Float",
			csvContent:  "ts,height,velocity,acceleration,motor_designation\n1.0,ten,5.0,1.0,A8-3",
			expectedErr: "invalid float value 'ten'",
		},
		{
			name:        "Empty File",
			csvContent:  "",
			expectedErr: "failed to read header", // EOF error on header read
		},
		{
			name:        "Header Only",
			csvContent:  "ts,height,velocity,acceleration,motor_designation\n",
			expectedErr: "no data rows found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempCSV(t, tt.csvContent)
			data, err := bench.LoadFlightInfo(filePath)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, data)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, data)
				if tt.name == "Header Only" {
					assert.Empty(t, data)
				}
			}
		})
	}
}

func TestLoadEventInfo(t *testing.T) {
	tests := []struct {
		name         string
		csvContent   string
		expectedData []bench.EventInfo
		expectedErr  string
	}{
		{
			name: "Success",
			// Now expecting 3 columns: idx, ts, event. LoadEventInfo uses 2nd (ts) and 3rd (event).
			csvContent: `idx,ts,event
0,0.0,LAUNCH
1,30.5,APOGEE
2,60.2,LANDED
`,
			expectedData: []bench.EventInfo{
				{Timestamp: 0.0, Event: "LAUNCH"},
				{Timestamp: 30.5, Event: "APOGEE"},
				{Timestamp: 60.2, Event: "LANDED"},
			},
			expectedErr: "",
		},
		{
			name:        "Empty File",
			csvContent:  "",
			expectedErr: "failed to read header from test.csv",
		},
		{
			name: "Header Only",
			csvContent: `idx,ts,event
`,
			expectedErr: "no data rows found in test.csv",
		},
		{
			name: "Wrong Column Count",
			// Providing 2 columns when at least 3 are expected
			csvContent: `ts,event
1.0,MY_EVENT
`,
			expectedErr: "unexpected number of columns in test.csv, row 2 (1-based data): got 2, want at least 3",
		},
		{
			name: "Invalid Timestamp Float",
			// 3 columns, but timestamp (2nd col) is invalid
			csvContent: `idx,ts,event
0,invalid,LAUNCH_ATTEMPT
`,
			expectedErr: "invalid float value 'invalid' in test.csv, row 2, column Timestamp",
		},
		{
			name: "Wrong Column Name", // LoadEventInfo doesn't check header names, only position/count.
			// This test will pass if there are 3+ columns and data types are correct.
			csvContent: `id,timestamp_val,event_text
0,1.0,EVENT
`,
			expectedData: []bench.EventInfo{{Timestamp: 1.0, Event: "EVENT"}},
			expectedErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempCSV(t, tt.csvContent)
			data, err := bench.LoadEventInfo(filePath)

			if tt.expectedErr != "" {
				require.Error(t, err, "Expected an error for test case: %s", tt.name)
				assert.ErrorContains(t, err, tt.expectedErr, "Error message mismatch for test case: %s", tt.name)
				assert.Nil(t, data, "Data should be nil on error for test case: %s", tt.name)
			} else {
				require.NoError(t, err, "Did not expect an error for test case: %s, got: %v", tt.name, err)
				assert.Equal(t, tt.expectedData, data, "Data mismatch for test case: %s", tt.name)
			}
		})
	}
}

func TestLoadFlightStates(t *testing.T) {
	tests := []struct {
		name         string
		csvContent   string
		expectedData []bench.FlightState
		expectedErr  string
	}{
		{
			name: "Success",
			csvContent: `ts,state
0.1,PRELAUNCH
10.5,POWERED_ASCENT
25.2,COAST`,
			expectedData: []bench.FlightState{
				{Timestamp: 0.1, State: "PRELAUNCH"},
				{Timestamp: 10.5, State: "POWERED_ASCENT"},
				{Timestamp: 25.2, State: "COAST"},
			},
			expectedErr: "",
		},
		{
			name: "Wrong Column Count",
			csvContent: `ts,state,extra
0.1,PRELAUNCH,oops`,
			expectedErr: "unexpected number of columns",
		},
		{
			name: "Invalid Timestamp Float",
			csvContent: `ts,state
0.x,PRELAUNCH`,
			expectedErr: "invalid float value '0.x'",
		},
		{
			name:        "Empty File",
			csvContent:  "",
			expectedErr: "failed to read header",
		},
		{
			name: "Header Only",
			csvContent: `ts,state
`,
			expectedErr: "no data rows found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempCSV(t, tt.csvContent)
			data, err := bench.LoadFlightStates(filePath)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, data)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedData, data)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name        string
		inputStr    string
		rowIdx      int
		colName     string
		fileName    string
		expectedVal float64
		expectedErr string // Substring of the expected error
	}{
		{"Valid Float", "123.45", 5, "TestCol", "test.csv", 123.45, ""},
		{"Valid Negative Float", "-0.99", 1, "NegVal", "neg.csv", -0.99, ""},
		{"Invalid Float String", "abc", 10, "Alpha", "alpha.csv", 0, "must be a valid number: strconv.ParseFloat: parsing \"abc\": invalid syntax"},
		{"Empty String", "", 2, "Empty", "empty.csv", 0, "is required"},
		{"Float with Extra Chars", "1.2x", 8, "Extra", "extra.csv", 0, "must be a valid number: strconv.ParseFloat: parsing \"1.2x\": invalid syntax"},
		{"NaN String", "NaN", 3, "NotNum", "nan.csv", math.NaN(), ""},
		{"Inf String", "Inf", 4, "Infinite", "inf.csv", math.Inf(1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := utils.ParseFloat(tt.inputStr, tt.colName)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				// Special handling for NaN comparison
				if math.IsNaN(tt.expectedVal) {
					assert.True(t, math.IsNaN(val), "Expected NaN, got %v", val)
				} else {
					assert.Equal(t, tt.expectedVal, val)
				}
			}
		})
	}
}
