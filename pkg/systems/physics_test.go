package systems_test

import (
	"io"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
)

// Create a logger that discards output for tests
var testLogger = logf.New(logf.Opts{Writer: io.Discard})

// TEST: GIVEN a new physics system WHEN initialized THEN has correct default values
func TestNewPhysicsSystem(t *testing.T) {
	cfg := &config.Engine{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						GravitationalAccel: 9.81,
					},
				},
			},
		},
		Simulation: config.Simulation{
			GroundTolerance: 0.1,
		},
	}

	system := systems.NewPhysicsSystem(&ecs.World{}, cfg, testLogger, 1)
	assert.NotNil(t, system)
	assert.Equal(t, "PhysicsSystem", system.String())
}

// TEST: GIVEN an entity with invalid mass WHEN calculating net force THEN returns zero
func TestCalculateNetForce_InvalidMass(t *testing.T) {
	cfg := &config.Engine{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						GravitationalAccel: 9.81,
					},
				},
			},
		},
	}

	system := systems.NewPhysicsSystem(&ecs.World{}, cfg, testLogger, 1)
	entity := &states.PhysicsState{
		Mass:         &types.Mass{Value: 0},
		Position:     &types.Position{},
		Velocity:     &types.Velocity{},
		Acceleration: &types.Acceleration{},
		Nosecone:     &components.Nosecone{},
		Bodytube:     &components.Bodytube{},
	}

	system.Add(entity)
	err := system.Update(0.01)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, entity.Acceleration.Vec.Y)
}

// TEST: GIVEN an entity with rotation WHEN updating THEN updates angular state
func TestUpdate_AngularMotion(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)

	// Create minimal valid motor data
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0.0, 0.0}, {1.0, 0.0}},
		TotalMass: 0.1,
		BurnTime:  1.0,
	}
	motor, err := components.NewMotor(ecs.NewBasic(), md, testLogger)
	assert.NoError(t, err)
	e := ecs.NewBasic()

	// Create the state with necessary fields, including Nosecone/Bodytube
	entity := &states.PhysicsState{
		Entity:       &e,                       // Need a basic entity ID
		Mass:         &types.Mass{Value: 10.0}, // Use a mass
		Position:     &types.Position{Vec: types.Vector3{Y: 10}},
		Velocity:     &types.Velocity{}, // Start stationary for simplicity
		Acceleration: &types.Acceleration{},
		// Initialize Orientation, AngularVelocity, AngularAcceleration
		Orientation:         &types.Orientation{Quat: *types.NewQuaternion(0, 0, 0, 1)}, // Start level
		AngularVelocity:     &types.Vector3{Y: 1.0},                                     // Initial angular velocity around Y
		AngularAcceleration: &types.Vector3{Y: 0.5},                                     // Initial angular acceleration around Y
		Motor:               motor,
		// Add dummy Nosecone and Bodytube
		Nosecone: &components.Nosecone{Length: 0.5, Radius: 0.1},
		Bodytube: &components.Bodytube{Length: 1.5, Radius: 0.1},
	}

	system.Add(entity)
	err = system.Update(0.01)
	assert.NoError(t, err)

	// Check if angular velocity increased as expected based on initial acceleration
	// The updateEntityState function integrates AngularVelocity using AngularAcceleration
	expectedAngularVelocityY := 1.0 + (0.5 * 0.01)
	assert.InDelta(t, expectedAngularVelocityY, entity.AngularVelocity.Y, 1e-9, "Angular velocity Y did not update correctly based on initial angular acceleration")

	// Also check orientation changed (simple check: not identity anymore)
	assert.False(t, entity.Orientation.Quat.IsIdentity(), "Orientation should change after angular update")
}

// TEST: GIVEN an entity with invalid timestep WHEN updating THEN returns error
func TestUpdate_InvalidTimestep(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)
	err := system.Update(0)
	assert.Error(t, err)
}

// TEST: GIVEN a system with invalid entity WHEN updating THEN returns error
func TestUpdate_InvalidEntity(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)
	entity := &states.PhysicsState{
		// Missing required fields
	}

	system.Add(entity)
	err := system.Update(0.01)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity missing required vectors")
}

// TEST: GIVEN a system with entity WHEN removing entity THEN entity is removed
func TestRemoveEntity(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)
	e := ecs.NewBasic()
	entity := &states.PhysicsState{
		Entity: &e,
		Mass:   &types.Mass{Value: 1.0},
	}

	system.Add(entity)
	system.Remove(e)
}

// TEST: GIVEN an entity with missing components WHEN updating THEN returns error
func TestUpdate_MissingComponents(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)
	e := ecs.NewBasic()
	entity := &states.PhysicsState{Entity: &e} // Missing Mass, Position etc.
	system.Add(entity)

	err := system.Update(0.01)
	// The system's validateEntity should catch this
	assert.Error(t, err, "Update should error if required state components are missing")
	// Test missing Nosecone/Bodytube (used by calculateReferenceArea within calculateNetForce)
	entityWithMass := &states.PhysicsState{
		Entity: &e,
		Mass:   &types.Mass{Value: 1.0}, Position: &types.Position{}, Velocity: &types.Velocity{}, Acceleration: &types.Acceleration{},
		// Missing Nosecone/Bodytube
	}
	system.Remove(e) // Remove the previous invalid entity before adding a new one
	system.Add(entityWithMass)
	// NOTE: calculateReferenceArea inside physics.go *will* panic if Nosecone or Bodytube are nil.
	// This test case reveals that calculateNetForce/calculateReferenceArea needs nil checks.
	// For now, we expect a panic. We should fix this in physics.go later.
	err = system.Update(0.01)
	assert.Error(t, err, "Update should error if geometry components needed for drag are missing")
}

// TEST: GIVEN an entity at ground level WHEN updating THEN ground collision handled
func TestUpdate_GroundCollision(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)
	e := ecs.NewBasic()
	entity := &states.PhysicsState{
		Entity:       &e,
		Mass:         &types.Mass{Value: 1.0},
		Position:     &types.Position{Vec: types.Vector3{Y: 0.0}},  // Start at ground
		Velocity:     &types.Velocity{Vec: types.Vector3{Y: -1.0}}, // Moving downwards
		Acceleration: &types.Acceleration{},
		Orientation:  &types.Orientation{Quat: types.Quaternion{X: 0, Y: 0, Z: 0, W: 1}}, AngularVelocity: &types.Vector3{}, AngularAcceleration: &types.Vector3{},
		Nosecone: &components.Nosecone{Length: 0.5, Radius: 0.1}, Bodytube: &components.Bodytube{Length: 1.5, Radius: 0.1}, // Placeholders
	}

	system.Add(entity)
	err := system.Update(0.01)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, entity.Velocity.Vec.Y, "Velocity should be zero after ground collision")
	assert.Equal(t, 0.0, entity.Position.Vec.Y, "Position should remain at ground level after collision")
}
