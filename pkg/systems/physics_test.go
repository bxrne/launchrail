package systems_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/assert"
	"github.com/zerodha/logf"
)

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

	system := systems.NewPhysicsSystem(&ecs.World{}, cfg)
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

	system := systems.NewPhysicsSystem(&ecs.World{}, cfg)
	entity := &states.PhysicsState{
		Mass:         &types.Mass{Value: 0},
		Position:     &types.Position{},
		Velocity:     &types.Velocity{},
		Acceleration: &types.Acceleration{},
	}

	system.Add(entity)
	err := system.Update(0.01)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, entity.Acceleration.Vec.Y)
}

// TEST: GIVEN an entity with rotation WHEN updating THEN updates angular state
func TestUpdate_AngularMotion(t *testing.T) {
	system := systems.NewPhysicsSystem(&ecs.World{}, &config.Engine{})

	// Create minimal valid motor data
	logger := logf.New(logf.Opts{})
	md := &thrustcurves.MotorData{
		Thrust:    [][]float64{{0.0, 0.0}, {1.0, 0.0}}, // Minimal 2-point curve
		TotalMass: 0.1,                              // Example mass
		BurnTime:  1.0,                              // Match thrust curve
	}
	motor, err := components.NewMotor(ecs.NewBasic(), md, logger)
	assert.NoError(t, err, "Failed to create test motor")

	entity := &states.PhysicsState{
		Mass:                &types.Mass{Value: 1.0},
		Position:            &types.Position{Vec: types.Vector3{Y: 10}},
		Velocity:            &types.Velocity{},
		Acceleration:        &types.Acceleration{},
		Orientation:         &types.Orientation{},
		AngularVelocity:     &types.Vector3{Y: 1.0},
		AngularAcceleration: &types.Vector3{Y: 0.5},
		Motor:               motor, // Use constructed motor
	}

	system.Add(entity)
	err = system.Update(0.01) // Use the error from NewMotor creation
	assert.NoError(t, err)
	// Check if angular velocity increased as expected
	expectedAngularVelocityY := 1.0 + (0.5 * 0.01)
	assert.InDelta(t, expectedAngularVelocityY, entity.AngularVelocity.Y, 1e-9, "Angular velocity Y did not update correctly")
}

// TEST: GIVEN an entity with invalid timestep WHEN updating THEN returns error
func TestUpdate_InvalidTimestep(t *testing.T) {
	system := systems.NewPhysicsSystem(&ecs.World{}, &config.Engine{})
	err := system.Update(0)
	assert.Error(t, err)
}

// TEST: GIVEN a system with invalid entity WHEN updating THEN returns error
func TestUpdate_InvalidEntity(t *testing.T) {
	system := systems.NewPhysicsSystem(&ecs.World{}, &config.Engine{})
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
	system := systems.NewPhysicsSystem(&ecs.World{}, &config.Engine{})
	entity := &states.PhysicsState{
		Entity: &ecs.BasicEntity{},
		Mass:   &types.Mass{Value: 1.0},
	}

	system.Add(entity)
	system.Remove(*entity.Entity)
}
