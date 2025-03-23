package systems_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a new physics system WHEN initialized THEN has correct default values
func TestNewPhysicsSystem(t *testing.T) {
	cfg := &config.Config{
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
	assert.Equal(t, 1, system.Priority())
	assert.Equal(t, "PhysicsSystem", system.String())
}

// TEST: GIVEN an entity with invalid mass WHEN calculating net force THEN returns zero
func TestCalculateNetForce_InvalidMass(t *testing.T) {
	cfg := &config.Config{
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
	cfg := &config.Config{
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
		Mass:                &types.Mass{Value: 1.0},
		Position:            &types.Position{Vec: types.Vector3{Y: 10}},
		Velocity:            &types.Velocity{},
		Acceleration:        &types.Acceleration{},
		Orientation:         &types.Orientation{},
		AngularVelocity:     &types.Vector3{Y: 1.0},
		AngularAcceleration: &types.Vector3{Y: 0.5},
	}

	system.Add(entity)
	err := system.Update(0.01)
	assert.NoError(t, err)
	assert.Greater(t, entity.AngularVelocity.Y, 1.0)
}

// TEST: GIVEN an entity with invalid timestep WHEN updating THEN returns error
func TestUpdate_InvalidTimestep(t *testing.T) {
	system := systems.NewPhysicsSystem(&ecs.World{}, &config.Config{})
	err := system.Update(0)
	assert.Error(t, err)
}

// TEST: GIVEN a system with invalid entity WHEN updating THEN returns error
func TestUpdate_InvalidEntity(t *testing.T) {
	system := systems.NewPhysicsSystem(&ecs.World{}, &config.Config{})
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
	system := systems.NewPhysicsSystem(&ecs.World{}, &config.Config{})
	entity := &states.PhysicsState{
		Entity: &ecs.BasicEntity{},
		Mass:   &types.Mass{Value: 1.0},
	}

	system.Add(entity)
	system.Remove(*entity.Entity)
}
