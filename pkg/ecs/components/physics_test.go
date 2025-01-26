package components_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/stretchr/testify/assert"
)

func TestPhysics_InitialState(t *testing.T) {
	physics := components.NewPhysics(9.8, 10.0)

	assert.Equal(t, 10.0, physics.Mass, "Mass should be initialized correctly")
	assert.Equal(t, types.Vector3{}, physics.Position, "Initial position should be a zero vector")
	assert.Equal(t, types.Vector3{}, physics.Velocity, "Initial velocity should be a zero vector")
	assert.Equal(t, types.Vector3{X: 0, Y: -9.8, Z: 0}, physics.Gravity, "Gravity should be initialized correctly")
}

func TestPhysics_AddForce(t *testing.T) {
	physics := components.NewPhysics(9.8, 10.0)
	force := types.Vector3{X: 1, Y: 0, Z: 0}

	physics.AddForce(force)
	assert.Len(t, physics.Forces, 1, "Forces should contain one element after adding a force")
	assert.Equal(t, force, physics.Forces[0], "Added force should match the input")
}

func TestPhysics_Update_ForceApplication(t *testing.T) {
	physics := components.NewPhysics(9.8, 10.0) // Gravity: -9.8, Mass: 10
	physics.AddForce(types.Vector3{X: 10, Y: 0, Z: 0})

	err := physics.Update(1.0)
	assert.NoError(t, err, "Update should not return an error")

	expectedVelocity := types.Vector3{X: 1, Y: -0.98, Z: 0} // Gravity scaled by mass
	assert.Equal(t, expectedVelocity.Round(2), physics.Velocity.Round(2), "Velocity should reflect applied forces and gravity")

	expectedPosition := types.Vector3{X: 1, Y: -0.98, Z: 0} // Position updated based on velocity
	assert.Equal(t, expectedPosition.Round(2), physics.Position.Round(2), "Position should update based on velocity")

	assert.Empty(t, physics.Forces, "Forces should be cleared after update")
}

func TestPhysics_Update_MultipleForces(t *testing.T) {
	physics := components.NewPhysics(9.8, 10.0)
	physics.AddForce(types.Vector3{X: 5, Y: 10, Z: 0})
	physics.AddForce(types.Vector3{X: -2, Y: -5, Z: 0})

	err := physics.Update(1.0)
	assert.NoError(t, err, "Update should not return an error")
	assert.Equal(t, types.Vector3{X: 0.3, Y: -0.48, Z: 0}, physics.Velocity.Round(2), "Velocity should account for all forces")
	assert.Equal(t, types.Vector3{X: 0.3, Y: -0.48, Z: 0}, physics.Position.Round(2), "Position should update based on combined forces")
	assert.Empty(t, physics.Forces, "Forces should be cleared after update")
}

func TestPhysics_Update_InvalidDt(t *testing.T) {
	physics := components.NewPhysics(9.8, 10.0)

	err := physics.Update(0.0)
	assert.Error(t, err, "Update with zero dt should return an error")
	assert.EqualError(t, err, "invalid timestep: dt must be > 0", "Error message should match expected")

	err = physics.Update(-1.0)
	assert.Error(t, err, "Update with negative dt should return an error")
	assert.EqualError(t, err, "invalid timestep: dt must be > 0", "Error message should match expected")
}

func TestPhysics_PanicOnInvalidMass(t *testing.T) {
	assert.Panics(t, func() {
		components.NewPhysics(9.8, 0.0)
	}, "Creating a Physics component with zero mass should panic")

	assert.Panics(t, func() {
		components.NewPhysics(9.8, -5.0)
	}, "Creating a Physics component with negative mass should panic")
}
