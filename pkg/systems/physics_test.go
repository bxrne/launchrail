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
	"github.com/stretchr/testify/require"
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

	// Create a basic entity ID
	basicEntity := ecs.NewBasic()

	entity := &states.PhysicsState{
		BasicEntity:  basicEntity,
		Mass:         nil, // Explicitly set to nil
		Position:     &types.Position{},
		Velocity:     &types.Velocity{},
		Acceleration: &types.Acceleration{},
		Nosecone:     &components.Nosecone{},
		Bodytube:     &components.Bodytube{},
	}

	system.Add(entity)
	t.Logf("[Test] Before Update: entity.Mass is nil? %v", entity.Mass == nil)
	err := system.UpdateWithError(0.01)
	assert.Error(t, err) // Assert that an error occurred
	if err == nil {
		t.Fatal("CalculateNetForce unexpectedly returned nil error, stopping test to prevent panic")
	}
	assert.Contains(t, err.Error(), "entity missing mass") // Check for the correct error substring
}

// TEST: GIVEN an entity with rotation WHEN updating THEN updates angular state
func TestUpdate_AngularMotion(t *testing.T) {
	world := &ecs.World{}
	// Provide a config with gravity for the test
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
	system := systems.NewPhysicsSystem(world, cfg, testLogger, 1)

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
		BasicEntity:  e,
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
	err = system.UpdateWithError(0.01)
	assert.NoError(t, err)

	// Assertions:
	require.NoError(t, err, "Update should not return an error")

	// Check if force was applied correctly (e.g., gravity)
	require.InDelta(t, 0, entity.AccumulatedForce.X, 1e-9, "Force X should be zero")
	require.InDelta(t, -98.1, entity.AccumulatedForce.Y, 1e-9, "Force Y should be gravity * mass")
	require.InDelta(t, 0, entity.AccumulatedForce.Z, 1e-9, "Force Z should be zero")

	// Angular velocity and orientation are no longer updated by PhysicsSystem directly.
	// Integration happens in the main simulation loop.
	// Removing checks for angular velocity and orientation changes here.
	/*
		require.InDelta(t, 1.005, entity.AngularVelocity.Y, 1e-9, "Angular velocity Y did not update correctly based on initial angular acceleration")
		require.False(t, entity.Orientation.Quat.IsIdentity(), "Orientation should change after angular update")
	*/
}

// TEST: GIVEN an entity with invalid timestep WHEN updating THEN returns error
func TestUpdate_InvalidTimestep(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)
	err := system.UpdateWithError(0)
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
	err := system.UpdateWithError(0.01)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity missing required vectors")
}

// TEST: GIVEN a system with entity WHEN removing entity THEN entity is removed
func TestRemoveEntity(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)
	e := ecs.NewBasic()
	entity := &states.PhysicsState{
		BasicEntity: e,
		Mass:        &types.Mass{Value: 1.0},
	}

	system.Add(entity)
	system.Remove(e)
}

// TEST: GIVEN an entity with missing components WHEN updating THEN returns error
func TestUpdate_MissingComponents(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewPhysicsSystem(world, &config.Engine{}, testLogger, 1)
	e := ecs.NewBasic()
	entity := &states.PhysicsState{BasicEntity: e} // Missing Mass, Position etc.
	system.Add(entity)

	err := system.UpdateWithError(0.01)
	// The system's validateEntity should catch this
	assert.Error(t, err, "Update should error if required state components are missing")
	// Test missing Nosecone/Bodytube (used by calculateReferenceArea within calculateNetForce)
	entityWithMass := &states.PhysicsState{
		BasicEntity: e,
		Mass:        &types.Mass{Value: 1.0}, Position: &types.Position{}, Velocity: &types.Velocity{}, Acceleration: &types.Acceleration{},
		// Missing Nosecone/Bodytube
	}
	system.Remove(e) // Remove the previous invalid entity before adding a new one
	system.Add(entityWithMass)
	// Set nonzero velocity to trigger drag/geometry checks
	if entityWithMass.Velocity != nil {
		entityWithMass.Velocity.Vec.Y = 1
	}
	// NOTE: calculateReferenceArea inside physics.go *will* panic if Nosecone or Bodytube are nil.
	// This test case reveals that calculateNetForce/calculateReferenceArea needs nil checks.
	// For now, we expect a panic. We should fix this in physics.go later.
	err = system.UpdateWithError(0.01)
	assert.Error(t, err, "Update should error if geometry components needed for drag are missing")
}

// TEST: GIVEN an entity at ground level WHEN updating THEN ground collision handled
func TestUpdate_GroundCollision(t *testing.T) {
	world := &ecs.World{}
	// Provide a config with gravity for the test
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
			GroundTolerance: 0.1, // Also provide ground tolerance from config
		},
	}
	system := systems.NewPhysicsSystem(world, cfg, testLogger, 1)
	e := ecs.NewBasic()
	entity := &states.PhysicsState{
		BasicEntity:  e,
		Mass:         &types.Mass{Value: 1.0},
		Position:     &types.Position{Vec: types.Vector3{Y: 0.0}},  // Start at ground
		Velocity:     &types.Velocity{Vec: types.Vector3{Y: -1.0}}, // Moving downwards
		Acceleration: &types.Acceleration{},
		Orientation:  &types.Orientation{Quat: types.Quaternion{X: 0, Y: 0, Z: 0, W: 1}}, AngularVelocity: &types.Vector3{}, AngularAcceleration: &types.Vector3{},
		Nosecone: &components.Nosecone{Length: 0.5, Radius: 0.1}, Bodytube: &components.Bodytube{Length: 1.5, Radius: 0.1}, // Placeholders
	}

	system.Add(entity)
	err := system.UpdateWithError(0.01)
	assert.NoError(t, err)

	// Assertions:
	// PhysicsSystem no longer handles ground collision directly; this happens in simulation loop.
	// We can check that forces were applied, but velocity/position clamping won't happen here.
	// For example, check accumulated force (should include gravity)
	require.InDelta(t, -9.81*entity.Mass.Value, entity.AccumulatedForce.Y, 1e-9, "Gravity should still be accumulated")

	// Removing checks for velocity/position clamping, as it's done in simulation loop.
	/*
		require.Equal(t, 0.0, entity.Position.Vec.Y, "Position should be clamped to ground")
		require.Equal(t, 0.0, entity.Velocity.Vec.Y, "Velocity should be zero after ground collision")
		require.Equal(t, 0.0, entity.Velocity.Vec.X, "Velocity should be zero after ground collision")
		require.Equal(t, 0.0, entity.Velocity.Vec.Z, "Velocity should be zero after ground collision")
	*/
}
