package systems

import "github.com/bxrne/launchrail/pkg/states"

// System defines the interface that all systems must implement
type System interface {
	// Update updates the system state
	Update(dt float64) error

	// Add adds entities to the system
	Add(pe *states.PhysicsState)
	// Priority returns the system priority for execution order
	Priority() int
}
