package systems_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST: GIVEN a new RulesSystem WHEN NewRulesSystem is called THEN a new RulesSystem is returned
func TestNewRulesSystem(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewRulesSystem(world)
	require.NotNil(t, system)
}

// TEST: GIVEN a RulesSystem WHEN Add is called THEN the entity is added to the system
func TestRulesSystem_Add(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewRulesSystem(world)
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
}

// TEST: GIVEN a RulesSystem WHEN Priority is called THEN the correct priority is returned
func TestRulesSystem_Priority(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewRulesSystem(world)
	assert.Equal(t, 100, system.Priority())
}

// TEST: GIVEN a RulesSystem WHEN Update is called with various flight conditions THEN appropriate events are detected
func TestRulesSystem_Update(t *testing.T) {
	tests := []struct {
		name          string
		position      components.Position
		velocity      components.Velocity
		motorState    string
		expectedEvent systems.Event
		description   string
	}{
		{
			name:          "Pre-apogee ascending",
			position:      components.Position{Y: 100},
			velocity:      components.Velocity{Y: 10},
			motorState:    "BURNOUT",
			expectedEvent: systems.None,
			description:   "Should not detect apogee while ascending",
		},
		{
			name:          "Apogee detection",
			position:      components.Position{Y: 100},
			velocity:      components.Velocity{Y: -0.1},
			motorState:    "BURNOUT",
			expectedEvent: systems.Apogee,
			description:   "Should detect apogee when velocity turns negative",
		},
		{
			name:          "Post-apogee descending",
			position:      components.Position{Y: 50},
			velocity:      components.Velocity{Y: -10},
			motorState:    "BURNOUT",
			expectedEvent: systems.None,
			description:   "Should not detect any event during descent",
		},
		{
			name:          "Landing detection",
			position:      components.Position{Y: 0},
			velocity:      components.Velocity{Y: -5},
			motorState:    "BURNOUT",
			expectedEvent: systems.Land,
			description:   "Should detect landing when reaching ground with negative velocity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := &ecs.World{}
			system := systems.NewRulesSystem(world)
			e := ecs.NewBasic()

			// Create position, velocity and motor with initial states
			pos := tt.position
			vel := tt.velocity
			motor := &components.Motor{}
			motor.SetState(tt.motorState)

			// Create physics entity with test conditions
			entity := systems.PhysicsEntity{
				Entity:       &e,
				Position:     &pos,
				Velocity:     &vel,
				Acceleration: &components.Acceleration{},
				Mass:         &components.Mass{},
				Motor:        motor,
			}

			// Add entity to system
			system.Add(&entity)

			// If testing landing conditions, need to simulate apogee first
			if tt.expectedEvent == systems.Land {
				// First simulate apogee
				entity.Position.Y = 100
				entity.Velocity.Y = -0.1
				entity.Motor.SetState("BURNOUT")
				err := system.Update(0.016)
				assert.NoError(t, err)

				// Then simulate landing conditions
				entity.Position.Y = 0
				entity.Velocity.Y = -5
			}

			// Run the update
			err := system.Update(0.016)
			assert.NoError(t, err)

			// Verify state based on expected event
			switch tt.expectedEvent {
			case systems.Apogee:
				assert.True(t, entity.Velocity.Y < 0, "Velocity should be negative at apogee")
				assert.Equal(t, "BURNOUT", entity.Motor.GetState(), "Motor should be burned out at apogee")
			case systems.Land:
				assert.Equal(t, float64(0), entity.Position.Y, "Position should be 0 at landing")
				assert.Equal(t, float64(0), entity.Velocity.Y, "Velocity should be 0 at landing")
				assert.Equal(t, float64(0), entity.Acceleration.Y, "Acceleration should be 0 at landing")
			case systems.None:
				if tt.name == "Pre-apogee ascending" {
					assert.True(t, entity.Velocity.Y > 0, "Velocity should be positive while ascending")
				}
			}
		})
	}
}

// TEST: GIVEN a RulesSystem WHEN Remove is called THEN the entity is removed from the system
func TestRulesSystem_Remove(t *testing.T) {
	world := &ecs.World{}
	system := systems.NewRulesSystem(world)
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
