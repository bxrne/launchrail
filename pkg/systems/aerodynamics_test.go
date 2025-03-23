package systems_test

import (
	"math"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
)

// TEST: GIVEN a new aerodynamic system WHEN getting air density at sea level THEN returns correct value
func TestAerodynamicSystem_GetAirDensity_SeaLevel(t *testing.T) {
	cfg := &config.Config{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						SeaLevelDensity: 1.225,
					},
				},
			},
		},
	}

	system := systems.NewAerodynamicSystem(nil, 1, cfg)
	density := system.GetAirDensity(0)

	if math.Abs(density-1.225) > 0.001 {
		t.Errorf("Expected sea level density 1.225, got %f", density)
	}
}

// TEST: GIVEN a moving rocket WHEN calculating drag THEN returns correct drag force
func TestAerodynamicSystem_CalculateDrag(t *testing.T) {
	cfg := &config.Config{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						SeaLevelDensity: 1.225,
					},
				},
			},
		},
	}

	system := systems.NewAerodynamicSystem(nil, 1, cfg)

	// Create test entity
	entity := &states.PhysicsState{
		Position: &types.Position{Vec: types.Vector3{Y: 0}},   // Sea level
		Velocity: &types.Velocity{Vec: types.Vector3{Y: 100}}, // 100 m/s upward
		Nosecone: &components.Nosecone{Radius: 0.1},
		Bodytube: &components.Bodytube{Radius: 0.1, Length: 1.0},
	}

	dragForce := system.CalculateDrag(*entity)

	// Verify drag force is negative (opposing motion)
	if dragForce.Y >= 0 {
		t.Errorf("Expected negative drag force, got %f", dragForce.Y)
	}
}

// TEST: GIVEN a rocket at different altitudes WHEN getting speed of sound THEN returns correct values
func TestAerodynamicSystem_GetSpeedOfSound(t *testing.T) {
	cfg := &config.Config{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						SeaLevelTemperature: 288.15,
					},
				},
			},
		},
	}

	system := systems.NewAerodynamicSystem(nil, 1, cfg)

	// Test at sea level
	speedAtSeaLevel := system.GetSpeedOfSound(0)
	expectedSpeed := 340.29 // Approximate speed of sound at sea level

	if math.Abs(speedAtSeaLevel-expectedSpeed) > 1.0 {
		t.Errorf("Expected speed of sound at sea level to be close to %f, got %f", expectedSpeed, speedAtSeaLevel)
	}
}

// TEST: GIVEN a system with multiple entities WHEN updating THEN processes all entities
func TestAerodynamicSystem_Update(t *testing.T) {
	world := ecs.World{}
	cfg := &config.Config{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						SeaLevelDensity: 1.225,
					},
				},
			},
		},
	}

	system := systems.NewAerodynamicSystem(&world, 2, cfg)

	// Add test entities
	entity1 := &states.PhysicsState{
		Position:            &types.Position{},
		Velocity:            &types.Velocity{Vec: types.Vector3{Y: 100}},
		Acceleration:        &types.Acceleration{},
		Mass:                &types.Mass{Value: 1.0},
		Nosecone:            &components.Nosecone{Radius: 0.1},
		Bodytube:            &components.Bodytube{Radius: 0.1, Length: 1.0},
		Orientation:         &types.Orientation{},
		AngularVelocity:     &types.Vector3{},
		AngularAcceleration: &types.Vector3{},
	}

	system.Add(entity1)

	err := system.Update(0.01) // 10ms timestep
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Verify that forces were applied
	if entity1.Acceleration.Vec.Y == 0 {
		t.Error("Expected non-zero acceleration after update")
	}
}
