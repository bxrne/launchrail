package components_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"

	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

func createTestMotor() (*components.Motor, *thrustcurves.MotorData) {
	// Create simple thrust curve data
	thrustData := [][]float64{
		{0.0, 10.0}, // Initial thrust
		{1.0, 10.0}, // Constant thrust for 1 second
		{2.0, 0.0},  // Burnout
	}

	motorData := &thrustcurves.MotorData{
		Thrust:    thrustData,
		TotalMass: 12.0, // Initial mass (10kg propellant + 2kg casing)
		WetMass:   10.0, // Propellant mass
		BurnTime:  2.0,  // 2 second burn
		AvgThrust: 10.0, // Average thrust
	}

	logger := logf.New(logf.Opts{})
	motor, err := components.NewMotor(ecs.NewBasic(), motorData, logger)

	if err != nil {
		panic(err)
	}

	return motor, motorData
}

// TEST: GIVEN a new Motor WHEN NewMotor is called THEN a new Motor is returned
func TestNewMotor(t *testing.T) {
	logger := logf.New(logf.Opts{})
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0.0, 10.0}, {1.0, 5.0}, {2.0, 0.0}},
		TotalMass: 2.0,
		BurnTime:  2.0,
		AvgThrust: 7.5,
	}

	motor, err := components.NewMotor(ecs.BasicEntity{}, md, logger)
	require.NoError(t, err)
	require.NotNil(t, motor)

	// Calculate expected thrust with efficiency factors: 0.85 * 0.90 * 0.97 = 0.74025

	assert.InDelta(t, motor.GetThrust(), 10.0, 0.0001) // Initial thrust is raw value (no efficiency factor applied)
	assert.Equal(t, 2.0, motor.GetMass())
}

// TEST: GIVEN a motor with constant thrust WHEN Update is called THEN thrust and mass are correctly calculated
func TestMotorUpdate(t *testing.T) {
	motor, _ := createTestMotor()

	// Calculate expected thrust with efficiency factors
	efficiencyFactor := 0.85 * 0.90 * 0.97 // nozzle * combustion * friction
	expectedThrust := 10.0 * efficiencyFactor

	// Test initial conditions
	assert.InDelta(t, motor.GetThrust(), expectedThrust, 0.0001)
	assert.Equal(t, 12.0, motor.GetMass()) // 10kg propellant + 2kg casing

	// Update with 0.5 second step
	err := motor.Update(0.5)
	assert.NoError(t, err)

	// After 0.5s (25% of burn time), mass should be reduced by 25% of propellant mass
	assert.InDelta(t, motor.GetThrust(), expectedThrust, 0.0001)
	expectedMass := 12.0 - (10.0 * 0.5 / 2.0) // Initial - (PropellantMass * TimeElapsed/BurnTime)
	assert.Equal(t, expectedMass, motor.GetMass())
	t.Logf("State after first update: %s", motor.GetState())

	// Update to burnout
	err = motor.Update(1.5)
	assert.NoError(t, err)

	t.Logf("State after second update: %s", motor.GetState())
	assert.Equal(t, 0.0, motor.GetThrust())
	assert.Equal(t, motor.GetCasingMass(), motor.GetMass()) // Only casing mass remains
}

// TEST: GIVEN a Motor WHEN Update is called THEN the Motor is updated
func TestMotorBurnout(t *testing.T) {
	id := ecs.BasicEntity{}
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0.0, 10.0}, {1.0, 5.0}, {2.0, 0.0}}, // Ends at 2s
		BurnTime:  2.0,
		AvgThrust: 5.0,
		TotalMass: 1.0,
	}
	logger := logf.New(logf.Opts{})
	motor, err := components.NewMotor(id, md, logger)
	require.NoError(t, err)

	require.NotNil(t, motor)

	// Simulate full burn
	for i := 0; i < 30; i++ { // Ensures we pass burn time
		err := motor.Update(0.1)
		assert.NoError(t, err)
	}

	assert.True(t, motor.IsCoasting(), "Motor should be coasting after burn")
	assert.Zero(t, motor.GetThrust(), "Thrust should be zero after burnout")
}

// TEST: GIVEN a Motor WHEN Update is called THEN the Motor is updated
func TestMotorReset(t *testing.T) {
	logger := logf.New(logf.Opts{})
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0.0, 10.0}, {1.0, 5.0}, {2.0, 0.0}},
		TotalMass: 2.0,
		BurnTime:  2.0,
		AvgThrust: 7.5,
	}

	motor, err := components.NewMotor(ecs.BasicEntity{}, md, logger)
	require.NoError(t, err)

	_ = motor.Update(1.5)
	motor.Reset()

	assert.Equal(t, 10.0, motor.GetThrust()) // Reset should restore initial thrust
	assert.Equal(t, 2.0, motor.GetMass())
	assert.False(t, motor.IsCoasting())
	assert.Equal(t, "IGNITED", motor.GetState()) // Check FSM state
}

// TEST: GIVEN a Motor WHEN Update is called THEN the Motor is updated
func TestInvalidUpdate(t *testing.T) {
	logger := logf.New(logf.Opts{})
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0.0, 10.0}, {1.0, 5.0}, {2.0, 0.0}},
		TotalMass: 2.0,
		BurnTime:  2.0,
		AvgThrust: 7.5,
	}

	motor, err := components.NewMotor(ecs.BasicEntity{}, md, logger)
	require.NoError(t, err)
	err = motor.Update(-0.1) // Invalid negative timestep
	assert.Error(t, err)
}

func TestMotorBurnsFullDuration(t *testing.T) {
	motor, motorData := createTestMotor()

	totalSteps := int(motorData.BurnTime / 0.1) // Simulate in 0.1s steps
	for i := 0; i < totalSteps; i++ {
		err := motor.Update(0.1)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, motor.GetThrust(), 0.0, "Thrust should never be negative")
	}

	// Ensure motor burns for full duration
	assert.InDelta(t, motorData.BurnTime, motor.GetElapsedTime(), 0.000001, "Motor should burn for full burnTime")
	assert.Zero(t, motor.GetThrust(), "Thrust should be zero after burnout")
	assert.True(t, motor.IsCoasting(), "Motor should be coasting after burn")
}
