package systems_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/stretchr/testify/require"
)

// TEST: GIVEN a new AerodynamicSystem WHEN NewAerodynamicSystem is called THEN a new AerodynamicSystem is returned
func TestNewAerodynamicSystem(t *testing.T) {
	world := &ecs.World{}
	workers := 1
	cfg := &config.Config{}
	aero := systems.NewAerodynamicSystem(world, workers, cfg)
	require.NotNil(t, aero)
}

// TEST: GIVEN an AerodynamicSystem WHEN CalculateDrag is called THEN the drag force is calculated
func TestAerodynamicSystem_CalculateDrag(t *testing.T) {
	world := &ecs.World{}
	workers := 1
	cfg := &config.Config{}
	aero := systems.NewAerodynamicSystem(world, workers, cfg)
	require.NotNil(t, aero)

	// Create a physics entity
	entity := systems.PhysicsEntity{
		Entity:       &ecs.BasicEntity{},
		Position:     &components.Position{Y: 0},
		Velocity:     &components.Velocity{X: 0},
		Acceleration: &components.Acceleration{},
		Mass:         &components.Mass{},
		Motor:        &components.Motor{},
		Bodytube:     &components.Bodytube{},
		Nosecone:     &components.Nosecone{},
	}

	// Calculate drag
	drag := aero.CalculateDrag(entity)
	require.NotNil(t, drag)
}

// TEST: GIVEN an AerodynamicSystem WHEN Update is called THEN the system state is updated
func TestAerodynamicSystem_Update(t *testing.T) {
	world := &ecs.World{}
	workers := 1
	cfg := &config.Config{}
	aero := systems.NewAerodynamicSystem(world, workers, cfg)
	require.NotNil(t, aero)

	err := aero.Update(0.1)
	require.NoError(t, err)
}

// TEST: GIVEN a new AerodynamicsSystem WHEN Add is called THEN the entity is added to the system
func TestAerodynamicSystem_Add(t *testing.T) {
	world := &ecs.World{}
	workers := 1
	cfg := &config.Config{}
	aero := systems.NewAerodynamicSystem(world, workers, cfg)
	require.NotNil(t, aero)

	// Create a physics entity
	entity := systems.PhysicsEntity{
		Entity:       &ecs.BasicEntity{},
		Position:     &components.Position{Y: 0},
		Velocity:     &components.Velocity{X: 0},
		Acceleration: &components.Acceleration{},
		Mass:         &components.Mass{},
		Motor:        &components.Motor{},
		Bodytube:     &components.Bodytube{},
		Nosecone:     &components.Nosecone{},
	}

	aero.Add(&entity)
}

// TEST: GIVEN a new AerodynamicsSystem WHEN Priority is called THEN the system priority is returned
func TestAerodynamicSystem_Priority(t *testing.T) {
	world := &ecs.World{}
	workers := 1
	cfg := &config.Config{}
	aero := systems.NewAerodynamicSystem(world, workers, cfg)
	require.NotNil(t, aero)

	priority := aero.Priority()
	require.Equal(t, 2, priority)
}

// TEST: GIVEN a new AerodynamicsSystem WHEN GetSpeedOfSound is called THEN the speed of sound is returned
func TestAerodynamicSystem_GetSpeedOfSound(t *testing.T) {
	world := &ecs.World{}
	workers := 1
	cfg := &config.Config{}
	aero := systems.NewAerodynamicSystem(world, workers, cfg)
	require.NotNil(t, aero)

	speed := aero.GetSpeedOfSound(20)
	require.Equal(t, float32(340.29), speed)
}
