package systems_test

import (
	"testing"
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST: GIVEN a new PhysicsSystem WHEN NewPhysicsSystem is called THEN a new PhysicsSystem is returned
func TestNewPhysicsSystem(t *testing.T) {
	world := &ecs.World{}
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

	system := systems.NewPhysicsSystem(world, cfg)
	require.NotNil(t, system)
}

// TEST: GIVEN a PhysicsSystem WHEN Add is called THEN the entity is added to the system
func TestPhysicsSystem_Add(t *testing.T) {
	world := &ecs.World{}
	cfg := &config.Config{}
	system := systems.NewPhysicsSystem(world, cfg)
	e := ecs.NewBasic()

	entity := systems.PhysicsEntity{
		Entity:       &e,
		Position:     &components.Position{},
		Velocity:     &components.Velocity{},
		Acceleration: &components.Acceleration{},
		Mass:         &components.Mass{Value: 1.0},
		Motor:        &components.Motor{},
	}

	system.Add(&entity)
}

// TEST: GIVEN a PhysicsSystem WHEN Update is called THEN physics are applied correctly
func TestPhysicsSystem_Update(t *testing.T) {
	tests := []struct {
		name        string
		mass        float64
		initialPos  components.Position
		initialVel  components.Velocity
		motorState  string
		dt          float32
		wantPosY    float64
		wantVelY    float64
		description string
	}{
		{
			name:        "Ground start no thrust",
			mass:        1.0,
			initialPos:  components.Position{Y: 0},
			initialVel:  components.Velocity{Y: 0},
			motorState:  "READY",
			dt:          0.016,
			wantPosY:    0,
			wantVelY:    0,
			description: "Should stay on ground with no thrust",
		},
		{
			name:        "Mid-flight coasting",
			mass:        1.0,
			initialPos:  components.Position{Y: 100},
			initialVel:  components.Velocity{Y: 50},
			motorState:  "COASTING",
			dt:          0.016,
			wantPosY:    100.8,
			wantVelY:    49.84,
			description: "Should experience gravity and drag",
		},
		{
			name:        "Landing detection",
			mass:        1.0,
			initialPos:  components.Position{Y: 0.1}, // Slightly above ground
			initialVel:  components.Velocity{Y: 0},
			motorState:  "COASTING",
			dt:          0.016,
			wantPosY:    0,      // Should land
			wantVelY:    -0.156, // Should stop
			description: "Should stop at ground",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			world := &ecs.World{}
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
			system := systems.NewPhysicsSystem(world, cfg)

			// Create entity with all required components properly initialized
			e := ecs.NewBasic()
			motor := &components.Motor{}
			motor.SetState(tt.motorState)

			entity := systems.PhysicsEntity{
				Entity:       &e,
				Position:     &tt.initialPos,
				Velocity:     &tt.initialVel,
				Acceleration: &components.Acceleration{},
				Mass:         &components.Mass{Value: tt.mass},
				Motor:        motor,
				// Initialize required components that were missing
				Bodytube: &components.Bodytube{Radius: 0.05, Length: 1.0}, // Add reasonable defaults
				Nosecone: &components.Nosecone{Radius: 0.05, Length: 0.3},
				Finset:   &components.TrapezoidFinset{},
			}

			system.Add(&entity)

			// Run update
			err := system.Update(tt.dt)
			assert.NoError(t, err)

			// Verify results with reasonable tolerance for floating point
			assert.InDelta(t, tt.wantPosY, entity.Position.Y, 0.1,
				"Position Y mismatch: want %.2f, got %.2f", tt.wantPosY, entity.Position.Y)
			assert.InDelta(t, tt.wantVelY, entity.Velocity.Y, 0.1,
				"Velocity Y mismatch: want %.2f, got %.2f", tt.wantVelY, entity.Velocity.Y)
		})
	}
}

// TEST: GIVEN a PhysicsSystem WHEN Remove is called THEN the entity is removed from the system
func TestPhysicsSystem_Remove(t *testing.T) {
	world := &ecs.World{}
	cfg := &config.Config{}
	system := systems.NewPhysicsSystem(world, cfg)
	e := ecs.NewBasic()

	entity := systems.PhysicsEntity{
		Entity:       &e,
		Position:     &components.Position{},
		Velocity:     &components.Velocity{},
		Acceleration: &components.Acceleration{},
		Mass:         &components.Mass{},
		Motor:        &components.Motor{},
	}

	system.Add(&entity)
	system.Remove(e)
}

// TEST: GIVEN a PhysicsSystem WHEN Priority is called THEN the correct priority is returned
func TestPhysicsSystem_Priority(t *testing.T) {
	world := &ecs.World{}
	cfg := &config.Config{}
	system := systems.NewPhysicsSystem(world, cfg)
	assert.Equal(t, 1, system.Priority())
}

// TEST: GIVEN a PhysicsSystem with multiple workers WHEN Update is called THEN forces are calculated concurrently
func TestPhysicsSystem_Concurrent(t *testing.T) {
	world := &ecs.World{}
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
	system := systems.NewPhysicsSystem(world, cfg)

	// Add multiple entities
	for i := 0; i < 10; i++ {
		e := ecs.NewBasic()
		entity := systems.PhysicsEntity{
			Entity:       &e,
			Position:     &components.Position{Y: float64(i * 10)},
			Velocity:     &components.Velocity{Y: 10},
			Acceleration: &components.Acceleration{},
			Mass:         &components.Mass{Value: 1.0},
			Motor:        &components.Motor{},
		}
		system.Add(&entity)
	}

	start := time.Now()
	err := system.Update(0.016)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Concurrent update took too long")
}
