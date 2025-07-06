package reporting_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/reporting"
	"github.com/stretchr/testify/assert"
)

// TestRecoverySystemConstants tests recovery system type constants
func TestRecoverySystemConstants(t *testing.T) {
	assert.Equal(t, "Drogue Parachute", reporting.RecoverySystemDrogue)
	assert.Equal(t, "Main Parachute", reporting.RecoverySystemMain)

	// Test that constants are not empty
	assert.NotEmpty(t, reporting.RecoverySystemDrogue)
	assert.NotEmpty(t, reporting.RecoverySystemMain)
}

// TestEventConstants tests event type constants
func TestEventConstants(t *testing.T) {
	assert.Equal(t, "Launch", reporting.EventLaunch)
	assert.Equal(t, "Rail Exit", reporting.EventRailExit)
	assert.Equal(t, "Apogee", reporting.EventApogee)
	assert.Equal(t, "Touchdown", reporting.EventTouchdown)
	assert.Equal(t, "Burnout", reporting.EventBurnout)
	assert.Equal(t, "Deployment", reporting.EventDeployment)

	// Test that all event constants are non-empty
	events := []string{
		reporting.EventLaunch,
		reporting.EventRailExit,
		reporting.EventApogee,
		reporting.EventTouchdown,
		reporting.EventBurnout,
		reporting.EventDeployment,
	}

	for _, event := range events {
		assert.NotEmpty(t, event, "Event constant should not be empty")
	}
}

// TestStatusConstants tests status value constants
func TestStatusConstants(t *testing.T) {
	assert.Equal(t, "DEPLOYED", reporting.StatusDeployed)
	assert.Equal(t, "SAFE", reporting.StatusSafe)
	assert.Equal(t, "ARMED", reporting.StatusArmed)

	// Test that constants are not empty
	assert.NotEmpty(t, reporting.StatusDeployed)
	assert.NotEmpty(t, reporting.StatusSafe)
	assert.NotEmpty(t, reporting.StatusArmed)
}

// TestColumnConstants tests column header and label constants
func TestColumnConstants(t *testing.T) {
	assert.Equal(t, "Time (s)", reporting.ColumnTimeSeconds)
	assert.Equal(t, "Altitude (m)", reporting.ColumnAltitude)
	assert.Equal(t, "Velocity (m/s)", reporting.ColumnVelocity)
	assert.Equal(t, "Acceleration (m/s²)", reporting.ColumnAcceleration)
	assert.Equal(t, "Thrust (N)", reporting.ColumnThrustNewtons)
	assert.Equal(t, "Event", reporting.ColumnEventName)
	assert.Equal(t, "Status", reporting.ColumnEventStatus)
	assert.Equal(t, "Component", reporting.ColumnEventComponent)

	// Test that all column constants are non-empty
	columns := []string{
		reporting.ColumnTimeSeconds,
		reporting.ColumnAltitude,
		reporting.ColumnVelocity,
		reporting.ColumnAcceleration,
		reporting.ColumnThrustNewtons,
		reporting.ColumnEventName,
		reporting.ColumnEventStatus,
		reporting.ColumnEventComponent,
	}

	for _, column := range columns {
		assert.NotEmpty(t, column, "Column constant should not be empty")
	}
}

// TestDefaultValueConstants tests default value constants
func TestDefaultValueConstants(t *testing.T) {
	assert.Equal(t, 20.0, reporting.DefaultDescentRateDrogue)
	assert.Equal(t, 5.0, reporting.DefaultDescentRateMain)
	assert.Equal(t, 300.0, reporting.DefaultMainDeployAltitude)

	// Test that default values are positive
	assert.Positive(t, reporting.DefaultDescentRateDrogue)
	assert.Positive(t, reporting.DefaultDescentRateMain)
	assert.Positive(t, reporting.DefaultMainDeployAltitude)

	// Test logical relationships between defaults
	assert.Greater(t, reporting.DefaultDescentRateDrogue, reporting.DefaultDescentRateMain,
		"Drogue descent rate should be higher than main parachute descent rate")
}

// TestConstantsConsistency tests overall consistency of constants
func TestConstantsConsistency(t *testing.T) {
	// Test that event constants don't contain duplicate values
	events := map[string]string{
		"Launch":     reporting.EventLaunch,
		"Rail Exit":  reporting.EventRailExit,
		"Apogee":     reporting.EventApogee,
		"Touchdown":  reporting.EventTouchdown,
		"Burnout":    reporting.EventBurnout,
		"Deployment": reporting.EventDeployment,
	}

	seenValues := make(map[string]bool)
	for name, value := range events {
		assert.False(t, seenValues[value], "Event constant %s has duplicate value: %s", name, value)
		seenValues[value] = true
	}

	// Test that status constants don't contain duplicate values
	statuses := map[string]string{
		"Deployed": reporting.StatusDeployed,
		"Safe":     reporting.StatusSafe,
		"Armed":    reporting.StatusArmed,
	}

	seenStatusValues := make(map[string]bool)
	for name, value := range statuses {
		assert.False(t, seenStatusValues[value], "Status constant %s has duplicate value: %s", name, value)
		seenStatusValues[value] = true
	}
}
