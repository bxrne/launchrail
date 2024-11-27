package systems

import (
	"github.com/bxrne/launchrail/pkg/entities"
)

// AeroSystem represents the aerodynamics system.
type AeroSystem struct {
	ecs *entities.ECS
}

// NewAeroSystem creates a new aerodynamics system.
func NewAeroSystem(ecs *entities.ECS) *AeroSystem {
	return &AeroSystem{ecs: ecs}
}

// Update updates the aerodynamics system.
func (as *AeroSystem) Update() {
	// Implement aerodynamics update logic here
}
