package reporting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
)

func TestFindColumnIndices(t *testing.T) {
	testCases := []struct {
		name                  string
		headers               []string
		expectedTimeIdx       int
		expectedEventNameIdx  int
		expectedStatusIdx     int
		expectedParaStatusIdx int
		expectedParaTypeIdx   int
	}{
		{
			name:                  "Standard headers",
			headers:               []string{"Event", "Time", "Status", "Parachute_Status", "Parachute_Type"},
			expectedTimeIdx:       1,
			expectedEventNameIdx:  0,
			expectedStatusIdx:     2,
			expectedParaStatusIdx: 3,
			expectedParaTypeIdx:   4,
		},
		{
			name:                  "Mixed capitalization",
			headers:               []string{"event", "TIME", "status", "parachute_status", "parachute_type"},
			expectedTimeIdx:       1,
			expectedEventNameIdx:  0,
			expectedStatusIdx:     2,
			expectedParaStatusIdx: 3,
			expectedParaTypeIdx:   4,
		},
		{
			name:                  "Alternative naming",
			headers:               []string{"Name", "Time (s)", "State", "Chute Status", "Chute Type"},
			expectedTimeIdx:       1,
			expectedEventNameIdx:  0,
			expectedStatusIdx:     2,
			expectedParaStatusIdx: 3,
			expectedParaTypeIdx:   4,
		},
		{
			name:                  "Missing columns",
			headers:               []string{"Data1", "Data2", "Data3"},
			expectedTimeIdx:       1, // Default
			expectedEventNameIdx:  0, // Default
			expectedStatusIdx:     -1,
			expectedParaStatusIdx: -1,
			expectedParaTypeIdx:   -1,
		},
		{
			name:                  "Empty headers",
			headers:               []string{},
			expectedTimeIdx:       1, // Default
			expectedEventNameIdx:  0, // Default
			expectedStatusIdx:     -1,
			expectedParaStatusIdx: -1,
			expectedParaTypeIdx:   -1,
		},
	}

	logObj := logf.New(logf.Opts{Level: logf.ErrorLevel})
	logger := &logObj

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			timeIdx, eventNameIdx, statusIdx, paraStatusIdx, paraTypeIdx := findColumnIndices(tc.headers, logger)
			
			assert.Equal(t, tc.expectedTimeIdx, timeIdx, "Time index mismatch")
			assert.Equal(t, tc.expectedEventNameIdx, eventNameIdx, "Event name index mismatch")
			assert.Equal(t, tc.expectedStatusIdx, statusIdx, "Status index mismatch")
			assert.Equal(t, tc.expectedParaStatusIdx, paraStatusIdx, "Parachute status index mismatch")
			assert.Equal(t, tc.expectedParaTypeIdx, paraTypeIdx, "Parachute type index mismatch")
		})
	}
}

func TestParseDeploymentTime(t *testing.T) {
	testCases := []struct {
		input    string
		expected float64
		success  bool
	}{
		{"10.5", 10.5, true},
		{"0", 0.0, true},
		{"-1.2", -1.2, true},
		{"invalid", 0.0, false},
		{"", 0.0, false},
	}

	logObj := logf.New(logf.Opts{Level: logf.ErrorLevel})
	logger := &logObj

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, ok := parseDeploymentTime(tc.input, logger)
			
			assert.Equal(t, tc.success, ok, "Success flag mismatch")
			if tc.success {
				assert.Equal(t, tc.expected, result, "Parsed value mismatch")
			}
		})
	}
}

