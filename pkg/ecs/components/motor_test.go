package components_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/assert"
)

func TestNewMotor(t *testing.T) {
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0, 10}, {1, 20}, {2, 15}},
		TotalMass: 50,
		BurnTime:  2.0,
		AvgThrust: 15,
	}

	motor := components.NewMotor(1, md)

	assert.NotNil(t, motor, "Motor should be created successfully")
	assert.Equal(t, float64(10), motor.GetThrust(), "Initial thrust should match the first point in the thrust curve")
	assert.Equal(t, float64(50), motor.GetMass(), "Initial mass should match the provided TotalMass")
	assert.Equal(t, "idle", motor.FSM.GetState(), "Motor FSM should start in the 'idle' state")
}

func TestMotor_UpdateBurningState(t *testing.T) {
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0, 10}, {1, 20}, {2, 15}},
		TotalMass: 50,
		BurnTime:  2.0,
		AvgThrust: 15,
	}

	motor := components.NewMotor(1, md)

	// Simulate motor update during burn
	err := motor.Update(0.5)
	assert.NoError(t, err, "Motor update should not produce an error")
	assert.Equal(t, "burning", motor.FSM.GetState(), "Motor FSM should transition to 'burning' state")
	assert.Greater(t, motor.GetThrust(), 0.0, "Motor thrust should be greater than zero during burn")
	assert.Less(t, motor.GetMass(), float64(50), "Motor mass should decrease during burn")
}

func TestMotor_UpdateIdleState(t *testing.T) {
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0, 10}, {1, 20}, {2, 15}},
		TotalMass: 50,
		BurnTime:  2.0,
		AvgThrust: 15,
	}

	motor := components.NewMotor(1, md)

	// Simulate motor burn to completion
	motor.Update(2.5)

	assert.Equal(t, "idle", motor.FSM.GetState(), "Motor FSM should transition to 'idle' state after burn completion")
	assert.Equal(t, float64(0), motor.GetThrust(), "Motor thrust should be zero in the 'idle' state")
	assert.Equal(t, float64(50), motor.GetMass(), "Motor mass should stop decreasing after burn")
}

func TestMotor_UpdateInvalidTimestep(t *testing.T) {
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0, 10}, {1, 20}, {2, 15}},
		TotalMass: 50,
		BurnTime:  2.0,
		AvgThrust: 15,
	}

	motor := components.NewMotor(1, md)

	// Simulate invalid timestep
	err := motor.Update(-1.0)
	assert.Error(t, err, "Motor update with invalid timestep should produce an error")
	assert.Equal(t, "idle", motor.FSM.GetState(), "Motor FSM should remain in 'idle' state")
}

func TestMotor_ThrustInterpolation(t *testing.T) {
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0, 10}, {1, 20}, {2, 15}},
		TotalMass: 50,
		BurnTime:  2.0,
		AvgThrust: 15,
	}

	motor := components.NewMotor(1, md)

	// Simulate motor update to test thrust interpolation
	motor.Update(0.5)
	assert.Equal(t, 15.0, motor.GetThrust(), "Thrust should be interpolated correctly at t=0.5")

	motor.Update(0.5)
	assert.Equal(t, 20.0, motor.GetThrust(), "Thrust should be interpolated correctly at t=1.0")

	motor.Update(0.5)
	assert.Equal(t, 17.5, motor.GetThrust(), "Thrust should be interpolated correctly at t=1.5")
}

func TestMotor_StringRepresentation(t *testing.T) {
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0, 10}, {1, 20}},
		TotalMass: 50,
		BurnTime:  2.0,
		AvgThrust: 15,
	}

	motor := components.NewMotor(1, md)

	expectedString := "Motor{ID: 1, Position: Vector3{X: 0.00, Y: 0.00, Z: 0.00}, Mass: 50.000000, Thrust: 10.000000}"
	assert.Equal(t, expectedString, motor.String(), "String representation of the Motor should match the expected format")
}
