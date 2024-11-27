package systems

import (
	"github.com/bxrne/launchrail/pkg/entities"
)

// Simulation represents the simulation.
type Simulation struct {
	ecs           *entities.ECS
	physicsSystem *PhysicsSystem
	aeroSystem    *AeroSystem
}

// NewSimulation creates a new simulation.
func NewSimulation(ecs *entities.ECS) *Simulation {
	return &Simulation{
		ecs:           ecs,
		physicsSystem: NewPhysicsSystem(ecs),
		aeroSystem:    NewAeroSystem(ecs),
	}
}

// Run runs the simulation.
func (s *Simulation) Run() {
	// Implement simulation run logic here
	// Example:
	for i := 0; i < 100; i++ { // Simulate 100 ticks
		s.physicsSystem.Update()
		s.aeroSystem.Update()
	}
}