func TestDetermineParachuteType(t *testing.T) {
	testCases := []struct {
		name          string
		row           []string
		eventNameIdx  int
		parachuteTypeIdx int
		defaultType   string
		expected      string
	}{
		{
			name:          "Use explicit type",
			row:           []string{"Deploy", "10.5", "DEPLOYED", "Drogue"},
			eventNameIdx:  0,
			parachuteTypeIdx: 3,
			defaultType:   "Parachute",
			expected:      "Drogue",
		},
		{
			name:          "Detect drogue from event name",
			row:           []string{"Drogue Deploy", "10.5", "DEPLOYED", ""},
			eventNameIdx:  0,
			parachuteTypeIdx: 3,
			defaultType:   "Parachute",
			expected:      RecoverySystemDrogue,
		},
		{
			name:          "Detect main from event name",
			row:           []string{"Main Parachute", "10.5", "DEPLOYED", ""},
			eventNameIdx:  0,
			parachuteTypeIdx: 3,
			defaultType:   "Parachute",
			expected:      RecoverySystemMain,
		},
		{
			name:          "Use default when no type info",
			row:           []string{"Parachute", "10.5", "DEPLOYED", ""},
			eventNameIdx:  0,
			parachuteTypeIdx: 3,
			defaultType:   "Default Chute",
			expected:      "Default Chute",
		},
		{
			name:          "Invalid indices",
			row:           []string{"Parachute"},
			eventNameIdx:  5,
			parachuteTypeIdx: 5,
			defaultType:   "Default Chute",
			expected:      "Default Chute",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := determineParachuteType(tc.row, tc.eventNameIdx, tc.parachuteTypeIdx, tc.defaultType)
			assert.Equal(t, tc.expected, result, "Parachute type detection mismatch")
		})
	}
}

func TestGetDescentRate(t *testing.T) {
	testCases := []struct {
		parachuteType string
		expected      float64
	}{
		{RecoverySystemDrogue, DefaultDescentRateDrogue},
		{RecoverySystemMain, DefaultDescentRateMain},
		{"drogue parachute", DefaultDescentRateDrogue},
		{"MAIN PARACHUTE", DefaultDescentRateMain},
		{"drogue something", DefaultDescentRateDrogue},
		{"something main", DefaultDescentRateMain},
		{"Other", 15.0}, // Default fallback
	}

	for _, tc := range testCases {
		t.Run(tc.parachuteType, func(t *testing.T) {
			result := getDescentRate(tc.parachuteType)
			assert.Equal(t, tc.expected, result, "Descent rate mismatch")
		})
	}
}

func TestIsDeployedStatus(t *testing.T) {
	testCases := []struct {
		status   string
		expected bool
	}{
		{StatusDeployed, true},
		{"DEPLOYED", true},
		{"deployed", true},
		{" DEPLOYED ", true},
		{"TRUE", true},
		{"1", true},
		{"ARMED", false},
		{"SAFE", false},
		{"0", false},
		{"FALSE", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.status, func(t *testing.T) {
			result := isDeployedStatus(tc.status)
			assert.Equal(t, tc.expected, result, "Deployed status detection mismatch")
		})
	}
}

func TestProcessParachuteDeployment(t *testing.T) {
	logObj := logf.New(logf.Opts{Level: logf.ErrorLevel})
	logger := &logObj
	parachuteMap := make(map[string]RecoverySystemData)
	
	// Test for drogue parachute
	processParachuteDeployment(RecoverySystemDrogue, 10.5, logger, parachuteMap)
	
	drogue, exists := parachuteMap[RecoverySystemDrogue]
	assert.True(t, exists, "Drogue entry should be created")
	assert.Equal(t, RecoverySystemDrogue, drogue.Type, "Type should match")
	assert.Equal(t, 10.5, drogue.Deployment, "Deployment time should match")
	assert.Equal(t, DefaultDescentRateDrogue, drogue.DescentRate, "Descent rate should match")
	
	// Test for main parachute
	processParachuteDeployment(RecoverySystemMain, 15.5, logger, parachuteMap)
	
	main, exists := parachuteMap[RecoverySystemMain]
	assert.True(t, exists, "Main entry should be created")
	assert.Equal(t, RecoverySystemMain, main.Type, "Type should match")
	assert.Equal(t, 15.5, main.Deployment, "Deployment time should match")
	assert.Equal(t, DefaultDescentRateMain, main.DescentRate, "Descent rate should match")
	
	// Test for updating existing entry
	processParachuteDeployment(RecoverySystemDrogue, 12.0, logger, parachuteMap)
	
	drogue, exists = parachuteMap[RecoverySystemDrogue]
	assert.True(t, exists, "Drogue entry should still exist")
	assert.Equal(t, 12.0, drogue.Deployment, "Deployment time should be updated")
}
