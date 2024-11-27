package systems

import (
	"github.com/bxrne/launchrail/pkg/entities"
)

// PhysicsSystem represents the physics system.
type PhysicsSystem struct {
	ecs *entities.ECS
}

// NewPhysicsSystem creates a new physics system.
func NewPhysicsSystem(ecs *entities.ECS) *PhysicsSystem {
	return &PhysicsSystem{ecs: ecs}
}

// Update updates the physics system.
func (ps *PhysicsSystem) Update() {
	// Implement physics update logic here
}
