package systems

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs"
)

type ParasiteSystem struct{}

func NewParasiteSystem() *ParasiteSystem {
	return &ParasiteSystem{}
}

func (s *ParasiteSystem) Priority() int {
	return 0 // Runs first
}

func (s *ParasiteSystem) Update(world *ecs.World, dt float64) error {
	// Query entities that have all required components
	entities := world.Query(
		ecs.ComponentPhysics,
		ecs.ComponentAerodynamics,
	)

	for _, entity := range entities {
		// Get components
		physComp, _ := world.GetComponent(entity, ecs.ComponentPhysics)
		aeroComp, _ := world.GetComponent(entity, ecs.ComponentAerodynamics)

		// pull and print the ComponentPhysics
		fmt.Println(physComp)
		fmt.Println(aeroComp)
	}

	return nil
}
