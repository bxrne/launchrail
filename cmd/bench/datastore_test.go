package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

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

	data, err := LoadFlightInfo(filePath)
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
			name:       "Missing Column",
			csvContent: "ts,height,velocity\n1.0,10.0,5.0",
			expectedErr: "unexpected number of columns",
		},
		{
			name:       "Invalid Float",
			csvContent: "ts,height,velocity,acceleration,motor_designation\n1.0,ten,5.0,1.0,A8-3",
			expectedErr: "invalid float value 'ten'",
		},
		{
			name:       "Empty File",
			csvContent: "",
			expectedErr: "failed to read header", // EOF error on header read
		},
		{
			name:       "Header Only",
			csvContent: "ts,height,velocity,acceleration,motor_designation\n",
			expectedErr: "no data rows found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempCSV(t, tt.csvContent)
			data, err := LoadFlightInfo(filePath)

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
		name        string
		csvContent  string
		expectedData []EventInfo
		expectedErr string
	}{
		{
			name: "Success",
			csvContent: `timestamp,event,outidx
1.5,LAUNCH,0
10.2,APOGEE,1
20.8,LANDING,2`, // Note: outidx is parsed as int
			expectedData: []EventInfo{
				{Timestamp: 1.5, Event: "LAUNCH"},
				{Timestamp: 10.2, Event: "APOGEE"},
				{Timestamp: 20.8, Event: "LANDING"},
			},
			expectedErr: "",
		},
		{
			name:       "Wrong Column Count",
			csvContent: `timestamp,event
1.5,LAUNCH`, // Missing outidx
			expectedErr: "unexpected number of columns",
		},
		{
			name:       "Invalid Timestamp Float",
			csvContent: `timestamp,event,outidx
1.5x,LAUNCH,0`, 
			expectedErr: "invalid float value '1.5x'",
		},
		// Removed Invalid_OutIdx_Int test case as the column is now ignored
		// {
		// 	name:       "Invalid OutIdx Int",
		// 	csvContent: `timestamp,event,outidx
		// 1.5,LAUNCH,zero`, // 'zero' cannot be parsed as int
		// 	expectedErr: "invalid integer value 'zero'",
		// },
		{
			name:       "Empty File",
			csvContent: "",
			expectedErr: "failed to read header",
		},
		{
			name:       "Header Only",
			csvContent: `timestamp,event,outidx
`,
			expectedErr: "no data rows found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempCSV(t, tt.csvContent)
			data, err := LoadEventInfo(filePath)

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

func TestLoadFlightStates(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		expectedData []FlightState
		expectedErr string
	}{
		{
			name: "Success",
			csvContent: `ts,state
0.1,PRELAUNCH
10.5,POWERED_ASCENT
25.2,COAST`, 
			expectedData: []FlightState{
				{Timestamp: 0.1, State: "PRELAUNCH"},
				{Timestamp: 10.5, State: "POWERED_ASCENT"},
				{Timestamp: 25.2, State: "COAST"},
			},
			expectedErr: "",
		},
		{
			name:       "Wrong Column Count",
			csvContent: `ts,state,extra
0.1,PRELAUNCH,oops`, 
			expectedErr: "unexpected number of columns",
		},
		{
			name:       "Invalid Timestamp Float",
			csvContent: `ts,state
0.x,PRELAUNCH`, 
			expectedErr: "invalid float value '0.x'",
		},
		{
			name:       "Empty File",
			csvContent: "",
			expectedErr: "failed to read header",
		},
		{
			name:       "Header Only",
			csvContent: `ts,state
`,
			expectedErr: "no data rows found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempCSV(t, tt.csvContent)
			data, err := LoadFlightStates(filePath)

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
		{"Invalid Float String", "abc", 10, "Alpha", "alpha.csv", 0, "invalid float value 'abc'"},
		{"Empty String", "", 2, "Empty", "empty.csv", 0, "invalid float value ''"},
		{"Float with Extra Chars", "1.2x", 8, "Extra", "extra.csv", 0, "invalid float value '1.2x'"},
		{"NaN String", "NaN", 3, "NotNum", "nan.csv", math.NaN(), ""},
		{"Inf String", "Inf", 4, "Infinite", "inf.csv", math.Inf(1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := parseFloat(tt.inputStr, tt.rowIdx, tt.colName, tt.fileName)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr)
				// Check that row, col, filename are in error message
				assert.ErrorContains(t, err, fmt.Sprintf("row %d", tt.rowIdx+2)) // +2 for 1-based and header
				assert.ErrorContains(t, err, fmt.Sprintf("column %s", tt.colName))
				assert.ErrorContains(t, err, filepath.Base(tt.fileName))
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
