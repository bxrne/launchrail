package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
)

// System defines the interface that all systems must implement
type System interface {
	// Update updates the system state
	Update(dt float32) error

	// Add adds entities to the system
	Add(se *SystemEntity)

	// Priority returns the system priority for execution order
	Priority() int
}

// SystemEntity represents an entity with physics components (Meta rocket)
type SystemEntity struct {
	Entity   *ecs.BasicEntity
	Pos      *components.Position
	Vel      *components.Velocity
	Acc      *components.Acceleration
	Mass     *components.Mass
	Motor    *components.Motor
	Bodytube *components.Bodytube
	Nosecone *components.Nosecone
	Finset   *components.TrapezoidFinset
}
