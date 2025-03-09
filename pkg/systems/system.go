package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/types"
)

// System defines the interface that all systems must implement
type System interface {
	// Update updates the system state
	Update(dt float64) error

	// Add adds entities to the system
	Add(pe *PhysicsEntity)
	// Priority returns the system priority for execution order
	Priority() int
}

// PhysicsEntity represents an entity with physics components (Meta rocket but could be reused for payload?)
type PhysicsEntity struct {
	Entity       *ecs.BasicEntity
	Position     *types.Position
	Velocity     *types.Velocity
	Acceleration *types.Acceleration
	Orietation   *types.Orientation
	Mass         *types.Mass
	Motor        *components.Motor
	Bodytube     *components.Bodytube
	Nosecone     *components.Nosecone
	Finset       *components.TrapezoidFinset
	Parachute    *components.Parachute
}
