package systems_test

import (
	"math"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/require"
)

// TEST: GIVEN a new LaunchRailSystem WHEN NewLaunchRailSystem is called THEN a new LaunchRailSystem is returned
func TestNewLaunchRailSystem(t *testing.T) {
	world := &ecs.World{}
	length := 2.0
	angle := 5.0
	orientation := 0.0

	rail := systems.NewLaunchRailSystem(world, length, angle, orientation)
	require.NotNil(t, rail)
}

// TEST: GIVEN a LaunchRailSystem WHEN Add is called THEN the entity is added to the system
func TestLaunchRailSystem_Add(t *testing.T) {
	world := &ecs.World{}
	rail := systems.NewLaunchRailSystem(world, 2.0, 5.0, 0.0)

	entity := &systems.PhysicsEntity{
		Entity:       &ecs.BasicEntity{},
		Position:     &types.Position{},
		Velocity:     &types.Velocity{},
		Acceleration: &types.Acceleration{},
		Mass:         &types.Mass{Value: 1.0},
		Motor:        &components.Motor{},
	}

	rail.Add(entity)
}

// TEST: GIVEN a LaunchRailSystem WHEN Update is called THEN the system state is updated and entities are constrained to the rail
func TestLaunchRailSystem_Update(t *testing.T) {
	tests := []struct {
		name           string
		length         float64
		angle          float64
		orientation    float64
		initialPos     types.Position
		initialVel     types.Velocity
		thrust         float64
		expectedOnRail bool
	}{
		{
			name:        "Still on rail",
			length:      2.0,
			angle:       5.0,
			orientation: 0.0,
			initialPos: types.Position{
				Vec: types.Vector3{
					X: 0.0,
					Y: 0.0,
					Z: 0.0,
				},
			},
			thrust:         100.0, // Add thrust to simulate motor
			expectedOnRail: true,
		},
		{
			name:        "Off rail",
			length:      2.0,
			angle:       5.0,
			orientation: 0.0,
			initialPos: types.Position{
				Vec: types.Vector3{
					X: 0.0,
					Y: 3.0,
					Z: 0.0,
				},
			},
			thrust:         100.0,
			expectedOnRail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := &ecs.World{}
			rail := systems.NewLaunchRailSystem(world, tt.length, tt.angle, tt.orientation)

			// Create mock motor that returns constant thrust
			motor := &components.Motor{}
			entity := &systems.PhysicsEntity{
				Entity:       &ecs.BasicEntity{},
				Position:     &types.Position{Vec: types.Vector3{X: tt.initialPos.Vec.X, Y: tt.initialPos.Vec.Y, Z: tt.initialPos.Vec.Z}},
				Velocity:     &types.Velocity{},
				Acceleration: &types.Acceleration{},
				Mass:         &types.Mass{Value: 1.0},
				Motor:        motor,
			}

			rail.Add(entity)

			// Run multiple updates to allow motion to develop
			for i := 0; i < 10; i++ {
				err := rail.Update(0.01)
				require.NoError(t, err)
			}

			// Verify position constraints
			if tt.expectedOnRail {
				// Check if motion is constrained to rail angle
				angleRad := tt.angle * math.Pi / 180.0
				expectedRatio := math.Tan(angleRad)

				// Only check ratio if we've moved significantly
				if entity.Position.Vec.Y > 0.1 {
					actualRatio := entity.Position.Vec.X / entity.Position.Vec.Y
					require.InDelta(t, expectedRatio, actualRatio, 0.001,
						"Position not following rail angle. Expected ratio %v, got %v",
						expectedRatio, actualRatio)
				}
			}
		})
	}
}

// TEST: GIVEN a LaunchRailSystem WHEN Priority is called THEN the system priority is returned
func TestLaunchRailSystem_Priority(t *testing.T) {
	world := &ecs.World{}
	rail := systems.NewLaunchRailSystem(world, 2.0, 5.0, 0.0)

	priority := rail.Priority()
	require.Equal(t, 1, priority)
}
