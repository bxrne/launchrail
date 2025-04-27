package main

import (
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
	csvContent := `ts,height,velocity,acceleration
-0.75,0.45,0.14,-0.02
-0.74,0.45,0.14,-0.02
21.22,7444.18,-10.90,-9.8
` // Added apogee data point
	filePath := createTempCSV(t, csvContent)

	data, err := LoadFlightInfo(filePath)
	require.NoError(t, err, "LoadFlightInfo failed")
	require.Len(t, data, 3, "Incorrect number of records loaded")

	assert.Equal(t, -0.75, data[0].Timestamp)
	assert.Equal(t, 0.45, data[0].Height)
	assert.Equal(t, 0.14, data[0].Velocity)
	assert.Equal(t, -0.02, data[0].Acceleration)

	assert.Equal(t, 21.22, data[2].Timestamp)
	assert.Equal(t, 7444.18, data[2].Height)
	assert.Equal(t, -10.90, data[2].Velocity)
	assert.Equal(t, -9.8, data[2].Acceleration)
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
			csvContent: "ts,height,velocity,acceleration\n1.0,ten,5.0,1.0",
			expectedErr: "invalid float value 'ten'",
		},
		{
			name:       "Empty File",
			csvContent: "",
			expectedErr: "failed to read header", // EOF error on header read
		},
		{
			name:       "Header Only",
			csvContent: "ts,height,velocity,acceleration\n",
			expectedErr: "", // Should load zero records without error
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
	csvContent := `ts,event
0.1,LAUNCH
5.2,BURNOUT
21.2,APOGEE
30.5,DROGUE_DEPLOY
45.1,MAIN_DEPLOY
60.0,LANDED
`
	filePath := createTempCSV(t, csvContent)

	data, err := LoadEventInfo(filePath)
	require.NoError(t, err, "LoadEventInfo failed")
	require.Len(t, data, 6, "Incorrect number of events loaded")

	assert.Equal(t, 0.1, data[0].Timestamp)
	assert.Equal(t, "LAUNCH", data[0].Event)
	assert.Equal(t, 60.0, data[5].Timestamp)
	assert.Equal(t, "LANDED", data[5].Event)

	// Test error case (invalid float)
	invalidCsv := `ts,event
not_a_float,FAIL
`
	invalidFilePath := createTempCSV(t, invalidCsv)
	_, err = LoadEventInfo(invalidFilePath)
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid float value 'not_a_float'")

	// Test error case (missing column)
	missingColCsv := `ts
1.0
`
	missingColFilePath := createTempCSV(t, missingColCsv)
	_, err = LoadEventInfo(missingColFilePath)
	require.Error(t, err)
	assert.ErrorContains(t, err, "unexpected number of columns")
}

func TestLoadFlightStates(t *testing.T) {
	csvContent := `ts,state
0.0,PRELAUNCH
0.1,POWERED_ASCENT
5.2,COAST
21.2,APOGEE_STATE
21.3,DROGUE_DESCENT
45.1,MAIN_DESCENT
60.0,LANDED_STATE
` // Using distinct state names
	filePath := createTempCSV(t, csvContent)

	data, err := LoadFlightStates(filePath)
	require.NoError(t, err, "LoadFlightStates failed")
	require.Len(t, data, 7, "Incorrect number of states loaded")

	assert.Equal(t, 0.0, data[0].Timestamp)
	assert.Equal(t, "PRELAUNCH", data[0].State)
	assert.Equal(t, 60.0, data[6].Timestamp)
	assert.Equal(t, "LANDED_STATE", data[6].State)

	// Test error case (invalid float)
	invalidCsv := `ts,state
bad_ts,FAIL_STATE
`
	invalidFilePath := createTempCSV(t, invalidCsv)
	_, err = LoadFlightStates(invalidFilePath)
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid float value 'bad_ts'")

	// Test error case (missing column)
	missingColCsv := `ts
1.0
`
	missingColFilePath := createTempCSV(t, missingColCsv)
	_, err = LoadFlightStates(missingColFilePath)
	require.Error(t, err)
	assert.ErrorContains(t, err, "unexpected number of columns")
}
