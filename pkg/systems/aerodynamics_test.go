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
