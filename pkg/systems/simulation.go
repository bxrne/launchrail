package systems

import (
	"github.com/bxrne/launchrail/pkg/entities"
	"time"
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
	tickDuration := 100 * time.Millisecond // Example tick duration
	for i := 0; i < 100; i++ {             // Simulate 100 ticks
		s.physicsSystem.Update(tickDuration)
		s.aeroSystem.Update()
		time.Sleep(tickDuration) // Simulate real-time
	}
}

// GetStates returns the stored states from the physics system.
func (s *Simulation) GetStates() []map[entities.Entity]State {
	return s.physicsSystem.GetStates()
}
