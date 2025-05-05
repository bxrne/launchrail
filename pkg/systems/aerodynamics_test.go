package systems_test

import (
	"io"
	"math"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

func TestCalculateAerodynamicMoment_ZeroVelocity(t *testing.T) {
	testLogger := logf.New(logf.Opts{Writer: io.Discard})
	minCfg := &config.Engine{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						// Provide default values if needed, otherwise zero values are okay for ISA model
					},
				},
			},
		},
	}
	system := systems.NewAerodynamicSystem(&ecs.World{}, 1, minCfg, testLogger) // Pass empty world and logger

	entity := states.PhysicsState{
		Position: &types.Position{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
		Velocity: &types.Velocity{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
		Nosecone: &components.Nosecone{Length: 0.5, Shape: "ogive"},
		Bodytube: &components.Bodytube{Length: 1.0, Radius: 0.05},
	}
	moment := system.CalculateAerodynamicMoment(entity)
	if moment.Y != 0 {
		t.Errorf("Expected zero moment for zero velocity, got %v", moment.Y)
	}
}

func TestCalculateInertia_Cylinder(t *testing.T) {
	entity := &states.PhysicsState{
		Bodytube: &components.Bodytube{Radius: 0.1, Length: 1.0},
		Mass:     &types.Mass{Value: 2.0},
	}
	inertia := systems.CalculateInertia(entity)
	if inertia <= 0 {
		t.Errorf("Expected positive inertia, got %v", inertia)
	}
}

func TestAerodynamicSystem_GetAirDensity_SeaLevel(t *testing.T) {
	testLogger := logf.New(logf.Opts{Writer: io.Discard})
	cfg := &config.Engine{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						SeaLevelDensity:     1.225,
						SeaLevelPressure:    101325,
						SeaLevelTemperature: 288.15,
						SpecificGasConstant:  287.058,
						GravitationalAccel:   9.80665,
						RatioSpecificHeats:   1.4,
						TemperatureLapseRate: -0.0065,
					},
				},
			},
		},
	}

	system := systems.NewAerodynamicSystem(&ecs.World{}, 1, cfg, testLogger) // Pass empty world and logger
	density := system.GetAirDensity(0)

	if math.Abs(density-1.225) > 0.001 {
		t.Errorf("Expected sea level density 1.225, got %f", density)
	}
}

func TestAerodynamicSystem_CalculateDrag(t *testing.T) {
	testLogger := logf.New(logf.Opts{Writer: io.Discard})
	cfg := &config.Engine{
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

	system := systems.NewAerodynamicSystem(&ecs.World{}, 1, cfg, testLogger) // Pass empty world and logger

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

func TestAerodynamicSystem_GetSpeedOfSound(t *testing.T) {
	testLogger := logf.New(logf.Opts{Writer: io.Discard})
	cfg := &config.Engine{
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

	system := systems.NewAerodynamicSystem(&ecs.World{}, 1, cfg, testLogger) // Pass empty world and logger

	// Test at sea level
	speedAtSeaLevel := system.GetSpeedOfSound(0)
	expectedSpeed := 340.29 // Approximate speed of sound at sea level

	if math.Abs(speedAtSeaLevel-expectedSpeed) > 1.0 {
		t.Errorf("Expected speed of sound at sea level to be close to %f, got %f", expectedSpeed, speedAtSeaLevel)
	}
}

func TestAerodynamicSystem_Update(t *testing.T) {
	world := &ecs.World{} // Use pointer to struct literal
	testLogger := logf.New(logf.Opts{Writer: io.Discard})
	cfg := &config.Engine{
		Options: config.Options{
			Launchsite: config.Launchsite{
				Atmosphere: config.Atmosphere{
					ISAConfiguration: config.ISAConfiguration{
						SeaLevelDensity:     1.225,
						SeaLevelPressure:    101325, // Add pressure
						SeaLevelTemperature: 288.15, // Add temperature
					},
				},
			},
		},
	}

	system := systems.NewAerodynamicSystem(world, 2, cfg, testLogger)

	// Create a basic ECS entity
	basicEntity := ecs.NewBasic() // Use ecs.NewBasic() to create the entity ID

	// Create PhysicsState and link to basicEntity
	entity1 := &states.PhysicsState{
		Entity:              &basicEntity, // Assign pointer to the basic entity
		Position:            &types.Position{},
		Velocity:            &types.Velocity{Vec: types.Vector3{Y: 100}},
		Acceleration:        &types.Acceleration{}, // Initialize Acceleration
		Mass:                &types.Mass{Value: 1.0},
		Nosecone:            &components.Nosecone{Radius: 0.1},
		Bodytube:            &components.Bodytube{Radius: 0.1, Length: 1.0},
		Orientation:         &types.Orientation{}, // Initialize Orientation
		AngularVelocity:     &types.Vector3{}, // Initialize AngularVelocity
		AngularAcceleration: &types.Vector3{}, // Initialize AngularAcceleration
		AccumulatedForce:    types.Vector3{}, // Initialize AccumulatedForce
		AccumulatedMoment:   types.Vector3{}, // Initialize AccumulatedMoment
	}

	system.Add(entity1) // Add the PhysicsState wrapper to the system's list

	err := system.Update(0.01) // 10ms timestep
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Verify that forces were accumulated (Update doesn't modify Acceleration directly)
	if entity1.AccumulatedForce.Magnitude() == 0 {
		t.Error("Expected non-zero accumulated force after update")
	}
	// Optional: Check AccumulatedMoment too
}
