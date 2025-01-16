package components_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
)

// TEST: GIVEN a motor component WHEN String is called THEN the correct representation is returned.
func TestMotorString(t *testing.T) {
	thrustcurve := &thrustcurves.MotorData{
		Designation: "TestMotor",
		ID:          "123",
		Thrust:      [][]float64{{0, 100}, {1, 200}},
	}
	motor := components.NewMotor(thrustcurve, 50.0)

	expected := "Motor{Position: {0 0 0}, Mass: 50.00, Thrust: 100.00}"
	if got := motor.String(); got != expected {
		t.Errorf("String() = %v, want %v", got, expected)
	}
}

// TEST: GIVEN a motor component WHEN GetThrust is called THEN the correct thrust is returned.
func TestMotorGetThrust(t *testing.T) {
	thrustcurve := &thrustcurves.MotorData{
		Designation: "TestMotor",
		Thrust:      [][]float64{{0, 100}, {1, 200}},
	}
	motor := components.NewMotor(thrustcurve, 50.0)

	if got := motor.GetThrust(); got != 100 {
		t.Errorf("GetThrust() = %v, want %v", got, 100)
	}

	// Update the motor to change thrust
	motor.Update(0.5)

	if got := motor.GetThrust(); got != 100 {
		t.Errorf("GetThrust() = %v, want %v", got, 100)
	}
}

// TEST: GIVEN a motor component and a delta time WHEN Update is called THEN the component is updated.
func TestMotorUpdate(t *testing.T) {
	thrustcurve := &thrustcurves.MotorData{
		Designation: "TestMotor",
		Thrust:      [][]float64{{0, 100}, {1, 200}, {2, 300}},
	}
	motor := components.NewMotor(thrustcurve, 50.0)

	// initial state
	if motor.GetThrust() != 100 {
		t.Errorf("thrust = %v, want %v", motor.GetThrust(), 100)
	}
	if motor.Mass != 50.0 {
		t.Errorf("Mass = %v, want %v", motor.Mass, 50.0)
	}

	motor.Update(10.0)

	if motor.GetThrust() != 300 {
		t.Errorf("thrust = %v, want %v", motor.GetThrust(), 300)
	}

}

// TEST: GIVEN a motor component WHEN NewMotor is called THEN a new Motor instance is returned.
func TestNewMotor(t *testing.T) {
	thrustcurve := &thrustcurves.MotorData{
		Designation: "TestMotor",
		Thrust:      [][]float64{{0, 100}},
	}
	motor := components.NewMotor(thrustcurve, 50.0)

	if motor.Position != (types.Vector3{X: 0, Y: 0, Z: 0}) {
		t.Errorf("Position = %v, want %v", motor.Position, types.Vector3{X: 0, Y: 0, Z: 0})
	}
	if motor.Mass != 50.0 {
		t.Errorf("Mass = %v, want %v", motor.Mass, 50.0)
	}
	if motor.GetThrust() != 100 {
		t.Errorf("thrust = %v, want %v", motor.GetThrust(), 100)
	}
}
