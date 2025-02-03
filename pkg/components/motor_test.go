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

// TEST: GIVEN a new Motor WHEN NewMotor is called THEN a new Motor is returned
func TestNewMotor(t *testing.T) {
	logger := logf.New(logf.Opts{})
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0.0, 10.0}, {1.0, 5.0}, {2.0, 0.0}},
		TotalMass: 2.0,
		BurnTime:  2.0,
		AvgThrust: 7.5,
	}

	motor := components.NewMotor(ecs.BasicEntity{}, md, logger)
	require.NotNil(t, motor)
	assert.Equal(t, 10.0, motor.GetThrust()) // Initial thrust should be first thrust point
	assert.Equal(t, 2.0, motor.GetMass())
}

// TEST: GIVEN a Motor WHEN Update is called THEN the Motor is updated
func TestMotorUpdate(t *testing.T) {
	logger := logf.New(logf.Opts{})
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0.0, 10.0}, {1.0, 5.0}, {2.0, 0.0}},
		TotalMass: 2.0,
		BurnTime:  2.0,
		AvgThrust: 7.5,
	}

	motor := components.NewMotor(ecs.BasicEntity{}, md, logger)
	err := motor.Update(0.5)
	assert.NoError(t, err)
	assert.Equal(t, 10.0, motor.GetThrust()) // Correct expected thrust after 0.5 seconds
	assert.Equal(t, motor.GetThrust(), 10.0)
	assert.Less(t, motor.GetMass(), 2.0)         // Mass should decrease
	assert.Equal(t, "BURNING", motor.GetState()) // Check FSM state
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
	motor := components.NewMotor(id, md, logger)

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

	motor := components.NewMotor(ecs.BasicEntity{}, md, logger)
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

	motor := components.NewMotor(ecs.BasicEntity{}, md, logger)
	err := motor.Update(-0.1) // Invalid negative timestep
	assert.Error(t, err)
}
