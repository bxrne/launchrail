package components_test

import (
	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAerodynamics(t *testing.T) {
	dragCoefficient := 0.5
	area := 1.0

	// Create a new Aerodynamics component
	aero := components.NewAerodynamics(dragCoefficient, area)

	// Assert the values are correctly set
	assert.NotNil(t, aero, "Aerodynamics component should be created")
	assert.Equal(t, dragCoefficient, aero.DragCoefficient, "Drag coefficient should match")
	assert.Equal(t, area, aero.Area, "Area should match")
}

func TestAerodynamics_CalculateDrag(t *testing.T) {
	// Test for different velocities
	aero := components.NewAerodynamics(0.5, 1.0)

	// Test drag at zero velocity (should result in zero drag)
	velocityZero := types.Vector3{X: 0, Y: 0, Z: 0}
	dragZero := aero.CalculateDrag(velocityZero)
	assert.Equal(t, types.Vector3{}, dragZero, "Drag should be zero for zero velocity")

	// Test drag at non-zero velocity
	velocity := types.Vector3{X: 10, Y: 0, Z: 0}
	drag := aero.CalculateDrag(velocity)

	// Calculate the expected drag magnitude
	expectedDragMagnitude := 0.5 * aero.DragCoefficient * 1.225 * aero.Area * velocity.Magnitude() * velocity.Magnitude()

	// Ensure the drag magnitude is correct
	assert.Equal(t, drag.Magnitude(), expectedDragMagnitude, "Calculated drag force magnitude should match")

	// Ensure the direction is opposite of velocity
	normalizedVel := velocity.DivideScalar(velocity.Magnitude())
	expectedDragForce := normalizedVel.MultiplyScalar(-expectedDragMagnitude)
	assert.Equal(t, drag, expectedDragForce, "Drag force should be in the opposite direction of velocity")
}

func TestAerodynamics_String(t *testing.T) {
	dragCoefficient := 0.5
	area := 1.0

	// Create a new Aerodynamics component
	aero := components.NewAerodynamics(dragCoefficient, area)

	// Expected string format
	expected := "Aerodynamics{DragCoefficient: 0.50}"

	// Check if String() method matches
	assert.Equal(t, expected, aero.String(), "String representation should match expected format")
}
