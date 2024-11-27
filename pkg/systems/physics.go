package systems

import (
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/entities"
	"time"
)

// PhysicsSystem represents the physics system.
type PhysicsSystem struct {
	ecs          *entities.ECS
	timeElapsed  time.Duration
	stateManager *StateManager
}

// NewPhysicsSystem creates a new physics system.
func NewPhysicsSystem(ecs *entities.ECS) *PhysicsSystem {
	return &PhysicsSystem{
		ecs:          ecs,
		stateManager: NewStateManager(),
	}
}

// Update updates the physics system.
func (ps *PhysicsSystem) Update(deltaTime time.Duration) {
	ps.timeElapsed += deltaTime

	// Apply phantom thrust for the first 5 seconds
	if ps.timeElapsed <= 5*time.Second {
		ps.applyPhantomThrust(deltaTime)
	}

	// Update accelerations, velocities, and positions
	ps.updateAccelerations()
	ps.updateVelocities(deltaTime)
	ps.updatePositions(deltaTime)

	// Store the current state
	ps.stateManager.StoreState(ps.ecs)
}

// applyPhantomThrust applies a phantom thrust to the rocket.
func (ps *PhysicsSystem) applyPhantomThrust(deltaTime time.Duration) {
	for entity := entities.Entity(1); entity < ps.ecs.GetNextEntity(); entity++ {
		if acceleration, ok := ps.ecs.GetComponent(entity, "Acceleration").(*components.Acceleration); ok {
			// Apply a constant thrust in the Z direction
			acceleration.Z += 10.0 * deltaTime.Seconds() // Example thrust
		}
	}
}

// updateAccelerations updates the accelerations based on forces.
func (ps *PhysicsSystem) updateAccelerations() {
	// For simplicity, we assume no other forces except thrust and drag (handled in AeroSystem)
}

// updateVelocities updates the velocities based on accelerations.
func (ps *PhysicsSystem) updateVelocities(deltaTime time.Duration) {
	for entity := entities.Entity(1); entity < ps.ecs.GetNextEntity(); entity++ {
		if velocity, ok := ps.ecs.GetComponent(entity, "Velocity").(*components.Velocity); ok {
			if acceleration, ok := ps.ecs.GetComponent(entity, "Acceleration").(*components.Acceleration); ok {
				// Update velocity based on acceleration
				velocity.X += acceleration.X * deltaTime.Seconds()
				velocity.Y += acceleration.Y * deltaTime.Seconds()
				velocity.Z += acceleration.Z * deltaTime.Seconds()
			}
		}
	}
}

// updatePositions updates the positions based on velocities.
func (ps *PhysicsSystem) updatePositions(deltaTime time.Duration) {
	for entity := entities.Entity(1); entity < ps.ecs.GetNextEntity(); entity++ {
		if velocity, ok := ps.ecs.GetComponent(entity, "Velocity").(*components.Velocity); ok {
			if position, ok := ps.ecs.GetComponent(entity, "Position").(*components.Position); ok {
				// Update position based on velocity
				position.X += velocity.X * deltaTime.Seconds()
				position.Y += velocity.Y * deltaTime.Seconds()
				position.Z += velocity.Z * deltaTime.Seconds()
			}
		}
	}
}

// GetStates returns the stored states.
func (ps *PhysicsSystem) GetStates() []map[entities.Entity]State {
	return ps.stateManager.GetStates()
}
