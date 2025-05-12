package systems

import "github.com/bxrne/launchrail/pkg/states"

// System defines the interface that all systems must implement
type System interface {
	// UpdateWithError updates the system state
	UpdateWithError(dt float64) error

	// Add adds entities to the system
	Add(pe *states.PhysicsState)
}
